package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestManagerStartsFullParallelSafeRabbitStreamTopology(t *testing.T) {
	backend := &fakeBackend{}
	files := &recordingServiceFiles{}
	requests := make([]map[string]any, 0, 5)
	manager := Manager{
		Process: backend.run, Token: fixedToken, Wait: successfulWait, Workspace: t.TempDir(), Files: files,
		Secret: func(size int) (string, error) { return strings.Repeat("a", size*2), nil },
		HTTPRequest: func(_ context.Context, _, _ string, body []byte, _ map[string]string) (int, error) {
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("proxy payload is invalid JSON: %q: %v", body, err)
			}
			requests = append(requests, payload)
			return 201, nil
		},
	}
	lease, err := manager.Start(context.Background(), []string{"rabbitstream"})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if len(requests) != 5 {
		t.Fatalf("proxy requests = %#v", requests)
	}
	environment := lease.Environment()
	for key, want := range map[string]string{
		"RABBITSTREAM_TEST_PORT":              "40001",
		"RABBITSTREAM_CLUSTER_PORTS":          "40003,40004,40005",
		"RABBITSTREAM_TLS_PORT":               "40006",
		"RABBITSTREAM_CLUSTER_PROJECT":        "codex-rabbitstream-task-cluster",
		"RABBITSTREAM_TEST_RESTART_CONTAINER": "codex-rabbitstream-tasksingle-rabbit",
		"RABBITSTREAM_UPGRADE_FROM_VERSION":   "4.3.4",
		"RABBITSTREAM_UPGRADE_TO_VERSION":     "4.3.5",
	} {
		if environment[key] != want {
			t.Fatalf("Environment()[%q] = %q, want %q", key, environment[key], want)
		}
	}
	if !strings.Contains(environment["RABBITSTREAM_CLUSTER_CONTAINERS"], "40003=codex-rabbitstream-task-cluster-rabbit1-1") ||
		!strings.HasSuffix(environment["RABBITSTREAM_CLUSTER_COMPOSE"], "/rabbitstream-task/compose.yaml") ||
		!strings.HasSuffix(environment["RABBITSTREAM_TLS_RUNTIME"], "/rabbitstream-task/tls") {
		t.Fatalf("topology environment = %#v", environment)
	}
	identities := lease.Identities()
	if len(identities) != 1 || !strings.Contains(identities["rabbitstream"], rabbitStreamOldImage) ||
		!strings.Contains(identities["rabbitstream"], rabbitStreamImage) || !strings.Contains(identities["rabbitstream"], toxiproxyImage) {
		t.Fatalf("Identities() = %#v", identities)
	}
	compose := files.writes[environment["RABBITSTREAM_CLUSTER_COMPOSE"]]
	var decoded map[string]any
	if err := yaml.Unmarshal(compose, &decoded); err != nil {
		t.Fatalf("generated Compose is invalid YAML: %v\n%s", err, compose)
	}
	for _, required := range []string{
		rabbitStreamOldImage, rabbitStreamImage, "RABBITSTREAM_VOLUME_TLS_CERTS",
	} {
		if !strings.Contains(string(compose), required) {
			t.Fatalf("generated Compose is missing %q", required)
		}
	}
	for name, required := range map[string]string{
		"rabbit1.conf":      "stream.advertised_port = 40003",
		"rabbit2.conf":      "stream.advertised_port = 40004",
		"rabbit3.conf":      "stream.advertised_port = 40005",
		"tls-rabbitmq.conf": "stream.advertised_tls_port = 40006",
	} {
		if content := files.writes[filepath.Join(filepath.Dir(environment["RABBITSTREAM_CLUSTER_COMPOSE"]), name)]; !strings.Contains(string(content), required) {
			t.Fatalf("%s is missing %q: %s", name, required, content)
		}
	}
	commands := strings.Join(backend.commands, "\n")
	if strings.Contains(commands, "127.0.0.1:15561:") || strings.Contains(commands, environment["RABBITSTREAM_TLS_PASSWORD"]) {
		t.Fatalf("commands use fixed ports or expose credentials: %s", commands)
	}
	composeEnvironment := map[string]string(nil)
	for index, command := range backend.commands {
		if strings.HasPrefix(command, "docker compose ") {
			composeEnvironment = backend.environments[index]
		}
	}
	if composeEnvironment["RABBITSTREAM_USER"] != environment["RABBITSTREAM_TEST_USER"] ||
		composeEnvironment["RABBITSTREAM_PASSWORD"] != environment["RABBITSTREAM_TEST_PASSWORD"] {
		t.Fatal("standalone and full topology credentials differ")
	}
	for _, target := range []string{"ca.pem", "client.pem", "client-key.pem", "untrusted-client.pem", "untrusted-client-key.pem"} {
		if _, ok := files.modes[filepath.Join(environment["RABBITSTREAM_TLS_RUNTIME"], target)]; !ok {
			t.Fatalf("TLS mode missing for %s", target)
		}
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	cleanup := strings.Join(backend.commands, "\n")
	for _, expected := range []string{
		"docker volume rm --force codex-rabbitstream-task-tls-data",
		"docker network rm codex-rabbitstream-task-network",
		"docker network rm codex-rabbitstream-tasksingle",
	} {
		if !strings.Contains(cleanup, expected) {
			t.Fatalf("cleanup missing %q", expected)
		}
	}
}

