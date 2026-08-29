package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const rabbitStreamOldImage = "rabbitmq@sha256:eb5295d083325da5929a5ade766684d4019ffd2bce8bc7e43d6f9a05cafc8646"

type rabbitStreamTopology struct {
	base       string
	project    string
	network    string
	proxy      string
	directory  string
	compose    string
	runtime    string
	ports      map[int]string
	containers map[string]string
	volumes    map[string]string
}

func startRabbitStream(ctx context.Context, manager Manager, lease *Lease, token string) error {
	workspaceValid := [2]bool{manager.Workspace != "", filepath.IsAbs(manager.Workspace)}
	if workspaceValid != [2]bool{true, true} {
		return errors.New("RabbitMQ Streams requires an absolute task workspace")
	}
	if err := startRabbitStreamStandalone(ctx, manager, lease, token+"single"); err != nil {
		return err
	}
	standaloneIdentity := lease.identities["rabbitstream-standalone"]
	delete(lease.identities, "rabbitstream-standalone")

	secret := manager.Secret
	if secret == nil {
		secret = randomHex
	}
	credentials, err := newRabbitStreamCredentials(
		secret,
		lease.environment["RABBITSTREAM_TEST_USER"],
		lease.environment["RABBITSTREAM_TEST_PASSWORD"],
	)
	if err != nil {
		return err
	}
	topology := newRabbitStreamTopology(manager.Workspace, token)
	files := manager.Files
	if files == nil {
		files = operatingFiles{}
	}
	if err := files.MkdirAll(topology.directory, 0o700); err != nil {
		return fmt.Errorf("create RabbitMQ Streams workspace: %w", err)
	}
	if err := files.MkdirAll(topology.runtime, 0o700); err != nil {
		return fmt.Errorf("create RabbitMQ Streams TLS runtime: %w", err)
	}
	if err := manager.createRabbitStreamResources(ctx, lease, topology); err != nil {
		return err
	}
	for _, port := range []int{15561, 15562, 15563, 15571, 8474} {
		published, err := manager.containerPort(ctx, topology.proxy, port)
		if err != nil {
			return fmt.Errorf("read RabbitMQ Streams topology port %d: %w", port, err)
		}
		topology.ports[port] = published
	}
	if err := writeRabbitStreamFixtureFiles(files, topology); err != nil {
		return err
	}
	compose := renderRabbitStreamCompose()
	if err := files.WriteFile(topology.compose, []byte(compose), 0o600); err != nil {
		return fmt.Errorf("write RabbitMQ Streams topology: %w", err)
	}
	composeEnvironment := topology.environment(credentials)
	for _, name := range []string{"rabbit1", "rabbit2", "rabbit3", "certgen", "rabbit-tls"} {
		lease.containers = append(lease.containers, topology.containers[name])
	}
	diagnostic := newBoundedDiagnosticWriter(
		credentials.user,
		credentials.password,
		credentials.cookie,
		credentials.restrictedUser,
		credentials.restrictedPassword,
	)
	if err := manager.Process(ctx, "docker", []string{
		"compose", "--project-directory", topology.directory, "-f", topology.compose,
		"-p", topology.project, "up", "-d", "--wait",
	}, composeEnvironment, io.Discard, &diagnostic); err != nil {
		manager.appendRabbitStreamContainerLogs(ctx, topology, &diagnostic)
		return fmt.Errorf("start RabbitMQ Streams topology: %w", serviceProcessError(err, &diagnostic))
	}
	if err := manager.Process(ctx, "docker", []string{
		"exec", topology.containers["rabbit1"], "rabbitmqctl", "await_online_nodes", "3",
	}, nil, io.Discard, io.Discard); err != nil {
		return fmt.Errorf("form RabbitMQ Streams cluster: %w", err)
	}
	if err := manager.configureRabbitStreamProxies(ctx, topology); err != nil {
		return err
	}
	if err := manager.configureRabbitStreamTLS(ctx, topology, credentials, files); err != nil {
		return err
	}
	identities, err := manager.rabbitStreamIdentities(ctx, topology)
	if err != nil {
		return err
	}
	lease.identities["rabbitstream"] = standaloneIdentity + ";" + strings.Join(identities, ";")
	maps.Copy(lease.environment, topology.testEnvironment(credentials))
	return nil
}

