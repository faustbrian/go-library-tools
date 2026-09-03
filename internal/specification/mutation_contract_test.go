package specification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"net/http"
	"net/netip"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

type staticAuthorityResolver struct {
	addresses []netip.Addr
	err       error
}

type sequenceAuthorityResolver struct {
	addresses [][]netip.Addr
	calls     int
}

func (resolver *sequenceAuthorityResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	index := resolver.calls
	resolver.calls++
	return resolver.addresses[index], nil
}

func (resolver staticAuthorityResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return resolver.addresses, resolver.err
}

var publicTestAuthorityResolver = staticAuthorityResolver{addresses: []netip.Addr{netip.MustParseAddr("8.8.8.8")}}

func fetchTestAuthority(ctx context.Context, client *http.Client, item authority) error {
	return fetchAuthorityResolved(ctx, client, publicTestAuthorityResolver, item)
}

func TestAuthorityFetchBoundaryContract(t *testing.T) {
	body := strings.Repeat("x", maximumAuthoritySize)
	digest := sha256.Sum256([]byte(body))
	item := authority{ID: "source", URL: "https://example.com/source", SHA256: hex.EncodeToString(digest[:])}
	if err := fetchTestAuthority(context.Background(), client(response(http.StatusOK, body, nil, nil), nil), item); err != nil {
		t.Fatalf("fetchTestAuthority(maximum body) error = %v", err)
	}

	for _, status := range []int{http.StatusContinue, http.StatusMultipleChoices} {
		err := fetchTestAuthority(context.Background(), client(response(status, body, nil, nil), nil), item)
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", status)) {
			t.Fatalf("fetchTestAuthority(status %d) error = %v", status, err)
		}
	}
	for _, status := range []int{http.StatusOK, http.StatusMultipleChoices - 1} {
		if err := fetchTestAuthority(context.Background(), client(response(status, body, nil, nil), nil), item); err != nil {
			t.Fatalf("fetchTestAuthority(status %d) error = %v", status, err)
		}
	}
}

func TestAuthorityFetchIdentifiesTheSpecificationMonitor(t *testing.T) {
	body := "authority"
	digest := sha256.Sum256([]byte(body))
	var userAgent string
	transport := roundTripper(func(request *http.Request) (*http.Response, error) {
		userAgent = request.Header.Get("User-Agent")
		return response(http.StatusOK, body, nil, nil), nil
	})
	err := fetchTestAuthority(context.Background(), &http.Client{Transport: transport}, authority{
		ID: "source", URL: "https://example.com/source", SHA256: hex.EncodeToString(digest[:]),
	})
	if err != nil {
		t.Fatalf("fetchTestAuthority() error = %v", err)
	}
	if userAgent != "go-library-tools (+https://github.com/faustbrian/go-library-tools)" {
		t.Fatalf("specification authority User-Agent = %q", userAgent)
	}
}

func TestAuthorityFetchRejectsPrivateDestinationsBeforeTransport(t *testing.T) {
	called := false
	transport := roundTripper(func(_ *http.Request) (*http.Response, error) {
		called = true
		return response(http.StatusOK, "private", nil, nil), nil
	})
	digest := sha256.Sum256([]byte("private"))
	err := fetchAuthority(context.Background(), &http.Client{Transport: transport}, authority{
		ID: "source", URL: "https://127.0.0.1/source", SHA256: hex.EncodeToString(digest[:]),
	})
	if err == nil || !strings.Contains(err.Error(), "public network destination") {
		t.Fatalf("fetchTestAuthority(private destination) error = %v", err)
	}
	if called {
		t.Fatal("fetchTestAuthority(private destination) called the transport")
	}
}

func TestAuthorityRedirectBoundaryContract(t *testing.T) {
	body := "authority"
	digest := sha256.Sum256([]byte(body))
	for redirects := 9; redirects <= 10; redirects++ {
		calls := 0
		transport := roundTripper(func(request *http.Request) (*http.Response, error) {
			calls++
			if calls <= redirects {
				return &http.Response{
					StatusCode: http.StatusFound,
					Header:     http.Header{"Location": []string{fmt.Sprintf("/redirect/%d", calls)}},
					Body:       &faultBody{Reader: strings.NewReader("")},
					Request:    request,
				}, nil
			}
			return response(http.StatusOK, body, nil, nil), nil
		})
		err := fetchTestAuthority(context.Background(), &http.Client{Transport: transport}, authority{ID: "source", URL: "https://example.com/source", SHA256: hex.EncodeToString(digest[:])})
		if redirects == 9 && err != nil {
			t.Fatalf("fetchTestAuthority(nine redirects) error = %v", err)
		}
		if redirects == 10 && (err == nil || !strings.Contains(err.Error(), "stopped after 10 redirects")) {
			t.Fatalf("fetchTestAuthority(ten redirects) error = %v", err)
		}
	}
	for _, location := range []string{"http://example.com/final", "https://other.example/final"} {
		transport := roundTripper(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{location}},
				Body:       &faultBody{Reader: strings.NewReader("")},
				Request:    request,
			}, nil
		})
		err := fetchTestAuthority(context.Background(), &http.Client{Transport: transport}, authority{ID: "source", URL: "https://example.com/source", SHA256: hex.EncodeToString(digest[:])})
		if err == nil || !strings.Contains(err.Error(), "redirect escaped its pinned HTTPS authority") {
			t.Fatalf("fetchTestAuthority(redirect %q) error = %v", location, err)
		}
	}
}

