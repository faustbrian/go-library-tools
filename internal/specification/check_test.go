package specification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestCheckAcceptsCompleteResolvedDecision(t *testing.T) {
	root, catalog := validFixture(t)
	report, err := Check(root, catalog)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if report.Modules != 1 || report.Decisions != 1 || report.Unresolved != 0 {
		t.Fatalf("Check() report = %#v", report)
	}
}

func TestCheckRequiresStructuredDecisionAndConformanceManifests(t *testing.T) {
	for _, name := range []string{"decisions.json", "conformance.json"} {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			if err := os.Remove(filepath.Join(root, "specification", name)); err != nil {
				t.Fatal(err)
			}
			if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), name) {
				t.Fatalf("Check() error = %v", err)
			}
		})
	}
}

func TestCheckRejectsInvalidDecisionHistoryArtifacts(t *testing.T) {
	malformed := "{"
	schema := `{"schema_version":2,"entries":[]}`
	stale := `{"schema_version":1,"entries":[]}`
	for name, content := range map[string]*string{"missing": nil, "malformed": &malformed, "schema": &schema, "stale": &stale} {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			path := filepath.Join(root, "specification/decision-history.json")
			if content == nil {
				removeFile(t, path)
			} else {
				write(t, path, *content)
			}
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckBindsDecisionsToVersionedSourceAuthorities(t *testing.T) {
	root, catalog := validFixture(t)
	decisionsPath := filepath.Join(root, "specification/decisions.json")
	decisions, err := os.ReadFile(decisionsPath)
	if err != nil {
		t.Fatal(err)
	}
	write(t, decisionsPath, strings.Replace(string(decisions), `"version":"RFC 9110"`, `"version":"RFC 9110-obsolete"`, 1))

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "does not match authority version") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRequiresAttributableDecisionEvidence(t *testing.T) {
	root, catalog := validFixture(t)
	for _, relative := range []string{"specification/decisions.json", "specification/conformance.json"} {
		path := filepath.Join(root, relative)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		write(t, path, strings.ReplaceAll(string(data), "testdata/request-target.json", "not-a-fixture"))
	}

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "fixture evidence") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRequiresExactConformanceBindings(t *testing.T) {
	root, catalog := validFixture(t)
	path := filepath.Join(root, "specification/conformance.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), `"fixtures":["testdata/request-target.json"]`, `"fixtures":["testdata/request-target.json","testdata/peer.tsv"]`, 1))

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "fixtures differ") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsUnknownNormativeAuthorityBinding(t *testing.T) {
	root, catalog := validFixture(t)
	path := filepath.Join(root, "specification/conformance.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), `"authoritative_sources":["rfc9110-source"]`, `"authoritative_sources":["rfc9110-source","unknown-source"]`, 1))

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "unknown authoritative source") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsInvalidRequirementStrength(t *testing.T) {
	root, catalog := validFixture(t)
	path := filepath.Join(root, "specification/decisions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), `"requirement_strength":"not specified"`, `"requirement_strength":"popular behavior"`, 1))

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "requirement_strength") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsIncompleteChangeControlContract(t *testing.T) {
	root, catalog := validFixture(t)
	write(t, filepath.Join(root, "COMPATIBILITY.md"), "# Compatibility\n")

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "does not link") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRequiresDecisionSpecificChangelogRecord(t *testing.T) {
	root, catalog := validFixture(t)
	write(t, filepath.Join(root, "CHANGELOG.md"), "# Changelog\n\n[Decision register](docs/specification-decisions.md)\n")
	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "changelog") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRequiresSupersededDecisionReplacementLink(t *testing.T) {
	root, catalog := validFixture(t)
	decisionPath := filepath.Join(root, "specification/decisions.json")
	decisionData, err := os.ReadFile(decisionPath)
	if err != nil {
		t.Fatal(err)
	}
	decisions, err := loadDecisions(decisionData)
	if err != nil {
		t.Fatal(err)
	}
	replacement := decisions.Decisions[0]
	replacement.ID = "EXAMPLE-DEC-002"
	replacement.Title = "Replacement policy"
	decisions.Decisions[0].Status = "superseded"
	decisions.Decisions[0].Replacement = replacement.ID
	decisions.Decisions = append(decisions.Decisions, replacement)
	writeJSON(t, decisionPath, decisions)

	conformancePath := filepath.Join(root, "specification/conformance.json")
	conformanceData, err := os.ReadFile(conformancePath)
	if err != nil {
		t.Fatal(err)
	}
	conformance, err := loadConformance(conformanceData)
	if err != nil {
		t.Fatal(err)
	}
	replacementRow := conformance.Decisions[0]
	replacementRow.ID = replacement.ID
	conformance.Decisions = append(conformance.Decisions, replacementRow)
	writeJSON(t, conformancePath, conformance)

	registerPath := filepath.Join(root, "docs/specification-decisions.md")
	register, err := os.ReadFile(registerPath)
	if err != nil {
		t.Fatal(err)
	}
	registerText := strings.ReplaceAll(string(register), `"status":"resolved"`, `"status":"superseded"`)
	registerText += "\nReplacement: EXAMPLE-DEC-002\n"
	replacementEntry := strings.ReplaceAll(string(register), "EXAMPLE-DEC-001", "EXAMPLE-DEC-002")
	replacementEntry = strings.ReplaceAll(replacementEntry, "Request target normalization", "Replacement policy")
	write(t, registerPath, registerText+"\n"+replacementEntry)
	changelogPath := filepath.Join(root, "CHANGELOG.md")
	changelog, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatal(err)
	}
	write(t, changelogPath, string(changelog)+"\n- EXAMPLE-DEC-002: Replacement policy.\n")
	matrixPath := filepath.Join(root, "specification/README.md")
	matrix, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatal(err)
	}
	write(t, matrixPath, string(matrix)+"\nEXAMPLE-DEC-002\n")

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "replacement link") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsInvalidSupersessionTargets(t *testing.T) {
	for name, replacement := range map[string]string{"missing replacement": "", "unknown replacement": "EXAMPLE-DEC-999"} {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			mutateDecisionManifest(t, root, func(manifest *decisionManifest) {
				manifest.Decisions[0].Status = "superseded"
				manifest.Decisions[0].Replacement = replacement
			})
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsDecisionSetDrift(t *testing.T) {
	tests := map[string]func(*testing.T, string){
		"Markdown register": func(t *testing.T, root string) {
			path := filepath.Join(root, "docs/specification-decisions.md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			write(t, path, string(data)+"\n## EXAMPLE-DEC-999: Unknown\n")
		},
		"conformance matrix": func(t *testing.T, root string) {
			path := filepath.Join(root, "specification/README.md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			write(t, path, string(data)+"\nEXAMPLE-DEC-999\n")
		},
		"conformance JSON": func(t *testing.T, root string) {
			path := filepath.Join(root, "specification/conformance.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := loadConformance(data)
			if err != nil {
				t.Fatal(err)
			}
			orphan := manifest.Decisions[0]
			orphan.ID = "EXAMPLE-DEC-999"
			manifest.Decisions = append(manifest.Decisions, orphan)
			writeJSON(t, path, manifest)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			mutate(t, root)
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsInvalidChangeControl(t *testing.T) {
	tests := map[string]func(*testing.T, string, *decisionManifest){
		"empty path": func(_ *testing.T, _ string, manifest *decisionManifest) { manifest.ChangeControl.README = "" },
		"invalid path": func(_ *testing.T, _ string, manifest *decisionManifest) {
			manifest.ChangeControl.README = "../README.md"
		},
		"missing document": func(_ *testing.T, _ string, manifest *decisionManifest) {
			manifest.ChangeControl.Compatibility = "missing/COMPATIBILITY.md"
		},
		"missing template": func(_ *testing.T, _ string, manifest *decisionManifest) {
			manifest.ChangeControl.PullRequestTemplate = ".github/missing.md"
		},
		"incomplete template": func(t *testing.T, root string, _ *decisionManifest) {
			write(t, filepath.Join(root, ".github/pull_request_template.md"), "## Pull Request\n")
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			mutateDecisionManifest(t, root, func(manifest *decisionManifest) { mutate(t, root, manifest) })
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}

	root, _ := validFixture(t)
	manifestData, err := os.ReadFile(filepath.Join(root, "specification/decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := loadDecisions(manifestData)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateChangeControl(root, inventory.Module{Directory: "nested"}, []inventory.Module{{Directory: "."}, {Directory: "nested"}}, manifest.ChangeControl, "nested/docs/specification-decisions.md", manifest.Decisions); err == nil {
		t.Fatal("validateChangeControl(outside module) error = nil")
	}
}

func TestCheckRequiresDecisionArtifactsForEachSpecificationBackedModule(t *testing.T) {
	root, catalog := validFixture(t)
	second := catalog.Modules[0]
	second.Directory = "nested"
	second.ModulePath = "example.com/nested"
	second.Provenance = json.RawMessage(`["nested/specification/manifest.tsv"]`)
	catalog.Modules = append(catalog.Modules, second)
	write(t, filepath.Join(root, "nested/specification/manifest.tsv"), "id\tversion\tsections\trole\turl\tsha256\tstatus\n"+
		"rfc9110\tRFC-9110\t7.1\tHTTP-semantics\thttps://www.rfc-editor.org/rfc/rfc9110.txt\t"+
		"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a\tpinned\n")

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), `module "nested"`) {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckBlocksUnresolvedDecisionAtPublicBoundary(t *testing.T) {
	root, catalog := validFixture(t)
	path := filepath.Join(root, "specification/decisions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), `"status":"resolved"`, `"status":"unresolved"`, 1))
	refreshDecisionArtifacts(t, root)

	report, err := Check(root, catalog)
	if err == nil || report.Unresolved != 1 || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("Check() = %#v, %v", report, err)
	}
	if onlineReport, onlineErr := CheckOnline(context.Background(), root, catalog, nil); onlineErr == nil || onlineReport.Unresolved != 1 || !strings.Contains(onlineErr.Error(), "unresolved") {
		t.Fatalf("CheckOnline() = %#v, %v", onlineReport, onlineErr)
	}
}

func TestCheckAcceptsHyphenatedDecisionOwner(t *testing.T) {
	root, catalog := validFixture(t)
	for _, relative := range []string{"CHANGELOG.md", "docs/specification-decisions.md", "specification/README.md", "specification/decisions.json", "specification/conformance.json"} {
		path := filepath.Join(root, relative)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		write(t, path, strings.ReplaceAll(string(data), "EXAMPLE-DEC-001", "HTTP-SIG-DEC-001"))
	}
	refreshDecisionArtifacts(t, root)

	if _, err := Check(root, catalog); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsCommentedExecutableEvidence(t *testing.T) {
	root, catalog := validFixture(t)
	write(t, filepath.Join(root, "contract_test.go"), "package example\n\n// func TestRequestTargetContract(t *testing.T) {}\n")

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "executable evidence") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsDuplicateAndOrphanMatrixDecisions(t *testing.T) {
	for name, matrix := range map[string]string{
		"duplicate": "# Matrix\n\n| Decision | Evidence |\n| --- | --- |\n| EXAMPLE-DEC-001 | `TestRequestTargetContract` |\n| EXAMPLE-DEC-001 | `TestRequestTargetContract` |\n",
		"orphan":    "# Matrix\n\n| Decision | Evidence |\n| --- | --- |\n| EXAMPLE-DEC-001 | `TestRequestTargetContract` |\n| EXAMPLE-DEC-999 | `TestRequestTargetContract` |\n",
	} {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			write(t, filepath.Join(root, "specification/README.md"), matrix)
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsMalformedSourceManifest(t *testing.T) {
	root, catalog := validFixture(t)
	write(t, filepath.Join(root, "specification/manifest.tsv"), "not a source manifest\n")

	if _, err := Check(root, catalog); err == nil {
		t.Fatal("Check() error = nil")
	}
}

func TestCheckRequiresFreshAuthorityMonitoring(t *testing.T) {
	root, catalog := validFixture(t)
	if err := os.Remove(filepath.Join(root, "specification/monitoring.json")); err != nil {
		t.Fatal(err)
	}

	if _, err := Check(root, catalog); err == nil {
		t.Fatal("Check() error = nil")
	}
}

func TestCheckRequiresSourceAndChangeMonitoringForEveryDeclaredSpecification(t *testing.T) {
	root, catalog := validFixture(t)
	path := filepath.Join(root, "specification/monitoring.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), `"kind":"specification"`, `"kind":"registry"`, 1))
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), `"kind":"registry"`, `"kind":"errata"`, 1))

	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "source and change authorities") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckOnlineDetectsChangedAuthority(t *testing.T) {
	root, catalog := validFixture(t)
	body := "current errata"
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(body))
	}))
	defer server.Close()
	digest := sha256.Sum256([]byte(body))
	write(t, filepath.Join(root, "specification/monitoring.json"), `{"schema_version":1,"reviewed_at":"`+
		time.Now().UTC().Format("2006-01-02")+
		`","review_interval_days":90,"authorities":[{"id":"rfc9110-source","kind":"specification","version":"RFC 9110","url":"`+
		server.URL+`","sha256":"`+hex.EncodeToString(digest[:])+`","specifications":["RFC 9110"]},{"id":"errata","kind":"errata","url":"`+
		server.URL+`","sha256":"`+hex.EncodeToString(digest[:])+`","specifications":["RFC 9110"]}]}`)
	refreshModuleDecisionArtifacts(t, root, "", server.URL)
	serverClient := server.Client()
	baseTransport, ok := serverClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("server transport = %T", serverClient.Transport)
	}
	transport := baseTransport.Clone()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	serverClient.Transport = transport

	if _, err := checkOnline(context.Background(), root, catalog, serverClient, publicTestAuthorityResolver); err != nil {
		t.Fatalf("CheckOnline() error = %v", err)
	}
	body = "new errata"
	if _, err := checkOnline(context.Background(), root, catalog, serverClient, publicTestAuthorityResolver); err == nil {
		t.Fatal("CheckOnline(changed authority) error = nil")
	}
}

func TestCheckOnlineProbesRestrictedNormativeContent(t *testing.T) {
	root, catalog := validFixture(t)
	body := "current release metadata"
	var restrictedRequests int
	restrictedStatus := http.StatusForbidden
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/restricted" {
			restrictedRequests++
			http.Error(writer, "licensed publication", restrictedStatus)
			return
		}
		_, _ = writer.Write([]byte(body))
	}))
	defer server.Close()
	digest := sha256.Sum256([]byte(body))
	restrictedURL := server.URL + "/restricted"
	write(t, filepath.Join(root, "specification/monitoring.json"), `{"schema_version":1,"reviewed_at":"`+
		time.Now().UTC().Format("2006-01-02")+
		`","review_interval_days":90,"authorities":[{"id":"rfc9110-source","kind":"specification","version":"RFC 9110","url":"`+
		restrictedURL+`","access":"restricted","expected_status":403,"unavailable_reason":"The licensed normative publication is not publicly retrievable.","specifications":["RFC 9110"]},{"id":"rfc9110-releases","kind":"releases","url":"`+
		server.URL+`/releases","sha256":"`+hex.EncodeToString(digest[:])+`","specifications":["RFC 9110"]}]}`)
	refreshModuleDecisionArtifacts(t, root, "", restrictedURL)
	serverClient := server.Client()
	baseTransport, ok := serverClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("server transport = %T", serverClient.Transport)
	}
	transport := baseTransport.Clone()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	serverClient.Transport = transport

	if _, err := checkOnline(context.Background(), root, catalog, serverClient, publicTestAuthorityResolver); err != nil {
		t.Fatalf("CheckOnline(restricted source) error = %v", err)
	}
	if restrictedRequests != 1 {
		t.Fatalf("restricted authority requests = %d, want 1", restrictedRequests)
	}
	restrictedStatus = http.StatusUnauthorized
	if _, err := checkOnline(context.Background(), root, catalog, serverClient, publicTestAuthorityResolver); err == nil || !strings.Contains(err.Error(), "want 403") {
		t.Fatalf("CheckOnline(changed restricted status) error = %v", err)
	}
}

func TestFetchAuthorityRejectsRedirectOutsidePinnedHTTPSAuthority(t *testing.T) {
	body := "redirected authority"
	destination := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(body))
	}))
	defer destination.Close()
	source := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, destination.URL, http.StatusFound)
	}))
	defer source.Close()
	digest := sha256.Sum256([]byte(body))

	err := fetchTestAuthority(context.Background(), source.Client(), authority{ID: "source", URL: source.URL, SHA256: hex.EncodeToString(digest[:])})
	if err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("fetchTestAuthority() error = %v", err)
	}
}

