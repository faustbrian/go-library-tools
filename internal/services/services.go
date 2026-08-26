// Package services owns isolated lifecycle management for generic test
// services used by library verification.
package services

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maximumPortOutput = 4_096
	defaultAttempts   = 90
	defaultWait       = time.Second
	cleanupTimeout    = time.Duration(30_000_000_000)
	postgresImage     = "postgres:18.4-alpine"
	valkeyImage       = "valkey/valkey:9.1.0-alpine"
	redisImage        = "redis:8.6.4-alpine"
	natsImage         = "nats:2.14.2-alpine"
	nsqImage          = "nsqio/nsq:v1.3.0"
	rabbitMQImage     = "rabbitmq:4.3.2-management-alpine"
)

var tokenRE = regexp.MustCompile(`^[a-z0-9]{4,32}$`)

// Process executes one external command without shell interpretation.
type Process func(context.Context, string, []string, map[string]string, io.Writer, io.Writer) error

// Probe verifies a host TCP endpoint.
type Probe func(context.Context, string, string) error

// Wait pauses between bounded readiness attempts.
type Wait func(context.Context, time.Duration) error

// Token creates a collision-resistant task identifier.
type Token func() (string, error)

// HTTPProbe verifies one HTTP readiness endpoint.
type HTTPProbe func(context.Context, string) error

// HTTPRequest performs one bounded fixture-control request.
type HTTPRequest func(context.Context, string, string, []byte, map[string]string) (int, error)

// Secret returns a cryptographically random hexadecimal fixture secret.
type Secret func(int) (string, error)

// FileSystem owns task-local fixture files.
type FileSystem interface {
	MkdirAll(string, os.FileMode) error
	WriteFile(string, []byte, os.FileMode) error
	Chmod(string, os.FileMode) error
}

// Manager starts one isolated set of generic service fixtures.
type Manager struct {
	Process     Process
	Probe       Probe
	Wait        Wait
	Token       Token
	HTTPProbe   HTTPProbe
	HTTPRequest HTTPRequest
	Secret      Secret
	Files       FileSystem
	Workspace   string
	Attempts    int
	OpenSearch  *OpenSearchImages
}

// Lease exposes fixture environment and owns exact started resources.
type Lease struct {
	environment map[string]string
	identities  map[string]string
	containers  []string
	networks    []string
	volumes     []string
	process     Process
	closeOnce   sync.Once
	closeErr    error
}

type definition struct {
	name          string
	image         string
	port          int
	runArguments  func(string) []string
	readyCommand  []string
	environment   func(string) map[string]string
	requiresProbe bool
	requiresHTTP  bool
	start         func(context.Context, Manager, *Lease, string) error
}

var catalog = map[string]definition{
	"postgresql": {
		name: "postgresql", image: postgresImage, port: 5432,
		runArguments: func(string) []string {
			return []string{"-e", "POSTGRES_USER=golib", "-e", "POSTGRES_PASSWORD=golib", "-e", "POSTGRES_DB=golib"}
		},
		readyCommand: []string{"pg_isready", "-U", "golib", "-d", "golib"},
		environment:  postgresEnvironment,
	},
	"valkey": {
		name: "valkey", image: valkeyImage, port: 6379,
		readyCommand: []string{"valkey-cli", "ping"}, environment: valkeyEnvironment,
	},
	"redis": {
		name: "redis", image: redisImage, port: 6379,
		readyCommand: []string{"redis-cli", "ping"}, environment: redisEnvironment,
	},
	"nats": {
		name: "nats", image: natsImage, port: 4222,
		requiresProbe: true, environment: func(port string) map[string]string {
			return map[string]string{"NATS_URL": "nats://127.0.0.1:" + port}
		},
	},
	"nsq": {
		name: "nsq", image: nsqImage, port: 4150,
		requiresProbe: true, environment: func(port string) map[string]string {
			return map[string]string{"NSQD_TCP_ADDRESS": "127.0.0.1:" + port}
		},
	},
	"rabbitmq": {
		name: "rabbitmq", image: rabbitMQImage, port: 5672,
		runArguments: func(name string) []string { return []string{"--hostname", name, "--user", "rabbitmq"} },
		readyCommand: []string{"rabbitmq-diagnostics", "-q", "ping"}, environment: func(port string) map[string]string {
			return map[string]string{"RABBITMQ_URL": "amqp://guest:guest@127.0.0.1:" + port + "/"}
		},
	},
	"rabbitstream-standalone": {
		name: "rabbitstream-standalone", start: startRabbitStreamStandalone,
	},
	"rabbitstream": {
		name: "rabbitstream", start: startRabbitStream,
	},
}

