package specification

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestWalkJSONValueConsumesCompleteArray(t *testing.T) {
	decoder := json.NewDecoder(strings.NewReader(`[1,2]`))
	if err := walkJSONValue(decoder); err != nil {
		t.Fatal(err)
	}
	if token, err := decoder.Token(); !errors.Is(err, io.EOF) {
		t.Fatalf("decoder.Token() = %v, %v, want EOF", token, err)
	}
}

func TestLoadMonitoringRejectsInvalidPolicies(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	authorities := `[{"id":"rfc9110-source","kind":"specification","version":"RFC 9110","url":"https://example.com/specification","sha256":"` + strings.Repeat("a", 64) + `","specifications":["RFC 9110"]},{"id":"rfc9110-errata","kind":"errata","url":"https://example.com/errata","sha256":"` + strings.Repeat("b", 64) + `","specifications":["RFC 9110"]}]`
	valid := `{"schema_version":1,"reviewed_at":"2026-08-30","review_interval_days":90,"authorities":` + authorities + `}`
	tests := map[string]string{
		"duplicate key":     strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		"unknown field":     strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"extra":true`, 1),
		"multiple values":   valid + `{}`,
		"trailing data":     valid + `x`,
		"schema":            strings.Replace(valid, `"schema_version":1`, `"schema_version":2`, 1),
		"date":              strings.Replace(valid, `"2026-08-30"`, `"not-a-date"`, 1),
		"future":            strings.Replace(valid, `"2026-08-30"`, `"2026-08-31"`, 1),
		"zero interval":     strings.Replace(valid, `"review_interval_days":90`, `"review_interval_days":0`, 1),
		"long interval":     strings.Replace(valid, `"review_interval_days":90`, `"review_interval_days":91`, 1),
		"stale":             strings.Replace(strings.Replace(valid, `"2026-08-30"`, `"2026-01-01"`, 1), `"review_interval_days":90`, `"review_interval_days":1`, 1),
		"no authorities":    strings.Replace(valid, authorities, `[]`, 1),
		"duplicate":         strings.Replace(valid, `"id":"rfc9110-source"`, `"id":"rfc9110-errata"`, 1),
		"empty identifier":  strings.Replace(valid, `"id":"rfc9110-source"`, `"id":""`, 1),
		"invalid URL":       strings.Replace(valid, `https://example.com/specification`, `http://example.com/specification`, 1),
		"URL userinfo":      strings.Replace(valid, `https://example.com/specification`, `https://user@example.com/specification`, 1),
		"URL fragment":      strings.Replace(valid, `https://example.com/specification`, `https://example.com/specification#section`, 1),
		"invalid digest":    strings.Replace(valid, strings.Repeat("a", 64), "invalid", 1),
		"unsupported kind":  strings.Replace(valid, `"kind":"specification"`, `"kind":"blog"`, 1),
		"no change feed":    strings.Replace(valid, `"kind":"errata"`, `"kind":"specification"`, 1),
		"missing binding":   strings.Replace(valid, `,"specifications":["RFC 9110"]`, "", 1),
		"missing version":   strings.Replace(valid, `"version":"RFC 9110",`, "", 1),
		"unknown binding":   strings.Replace(valid, `"specifications":["RFC 9110"]`, `"specifications":["RFC 9999"]`, 1),
		"duplicate binding": strings.Replace(valid, `["RFC 9110"]`, `["RFC 9110","RFC 9110"]`, 1),
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "specification/monitoring.json"), content)
			if _, err := loadMonitoring(root, "specification/monitoring.json", []string{"RFC 9110"}, now); err == nil {
				t.Fatal("loadMonitoring() error = nil")
			}
		})
	}
	for name, specifications := range map[string][]string{
		"empty declared specification":     {""},
		"duplicate declared specification": {"RFC 9110", "RFC 9110"},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "specification/monitoring.json"), valid)
			if _, err := loadMonitoring(root, "specification/monitoring.json", specifications, now); err == nil {
				t.Fatal("loadMonitoring() error = nil")
			}
		})
	}
	if _, err := loadMonitoring(t.TempDir(), "specification/monitoring.json", []string{"RFC 9110"}, now); err == nil {
		t.Fatal("loadMonitoring(missing) error = nil")
	}
}

