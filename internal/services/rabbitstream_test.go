package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestManagerStartsParallelSafeRabbitStreamStandalone(t *testing.T) {
	backend := &fakeBackend{}
	files := &recordingServiceFiles{}
	var requestMethod, requestURL string
	var requestBody []byte
	var requestHeaders map[string]string
	manager := Manager{
		Process: backend.run, Token: fixedToken, Wait: successfulWait, Workspace: t.TempDir(), Files: files,
		Secret: func(size int) (string, error) { return strings.Repeat("a", size*2), nil },
		HTTPRequest: func(_ context.Context, method, url string, body []byte, headers map[string]string) (int, error) {
			requestMethod, requestURL = method, url
			requestBody, requestHeaders = append([]byte(nil), body...), clone(headers)
			return http.StatusCreated, nil
		},
	}
	lease, err := manager.Start(context.Background(), []string{"rabbitstream-standalone"})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	wantEnvironment := map[string]string{
		"RABBITSTREAM_TEST_HOST":              "localhost",
		"RABBITSTREAM_TEST_PORT":              "40001",
		"RABBITSTREAM_TEST_USER":              "rabbitstream-aaaaaaaaaaaaaaaa",
		"RABBITSTREAM_TEST_PASSWORD":          strings.Repeat("a", 48),
		"RABBITSTREAM_TEST_RESTART_CONTAINER": "codex-rabbitstream-task-rabbit",
		"RABBITSTREAM_TEST_PROXY_API":         "http://127.0.0.1:40002",
		"RABBITSTREAM_TEST_PROXY_NAME":        "rabbitstream",
	}
	if !reflect.DeepEqual(lease.Environment(), wantEnvironment) {
		t.Fatalf("Environment() = %#v", lease.Environment())
	}
	wantIdentity := identity(rabbitStreamImage) + ";" + identity(toxiproxyImage)
	if lease.Identities()["rabbitstream-standalone"] != wantIdentity {
		t.Fatalf("Identities() = %#v", lease.Identities())
	}
	if requestMethod != http.MethodPost || requestURL != "http://127.0.0.1:40002/proxies" ||
		requestHeaders["Content-Type"] != "application/json" ||
		string(requestBody) != `{"name":"rabbitstream","listen":"0.0.0.0:15552","upstream":"rabbit:5552"}` {
		t.Fatalf("proxy request = %s %s %s %#v", requestMethod, requestURL, requestBody, requestHeaders)
	}
	commands := strings.Join(backend.commands, "\n")
	for _, expected := range []string{
		"docker network create --label golib.task=task codex-rabbitstream-task",
		"-p 127.0.0.1::15552 -p 127.0.0.1::8474",
		"target=/etc/rabbitmq/enabled_plugins,readonly",
	} {
		if !strings.Contains(commands, expected) {
			t.Fatalf("commands missing %q: %s", expected, commands)
		}
	}
	if strings.Contains(commands, "127.0.0.1:15552:") || strings.Contains(commands, "127.0.0.1:8474:") {
		t.Fatalf("commands contain fixed host ports: %s", commands)
	}
	if strings.Contains(commands, "RABBITMQ_DEFAULT_PASS=") || strings.Contains(commands, wantEnvironment["RABBITSTREAM_TEST_PASSWORD"]) {
		t.Fatalf("commands expose fixture credentials: %s", commands)
	}
	credentialEnvironment := map[string]string(nil)
	for index, command := range backend.commands {
		if strings.Contains(command, "--name codex-rabbitstream-task-rabbit") {
			credentialEnvironment = backend.environments[index]
		}
	}
	if credentialEnvironment["RABBITMQ_DEFAULT_PASS"] != wantEnvironment["RABBITSTREAM_TEST_PASSWORD"] {
		t.Fatalf("broker environment is incomplete")
	}
	var configuration []byte
	for path, content := range files.writes {
		if filepath.Base(path) == "rabbitmq.conf" {
			configuration = content
		}
	}
	if !strings.Contains(string(configuration), "stream.advertised_port = 40001") {
		t.Fatalf("standalone configuration = %q", configuration)
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	wantTail := []string{
		"docker rm --force codex-rabbitstream-task-rabbit",
		"docker rm --force codex-rabbitstream-task-proxy",
		"docker network rm codex-rabbitstream-task",
	}
	if got := backend.commands[len(backend.commands)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("cleanup commands = %#v", got)
	}
}

func TestRabbitStreamStandaloneAcceptsExistingProxyAndRejectsControlFailure(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		err    error
		ok     bool
	}{
		{"conflict", http.StatusConflict, nil, true},
		{"request failure", 0, errors.New("control unavailable"), false},
		{"unexpected status", http.StatusOK, nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			backend := &fakeBackend{}
			manager := Manager{
				Process: backend.run, Token: fixedToken, Wait: successfulWait, Attempts: 1,
				Workspace: t.TempDir(), Files: &recordingServiceFiles{},
				Secret: func(size int) (string, error) { return strings.Repeat("a", size*2), nil },
				HTTPRequest: func(context.Context, string, string, []byte, map[string]string) (int, error) {
					return test.status, test.err
				},
			}
			lease, err := manager.Start(context.Background(), []string{"rabbitstream-standalone"})
			if test.ok {
				if err != nil {
					t.Fatalf("Start() error = %v", err)
				}
				if err := lease.Close(context.Background()); err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "configure RabbitMQ Streams proxy") {
				t.Fatalf("Start() error = %v", err)
			}
			if len(backend.removed()) != 2 || !strings.Contains(strings.Join(backend.commands, "\n"), "docker network rm") {
				t.Fatalf("cleanup commands = %#v", backend.commands)
			}
		})
	}
}

