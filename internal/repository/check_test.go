package repository_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repository"
)

func TestCheckAcceptsStandaloneAndMultiModuleRepositories(t *testing.T) {
	root, catalog := fixture(t)
	if err := repository.Check(root, catalog); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	nested := inventory.Module{Directory: "nested", ModulePath: "example/nested", GoVersion: "1.27.0"}
	write(t, filepath.Join(root, "nested", "go.mod"), "module example/nested\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "go.work"), "go 1.27.0\n\nuse (\n\t.\n\t./nested\n)\n")
	catalog.Modules = append(catalog.Modules, nested)
	if err := repository.Check(root, catalog); err != nil {
		t.Fatalf("Check() multi-module error = %v", err)
	}
}

func TestCheckAcceptsUnreleasableFixtureOutsideRepositoryNamespace(t *testing.T) {
	root, catalog := fixture(t)
	fixtureModule := inventory.Module{
		Directory:  "testdata/fixture",
		ModulePath: "example.com/repository-fixture",
		GoVersion:  "1.27.0",
		Releasable: false,
	}
	write(t, filepath.Join(root, "testdata", "fixture", "go.mod"), "module example.com/repository-fixture\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "go.work"), "go 1.27.0\n\nuse (\n\t.\n\t./testdata/fixture\n)\n")
	catalog.Modules = append(catalog.Modules, fixtureModule)

	if err := repository.Check(root, catalog); err != nil {
		t.Fatalf("Check() fixture module error = %v", err)
	}
}

func TestCheckRequiresSpecificationEvidenceForSpecificationBackedModule(t *testing.T) {
	root, catalog := fixture(t)
	catalog.Modules[0].Specifications = []string{"RFC 9110"}
	catalog.Modules[0].Provenance = []byte(`["specification/manifest.tsv"]`)
	write(t, filepath.Join(root, "specification", "manifest.tsv"), "id\tversion\tsections\trole\turl\tsha256\tstatus\n"+
		"rfc9110\tRFC-9110\t7.1\tHTTP\thttps://www.rfc-editor.org/rfc/rfc9110.txt\t21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a\tpinned\n")
	write(t, filepath.Join(root, "specification", "monitoring.json"), `{"schema_version":1,"reviewed_at":"`+time.Now().UTC().Format("2006-01-02")+`","review_interval_days":90,"authorities":[{"id":"source","kind":"specification","version":"RFC 9110","url":"https://www.rfc-editor.org/rfc/rfc9110.txt","sha256":"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a","specifications":["RFC 9110"]},{"id":"errata","kind":"errata","url":"https://www.rfc-editor.org/errata/rfc9110","sha256":"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a","specifications":["RFC 9110"]}]}`)

	err := repository.Check(root, catalog)
	if err == nil || !strings.Contains(err.Error(), "specification decision register") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckBlocksUnresolvedSpecificationDecision(t *testing.T) {
	root, catalog := fixture(t)
	catalog.Modules[0].Specifications = []string{"RFC 9110"}
	catalog.Modules[0].Provenance = []byte(`["specification/manifest.tsv"]`)
	write(t, filepath.Join(root, "specification", "manifest.tsv"), "id\tversion\tsections\trole\turl\tsha256\tstatus\n"+
		"rfc9110\tRFC-9110\t7.1\tHTTP\thttps://www.rfc-editor.org/rfc/rfc9110.txt\t21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a\tpinned\n")
	write(t, filepath.Join(root, "specification", "monitoring.json"), `{"schema_version":1,"reviewed_at":"`+time.Now().UTC().Format("2006-01-02")+`","review_interval_days":90,"authorities":[{"id":"source","kind":"specification","version":"RFC 9110","url":"https://www.rfc-editor.org/rfc/rfc9110.txt","sha256":"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a","specifications":["RFC 9110"]},{"id":"errata","kind":"errata","url":"https://www.rfc-editor.org/errata/rfc9110","sha256":"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a","specifications":["RFC 9110"]}]}`)
	write(t, filepath.Join(root, "specification", "README.md"), "# Matrix\n\nEXAMPLE-DEC-001\n")
	write(t, filepath.Join(root, "docs", "specification-decisions.md"), "# Decisions\n\n## EXAMPLE-DEC-001: Unknown behavior\n\n"+
		"- **Status, owner, and classification:** `unresolved`; maintainers; omission.\n"+
		"- **Source and issue:** [RFC 9110 section 7.1](https://www.rfc-editor.org/rfc/rfc9110.html#section-7.1) omits the behavior.\n"+
		"- **Interpretations and peer behavior:** Accept or reject; peers disagree.\n"+
		"- **Selected behavior, security and resource consequences, compatibility and wire consequences:** No behavior is selected; security, resource, compatibility, and wire consequences remain under review.\n"+
		"- **Evidence, public surface, upstream, and reconsideration:** Evidence and public surface are under review; upstream has no decision; reconsider when clarified.\n")
	write(t, filepath.Join(root, "specification", "decisions.json"), `{"schema_version":1,"decisions":[{"id":"EXAMPLE-DEC-001","title":"Unknown behavior","status":"unresolved","owner":"maintainers","classification":"omission","decision_scope":"defensive","specification":"RFC 9110","version":"RFC 9110","source_authority":"source","section":"7.1","requirement_strength":"not specified","issue":"The behavior is omitted.","interpretations":["Accept","Reject"],"peer_behavior":"Peers disagree.","selected_behavior":"No behavior is selected pending review.","rationale":"The ambiguity remains visible.","security_consequences":"Under review.","resource_consequences":"Under review.","compatibility_consequences":"Under review.","wire_consequences":"Under review.","executable_evidence":["TestContract"],"fixture_evidence":["testdata/contract.json"],"fuzz_evidence":["FuzzContract"],"interoperability_evidence":["testdata/peer.tsv"],"public_apis":["Contract"],"documentation":["docs/specification-decisions.md"],"upstream_status":"No upstream decision exists.","reconsider_when":"The specification clarifies the behavior."}]}`)
	writeSpecificationChangeControl(t, root)
	write(t, filepath.Join(root, "specification", "conformance.json"), `{"schema_version":1,"decisions":[{"id":"EXAMPLE-DEC-001","authoritative_sources":["source"],"executable_evidence":["TestContract"],"fixtures":["testdata/contract.json"],"fuzz":["FuzzContract"],"differential_evidence":["testdata/peer.tsv"],"differential_classification":"specification ambiguity","public_behavior":["Behavior remains unresolved."]}]}`)
	write(t, filepath.Join(root, "testdata", "contract.json"), "{}\n")
	write(t, filepath.Join(root, "testdata", "peer.tsv"), "peer\tbehavior\nexample\tunknown\n")
	write(t, filepath.Join(root, "contract_test.go"), "package fixture\n\nimport \"testing\"\n\nfunc TestContract(t *testing.T) {}\nfunc FuzzContract(f *testing.F) {}\n")

	err := repository.Check(root, catalog)
	if err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("Check() error = %v", err)
	}
}

func writeSpecificationChangeControl(t *testing.T, root string) {
	t.Helper()
	decisionPath := filepath.Join(root, "specification", "decisions.json")
	data, err := os.ReadFile(decisionPath)
	if err != nil {
		t.Fatal(err)
	}
	control := `"change_control":{"readme":"README.md","conformance":"specification/README.md","compatibility":"COMPATIBILITY.md","contribution":"CONTRIBUTING.md","changelog":"CHANGELOG.md","pull_request_template":".github/pull_request_template.md"},`
	write(t, decisionPath, strings.Replace(string(data), `"decisions":`, control+`"decisions":`, 1))
	for path, link := range map[string]string{
		"README.md":        "[Specification decisions](docs/specification-decisions.md)\n",
		"COMPATIBILITY.md": "[Specification decisions](docs/specification-decisions.md)\n",
		"CONTRIBUTING.md":  "[Specification decisions](docs/specification-decisions.md)\n",
		"CHANGELOG.md":     "## Specification Decisions\n\nEXAMPLE-DEC-001: [Decision register](docs/specification-decisions.md)\n",
	} {
		existing, _ := os.ReadFile(filepath.Join(root, path))
		write(t, filepath.Join(root, path), string(existing)+"\n"+link)
	}
	matrixPath := filepath.Join(root, "specification", "README.md")
	matrix, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatal(err)
	}
	write(t, matrixPath, string(matrix)+"\n[Specification decisions](../docs/specification-decisions.md)\n")
	write(t, filepath.Join(root, ".github", "pull_request_template.md"), "## Specification Decisions\n\n- Decision identifier\n- Compatibility impact\n- Changelog entry\n- Superseded identifiers remain in the register\n")
	finalizeSpecificationFixture(t, root, "https://www.rfc-editor.org/rfc/rfc9110.txt")
}

func finalizeSpecificationFixture(t *testing.T, root, authorityURL string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "specification/decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Decisions []map[string]any `json:"decisions"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	registerPath := filepath.Join(root, "docs/specification-decisions.md")
	register, err := os.ReadFile(registerPath)
	if err != nil {
		t.Fatal(err)
	}
	write(t, registerPath, string(register)+"\n"+string(data)+"\n"+authorityURL+"\n")
	var changelog strings.Builder
	changelog.WriteString("# Changelog\n\n")
	history := map[string]any{"schema_version": 1, "entries": []any{}}
	for _, item := range manifest.Decisions {
		encoded, _ := json.Marshal(item)
		sum := sha256.Sum256(encoded)
		_, _ = fmt.Fprintf(&changelog, "- %s sha256:%x\n", item["id"], sum)
		entries, ok := history["entries"].([]any)
		if !ok {
			t.Fatal("decision history entries are not an array")
		}
		history["entries"] = append(entries, map[string]any{"id": item["id"], "digests": []string{hex.EncodeToString(sum[:])}})
	}
	write(t, filepath.Join(root, "CHANGELOG.md"), changelog.String()+"\n[Decisions](docs/specification-decisions.md)\n")
	historyData, _ := json.Marshal(history)
	write(t, filepath.Join(root, "specification/decision-history.json"), string(historyData))
}

func TestCheckRejectsRepositoryContractViolations(t *testing.T) {
	tests := map[string]func(*testing.T, string, *inventory.Inventory){
		"relative root": func(_ *testing.T, _ string, catalog *inventory.Inventory) { catalog.Repository = "relative-root" },
		"version mismatch": func(t *testing.T, root string, _ *inventory.Inventory) {
			write(t, filepath.Join(root, ".go-version"), "1.26.0\n")
		},
		"missing version": func(t *testing.T, root string, _ *inventory.Inventory) { remove(t, filepath.Join(root, ".go-version")) },
		"duplicate module": func(_ *testing.T, _ string, catalog *inventory.Inventory) {
			catalog.Modules = append(catalog.Modules, catalog.Modules[0])
		},
		"legacy tooling": func(t *testing.T, root string, _ *inventory.Inventory) {
			if err := os.Mkdir(filepath.Join(root, ".golib"), 0o700); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := fixture(t)
			mutate(t, root, &catalog)
			checkRoot := root
			if name == "relative root" {
				checkRoot = "."
			}
			if err := repository.Check(checkRoot, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsInvalidModules(t *testing.T) {
	tests := map[string]func(*testing.T, string, *inventory.Module){
		"empty directory":     func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = "" },
		"absolute directory":  func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = "/tmp" },
		"backslash directory": func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = `nested\module` },
		"unclean directory":   func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = "nested/.." },
		"outside namespace":   func(_ *testing.T, _ string, module *inventory.Module) { module.ModulePath = "other/module" },
		"missing go.mod":      func(t *testing.T, root string, _ *inventory.Module) { remove(t, filepath.Join(root, "go.mod")) },
		"malformed go.mod": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module\n")
		},
		"module mismatch": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module example/other\n\ngo 1.27.0\n")
		},
		"missing go version": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module example\n")
		},
		"go version mismatch": func(_ *testing.T, _ string, module *inventory.Module) { module.GoVersion = "1.26.0" },
		"replace directive": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.0\nreplace example/old => ./old\n")
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := fixture(t)
			mutate(t, root, &catalog.Modules[0])
			if err := repository.Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsInvalidWorkspaces(t *testing.T) {
	tests := map[string]string{
		"missing":          "",
		"malformed":        "go [",
		"version mismatch": "go 1.26.0\nuse (\n.\n./nested\n)\n",
		"non-local":        "go 1.27.0\nuse ../other\n",
		"parent":           "go 1.27.0\nuse ..\n",
		"absolute":         "go 1.27.0\nuse /tmp/other\n",
		"nested non-local": "go 1.27.0\nuse ../other/nested\n",
		"root only":        "go 1.27.0\nuse .\n",
		"mismatch":         "go 1.27.0\nuse ./other\n",
	}
	for name, workspace := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := fixture(t)
			catalog.Modules = append(catalog.Modules, inventory.Module{Directory: "nested", ModulePath: "example/nested", GoVersion: "1.27.0"})
			write(t, filepath.Join(root, "nested", "go.mod"), "module example/nested\n\ngo 1.27.0\n")
			if workspace != "" {
				write(t, filepath.Join(root, "go.work"), workspace)
			}
			if err := repository.Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsUnreadableWorkspace(t *testing.T) {
	root, catalog := fixture(t)
	if err := os.Symlink("missing", filepath.Join(root, "go.work")); err != nil {
		t.Fatal(err)
	}
	if err := repository.Check(root, catalog); err == nil || !strings.Contains(err.Error(), "read go.work") {
		t.Fatalf("Check() error = %v", err)
	}
}

func fixture(t *testing.T) (string, inventory.Inventory) {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, ".go-version"), "1.27.0\n")
	write(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.0\n")
	return root, inventory.Inventory{Repository: "example", GoVersion: "1.27.0", Modules: []inventory.Module{{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Releasable: true}}}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func remove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}