func TestAuthorityRedirectRevalidatesPublicResolution(t *testing.T) {
	body := "authority"
	requests := 0
	transport := roundTripper(func(request *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{"/final"}},
			Body:       &faultBody{Reader: strings.NewReader("")},
			Request:    request,
		}, nil
	})
	resolver := &sequenceAuthorityResolver{addresses: [][]netip.Addr{
		{netip.MustParseAddr("8.8.8.8")},
		{netip.MustParseAddr("10.0.0.1")},
	}}
	digest := sha256.Sum256([]byte(body))
	err := fetchAuthorityResolved(context.Background(), &http.Client{Transport: transport}, resolver, authority{
		ID: "source", URL: "https://example.com/source", SHA256: hex.EncodeToString(digest[:]),
	})
	if err == nil || !strings.Contains(err.Error(), "public network destination") || resolver.calls != 2 || requests != 1 {
		t.Fatalf("fetchAuthorityResolved(redirect resolution) = %v, resolver calls %d, requests %d", err, resolver.calls, requests)
	}
}

func TestAuthorityFetchRejectsRebindingBeforeDial(t *testing.T) {
	resolver := &sequenceAuthorityResolver{addresses: [][]netip.Addr{
		{netip.MustParseAddr("8.8.8.8")},
		{netip.MustParseAddr("10.0.0.1")},
	}}
	dialed := false
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Fatalf("default transport = %T", http.DefaultTransport)
	}
	transport := baseTransport.Clone()
	transport.DialContext = func(context.Context, string, string) (net.Conn, error) {
		dialed = true
		return nil, errors.New("dialed unvalidated destination")
	}
	client, err := secureAuthorityClient(&http.Client{Transport: transport}, resolver)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("authority"))
	err = fetchAuthorityResolved(context.Background(), client, resolver, authority{
		ID: "source", URL: "https://example.com/source", SHA256: hex.EncodeToString(digest[:]),
	})
	if err == nil || !strings.Contains(err.Error(), "public network destination") || resolver.calls != 2 || dialed {
		t.Fatalf("fetchAuthorityResolved(rebinding) = %v, resolver calls %d, dialed %t", err, resolver.calls, dialed)
	}
}

func TestCheckAggregatesMixedModuleReports(t *testing.T) {
	root, catalog := validFixture(t)
	catalog.Modules = append([]inventory.Module{{Directory: "empty"}}, catalog.Modules...)
	catalog.Modules[1].Directory = "nested"
	catalog.Modules[1].ModulePath = "example.com/nested"
	catalog.Modules[1].Provenance = []byte(`["nested/specification/manifest.tsv"]`)
	copyFixtureModule(t, root, "nested")
	report, err := Check(root, catalog)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if report != (Report{Modules: 1, Decisions: 1}) {
		t.Fatalf("Check() report = %#v", report)
	}
}

func TestCheckAggregatesMultipleSpecificationModules(t *testing.T) {
	root, catalog := validFixture(t)
	nested := catalog.Modules[0]
	nested.Directory = "nested"
	nested.ModulePath = "example.com/nested"
	nested.Provenance = []byte(`["nested/specification/manifest.tsv"]`)
	copyFixtureModule(t, root, "nested")
	catalog.Modules = append(catalog.Modules, nested)
	for _, prefix := range []string{"", "nested"} {
		path := filepath.Join(root, prefix, "specification/decisions.json")
		manifestData := readFile(t, path)
		manifest, err := loadDecisions(manifestData)
		if err != nil {
			t.Fatal(err)
		}
		manifest.Decisions[0].Status = "unresolved"
		writeJSON(t, path, manifest)
		refreshModuleDecisionArtifacts(t, root, prefix, "https://www.rfc-editor.org/rfc/rfc9110.txt")
	}
	report, err := Check(root, catalog)
	if err == nil || !strings.Contains(err.Error(), "2 unresolved") {
		t.Fatalf("Check() error = %v", err)
	}
	if report != (Report{Modules: 2, Decisions: 2, Unresolved: 2}) {
		t.Fatalf("Check() report = %#v", report)
	}
}

