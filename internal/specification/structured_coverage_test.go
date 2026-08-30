package specification

import (
	"encoding/json"
	"go/ast"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestStructuredManifestDecodersRejectEveryInvalidBoundary(t *testing.T) {
	for name, data := range map[string]string{
		"malformed": `{`,
		"schema":    `{"schema_version":2,"decisions":[{}]}`,
		"empty":     `{"schema_version":1,"decisions":[]}`,
	} {
		t.Run("decisions "+name, func(t *testing.T) {
			if _, err := loadDecisions([]byte(data)); err == nil {
				t.Fatal("loadDecisions() error = nil")
			}
		})
		t.Run("conformance "+name, func(t *testing.T) {
			if _, err := loadConformance([]byte(data)); err == nil {
				t.Fatal("loadConformance() error = nil")
			}
		})
	}
	for name, data := range map[string]string{
		"empty input":    ``,
		"nested invalid": `[{"key":`,
		"duplicate":      `{"schema_version":1,"schema_version":1}`,
		"multiple":       `{} {}`,
		"trailing":       `{}x`,
	} {
		t.Run(name, func(t *testing.T) {
			var target map[string]any
			if err := decodeStrictJSON([]byte(data), &target); err == nil {
				t.Fatal("decodeStrictJSON() error = nil")
			}
		})
	}

	for name, relative := range map[string]string{
		"decisions":   "specification/decisions.json",
		"conformance": "specification/conformance.json",
	} {
		t.Run("check "+name, func(t *testing.T) {
			root, catalog := validFixture(t)
			write(t, filepath.Join(root, relative), "{")
			if _, err := Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestDecisionLedgerRejectsEveryInvalidBoundary(t *testing.T) {
	item := decision{ID: "EXAMPLE-DEC-001"}
	digest := decisionDigest(item)
	valid := decisionHistory{SchemaVersion: 1, Entries: []decisionHistoryEntry{{ID: item.ID, Digests: []string{digest}}}}
	if err := validateDecisionLedger(valid, []decision{item}); err != nil {
		t.Fatal(err)
	}
	for name, history := range map[string]decisionHistory{
		"empty":           {Entries: []decisionHistoryEntry{{}}},
		"duplicate":       {Entries: []decisionHistoryEntry{{ID: item.ID, Digests: []string{digest}}, {ID: item.ID, Digests: []string{digest}}}},
		"invalid digest":  {Entries: []decisionHistoryEntry{{ID: item.ID, Digests: []string{"invalid"}}}},
		"missing current": {Entries: nil},
		"orphan":          {Entries: []decisionHistoryEntry{{ID: item.ID, Digests: []string{digest}}, {ID: "OLD-DEC-001", Digests: []string{digest}}}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateDecisionLedger(history, []decision{item}); err == nil {
				t.Fatal("error = nil")
			}
		})
	}
}

func TestValidateDecisionRejectsEveryStructuredBoundary(t *testing.T) {
	root, catalog, item, authorities, evidence := structuredFixture(t)
	module := catalog.Modules[0]
	tests := map[string]func(*decision, map[string]authority){
		"identifier":     func(value *decision, _ map[string]authority) { value.ID = "invalid" },
		"empty field":    func(value *decision, _ map[string]authority) { value.Title = "" },
		"status":         func(value *decision, _ map[string]authority) { value.Status = "pending" },
		"classification": func(value *decision, _ map[string]authority) { value.Classification = "policy" },
		"scope":          func(value *decision, _ map[string]authority) { value.DecisionScope = "local" },
		"strength":       func(value *decision, _ map[string]authority) { value.RequirementStrength = "sometimes" },
		"defensive MUST": func(value *decision, _ map[string]authority) { value.RequirementStrength = "MUST" },
		"recommended MUST": func(value *decision, _ map[string]authority) {
			value.DecisionScope, value.RequirementStrength = "recommended", "MUST"
		},
		"recommendation MUST": func(value *decision, items map[string]authority) {
			value.DecisionScope, value.RequirementStrength = "normative", "MUST"
			source := items[value.SourceAuthority]
			source.Kind = "recommendation"
			items[value.SourceAuthority] = source
		},
		"authority": func(value *decision, _ map[string]authority) { value.SourceAuthority = "missing" },
		"authority kind": func(value *decision, items map[string]authority) {
			source := items[value.SourceAuthority]
			source.Kind = "errata"
			items[value.SourceAuthority] = source
		},
		"coverage":        func(value *decision, _ map[string]authority) { value.Specification = "RFC 9999" },
		"version":         func(value *decision, _ map[string]authority) { value.Version = "RFC 9999" },
		"interpretations": func(value *decision, _ map[string]authority) { value.Interpretations = nil },
		"evidence list":   func(value *decision, _ map[string]authority) { value.PublicAPIs = nil },
		"resolved evidence": func(value *decision, _ map[string]authority) {
			value.ExecutableEvidence = nil
		},
		"resolved fuzz only": func(value *decision, _ map[string]authority) {
			value.ExecutableEvidence = []string{"FuzzRequestTargetContract"}
		},
		"missing evidence": func(value *decision, _ map[string]authority) { value.ExecutableEvidence = []string{"TestMissing"} },
		"fuzz prefix":      func(value *decision, _ map[string]authority) { value.FuzzEvidence = []string{"TestContract"} },
		"missing fuzz":     func(value *decision, _ map[string]authority) { value.FuzzEvidence = []string{"FuzzMissing"} },
		"outside file":     func(value *decision, _ map[string]authority) { value.FixtureEvidence = []string{"../fixture"} },
		"missing file": func(value *decision, _ map[string]authority) {
			value.FixtureEvidence = []string{"testdata/missing.json"}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := item
			items := make(map[string]authority, len(authorities))
			maps.Copy(items, authorities)
			mutate(&value, items)
			if err := validateDecision(root, module, catalog.Modules, value, items, evidence); err == nil {
				t.Fatal("validateDecision() error = nil")
			}
		})
	}
}

func TestValidateDecisionAcceptsNonMandatoryPolicies(t *testing.T) {
	for name, policy := range map[string]struct {
		classification string
		scope          string
		strength       string
	}{
		"permissive application policy": {"optional behavior", "application-policy", "MAY"},
		"defensive recommendation":      {"interoperability policy", "defensive", "SHOULD"},
	} {
		t.Run(name, func(t *testing.T) {
			root, catalog, item, authorities, evidence := structuredFixture(t)
			item.Classification = policy.classification
			item.DecisionScope = policy.scope
			item.RequirementStrength = policy.strength

			if err := validateDecision(root, catalog.Modules[0], catalog.Modules, item, authorities, evidence); err != nil {
				t.Fatalf("validateDecision() error = %v", err)
			}
		})
	}
}

func TestOptionalEvidenceLanesRecordTheirHonestDisposition(t *testing.T) {
	root, catalog, item, authorities, evidence := structuredFixture(t)
	item.FixtureEvidence = nil
	item.FuzzEvidence = nil
	item.InteroperabilityEvidence = nil
	if err := validateDecision(root, catalog.Modules[0], catalog.Modules, item, authorities, evidence); err != nil {
		t.Fatalf("validateDecision(optional evidence omitted) error = %v", err)
	}

	fixtureRoot, _ := validFixture(t)
	manifest, err := loadConformance(readFile(t, filepath.Join(fixtureRoot, "specification/conformance.json")))
	if err != nil {
		t.Fatal(err)
	}
	row := manifest.Decisions[0]
	row.Fixtures = nil
	row.Fuzz = nil
	row.DifferentialEvidence = nil
	row.DifferentialClassification = "not assessed"
	if err := validateConformance(item, row, authorities, evidence); err != nil {
		t.Fatalf("validateConformance(optional evidence omitted) error = %v", err)
	}

	item.Status = "unresolved"
	item.ExecutableEvidence = nil
	row.ExecutableEvidence = nil
	if err := validateDecision(root, catalog.Modules[0], catalog.Modules, item, authorities, evidence); err != nil {
		t.Fatalf("validateDecision(unresolved without executable evidence) error = %v", err)
	}
	if err := validateConformance(item, row, authorities, evidence); err != nil {
		t.Fatalf("validateConformance(unresolved without executable evidence) error = %v", err)
	}
}

func TestDifferentialDispositionMatchesEvidence(t *testing.T) {
	_, _, item, authorities, evidence := structuredFixture(t)
	root, _ := validFixture(t)
	manifest, err := loadConformance(readFile(t, filepath.Join(root, "specification/conformance.json")))
	if err != nil {
		t.Fatal(err)
	}
	base := manifest.Decisions[0]

	for classification := range allowedDifferentialClassifications {
		t.Run("evidence "+classification, func(t *testing.T) {
			row := base
			row.DifferentialClassification = classification
			if err := validateConformance(item, row, authorities, evidence); err != nil {
				t.Fatalf("validateConformance() error = %v", err)
			}
		})
		t.Run("no evidence "+classification, func(t *testing.T) {
			row := base
			row.DifferentialEvidence = nil
			row.DifferentialClassification = classification
			itemWithoutDifferential := item
			itemWithoutDifferential.InteroperabilityEvidence = nil
			if err := validateConformance(itemWithoutDifferential, row, authorities, evidence); err == nil {
				t.Fatal("validateConformance() error = nil")
			}
		})
	}

	row := base
	row.DifferentialClassification = "not assessed"
	if err := validateConformance(item, row, authorities, evidence); err == nil {
		t.Fatal("validateConformance(nonempty evidence marked not assessed) error = nil")
	}
	row.DifferentialEvidence = nil
	item.InteroperabilityEvidence = nil
	if err := validateConformance(item, row, authorities, evidence); err != nil {
		t.Fatalf("validateConformance(empty evidence marked not assessed) error = %v", err)
	}
}

func TestSupersededEvidenceRemainsAuditableWithoutCurrentArtifacts(t *testing.T) {
	root, catalog, item, authorities, evidence := structuredFixture(t)
	item.Status = "superseded"
	item.ExecutableEvidence = []string{"TestRetiredContract"}
	item.FixtureEvidence = []string{"testdata/retired.json"}
	item.FuzzEvidence = []string{"FuzzRetiredContract"}
	item.InteroperabilityEvidence = []string{"testdata/retired-peer.tsv"}
	if err := validateDecision(root, catalog.Modules[0], catalog.Modules, item, authorities, evidence); err != nil {
		t.Fatalf("validateDecision(superseded historical evidence) error = %v", err)
	}

	fixtureRoot, _ := validFixture(t)
	manifest, err := loadConformance(readFile(t, filepath.Join(fixtureRoot, "specification/conformance.json")))
	if err != nil {
		t.Fatal(err)
	}
	row := manifest.Decisions[0]
	row.ExecutableEvidence = append([]string(nil), item.ExecutableEvidence...)
	row.Fixtures = append([]string(nil), item.FixtureEvidence...)
	row.Fuzz = append([]string(nil), item.FuzzEvidence...)
	row.DifferentialEvidence = append([]string(nil), item.InteroperabilityEvidence...)
	if err := validateConformance(item, row, authorities, evidence); err != nil {
		t.Fatalf("validateConformance(superseded historical evidence) error = %v", err)
	}
	item.Documentation = []string{"docs/retired.md"}
	if err := validateDecision(root, catalog.Modules[0], catalog.Modules, item, authorities, evidence); err == nil {
		t.Fatal("validateDecision(superseded missing documentation) error = nil")
	}
}

func TestAdditionalAuthoritativeSourcesCoverDeclaredSpecifications(t *testing.T) {
	_, _, item, authorities, evidence := structuredFixture(t)
	additional := authority{
		ID: "extension-source", Kind: "extension", Version: "Extension 1", URL: "https://example.com/extension",
		Specifications: []string{"Example Extension"},
	}
	authorities[additional.ID] = additional
	root, _ := validFixture(t)
	manifest, err := loadConformance(readFile(t, filepath.Join(root, "specification/conformance.json")))
	if err != nil {
		t.Fatal(err)
	}
	row := manifest.Decisions[0]
	row.AuthoritativeSources = append(row.AuthoritativeSources, additional.ID)
	if err := validateConformance(item, row, authorities, evidence); err != nil {
		t.Fatalf("validateConformance(additional specification source) error = %v", err)
	}

	record, err := json.Marshal(documentedAuthority{ID: additional.ID, Version: additional.Version, URL: additional.URL, Specifications: additional.Specifications})
	if err != nil {
		t.Fatal(err)
	}
	body := "`" + string(record) + "`"
	if err := validateAdditionalAuthoritativeSourceDocumentation(item, row, authorities, body); err != nil {
		t.Fatalf("validateAdditionalAuthoritativeSourceDocumentation() error = %v", err)
	}
	for name, missing := range map[string]string{
		"identifier":    additional.ID,
		"version":       additional.Version,
		"URL":           additional.URL,
		"specification": additional.Specifications[0],
	} {
		t.Run(name, func(t *testing.T) {
			incomplete := strings.Replace(body, missing, "omitted", 1)
			if err := validateAdditionalAuthoritativeSourceDocumentation(item, row, authorities, incomplete); err == nil {
				t.Fatal("validateAdditionalAuthoritativeSourceDocumentation() error = nil")
			}
		})
	}
	row.AuthoritativeSources = append(row.AuthoritativeSources, "missing")
	if err := validateAdditionalAuthoritativeSourceDocumentation(item, row, authorities, body); err == nil {
		t.Fatal("validateAdditionalAuthoritativeSourceDocumentation(unknown source) error = nil")
	}
}

func TestValidateConformanceRejectsEveryStructuredBoundary(t *testing.T) {
	_, _, item, authorities, evidence := structuredFixture(t)
	root, _ := validFixture(t)
	manifestData := readFile(t, filepath.Join(root, "specification/conformance.json"))
	manifest, err := loadConformance(manifestData)
	if err != nil {
		t.Fatal(err)
	}
	row := manifest.Decisions[0]
	tests := map[string]func(*conformanceDecision, map[string]authority){
		"empty list": func(value *conformanceDecision, _ map[string]authority) { value.PublicBehavior = nil },
		"primary source": func(value *conformanceDecision, _ map[string]authority) {
			value.AuthoritativeSources = []string{"other"}
		},
		"unknown source": func(value *conformanceDecision, _ map[string]authority) {
			value.AuthoritativeSources = append(value.AuthoritativeSources, "missing")
		},
		"wrong source kind": func(_ *conformanceDecision, items map[string]authority) {
			source := items[item.SourceAuthority]
			source.Kind = "errata"
			items[item.SourceAuthority] = source
		},
		"source coverage": func(_ *conformanceDecision, items map[string]authority) {
			source := items[item.SourceAuthority]
			source.Specifications = []string{"other"}
			items[item.SourceAuthority] = source
		},
		"set mismatch": func(value *conformanceDecision, _ map[string]authority) {
			value.Fixtures = append(value.Fixtures, "extra")
		},
		"classification": func(value *conformanceDecision, _ map[string]authority) { value.DifferentialClassification = "unknown" },
		"missing evidence": func(value *conformanceDecision, _ map[string]authority) {
			value.ExecutableEvidence = []string{"TestMissing"}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := row
			decisionValue := item
			items := make(map[string]authority, len(authorities))
			maps.Copy(items, authorities)
			if name == "missing evidence" {
				decisionValue.ExecutableEvidence = []string{"TestMissing"}
			}
			mutate(&value, items)
			if err := validateConformance(decisionValue, value, items, evidence); err == nil {
				t.Fatal("validateConformance() error = nil")
			}
		})
	}
	if sameStringSet([]string{"a"}, []string{"a", "b"}) || sameStringSet([]string{"a"}, []string{"b"}) || !sameStringSet([]string{"a", "b"}, []string{"b", "a"}) {
		t.Fatal("sameStringSet() does not enforce exact set equality")
	}
	if err := validateTextList("ID", "values", []string{" "}); err == nil {
		t.Fatal("validateTextList(blank) error = nil")
	}
	if err := validateTextList("ID", "values", []string{"a", "a"}); err == nil {
		t.Fatal("validateTextList(duplicate) error = nil")
	}
	if slicesContains([]string{"a"}, "b") {
		t.Fatal("slicesContains() = true")
	}
}

func TestStructuredManifestContractRejectsCrossSurfaceDrift(t *testing.T) {
	tests := map[string]func(string, *decisionManifest, *conformanceManifest, *[]byte, map[string]struct{}){
		"conformance identifier": func(_ string, _ *decisionManifest, conformance *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			conformance.Decisions[0].ID = "invalid"
		},
		"duplicate conformance": func(_ string, _ *decisionManifest, conformance *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			conformance.Decisions = append(conformance.Decisions, conformance.Decisions[0])
		},
		"duplicate decision": func(_ string, manifest *decisionManifest, _ *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			manifest.Decisions = append(manifest.Decisions, manifest.Decisions[0])
		},
		"missing register": func(_ string, _ *decisionManifest, _ *conformanceManifest, register *[]byte, _ map[string]struct{}) {
			*register = []byte("# Decisions\n")
		},
		"heading only register": func(_ string, _ *decisionManifest, _ *conformanceManifest, register *[]byte, _ map[string]struct{}) {
			*register = []byte("## EXAMPLE-DEC-001: Request target normalization\nstatus source and issue interpretations and peer behavior selected behavior security resource compatibility wire evidence public upstream reconsider\n")
		},
		"title mismatch": func(_ string, manifest *decisionManifest, _ *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			manifest.Decisions[0].Title = "Different"
		},
		"missing matrix": func(_ string, _ *decisionManifest, _ *conformanceManifest, _ *[]byte, matrix map[string]struct{}) {
			delete(matrix, "EXAMPLE-DEC-001")
		},
		"missing conformance": func(_ string, _ *decisionManifest, conformance *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			conformance.Decisions = nil
		},
		"superseded self": func(_ string, manifest *decisionManifest, _ *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			manifest.Decisions[0].Status = "superseded"
			manifest.Decisions[0].Replacement = manifest.Decisions[0].ID
		},
		"superseded unknown": func(_ string, manifest *decisionManifest, _ *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			manifest.Decisions[0].Status = "superseded"
			manifest.Decisions[0].Replacement = "EXAMPLE-DEC-999"
		},
		"unknown register": func(_ string, _ *decisionManifest, _ *conformanceManifest, register *[]byte, _ map[string]struct{}) {
			orphan := strings.ReplaceAll(string(*register), "EXAMPLE-DEC-001", "EXAMPLE-DEC-999")
			orphan = strings.ReplaceAll(orphan, "Request target normalization", "Unknown")
			*register = append(*register, []byte("\n"+orphan)...)
		},
		"unknown matrix": func(_ string, _ *decisionManifest, _ *conformanceManifest, _ *[]byte, matrix map[string]struct{}) {
			matrix["EXAMPLE-DEC-999"] = struct{}{}
		},
		"unknown conformance": func(_ string, _ *decisionManifest, conformance *conformanceManifest, _ *[]byte, _ map[string]struct{}) {
			row := conformance.Decisions[0]
			row.ID = "EXAMPLE-DEC-999"
			conformance.Decisions = append(conformance.Decisions, row)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog, manifest, conformance, policy, register, matrix, evidence := structuredContractFixture(t)
			mutate(root, &manifest, &conformance, &register, matrix)
			if _, _, err := validateDecisionContract(root, catalog.Modules[0], catalog.Modules, policy, manifest, conformance, register, matrix, evidence); err == nil {
				t.Fatal("validateDecisionContract() error = nil")
			}
		})
	}
}

func TestValidateChangeControlRejectsEveryFailureBoundary(t *testing.T) {
	root, catalog, manifest, _, _, _, _, _ := structuredContractFixture(t)
	module := catalog.Modules[0]
	modules := append([]inventory.Module(nil), catalog.Modules...)
	modules = append(modules, inventory.Module{Directory: "nested"})
	registerPath := "docs/specification-decisions.md"
	tests := map[string]func(*changeControl, *inventory.Module){
		"empty":            func(control *changeControl, _ *inventory.Module) { control.README = "" },
		"invalid":          func(control *changeControl, _ *inventory.Module) { control.README = "/absolute" },
		"outside":          func(_ *changeControl, value *inventory.Module) { value.Directory = "nested" },
		"missing document": func(control *changeControl, _ *inventory.Module) { control.README = "missing.md" },
		"missing link": func(control *changeControl, _ *inventory.Module) {
			write(t, filepath.Join(root, "docs/no-link.md"), "# No link\n")
			control.Compatibility = "docs/no-link.md"
		},
		"missing template": func(control *changeControl, _ *inventory.Module) { control.PullRequestTemplate = ".github/missing.md" },
		"incomplete template": func(control *changeControl, _ *inventory.Module) {
			write(t, filepath.Join(root, ".github/incomplete.md"), "## Specification Decisions\n")
			control.PullRequestTemplate = ".github/incomplete.md"
		},
		"nested contribution": func(control *changeControl, _ *inventory.Module) {
			write(t, filepath.Join(root, "nested/CONTRIBUTING.md"), "[Decisions](../docs/specification-decisions.md)\n")
			control.Contribution = "nested/CONTRIBUTING.md"
		},
		"nested template": func(control *changeControl, _ *inventory.Module) {
			write(t, filepath.Join(root, "nested/.github/pull_request_template.md"), "## Specification Decisions\n\nDecision identifier, compatibility, changelog, and superseded history.\n")
			control.PullRequestTemplate = "nested/.github/pull_request_template.md"
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			control := manifest.ChangeControl
			value := module
			mutate(&control, &value)
			if err := validateChangeControl(root, value, modules, control, registerPath, manifest.Decisions); err == nil {
				t.Fatal("validateChangeControl() error = nil")
			}
		})
	}
}