func (manager Manager) appendRabbitStreamContainerLogs(
	ctx context.Context,
	topology rabbitStreamTopology,
	diagnostic *boundedDiagnosticWriter,
) {
	for _, name := range rabbitStreamDiagnosticContainers(topology, string(diagnostic.value)) {
		stdout := newBoundedDiagnosticWriter(diagnostic.secrets...)
		stderr := newBoundedDiagnosticWriter(diagnostic.secrets...)
		_ = manager.Process(ctx, "docker", []string{
			"logs", "--tail", "120", topology.containers[name],
		}, nil, &stdout, &stderr)
		output := strings.TrimSpace(strings.Join([]string{stdout.String(), stderr.String()}, "\n"))
		if output != "" {
			_, _ = fmt.Fprintf(diagnostic, "\n[%s logs]\n%s", name, output)
		}
	}
}

func rabbitStreamDiagnosticContainers(topology rabbitStreamTopology, diagnostic string) []string {
	names := []string{"rabbit1", "rabbit2", "rabbit3", "certgen", "rabbit-tls"}
	for _, name := range names {
		if strings.Contains(diagnostic, topology.containers[name]+" exited") {
			return []string{name}
		}
	}
	return nil
}

type rabbitStreamCredentials struct {
	user               string
	password           string
	cookie             string
	restrictedUser     string
	restrictedPassword string
}

func newRabbitStreamCredentials(secret Secret, user, password string) (rabbitStreamCredentials, error) {
	userSecret := strings.TrimPrefix(user, "rabbitstream-")
	valid := [5]bool{userSecret != user, len(userSecret) == 16, hexSecret(userSecret), len(password) == 48, hexSecret(password)}
	if valid != [5]bool{true, true, true, true, true} {
		return rabbitStreamCredentials{}, errors.New("RabbitMQ Streams shared credentials are malformed")
	}
	sizes := []int{32, 8, 24}
	values := make([]string, len(sizes))
	labels := []string{"cookie", "restricted username", "restricted password"}
	for index, size := range sizes {
		value, err := secret(size)
		if err != nil {
			return rabbitStreamCredentials{}, fmt.Errorf("create RabbitMQ Streams %s: %w", labels[index], err)
		}
		if len(value) != size*2 || !hexSecret(value) {
			return rabbitStreamCredentials{}, errors.New("RabbitMQ Streams secret generator returned malformed data")
		}
		values[index] = value
	}
	return rabbitStreamCredentials{
		user: user, password: password, cookie: values[0],
		restrictedUser: "restricted-" + values[1], restrictedPassword: values[2],
	}, nil
}

func newRabbitStreamTopology(workspace, token string) rabbitStreamTopology {
	base := "codex-rabbitstream-" + token
	project := base + "-cluster"
	directory := filepath.Join(workspace, "rabbitstream-"+token)
	containers := map[string]string{
		"rabbit1": project + "-rabbit1-1", "rabbit2": project + "-rabbit2-1",
		"rabbit3": project + "-rabbit3-1", "certgen": project + "-certgen-1",
		"rabbit-tls": project + "-rabbit-tls-1",
	}
	volumes := map[string]string{
		"rabbit1": base + "-rabbit1-data", "rabbit2": base + "-rabbit2-data",
		"rabbit3": base + "-rabbit3-data", "tls-certs": base + "-tls-certs",
		"tls-data": base + "-tls-data",
	}
	return rabbitStreamTopology{
		base: base, project: project, network: base + "-network", proxy: base + "-proxy",
		directory: directory, compose: filepath.Join(directory, "compose.yaml"),
		runtime: filepath.Join(directory, "tls"), ports: make(map[int]string),
		containers: containers, volumes: volumes,
	}
}