// Start validates the complete selection before creating task-owned containers.
func (manager Manager) Start(ctx context.Context, services []string) (*Lease, error) {
	if manager.Process == nil {
		return nil, errors.New("service process backend is required")
	}
	definitions, err := manager.selectedDefinitions(services)
	if err != nil {
		return nil, err
	}
	token := manager.Token
	if token == nil {
		token = randomToken
	}
	identifier, err := token()
	if err != nil {
		return nil, fmt.Errorf("create service token: %w", err)
	}
	if !tokenRE.MatchString(identifier) {
		return nil, errors.New("service token is malformed")
	}
	lease := &Lease{
		environment: make(map[string]string), identities: make(map[string]string),
		process: manager.Process,
	}
	for _, service := range definitions {
		start := manager.start
		if service.start != nil {
			start = func(ctx context.Context, lease *Lease, _ definition, token string) error {
				return service.start(ctx, manager, lease, token)
			}
		}
		if err := start(ctx, lease, service, identifier); err != nil {
			cleanupContext, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
			cleanupErr := lease.Close(cleanupContext)
			cancel()
			return nil, errors.Join(err, cleanupErr)
		}
	}
	return lease, nil
}

func (manager Manager) selectedDefinitions(names []string) ([]definition, error) {
	if len(names) == 0 {
		return nil, errors.New("service selection is empty")
	}
	seen := make(map[string]struct{}, len(names))
	result := make([]definition, 0, len(names))
	for _, name := range names {
		service, exists := catalog[name]
		if name == "opensearch" && manager.OpenSearch != nil {
			custom, customErr := manager.OpenSearch.definition()
			if customErr != nil {
				return nil, customErr
			}
			service = custom
			exists = true
		}
		if !exists {
			return nil, fmt.Errorf("unsupported service %q", name)
		}
		if _, duplicate := seen[name]; duplicate {
			return nil, fmt.Errorf("duplicate service %q", name)
		}
		seen[name] = struct{}{}
		result = append(result, service)
	}
	slices.SortFunc(result, func(left, right definition) int { return strings.Compare(left.name, right.name) })
	return result, nil
}

func (manager Manager) start(ctx context.Context, lease *Lease, service definition, token string) error {
	name := "golib-" + service.name + "-" + token
	arguments := []string{"run", "--detach", "--name", name, "-p", "127.0.0.1::" + strconv.Itoa(service.port)}
	if service.runArguments != nil {
		arguments = append(arguments, service.runArguments(name)...)
	}
	arguments = append(arguments, service.image)
	if service.name == "nsq" {
		arguments = append(arguments, "/nsqd", "--broadcast-address=127.0.0.1")
	}
	if err := manager.Process(ctx, "docker", arguments, nil, io.Discard, io.Discard); err != nil {
		return fmt.Errorf("start %s: %w", service.name, err)
	}
	lease.containers = append(lease.containers, name)
	imageIdentity, err := manager.imageIdentity(ctx, name, service.image)
	if err != nil {
		return fmt.Errorf("read %s image identity: %w", service.name, err)
	}
	var portOutput boundedWriter
	if err := manager.Process(ctx, "docker", []string{"port", name, strconv.Itoa(service.port) + "/tcp"}, nil, &portOutput, io.Discard); err != nil {
		return fmt.Errorf("read %s port: %w", service.name, err)
	}
	port, err := publishedPort(strings.NewReader(portOutput.String()))
	if err != nil {
		return fmt.Errorf("parse %s port: %w", service.name, err)
	}
	if err := manager.waitReady(ctx, service, name, port); err != nil {
		return err
	}
	for key, value := range service.environment(port) {
		lease.environment[key] = value
	}
	lease.identities[service.name] = imageIdentity
	return nil
}

func (manager Manager) imageIdentity(ctx context.Context, container, reference string) (string, error) {
	var output boundedWriter
	if err := manager.Process(ctx, "docker", []string{"inspect", "--format={{.Image}}", container}, nil, &output, io.Discard); err != nil {
		return "", err
	}
	digest := strings.TrimSpace(output.String())
	if len(digest) != 71 || !strings.HasPrefix(digest, "sha256:") {
		return "", errors.New("container image identity is malformed")
	}
	for _, character := range digest[len("sha256:"):] {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return "", errors.New("container image identity is malformed")
		}
	}
	return reference + "#" + digest, nil
}

func (manager Manager) waitReady(ctx context.Context, service definition, name, port string) error {
	attempts := manager.Attempts
	if attempts == 0 {
		attempts = defaultAttempts
	}
	if attempts < 1 {
		return errors.New("service readiness attempts must be positive")
	}
	probe := manager.Probe
	if probe == nil {
		probe = tcpProbe
	}
	wait := manager.Wait
	if wait == nil {
		wait = waitContext
	}
	var last error
	httpReady := manager.HTTPProbe
	if httpReady == nil {
		httpReady = httpProbe
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if service.requiresHTTP {
			last = httpReady(ctx, "http://127.0.0.1:"+port+"/")
		} else if service.requiresProbe {
			last = probe(ctx, "tcp", net.JoinHostPort("127.0.0.1", port))
		} else {
			arguments := append([]string{"exec", name}, service.readyCommand...)
			last = manager.Process(ctx, "docker", arguments, nil, io.Discard, io.Discard)
		}
		if last == nil {
			return nil
		}
		if shouldRetry(attempt, attempts) {
			if err := wait(ctx, defaultWait); err != nil {
				return fmt.Errorf("wait for %s: %w", service.name, err)
			}
		}
	}
	return fmt.Errorf("service %s did not become ready: %w", service.name, last)
}

