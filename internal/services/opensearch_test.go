package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

const validOpenSearchLock = `opensearch_image_repository='opensearchproject/opensearch'
opensearch_old_version='2.19.6'
opensearch_old_digest='sha256:8690b204fe914c60ca76d451ac73bc0481e034d32d3779944c8caca56a2b003f'
opensearch_new_version='3.8.0'
opensearch_new_digest='sha256:bcc1797519726ceb6d651d4a3e60b7c30da91793914a8dfe75fd441d4f641509'
`

func TestParseOpenSearchImagesAcceptsStrictDigestLock(t *testing.T) {
	images, err := ParseOpenSearchImages(strings.NewReader(validOpenSearchLock))
	if err != nil {
		t.Fatalf("ParseOpenSearchImages() error = %v", err)
	}
	if images.Repository != "opensearchproject/opensearch" || images.OldVersion != "2.19.6" || images.NewVersion != "3.8.0" ||
		images.OldImage() != "opensearchproject/opensearch@sha256:8690b204fe914c60ca76d451ac73bc0481e034d32d3779944c8caca56a2b003f" ||
		images.NewImage() != "opensearchproject/opensearch@sha256:bcc1797519726ceb6d651d4a3e60b7c30da91793914a8dfe75fd441d4f641509" {
		t.Fatalf("images = %#v", images)
	}
}

func TestParseOpenSearchImagesRejectsUntrustedLocks(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	base := "opensearch_image_repository='repo/name'\nopensearch_old_version='1.0.0'\nopensearch_old_digest='" + digest + "'\nopensearch_new_version='2.0.0'\nopensearch_new_digest='sha256:" + strings.Repeat("b", 64) + "'\n"
	for _, test := range []struct {
		name string
		body string
	}{
		{"read", "read failure"},
		{"oversized", strings.Repeat("x", maximumOpenSearchLock+1)},
		{"blank", "\n" + base},
		{"comment", "# comment\n" + base},
		{"unknown", base + "other='value'\n"},
		{"duplicate", base + "opensearch_new_version='3.0.0'\n"},
		{"unquoted", strings.Replace(base, "repo/name'", "repo/name", 1)},
		{"missing", strings.Replace(base, "opensearch_new_version='2.0.0'\n", "", 1)},
		{"repository", strings.Replace(base, "repo/name", "../repo", 1)},
		{"version", strings.Replace(base, "2.0.0", "latest", 1)},
		{"digest", strings.Replace(base, strings.Repeat("b", 64), "bad", 1)},
		{"same version", strings.Replace(base, "2.0.0", "1.0.0", 1)},
		{"same digest", strings.Replace(base, "sha256:"+strings.Repeat("b", 64), digest, 1)},
		{"injection", strings.Replace(base, "repo/name", "repo/name$(touch bad)", 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.body)
			if test.name == "read" {
				if _, err := ParseOpenSearchImages(failingReader{}); err == nil {
					t.Fatal("ParseOpenSearchImages() error = nil")
				}
				return
			}
			if _, err := ParseOpenSearchImages(reader); err == nil {
				t.Fatal("ParseOpenSearchImages() error = nil")
			}
		})
	}
}