func TestCheckAcceptsMultipleDecisionsInOneRegister(t *testing.T) {
	root, catalog := validFixture(t)
	addSecondDecision(t, root)
	report, err := Check(root, catalog)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if report != (Report{Modules: 1, Decisions: 2}) {
		t.Fatalf("Check() report = %#v", report)
	}
}

func TestCheckAcceptsEveryEmptyProvenanceEncodingWithoutSpecifications(t *testing.T) {
	for _, provenance := range []string{"", "null", "[]"} {
		report, err := Check(t.TempDir(), inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Provenance: []byte(provenance)}}})
		if err != nil || report != (Report{}) {
			t.Fatalf("Check(provenance %q) = %#v, %v", provenance, report, err)
		}
	}
}

func TestCheckOnlineReturnsImmediatelyWithoutSpecificationModules(t *testing.T) {
	report, err := CheckOnline(context.Background(), t.TempDir(), inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}}, nil)
	if err != nil || report != (Report{}) {
		t.Fatalf("CheckOnline(no specification modules) = %#v, %v", report, err)
	}
}

func TestStrictJSONDistinguishesTrailingContracts(t *testing.T) {
	for input, want := range map[string]string{
		`{} {}`: "multiple JSON values",
		`{} x`:  "decode trailing data:",
	} {
		var target map[string]any
		err := decodeStrictJSON([]byte(input), &target)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("decodeStrictJSON(%q) error = %v", input, err)
		}
	}
}

func TestCheckRejectsEveryInvalidConformanceIdentifierShape(t *testing.T) {
	for _, identifier := range []string{"invalid", "EXAMPLE-DEC-001-suffix"} {
		root, catalog := validFixture(t)
		path := filepath.Join(root, "specification/conformance.json")
		data := readFile(t, path)
		manifest, err := loadConformance(data)
		if err != nil {
			t.Fatal(err)
		}
		manifest.Decisions[0].ID = identifier
		writeJSON(t, path, manifest)
		if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "conformance decision has invalid identifier") {
			t.Fatalf("Check(conformance %q) error = %v", identifier, err)
		}
	}
}