// Environment returns an independent copy of the fixture environment.
func (lease *Lease) Environment() map[string]string { return clone(lease.environment) }

// Identities returns immutable image identities for evidence binding.
func (lease *Lease) Identities() map[string]string { return clone(lease.identities) }

// Close removes only containers started by this lease, in reverse order.
func (lease *Lease) Close(ctx context.Context) error {
	lease.closeOnce.Do(func() {
		failures := make([]error, 0)
		for index := len(lease.containers) - 1; index >= 0; index-- {
			name := lease.containers[index]
			if err := lease.process(ctx, "docker", []string{"rm", "--force", name}, nil, io.Discard, io.Discard); err != nil {
				failures = append(failures, fmt.Errorf("remove service container %s: %w", name, err))
			}
		}
		for index := len(lease.volumes) - 1; index >= 0; index-- {
			name := lease.volumes[index]
			if err := lease.process(ctx, "docker", []string{"volume", "rm", "--force", name}, nil, io.Discard, io.Discard); err != nil {
				failures = append(failures, fmt.Errorf("remove service volume %s: %w", name, err))
			}
		}
		for index := len(lease.networks) - 1; index >= 0; index-- {
			name := lease.networks[index]
			if err := lease.process(ctx, "docker", []string{"network", "rm", name}, nil, io.Discard, io.Discard); err != nil {
				failures = append(failures, fmt.Errorf("remove service network %s: %w", name, err))
			}
		}
		lease.closeErr = errors.Join(failures...)
	})
	return lease.closeErr
}

type operatingFiles struct{}

func (operatingFiles) MkdirAll(path string, mode os.FileMode) error { return os.MkdirAll(path, mode) }
func (operatingFiles) WriteFile(path string, data []byte, mode os.FileMode) error {
	return os.WriteFile(path, data, mode)
}
func (operatingFiles) Chmod(path string, mode os.FileMode) error { return os.Chmod(path, mode) }

func postgresEnvironment(port string) map[string]string {
	dsn := "postgres://golib:golib@127.0.0.1:" + port + "/golib?sslmode=disable"
	return map[string]string{
		"POSTGRES_URL": dsn, "POSTGRES_VERSION": "18.4", "OUTBOX_POSTGRES_VERSION": "18.4",
		"STATE_MACHINE_POSTGRES_VERSION": "18.4", "FEATURE_FLAGS_POSTGRES_DSN": dsn,
		"TEST_DATABASE_URL": dsn, "DATABASE_URL": dsn, "TEMPORAL_POSTGRES_DSN": dsn,
	}
}

func valkeyEnvironment(port string) map[string]string {
	address := "127.0.0.1:" + port
	return map[string]string{
		"VALKEY_ADDR": address, "VALKEY_ADDRESS": address, "FEATURE_FLAGS_VALKEY_ADDRESS": address,
		"TEST_VALKEY_ADDRESS": address, "CACHE_VALKEY_IMAGE": valkeyImage,
	}
}

func redisEnvironment(port string) map[string]string {
	address := "127.0.0.1:" + port
	return map[string]string{
		"REDIS_ADDR": address, "TEST_REDIS_ADDRESS": address, "CACHE_REDIS_IMAGE": redisImage,
	}
}

func publishedPort(reader io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumPortOutput+1))
	if err != nil {
		return "", fmt.Errorf("read published port: %w", err)
	}
	if len(data) > maximumPortOutput {
		return "", errors.New("published port output exceeds limit")
	}
	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(string(data))))
	if !scanner.Scan() {
		return "", errors.New("published port output is empty")
	}
	line := scanner.Text()
	if scanner.Scan() {
		return "", errors.New("published port output is ambiguous")
	}
	host, port, err := net.SplitHostPort(line)
	if err != nil {
		return "", errors.New("published port is not bound to 127.0.0.1")
	}
	if host != "127.0.0.1" {
		return "", errors.New("published port is not bound to 127.0.0.1")
	}
	number, err := strconv.Atoi(port)
	if err != nil {
		return "", errors.New("published port is invalid")
	}
	if number < 1 {
		return "", errors.New("published port is invalid")
	}
	if number > 65535 {
		return "", errors.New("published port is invalid")
	}
	return port, nil
}

func randomToken() (string, error) {
	return randomTokenFrom(rand.Reader)
}

func randomTokenFrom(reader io.Reader) (string, error) {
	value := make([]byte, 8)
	if _, err := io.ReadFull(reader, value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func tcpProbe(ctx context.Context, network, address string) error {
	connection, err := (&net.Dialer{Timeout: time.Second}).DialContext(ctx, network, address)
	if err != nil {
		return err
	}
	return connection.Close()
}

func waitContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func clone(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

type boundedWriter struct {
	value strings.Builder
}

func (writer *boundedWriter) Write(value []byte) (int, error) {
	if writer.value.Len()+len(value) > maximumPortOutput {
		return 0, errors.New("service output exceeds limit")
	}
	return writer.value.Write(value)
}

func (writer *boundedWriter) String() string { return writer.value.String() }
