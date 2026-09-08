package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/jcs"
)

func TestResolvedReferenceGraphIncludesInternalAndTransitiveExternalReferences(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	a := []byte(`{"$id":"https://github.com/faustbrian/go-library-tools/schema/a.schema.json","$ref":"b.schema.json#/$defs/b","allOf":[{"$ref":"#/$defs/a"}],"$defs":{"a":{"type":"string"}}}`)
	b := []byte(`{"$id":"https://github.com/faustbrian/go-library-tools/schema/b.schema.json","$defs":{"b":{"type":"string"}}}`)
	if err := os.WriteFile(filepath.Join(root, "schema", "a.schema.json"), a, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "schema", "b.schema.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := resolvedReferenceGraph(filepath.Join(root, "schema"), "schema/a.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	bDigest := sha256.Sum256(b)
	wantFragment := `"resolved_schema_bytes_sha256":"sha256:` + hex.EncodeToString(bDigest[:]) + `","resolved_schema_path":"schema/b.schema.json"`
	if !contains(string(got), wantFragment) || !contains(string(got), `"schema_path":"schema/a.schema.json"`) || !contains(string(got), `"schema_path":"schema/b.schema.json"`) {
		t.Fatalf("graph = %s", got)
	}
}

func TestPublishedResolvedReferenceGraphsMatchEveryVersionedSchema(t *testing.T) {
	t.Parallel()

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	schemaRoot := filepath.Join(repositoryRoot, "schema")
	paths, err := filepath.Glob(filepath.Join(schemaRoot, "*-v*.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 26 {
		t.Fatalf("versioned schema count = %d, want 26", len(paths))
	}
	for _, path := range paths {
		name := filepath.Base(path)
		want, err := resolvedReferenceGraph(schemaRoot, filepath.ToSlash(filepath.Join("schema", name)))
		if err != nil {
			t.Fatalf("generate %s: %v", name, err)
		}
		asset := filepath.Join(repositoryRoot, "release", strings.TrimSuffix(name, ".schema.json")+"-resolved-reference-graph.json")
		got, err := os.ReadFile(asset)
		if err != nil {
			t.Fatalf("read %s: %v", asset, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s is stale", filepath.Base(asset))
		}
	}
}

func TestResolvedReferenceGraphRejectsDuplicateMembers(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"$id":"https://github.com/faustbrian/go-library-tools/schema/a.schema.json","$id":"https://github.com/faustbrian/go-library-tools/schema/a.schema.json"}`)
	if err := os.WriteFile(filepath.Join(root, "schema", "a.schema.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := resolvedReferenceGraph(filepath.Join(root, "schema"), "schema/a.schema.json"); err == nil {
		t.Fatal("duplicate member accepted")
	}
}

func TestHistoricalRewriteChangesOnlyRootIdentityAndReferences(t *testing.T) {
	t.Parallel()
	input := []byte(`{"$id":"https://github.com/faustbrian/go-library-tools/schema/old.schema.json","description":"old.schema.json","$defs":{"x":{"$ref":"old.schema.json#/$defs/x","const":"old.schema.json"},"fragment":{"$ref":"#/$defs/old.schema.json"}}}`)
	got, err := rewriteHistoricalSchema(input, map[string]string{"old.schema.json": "new.schema.json"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !contains(text, `"$id":"https://github.com/faustbrian/go-library-tools/schema/new.schema.json"`) || !contains(text, `"$ref":"new.schema.json#/$defs/x"`) {
		t.Fatalf("authorized fields not rewritten: %s", text)
	}
	if !contains(text, `"description":"old.schema.json"`) || !contains(text, `"const":"old.schema.json"`) {
		t.Fatalf("non-reference string changed: %s", text)
	}
	if !contains(text, `"$ref":"#/$defs/old.schema.json"`) {
		t.Fatalf("reference fragment text changed: %s", text)
	}
}

func TestGraphBudgetExactCumulativeBoundaries(t *testing.T) {
	t.Parallel()
	const requiredAllocationBytes = int64(72 << 20)
	if maximumGraphAllocationBytes != requiredAllocationBytes {
		t.Fatalf("graph allocation budget = %d, want exact 72 MiB", maximumGraphAllocationBytes)
	}
	for _, size := range []int64{maximumGraphSchemaBytes - 1, maximumGraphSchemaBytes} {
		budget := &graphBudget{}
		if err := budget.chargeSchema(size); err != nil {
			t.Fatalf("schema byte budget %d rejected: %v", size, err)
		}
	}
	if err := (&graphBudget{}).chargeSchema(maximumGraphSchemaBytes + 1); err == nil {
		t.Fatal("above cumulative schema byte budget accepted")
	}
	for _, size := range []int64{requiredAllocationBytes - 1, requiredAllocationBytes} {
		budget := &graphBudget{}
		if err := budget.chargeAllocation(size); err != nil {
			t.Fatalf("allocation budget %d rejected: %v", size, err)
		}
	}
	if err := (&graphBudget{}).chargeAllocation(requiredAllocationBytes + 1); err == nil {
		t.Fatal("above cumulative allocation budget accepted")
	}
}

func TestResolvedReferenceGraphAllocationBoundaryIsEndToEnd(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"$id":"https://github.com/faustbrian/go-library-tools/schema/a.schema.json","items":[{"type":"string"},{"type":"integer"}]}`)
	if err := os.WriteFile(filepath.Join(root, "schema", "a.schema.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	measurement := &graphBudget{allocationLimit: maximumGraphAllocationBytes}
	want, err := resolvedReferenceGraphWithBudget(filepath.Join(root, "schema"), "schema/a.schema.json", measurement)
	if err != nil {
		t.Fatal(err)
	}
	if measurement.allocationBytes == 0 {
		t.Fatal("graph reported no allocation capacity")
	}
	for _, test := range []struct {
		name     string
		limit    int64
		accepted bool
	}{
		{name: "below", limit: measurement.allocationBytes - 1, accepted: false},
		{name: "exact", limit: measurement.allocationBytes, accepted: true},
		{name: "above", limit: measurement.allocationBytes + 1, accepted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			budget := &graphBudget{allocationLimit: test.limit}
			got, err := resolvedReferenceGraphWithBudget(filepath.Join(root, "schema"), "schema/a.schema.json", budget)
			if test.accepted != (err == nil) {
				t.Fatalf("limit=%d accepted=%v error=%v", test.limit, test.accepted, err)
			}
			if err == nil && !bytes.Equal(got, want) {
				t.Fatalf("graph = %s, want %s", got, want)
			}
		})
	}
}

func TestResolvedReferenceGraphEnforcesCanonicalJSONDepthBoundary(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		arrays   int
		accepted bool
	}{
		{name: "below", arrays: 62, accepted: true},
		{name: "exact", arrays: 63, accepted: true},
		{name: "above", arrays: 64, accepted: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			schema := `{"$id":"https://github.com/faustbrian/go-library-tools/schema/a.schema.json","x":` + strings.Repeat("[", test.arrays) + `0` + strings.Repeat("]", test.arrays) + `}`
			if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "schema", "a.schema.json"), []byte(schema), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := resolvedReferenceGraph(filepath.Join(root, "schema"), "schema/a.schema.json")
			if test.accepted != (err == nil) {
				t.Fatalf("accepted=%v error=%v", test.accepted, err)
			}
		})
	}
}

