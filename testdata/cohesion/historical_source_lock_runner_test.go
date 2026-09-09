package cohesion

// This file is an injected historical runner. Copy it into internal/cohesion
// in a clean checkout of commit 89536ffa95c8d665537c23e05566e1225b0e36a6,
// then run only TestHistoricalSourceLockObservationRunner with
// HISTORICAL_OBSERVATION_PATH set to a nonexistent file beneath an existing
// task-owned, non-symlink directory.

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const historicalSourceCommit = "89536ffa95c8d665537c23e05566e1225b0e36a6"
const historicalSourceLockSHA256 = "c86bb9b2ee8c4f7464e378eb2025b1d3637f2fc4ad2fb7df1bf689a0c3e2246f"
const loadSourceLockEntryPoint = "internal/cohesion/source_lock.go#LoadSourceLock(path string) (SourceLock, error)"
const verifyPolicyEntryPoint = "internal/cohesion/source_lock.go#SourceLock.VerifyPolicy(repository, version, checksumsSHA256 string) error"

// historicalObservation is a raw, non-authoritative observation. The separate
// value encodings let the later independent extractor derive semantic corpus
// rows without this historical runner computing corpus hashes or baselines.
// Accepted rows carry both value pairs; rejected rows carry only ErrorClass.
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

// verifyPolicyInvocation is the runner's deterministic byte representation of
// VerifyPolicy's three non-receiver arguments. It is an observation envelope,
// not a historical product artifact.
type verifyPolicyInvocation struct {
	ChecksumsSHA256 string `json:"checksums_sha256"`
	Repository      string `json:"repository"`
	Version         string `json:"version"`
}

type sourceLockInvocation struct {
	ContentBase64 string `json:"content_base64"`
	PadToBytes    int    `json:"pad_to_bytes,omitempty"`
}