func TestDecisionContractRejectsUndocumentedAdditionalAuthoritativeSource(t *testing.T) {
	root, catalog := validFixture(t)
	decisionManifest, err := loadDecisions(readFile(t, filepath.Join(root, "specification/decisions.json")))
	if err != nil {
		t.Fatal(err)
	}
	conformanceManifest, err := loadConformance(readFile(t, filepath.Join(root, "specification/conformance.json")))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := loadMonitoring(root, "specification/monitoring.json", catalog.Modules[0].Specifications, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	additional := authority{ID: "extension-source", Kind: "extension", Version: "Extension 1", URL: "https://example.com/extension", Specifications: []string{"Example Extension"}}
	policy.Authorities = append(policy.Authorities, additional)
	conformanceManifest.Decisions[0].AuthoritativeSources = append(conformanceManifest.Decisions[0].AuthoritativeSources, additional.ID)
	matrix, err := matrixDecisions(readFile(t, filepath.Join(root, "specification/README.md")))
	if err != nil {
		t.Fatal(err)
	}
	evidence := map[string]struct{}{"TestRequestTargetContract": {}, "FuzzRequestTargetContract": {}}
	_, _, err = validateDecisionContract(root, catalog.Modules[0], catalog.Modules, policy, decisionManifest, conformanceManifest, readFile(t, filepath.Join(root, "docs/specification-decisions.md")), matrix, evidence)
	if err == nil || !strings.Contains(err.Error(), "omits additional authoritative source") {
		t.Fatalf("validateDecisionContract() error = %v", err)
	}
}

func TestDecisionContractRejectsAdditionalAuthorityDocumentedOutsideDecision(t *testing.T) {
	root, catalog := validFixture(t)
	decisionManifest, err := loadDecisions(readFile(t, filepath.Join(root, "specification/decisions.json")))
	if err != nil {
		t.Fatal(err)
	}
	conformanceManifest, err := loadConformance(readFile(t, filepath.Join(root, "specification/conformance.json")))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := loadMonitoring(root, "specification/monitoring.json", catalog.Modules[0].Specifications, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	additional := authority{ID: "extension-source", Kind: "extension", Version: "Extension 1", URL: "https://example.com/extension", Specifications: []string{"Example Extension"}}
	policy.Authorities = append(policy.Authorities, additional)
	conformanceManifest.Decisions[0].AuthoritativeSources = append(conformanceManifest.Decisions[0].AuthoritativeSources, additional.ID)
	register := append(readFile(t, filepath.Join(root, "docs/specification-decisions.md")), []byte("\n## Appendix\n\nextension-source\nExtension 1\nhttps://example.com/extension\nExample Extension\n")...)
	matrix, err := matrixDecisions(readFile(t, filepath.Join(root, "specification/README.md")))
	if err != nil {
		t.Fatal(err)
	}
	evidence := map[string]struct{}{"TestRequestTargetContract": {}, "FuzzRequestTargetContract": {}}
	_, _, err = validateDecisionContract(root, catalog.Modules[0], catalog.Modules, policy, decisionManifest, conformanceManifest, register, matrix, evidence)
	if err == nil || !strings.Contains(err.Error(), "omits additional authoritative source") {
		t.Fatalf("validateDecisionContract(appendix authority) error = %v", err)
	}
}

func TestOptionalEvidenceAndDocumentationFailureBoundaries(t *testing.T) {
	root, catalog, item, authorities, evidence := structuredFixture(t)
	item.FixtureEvidence = []string{" "}
	if err := validateDecision(root, catalog.Modules[0], catalog.Modules, item, authorities, evidence); err == nil || !strings.Contains(err.Error(), "blank fixture_evidence") {
		t.Fatalf("validateDecision(blank optional evidence) error = %v", err)
	}
	_, _, item, authorities, evidence = structuredFixture(t)
	item.Documentation = []string{"docs/missing.md"}
	if err := validateDecision(root, catalog.Modules[0], catalog.Modules, item, authorities, evidence); err == nil || !strings.Contains(err.Error(), "docs/missing.md") {
		t.Fatalf("validateDecision(missing documentation) error = %v", err)
	}
	fixtureRoot, _ := validFixture(t)
	manifest, err := loadConformance(readFile(t, filepath.Join(fixtureRoot, "specification/conformance.json")))
	if err != nil {
		t.Fatal(err)
	}
	row := manifest.Decisions[0]
	row.Fixtures = []string{" "}
	if err := validateConformance(item, row, authorities, evidence); err == nil || !strings.Contains(err.Error(), "blank conformance fixtures") {
		t.Fatalf("validateConformance(blank optional evidence) error = %v", err)
	}
}

func TestDecisionLedgerConditionBoundaries(t *testing.T) {
	item := decision{ID: "EXAMPLE-DEC-001", Title: "Example"}
	digest := decisionDigest(item)
	tests := []struct {
		name    string
		history decisionHistory
		want    string
	}{
		{"empty identifier", decisionHistory{Entries: []decisionHistoryEntry{{Digests: []string{digest}}}}, "incomplete entry"},
		{"empty digests", decisionHistory{Entries: []decisionHistoryEntry{{ID: item.ID}}}, "incomplete entry"},
		{"missing current", decisionHistory{}, "does not contain current"},
		{"wrong current", decisionHistory{Entries: []decisionHistoryEntry{{ID: item.ID, Digests: []string{strings.Repeat("0", 64)}}}}, "does not contain current"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDecisionLedger(test.history, []decision{item})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validateDecisionLedger() error = %v", err)
			}
		})
	}
}

func TestChangelogScansEveryDecisionIdentifier(t *testing.T) {
	root := t.TempDir()
	item := decision{ID: "EXAMPLE-DEC-001", Title: "Example"}
	write(t, filepath.Join(root, "CHANGELOG.md"), item.ID+" sha256:"+decisionDigest(item)+"\nEXAMPLE-DEC-999\n")
	err := validateDecisionHistory(root, "CHANGELOG.md", []decision{item})
	if err == nil || !strings.Contains(err.Error(), `preserves decision "EXAMPLE-DEC-999"`) {
		t.Fatalf("validateDecisionHistory() error = %v", err)
	}
}

func TestSupersededDecisionCannotReplaceItself(t *testing.T) {
	root, catalog := validFixture(t)
	mutateDecisionManifest(t, root, func(manifest *decisionManifest) {
		manifest.Decisions[0].Status = "superseded"
		manifest.Decisions[0].Replacement = manifest.Decisions[0].ID
	})
	refreshDecisionArtifacts(t, root)
	if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "has no replacement") {
		t.Fatalf("Check(self replacement) error = %v", err)
	}
}