func TestFetchAuthorityAllowsRedirectWithinPinnedHTTPSAuthority(t *testing.T) {
	body := "redirected authority"
	server := httptest.NewTLSServer(nil)
	server.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/source" {
			http.Redirect(writer, request, "/final", http.StatusFound)
			return
		}
		_, _ = writer.Write([]byte(body))
	})
	defer server.Close()
	digest := sha256.Sum256([]byte(body))
	if err := fetchTestAuthority(context.Background(), server.Client(), authority{ID: "source", URL: server.URL + "/source", SHA256: hex.EncodeToString(digest[:])}); err != nil {
		t.Fatalf("fetchTestAuthority() error = %v", err)
	}
}

func TestFetchAuthorityBoundsSameAuthorityRedirects(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/loop", http.StatusFound)
	}))
	defer server.Close()
	if err := fetchTestAuthority(context.Background(), server.Client(), authority{ID: "source", URL: server.URL + "/loop", SHA256: strings.Repeat("0", 64)}); err == nil || !strings.Contains(err.Error(), "10 redirects") {
		t.Fatalf("fetchTestAuthority() error = %v", err)
	}
}

func TestCheckHandlesDiscoveryAndRepositoryFailures(t *testing.T) {
	root, catalog := validFixture(t)
	noSpecifications := catalog
	noSpecifications.Modules = append([]inventory.Module(nil), catalog.Modules...)
	noSpecifications.Modules[0].Specifications = nil
	if _, err := Check(root, noSpecifications); err == nil || !strings.Contains(err.Error(), "does not declare specifications") {
		t.Fatalf("Check(no specification declaration) error = %v", err)
	}
	if _, err := CheckOnline(context.Background(), root, noSpecifications, nil); err == nil || !strings.Contains(err.Error(), "does not declare specifications") {
		t.Fatalf("CheckOnline(no specification declaration) error = %v", err)
	}
	artifactsWithoutDeclaration := noSpecifications
	artifactsWithoutDeclaration.Modules = append([]inventory.Module(nil), noSpecifications.Modules...)
	artifactsWithoutDeclaration.Modules[0].Provenance = nil
	if _, err := Check(root, artifactsWithoutDeclaration); err == nil || !strings.Contains(err.Error(), "specification artifacts") {
		t.Fatalf("Check(undeclared artifacts) error = %v", err)
	}
	invalidDiscovery := inventory.Inventory{Modules: []inventory.Module{{Directory: strings.Repeat("x", 4096)}}}
	if _, err := Check(t.TempDir(), invalidDiscovery); err == nil || !strings.Contains(err.Error(), "specification discovery") {
		t.Fatalf("Check(invalid discovery path) error = %v", err)
	}
	emptyRoot := t.TempDir()
	if report, err := Check(emptyRoot, inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}}); err != nil || report != (Report{}) {
		t.Fatalf("Check(non-specification module) = %#v, %v", report, err)
	}
	if _, err := CheckOnline(context.Background(), root, catalog, nil); err == nil {
		t.Fatal("CheckOnline(nil client) error = nil")
	}
	customTransport := roundTripper(func(*http.Request) (*http.Response, error) { return nil, nil })
	if _, err := checkOnline(context.Background(), root, catalog, &http.Client{Transport: customTransport}, publicTestAuthorityResolver); err == nil || !strings.Contains(err.Error(), "pinned destination dialing") {
		t.Fatalf("checkOnline(custom transport) error = %v", err)
	}

	second := catalog.Modules[0]
	second.Directory = "nested"
	second.ModulePath = "example.com/nested"
	second.Provenance = json.RawMessage(`["nested/specification/manifest.tsv"]`)
	for _, relative := range []string{"README.md", "COMPATIBILITY.md", "CHANGELOG.md", "docs/specification-decisions.md", "specification/README.md", "specification/manifest.tsv", "specification/monitoring.json", "specification/decisions.json", "specification/conformance.json", "testdata/request-target.json", "testdata/peer.tsv"} {
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatal(err)
		}
		content := strings.ReplaceAll(string(data), "EXAMPLE-DEC-001", "NESTED-DEC-001")
		content = strings.ReplaceAll(content, `"docs/specification-decisions.md"`, `"nested/docs/specification-decisions.md"`)
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
	catalog.Modules = append(catalog.Modules, second)
	if report, err := Check(root, catalog); err != nil || report.Modules != 2 {
		t.Fatalf("Check(two modules) = %#v, %v", report, err)
	}
	catalog.Modules[1].Provenance = json.RawMessage(`["specification/manifest.tsv"]`)
	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "outside module") {
		t.Fatalf("Check(cross-module provenance) error = %v", err)
	}

	mutations := map[string]func(*testing.T, string){
		"missing register": func(t *testing.T, root string) { removeFile(t, filepath.Join(root, "docs/specification-decisions.md")) },
		"missing matrix":   func(t *testing.T, root string) { removeFile(t, filepath.Join(root, "specification/README.md")) },
		"no decisions": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "docs/specification-decisions.md"), "# Decisions\n")
		},
		"decision missing from matrix": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "specification/README.md"), "# Matrix\n\nOTHER-DEC-001\n")
		},
		"duplicate decision": func(t *testing.T, root string) {
			data, err := os.ReadFile(filepath.Join(root, "docs/specification-decisions.md"))
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(root, "docs/specification-decisions.md"), string(data)+"\n"+strings.TrimPrefix(string(data), "# Decisions\n"))
		},
		"invalid entry": func(t *testing.T, root string) {
			path := filepath.Join(root, "specification/decisions.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			write(t, path, strings.Replace(string(data), `"owner":"example maintainers"`, `"owner":""`, 1))
		},
		"evidence walk": func(t *testing.T, root string) {
			if err := os.Symlink("missing", filepath.Join(root, "evidence-link")); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			mutate(t, root)
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}

	root, catalog = validFixture(t)
	path := filepath.Join(root, "specification/decisions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), `"status":"resolved"`, `"status":"unresolved"`, 1))
	refreshDecisionArtifacts(t, root)
	if report, err := Check(root, catalog); err == nil || report.Unresolved != 1 || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("Check(unresolved) = %#v, %v", report, err)
	}
}

