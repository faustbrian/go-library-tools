package services

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestManagerStartsGenericFixturesAndCleansExactContainers(t *testing.T) {
	backend := &fakeBackend{}
	probes := 0
	manager := Manager{
		Process: backend.run,
		Probe: func(context.Context, string, string) error {
			probes++
			return nil
		},
		Wait:     func(context.Context, time.Duration) error { return nil },
		Token:    func() (string, error) { return "task123", nil },
		Attempts: 2,
	}
	requested := []string{"valkey", "postgresql", "rabbitmq", "redis", "nats", "nsq"}
	lease, err := manager.Start(context.Background(), requested)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if probes != 2 {
		t.Fatalf("TCP probes = %d", probes)
	}
	wantIdentities := map[string]string{
		"nats": identity(natsImage), "nsq": identity(nsqImage),
		"postgresql": identity(postgresImage), "rabbitmq": identity(rabbitMQImage),
		"redis": identity(redisImage), "valkey": identity(valkeyImage),
	}
	if !reflect.DeepEqual(lease.Identities(), wantIdentities) {
		t.Fatalf("Identities() = %#v", lease.Identities())
	}
	environment := lease.Environment()
	for _, name := range []string{
		"POSTGRES_URL", "POSTGRES_VERSION", "OUTBOX_POSTGRES_VERSION", "STATE_MACHINE_POSTGRES_VERSION",
		"FEATURE_FLAGS_POSTGRES_DSN", "TEST_DATABASE_URL", "DATABASE_URL", "TEMPORAL_POSTGRES_DSN",
		"VALKEY_ADDR", "VALKEY_ADDRESS", "FEATURE_FLAGS_VALKEY_ADDRESS", "TEST_VALKEY_ADDRESS", "CACHE_VALKEY_IMAGE",
		"REDIS_ADDR", "TEST_REDIS_ADDRESS", "CACHE_REDIS_IMAGE", "NATS_URL", "NSQD_TCP_ADDRESS", "RABBITMQ_URL",
	} {
		if environment[name] == "" {
			t.Fatalf("Environment()[%q] is empty", name)
		}
	}
	environment["POSTGRES_URL"] = "changed"
	if lease.Environment()["POSTGRES_URL"] == "changed" {
		t.Fatal("Environment() returned mutable state")
	}
	identities := lease.Identities()
	identities["nats"] = "changed"
	if lease.Identities()["nats"] == "changed" {
		t.Fatal("Identities() returned mutable state")
	}

	if err := lease.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatalf("Close() second error = %v", err)
	}
	removed := backend.removed()
	wantRemoved := []string{
		"golib-valkey-task123", "golib-redis-task123", "golib-rabbitmq-task123",
		"golib-postgresql-task123", "golib-nsq-task123", "golib-nats-task123",
	}
	if !reflect.DeepEqual(removed, wantRemoved) {
		t.Fatalf("removed = %#v, want %#v", removed, wantRemoved)
	}
	if !strings.Contains(strings.Join(backend.commands, "\n"), "nsqio/nsq:v1.3.0 /nsqd --broadcast-address=127.0.0.1") {
		t.Fatal("NSQ daemon arguments missing")
	}
}

func TestManagerValidatesCompleteRequestBeforeStarting(t *testing.T) {
	for _, test := range []struct {
		name     string
		services []string
		token    Token
		want     string
	}{
		{"empty", nil, nil, "service selection is empty"},
		{"unknown", []string{"unknown"}, nil, "unsupported service"},
		{"duplicate", []string{"redis", "redis"}, nil, "duplicate service"},
		{"token failure", []string{"redis"}, func() (string, error) { return "", errors.New("entropy") }, "create service token"},
		{"unsafe token", []string{"redis"}, func() (string, error) { return "../unsafe", nil }, "service token is malformed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			backend := &fakeBackend{}
			manager := Manager{Process: backend.run, Probe: successfulProbe, Wait: successfulWait, Token: test.token}
			_, err := manager.Start(context.Background(), test.services)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Start() error = %v, want %q", err, test.want)
			}
			if len(backend.commands) != 0 {
				t.Fatalf("commands = %#v", backend.commands)
			}
		})
	}
	manager := Manager{}
	if _, err := manager.Start(context.Background(), []string{"redis"}); err == nil || !strings.Contains(err.Error(), "process backend") {
		t.Fatalf("Start() missing backend error = %v", err)
	}
	backend := &fakeBackend{}
	manager = Manager{Process: backend.run, Probe: successfulProbe, Wait: successfulWait}
	lease, err := manager.Start(context.Background(), []string{"redis"})
	if err != nil {
		t.Fatalf("Start() default token error = %v", err)
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatalf("Close() default token error = %v", err)
	}
}