func TestFullRabbitStreamValidatesWorkspaceAndCredentials(t *testing.T) {
	backend := &fakeBackend{}
	if _, err := (Manager{Process: backend.run, Token: fixedToken}).Start(context.Background(), []string{"rabbitstream"}); err == nil || !strings.Contains(err.Error(), "absolute task workspace") {
		t.Fatalf("Start(relative workspace) error = %v", err)
	}
	manager := Manager{
		Process: backend.run, Token: fixedToken, Workspace: t.TempDir(),
		Secret: func(int) (string, error) { return "", errors.New("standalone entropy") },
	}
	if _, err := manager.Start(context.Background(), []string{"rabbitstream"}); err == nil || !strings.Contains(err.Error(), "username") {
		t.Fatalf("Start(standalone failure) error = %v", err)
	}
	if _, err := newRabbitStreamCredentials(func(size int) (string, error) {
		return strings.Repeat("a", size*2), nil
	}, "invalid", "invalid"); err == nil || !strings.Contains(err.Error(), "shared credentials") {
		t.Fatalf("newRabbitStreamCredentials(invalid shared) error = %v", err)
	}
	for _, test := range []struct {
		name   string
		failAt int
		value  string
		want   string
	}{
		{"cookie", 4, "", "cookie"}, {"restricted username", 5, "", "restricted username"},
		{"restricted password", 6, "", "restricted password"}, {"malformed", 0, "aa", "malformed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			manager := Manager{
				Process: (&fakeBackend{}).run, Token: fixedToken, Wait: successfulWait,
				Workspace: t.TempDir(), Files: &recordingServiceFiles{},
				Secret: func(size int) (string, error) {
					calls++
					if calls == test.failAt {
						return "", errors.New("entropy")
					}
					if test.value != "" && calls >= 4 {
						return test.value, nil
					}
					return strings.Repeat("a", size*2), nil
				},
				HTTPRequest: successfulServiceRequest,
			}
			if _, err := manager.Start(context.Background(), []string{"rabbitstream"}); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Start() error = %v", err)
			}
		})
	}
}

func TestRabbitStreamCredentialsRejectEachSharedField(t *testing.T) {
	secret := func(size int) (string, error) { return strings.Repeat("a", size*2), nil }
	validUser := "rabbitstream-" + strings.Repeat("a", 16)
	validPassword := strings.Repeat("b", 48)
	for name, values := range map[string][2]string{
		"prefix": {strings.Repeat("a", 16), validPassword}, "user length": {"rabbitstream-aa", validPassword},
		"user hex": {"rabbitstream-zzzzzzzzzzzzzzzz", validPassword}, "password length": {validUser, "bb"},
		"password hex": {validUser, strings.Repeat("z", 48)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := newRabbitStreamCredentials(secret, values[0], values[1]); err == nil {
				t.Fatal("accepted malformed credentials")
			}
		})
	}
}