func TestLoadMonitoringRejectsUnboundedAuthoritySet(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	items := make([]string, 0, 66)
	for index := range 65 {
		items = append(items, fmt.Sprintf(`{"id":"source-%d","kind":"specification","version":"RFC 9110","url":"https://example.com/specification/%d","sha256":"%s","specifications":["RFC 9110"]}`, index, index, strings.Repeat("a", 64)))
	}
	items = append(items, `{"id":"errata","kind":"errata","url":"https://example.com/errata","sha256":"`+strings.Repeat("b", 64)+`","specifications":["RFC 9110"]}`)
	policy := `{"schema_version":1,"reviewed_at":"2026-08-30","review_interval_days":90,"authorities":[` + strings.Join(items, ",") + `]}`
	root := t.TempDir()
	write(t, filepath.Join(root, "specification/monitoring.json"), policy)

	if _, err := loadMonitoring(root, "specification/monitoring.json", []string{"RFC 9110"}, now); err == nil || !strings.Contains(err.Error(), "too many authorities") {
		t.Fatalf("loadMonitoring() error = %v", err)
	}
}

func TestValidateSourceManifestFormats(t *testing.T) {
	digest := strings.Repeat("a", 64)
	validTSV := "id\tversion\turl\tsha256\tstatus\nsource\tv1\thttps://example.com/spec\t" + digest + "\tpinned\n"
	tests := []struct {
		name string
		path string
		data string
		ok   bool
	}{
		{"TSV", "manifest.tsv", validTSV, true},
		{"TSV no rows", "manifest.tsv", strings.Split(validTSV, "\n")[0], false},
		{"TSV missing column", "manifest.tsv", strings.Replace(validTSV, "version\t", "", 1), false},
		{"TSV duplicate column", "manifest.tsv", "id\tversion\tid\turl\tsha256\tstatus\nsource\tv1\tsource\thttps://example.com/spec\t" + digest + "\tpinned\n", false},
		{"TSV field count", "manifest.tsv", validTSV + "short\n", false},
		{"TSV invalid row", "manifest.tsv", strings.Replace(validTSV, "https://", "http://", 1), false},
		{"TSV URL userinfo", "manifest.tsv", strings.Replace(validTSV, "https://example.com", "https://user@example.com", 1), false},
		{"TSV URL fragment", "manifest.tsv", strings.Replace(validTSV, "https://example.com/spec", "https://example.com/spec#section", 1), false},
		{"JSON", "manifest.json", `{"url":"https://example.com","sha256":"` + digest + `"}`, true},
		{"JSON duplicate key", "manifest.json", `{"url":"https://example.com","sha256":"` + digest + `","sha256":"` + digest + `"}`, false},
		{"JSON array", "manifest.json", `[{"url":"https://example.com","sha256":"` + digest + `"}]`, true},
		{"JSON array invalid", "manifest.json", `[{"url":"https://example.com","sha256":"invalid"}]`, false},
		{"JSON unrelated digest", "manifest.json", `{"metadata":{"sourceSha256":"` + digest + `"}}`, false},
		{"JSON malformed", "manifest.json", `{`, false},
		{"JSON no digest", "manifest.json", `{}`, false},
		{"JSON invalid digest", "manifest.json", `{"url":"https://example.com","sha256":"invalid"}`, false},
		{"JSON non-string digest", "manifest.json", `{"sha256":1}`, false},
		{"JSON nested invalid digest", "manifest.json", `{"sources":[{"sha256":"invalid"}]}`, false},
		{"unsupported", "manifest.yaml", "sources: []", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateSourceManifest(test.path, []byte(test.data))
			if (err == nil) != test.ok {
				t.Fatalf("validateSourceManifest() error = %v", err)
			}
		})
	}
}