func TestManagerStartsOpenSearchFromRepositoryLock(t *testing.T) {
	images, err := ParseOpenSearchImages(strings.NewReader(validOpenSearchLock))
	if err != nil {
		t.Fatal(err)
	}
	backend := &fakeBackend{}
	urls := []string{}
	manager := Manager{
		Process: backend.run, Probe: successfulProbe, Wait: successfulWait, Token: fixedToken,
		OpenSearch: &images,
		HTTPProbe: func(_ context.Context, url string) error {
			urls = append(urls, url)
			return nil
		},
	}
	lease, err := manager.Start(context.Background(), []string{"opensearch"})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	wantEnvironment := map[string]string{"OPENSEARCH_URL": "http://127.0.0.1:40001", "OPENSEARCH_EXPECTED_VERSION": "3.8.0"}
	if !reflect.DeepEqual(lease.Environment(), wantEnvironment) || !reflect.DeepEqual(urls, []string{"http://127.0.0.1:40001/"}) {
		t.Fatalf("environment/URLs = %#v, %#v", lease.Environment(), urls)
	}
	if identity := lease.Identities()["opensearch"]; !strings.HasPrefix(identity, images.NewImage()+"#sha256:") {
		t.Fatalf("identity = %q", identity)
	}
	commands := strings.Join(backend.commands, "\n")
	for _, argument := range []string{"--cpus=1", "--memory=1g", "--pids-limit=512", "--ulimit nofile=1024:1024", "DISABLE_SECURITY_PLUGIN=true"} {
		if !strings.Contains(commands, argument) {
			t.Fatalf("commands missing %q: %s", argument, commands)
		}
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestManagerRejectsMissingAndInvalidOpenSearchPolicy(t *testing.T) {
	backend := &fakeBackend{}
	manager := Manager{Process: backend.run, Probe: successfulProbe, Wait: successfulWait, Token: fixedToken}
	if _, err := manager.Start(context.Background(), []string{"opensearch"}); err == nil || !strings.Contains(err.Error(), "unsupported service") {
		t.Fatalf("Start(missing OpenSearch) error = %v", err)
	}
	manager.OpenSearch = &OpenSearchImages{}
	if _, err := manager.Start(context.Background(), []string{"opensearch"}); err == nil || !strings.Contains(err.Error(), "OpenSearch image policy") {
		t.Fatalf("Start(invalid OpenSearch) error = %v", err)
	}
}

func TestHTTPProbeAcceptsSuccessAndRejectsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" || request.URL.Path == "/ok" {
			response.WriteHeader(http.StatusNoContent)
			return
		}
		if request.URL.Path == "/redirect" {
			http.Redirect(response, request, "/ok", http.StatusFound)
			return
		}
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	if err := httpProbe(context.Background(), server.URL+"/ok"); err != nil {
		t.Fatalf("httpProbe(success) error = %v", err)
	}
	if err := httpProbe(context.Background(), server.URL+"/fail"); err == nil {
		t.Fatal("httpProbe(failure) error = nil")
	}
	if err := httpProbe(context.Background(), server.URL+"/redirect"); err == nil || !strings.Contains(err.Error(), "redirects are not allowed") {
		t.Fatalf("httpProbe(redirect) error = %v", err)
	}
	if err := httpProbe(context.Background(), "://invalid"); err == nil {
		t.Fatal("httpProbe(invalid URL) error = nil")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := httpProbe(canceled, server.URL+"/ok"); !errors.Is(err, context.Canceled) {
		t.Fatalf("httpProbe(canceled) error = %v", err)
	}
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	definition := definition{name: "opensearch", requiresHTTP: true}
	if err := (Manager{Process: (&fakeBackend{}).run, Attempts: 1}).waitReady(context.Background(), definition, "container", parsed.Port()); err != nil {
		t.Fatalf("waitReady(default HTTP) error = %v", err)
	}
}

func TestOpenSearchReadinessUsesDefaultsAndReportsWaitFailure(t *testing.T) {
	images, err := ParseOpenSearchImages(strings.NewReader(validOpenSearchLock))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := images.definition()
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("not ready")
	manager := Manager{
		Process: (&fakeBackend{}).run, Attempts: 2,
		HTTPProbe: func(context.Context, string) error { return failure },
		Wait:      func(context.Context, time.Duration) error { return errors.New("stop wait") },
	}
	if err := manager.waitReady(context.Background(), definition, "container", "9200"); err == nil || !strings.Contains(err.Error(), "wait for opensearch") {
		t.Fatalf("waitReady(OpenSearch) error = %v", err)
	}
}