func TestHistoricalSourceLockObservationRunner(t *testing.T) {
	observationPath := os.Getenv("HISTORICAL_OBSERVATION_PATH")
	if observationPath == "" {
		t.Skip("set HISTORICAL_OBSERVATION_PATH to run the injected historical observer")
	}
	if !filepath.IsAbs(observationPath) {
		t.Fatal("HISTORICAL_OBSERVATION_PATH must be absolute")
	}
	parent := filepath.Dir(filepath.Clean(observationPath))
	info, err := os.Lstat(parent)
	if err != nil {
		t.Fatalf("stat observation parent: %v", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("observation parent must be an existing non-symlink directory")
	}
	if _, err := os.Lstat(observationPath); !os.IsNotExist(err) {
		t.Fatalf("observation path must not already exist: %v", err)
	}
	historicalRequireExactSources(t)

	valid, err := os.ReadFile("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(valid)); got != historicalSourceLockSHA256 {
		t.Fatalf("unexpected historical source-lock bytes: %s", got)
	}

	var decoded map[string]any
	if err := json.Unmarshal(valid, &decoded); err != nil {
		t.Fatal(err)
	}
	repositories := decoded["repositories"].([]any)
	firstConsumer := repositories[0].(map[string]any)
	firstTooling := firstConsumer["tooling"].(map[string]any)
	consumerRepo := firstConsumer["repository"].(string)
	consumerVersion := firstTooling["version"].(string)
	consumerDigest := firstTooling["checksums_sha256"].(string)

	observations := make([]historicalObservation, 0, 20)
	observeLoadSized := func(caseID string, input []byte, padToBytes int) {
		invocation := sourceLockInvocation{ContentBase64: base64.StdEncoding.EncodeToString(input), PadToBytes: padToBytes}
		materialized := append([]byte(nil), input...)
		if padToBytes != 0 {
			materialized = append(materialized, []byte(strings.Repeat(" ", padToBytes-len(materialized)))...)
		}
		path := filepath.Join(t.TempDir(), "sources.json")
		if err := os.WriteFile(path, materialized, 0o600); err != nil {
			t.Fatal(err)
		}
		value, err := LoadSourceLock(path)
		if err != nil {
			observations = append(observations, historicalRejectedObservation(t, caseID, loadSourceLockEntryPoint, historicalMarshal(t, invocation), err))
			return
		}
		decodedBytes := historicalMarshal(t, value)
		// LoadSourceLock has no emitter. Its historical command-visible mapping is
		// the successful `cohesion sources check` stdout contract in cli.go.
		emittedBytes := historicalMarshal(t, fmt.Sprintf("cohesion sources check passed: %d repositories\n", len(value.Repositories)))
		observations = append(observations, historicalAcceptedObservation(caseID, loadSourceLockEntryPoint, historicalMarshal(t, invocation), "json", decodedBytes, "json", emittedBytes))
	}
	observeLoad := func(caseID string, input []byte) { observeLoadSized(caseID, input, 0) }

	observeLoad("load-reviewed", valid)
	observeLoad("load-surrounding-whitespace", append([]byte(" \n\t"), append(valid, []byte("\r\n ")...)...))
	observeLoad("load-duplicate-key-last-wins", append([]byte(`{"schema_version":999,`), valid[1:]...))
	observeLoad("load-escaped-lexeme", []byte(strings.Replace(string(valid), "github.com/faustbrian/go-adaptive-throttle", `github\u002ecom/faustbrian/go-adaptive-throttle`, 1)))
	observeLoad("load-bom", append([]byte{0xef, 0xbb, 0xbf}, valid...))
	observeLoad("load-malformed-json", []byte("{"))
	observeLoadSized("load-oversized", []byte(`{}`), int(maximumSourceLockSize)+1)
	observeLoad("load-trailing-value", append(append([]byte{}, valid...), []byte(` {}`)...))
	observeLoad("load-invalid-utf8-tail", append(append([]byte{}, valid...), 0xff))
	observeLoad("load-wrong-schema-version", []byte(strings.Replace(string(valid), `"schema_version": 1`, `"schema_version": 2`, 1)))
	observeLoad("load-count-mismatch", []byte(strings.Replace(string(valid), `"repository_count": 92`, `"repository_count": 91`, 1)))
	observeLoad("load-unknown-member", append([]byte(`{"unexpected":true,`), valid[1:]...))

	unsorted := historicalCloneJSON(t, decoded)
	unsortedRepos := unsorted["repositories"].([]any)
	unsortedRepos[0], unsortedRepos[1] = unsortedRepos[1], unsortedRepos[0]
	observeLoad("load-unsorted", historicalMarshal(t, unsorted))

	duplicate := historicalCloneJSON(t, decoded)
	duplicateRepos := duplicate["repositories"].([]any)
	duplicateRepos[1] = duplicateRepos[0]
	observeLoad("load-duplicate-repository", historicalMarshal(t, duplicate))

	missingRelease := historicalCloneJSON(t, decoded)
	for _, raw := range missingRelease["repositories"].([]any) {
		repository := raw.(map[string]any)
		if repository["repository"] == "github.com/faustbrian/go-library-tools" {
			repository["source"] = map[string]any{"commit": strings.Repeat("a", 40), "kind": "commit"}
			repository["tooling"] = map[string]any{"checksums_sha256": strings.Repeat("b", 64), "version": "v1.5.0"}
		}
	}
	observeLoad("load-missing-release-source", historicalMarshal(t, missingRelease))

	lock, err := LoadSourceLock("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	observePolicy := func(caseID string, invocation verifyPolicyInvocation) {
		input := historicalMarshal(t, invocation)
		err := lock.VerifyPolicy(invocation.Repository, invocation.Version, invocation.ChecksumsSHA256)
		if err != nil {
			observations = append(observations, historicalRejectedObservation(t, caseID, verifyPolicyEntryPoint, input, err))
			return
		}
		// A nil error is represented independently as JSON null. VerifyPolicy has
		// no emitter; its command-visible mapping is the successful verification
		// stdout contract in cli.go, not a remarshal of the API result.
		decodedBytes := historicalMarshal(t, nil)
		emittedBytes := historicalMarshal(t, "cohesion source policy verified: "+invocation.Repository+"\n")
		observations = append(observations, historicalAcceptedObservation(caseID, verifyPolicyEntryPoint, input, "json", decodedBytes, "json", emittedBytes))
	}
	observePolicy("policy-exact-consumer", verifyPolicyInvocation{consumerDigest, consumerRepo, consumerVersion})
	observePolicy("policy-release-source-ignores-identity", verifyPolicyInvocation{"", "github.com/faustbrian/go-library-tools", "v0.0.0"})
	observePolicy("policy-unknown-repository", verifyPolicyInvocation{consumerDigest, "github.com/faustbrian/go-unknown", consumerVersion})
	observePolicy("policy-wrong-version", verifyPolicyInvocation{consumerDigest, consumerRepo, "v9.9.9"})
	observePolicy("policy-wrong-digest", verifyPolicyInvocation{strings.Repeat("c", 64), consumerRepo, consumerVersion})

	sort.Slice(observations, func(i, j int) bool { return observations[i].CaseID < observations[j].CaseID })
	data, err := json.MarshalIndent(observations, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	historicalWriteExclusive(t, observationPath, append(data, '\n'))
}

func historicalRequireExactSources(t *testing.T) {
	t.Helper()
	for path, want := range map[string]string{
		"source_lock.go":              "a774383c75967a050aa0b1097c78e9ff70d0c53071f8f0c4f7935051c8683035",
		"schema.go":                   "3dba96afe750913712e6dee284ef9cbf1d6814f77d17adcf2c67fc09f068e80b",
		"sources_schema_generated.go": "48298a6d20de24769ca27e7415635c8273994feae7b24f9449d3626ce10f0e0e",
		"../../schema/cohesion-sources.schema.json": "bb86d9335a5362000111ba38aecda145bc3ac0e3329719f003fe4c8b4684288a",
		"../../release/cohesion-sources.json":       historicalSourceLockSHA256,
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read exact historical source %s: %v", path, err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != want {
			t.Fatalf("historical commit %s source %s SHA-256 = %s, want %s", historicalSourceCommit, path, got, want)
		}
	}
}

func historicalAcceptedObservation(caseID, entryPoint string, input []byte, decodedKind string, decoded []byte, emittedKind string, emitted []byte) historicalObservation {
	decodedBase64 := base64.StdEncoding.EncodeToString(decoded)
	emittedBase64 := base64.StdEncoding.EncodeToString(emitted)
	return historicalObservation{
		CaseID: caseID, DecodedValueBase64: decodedBase64, DecodedValueKind: decodedKind,
		EmittedValueBase64: emittedBase64, EmittedValueKind: emittedKind,
		EntryPoint: entryPoint, InputBase64: base64.StdEncoding.EncodeToString(input), Outcome: "accepted",
	}
}

func historicalRejectedObservation(t *testing.T, caseID, entryPoint string, input []byte, err error) historicalObservation {
	t.Helper()
	errorClass := historicalErrorClass(t, err)
	return historicalObservation{
		CaseID: caseID, EntryPoint: entryPoint, ErrorClass: errorClass,
		InputBase64: base64.StdEncoding.EncodeToString(input), Outcome: "rejected",
	}
}

func historicalErrorClass(t *testing.T, err error) string {
	t.Helper()
	message := err.Error()
	switch {
	case strings.Contains(message, "source lock schema is invalid"):
		return "schema-invalid"
	case strings.Contains(message, "exceed maximum size"):
		return "input-too-large"
	case strings.Contains(message, "decode cohesion source lock"):
		return "decode-invalid"
	case strings.Contains(message, "repository_count"):
		return "repository-count-mismatch"
	case strings.Contains(message, "strictly sorted"):
		return "repository-order-invalid"
	case strings.Contains(message, "exactly one release source"):
		return "release-source-count-invalid"
	case strings.Contains(message, "policy does not match"):
		return "policy-mismatch"
	case strings.Contains(message, "repository is not locked"):
		return "repository-not-locked"
	default:
		t.Fatalf("unclassified historical error: %v", err)
		return ""
	}
}

func historicalMarshal(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func historicalCloneJSON(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	data := historicalMarshal(t, value)
	var clone map[string]any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func historicalWriteExclusive(t *testing.T, path string, data []byte) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