func TestChangeControlTemplateRequiresEveryReviewPrompt(t *testing.T) {
	required := []string{"## specification decisions", "decision identifier", "compatibility", "changelog", "superseded"}
	for _, omitted := range required {
		t.Run(omitted, func(t *testing.T) {
			root, catalog := validFixture(t)
			parts := make([]string, 0, len(required)-1)
			for _, value := range required {
				if value != omitted {
					parts = append(parts, value)
				}
			}
			write(t, filepath.Join(root, ".github/pull_request_template.md"), strings.Join(parts, "\n"))
			if _, err := Check(root, catalog); err == nil || !strings.Contains(err.Error(), "required decision review section") {
				t.Fatalf("Check(template without %q) error = %v", omitted, err)
			}
		})
	}
}

func TestMarkdownLinkBoundaryContract(t *testing.T) {
	for name, data := range map[string]string{
		"malformed":    `[bad](%)`,
		"scheme":       `[bad](https://example.com/docs/specification-decisions.md)`,
		"host":         `[bad](//example.com/docs/specification-decisions.md)`,
		"empty path":   `[bad](#fragment)`,
		"wrong target": `[bad](other.md)`,
	} {
		t.Run(name, func(t *testing.T) {
			if documentLinksTo("README.md", []byte(data), "docs/specification-decisions.md") {
				t.Fatalf("documentLinksTo(%q) = true", data)
			}
		})
	}
	data := []byte(`[bad](https://example.com/other) [good](docs/specification-decisions.md)`)
	if !documentLinksTo("README.md", data, "docs/specification-decisions.md") {
		t.Fatal("documentLinksTo() did not continue to a valid link")
	}
	if documentLinksToDecision([]byte(`[bad](%)`), "EXAMPLE-DEC-001", "Replacement") {
		t.Fatal("documentLinksToDecision(malformed) = true")
	}
	if !documentLinksToDecision([]byte(`[bad](#other) [good](#example-dec-001-replacement)`), "EXAMPLE-DEC-001", "Replacement") {
		t.Fatal("documentLinksToDecision() did not continue to a valid link")
	}
	if documentLinksToDecision([]byte(`[external](https://example.com/#example-dec-001-replacement)`), "EXAMPLE-DEC-001", "Replacement") {
		t.Fatal("documentLinksToDecision() accepted an external replacement link")
	}
	if anchor := decisionHeadingAnchor("EXAMPLE-DEC-001", "Replacement: policy!"); anchor != "example-dec-001-replacement-policy" {
		t.Fatalf("decisionHeadingAnchor() = %q", anchor)
	}
}

func TestDecisionIdentifierAndPolicyConditionBoundaries(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"fixture", "peer", "docs"} {
		write(t, filepath.Join(root, path), path)
	}
	base := decision{
		ID: "EXAMPLE-DEC-001", Title: "Title", Status: "unresolved", Owner: "Owner", Classification: "omission",
		DecisionScope: "defensive", Specification: "RFC 9110", Version: "RFC 9110", SourceAuthority: "source",
		Section: "1", RequirementStrength: "not specified", Issue: "Issue", Interpretations: []string{"One"},
		PeerBehavior: "Peer", SelectedBehavior: "Selected", Rationale: "Rationale", SecurityConsequences: "Security",
		ResourceConsequences: "Resource", CompatibilityConsequences: "Compatibility", WireConsequences: "Wire",
		FixtureEvidence: []string{"fixture"}, FuzzEvidence: []string{"FuzzContract"}, InteroperabilityEvidence: []string{"peer"},
		PublicAPIs: []string{"Parse"}, Documentation: []string{"docs"}, UpstreamStatus: "None", ReconsiderWhen: "Later",
	}
	authorities := map[string]authority{"source": {ID: "source", Kind: "specification", Version: "RFC 9110", Specifications: []string{"RFC 9110"}}}
	evidence := map[string]struct{}{"FuzzContract": {}}
	for _, identifier := range []string{"invalid", "EXAMPLE-DEC-001-suffix"} {
		item := base
		item.ID = identifier
		if err := validateDecision(root, inventory.Module{Directory: "."}, nil, item, authorities, evidence); err == nil || !strings.Contains(err.Error(), "invalid identifier") {
			t.Fatalf("validateDecision(%q) error = %v", identifier, err)
		}
	}
	item := base
	item.Status = "resolved"
	if err := validateDecision(root, inventory.Module{Directory: "."}, nil, item, authorities, evidence); err == nil || !strings.Contains(err.Error(), "no executable evidence") {
		t.Fatalf("validateDecision(resolved without evidence) error = %v", err)
	}
	item = base
	item.DecisionScope = "normative"
	item.RequirementStrength = "MUST"
	if err := validateDecision(root, inventory.Module{Directory: "."}, nil, item, authorities, evidence); err != nil {
		t.Fatalf("validateDecision(normative MUST) error = %v", err)
	}
	item.RequirementStrength = "REQUIRED"
	if err := validateDecision(root, inventory.Module{Directory: "."}, nil, item, authorities, evidence); err != nil {
		t.Fatalf("validateDecision(normative REQUIRED) error = %v", err)
	}
	item.DecisionScope = "defensive"
	if err := validateDecision(root, inventory.Module{Directory: "."}, nil, item, authorities, evidence); err == nil || !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("validateDecision(defensive REQUIRED) error = %v", err)
	}
	authorities["source"] = authority{ID: "source", Kind: "recommendation", Version: "RFC 9110", Specifications: []string{"RFC 9110"}}
	item.DecisionScope = "normative"
	item.RequirementStrength = "SHOULD"
	if err := validateDecision(root, inventory.Module{Directory: "."}, nil, item, authorities, evidence); err == nil || !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("validateDecision(normative recommendation) error = %v", err)
	}
}