func TestStructuredHelperFailureBoundaries(t *testing.T) {
	if err := validateDecisionHistory(t.TempDir(), "missing.md", nil); err == nil {
		t.Fatal("validateDecisionHistory(missing) error = nil")
	}
	historyRoot := t.TempDir()
	write(t, filepath.Join(historyRoot, "CHANGELOG.md"), "EXAMPLE-DEC-999\n")
	if err := validateDecisionHistory(historyRoot, "CHANGELOG.md", nil); err == nil {
		t.Fatal("validateDecisionHistory(orphan) error = nil")
	}
	if repositoryPath("/absolute") || repositoryPath(".") || repositoryPath("../escape") || !repositoryPath("docs/file.md") {
		t.Fatal("repositoryPath() boundary mismatch")
	}
	if moduleOwnsOrSharesPath("nested", "/absolute", []inventory.Module{{Directory: "."}, {Directory: "nested"}}) {
		t.Fatal("moduleOwnsOrSharesPath() accepted an invalid repository path")
	}
	if documentLinksTo("README.md", []byte("[x](https://example.com) [y](#fragment)"), "docs/target.md") {
		t.Fatal("documentLinksTo() accepted a non-local link")
	}
	if !documentLinksTo("README.md", []byte(`[target](<docs/target.md> "Target")`), "docs/target.md") {
		t.Fatal("documentLinksTo() rejected a valid titled angle-bracket link")
	}
	if documentLinksToDecision([]byte("[x](https://example.com)"), "EXAMPLE-DEC-001", "Replacement") {
		t.Fatal("documentLinksToDecision() accepted a missing fragment")
	}
	if !documentLinksToDecision([]byte("[replacement](#EXAMPLE-DEC-002-Replacement-policy)"), "EXAMPLE-DEC-002", "Replacement policy") {
		t.Fatal("documentLinksToDecision() rejected a replacement link")
	}
	if documentLinksToDecision([]byte("[replacement](#EXAMPLE-DEC-002-invalid)"), "EXAMPLE-DEC-002", "Replacement policy") {
		t.Fatal("documentLinksToDecision() accepted an inexact replacement link")
	}
	if _, err := validateJSONPins(map[string]any{"sha256": strings.Repeat("a", 64), "version": 1}); err == nil {
		t.Fatal("validateJSONPins(identity) error = nil")
	}
	if jsonPinHasIdentity(map[string]any{"version": 1}, "sha256") {
		t.Fatal("jsonPinHasIdentity() = true")
	}
	if pins, err := validateJSONPins([]any{map[string]any{"id": "source", "sha256": strings.Repeat("a", 64)}}); err != nil || pins != 1 {
		t.Fatalf("validateJSONPins(array) = %d, %v", pins, err)
	}
	function := &ast.FuncDecl{
		Name: ast.NewIdent("TestNestedSelector"),
		Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{Type: &ast.StarExpr{X: &ast.SelectorExpr{
			X: &ast.SelectorExpr{X: ast.NewIdent("example"), Sel: ast.NewIdent("testing")}, Sel: ast.NewIdent("T"),
		}}}}}},
	}
	if isExecutableEvidence(&ast.File{}, function) {
		t.Fatal("isExecutableEvidence() accepted a nested selector")
	}
}

