package cohesion_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestSchemaV3PublishesBootstrapVersionedSchemaIdentities(t *testing.T) {
	t.Parallel()

	files := []string{
		"cohesion-schema-v3-decision-review-v1.schema.json",
		"cohesion-schema-v3-decision-freeze-v1.schema.json",
		"cohesion-contract-freeze-v1.schema.json",
		"cohesion-contract-review-v1.schema.json",
	}

	for _, file := range files {
		file := file
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join("..", "..", "schema", file)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read frozen schema %s: %v", file, err)
			}

			var document struct {
				Schema               string `json:"$schema"`
				ID                   string `json:"$id"`
				AdditionalProperties *bool  `json:"additionalProperties"`
			}
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatalf("decode frozen schema %s: %v", file, err)
			}
			wantID := "https://github.com/faustbrian/go-library-tools/schema/" + file
			if document.Schema == "" || document.ID != wantID {
				t.Fatalf("schema identity = (%q, %q), want a declared dialect and %q", document.Schema, document.ID, wantID)
			}
			if document.AdditionalProperties == nil || *document.AdditionalProperties {
				t.Fatal("top-level schema must reject unknown properties")
			}
		})
	}
}

func TestSchemaV3PublishesTheExactVersionedSchemaSet(t *testing.T) {
	t.Parallel()

	files := []string{
		"modules-v1.schema.json",
		"modules-v2.schema.json",
		"modules-v3.schema.json",
		"cohesion-catalog-v1.schema.json",
		"cohesion-catalog-v2.schema.json",
		"cohesion-inputs-v1.schema.json",
		"cohesion-inputs-v2.schema.json",
		"cohesion-sources-v1.schema.json",
		"cohesion-sources-v2.schema.json",
		"cohesion-delivery-evidence-v1.schema.json",
		"cohesion-authorization-registry-v1.schema.json",
		"cohesion-authorization-record-v1.schema.json",
		"cohesion-residual-inventory-v1.schema.json",
		"cohesion-residual-register-v1.schema.json",
		"cohesion-gate-policy-v1.schema.json",
		"cohesion-go-toolchains-v1.schema.json",
		"cohesion-go-official-downloads-oracle-v1.schema.json",
		"cohesion-go-official-downloads-oracle-review-v1.schema.json",
		"cohesion-engineering-identities-v1.schema.json",
		"cohesion-schema-provenance-v1.schema.json",
		"cohesion-schema-provenance-v2.schema.json",
		"cohesion-schema-v3-decision-review-v1.schema.json",
		"cohesion-schema-v3-decision-freeze-v1.schema.json",
		"cohesion-contract-freeze-v1.schema.json",
		"cohesion-contract-review-v1.schema.json",
		"cohesion-diagnostic-v1.schema.json",
	}
	wantFiles := append([]string(nil), files...)
	slices.Sort(wantFiles)
	paths, err := filepath.Glob(filepath.Join("..", "..", "schema", "*-v*.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	actualFiles := make([]string, 0, len(paths))
	for _, path := range paths {
		actualFiles = append(actualFiles, filepath.Base(path))
	}
	slices.Sort(actualFiles)
	if !slices.Equal(actualFiles, wantFiles) {
		t.Fatalf("versioned schema files = %q, want exact set %q", actualFiles, wantFiles)
	}

	compiler := jsonschema.NewCompiler()
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join("..", "..", "schema", file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		var identity struct {
			ID string `json:"$id"`
		}
		if err := json.Unmarshal(data, &identity); err != nil {
			t.Fatalf("decode %s: %v", file, err)
		}
		wantID := "https://github.com/faustbrian/go-library-tools/schema/" + file
		if identity.ID != wantID {
			t.Fatalf("%s $id = %q, want %q", file, identity.ID, wantID)
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		if err := compiler.AddResource(wantID, document); err != nil {
			t.Fatalf("add %s: %v", file, err)
		}
	}
	for _, file := range files {
		identity := "https://github.com/faustbrian/go-library-tools/schema/" + file
		if _, err := compiler.Compile(identity); err != nil {
			t.Fatalf("compile %s: %v", file, err)
		}
	}
}

func TestSchemaProvenanceEntryPointGrammarAcceptsFunctionsAndReceiverMethods(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("..", "..", "schema", "cohesion-schema-provenance-v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Definitions map[string]struct {
			Pattern string `json:"pattern"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	pattern, err := regexp.Compile(document.Definitions["entry_point"].Pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, entryPoint := range []string{
		"internal/inventory/inventory.go#Load(root string) (Inventory, error)",
		"internal/cohesion/source_lock.go#SourceLock.VerifyPolicy(repository, version, checksumsSHA256 string) error",
	} {
		if !pattern.MatchString(entryPoint) {
			t.Fatalf("entry-point grammar rejects %q", entryPoint)
		}
	}
	for _, entryPoint := range []string{
		"internal/cohesion/source_lock.go#load(root string)",
		"internal/cohesion/source_lock.go#SourceLock.verifyPolicy()",
		"internal/cohesion/source_lock.go#SourceLock.Inner.VerifyPolicy()",
		"internal/cohesion/source_lock.go#SourceLock.Field",
	} {
		if pattern.MatchString(entryPoint) {
			t.Fatalf("entry-point grammar accepted non-exported or non-callable path %q", entryPoint)
		}
	}
}

func TestSchemaProvenanceNewContractsBindReviewedOracleLineage(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("..", "..", "schema", "cohesion-schema-provenance-v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	const identity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-schema-provenance-v1.schema.json"
	if err := compiler.AddResource(identity, document); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile(identity + "#/$defs/new_contract")
	if err != nil {
		t.Fatal(err)
	}

	digest := "sha256:" + strings.Repeat("a", 64)
	commit := strings.Repeat("b", 40)
	for _, base := range []string{
		"modules-v3",
		"cohesion-catalog-v2",
		"cohesion-inputs-v2",
		"cohesion-sources-v2",
		"cohesion-delivery-evidence-v1",
		"cohesion-authorization-registry-v1",
		"cohesion-authorization-record-v1",
		"cohesion-residual-inventory-v1",
		"cohesion-residual-register-v1",
		"cohesion-gate-policy-v1",
		"cohesion-go-toolchains-v1",
		"cohesion-go-official-downloads-oracle-v1",
		"cohesion-go-official-downloads-oracle-review-v1",
		"cohesion-engineering-identities-v1",
		"cohesion-schema-provenance-v1",
		"cohesion-schema-v3-decision-review-v1",
		"cohesion-schema-v3-decision-freeze-v1",
		"cohesion-contract-freeze-v1",
		"cohesion-contract-review-v1",
		"cohesion-diagnostic-v1",
	} {
		base := base
		t.Run(base, func(t *testing.T) {
			t.Parallel()

			row := map[string]any{
				"kind":                            "new-contract",
				"decision_sha256":                 "sha256:81f90106d873e8635f4aa4a8d120d5a074fdc0eb2146c767635e9aa9879d5776",
				"decision_freeze":                 map[string]any{"repository": "github.com/faustbrian/go-library-tools", "release": "v1.5.6", "release_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.5.6/cohesion-schema-v3-decision-freeze.json", "asset": "cohesion-schema-v3-decision-freeze.json", "schema_id": "urn:golib:cohesion:schema-v3-decision-freeze:v1", "bytes_sha256": digest},
				"goal_contract_release":           map[string]any{"goal_id": "golib-cohesion-v1", "requirements_sha256": "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c", "repository": "github.com/faustbrian/go-library-tools", "path": "docs/ecosystem/goals/cohesion-v1.md", "release": "v1.5.5", "tag_object_sha": "6e54f464376865a60c618cce8f07fe70bcfabed0", "peeled_commit": "66d2874dc98afe43dd7dad379d6f5e7d1b613111"},
				"goal_contract_sha256":            "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c",
				"oracle_source":                   map[string]any{"repository": "github.com/faustbrian/go-library-tools", "path": "testdata/cohesion/forward-oracles/" + base + "-forward-oracle.json", "source_revision": commit, "bytes_sha256": digest},
				"oracle_review":                   map[string]any{"repository": "github.com/faustbrian/go-library-tools", "path": "testdata/cohesion/forward-oracle-reviews/" + base + "-forward-oracle-review.json", "source_revision": commit, "bytes_sha256": digest},
				"first_implementation_commit":     commit,
				"forward_oracle":                  map[string]any{"repository": "github.com/faustbrian/go-library-tools", "release": "v1.5.6", "release_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.5.6/" + base + "-forward-oracle.json", "asset": base + "-forward-oracle.json", "bytes_sha256": digest},
				"forward_oracle_sha256":           digest,
				"schema_path":                     "schema/" + base + ".schema.json",
				"schema_bytes_sha256":             digest,
				"schema_id":                       "https://github.com/faustbrian/go-library-tools/schema/" + base + ".schema.json",
				"resolved_reference_graph":        map[string]any{"repository": "github.com/faustbrian/go-library-tools", "release": "v1.5.6", "release_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.5.6/" + base + "-resolved-reference-graph.json", "asset": base + "-resolved-reference-graph.json", "bytes_sha256": digest},
				"resolved_reference_graph_sha256": digest,
			}
			if err := compiled.Validate(row); err != nil {
				t.Fatalf("valid reviewed oracle lineage rejected: %v", err)
			}

			wrong := cloneJSONValue(t, row)
			wrong["oracle_review"].(map[string]any)["path"] = "testdata/cohesion/forward-oracle-reviews/wrong-forward-oracle-review.json"
			if err := compiled.Validate(wrong); err == nil {
				t.Fatal("mismatched oracle review path accepted")
			}
		})
	}
}

func cloneJSONValue(t *testing.T, value map[string]any) map[string]any {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var clone map[string]any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func TestSchemaV3BootstrapControlAssetsMatchFrozenBytesAndSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		asset  string
		schema string
		digest string
	}{
		{"cohesion-contract-review.json", "cohesion-contract-review-v1.schema.json", "0793cd5175f84fdbafe4b13590d1cb48cd2450bcd2b0a4189986fcdf3e78df9a"},
		{"cohesion-contract-freeze.json", "cohesion-contract-freeze-v1.schema.json", "d7ecce05defd00a370a9adbfc11963d0265c000e724e77a1ab44f3743604e5f4"},
		{"cohesion-schema-v3-decision-review.json", "cohesion-schema-v3-decision-review-v1.schema.json", "c371d51fc8bf5b314d8c7cf51e8b96b7d4714397145cce09cc723f8370fa628d"},
		{"cohesion-schema-v3-decision-freeze.json", "cohesion-schema-v3-decision-freeze-v1.schema.json", "a523f5972f4d75061d935e31773a7042ebf47d3dfb4541843d478dcc4992851f"},
	}
	compiled := compileSchemaV3BootstrapSet(t)
	for _, test := range tests {
		test := test
		t.Run(test.asset, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join("..", "..", "release", test.asset))
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(data)
			if got := hex.EncodeToString(digest[:]); got != test.digest {
				t.Fatalf("asset SHA-256 = %s, want frozen %s", got, test.digest)
			}
			var document any
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatal(err)
			}
			identity := "https://github.com/faustbrian/go-library-tools/schema/" + test.schema
			if err := compiled[identity].Validate(document); err != nil {
				t.Fatalf("validate frozen asset: %v", err)
			}
		})
	}
}

func TestContractReviewSchemaAcceptsRejectedReviewWithFindings(t *testing.T) {
	t.Parallel()

	compiled := compileSchemaV3BootstrapSet(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "release", "cohesion-contract-review.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	review := document["review"].(map[string]any)
	review["outcome"] = "rejected"
	review["findings"] = []any{"contract mismatch"}
	identity := "https://github.com/faustbrian/go-library-tools/schema/cohesion-contract-review-v1.schema.json"
	if err := compiled[identity].Validate(document); err != nil {
		t.Fatalf("validate rejected contract review: %v", err)
	}
}

func compileSchemaV3BootstrapSet(t *testing.T) map[string]*jsonschema.Schema {
	t.Helper()

	files := []string{
		"cohesion-contract-review-v1.schema.json",
		"cohesion-contract-freeze-v1.schema.json",
		"cohesion-schema-v3-decision-review-v1.schema.json",
		"cohesion-schema-v3-decision-freeze-v1.schema.json",
	}
	compiler := jsonschema.NewCompiler()
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join("..", "..", "schema", file))
		if err != nil {
			t.Fatal(err)
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("decode %s: %v", file, err)
		}
		identity := "https://github.com/faustbrian/go-library-tools/schema/" + file
		if err := compiler.AddResource(identity, document); err != nil {
			t.Fatalf("add %s: %v", file, err)
		}
	}

	result := make(map[string]*jsonschema.Schema, len(files))
	for _, file := range files {
		identity := "https://github.com/faustbrian/go-library-tools/schema/" + file
		compiled, err := compiler.Compile(identity)
		if err != nil {
			t.Fatalf("compile %s: %v", file, err)
		}
		result[identity] = compiled
	}
	return result
}

func TestSchemaV3LeavesHistoricalUnversionedSchemaBytesUntouched(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"modules.schema.json":          "a72d03ea77f9134b516ef61426ba7fa1a95fd9505c68d5d2288aee6472b42cd7",
		"cohesion-catalog.schema.json": "674fcd9c3abd6c985de6b681883b4f822a510e89a84dd01e8d6249a56af3af49",
		"cohesion-inputs.schema.json":  "b8d7977be639117e349156bdd9d37ec475832151e0d56fc5159ea940cf4f5fca",
		"cohesion-sources.schema.json": "bb86d9335a5362000111ba38aecda145bc3ac0e3329719f003fe4c8b4684288a",
	}

	for file, wantDigest := range want {
		data, err := os.ReadFile(filepath.Join("..", "..", "schema", file))
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(data)
		if got := hex.EncodeToString(digest[:]); got != wantDigest {
			t.Fatalf("%s SHA-256 = %s, want historical %s", file, got, wantDigest)
		}
	}
}

func TestSchemaV3PublishesClosedHistoricalVersionedSchemas(t *testing.T) {
	t.Parallel()

	files := []string{
		"modules-v1.schema.json",
		"modules-v2.schema.json",
		"cohesion-catalog-v1.schema.json",
		"cohesion-inputs-v1.schema.json",
		"cohesion-sources-v1.schema.json",
	}
	compiler := jsonschema.NewCompiler()
	for _, file := range files {
		path := filepath.Join("..", "..", "schema", file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read versioned historical schema %s: %v", file, err)
		}
		var identity struct {
			ID                   string `json:"$id"`
			AdditionalProperties *bool  `json:"additionalProperties"`
		}
		if err := json.Unmarshal(data, &identity); err != nil {
			t.Fatalf("decode versioned historical schema %s: %v", file, err)
		}
		wantID := "https://github.com/faustbrian/go-library-tools/schema/" + file
		if identity.ID != wantID {
			t.Fatalf("%s $id = %q, want %q", file, identity.ID, wantID)
		}
		if identity.AdditionalProperties == nil || *identity.AdditionalProperties {
			t.Fatalf("%s top level is not closed", file)
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("parse versioned historical schema %s: %v", file, err)
		}
		if err := compiler.AddResource(wantID, document); err != nil {
			t.Fatalf("add versioned historical schema %s: %v", file, err)
		}
	}
	for _, file := range files {
		identity := "https://github.com/faustbrian/go-library-tools/schema/" + file
		if _, err := compiler.Compile(identity); err != nil {
			t.Fatalf("compile versioned historical schema %s: %v", file, err)
		}
	}
	modulesV1, err := compiler.Compile("https://github.com/faustbrian/go-library-tools/schema/modules-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	minimalHistorical := map[string]any{
		"schema_version": 1,
		"repository":     "github.com/faustbrian/example",
		"modules": []any{map[string]any{
			"directory":   ".",
			"module_path": "github.com/faustbrian/example",
			"packages":    []any{},
		}},
	}
	if err := modulesV1.Validate(minimalHistorical); err != nil {
		t.Fatalf("modules-v1 rejects released minimal manifest: %v", err)
	}
	nullableHistorical := map[string]any{
		"schema_version": 1,
		"repository":     "github.com/faustbrian/example",
		"go_version":     nil,
		"modules": []any{map[string]any{
			"directory":     ".",
			"go_version":    nil,
			"releasable":    nil,
			"gates":         nil,
			"test_tags":     nil,
			"packages":      []any{nil},
			"goal_evidence": []any{nil},
			"family_order":  nil,
			"provenance":    nil,
		}},
	}
	encoded, err := json.Marshal(nullableHistorical)
	if err != nil {
		t.Fatal(err)
	}
	var legacyDecoded inventory.Inventory
	if err := json.Unmarshal(encoded, &legacyDecoded); err != nil {
		t.Fatalf("released encoding/json rejects nullable manifest: %v", err)
	}
	if err := modulesV1.Validate(nullableHistorical); err != nil {
		t.Fatalf("modules-v1 rejects released nullable manifest: %v", err)
	}
}

func TestHistoricalSchemaSupportAssetsCoverEveryVersionedSnapshot(t *testing.T) {
	t.Parallel()

	bases := []string{
		"modules-v1",
		"modules-v2",
		"cohesion-catalog-v1",
		"cohesion-inputs-v1",
		"cohesion-sources-v1",
	}
	suffixes := []string{
		"historical-baseline.json",
		"accepted-corpus.json",
		"rejected-corpus.json",
		"resolved-reference-graph.json",
	}
	for _, base := range bases {
		for _, suffix := range suffixes {
			path := filepath.Join("..", "..", "release", base+"-"+suffix)
			if _, err := os.ReadFile(path); err != nil {
				t.Errorf("read historical support asset %s: %v", path, err)
			}
		}
	}
	for _, path := range []string{
		// v1 remains a historical schema contract; its superseded aggregate
		// control asset is intentionally not required for v2 migration.
		filepath.Join("..", "..", "schema", "cohesion-schema-provenance-v1.schema.json"),
		filepath.Join("..", "..", "schema", "cohesion-schema-provenance-v2.schema.json"),
		filepath.Join("..", "..", "release", "cohesion-schema-provenance-v2-resolved-reference-graph.json"),
	} {
		if _, err := os.ReadFile(path); err != nil {
			t.Errorf("read historical provenance control %s: %v", path, err)
		}
	}
}
