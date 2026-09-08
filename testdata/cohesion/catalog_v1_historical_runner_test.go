package cohesion

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

// This file is injected into internal/cohesion at b06e903. It emits raw,
// independently digestible observations; it intentionally computes no oracle
// digests itself.
type historicalObservation struct {
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

type projectInvocation struct {
	Catalog  inventory.Inventory `json:"catalog"`
	View     string              `json:"view"`
	Identity Identity            `json:"identity"`
}

func TestHistoricalCatalogV1Runner(t *testing.T) {
	output := os.Getenv("HISTORICAL_OBSERVATION_PATH")
	if output == "" {
		t.Skip("set HISTORICAL_OBSERVATION_PATH to run the injected historical runner")
	}

	identity := historicalIdentityUnpublished()
	catalog := inventory.Inventory{
		SchemaVersion: 2,
		Repository:    "github.com/faustbrian/go-example",
		GoVersion:     "1.27.0",
		Modules: []inventory.Module{
			historicalCatalogModule("github.com/faustbrian/go-example/zeta", "public library", true, "tooling", "active"),
			historicalCatalogModule("github.com/faustbrian/go-example/alpha", "adapter", true, "foundations", "deprecated"),
			historicalCatalogModule("github.com/faustbrian/go-example/planned", "public library", true, "service-edge", "planned"),
			historicalCatalogModule("github.com/faustbrian/go-example/tool", "public tool", true, "tooling", "active"),
			historicalCatalogModule("github.com/faustbrian/go-example/internal", "fixture", false, "", ""),
		},
	}

	observations := make([]historicalObservation, 0, 9)
	for _, test := range []struct {
		id       string
		view     string
		identity Identity
	}{
		{id: "project-consumer-unpublished", view: "consumer", identity: identity},
		{id: "project-engineering-unpublished", view: "engineering", identity: identity},
		{id: "project-engineering-published", view: "engineering", identity: Identity{
			DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
			SourceIdentity: "v1.2.3", ToolingVersion: "v1.2.3", PublicationStatus: "published",
		}},
	} {
		invocation := projectInvocation{Catalog: catalog, View: test.view, Identity: test.identity}
		envelope, err := Project(invocation.Catalog, invocation.View, invocation.Identity)
		var decoded, emitted []byte
		if err == nil {
			// Decoded meaning is the returned Envelope marshalled as JSON. Emitted
			// meaning independently follows the historical CLI's indented
			// json.Encoder path, including its terminating newline.
			decoded = mustHistoricalJSON(t, envelope)
			emitted = historicalCatalogJSONEmission(t, envelope)
		}
		observation := observeHistorical(test.id, historicalProjectEntryPoint, mustHistoricalJSON(t, invocation), decoded, "json", emitted, "json", err, classifyCatalogV1Error)
		observations = append(observations, observation)
	}

	for _, test := range []struct {
		id       string
		view     string
		identity Identity
	}{
		{id: "project-reject-view", view: "invalid", identity: identity},
		{id: "project-reject-design-identity", view: "engineering", identity: Identity{
			DesignLanguageVersion: "1.0", DesignLanguageSHA256: "invalid", SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished",
		}},
		{id: "project-reject-publication-identity", view: "engineering", identity: Identity{
			DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "v1.2.3", PublicationStatus: "published",
		}},
	} {
		invocation := projectInvocation{Catalog: catalog, View: test.view, Identity: test.identity}
		_, err := Project(invocation.Catalog, invocation.View, invocation.Identity)
		observations = append(observations, observeHistorical(test.id, historicalProjectEntryPoint, mustHistoricalJSON(t, invocation), nil, "", nil, "", err, classifyCatalogV1Error))
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

const historicalProjectEntryPoint = "internal/cohesion/catalog.go#Project(catalog inventory.Inventory, view string, identity Identity) (Envelope, error)"

func historicalIdentityUnpublished() Identity {
	return Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished",
	}
}

func historicalCatalogModule(path, kind string, releasable bool, family, lifecycle string) inventory.Module {
	module := inventory.Module{Directory: path, ModulePath: path, GoVersion: "1.27.0", Kind: kind, Releasable: releasable}
	if family != "" {
		module.Cohesion = &inventory.Cohesion{Family: family, LifecycleStatus: lifecycle, Responsibility: "Responsibility for " + path + "."}
	}
	return module
}

func observeHistorical(caseID, entryPoint string, input, decoded []byte, decodedKind string, emitted []byte, emittedKind string, err error, classify func(error) string) historicalObservation {
	result := historicalObservation{CaseID: caseID, EntryPoint: entryPoint, InputBase64: base64.StdEncoding.EncodeToString(input)}
	if err != nil {
		result.Outcome = "rejected"
		result.ErrorClass = classify(err)
		return result
	}
	result.Outcome = "accepted"
	result.DecodedValueKind = decodedKind
	result.DecodedValueBase64 = base64.StdEncoding.EncodeToString(decoded)
	result.EmittedValueKind = emittedKind
	result.EmittedValueBase64 = base64.StdEncoding.EncodeToString(emitted)
	return result
}

func historicalCatalogJSONEmission(t *testing.T, value any) []byte {
	t.Helper()
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func mustHistoricalJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func classifyCatalogV1Error(err error) string {
	message := err.Error()
	switch {
	case strings.Contains(message, "catalog view must be consumer or engineering"):
		return "catalog-view"
	case strings.Contains(message, "design-language identity is invalid"):
		return "design-language-identity"
	case strings.Contains(message, "tooling publication identity is invalid"):
		return "tooling-publication-identity"
	case strings.Contains(message, "catalog modules do not match the envelope view"):
		return "catalog-module-shape"
	default:
		return "unexpected:" + message
	}
}