func (manager Manager) createRabbitStreamResources(ctx context.Context, lease *Lease, topology rabbitStreamTopology) error {
	if err := manager.Process(ctx, "docker", []string{
		"network", "create", "--label", "golib.task=" + topology.base, topology.network,
	}, nil, io.Discard, io.Discard); err != nil {
		return fmt.Errorf("create RabbitMQ Streams topology network: %w", err)
	}
	lease.networks = append(lease.networks, topology.network)
	volumeNames := []string{
		topology.volumes["rabbit1"], topology.volumes["rabbit2"], topology.volumes["rabbit3"],
		topology.volumes["tls-certs"], topology.volumes["tls-data"],
	}
	for _, name := range volumeNames {
		if err := manager.Process(ctx, "docker", []string{
			"volume", "create", "--label", "golib.task=" + topology.base, name,
		}, nil, io.Discard, io.Discard); err != nil {
			return fmt.Errorf("create RabbitMQ Streams topology volume %s: %w", name, err)
		}
		lease.volumes = append(lease.volumes, name)
	}
	arguments := []string{
		"run", "--detach", "--name", topology.proxy, "--network", topology.network,
		"--network-alias", "proxy",
	}
	for _, port := range []int{15561, 15562, 15563, 15571, 8474} {
		arguments = append(arguments, "-p", "127.0.0.1::"+strconv.Itoa(port))
	}
	arguments = append(arguments, toxiproxyImage, "-host=0.0.0.0")
	if err := manager.Process(ctx, "docker", arguments, nil, io.Discard, io.Discard); err != nil {
		return fmt.Errorf("start RabbitMQ Streams topology proxy: %w", err)
	}
	lease.containers = append(lease.containers, topology.proxy)
	return nil
}

func (topology rabbitStreamTopology) environment(credentials rabbitStreamCredentials) map[string]string {
	return map[string]string{
		"RABBITSTREAM_USER": credentials.user, "RABBITSTREAM_PASSWORD": credentials.password,
		"RABBITSTREAM_ERLANG_COOKIE": credentials.cookie, "RABBITSTREAM_NETWORK": topology.network,
		"RABBITSTREAM_VOLUME_RABBIT1":   topology.volumes["rabbit1"],
		"RABBITSTREAM_VOLUME_RABBIT2":   topology.volumes["rabbit2"],
		"RABBITSTREAM_VOLUME_RABBIT3":   topology.volumes["rabbit3"],
		"RABBITSTREAM_VOLUME_TLS_CERTS": topology.volumes["tls-certs"],
		"RABBITSTREAM_VOLUME_TLS_DATA":  topology.volumes["tls-data"],
	}
}

func (manager Manager) configureRabbitStreamProxies(ctx context.Context, topology rabbitStreamTopology) error {
	request := manager.HTTPRequest
	if request == nil {
		request = httpRequest
	}
	controlURL := "http://127.0.0.1:" + topology.ports[8474]
	for _, proxy := range []struct {
		name     string
		listen   int
		upstream string
	}{
		{"rabbit1", 15561, "rabbit1:5552"}, {"rabbit2", 15562, "rabbit2:5552"},
		{"rabbit3", 15563, "rabbit3:5552"}, {"rabbit-tls", 15571, "rabbit-tls:5551"},
	} {
		payload := fmt.Appendf(nil, `{"name":%q,"listen":%q,"upstream":%q}`,
			proxy.name, "0.0.0.0:"+strconv.Itoa(proxy.listen), proxy.upstream)
		if err := manager.configureToxiproxy(ctx, request, controlURL+"/proxies", payload); err != nil {
			return fmt.Errorf("configure %s: %w", proxy.name, err)
		}
	}
	return nil
}

