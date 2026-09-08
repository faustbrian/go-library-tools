package cohesion

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

// This file is injected into internal/cohesion at 903eb53. It emits raw,
// independently digestible observations and deliberately does not compute the
// final RFC 8785 corpus digests.
type historicalAggregateObservation struct {
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

type aggregateInvocationBundle struct {
	Identity    Identity                   `json:"identity"`
	Manifest    aggregateInputManifest     `json:"manifest"`
	Projections map[string]json.RawMessage `json:"projections"`
}

type aggregateValue struct {
	ConsumerJSON              json.RawMessage `json:"consumer_json"`
	ConsumerMarkdownBase64    string          `json:"consumer_markdown_base64"`
	EngineeringJSON           json.RawMessage `json:"engineering_json"`
	EngineeringMarkdownBase64 string          `json:"engineering_markdown_base64"`
}

func TestHistoricalInputsV1Runner(t *testing.T) {
	output := os.Getenv("HISTORICAL_OBSERVATION_PATH")
	if output == "" {
		t.Skip("set HISTORICAL_OBSERVATION_PATH to run the injected historical runner")
	}

	valid := historicalAggregateBundle(t, []string{"github.com/faustbrian/go-alpha", "github.com/faustbrian/go-beta"})
	cases := []struct {
		id     string
		bundle aggregateInvocationBundle
	}{
		{id: "aggregate-accept-two-sorted-repositories", bundle: valid},
		{id: "aggregate-reject-empty-repositories", bundle: mutateAggregateBundle(t, valid, func(bundle *aggregateInvocationBundle) {
			bundle.Manifest.Repositories = nil
		})},
		{id: "aggregate-reject-unsorted-repositories", bundle: mutateAggregateBundle(t, valid, func(bundle *aggregateInvocationBundle) {
			bundle.Manifest.Repositories[0], bundle.Manifest.Repositories[1] = bundle.Manifest.Repositories[1], bundle.Manifest.Repositories[0]
		})},
		{id: "aggregate-reject-unsafe-projection-path", bundle: mutateAggregateBundle(t, valid, func(bundle *aggregateInvocationBundle) {
			bundle.Manifest.Repositories[0].Projection = "safe/../alpha.json"
		})},
		{id: "aggregate-reject-projection-digest", bundle: mutateAggregateBundle(t, valid, func(bundle *aggregateInvocationBundle) {
			bundle.Manifest.Repositories[0].SHA256 = strings.Repeat("f", 64)
		})},
		{id: "aggregate-reject-projection-schema", bundle: mutateAggregateBundle(t, valid, func(bundle *aggregateInvocationBundle) {
			name := bundle.Manifest.Repositories[0].Projection
			var document map[string]any
			if err := json.Unmarshal(bundle.Projections[name], &document); err != nil {
				t.Fatal(err)
			}
			document["unexpected"] = true
			bundle.Projections[name] = mustAggregateJSON(t, document)
			bundle.Manifest.Repositories[0].SHA256 = bareSHA256(bundle.Projections[name])
		})},
		{id: "aggregate-reject-projection-identity", bundle: mutateAggregateBundle(t, valid, func(bundle *aggregateInvocationBundle) {
			name := bundle.Manifest.Repositories[0].Projection
			var document map[string]any
			if err := json.Unmarshal(bundle.Projections[name], &document); err != nil {
				t.Fatal(err)
			}
			document["repository"] = "github.com/faustbrian/go-wrong"
			bundle.Projections[name] = mustAggregateJSON(t, document)
			bundle.Manifest.Repositories[0].SHA256 = bareSHA256(bundle.Projections[name])
		})},
	}

	observations := make([]historicalAggregateObservation, 0, len(cases)+2)
	for _, test := range cases {
		input := mustAggregateJSON(t, test.bundle)
		artifacts, emitted, err := runHistoricalAggregateBundle(t, test.bundle)
		observation := historicalAggregateObservation{
			CaseID: test.id, EntryPoint: historicalAggregateEntryPoint,
			InputBase64: base64.StdEncoding.EncodeToString(input),
		}
		if err != nil {
			observation.Outcome = "rejected"
			observation.ErrorClass = classifyInputsV1Error(err)
		} else {
			observation.Outcome = "accepted"
			// Decoded meaning is an independent JSON representation of the
			// Artifacts returned by Aggregate. JSON artifacts remain JSON values;
			// Markdown artifacts retain their exact bytes as padded base64.
			value := aggregateValue{
				ConsumerJSON: artifacts.ConsumerJSON, ConsumerMarkdownBase64: base64.StdEncoding.EncodeToString(artifacts.ConsumerMarkdown),
				EngineeringJSON: artifacts.EngineeringJSON, EngineeringMarkdownBase64: base64.StdEncoding.EncodeToString(artifacts.EngineeringMarkdown),
			}
			observation.DecodedValueKind = "json"
			observation.DecodedValueBase64 = base64.StdEncoding.EncodeToString(mustAggregateJSON(t, value))
			// Emitted meaning is derived separately by GenerateAggregate, then by
			// reopening all four published files. It is not copied from the
			// Artifacts return value.
			observation.EmittedValueKind = "json"
			observation.EmittedValueBase64 = base64.StdEncoding.EncodeToString(emitted)
		}
		observations = append(observations, observation)
	}

	invalidIdentity := valid
	invalidIdentity.Identity.DesignLanguageSHA256 = "invalid"
	_, _, err := runHistoricalAggregateBundle(t, invalidIdentity)
	observations = append(observations, historicalAggregateObservation{
		CaseID: "aggregate-reject-generator-identity", EntryPoint: historicalAggregateEntryPoint,
		InputBase64: base64.StdEncoding.EncodeToString(mustAggregateJSON(t, invalidIdentity)),
		Outcome:     "rejected", ErrorClass: classifyInputsV1Error(err),
	})

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

const historicalAggregateEntryPoint = "internal/cohesion/aggregate.go#Aggregate(inputsPath string, identity Identity) (Artifacts, error)"

func historicalAggregateBundle(t *testing.T, repositories []string) aggregateInvocationBundle {
	t.Helper()
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	bundle := aggregateInvocationBundle{
		Identity:    identity,
		Manifest:    aggregateInputManifest{SchemaVersion: 1, DesignLanguage: aggregateDesignLanguage{Version: "1.0", SHA256: strings.Repeat("a", 64)}},
		Projections: make(map[string]json.RawMessage, len(repositories)),
	}
	for _, repository := range repositories {
		module := historicalAggregateModule(repository)
		envelope, err := Project(inventory.Inventory{Repository: repository, Modules: []inventory.Module{module}}, "engineering", identity)
		if err != nil {
			t.Fatal(err)
		}
		projection := mustAggregateJSON(t, envelope)
		name := strings.TrimPrefix(repository, "github.com/faustbrian/") + ".json"
		bundle.Projections[name] = projection
		bundle.Manifest.Repositories = append(bundle.Manifest.Repositories, aggregateRepositoryInput{
			Repository: repository, Projection: name, SHA256: bareSHA256(projection),
		})
	}
	return bundle
}

func historicalAggregateModule(repository string) inventory.Module {
	readme, changelog := "README.md", "CHANGELOG.md"
	api, ecosystem := "https://pkg.go.dev/"+repository, "https://example.com/ecosystem"
	return inventory.Module{
		Directory: ".", ModulePath: repository, GoVersion: "1.27.0",
		Kind: "public library", Purpose: "Historical aggregate fixture.", Lifecycle: "stable",
		Releasable: true, Version: "1.0.0", TagPrefix: "v", Gates: map[string]bool{},
		TestTags: []string{}, BuildTags: []string{}, RequiredServices: []string{},
		ExternalRuntimeDependencies: []string{}, InteroperabilityTools: []string{}, ConformanceCorpora: []string{},
		Specifications: []string{}, OwnedDependencies: []string{}, ReverseOwnedDependencies: []string{},
		Packages: []inventory.Package{{ModuleDirectory: ".", Directory: ".", Name: "library", ImportPath: repository, Kind: "public", BuildTags: []string{}}},
		Family:   "foundations", GoalFiles: []string{}, GoalEvidence: []inventory.GoalEvidence{}, Provenance: []byte("[]"),
		Cohesion: &inventory.Cohesion{
			Family: "foundations", SecondaryCapabilities: []string{"testing-and-conformance"},
			Responsibility: "Provide historical aggregate fixtures.", NonGoals: []string{"Own application state."},
			PublicPackageIdentifier: "library", PrimaryEntryPackages: []string{repository}, PackageSelection: map[string]string{},
			LifecycleStatus: "active", Maturity: "stable", ConstructionStyles: []string{"plain-function"}, LifecycleStyles: []string{"stateless"},
			Ownership:                 inventory.Ownership{Configuration: "caller", MutableInputs: []string{"copy"}, RuntimeResources: "none", BackgroundWork: "none"},
			OptionalOwnedDependencies: []string{}, Adapters: []string{}, Companions: []string{},
			SupportedGo:        inventory.SupportedGo{Minimum: "1.27.0", Tested: []string{"1.27.0"}},
			SupportedPlatforms: []string{"portable-go"}, SupportedBackends: []string{}, SupportedProtocols: []string{},
			Documentation:              inventory.Documentation{README: &readme, API: &api, Changelog: &changelog, PkgGoDev: &api, EcosystemIndex: &ecosystem},
			KnownGoodCompatibilitySets: []string{}, Delivery: inventory.Delivery{Implementation: "in-progress", Hardening: "in-progress", Release: "in-progress"},
		},
	}
}

func mutateAggregateBundle(t *testing.T, source aggregateInvocationBundle, mutate func(*aggregateInvocationBundle)) aggregateInvocationBundle {
	t.Helper()
	var clone aggregateInvocationBundle
	if err := json.Unmarshal(mustAggregateJSON(t, source), &clone); err != nil {
		t.Fatal(err)
	}
	mutate(&clone)
	return clone
}

func runHistoricalAggregateBundle(t *testing.T, bundle aggregateInvocationBundle) (Artifacts, []byte, error) {
	t.Helper()
	root := t.TempDir()
	for name, content := range bundle.Projections {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manifestPath := filepath.Join(root, "inputs.json")
	if err := os.WriteFile(manifestPath, mustAggregateJSON(t, bundle.Manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	artifacts, err := Aggregate(manifestPath, bundle.Identity)
	if err != nil {
		return Artifacts{}, nil, err
	}
	output := filepath.Join(root, "published")
	if err := GenerateAggregate(manifestPath, output, bundle.Identity); err != nil {
		t.Fatalf("GenerateAggregate accepted Aggregate input: %v", err)
	}
	emitted := aggregateValue{
		ConsumerJSON:              readHistoricalAggregateJSON(t, filepath.Join(output, "catalog-consumer.json")),
		ConsumerMarkdownBase64:    base64.StdEncoding.EncodeToString(readHistoricalAggregateBytes(t, filepath.Join(output, "catalog-consumer.md"))),
		EngineeringJSON:           readHistoricalAggregateJSON(t, filepath.Join(output, "catalog-engineering.json")),
		EngineeringMarkdownBase64: base64.StdEncoding.EncodeToString(readHistoricalAggregateBytes(t, filepath.Join(output, "catalog-engineering.md"))),
	}
	return artifacts, mustAggregateJSON(t, emitted), nil
}

func readHistoricalAggregateBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readHistoricalAggregateJSON(t *testing.T, path string) json.RawMessage {
	t.Helper()
	data := readHistoricalAggregateBytes(t, path)
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode generated JSON artifact %s: %v", filepath.Base(path), err)
	}
	// Remarshal the independently parsed value so observation semantics are
	// JSON rather than the generator's lexical whitespace.
	return mustAggregateJSON(t, value)
}

func mustAggregateJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func bareSHA256(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func classifyInputsV1Error(err error) string {
	if err == nil {
		return "unexpected-success"
	}
	message := err.Error()
	switch {
	case strings.Contains(message, "design-language identity is invalid"):
		return "generator-design-language-identity"
	case strings.Contains(message, "tooling publication identity is invalid"):
		return "generator-publication-identity"
	case strings.Contains(message, "read cohesion aggregation inputs"):
		return "read-inputs"
	case strings.Contains(message, "aggregation input schema is invalid"):
		return "inputs-schema"
	case strings.Contains(message, "requires schema_version 1"):
		return "inputs-schema-version"
	case strings.Contains(message, "design-language identity does not match"):
		return "inputs-design-language-identity"
	case strings.Contains(message, "repositories must be unique and sorted"):
		return "repository-order"
	case strings.Contains(message, "safe manifest-relative path"):
		return "projection-path"
	case strings.Contains(message, "projection digest does not match"):
		return "projection-digest"
	case strings.Contains(message, "projection schema is invalid"):
		return "projection-schema"
	case strings.Contains(message, "projection identity does not match"):
		return "projection-envelope-identity"
	case strings.Contains(message, "projection generator identity does not match"):
		return "projection-generator-identity"
	case strings.Contains(message, "contains no modules"):
		return "projection-empty-modules"
	case strings.Contains(message, "module repository identity does not match"):
		return "module-repository-identity"
	case strings.Contains(message, "module path identity does not match"):
		return "module-path-identity"
	case strings.Contains(message, "duplicate cohesion module"):
		return "duplicate-module"
	default:
		return "unexpected:" + message
	}
}
