package services

import (
	"bytes"
	"cmp"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const (
	rabbitStreamImage  = "rabbitmq@sha256:397fde82bc04522d88680b57cbf5d70caae715a76c957404e52e3f0fa056b8f3"
	toxiproxyImage     = "ghcr.io/shopify/toxiproxy@sha256:9378ed52a28bc50edc1350f936f518f31fa95f0d15917d6eb40b8e376d1a214e"
	maximumHTTPBody    = 65_536
	httpRequestTimeout = time.Duration(5_000_000_000)
)

func startRabbitStreamStandalone(ctx context.Context, manager Manager, lease *Lease, token string) error {
	workspaceValid := [2]bool{manager.Workspace != "", filepath.IsAbs(manager.Workspace)}
	if workspaceValid != [2]bool{true, true} {
		return errors.New("RabbitMQ Streams requires an absolute task workspace")
	}
	secret := secretGenerator(manager.Secret)
	userSecret, err := secret(8)
	if err != nil {
		return fmt.Errorf("create RabbitMQ Streams username: %w", err)
	}
	password, err := secret(24)
	if err != nil {
		return fmt.Errorf("create RabbitMQ Streams password: %w", err)
	}
	cookie, err := secret(32)
	if err != nil {
		return fmt.Errorf("create RabbitMQ Streams cookie: %w", err)
	}
	for _, credential := range []struct {
		value string
		bytes int
	}{{userSecret, 8}, {password, 24}, {cookie, 32}} {
		valid := [2]bool{len(credential.value) == credential.bytes*2, hexSecret(credential.value)}
		if valid != [2]bool{true, true} {
			return errors.New("RabbitMQ Streams secret generator returned malformed data")
		}
	}

	network := "codex-rabbitstream-" + token
	if err := manager.Process(ctx, "docker", []string{"network", "create", "--label", "golib.task=" + token, network}, nil, io.Discard, io.Discard); err != nil {
		return fmt.Errorf("create RabbitMQ Streams network: %w", err)
	}
	lease.networks = append(lease.networks, network)
	proxy := network + "-proxy"
	if err := manager.Process(ctx, "docker", []string{
		"run", "--detach", "--name", proxy, "--network", network, "--network-alias", "proxy",
		"-p", "127.0.0.1::15552", "-p", "127.0.0.1::8474", toxiproxyImage, "-host=0.0.0.0",
	}, nil, io.Discard, io.Discard); err != nil {
		return fmt.Errorf("start RabbitMQ Streams proxy: %w", err)
	}
	lease.containers = append(lease.containers, proxy)
	proxyIdentity, err := manager.imageIdentity(ctx, proxy, toxiproxyImage)
	if err != nil {
		return fmt.Errorf("read RabbitMQ Streams proxy image identity: %w", err)
	}
	proxyPort, err := manager.containerPort(ctx, proxy, 15552)
	if err != nil {
		return fmt.Errorf("read RabbitMQ Streams proxy port: %w", err)
	}
	controlPort, err := manager.containerPort(ctx, proxy, 8474)
	if err != nil {
		return fmt.Errorf("read RabbitMQ Streams control port: %w", err)
	}

	username := "rabbitstream-" + userSecret
	rabbit := network + "-rabbit"
	files := manager.Files
	if files == nil {
		files = operatingFiles{}
	}
	directory := filepath.Join(manager.Workspace, "rabbitstream-standalone-"+token)
	if err := files.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create RabbitMQ Streams standalone workspace: %w", err)
	}
	enabledPlugins := filepath.Join(directory, "enabled_plugins")
	configuration := filepath.Join(directory, "rabbitmq.conf")
	if err := files.WriteFile(enabledPlugins, []byte("[rabbitmq_management,rabbitmq_stream].\n"), 0o644); err != nil {
		return fmt.Errorf("write RabbitMQ Streams standalone plugins: %w", err)
	}
	config := "stream.listeners.tcp.1 = 5552\nstream.advertised_host = localhost\nstream.advertised_port = " + proxyPort + "\n"
	if err := files.WriteFile(configuration, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write RabbitMQ Streams standalone configuration: %w", err)
	}
	rabbitEnvironment := map[string]string{
		"RABBITMQ_DEFAULT_USER":  username,
		"RABBITMQ_DEFAULT_PASS":  password,
		"RABBITMQ_ERLANG_COOKIE": cookie,
	}
	if err := manager.Process(ctx, "docker", []string{
		"run", "--detach", "--name", rabbit, "--hostname", "rabbit", "--network", network, "--network-alias", "rabbit",
		"-e", "RABBITMQ_DEFAULT_USER", "-e", "RABBITMQ_DEFAULT_PASS", "-e", "RABBITMQ_ERLANG_COOKIE",
		"--mount", "type=bind,source=" + enabledPlugins + ",target=/etc/rabbitmq/enabled_plugins,readonly",
		"--mount", "type=bind,source=" + configuration + ",target=/etc/rabbitmq/rabbitmq.conf,readonly",
		rabbitStreamImage,
	}, rabbitEnvironment, io.Discard, io.Discard); err != nil {
		return fmt.Errorf("start RabbitMQ Streams broker: %w", err)
	}
	lease.containers = append(lease.containers, rabbit)
	rabbitIdentity, err := manager.imageIdentity(ctx, rabbit, rabbitStreamImage)
	if err != nil {
		return fmt.Errorf("read RabbitMQ Streams broker image identity: %w", err)
	}
	readiness := definition{name: "rabbitstream-standalone", readyCommand: []string{"rabbitmq-diagnostics", "-q", "check_running"}}
	if err := manager.waitReady(ctx, readiness, rabbit, ""); err != nil {
		return err
	}

	controlURL := "http://127.0.0.1:" + controlPort
	request := manager.HTTPRequest
	if request == nil {
		request = httpRequest
	}
	payload := []byte(`{"name":"rabbitstream","listen":"0.0.0.0:15552","upstream":"rabbit:5552"}`)
	if err := manager.configureToxiproxy(ctx, request, controlURL+"/proxies", payload); err != nil {
		return err
	}

	lease.environment["RABBITSTREAM_TEST_HOST"] = "localhost"
	lease.environment["RABBITSTREAM_TEST_PORT"] = proxyPort
	lease.environment["RABBITSTREAM_TEST_USER"] = username
	lease.environment["RABBITSTREAM_TEST_PASSWORD"] = password
	lease.environment["RABBITSTREAM_TEST_RESTART_CONTAINER"] = rabbit
	lease.environment["RABBITSTREAM_TEST_PROXY_API"] = controlURL
	lease.environment["RABBITSTREAM_TEST_PROXY_NAME"] = "rabbitstream"
	lease.identities["rabbitstream-standalone"] = rabbitIdentity + ";" + proxyIdentity
	return nil
}