func TestFullRabbitStreamUsesDefaultProxyClient(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		if request.URL.Path != "/proxies" {
			t.Errorf("path = %q", request.URL.Path)
		}
		response.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	port := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	topology := newRabbitStreamTopology(t.TempDir(), "task")
	topology.ports[8474] = port
	if err := (Manager{}).configureRabbitStreamProxies(context.Background(), topology); err != nil || requests != 4 {
		t.Fatalf("configureRabbitStreamProxies() = %v, requests = %d", err, requests)
	}
}

func TestFullRabbitStreamReportsFileFailures(t *testing.T) {
	for _, test := range []struct {
		name      string
		operation string
		at        int
		want      string
	}{
		{"standalone workspace", "mkdir", 1, "create RabbitMQ Streams standalone workspace"},
		{"workspace", "mkdir", 2, "create RabbitMQ Streams workspace"},
		{"TLS runtime", "mkdir", 3, "create RabbitMQ Streams TLS runtime"},
		{"standalone plugins", "write", 1, "write RabbitMQ Streams standalone plugins"},
		{"standalone config", "write", 2, "write RabbitMQ Streams standalone configuration"},
		{"fixture", "write", 3, "write RabbitMQ Streams fixture file"},
		{"Compose", "write", 8, "write RabbitMQ Streams topology"},
		{"TLS mode", "chmod", 1, "protect RabbitMQ Streams TLS file"},
	} {
		t.Run(test.name, func(t *testing.T) {
			files := &recordingServiceFiles{failureOperation: test.operation, failureAt: test.at}
			manager := fullRabbitStreamManager(t, (&fakeBackend{}).run, files)
			if _, err := manager.Start(context.Background(), []string{"rabbitstream"}); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Start() error = %v", err)
			}
		})
	}
}

func TestFullRabbitStreamCleansPartialDockerFailures(t *testing.T) {
	for _, test := range []struct {
		name       string
		operation  string
		occurrence int
		want       string
	}{
		{"network", "network", 2, "topology network"},
		{"volume", "volume", 3, "topology volume"},
		{"proxy", "run", 3, "topology proxy"},
		{"first port", "port", 3, "topology port 15561"},
		{"last port", "port", 7, "topology port 8474"},
		{"Compose", "compose", 1, "start RabbitMQ Streams topology"},
		{"cluster", "exec", 2, "form RabbitMQ Streams cluster"},
		{"TLS user", "exec", 3, "TLS authorization"},
		{"TLS permissions", "exec", 4, "TLS authorization"},
		{"TLS copy", "cp", 3, "copy RabbitMQ Streams TLS file"},
		{"proxy identity", "inspect", 3, "proxy image identity"},
		{"last identity", "inspect", 8, "rabbit-tls image identity"},
	} {
		t.Run(test.name, func(t *testing.T) {
			backend := &occurrenceBackend{failureOperation: test.operation, failureOccurrence: test.occurrence}
			manager := fullRabbitStreamManager(t, backend.run, &recordingServiceFiles{})
			if _, err := manager.Start(context.Background(), []string{"rabbitstream"}); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Start() error = %v", err)
			}
			if test.name != "network" && !strings.Contains(strings.Join(backend.commands, "\n"), "docker network rm codex-rabbitstream-tasksingle") {
				t.Fatalf("standalone cleanup missing: %#v", backend.commands)
			}
		})
	}
}

func TestFullRabbitStreamReportsEachProxyFailure(t *testing.T) {
	for failAt := 2; failAt <= 5; failAt++ {
		t.Run(string(rune('0'+failAt)), func(t *testing.T) {
			calls := 0
			manager := fullRabbitStreamManager(t, (&fakeBackend{}).run, &recordingServiceFiles{})
			manager.Attempts = 1
			manager.HTTPRequest = func(context.Context, string, string, []byte, map[string]string) (int, error) {
				calls++
				if calls == failAt {
					return 500, nil
				}
				return 201, nil
			}
			if _, err := manager.Start(context.Background(), []string{"rabbitstream"}); err == nil || !strings.Contains(err.Error(), "configure") {
				t.Fatalf("Start() error = %v", err)
			}
		})
	}
}