func TestConfigureToxiproxyRetriesAndHonorsReadinessPolicy(t *testing.T) {
	attempts := 0
	manager := Manager{Attempts: 2, Wait: successfulWait}
	err := manager.configureToxiproxy(context.Background(), func(context.Context, string, string, []byte, map[string]string) (int, error) {
		attempts++
		if attempts == 1 {
			return 0, errors.New("not listening")
		}
		return http.StatusCreated, nil
	}, "http://fixture/proxies", nil)
	if err != nil || attempts != 2 {
		t.Fatalf("configureToxiproxy() = %v after %d attempts", err, attempts)
	}
	if err := (Manager{Attempts: -1}).configureToxiproxy(context.Background(), nil, "", nil); err == nil || !strings.Contains(err.Error(), "attempts must be positive") {
		t.Fatalf("configureToxiproxy(negative) error = %v", err)
	}
	waitFailure := errors.New("canceled wait")
	err = (Manager{Attempts: 2, Wait: func(context.Context, time.Duration) error { return waitFailure }}).configureToxiproxy(
		context.Background(),
		func(context.Context, string, string, []byte, map[string]string) (int, error) {
			return http.StatusServiceUnavailable, nil
		},
		"http://fixture/proxies", nil,
	)
	if !errors.Is(err, waitFailure) {
		t.Fatalf("configureToxiproxy(wait failure) error = %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	err = (Manager{Attempts: 2}).configureToxiproxy(
		canceled,
		func(context.Context, string, string, []byte, map[string]string) (int, error) {
			return 0, errors.New("not ready")
		},
		"http://fixture/proxies", nil,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("configureToxiproxy(default wait) error = %v", err)
	}
}

func TestRabbitStreamStandaloneUsesSecureRuntimeDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/proxies" {
			t.Errorf("request path = %q", request.URL.Path)
		}
		response.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	port := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	backend := &fakeBackend{portOutput: "127.0.0.1:" + port + "\n"}
	manager := Manager{Process: backend.run, Token: fixedToken, Wait: successfulWait, Workspace: t.TempDir()}
	lease, err := manager.Start(context.Background(), []string{"rabbitstream-standalone"})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	environment := lease.Environment()
	if !strings.HasPrefix(environment["RABBITSTREAM_TEST_USER"], "rabbitstream-") ||
		!hexSecret(strings.TrimPrefix(environment["RABBITSTREAM_TEST_USER"], "rabbitstream-")) ||
		!hexSecret(environment["RABBITSTREAM_TEST_PASSWORD"]) {
		t.Fatalf("generated credential shape is invalid")
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRabbitStreamStandaloneRejectsSecretFailures(t *testing.T) {
	if _, err := (Manager{Process: (&fakeBackend{}).run, Token: fixedToken}).Start(context.Background(), []string{"rabbitstream-standalone"}); err == nil || !strings.Contains(err.Error(), "absolute task workspace") {
		t.Fatalf("Start(missing workspace) error = %v", err)
	}
	failure := errors.New("entropy unavailable")
	for _, test := range []struct {
		name   string
		failAt int
		value  string
		want   string
	}{
		{"username", 1, "", "username"},
		{"password", 2, "", "password"},
		{"cookie", 3, "", "cookie"},
		{"malformed", 0, "not-hex", "malformed"},
		{"wrong length", 0, "aa", "malformed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			manager := Manager{Process: (&fakeBackend{}).run, Token: fixedToken, Workspace: t.TempDir(), Secret: func(size int) (string, error) {
				calls++
				if calls == test.failAt {
					return "", failure
				}
				if test.value != "" {
					return test.value, nil
				}
				return strings.Repeat("a", size*2), nil
			}}
			if _, err := manager.Start(context.Background(), []string{"rabbitstream-standalone"}); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Start() error = %v", err)
			}
		})
	}
}