func (manager Manager) configureToxiproxy(ctx context.Context, request HTTPRequest, url string, payload []byte) error {
	attempts := manager.Attempts
	if attempts == 0 {
		attempts = defaultAttempts
	}
	if attempts < 1 {
		return errors.New("service readiness attempts must be positive")
	}
	wait := manager.Wait
	if wait == nil {
		wait = waitContext
	}
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		status, err := request(ctx, http.MethodPost, url, payload, map[string]string{"Content-Type": "application/json"})
		accepted := err == nil && (status == http.StatusCreated || status == http.StatusConflict)
		if accepted {
			return nil
		}
		last = proxyReadinessError(status, err)
		if shouldRetry(attempt, attempts) {
			if err := wait(ctx, defaultWait); err != nil {
				return fmt.Errorf("wait for RabbitMQ Streams proxy: %w", err)
			}
		}
	}
	return fmt.Errorf("configure RabbitMQ Streams proxy: %w", last)
}

func secretGenerator(secret Secret) Secret {
	switch secret {
	case nil:
		return randomHex
	default:
		return secret
	}
}

func proxyReadinessError(status int, err error) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("unexpected HTTP %d", status)
}

func shouldRetry(attempt, attempts int) bool { return attempt+1 < attempts }

func (manager Manager) containerPort(ctx context.Context, container string, port int) (string, error) {
	var output boundedWriter
	if err := manager.Process(ctx, "docker", []string{"port", container, fmt.Sprintf("%d/tcp", port)}, nil, &output, io.Discard); err != nil {
		return "", err
	}
	return publishedPort(strings.NewReader(output.String()))
}

func randomHex(size int) (string, error) {
	return randomHexFrom(rand.Reader, size)
}

func randomHexFrom(reader io.Reader, size int) (string, error) {
	if size < 1 {
		return "", errors.New("secret size must be between 1 and 64 bytes")
	}
	if size > 64 {
		return "", errors.New("secret size must be between 1 and 64 bytes")
	}
	value := make([]byte, size)
	if _, err := io.ReadFull(reader, value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func hexSecret(value string) bool {
	if len(value) < 2 {
		return false
	}
	if len(value)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func httpRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) (int, error) {
	return httpRequestWithFactory(ctx, method, url, body, headers, http.NewRequestWithContext)
}

func httpRequestWithFactory(ctx context.Context, method, url string, body []byte, headers map[string]string, create func(context.Context, string, string, io.Reader) (*http.Request, error)) (int, error) {
	request, err := create(ctx, method, url, bytes.NewReader(body))
	switch request {
	case nil:
		return 0, errors.Join(errors.New("create fixture-control request"), err)
	}
	switch err {
	case nil:
	default:
		return 0, err
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	client := &http.Client{
		Transport:     &http.Transport{DisableKeepAlives: true},
		Timeout:       httpRequestTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirects are not allowed") },
	}
	response, err := client.Do(request)
	switch response {
	case nil:
		return 0, err
	}
	switch err {
	case nil:
	default:
		return 0, err
	}
	defer response.Body.Close()
	drainErr := drainHTTPBody(response.Body)
	switch drainErr {
	case nil:
	default:
		return 0, drainErr
	}
	return response.StatusCode, nil
}

func drainHTTPBody(reader io.Reader) error {
	written, err := io.Copy(io.Discard, io.LimitReader(reader, maximumHTTPBody+1))
	switch err {
	case nil:
	default:
		return err
	}
	switch cmp.Compare(written, int64(maximumHTTPBody)) {
	case 1:
		return errors.New("fixture-control response exceeds limit")
	default:
		return nil
	}
}
