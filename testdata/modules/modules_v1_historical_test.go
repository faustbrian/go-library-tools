package inventory_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const historicalV1ManifestLimit = 32 << 20
const historicalV1LoadEntryPoint = "internal/inventory/inventory.go#Load(root string, policy config.Config) (Inventory, error)"

type historicalV1Observation struct {
	CaseID             string `json:"case_id"`
	EntryPoint         string `json:"entry_point"`
	InputBase64        string `json:"input_base64"`
	Outcome            string `json:"outcome"`
	DecodedValueKind   string `json:"decoded_value_kind,omitempty"`
	DecodedValueBase64 string `json:"decoded_value_base64,omitempty"`
	EmittedValueKind   string `json:"emitted_value_kind,omitempty"`
	EmittedValueBase64 string `json:"emitted_value_base64,omitempty"`
	ErrorClass         string `json:"error_class,omitempty"`
}

type historicalV1Blob struct {
	Base64     string `json:"base64"`
	PadToBytes int    `json:"pad_to_bytes,omitempty"`
}

type historicalV1Invocation struct {
	Modules        historicalV1Blob            `json:"modules"`
	Packages       historicalV1Blob            `json:"packages"`
	Policy         config.Config               `json:"policy"`
	FixtureSources []historicalV1FixtureSource `json:"fixture_sources,omitempty"`
}

type historicalV1FixtureSource struct {
	Path        string `json:"path"`
	BytesSHA256 string `json:"bytes_sha256"`
}

type historicalV1Case struct {
	id       string
	input    historicalV1Invocation
	accepted bool
}