func TestFullRabbitStreamDefaultsAndOperatingFiles(t *testing.T) {
	root := t.TempDir()
	files := operatingFiles{}
	directory := filepath.Join(root, "nested")
	if err := files.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "file")
	if err := files.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := files.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o640 {
		t.Fatalf("file mode = %v, %v", info, err)
	}

	backend := &fakeBackend{}
	manager := fullRabbitStreamManager(t, backend.run, files)
	manager.Files = nil
	manager.Secret = nil
	manager.HTTPRequest = successfulServiceRequest
	manager.Process = func(ctx context.Context, name string, args []string, environment map[string]string, stdout, stderr io.Writer) error {
		if len(args) > 0 && args[0] == "cp" {
			return os.WriteFile(args[len(args)-1], []byte("certificate"), 0o600)
		}
		return backend.run(ctx, name, args, environment, stdout, stderr)
	}
	lease, err := manager.Start(context.Background(), []string{"rabbitstream"})
	if err != nil {
		t.Fatalf("Start(defaults) error = %v", err)
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestLeaseCloseReportsVolumeCleanupFailure(t *testing.T) {
	failure := errors.New("volume cleanup failed")
	lease := &Lease{
		volumes: []string{"codex-rabbitstream-task-data"},
		process: func(context.Context, string, []string, map[string]string, io.Writer, io.Writer) error { return failure },
	}
	if err := lease.Close(context.Background()); err == nil || !strings.Contains(err.Error(), "remove service volume") {
		t.Fatalf("Close() error = %v", err)
	}
}

func fullRabbitStreamManager(t *testing.T, process Process, files FileSystem) Manager {
	t.Helper()
	return Manager{
		Process: process, Token: fixedToken, Wait: successfulWait, Workspace: t.TempDir(), Files: files,
		Secret:      func(size int) (string, error) { return strings.Repeat("a", size*2), nil },
		HTTPRequest: successfulServiceRequest,
	}
}

func successfulServiceRequest(context.Context, string, string, []byte, map[string]string) (int, error) {
	return 201, nil
}

type recordingServiceFiles struct {
	writes           map[string][]byte
	modes            map[string]os.FileMode
	counts           map[string]int
	failureOperation string
	failureAt        int
}

func (files *recordingServiceFiles) fail(operation string) error {
	if files.counts == nil {
		files.counts = make(map[string]int)
	}
	files.counts[operation]++
	if files.failureOperation == operation && files.counts[operation] == files.failureAt {
		return errors.New("injected file failure")
	}
	return nil
}

func (files *recordingServiceFiles) MkdirAll(string, os.FileMode) error { return files.fail("mkdir") }

func (files *recordingServiceFiles) WriteFile(path string, data []byte, mode os.FileMode) error {
	if err := files.fail("write"); err != nil {
		return err
	}
	if files.writes == nil {
		files.writes = make(map[string][]byte)
	}
	if files.modes == nil {
		files.modes = make(map[string]os.FileMode)
	}
	files.writes[path] = append([]byte(nil), data...)
	files.modes[path] = mode
	return nil
}

func (files *recordingServiceFiles) Chmod(path string, mode os.FileMode) error {
	if err := files.fail("chmod"); err != nil {
		return err
	}
	if files.modes == nil {
		files.modes = make(map[string]os.FileMode)
	}
	files.modes[path] = mode
	return nil
}

func TestRabbitStreamTopologyDeterminism(t *testing.T) {
	left := newRabbitStreamTopology("/tmp/task", "token")
	right := newRabbitStreamTopology("/tmp/task", "token")
	if !reflect.DeepEqual(left, right) || left.project != "codex-rabbitstream-token-cluster" {
		t.Fatalf("topologies differ: %#v %#v", left, right)
	}
}