func TestMonitoringBoundaryContract(t *testing.T) {
	for _, interval := range []int{1, 90} {
		root := t.TempDir()
		writeMonitoringFixture(t, root, 64, interval)
		if _, err := loadMonitoring(root, "monitoring.json", []string{"RFC 9110"}, time.Date(2026, 8, 30, 23, 59, 0, 0, time.UTC)); err != nil {
			t.Fatalf("loadMonitoring(interval %d) error = %v", interval, err)
		}
	}
	root := t.TempDir()
	writeMonitoringFixture(t, root, 65, 90)
	if _, err := loadMonitoring(root, "monitoring.json", []string{"RFC 9110"}, time.Date(2026, 8, 30, 23, 59, 0, 0, time.UTC)); err == nil || !strings.Contains(err.Error(), "too many authorities") {
		t.Fatalf("loadMonitoring(65 authorities) error = %v", err)
	}
}

func TestSourceManifestAndJSONPinBoundaryContract(t *testing.T) {
	digest := strings.Repeat("a", 64)
	valid := "id\tversion\turl\tsha256\tstatus\nsource\tv1\thttps://example.com/spec\t" + digest + "\tpinned\n"
	invalidSecond := valid + "source2\tv2\thttp://example.com/spec\t" + digest + "\tpinned\n"
	if err := validateSourceManifest("manifest.tsv", []byte(invalidSecond)); err == nil || !strings.Contains(err.Error(), "row 3") {
		t.Fatalf("validateSourceManifest(second invalid row) error = %v", err)
	}
	shortSecond := valid + "short\n"
	if err := validateSourceManifest("manifest.tsv", []byte(shortSecond)); err == nil || !strings.Contains(err.Error(), "row 3") {
		t.Fatalf("validateSourceManifest(second short row) error = %v", err)
	}
	for name, row := range map[string]string{
		"empty id":      "\tv1\thttps://example.com/spec\t" + digest + "\tpinned\n",
		"empty version": "source\t\thttps://example.com/spec\t" + digest + "\tpinned\n",
		"bad URL":       "source\tv1\t://\t" + digest + "\tpinned\n",
		"bad digest":    "source\tv1\thttps://example.com/spec\tbad\tpinned\n",
		"bad status":    "source\tv1\thttps://example.com/spec\t" + digest + "\tmutable\n",
	} {
		t.Run(name, func(t *testing.T) {
			err := validateSourceManifest("manifest.tsv", []byte(strings.Split(valid, "\n")[0]+"\n"+row))
			if err == nil || !strings.Contains(err.Error(), "row 2") {
				t.Fatalf("validateSourceManifest(%s) error = %v", name, err)
			}
		})
	}
	value := map[string]any{
		"first":  map[string]any{"url": "https://example.com/one", "sha256": digest},
		"second": []any{map[string]any{"id": "two", "sha256": digest}},
	}
	if pins, err := validateJSONPins(value); err != nil || pins != 2 {
		t.Fatalf("validateJSONPins() = %d, %v", pins, err)
	}
	if pins, err := validateJSONPins([]any{
		map[string]any{"id": "one", "sha256": digest},
		map[string]any{"id": "two", "sha256": digest},
	}); err != nil || pins != 2 {
		t.Fatalf("validateJSONPins(array) = %d, %v", pins, err)
	}
	if jsonPinHasIdentity(map[string]any{"id": 1, "name": " "}, "sha256") {
		t.Fatal("jsonPinHasIdentity(non-text identities) = true")
	}
	if !jsonPinHasIdentity(map[string]any{"id": 1, "name": " ", "source": "rfc"}, "sha256") {
		t.Fatal("jsonPinHasIdentity() did not continue to a valid identity")
	}
}