func TestCheckRejectsInvalidProvenanceDeclarations(t *testing.T) {
	tests := map[string]json.RawMessage{
		"missing":      nil,
		"null":         json.RawMessage(`null`),
		"malformed":    json.RawMessage(`{`),
		"empty":        json.RawMessage(`[]`),
		"missing file": json.RawMessage(`["specification/missing.tsv"]`),
		"unsupported":  json.RawMessage(`["specification/manifest.yaml"]`),
	}
	for name, provenance := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := validFixture(t)
			catalog.Modules[0].Provenance = provenance
			if name == "unsupported" {
				write(t, filepath.Join(root, "specification/manifest.yaml"), "sources: []\n")
			}
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func validFixture(t *testing.T) (string, inventory.Inventory) {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "docs/specification-decisions.md"), "# Decisions\n\n"+
		"## EXAMPLE-DEC-001: Request target normalization\n\n"+
		"- **Status, owner, and classification:** `resolved`; maintainers; defensive interoperability policy.\n"+
		"- **Source and issue:** [RFC 9110 section 7.1](https://www.rfc-editor.org/rfc/rfc9110.html#section-7.1), exact version RFC 9110. The specification omits an application normalization policy.\n"+
		"- **Interpretations and peer behavior:** Preserve exact bytes or normalize equivalent paths. Maintained peers use both policies.\n"+
		"- **Selected behavior, security and resource consequences, compatibility and wire consequences:** Preserve exact bytes because normalization can change signatures. Input is bounded; changing this alters wire compatibility.\n"+
		"- **Evidence, public surface, upstream, and reconsideration:** `TestRequestTargetContract` covers `Parse` and the public compatibility guide. No upstream issue exists. Reconsider when RFC 9110 is superseded.\n")
	write(t, filepath.Join(root, "README.md"), "# Example\n\n[Specification decisions](docs/specification-decisions.md)\n")
	write(t, filepath.Join(root, "COMPATIBILITY.md"), "# Compatibility\n\n[Specification decisions](docs/specification-decisions.md)\n")
	write(t, filepath.Join(root, "CONTRIBUTING.md"), "# Contributing\n\n[Root specification decisions](docs/specification-decisions.md)\n\n[Nested specification decisions](nested/docs/specification-decisions.md)\n")
	write(t, filepath.Join(root, "CHANGELOG.md"), "# Changelog\n\n## Specification Decisions\n\n- EXAMPLE-DEC-001: Request target normalization.\n\n[Decision register](docs/specification-decisions.md)\n")
	write(t, filepath.Join(root, ".github/pull_request_template.md"), "## Specification Decisions\n\n- Decision identifier\n- Compatibility impact\n- Changelog entry\n- Superseded identifiers remain in the register\n")
	write(t, filepath.Join(root, "specification/README.md"), "# Specification conformance matrix\n\n[Specification decisions](../docs/specification-decisions.md)\n\n| Source | Decisions | Evidence |\n| --- | --- | --- |\n| RFC 9110 | EXAMPLE-DEC-001 | `TestRequestTargetContract` |\n")
	write(t, filepath.Join(root, "specification/manifest.tsv"), "id\tversion\tsections\trole\turl\tsha256\tstatus\n"+
		"rfc9110\tRFC-9110\t7.1\tHTTP-semantics\thttps://www.rfc-editor.org/rfc/rfc9110.txt\t"+
		"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a\tpinned\n")
	write(t, filepath.Join(root, "specification/monitoring.json"), `{"schema_version":1,"reviewed_at":"`+
		time.Now().UTC().Format("2006-01-02")+
		`","review_interval_days":90,"authorities":[{"id":"rfc9110-source","kind":"specification","version":"RFC 9110","url":"https://www.rfc-editor.org/rfc/rfc9110.txt","sha256":"`+
		"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a"+
		`","specifications":["RFC 9110"]},{"id":"rfc9110-errata","kind":"errata","url":"https://www.rfc-editor.org/errata/rfc9110","sha256":"`+
		"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a"+
		`","specifications":["RFC 9110"]}]}`)
	write(t, filepath.Join(root, "specification/decisions.json"), `{"schema_version":1,"change_control":{"readme":"README.md","conformance":"specification/README.md","compatibility":"COMPATIBILITY.md","contribution":"CONTRIBUTING.md","changelog":"CHANGELOG.md","pull_request_template":".github/pull_request_template.md"},"decisions":[{"id":"EXAMPLE-DEC-001","title":"Request target normalization","status":"resolved","owner":"example maintainers","classification":"omission","decision_scope":"defensive","specification":"RFC 9110","version":"RFC 9110","source_authority":"rfc9110-source","section":"7.1","requirement_strength":"not specified","issue":"The specification omits application normalization policy.","interpretations":["Preserve bytes","Normalize equivalent paths"],"peer_behavior":"Maintained peers use both policies.","selected_behavior":"Preserve exact bytes.","rationale":"Normalization can change signatures.","security_consequences":"Prevents signature confusion.","resource_consequences":"Work remains bounded.","compatibility_consequences":"Existing callers retain byte-exact behavior.","wire_consequences":"Wire bytes are preserved.","executable_evidence":["TestRequestTargetContract"],"fixture_evidence":["testdata/request-target.json"],"fuzz_evidence":["FuzzRequestTargetContract"],"interoperability_evidence":["testdata/peer.tsv"],"public_apis":["Parse"],"documentation":["docs/specification-decisions.md"],"upstream_status":"No upstream issue exists.","reconsider_when":"RFC 9110 is superseded."}]}`)
	refreshDecisionArtifacts(t, root)
	write(t, filepath.Join(root, "specification/conformance.json"), `{"schema_version":1,"decisions":[{"id":"EXAMPLE-DEC-001","authoritative_sources":["rfc9110-source"],"executable_evidence":["TestRequestTargetContract"],"fixtures":["testdata/request-target.json"],"fuzz":["FuzzRequestTargetContract"],"differential_evidence":["testdata/peer.tsv"],"differential_classification":"deliberate policy difference","public_behavior":["Parse preserves request-target bytes."]}]}`)
	write(t, filepath.Join(root, "testdata/request-target.json"), "{}\n")
	write(t, filepath.Join(root, "testdata/peer.tsv"), "peer\tbehavior\nexample\tpreserve\n")
	write(t, filepath.Join(root, "contract_test.go"), "package example\n\nimport \"testing\"\n\nfunc TestRequestTargetContract(t *testing.T) {}\nfunc FuzzRequestTargetContract(f *testing.F) {}\n")

	catalog := inventory.Inventory{Modules: []inventory.Module{{
		Directory:      ".",
		Specifications: []string{"RFC 9110"},
		Provenance:     json.RawMessage(`["specification/manifest.tsv"]`),
	}}}
	return root, catalog
}

