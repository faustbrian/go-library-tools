package inventory_test

import (
	"bytes"
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

const historicalV2ManifestLimit = 32 << 20
const historicalV2LoadEntryPoint = "internal/inventory/inventory.go#Load(root string, policy config.Config) (Inventory, error)"
const historicalV2LoadSnapshotEntryPoint = "internal/inventory/inventory.go#LoadSnapshot(root string, policy config.Config) (Inventory, []byte, error)"

type historicalV2Observation struct {
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

type historicalV2Blob struct {
	Base64     string `json:"base64"`
	PadToBytes int    `json:"pad_to_bytes,omitempty"`
}

type historicalV2Invocation struct {
	Modules        historicalV2Blob            `json:"modules"`
	Packages       historicalV2Blob            `json:"packages"`
	Policy         config.Config               `json:"policy"`
	FixtureSources []historicalV2FixtureSource `json:"fixture_sources,omitempty"`
}

type historicalV2FixtureSource struct {
	Path        string `json:"path"`
	BytesSHA256 string `json:"bytes_sha256"`
}

type historicalV2Case struct {
	id       string
	input    historicalV2Invocation
	accepted bool
}

type historicalV2SnapshotValue struct {
	inventory.Inventory
	SnapshotBytes  int    `json:"snapshot_bytes"`
	SnapshotSHA256 string `json:"snapshot_sha256"`
}

func TestHistoricalModulesV2ObservationRunner(t *testing.T) {
	output := os.Getenv("HISTORICAL_OBSERVATION_PATH")
	if output == "" {
		t.Skip("set HISTORICAL_OBSERVATION_PATH to run the injected historical runner")
	}

	cases := historicalV2Cases(t)
	observations := make([]historicalV2Observation, 0, len(cases)*2)
	for _, test := range cases {
		observations = append(observations, historicalV2ObserveLoad(t, test))
		observations = append(observations, historicalV2ObserveSnapshot(t, test))
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

func historicalV2ObserveLoad(t *testing.T, test historicalV2Case) historicalV2Observation {
	t.Helper()
	inputBytes := historicalV2JSON(t, test.input)
	root, _ := historicalV2Materialize(t, test.input)
	value, err := inventory.Load(root, test.input.Policy)
	if test.accepted != (err == nil) {
		t.Fatalf("%s/load: accepted=%v, error=%v", test.id, test.accepted, err)
	}
	row := historicalV2Observation{
		CaseID: "v2-load-" + test.id, EntryPoint: historicalV2LoadEntryPoint,
		InputBase64: base64.StdEncoding.EncodeToString(inputBytes),
	}
	if err != nil {
		row.Outcome = "rejected"
		row.ErrorClass = historicalV2ErrorClass(t, err)
		return row
	}
	row.Outcome = "accepted"
	decodedValue, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	emittedValue := historicalV2JSON(t, value)
	row.DecodedValueKind, row.EmittedValueKind = "json", "json"
	row.DecodedValueBase64 = base64.StdEncoding.EncodeToString(decodedValue)
	row.EmittedValueBase64 = base64.StdEncoding.EncodeToString(emittedValue)
	return row
}

func historicalV2ObserveSnapshot(t *testing.T, test historicalV2Case) historicalV2Observation {
	t.Helper()
	inputBytes := historicalV2JSON(t, test.input)
	root, moduleBytes := historicalV2Materialize(t, test.input)
	catalog, snapshot, err := inventory.LoadSnapshot(root, test.input.Policy)
	if test.accepted != (err == nil) {
		t.Fatalf("%s/load-snapshot: accepted=%v, error=%v", test.id, test.accepted, err)
	}
	row := historicalV2Observation{
		CaseID: "v2-load-snapshot-" + test.id, EntryPoint: historicalV2LoadSnapshotEntryPoint,
		InputBase64: base64.StdEncoding.EncodeToString(inputBytes),
	}
	if err != nil {
		row.Outcome = "rejected"
		row.ErrorClass = historicalV2ErrorClass(t, err) + ":" + historicalV2SnapshotState(catalog, snapshot, moduleBytes)
		return row
	}
	value := historicalV2SnapshotObservationValue(catalog, snapshot)
	decodedValue, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	emittedValue := historicalV2JSON(t, historicalV2SnapshotObservationValue(catalog, append([]byte(nil), snapshot...)))
	row.Outcome = "accepted"
	row.DecodedValueKind, row.EmittedValueKind = "json", "json"
	row.DecodedValueBase64 = base64.StdEncoding.EncodeToString(decodedValue)
	row.EmittedValueBase64 = base64.StdEncoding.EncodeToString(emittedValue)
	return row
}

func TestHistoricalV2SnapshotObservationValueBoundsExactSnapshot(t *testing.T) {
	t.Parallel()

	snapshot := make([]byte, historicalV2ManifestLimit)
	value := historicalV2SnapshotObservationValue(inventory.Inventory{}, snapshot)
	if value.SnapshotBytes != len(snapshot) {
		t.Fatalf("snapshot bytes = %d, want %d", value.SnapshotBytes, len(snapshot))
	}
	if value.SnapshotSHA256 != historicalV2SHA256(snapshot) {
		t.Fatalf("snapshot digest = %s, want exact-byte digest", value.SnapshotSHA256)
	}
	encoded := historicalV2JSON(t, value)
	if len(encoded) > 1<<10 {
		t.Fatalf("bounded snapshot observation = %d bytes, want at most 1 KiB", len(encoded))
	}
	independent := historicalV2JSON(t, historicalV2SnapshotObservationValue(inventory.Inventory{}, append([]byte(nil), snapshot...)))
	if !bytes.Equal(encoded, independent) {
		t.Fatal("independent snapshot observation differs")
	}
}

func TestHistoricalV2SchemaVersionCasesMutateTheV2Field(t *testing.T) {
	t.Parallel()

	manifest := historicalV2Module("{}", "[]", "[]", "")
	for _, value := range []string{"2.5", "2e0", "-0"} {
		mutated := historicalV2WithSchemaVersion(t, manifest, value)
		if mutated == manifest || !strings.Contains(mutated, `"schema_version":`+value) {
			t.Fatalf("schema-version mutation %q did not alter the v2 field", value)
		}
	}
}

func historicalV2SnapshotObservationValue(catalog inventory.Inventory, snapshot []byte) historicalV2SnapshotValue {
	return historicalV2SnapshotValue{
		Inventory:      catalog,
		SnapshotBytes:  len(snapshot),
		SnapshotSHA256: historicalV2SHA256(snapshot),
	}
}

func historicalV2Cases(t *testing.T) []historicalV2Case {
	t.Helper()
	valid := historicalV2Module("{}", "[]", "[]", "")
	emptyPackages := `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]}`
	pkg := `{"module_directory":".","directory":".","name":"example","import_path":"github.com/faustbrian/example","kind":"public","production":true,"executable":false,"coverage_required":true,"build_required":true,"build_tags":[]}`
	base := historicalV2Policy()
	return []historicalV2Case{
		historicalV2RepositoryFixture(t),
		historicalV2NewCase("accept-canonical", valid, emptyPackages, base, true),
		historicalV2NewCase("accept-surrounding-whitespace", " \n\t"+valid+"\r\n ", emptyPackages, base, true),
		historicalV2NewCase("accept-duplicate-member-last-wins", strings.Replace(valid, `"repository":"github.com/faustbrian/example"`, `"repository":"wrong","repository":"github.com/faustbrian/example"`, 1), emptyPackages, base, true),
		historicalV2NewCase("accept-package-duplicate-member-last-wins", valid, strings.Replace(emptyPackages, `"repository":"github.com/faustbrian/example"`, `"repository":"wrong","repository":"github.com/faustbrian/example"`, 1), base, true),
		historicalV2NewCase("accept-invalid-utf8-replacement", strings.Replace(valid, `"releasable":true`, "\"purpose\":\"\xff\",\"releasable\":true", 1), emptyPackages, base, true),
		historicalV2NewCase("accept-lone-surrogate-replacement", strings.Replace(valid, `"releasable":true`, `"purpose":"\ud800","releasable":true`, 1), emptyPackages, base, true),
		historicalV2NewCase("accept-raw-provenance-negative-zero", historicalV2WithProvenance(valid, `-0`), emptyPackages, base, true),
		historicalV2NewCase("accept-raw-provenance-fraction", historicalV2WithProvenance(valid, `1.5`), emptyPackages, base, true),
		historicalV2NewCase("accept-raw-provenance-exponent", historicalV2WithProvenance(valid, `1e2`), emptyPackages, base, true),
		historicalV2NewCase("accept-depth-below-limit", historicalV2WithProvenance(valid, historicalV2NestedJSON(9996)), emptyPackages, base, true),
		historicalV2NewCase("accept-depth-at-limit", historicalV2WithProvenance(valid, historicalV2NestedJSON(9997)), emptyPackages, base, true),
		historicalV2NewCase("reject-depth-above-limit", historicalV2WithProvenance(valid, historicalV2NestedJSON(9998)), emptyPackages, base, false),
		historicalV2NewCase("reject-int-field-fraction", historicalV2WithSchemaVersion(t, valid, "2.5"), emptyPackages, base, false),
		historicalV2NewCase("reject-int-field-exponent", historicalV2WithSchemaVersion(t, valid, "2e0"), emptyPackages, base, false),
		historicalV2NewCase("reject-int-field-negative-zero", historicalV2WithSchemaVersion(t, valid, "-0"), emptyPackages, base, false),
		historicalV2NewCase("reject-bom", "\ufeff"+valid, emptyPackages, base, false),
		historicalV2NewCase("reject-truncated", "{", emptyPackages, base, false),
		historicalV2NewCase("reject-second-json-value", valid+` {}`, emptyPackages, base, false),
		historicalV2NewCase("reject-malformed-trailing-value", valid+` {`, emptyPackages, base, false),
		historicalV2NewCase("reject-unknown-inventory-member", strings.Replace(valid, `"modules":`, `"unknown":true,"modules":`, 1), emptyPackages, base, false),
		historicalV2NewCase("reject-unknown-module-member", strings.Replace(valid, `"packages":`, `"unknown":true,"packages":`, 1), emptyPackages, base, false),
		historicalV2NewCase("reject-unknown-package-inventory-member", valid, strings.Replace(emptyPackages, `"packages":`, `"unknown":true,"packages":`, 1), base, false),
		historicalV2NewSizedCase("accept-exact-size-limit", valid, historicalV2ManifestLimit, emptyPackages, base, true),
		historicalV2NewSizedCase("reject-size-limit-plus-one", valid, historicalV2ManifestLimit+1, emptyPackages, base, false),
		historicalV2NewCase("reject-duplicate-module-directory", historicalV2Module("{}", "[]", "[]", `,{"directory":".","module_path":"github.com/faustbrian/second","gates":{},"packages":[]}`), emptyPackages, base, false),
		historicalV2NewCase("reject-package-missing-from-packages", historicalV2Module("{}", "[]", "["+pkg+"]", ""), emptyPackages, base, false),
		historicalV2NewCase("reject-package-missing-from-modules", valid, historicalV2Packages(pkg), base, false),
		historicalV2NewCase("reject-package-value-divergence", historicalV2Module("{}", "[]", "["+pkg+"]", ""), historicalV2Packages(strings.Replace(pkg, `"name":"example"`, `"name":"other"`, 1)), base, false),
		historicalV2NewCase("reject-package-module-directory-divergence", historicalV2Module("{}", "[]", "["+strings.Replace(pkg, `"module_directory":"."`, `"module_directory":"nested"`, 1)+"]", ""), historicalV2Packages(pkg), base, false),
		historicalV2NewCase("reject-duplicate-module-package", historicalV2Module("{}", "[]", "["+pkg+","+pkg+"]", ""), historicalV2Packages(pkg), base, false),
		historicalV2NewCase("reject-duplicate-canonical-package", historicalV2Module("{}", "[]", "["+pkg+"]", ""), historicalV2Packages(pkg+","+pkg), base, false),
		historicalV2NewCase("reject-repository-divergence", valid, `{"schema_version":1,"repository":"github.com/faustbrian/other","packages":[]}`, base, false),
		historicalV2NewCase("reject-operation-unknown-module", valid, emptyPackages, historicalV2WithOperations(base, config.Operation{Module: "missing", Gate: "docs"}), false),
		historicalV2NewCase("reject-operation-disabled-gate", valid, emptyPackages, historicalV2WithOperations(base, config.Operation{Module: ".", Gate: "test"}), false),
		historicalV2NewCase("reject-required-fuzz-operation-missing", historicalV2Module(`{"fuzz":true}`, "[]", "[]", ""), emptyPackages, base, false),
		historicalV2NewCase("reject-interoperability-without-tools", valid, emptyPackages, historicalV2WithOperations(base, config.Operation{Module: ".", Gate: "interoperability"}), false),
		historicalV2NewCase("reject-api-baseline-unknown-module", historicalV2Module(`{"api_compatibility":true}`, "[]", "[]", ""), emptyPackages, historicalV2WithAPI(base, "missing"), false),
		historicalV2NewCase("reject-api-baseline-disabled", valid, emptyPackages, historicalV2WithAPI(base, "."), false),
		historicalV2NewCase("reject-mutation-import-unknown-module", historicalV2Module(`{"mutation":true}`, "[]", "[]", ""), emptyPackages, historicalV2WithMutation(base, "missing"), false),
		historicalV2NewCase("reject-mutation-import-disabled", valid, emptyPackages, historicalV2WithMutation(base, "."), false),
		historicalV2NewCase("accept-enabled-policy", historicalV2Module(`{"tests":true,"fuzz":true}`, `["reference"]`, "[]", ""), emptyPackages, historicalV2WithOperations(base, config.Operation{Module: ".", Gate: "test"}, config.Operation{Module: ".", Gate: "fuzz"}, config.Operation{Module: ".", Gate: "interoperability"}), true),
	}
}

func historicalV2RepositoryFixture(t *testing.T) historicalV2Case {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repositoryRoot := filepath.Clean(filepath.Join(workingDirectory, "../.."))
	modules := historicalV2ReadFixture(t, repositoryRoot, "modules.json")
	packages := historicalV2ReadFixture(t, repositoryRoot, "packages.json")
	policyBytes := historicalV2ReadFixture(t, repositoryRoot, ".golib.yaml")
	policy, err := config.Load(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	return historicalV2Case{
		id: "accept-frozen-repository-fixture",
		input: historicalV2Invocation{
			Modules:  historicalV2Blob{Base64: base64.StdEncoding.EncodeToString(modules)},
			Packages: historicalV2Blob{Base64: base64.StdEncoding.EncodeToString(packages)},
			Policy:   policy,
			FixtureSources: []historicalV2FixtureSource{
				{Path: ".golib.yaml", BytesSHA256: historicalV2SHA256(policyBytes)},
				{Path: "modules.json", BytesSHA256: historicalV2SHA256(modules)},
				{Path: "packages.json", BytesSHA256: historicalV2SHA256(packages)},
			},
		},
		accepted: true,
	}
}

func historicalV2ReadFixture(t *testing.T, root, relative string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, relative))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func historicalV2SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum)
}

func historicalV2NewCase(id, modules, packages string, policy config.Config, accepted bool) historicalV2Case {
	return historicalV2Case{id: id, input: historicalV2Invocation{Modules: historicalV2Bytes(modules), Packages: historicalV2Bytes(packages), Policy: policy}, accepted: accepted}
}

func historicalV2NewSizedCase(id, modules string, size int, packages string, policy config.Config, accepted bool) historicalV2Case {
	test := historicalV2NewCase(id, modules, packages, policy, accepted)
	test.input.Modules.PadToBytes = size
	return test
}

func historicalV2Bytes(value string) historicalV2Blob {
	return historicalV2Blob{Base64: base64.StdEncoding.EncodeToString([]byte(value))}
}

func historicalV2Materialize(t *testing.T, invocation historicalV2Invocation) (string, []byte) {
	t.Helper()
	root := t.TempDir()
	modules := historicalV2BlobBytes(t, invocation.Modules)
	packages := historicalV2BlobBytes(t, invocation.Packages)
	if err := os.WriteFile(filepath.Join(root, invocation.Policy.Manifests.Modules), modules, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, invocation.Policy.Manifests.Packages), packages, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, modules
}

func historicalV2BlobBytes(t *testing.T, blob historicalV2Blob) []byte {
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
	return data
}

func historicalV2Policy() config.Config {
	return config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
}

func historicalV2WithOperations(policy config.Config, operations ...config.Operation) config.Config {
	policy.Operations = operations
	return policy
}

func historicalV2WithAPI(policy config.Config, module string) config.Config {
	policy.API.Baselines = []config.APIBaseline{{Module: module}}
	return policy
}

func historicalV2WithMutation(policy config.Config, module string) config.Config {
	policy.Mutation.Imports = []config.MutationImport{{Module: module}}
	return policy
}

func historicalV2Module(gates, tools, packages, extraModule string) string {
	return fmt.Sprintf(`{"schema_version":2,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":false,"gates":%s,"interoperability_tools":%s,"packages":%s}%s]}`, gates, tools, packages, extraModule)
}

func historicalV2Packages(items string) string {
	return `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[` + items + `]}`
}

func historicalV2WithProvenance(manifest, raw string) string {
	return strings.Replace(manifest, `"packages":[]`, `"provenance":`+raw+`,"packages":[]`, 1)
}

func historicalV2WithSchemaVersion(t *testing.T, manifest, value string) string {
	t.Helper()
	const field = `"schema_version":2`
	if strings.Count(manifest, field) != 1 {
		t.Fatalf("module manifest schema-version field count = %d, want 1", strings.Count(manifest, field))
	}
	return strings.Replace(manifest, field, `"schema_version":`+value, 1)
}

func historicalV2NestedJSON(depth int) string {
	return strings.Repeat("[", depth) + "0" + strings.Repeat("]", depth)
}

func historicalV2JSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func historicalV2ErrorClass(t *testing.T, err error) string {
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
		t.Fatalf("unclassified historical v2 error: %v", err)
		return ""
	}
}

func historicalV2SnapshotState(catalog inventory.Inventory, snapshot, input []byte) string {
	if catalog.SchemaVersion == 0 && catalog.Repository == "" && catalog.Modules == nil && snapshot == nil {
		return "zero-catalog:nil-snapshot"
	}
	if catalog.SchemaVersion == 2 && catalog.Repository == "github.com/faustbrian/example" && bytes.Equal(snapshot, input) {
		return "decoded-catalog:exact-snapshot"
	}
	return "other-partial-state"
}