func TestExecutableEvidenceDiscoveryBoundaryContract(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "root_test.go"), "//go:build contract\n\npackage example\nimport \"testing\"\nfunc TestRoot(t *testing.T) {}\n")
	write(t, filepath.Join(root, "plain_test.go"), "package example\nimport \"testing\"\nfunc TestPlain(t *testing.T) {}\n")
	write(t, filepath.Join(root, "nested", "nested_test.go"), "package nested\nimport \"testing\"\nfunc TestNested(t *testing.T) {}\n")
	write(t, filepath.Join(root, "nested", ".git", "ignored_test.go"), "package ignored\nimport \"testing\"\nfunc TestIgnored(t *testing.T) {}\n")
	write(t, filepath.Join(root, "nested", ".verification", "ignored_test.go"), "package ignored\nimport \"testing\"\nfunc TestVerificationIgnored(t *testing.T) {}\n")
	modules := []inventory.Module{{Directory: ".", TestTags: []string{"contract"}}, {Directory: "nested"}}
	rootEvidence, err := testEvidence(root, ".", modules)
	if err != nil {
		t.Fatalf("testEvidence(root) error = %v", err)
	}
	if _, ok := rootEvidence["TestRoot"]; !ok {
		t.Fatalf("testEvidence(root) = %#v", rootEvidence)
	}
	if _, ok := rootEvidence["TestNested"]; ok {
		t.Fatalf("testEvidence(root) crossed nested module: %#v", rootEvidence)
	}
	emptyRootEvidence, err := testEvidence(root, "", nil)
	if err != nil {
		t.Fatalf("testEvidence(empty root directory) error = %v", err)
	}
	if _, ok := emptyRootEvidence["TestPlain"]; !ok {
		t.Fatalf("testEvidence(empty root directory) = %#v", emptyRootEvidence)
	}
	nestedEvidence, err := testEvidence(root, "nested", modules)
	if err != nil {
		t.Fatalf("testEvidence(nested) error = %v", err)
	}
	if _, ok := nestedEvidence["TestNested"]; !ok {
		t.Fatalf("testEvidence(nested) = %#v", nestedEvidence)
	}
	if _, ok := nestedEvidence["TestIgnored"]; ok {
		t.Fatalf("testEvidence(nested) included .git evidence: %#v", nestedEvidence)
	}
	if _, ok := nestedEvidence["TestVerificationIgnored"]; ok {
		t.Fatalf("testEvidence(nested) included .verification evidence: %#v", nestedEvidence)
	}
	dotGitRoot := filepath.Join(t.TempDir(), ".git")
	write(t, filepath.Join(dotGitRoot, "contract_test.go"), "package example\nimport \"testing\"\nfunc TestDotGitRoot(t *testing.T) {}\n")
	if evidence, err := testEvidence(dotGitRoot, ".", nil); err != nil {
		t.Fatalf("testEvidence(.git root) error = %v", err)
	} else if _, ok := evidence["TestDotGitRoot"]; !ok {
		t.Fatalf("testEvidence(.git root) = %#v", evidence)
	}
}

func TestExecutableEvidenceAcceptsExactMaximumSize(t *testing.T) {
	root := t.TempDir()
	prefix := "package example\n//"
	write(t, filepath.Join(root, "maximum_test.go"), prefix+strings.Repeat("x", maximumEvidenceSize-len(prefix)))
	if _, err := testEvidence(root, ".", nil); err != nil {
		t.Fatalf("testEvidence(exact maximum) error = %v", err)
	}
}

func TestValidateProvenanceRejectsBothMissingEncodings(t *testing.T) {
	for _, provenance := range [][]byte{nil, []byte("null")} {
		err := validateProvenance(t.TempDir(), inventory.Module{Directory: ".", Provenance: provenance}, nil)
		if err == nil || err.Error() != "source manifest is not declared" {
			t.Fatalf("validateProvenance(%q) error = %v", provenance, err)
		}
	}
}