func TestHistoricalModulesV1ObservationRunner(t *testing.T) {
	output := os.Getenv("HISTORICAL_OBSERVATION_PATH")
	if output == "" {
		t.Skip("set HISTORICAL_OBSERVATION_PATH to run the injected historical runner")
	}

	cases := historicalV1Cases(t)
	observations := make([]historicalV1Observation, 0, len(cases))
	for _, test := range cases {
		invocationBytes := historicalV1JSON(t, test.input)
		root := historicalV1Materialize(t, test.input)
		value, err := inventory.Load(root, test.input.Policy)
		if test.accepted != (err == nil) {
			t.Fatalf("%s: accepted=%v, error=%v", test.id, test.accepted, err)
		}
		row := historicalV1Observation{
			CaseID: test.id, EntryPoint: historicalV1LoadEntryPoint,
			InputBase64: base64.StdEncoding.EncodeToString(invocationBytes),
		}
		if err != nil {
			row.Outcome = "rejected"
			row.ErrorClass = historicalV1ErrorClass(t, err)
		} else {
			row.Outcome = "accepted"
			decodedValue, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			emittedValue := historicalV1JSON(t, value)
			row.DecodedValueKind, row.EmittedValueKind = "json", "json"
			row.DecodedValueBase64 = base64.StdEncoding.EncodeToString(decodedValue)
			row.EmittedValueBase64 = base64.StdEncoding.EncodeToString(emittedValue)
		}
		observations = append(observations, row)
	}

	sort.Slice(observations, func(i, j int) bool { return observations[i].CaseID < observations[j].CaseID })
	data, err := json.MarshalIndent(observations, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(output, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func historicalV1Cases(t *testing.T) []historicalV1Case {
	t.Helper()
	valid := historicalV1Module("{}", "[]", "[]", "")
	emptyPackages := `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]}`
	pkg := `{"module_directory":".","directory":".","name":"example","import_path":"github.com/faustbrian/example","kind":"public","production":true,"executable":false,"coverage_required":true,"build_required":true,"build_tags":[]}`
	base := historicalV1Policy()
	return []historicalV1Case{
		historicalV1RepositoryFixture(t),
		historicalV1NewCase("v1-load-accept-canonical", valid, emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-surrounding-whitespace", " \n\t"+valid+"\r\n ", emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-duplicate-member-last-wins", strings.Replace(valid, `"repository":"github.com/faustbrian/example"`, `"repository":"wrong","repository":"github.com/faustbrian/example"`, 1), emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-package-duplicate-member-last-wins", valid, strings.Replace(emptyPackages, `"repository":"github.com/faustbrian/example"`, `"repository":"wrong","repository":"github.com/faustbrian/example"`, 1), base, true),
		historicalV1NewCase("v1-load-accept-invalid-utf8-replacement", strings.Replace(valid, `"releasable":true`, "\"purpose\":\"\xff\",\"releasable\":true", 1), emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-lone-surrogate-replacement", strings.Replace(valid, `"releasable":true`, `"purpose":"\ud800","releasable":true`, 1), emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-raw-provenance-negative-zero", historicalV1WithProvenance(valid, `-0`), emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-raw-provenance-fraction", historicalV1WithProvenance(valid, `1.5`), emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-raw-provenance-exponent", historicalV1WithProvenance(valid, `1e2`), emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-depth-below-limit", historicalV1WithProvenance(valid, historicalV1NestedJSON(9996)), emptyPackages, base, true),
		historicalV1NewCase("v1-load-accept-depth-at-limit", historicalV1WithProvenance(valid, historicalV1NestedJSON(9997)), emptyPackages, base, true),
		historicalV1NewCase("v1-load-reject-depth-above-limit", historicalV1WithProvenance(valid, historicalV1NestedJSON(9998)), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-int-field-fraction", strings.Replace(valid, `"schema_version":1`, `"schema_version":1.5`, 1), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-int-field-exponent", strings.Replace(valid, `"schema_version":1`, `"schema_version":1e0`, 1), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-int-field-negative-zero", strings.Replace(valid, `"schema_version":1`, `"schema_version":-0`, 1), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-bom", "\ufeff"+valid, emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-truncated", "{", emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-second-json-value", valid+` {}`, emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-malformed-trailing-value", valid+` {`, emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-unknown-inventory-member", strings.Replace(valid, `"modules":`, `"unknown":true,"modules":`, 1), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-unknown-module-member", strings.Replace(valid, `"packages":`, `"unknown":true,"packages":`, 1), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-unknown-package-inventory-member", valid, strings.Replace(emptyPackages, `"packages":`, `"unknown":true,"packages":`, 1), base, false),
		historicalV1NewSizedCase("v1-load-accept-exact-size-limit", valid, historicalV1ManifestLimit, emptyPackages, base, true),
		historicalV1NewSizedCase("v1-load-reject-size-limit-plus-one", valid, historicalV1ManifestLimit+1, emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-duplicate-module-directory", historicalV1Module("{}", "[]", "[]", `,{"directory":".","module_path":"github.com/faustbrian/second","gates":{},"packages":[]}`), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-package-missing-from-packages", historicalV1Module("{}", "[]", "["+pkg+"]", ""), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-package-missing-from-modules", valid, historicalV1Packages(pkg), base, false),
		historicalV1NewCase("v1-load-reject-package-value-divergence", historicalV1Module("{}", "[]", "["+pkg+"]", ""), historicalV1Packages(strings.Replace(pkg, `"name":"example"`, `"name":"other"`, 1)), base, false),
		historicalV1NewCase("v1-load-reject-package-module-directory-divergence", historicalV1Module("{}", "[]", "["+strings.Replace(pkg, `"module_directory":"."`, `"module_directory":"nested"`, 1)+"]", ""), historicalV1Packages(pkg), base, false),
		historicalV1NewCase("v1-load-reject-duplicate-module-package", historicalV1Module("{}", "[]", "["+pkg+","+pkg+"]", ""), historicalV1Packages(pkg), base, false),
		historicalV1NewCase("v1-load-reject-duplicate-canonical-package", historicalV1Module("{}", "[]", "["+pkg+"]", ""), historicalV1Packages(pkg+","+pkg), base, false),
		historicalV1NewCase("v1-load-reject-repository-divergence", valid, `{"schema_version":1,"repository":"github.com/faustbrian/other","packages":[]}`, base, false),
		historicalV1NewCase("v1-load-reject-operation-unknown-module", valid, emptyPackages, historicalV1WithOperations(base, config.Operation{Module: "missing", Gate: "docs"}), false),
		historicalV1NewCase("v1-load-reject-operation-disabled-gate", valid, emptyPackages, historicalV1WithOperations(base, config.Operation{Module: ".", Gate: "test"}), false),
		historicalV1NewCase("v1-load-reject-required-fuzz-operation-missing", historicalV1Module(`{"fuzz":true}`, "[]", "[]", ""), emptyPackages, base, false),
		historicalV1NewCase("v1-load-reject-interoperability-without-tools", valid, emptyPackages, historicalV1WithOperations(base, config.Operation{Module: ".", Gate: "interoperability"}), false),
		historicalV1NewCase("v1-load-reject-api-baseline-unknown-module", historicalV1Module(`{"api_compatibility":true}`, "[]", "[]", ""), emptyPackages, historicalV1WithAPI(base, "missing"), false),
		historicalV1NewCase("v1-load-reject-api-baseline-disabled", valid, emptyPackages, historicalV1WithAPI(base, "."), false),
		historicalV1NewCase("v1-load-reject-mutation-import-unknown-module", historicalV1Module(`{"mutation":true}`, "[]", "[]", ""), emptyPackages, historicalV1WithMutation(base, "missing"), false),
		historicalV1NewCase("v1-load-reject-mutation-import-disabled", valid, emptyPackages, historicalV1WithMutation(base, "."), false),
		historicalV1NewCase("v1-load-accept-enabled-policy", historicalV1Module(`{"tests":true,"fuzz":true}`, `["reference"]`, "[]", ""), emptyPackages, historicalV1WithOperations(base, config.Operation{Module: ".", Gate: "test"}, config.Operation{Module: ".", Gate: "fuzz"}, config.Operation{Module: ".", Gate: "interoperability"}), true),
	}
}

func historicalV1RepositoryFixture(t *testing.T) historicalV1Case {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repositoryRoot := filepath.Clean(filepath.Join(workingDirectory, "../.."))
	modules := historicalV1ReadFixture(t, repositoryRoot, "modules.json")
	packages := historicalV1ReadFixture(t, repositoryRoot, "packages.json")
	policyBytes := historicalV1ReadFixture(t, repositoryRoot, ".golib.yaml")
	policy, err := config.Load(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	return historicalV1Case{
		id: "v1-load-accept-frozen-repository-fixture",
		input: historicalV1Invocation{
			Modules:  historicalV1Blob{Base64: base64.StdEncoding.EncodeToString(modules)},
			Packages: historicalV1Blob{Base64: base64.StdEncoding.EncodeToString(packages)},
			Policy:   policy,
			FixtureSources: []historicalV1FixtureSource{
				{Path: ".golib.yaml", BytesSHA256: historicalV1SHA256(policyBytes)},
				{Path: "modules.json", BytesSHA256: historicalV1SHA256(modules)},
				{Path: "packages.json", BytesSHA256: historicalV1SHA256(packages)},
			},
		},
		accepted: true,
	}
}

func historicalV1ReadFixture(t *testing.T, root, relative string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, relative))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func historicalV1SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum)
}

func historicalV1NewCase(id, modules, packages string, policy config.Config, accepted bool) historicalV1Case {
	return historicalV1Case{id: id, input: historicalV1Invocation{Modules: historicalV1Bytes(modules), Packages: historicalV1Bytes(packages), Policy: policy}, accepted: accepted}
}

func historicalV1NewSizedCase(id, modules string, size int, packages string, policy config.Config, accepted bool) historicalV1Case {
	test := historicalV1NewCase(id, modules, packages, policy, accepted)
	test.input.Modules.PadToBytes = size
	return test
}

func historicalV1Bytes(value string) historicalV1Blob {
	return historicalV1Blob{Base64: base64.StdEncoding.EncodeToString([]byte(value))}
}

func historicalV1Materialize(t *testing.T, invocation historicalV1Invocation) string {
	t.Helper()
	root := t.TempDir()
	historicalV1WriteBlob(t, filepath.Join(root, invocation.Policy.Manifests.Modules), invocation.Modules)
	historicalV1WriteBlob(t, filepath.Join(root, invocation.Policy.Manifests.Packages), invocation.Packages)
	return root
}

func historicalV1WriteBlob(t *testing.T, path string, blob historicalV1Blob) {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(blob.Base64)
	if err != nil {
		t.Fatal(err)
	}
	if blob.PadToBytes != 0 {
		if len(data) > blob.PadToBytes {
			t.Fatalf("fixture length %d exceeds requested size %d", len(data), blob.PadToBytes)
		}
		data = append(data, []byte(strings.Repeat(" ", blob.PadToBytes-len(data)))...)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func historicalV1Policy() config.Config {
	return config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
}

func historicalV1WithOperations(policy config.Config, operations ...config.Operation) config.Config {
	policy.Operations = operations
	return policy
}

func historicalV1WithAPI(policy config.Config, module string) config.Config {
	policy.API.Baselines = []config.APIBaseline{{Module: module}}
	return policy
}

func historicalV1WithMutation(policy config.Config, module string) config.Config {
	policy.Mutation.Imports = []config.MutationImport{{Module: module}}
	return policy
}

func historicalV1Module(gates, tools, packages, extraModule string) string {
	return fmt.Sprintf(`{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":%s,"interoperability_tools":%s,"packages":%s}%s]}`, gates, tools, packages, extraModule)
}

func historicalV1Packages(items string) string {
	return `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[` + items + `]}`
}

func historicalV1WithProvenance(manifest, raw string) string {
	return strings.Replace(manifest, `"packages":[]`, `"provenance":`+raw+`,"packages":[]`, 1)
}
func historicalV1NestedJSON(depth int) string {
	return strings.Repeat("[", depth) + "0" + strings.Repeat("]", depth)
}

func historicalV1JSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func historicalV1ErrorClass(t *testing.T, err error) string {
	t.Helper()
	if errors.Is(err, repositoryfile.ErrTooLarge) {
		return "manifest-too-large"
	}
	message := err.Error()
	switch {
	case strings.Contains(message, "unknown field"):
		return "unknown-field"
	case strings.Contains(message, "multiple JSON values"):
		return "multiple-json-values"
	case strings.Contains(message, "invalid character"), strings.Contains(message, "unexpected EOF"):
		return "invalid-json"
	case strings.Contains(message, "exceeded max depth"):
		return "invalid-json-depth"
	case strings.Contains(message, "cannot unmarshal number"):
		return "invalid-number-for-int"
	case strings.Contains(message, "schema_version must be 1"):
		return "schema-version"
	case strings.Contains(message, "duplicate module directory"):
		return "duplicate-module-directory"
	case strings.Contains(message, "duplicate module package"):
		return "duplicate-module-package"
	case strings.Contains(message, "duplicate package import"):
		return "duplicate-canonical-package"
	case strings.Contains(message, "missing from packages manifest"):
		return "package-missing-from-packages"
	case strings.Contains(message, "missing from module manifest"):
		return "package-missing-from-modules"
	case strings.Contains(message, "differs between canonical manifests"):
		return "package-value-divergence"
	case strings.Contains(message, "module directory does not match"):
		return "package-module-directory-divergence"
	case strings.Contains(message, "repository identities differ"):
		return "repository-identity-divergence"
	case strings.Contains(message, "operations[") && strings.Contains(message, "unknown module"):
		return "operation-unknown-module"
	case strings.Contains(message, "operations[") && strings.Contains(message, "not enabled"):
		return "operation-gate-disabled"
	case strings.Contains(message, "requires a typed operation"):
		return "required-operation-missing"
	case strings.Contains(message, "api.baselines[") && strings.Contains(message, "unknown module"):
		return "api-baseline-unknown-module"
	case strings.Contains(message, "api.baselines[") && strings.Contains(message, "not enabled"):
		return "api-baseline-disabled"
	case strings.Contains(message, "mutation.imports[") && strings.Contains(message, "unknown module"):
		return "mutation-import-unknown-module"
	case strings.Contains(message, "mutation.imports[") && strings.Contains(message, "not enabled"):
		return "mutation-import-disabled"
	default:
		t.Fatalf("unclassified historical v1 error: %v", err)
		return ""
	}
}