func TestManagerCleansStartedResourcesAfterFailures(t *testing.T) {
	failure := errors.New("injected failure")
	for _, test := range []struct {
		name    string
		backend *fakeBackend
		probe   Probe
		wait    Wait
		want    string
	}{
		{"run", &fakeBackend{failCommand: "run", failure: failure}, successfulProbe, successfulWait, "start redis"},
		{"inspect", &fakeBackend{failCommand: "inspect", failure: failure}, successfulProbe, successfulWait, "read redis image identity"},
		{"malformed identity", &fakeBackend{imageOutput: "sha256:bad"}, successfulProbe, successfulWait, "image identity is malformed"},
		{"port", &fakeBackend{failCommand: "port", failure: failure}, successfulProbe, successfulWait, "read redis port"},
		{"malformed port", &fakeBackend{portOutput: "not-a-port"}, successfulProbe, successfulWait, "parse redis port"},
		{"exec readiness", &fakeBackend{failCommand: "exec", failure: failure}, successfulProbe, successfulWait, "did not become ready"},
		{"TCP readiness", &fakeBackend{}, func(context.Context, string, string) error { return failure }, successfulWait, "did not become ready"},
		{"wait", &fakeBackend{failCommand: "exec", failure: failure}, successfulProbe, func(context.Context, time.Duration) error { return failure }, "wait for redis"},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := "redis"
			if test.name == "TCP readiness" {
				service = "nats"
			}
			manager := Manager{Process: test.backend.run, Probe: test.probe, Wait: test.wait, Token: fixedToken, Attempts: 2}
			_, err := manager.Start(context.Background(), []string{service})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Start() error = %v, want %q", err, test.want)
			}
			if test.name != "run" && len(test.backend.removed()) != 1 {
				t.Fatalf("removed = %#v", test.backend.removed())
			}
		})
	}
}

func TestLeaseCloseReportsAllCleanupFailuresConcurrently(t *testing.T) {
	failure := errors.New("cleanup failed")
	backend := &fakeBackend{failRemove: true, failure: failure}
	manager := Manager{Process: backend.run, Probe: successfulProbe, Wait: successfulWait, Token: fixedToken}
	lease, err := manager.Start(context.Background(), []string{"redis", "valkey"})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	errorsSeen := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Go(func() {
			errorsSeen <- lease.Close(context.Background())
		})
	}
	group.Wait()
	close(errorsSeen)
	for err := range errorsSeen {
		if err == nil || !strings.Contains(err.Error(), "remove service container") {
			t.Fatalf("Close() error = %v", err)
		}
	}
	if len(backend.removed()) != 2 {
		t.Fatalf("removed = %#v", backend.removed())
	}
}

func TestPublishedPortRejectsHostAndPortAmbiguity(t *testing.T) {
	for _, test := range []struct {
		value string
		want  string
		ok    bool
	}{
		{"127.0.0.1:49152\n", "49152", true},
		{"127.0.0.1:0", "", false},
		{"127.0.0.1:1", "1", true},
		{"127.0.0.1:not-a-port", "", false},
		{"127.0.0.1:65535", "65535", true},
		{"127.0.0.1:65536", "", false},
		{"", "", false},
		{"0.0.0.0:49152", "", false},
		{"127.0.0.1:49152\n127.0.0.1:49153", "", false},
		{strings.Repeat("x", maximumPortOutput+1), "", false},
	} {
		got, err := publishedPort(strings.NewReader(test.value))
		if test.ok && (err != nil || got != test.want) {
			t.Fatalf("publishedPort(%q) = %q, %v", test.value, got, err)
		}
		if !test.ok && err == nil {
			t.Fatalf("publishedPort(%q) error = nil", test.value)
		}
	}
	if _, err := publishedPort(failingReader{}); err == nil || !strings.Contains(err.Error(), "read published port") {
		t.Fatalf("publishedPort(read failure) error = %v", err)
	}
}