func TestExecutableEvidenceImportScanningContract(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"non-testing alias", `package example; import testing "fmt"; func TestContract(t *testing.T) {}`, false},
		{"default then dot", `package example; import ("testing"; . "testing"); func TestContract(t *T) {}`, true},
		{"blank then alias", `package example; import (_ "testing"; testpkg "testing"); func TestContract(t *testpkg.T) {}`, true},
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
			if got := isExecutableEvidence(file, function); got != test.want {
				t.Fatalf("isExecutableEvidence() = %v, want %v", got, test.want)
			}
		})
	}
}

func copyFixtureModule(t *testing.T, root, prefix string) {
	t.Helper()
	for _, relative := range []string{"README.md", "COMPATIBILITY.md", "CHANGELOG.md", "docs/specification-decisions.md", "specification/README.md", "specification/manifest.tsv", "specification/monitoring.json", "specification/decisions.json", "specification/conformance.json", "specification/decision-history.json", "testdata/request-target.json", "testdata/peer.tsv"} {
		data := readFile(t, filepath.Join(root, relative))
		content := strings.ReplaceAll(string(data), `"docs/specification-decisions.md"`, `"nested/docs/specification-decisions.md"`)
		content = strings.ReplaceAll(content, `"README.md"`, `"nested/README.md"`)
		content = strings.ReplaceAll(content, `"COMPATIBILITY.md"`, `"nested/COMPATIBILITY.md"`)
		content = strings.ReplaceAll(content, `"CHANGELOG.md"`, `"nested/CHANGELOG.md"`)
		content = strings.ReplaceAll(content, `"specification/README.md"`, `"nested/specification/README.md"`)
		content = strings.ReplaceAll(content, `"testdata/request-target.json"`, `"nested/testdata/request-target.json"`)
		content = strings.ReplaceAll(content, `"testdata/peer.tsv"`, `"nested/testdata/peer.tsv"`)
		content = strings.ReplaceAll(content, "EXAMPLE-DEC-001", "EXAMPLE-DEC-002")
		write(t, filepath.Join(root, prefix, relative), content)
	}
	write(t, filepath.Join(root, prefix, "contract_test.go"), "package nested\nimport \"testing\"\nfunc TestRequestTargetContract(t *testing.T) {}\nfunc FuzzRequestTargetContract(f *testing.F) {}\n")
	refreshModuleDecisionArtifacts(t, root, prefix, "https://www.rfc-editor.org/rfc/rfc9110.txt")
}

func addSecondDecision(t *testing.T, root string) {
	t.Helper()
	decisionPath := filepath.Join(root, "specification/decisions.json")
	decisions, err := loadDecisions(readFile(t, decisionPath))
	if err != nil {
		t.Fatal(err)
	}
	second := decisions.Decisions[0]
	second.ID = "EXAMPLE-DEC-002"
	second.Title = "Second policy"
	decisions.Decisions = append(decisions.Decisions, second)
	writeJSON(t, decisionPath, decisions)

	conformancePath := filepath.Join(root, "specification/conformance.json")
	conformance, err := loadConformance(readFile(t, conformancePath))
	if err != nil {
		t.Fatal(err)
	}
	secondRow := conformance.Decisions[0]
	secondRow.ID = second.ID
	conformance.Decisions = append(conformance.Decisions, secondRow)
	writeJSON(t, conformancePath, conformance)

	registerPath := filepath.Join(root, "docs/specification-decisions.md")
	entry := string(readFile(t, registerPath))
	entry = strings.ReplaceAll(entry, "EXAMPLE-DEC-001", second.ID)
	entry = strings.ReplaceAll(entry, "Request target normalization", second.Title)
	write(t, registerPath, string(readFile(t, registerPath))+"\n"+entry)
	write(t, filepath.Join(root, "specification/README.md"), string(readFile(t, filepath.Join(root, "specification/README.md")))+"\n"+second.ID+"\n")
	refreshDecisionArtifacts(t, root)
}

func writeMonitoringFixture(t *testing.T, root string, count, interval int) {
	t.Helper()
	authorities := make([]string, 0, count)
	for index := range count - 1 {
		authorities = append(authorities, fmt.Sprintf(`{"id":"source-%d","kind":"specification","version":"RFC 9110","url":"https://example.com/source/%d","sha256":"%s","specifications":["RFC 9110"]}`, index, index, strings.Repeat("a", 64)))
	}
	authorities = append(authorities, `{"id":"errata","kind":"errata","url":"https://example.com/errata","sha256":"`+strings.Repeat("b", 64)+`","specifications":["RFC 9110"]}`)
	write(t, filepath.Join(root, "monitoring.json"), fmt.Sprintf(`{"schema_version":1,"reviewed_at":"2026-08-30","review_interval_days":%d,"authorities":[%s]}`, interval, strings.Join(authorities, ",")))
}