func refreshDecisionArtifacts(t *testing.T, root string) {
	refreshModuleDecisionArtifacts(t, root, "", "https://www.rfc-editor.org/rfc/rfc9110.txt")
}

func refreshModuleDecisionArtifacts(t *testing.T, root, prefix, authorityURL string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, prefix, "specification/decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := loadDecisions(data)
	if err != nil {
		t.Fatal(err)
	}
	registerPath := filepath.Join(root, prefix, "docs/specification-decisions.md")
	register, err := os.ReadFile(registerPath)
	if err != nil {
		t.Fatal(err)
	}
	var content strings.Builder
	content.Write(register)
	for _, item := range manifest.Decisions {
		encoded, _ := json.Marshal(item)
		_, _ = fmt.Fprintf(&content, "\nDecision data: `%s`\n", encoded)
	}
	content.WriteString("Authoritative URL: " + authorityURL + "\n")
	write(t, registerPath, content.String())
	var lines strings.Builder
	lines.WriteString("# Changelog\n\n## Specification Decisions\n\n")
	history := decisionHistory{SchemaVersion: 1}
	for _, item := range manifest.Decisions {
		_, _ = fmt.Fprintf(&lines, "- %s sha256:%s\n", item.ID, decisionDigest(item))
		history.Entries = append(history.Entries, decisionHistoryEntry{ID: item.ID, Digests: []string{decisionDigest(item)}})
	}
	write(t, filepath.Join(root, prefix, "CHANGELOG.md"), lines.String()+"\n[Decision register](docs/specification-decisions.md)\n")
	writeJSON(t, filepath.Join(root, prefix, "specification/decision-history.json"), history)
}

func write(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, string(data)+"\n")
}

func mutateDecisionManifest(t *testing.T, root string, mutate func(*decisionManifest)) {
	t.Helper()
	path := filepath.Join(root, "specification/decisions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := loadDecisions(data)
	if err != nil {
		t.Fatal(err)
	}
	mutate(&manifest)
	writeJSON(t, path, manifest)
}

func removeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