func TestPublishedPortAcceptsExactOutputLimit(t *testing.T) {
	suffix := "127.0.0.1:49152"
	value := strings.Repeat(" ", maximumPortOutput-len(suffix)) + suffix
	if port, err := publishedPort(strings.NewReader(value)); err != nil || port != "49152" {
		t.Fatalf("publishedPort(exact limit) = %q, %v", port, err)
	}
}

func TestImageIdentityRejectsNonHexDigest(t *testing.T) {
	backend := &fakeBackend{imageOutput: "sha256:" + strings.Repeat("g", 64)}
	manager := Manager{Process: backend.run}
	if _, err := manager.imageIdentity(context.Background(), "container", redisImage); err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("imageIdentity() error = %v", err)
	}
}

func TestReadinessDefaultsAndRuntimeHelpers(t *testing.T) {
	manager := Manager{Process: (&fakeBackend{}).run, Attempts: -1}
	if err := manager.waitReady(context.Background(), catalog["redis"], "container", "1234"); err == nil || !strings.Contains(err.Error(), "attempts must be positive") {
		t.Fatalf("waitReady(negative attempts) error = %v", err)
	}

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := tcpProbe(context.Background(), "tcp", address); err != nil {
		t.Fatalf("tcpProbe() error = %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tcpProbe(context.Background(), "tcp", address); err == nil {
		t.Fatal("tcpProbe(closed) error = nil")
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitContext(canceled, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("waitContext(canceled) error = %v", err)
	}
	if err := waitContext(context.Background(), 0); err != nil {
		t.Fatalf("waitContext(timer) error = %v", err)
	}

	failing := &fakeBackend{failCommand: "exec", failure: errors.New("not ready")}
	manager = Manager{Process: failing.run, Attempts: 2}
	if err := manager.waitReady(canceled, catalog["redis"], "container", "1234"); !errors.Is(err, context.Canceled) {
		t.Fatalf("waitReady(default wait) error = %v", err)
	}

	token, err := randomToken()
	if err != nil || !tokenRE.MatchString(token) {
		t.Fatalf("randomToken() = %q, %v", token, err)
	}
	if _, err := randomTokenFrom(failingReader{}); err == nil {
		t.Fatal("randomTokenFrom() error = nil")
	}

	var writer boundedWriter
	if _, err := writer.Write([]byte("port")); err != nil {
		t.Fatalf("boundedWriter.Write() error = %v", err)
	}
	var exact boundedWriter
	if count, err := exact.Write([]byte(strings.Repeat("x", maximumPortOutput))); err != nil || count != maximumPortOutput {
		t.Fatalf("exact Write = %d, %v", count, err)
	}
	if _, err := writer.Write([]byte(strings.Repeat("x", maximumPortOutput))); err == nil {
		t.Fatal("boundedWriter.Write(overflow) error = nil")
	}
}

type fakeBackend struct {
	mu           sync.Mutex
	commands     []string
	environments []map[string]string
	port         int
	portOutput   string
	imageOutput  string
	failCommand  string
	failRemove   bool
	failure      error
}

func (backend *fakeBackend) run(_ context.Context, name string, args []string, environment map[string]string, stdout, _ io.Writer) error {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	command := name + " " + strings.Join(args, " ")
	backend.commands = append(backend.commands, command)
	backend.environments = append(backend.environments, clone(environment))
	operation := ""
	if len(args) > 0 {
		operation = args[0]
	}
	if operation == "rm" && backend.failRemove {
		return backend.failure
	}
	if operation == backend.failCommand {
		return backend.failure
	}
	if operation == "port" {
		value := backend.portOutput
		if value == "" {
			backend.port++
			value = "127.0.0.1:" + strconv.Itoa(40000+backend.port) + "\n"
		}
		_, _ = io.WriteString(stdout, value)
	}
	if operation == "inspect" {
		value := backend.imageOutput
		if value == "" {
			value = "sha256:" + strings.Repeat("a", 64)
		}
		_, _ = io.WriteString(stdout, value)
	}
	return nil
}

func (backend *fakeBackend) removed() []string {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	result := make([]string, 0)
	for _, command := range backend.commands {
		if name, found := strings.CutPrefix(command, "docker rm --force "); found {
			result = append(result, name)
		}
	}
	return result
}

func fixedToken() (string, error) { return "task", nil }

func successfulProbe(context.Context, string, string) error { return nil }

func successfulWait(context.Context, time.Duration) error { return nil }

func identity(image string) string { return image + "#sha256:" + strings.Repeat("a", 64) }

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