func TestTestEvidenceBoundsAndRejectsSymlinks(t *testing.T) {
	root := t.TempDir()
	if got, err := testEvidence(root, ".", nil); err != nil || len(got) != 0 {
		t.Fatalf("testEvidence(empty) = %#v, %v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".verification"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "nested", "contract_test.go"), "package example\n\nimport \"testing\"\n\nfunc TestContract(t *testing.T) {}\n")
	write(t, filepath.Join(root, "nested", "invalid_signature_test.go"), "package example\n\nfunc TestNotExecutable() {}\n")
	write(t, filepath.Join(root, "nested", "excluded_test.go"), "//go:build never\n\npackage example\n\nimport \"testing\"\n\nfunc TestExcluded(t *testing.T) {}\n")
	for _, directory := range []string{"vendor/example", "testdata", ".hidden", "_private"} {
		write(t, filepath.Join(root, directory, "ignored_test.go"), "package ignored\n\nimport \"testing\"\n\nfunc TestIgnored(t *testing.T) {}\n")
	}
	if got, err := testEvidence(root, ".", nil); err != nil {
		t.Fatalf("testEvidence() error = %v", err)
	} else if _, exists := got["TestContract"]; !exists {
		t.Fatalf("testEvidence() = %#v", got)
	} else if _, exists := got["TestNotExecutable"]; exists {
		t.Fatalf("testEvidence() includes invalid Go test signature: %#v", got)
	} else if _, exists := got["TestExcluded"]; exists {
		t.Fatalf("testEvidence() includes build-excluded Go test: %#v", got)
	} else if _, exists := got["TestIgnored"]; exists {
		t.Fatalf("testEvidence() includes undiscovered package evidence: %#v", got)
	}
	if err := os.Symlink("contract_test.go", filepath.Join(root, "nested", "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := testEvidence(root, ".", nil); err == nil {
		t.Fatal("testEvidence(symlink) error = nil")
	}
	if _, err := testEvidence(filepath.Join(root, "missing"), ".", nil); err == nil {
		t.Fatal("testEvidence(missing root) error = nil")
	}
	invalid := t.TempDir()
	write(t, filepath.Join(invalid, "invalid_test.go"), "package")
	if _, err := testEvidence(invalid, ".", nil); err == nil {
		t.Fatal("testEvidence(invalid Go) error = nil")
	}
	invalidBuild := t.TempDir()
	write(t, filepath.Join(invalidBuild, "invalid_test.go"), "//go:build (\n\npackage example\n")
	if _, err := testEvidence(invalidBuild, ".", nil); err == nil {
		t.Fatal("testEvidence(invalid build constraint) error = nil")
	}

	large := t.TempDir()
	write(t, filepath.Join(large, "large_test.go"), strings.Repeat("x", maximumEvidenceSize+1))
	if _, err := testEvidence(large, ".", nil); err == nil {
		t.Fatal("testEvidence(large file) error = nil")
	}
	aggregate := t.TempDir()
	write(t, filepath.Join(aggregate, "one_test.go"), "package example\n// "+strings.Repeat("x", maximumEvidenceSize/2))
	write(t, filepath.Join(aggregate, "two_test.go"), "package example\n// "+strings.Repeat("x", maximumEvidenceSize/2))
	if _, err := testEvidence(aggregate, ".", nil); err == nil {
		t.Fatal("testEvidence(aggregate) error = nil")
	}
}

func TestIsExecutableEvidenceRecognizesGoTestSignatures(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"test", `package example; import ("fmt"; "testing"); func TestContract(t *testing.T) { _ = fmt.Sprint() }`, true},
		{"bare test", `package example; import "testing"; func Test(t *testing.T) {}`, true},
		{"fuzz alias", `package example; import testpkg "testing"; func FuzzContract(f *testpkg.F) {}`, true},
		{"benchmark dot import", `package example; import . "testing"; func BenchmarkContract(b *B) {}`, true},
		{"blank import", `package example; import _ "testing"; func TestContract(t *testing.T) {}`, false},
		{"wrong dot type", `package example; import . "testing"; func TestContract(t *F) {}`, false},
		{"helper", `package example; import "testing"; func Helper(t *testing.T) {}`, false},
		{"lowercase suffix", `package example; import "testing"; func Testcontract(t *testing.T) {}`, false},
		{"unicode modifier suffix", `package example; import "testing"; func Testʰ(t *testing.T) {}`, true},
		{"unicode lowercase suffix", `package example; import "testing"; func Testé(t *testing.T) {}`, false},
		{"non-pointer", `package example; import "testing"; func TestContract(t testing.T) {}`, false},
		{"wrong selector", `package example; import "testing"; func TestContract(t *testing.F) {}`, false},
		{"result", `package example; import "testing"; func TestContract(t *testing.T) error { return nil }`, false},
		{"type parameters", `package example; import "testing"; func TestContract[V any](t *testing.T) {}`, false},
		{"no parameters", `package example; func TestContract() {}`, false},
		{"two parameters", `package example; import "testing"; func TestContract(t *testing.T, other int) {}`, false},
		{"two names", `package example; import "testing"; func TestContract(t, other *testing.T) {}`, false},
		{"method", `package example; import "testing"; type suite struct{}; func (suite) TestContract(t *testing.T) {}`, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "evidence_test.go", test.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			var function *ast.FuncDecl
			for _, declaration := range file.Decls {
				if candidate, ok := declaration.(*ast.FuncDecl); ok {
					function = candidate
				}
			}
			if function == nil {
				t.Fatal("fixture contains no function")
			}
			if got := isExecutableEvidence(file, function); got != test.want {
				t.Fatalf("isExecutableEvidence() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestDiscoveryAndMappingHelpers(t *testing.T) {
	root := t.TempDir()
	if found, err := hasSpecificationArtifacts(root, "."); err != nil || found {
		t.Fatalf("hasSpecificationArtifacts(empty) = %v, %v", found, err)
	}
	write(t, filepath.Join(root, "docs/specification-decisions.md"), "# Decisions\n")
	if found, err := hasSpecificationArtifacts(root, "."); err != nil || found {
		t.Fatalf("hasSpecificationArtifacts(register) = %v, %v", found, err)
	}
	if err := os.Remove(filepath.Join(root, "docs/specification-decisions.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "specification"), 0o700); err != nil {
		t.Fatal(err)
	}
	if found, err := hasSpecificationArtifacts(root, "."); err != nil || !found {
		t.Fatalf("hasSpecificationArtifacts(specification) = %v, %v", found, err)
	}
	if _, err := hasSpecificationArtifacts(root, strings.Repeat("x", 4096)); err == nil {
		t.Fatal("hasSpecificationArtifacts(invalid path) error = nil")
	}
	if _, err := matrixDecisions([]byte("# Matrix\n")); err == nil {
		t.Fatal("matrixDecisions(empty) error = nil")
	}

	modules := []inventory.Module{{Directory: "."}, {Directory: "nested"}}
	for _, path := range []string{"/absolute", ".", "..", "../escape", "nested/specification/manifest.tsv"} {
		if moduleOwnsPath(".", path, modules) {
			t.Errorf("moduleOwnsPath(root, %q) = true", path)
		}
	}
	if !moduleOwnsPath(".", "specification/manifest.tsv", modules) {
		t.Error("moduleOwnsPath(root manifest) = false")
	}
	if moduleOwnsPath("nested", "nested", modules) || !moduleOwnsPath("nested", "nested/specification/manifest.tsv", modules) {
		t.Error("moduleOwnsPath(nested) does not enforce the module boundary")
	}
}

func TestCheckRejectsDecisionIdentifierSharedAcrossModules(t *testing.T) {
	root, catalog := validFixture(t)
	nested := catalog.Modules[0]
	nested.Directory = "nested"
	nested.Provenance = []byte(`["nested/specification/manifest.tsv"]`)
	for _, relative := range []string{"README.md", "COMPATIBILITY.md", "CHANGELOG.md", "docs/specification-decisions.md", "specification/README.md", "specification/manifest.tsv", "specification/monitoring.json", "specification/decisions.json", "specification/conformance.json", "testdata/request-target.json", "testdata/peer.tsv"} {
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatal(err)
		}
		content := strings.ReplaceAll(string(data), `"docs/specification-decisions.md"`, `"nested/docs/specification-decisions.md"`)
		content = strings.ReplaceAll(content, `"README.md"`, `"nested/README.md"`)
		content = strings.ReplaceAll(content, `"COMPATIBILITY.md"`, `"nested/COMPATIBILITY.md"`)
		content = strings.ReplaceAll(content, `"CHANGELOG.md"`, `"nested/CHANGELOG.md"`)
		content = strings.ReplaceAll(content, `"specification/README.md"`, `"nested/specification/README.md"`)
		content = strings.ReplaceAll(content, `"testdata/request-target.json"`, `"nested/testdata/request-target.json"`)
		content = strings.ReplaceAll(content, `"testdata/peer.tsv"`, `"nested/testdata/peer.tsv"`)
		write(t, filepath.Join(root, "nested", relative), content)
	}
	write(t, filepath.Join(root, "nested/contract_test.go"), "package nested\n\nimport \"testing\"\n\nfunc TestRequestTargetContract(t *testing.T) {}\nfunc FuzzRequestTargetContract(f *testing.F) {}\n")
	refreshModuleDecisionArtifacts(t, root, "nested", "https://www.rfc-editor.org/rfc/rfc9110.txt")
	catalog.Modules = append(catalog.Modules, nested)
	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "shared by modules") {
		t.Fatalf("Check(shared decision) error = %v", err)
	}
}

func TestFetchAuthorityRejectsEveryFailureBoundary(t *testing.T) {
	body := "authority"
	digest := sha256.Sum256([]byte(body))
	item := authority{ID: "errata", URL: "https://example.com/errata", SHA256: hexDigest(digest)}
	if err := fetchAuthority(context.Background(), client(response(http.StatusOK, body, nil, nil), nil), item); err != nil {
		t.Fatalf("fetchAuthority() error = %v", err)
	}
	tests := []struct {
		name   string
		item   authority
		client *http.Client
	}{
		{"request", authority{ID: "bad", URL: "://"}, client(response(http.StatusOK, body, nil, nil), nil)},
		{"fetch", item, client(nil, errors.New("network"))},
		{"read", item, client(response(http.StatusOK, "", errors.New("read"), nil), nil)},
		{"close", item, client(response(http.StatusOK, body, nil, errors.New("close")), nil)},
		{"status", item, client(response(http.StatusBadGateway, body, nil, nil), nil)},
		{"large", item, client(response(http.StatusOK, strings.Repeat("x", maximumAuthoritySize+1), nil, nil), nil)},
		{"changed", authority{ID: item.ID, URL: item.URL, SHA256: strings.Repeat("0", 64)}, client(response(http.StatusOK, body, nil, nil), nil)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := fetchAuthority(context.Background(), test.client, test.item); err == nil {
				t.Fatal("fetchAuthority() error = nil")
			}
		})
	}
}

func hexDigest(value [32]byte) string {
	const digits = "0123456789abcdef"
	result := make([]byte, 64)
	for index, item := range value {
		result[index*2] = digits[item>>4]
		result[index*2+1] = digits[item&15]
	}
	return string(result)
}

type roundTripper func(*http.Request) (*http.Response, error)

func (function roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func client(result *http.Response, err error) *http.Client {
	return &http.Client{Transport: roundTripper(func(*http.Request) (*http.Response, error) {
		return result, err
	})}
}

func response(status int, content string, readErr, closeErr error) *http.Response {
	return &http.Response{StatusCode: status, Body: &faultBody{Reader: strings.NewReader(content), readErr: readErr, closeErr: closeErr}}
}

type faultBody struct {
	io.Reader
	readErr  error
	closeErr error
}

func (body *faultBody) Read(buffer []byte) (int, error) {
	if body.readErr != nil {
		return 0, body.readErr
	}
	return body.Reader.Read(buffer)
}

func (body *faultBody) Close() error {
	return body.closeErr
}