func TestResolvedReferenceGraphEnforcesCanonicalJSONStructuralBoundaries(t *testing.T) {
	identity := `"$id":"https://github.com/faustbrian/go-library-tools/schema/a.schema.json"`
	object := func(size int) string {
		var value strings.Builder
		value.WriteByte('{')
		for index := 0; index < size; index++ {
			if index != 0 {
				value.WriteByte(',')
			}
			value.WriteString(`"k`)
			value.WriteString(strconv.Itoa(index))
			value.WriteString(`":0`)
		}
		value.WriteByte('}')
		return value.String()
	}
	array := func(size int) string {
		return `[` + strings.TrimSuffix(strings.Repeat("0,", size), ",") + `]`
	}
	cases := []struct {
		name     string
		value    string
		accepted bool
	}{
		{name: "string-below", value: `"` + strings.Repeat("x", (1<<20)-1) + `"`, accepted: true},
		{name: "string-exact", value: `"` + strings.Repeat("x", 1<<20) + `"`, accepted: true},
		{name: "string-above", value: `"` + strings.Repeat("x", (1<<20)+1) + `"`, accepted: false},
		{name: "array-below", value: array(4095), accepted: true},
		{name: "array-exact", value: array(4096), accepted: true},
		{name: "array-above", value: array(4097), accepted: false},
		{name: "object-below", value: object(4095), accepted: true},
		{name: "object-exact", value: object(4096), accepted: true},
		{name: "object-above", value: object(4097), accepted: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := testResolveSchema(t, []byte(`{`+identity+`,"x":`+test.value+`}`))
			if test.accepted != (err == nil) {
				t.Fatalf("accepted=%v error=%v", test.accepted, err)
			}
		})
	}
	t.Run("file", func(t *testing.T) {
		base := []byte(`{` + identity + `}`)
		for _, size := range []int{maximumSchemaBytes - 1, maximumSchemaBytes} {
			data := append(append([]byte(nil), base...), bytes.Repeat([]byte(" "), size-len(base))...)
			if err := testResolveSchema(t, data); err != nil {
				t.Fatalf("file size %d rejected: %v", size, err)
			}
		}
		data := append(append([]byte(nil), base...), bytes.Repeat([]byte(" "), maximumSchemaBytes+1-len(base))...)
		if err := testResolveSchema(t, data); err == nil {
			t.Fatal("above maximum schema file accepted")
		}
	})
}