func TestRabbitStreamStandaloneCleansEveryPartialStart(t *testing.T) {
	for _, test := range []struct {
		name       string
		operation  string
		occurrence int
		want       string
	}{
		{"network", "network", 1, "create RabbitMQ Streams network"},
		{"proxy", "run", 1, "start RabbitMQ Streams proxy"},
		{"proxy identity", "inspect", 1, "proxy image identity"},
		{"proxy port", "port", 1, "proxy port"},
		{"control port", "port", 2, "control port"},
		{"broker", "run", 2, "start RabbitMQ Streams broker"},
		{"broker identity", "inspect", 2, "broker image identity"},
		{"readiness", "exec", 1, "did not become ready"},
	} {
		t.Run(test.name, func(t *testing.T) {
			backend := &occurrenceBackend{failureOperation: test.operation, failureOccurrence: test.occurrence}
			manager := Manager{
				Process: backend.run, Token: fixedToken, Attempts: 1,
				Workspace: t.TempDir(), Files: &recordingServiceFiles{},
				Secret: func(size int) (string, error) { return strings.Repeat("a", size*2), nil },
				HTTPRequest: func(context.Context, string, string, []byte, map[string]string) (int, error) {
					return http.StatusCreated, nil
				},
			}
			if _, err := manager.Start(context.Background(), []string{"rabbitstream-standalone"}); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Start() error = %v", err)
			}
			if test.name != "network" && !strings.Contains(strings.Join(backend.commands, "\n"), "docker network rm codex-rabbitstream-task") {
				t.Fatalf("cleanup commands = %#v", backend.commands)
			}
		})
	}
}

func TestRabbitStreamRuntimeHelpers(t *testing.T) {
	secret, err := randomHex(8)
	if err != nil || len(secret) != 16 || !hexSecret(secret) {
		t.Fatalf("randomHex() = %q, %v", secret, err)
	}
	for _, size := range []int{0, 65} {
		if _, err := randomHexFrom(strings.NewReader(""), size); err == nil {
			t.Fatalf("randomHexFrom(%d) error = nil", size)
		}
	}
	if _, err := randomHexFrom(failingReader{}, 1); err == nil {
		t.Fatal("randomHexFrom(read failure) error = nil")
	}
	for _, invalid := range []string{"", "a", "zz"} {
		if hexSecret(invalid) {
			t.Fatalf("hexSecret(%q) = true", invalid)
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/ok":
			if request.Method != http.MethodPost || request.Header.Get("X-Test") != "present" {
				t.Errorf("request = %s %#v", request.Method, request.Header)
			}
			body, readErr := io.ReadAll(request.Body)
			if readErr != nil || string(body) != "body" {
				t.Errorf("body = %q, %v", body, readErr)
			}
			response.WriteHeader(http.StatusCreated)
		case "/redirect":
			http.Redirect(response, request, "/ok", http.StatusFound)
		case "/large":
			_, _ = io.WriteString(response, strings.Repeat("x", maximumHTTPBody+1))
		case "/truncated":
			response.Header().Set("Content-Length", "10")
			_, _ = io.WriteString(response, "x")
		}
	}))
	defer server.Close()
	status, err := httpRequest(context.Background(), http.MethodPost, server.URL+"/ok", []byte("body"), map[string]string{"X-Test": "present"})
	if err != nil || status != http.StatusCreated {
		t.Fatalf("httpRequest(ok) = %d, %v", status, err)
	}
	for _, path := range []string{"/redirect", "/large", "/truncated"} {
		if _, err := httpRequest(context.Background(), http.MethodGet, server.URL+path, nil, nil); err == nil {
			t.Fatalf("httpRequest(%s) error = nil", path)
		}
	}
	if _, err := httpRequest(context.Background(), http.MethodGet, "://invalid", nil, nil); err == nil {
		t.Fatal("httpRequest(invalid URL) error = nil")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := httpRequest(canceled, http.MethodGet, server.URL+"/ok", nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("httpRequest(canceled) error = %v", err)
	}
}

func TestLeaseCloseReportsNetworkCleanupFailure(t *testing.T) {
	failure := errors.New("network cleanup failed")
	lease := &Lease{
		networks: []string{"codex-rabbitstream-task"},
		process:  func(context.Context, string, []string, map[string]string, io.Writer, io.Writer) error { return failure },
	}
	if err := lease.Close(context.Background()); err == nil || !strings.Contains(err.Error(), "remove service network") {
		t.Fatalf("Close() error = %v", err)
	}
}

type occurrenceBackend struct {
	commands          []string
	counts            map[string]int
	failureOperation  string
	failureOccurrence int
	port              int
}

func (backend *occurrenceBackend) run(_ context.Context, name string, args []string, _ map[string]string, stdout, _ io.Writer) error {
	command := name + " " + strings.Join(args, " ")
	backend.commands = append(backend.commands, command)
	operation := ""
	if len(args) > 0 {
		operation = args[0]
	}
	if backend.counts == nil {
		backend.counts = make(map[string]int)
	}
	backend.counts[operation]++
	if operation == backend.failureOperation && backend.counts[operation] == backend.failureOccurrence {
		return errors.New("injected failure")
	}
	if operation == "inspect" {
		_, _ = io.WriteString(stdout, "sha256:"+strings.Repeat("a", 64))
	}
	if operation == "port" {
		backend.port++
		_, _ = io.WriteString(stdout, "127.0.0.1:"+strconv.Itoa(40000+backend.port)+"\n")
	}
	return nil
}