func (manager Manager) configureRabbitStreamTLS(
	ctx context.Context,
	topology rabbitStreamTopology,
	credentials rabbitStreamCredentials,
	files FileSystem,
) error {
	tlsContainer := topology.containers["rabbit-tls"]
	authorizationEnvironment := map[string]string{
		"RABBITSTREAM_RESTRICTED_USER": credentials.restrictedUser,
		"RABBITSTREAM_RESTRICTED_PASS": credentials.restrictedPassword,
	}
	for _, script := range []string{
		`rabbitmqctl add_user "$RABBITSTREAM_RESTRICTED_USER" "$RABBITSTREAM_RESTRICTED_PASS"`,
		`rabbitmqctl set_permissions -p / "$RABBITSTREAM_RESTRICTED_USER" '^$' '^codex-rabbitstream-allowed-.*$' '^codex-rabbitstream-allowed-.*$'`,
	} {
		command := []string{
			"exec", "-e", "RABBITSTREAM_RESTRICTED_USER", "-e", "RABBITSTREAM_RESTRICTED_PASS",
			tlsContainer, "/bin/sh", "-ec", script,
		}
		if err := manager.Process(ctx, "docker", command, authorizationEnvironment, io.Discard, io.Discard); err != nil {
			return fmt.Errorf("configure RabbitMQ Streams TLS authorization: %w", err)
		}
	}
	certgen := topology.containers["certgen"]
	for _, name := range []string{"ca.pem", "client.pem", "client-key.pem", "untrusted-client.pem", "untrusted-client-key.pem"} {
		target := filepath.Join(topology.runtime, name)
		if err := manager.Process(ctx, "docker", []string{"cp", certgen + ":/certs/" + name, target}, nil, io.Discard, io.Discard); err != nil {
			return fmt.Errorf("copy RabbitMQ Streams TLS file %s: %w", name, err)
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(name, "key.pem") {
			mode = 0o600
		}
		if err := files.Chmod(target, mode); err != nil {
			return fmt.Errorf("protect RabbitMQ Streams TLS file %s: %w", name, err)
		}
	}
	return nil
}

func (manager Manager) rabbitStreamIdentities(ctx context.Context, topology rabbitStreamTopology) ([]string, error) {
	result := make([]string, 0, 6)
	for _, item := range []struct {
		name      string
		container string
		image     string
	}{
		{"proxy", topology.proxy, toxiproxyImage},
		{"rabbit1", topology.containers["rabbit1"], rabbitStreamOldImage},
		{"rabbit2", topology.containers["rabbit2"], rabbitStreamOldImage},
		{"rabbit3", topology.containers["rabbit3"], rabbitStreamOldImage},
		{"certgen", topology.containers["certgen"], rabbitStreamImage},
		{"rabbit-tls", topology.containers["rabbit-tls"], rabbitStreamImage},
	} {
		identity, err := manager.imageIdentity(ctx, item.container, item.image)
		if err != nil {
			return nil, fmt.Errorf("read RabbitMQ Streams %s image identity: %w", item.name, err)
		}
		result = append(result, identity)
	}
	return result, nil
}

func (topology rabbitStreamTopology) testEnvironment(credentials rabbitStreamCredentials) map[string]string {
	ports := []string{topology.ports[15561], topology.ports[15562], topology.ports[15563]}
	containers := make([]string, 0, len(ports))
	for index, port := range ports {
		containers = append(containers, port+"="+topology.containers["rabbit"+strconv.Itoa(index+1)])
	}
	return map[string]string{
		"RABBITSTREAM_CLUSTER_PORTS":        strings.Join(ports, ","),
		"RABBITSTREAM_CLUSTER_CONTAINERS":   strings.Join(containers, ","),
		"RABBITSTREAM_CLUSTER_PROJECT":      topology.project,
		"RABBITSTREAM_CLUSTER_COMPOSE":      topology.compose,
		"RABBITSTREAM_NETWORK":              topology.network,
		"RABBITSTREAM_VOLUME_RABBIT1":       topology.volumes["rabbit1"],
		"RABBITSTREAM_VOLUME_RABBIT2":       topology.volumes["rabbit2"],
		"RABBITSTREAM_VOLUME_RABBIT3":       topology.volumes["rabbit3"],
		"RABBITSTREAM_VOLUME_TLS_CERTS":     topology.volumes["tls-certs"],
		"RABBITSTREAM_VOLUME_TLS_DATA":      topology.volumes["tls-data"],
		"RABBITSTREAM_ERLANG_COOKIE":        credentials.cookie,
		"RABBITSTREAM_UPGRADE_IMAGE":        rabbitStreamImage,
		"RABBITSTREAM_UPGRADE_FROM_VERSION": "4.3.4",
		"RABBITSTREAM_UPGRADE_TO_VERSION":   "4.3.5",
		"RABBITSTREAM_TLS_HOST":             "localhost",
		"RABBITSTREAM_TLS_PORT":             topology.ports[15571],
		"RABBITSTREAM_TLS_USER":             credentials.user,
		"RABBITSTREAM_TLS_PASSWORD":         credentials.password,
		"RABBITSTREAM_TLS_RUNTIME":          topology.runtime,
		"RABBITSTREAM_RESTRICTED_USER":      credentials.restrictedUser,
		"RABBITSTREAM_RESTRICTED_PASSWORD":  credentials.restrictedPassword,
	}
}

func renderRabbitStreamCompose() string {
	return fmt.Sprintf(rabbitStreamComposeTemplate,
		rabbitStreamOldImage, rabbitStreamOldImage, rabbitStreamOldImage,
		rabbitStreamImage, rabbitStreamImage,
	)
}

func writeRabbitStreamFixtureFiles(files FileSystem, topology rabbitStreamTopology) error {
	contents := map[string]string{
		"enabled_plugins": "[rabbitmq_management,rabbitmq_stream].\n",
		"rabbit1.conf":    rabbitStreamClusterConfig(topology.ports[15561]),
		"rabbit2.conf":    rabbitStreamClusterConfig(topology.ports[15562]),
		"rabbit3.conf":    rabbitStreamClusterConfig(topology.ports[15563]),
		"tls-rabbitmq.conf": strings.Join([]string{
			"stream.listeners.tcp = none", "stream.listeners.ssl.1 = 5551",
			"ssl_options.cacertfile = /certs/ca.pem", "ssl_options.certfile = /certs/server.pem",
			"ssl_options.keyfile = /certs/server-key.pem", "ssl_options.verify = verify_peer",
			"ssl_options.fail_if_no_peer_cert = true", "ssl_options.versions.1 = tlsv1.3",
			"ssl_options.versions.2 = tlsv1.2", "stream.advertised_tls_host = localhost",
			"stream.advertised_tls_port = " + topology.ports[15571], "",
		}, "\n"),
	}
	for _, name := range []string{"enabled_plugins", "rabbit1.conf", "rabbit2.conf", "rabbit3.conf", "tls-rabbitmq.conf"} {
		if err := files.WriteFile(filepath.Join(topology.directory, name), []byte(contents[name]), 0o644); err != nil {
			return fmt.Errorf("write RabbitMQ Streams fixture file %s: %w", name, err)
		}
	}
	return nil
}

func rabbitStreamClusterConfig(port string) string {
	lines := []string{
		"cluster_formation.peer_discovery_backend = classic_config",
		"cluster_formation.classic_config.nodes.1 = rabbit@rabbit1",
		"cluster_formation.classic_config.nodes.2 = rabbit@rabbit2",
		"cluster_formation.classic_config.nodes.3 = rabbit@rabbit3",
		"cluster_formation.target_cluster_size_hint = 3",
		"stream.advertised_host = localhost",
		"stream.advertised_port = " + port,
		"",
	}
	return strings.Join(lines, "\n")
}

const rabbitStreamComposeTemplate = `services:
  rabbit1:
    image: ${RABBITSTREAM_IMAGE:-%s}
    hostname: rabbit1
    environment: &rabbit-environment
      RABBITMQ_DEFAULT_USER: ${RABBITSTREAM_USER:?required}
      RABBITMQ_DEFAULT_PASS: ${RABBITSTREAM_PASSWORD:?required}
      RABBITMQ_ERLANG_COOKIE: ${RABBITSTREAM_ERLANG_COOKIE:?required}
      RABBITMQ_SERVER_ADDITIONAL_ERL_ARGS: -rabbitmq_stream initial_cluster_size 3
    networks: {rabbitstream: {aliases: [rabbit1]}}
    volumes:
      - "./enabled_plugins:/etc/rabbitmq/enabled_plugins:ro"
      - "./rabbit1.conf:/etc/rabbitmq/rabbitmq.conf:ro"
      - {type: volume, source: rabbit1-data, target: /var/lib/rabbitmq/mnesia, volume: {nocopy: true}}
    healthcheck: &rabbit-health
      test: ["CMD", "gosu", "rabbitmq", "rabbitmq-diagnostics", "-q", "check_running"]
      interval: 1s
      timeout: 5s
      retries: 90
  rabbit2:
    image: ${RABBITSTREAM_IMAGE:-%s}
    hostname: rabbit2
    environment: *rabbit-environment
    networks: {rabbitstream: {aliases: [rabbit2]}}
    volumes:
      - "./enabled_plugins:/etc/rabbitmq/enabled_plugins:ro"
      - "./rabbit2.conf:/etc/rabbitmq/rabbitmq.conf:ro"
      - {type: volume, source: rabbit2-data, target: /var/lib/rabbitmq/mnesia, volume: {nocopy: true}}
    healthcheck: *rabbit-health
  rabbit3:
    image: ${RABBITSTREAM_IMAGE:-%s}
    hostname: rabbit3
    environment: *rabbit-environment
    networks: {rabbitstream: {aliases: [rabbit3]}}
    volumes:
      - "./enabled_plugins:/etc/rabbitmq/enabled_plugins:ro"
      - "./rabbit3.conf:/etc/rabbitmq/rabbitmq.conf:ro"
      - {type: volume, source: rabbit3-data, target: /var/lib/rabbitmq/mnesia, volume: {nocopy: true}}
    healthcheck: *rabbit-health
  certgen:
    image: %s
    entrypoint: ["/bin/sh", "-ec"]
    command: ["openssl req -x509 -newkey rsa:2048 -nodes -days 2 -sha256 -subj /CN=rabbitstream-test-ca -keyout /certs/ca-key.pem -out /certs/ca.pem; openssl req -newkey rsa:2048 -nodes -sha256 -subj /CN=localhost -addext subjectAltName=DNS:localhost,IP:127.0.0.1 -keyout /certs/server-key.pem -out /certs/server.csr; openssl x509 -req -days 2 -sha256 -copy_extensions copy -in /certs/server.csr -CA /certs/ca.pem -CAkey /certs/ca-key.pem -CAcreateserial -out /certs/server.pem; openssl req -newkey rsa:2048 -nodes -sha256 -subj /CN=rabbitstream-client -keyout /certs/client-key.pem -out /certs/client.csr; openssl x509 -req -days 2 -sha256 -in /certs/client.csr -CA /certs/ca.pem -CAkey /certs/ca-key.pem -CAcreateserial -out /certs/client.pem; openssl req -x509 -newkey rsa:2048 -nodes -days 2 -sha256 -subj /CN=untrusted-rabbitstream-client -keyout /certs/untrusted-client-key.pem -out /certs/untrusted-client.pem; chown -R 999:999 /certs; chmod 600 /certs/*-key.pem"]
    volumes: ["tls-certs:/certs"]
  rabbit-tls:
    image: %s
    hostname: rabbit-tls
    depends_on: {certgen: {condition: service_completed_successfully}}
    environment:
      RABBITMQ_DEFAULT_USER: ${RABBITSTREAM_USER:?required}
      RABBITMQ_DEFAULT_PASS: ${RABBITSTREAM_PASSWORD:?required}
      RABBITMQ_ERLANG_COOKIE: ${RABBITSTREAM_ERLANG_COOKIE:?required}
    networks: {rabbitstream: {aliases: [rabbit-tls]}}
    volumes:
      - "./enabled_plugins:/etc/rabbitmq/enabled_plugins:ro"
      - "./tls-rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf:ro"
      - "tls-certs:/certs:ro"
      - {type: volume, source: tls-data, target: /var/lib/rabbitmq/mnesia, volume: {nocopy: true}}
    healthcheck: *rabbit-health
networks:
  rabbitstream:
    external: true
    name: ${RABBITSTREAM_NETWORK:?required}
volumes:
  rabbit1-data: {external: true, name: "${RABBITSTREAM_VOLUME_RABBIT1:?required}"}
  rabbit2-data: {external: true, name: "${RABBITSTREAM_VOLUME_RABBIT2:?required}"}
  rabbit3-data: {external: true, name: "${RABBITSTREAM_VOLUME_RABBIT3:?required}"}
  tls-certs: {external: true, name: "${RABBITSTREAM_VOLUME_TLS_CERTS:?required}"}
  tls-data: {external: true, name: "${RABBITSTREAM_VOLUME_TLS_DATA:?required}"}
`