func TestGraphReferenceAndNodeCountBoundaries(t *testing.T) {
	t.Parallel()
	for _, size := range []int{maximumGraphUniqueReferences - 1, maximumGraphUniqueReferences} {
		values := make([]any, 0, size)
		for index := 0; index < size; index++ {
			values = append(values, jcs.Object{{Name: "$ref", Value: "schema-" + strconv.Itoa(index) + ".schema.json"}})
		}
		if _, err := collectReferences(values, &graphBudget{}); err != nil {
			t.Fatalf("reference count %d rejected: %v", size, err)
		}
	}
	values := make([]any, 0, maximumGraphUniqueReferences+1)
	for index := 0; index <= maximumGraphUniqueReferences; index++ {
		values = append(values, jcs.Object{{Name: "$ref", Value: "schema-" + strconv.Itoa(index) + ".schema.json"}})
	}
	if _, err := collectReferences(values, &graphBudget{}); err == nil {
		t.Fatal("above maximum reference count accepted")
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "schema", "a.schema.json")
	data := []byte(`{"$id":"https://github.com/faustbrian/go-library-tools/schema/a.schema.json"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	resolver := &graphResolver{repositoryRoot: root, schemaRoot: filepath.Join(root, "schema"), loaded: make([]*schemaDocument, 0, maximumGraphNodes)}
	for index := 0; index < maximumGraphNodes-1; index++ {
		resolver.loaded = append(resolver.loaded, &schemaDocument{path: "existing-" + strconv.Itoa(index)})
	}
	if _, err := resolver.load("schema/a.schema.json"); err != nil {
		t.Fatalf("exact maximum node count rejected: %v", err)
	}
	if _, err := resolver.load("schema/another.schema.json"); err == nil || !strings.Contains(err.Error(), "schema nodes") {
		t.Fatalf("above maximum node count not rejected by count bound: %v", err)
	}
}

func testResolveSchema(t *testing.T, data []byte) error {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "schema", "a.schema.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := resolvedReferenceGraph(filepath.Join(root, "schema"), "schema/a.schema.json")
	return err
}

func contains(value, fragment string) bool {
	for index := 0; index+len(fragment) <= len(value); index++ {
		if value[index:index+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