func structuredFixture(t *testing.T) (string, inventory.Inventory, decision, map[string]authority, map[string]struct{}) {
	t.Helper()
	root, catalog, manifest, _, policy, _, _, evidence := structuredContractFixture(t)
	authorities := make(map[string]authority, len(policy.Authorities))
	for _, item := range policy.Authorities {
		authorities[item.ID] = item
	}
	return root, catalog, manifest.Decisions[0], authorities, evidence
}

func structuredContractFixture(t *testing.T) (string, inventory.Inventory, decisionManifest, conformanceManifest, monitoring, []byte, map[string]struct{}, map[string]struct{}) {
	t.Helper()
	root, catalog := validFixture(t)
	decisionData := readFile(t, filepath.Join(root, "specification/decisions.json"))
	manifest, err := loadDecisions(decisionData)
	if err != nil {
		t.Fatal(err)
	}
	conformanceData := readFile(t, filepath.Join(root, "specification/conformance.json"))
	conformance, err := loadConformance(conformanceData)
	if err != nil {
		t.Fatal(err)
	}
	monitoringData := readFile(t, filepath.Join(root, "specification/monitoring.json"))
	var policy monitoring
	if err := json.Unmarshal(monitoringData, &policy); err != nil {
		t.Fatal(err)
	}
	register := readFile(t, filepath.Join(root, "docs/specification-decisions.md"))
	matrix, err := matrixDecisions(readFile(t, filepath.Join(root, "specification/README.md")))
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := testEvidence(root, ".", catalog.Modules)
	if err != nil {
		t.Fatal(err)
	}
	return root, catalog, manifest, conformance, policy, register, matrix, evidence
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
