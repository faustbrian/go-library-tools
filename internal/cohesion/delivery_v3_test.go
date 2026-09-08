//nolint:gocritic // whitespace key is an intentional hostile fixture
package cohesion

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDeriveGoalStatusExhaustivelyIgnoresReleaseState(t *testing.T) {
	t.Parallel()

	states := []DimensionState{
		DimensionNotStarted,
		DimensionInProgress,
		DimensionBlocked,
		DimensionNotApplicable,
		DimensionVerified,
	}
	for _, implementation := range states {
		for _, hardening := range states {
			want := independentlyDeriveGoalStatus(implementation, hardening)
			for _, release := range states {
				if got := deriveGoalStatus(implementation, hardening); got != want {
					t.Fatalf("deriveGoalStatus(%q, %q) with release %q = %q, want %q", implementation, hardening, release, got, want)
				}
			}
		}
	}
}

func TestEvidenceKindForDimensionStateUsesExactMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dimension string
		state     DimensionState
		want      string
		ok        bool
	}{
		{"implementation", DimensionVerified, "source-acceptance", true},
		{"hardening", DimensionVerified, "gate-attestation", true},
		{"release", DimensionVerified, "release-attestation", true},
		{"implementation", DimensionBlocked, "blocked-decision", true},
		{"hardening", DimensionNotApplicable, "not-applicable-decision", true},
		{"release", DimensionNotStarted, "", false},
		{"unknown", DimensionVerified, "", false},
	}
	for _, test := range tests {
		got, ok := evidenceKindForDimensionState(test.dimension, test.state)
		if got != test.want || ok != test.ok {
			t.Fatalf("evidenceKindForDimensionState(%q, %q) = %q, %t, want %q, %t", test.dimension, test.state, got, ok, test.want, test.ok)
		}
	}
}

func TestValidateManifestDeliveryRequiresDerivedGoalAndTerminalEvidence(t *testing.T) {
	t.Parallel()

	reference := EvidenceReference{
		ReceiptPath:   ".ai/evidence/cohesion.json",
		ReceiptSHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		EntryID:       "implementation-acceptance",
	}
	valid := ManifestDelivery{
		Goal: DeliveryGoal{
			ID:                 cohesionGoalID,
			RequirementsSHA256: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Status:             GoalComplete,
		},
		Implementation: DimensionVerified,
		Hardening:      DimensionNotApplicable,
		Release:        DimensionInProgress,
		Evidence: ManifestEvidence{
			Implementation: []EvidenceReference{reference},
			Hardening: []EvidenceReference{{
				ReceiptPath:   ".ai/evidence/not-applicable.json",
				ReceiptSHA256: "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				EntryID:       "hardening-not-applicable",
			}},
			Release: []EvidenceReference{},
		},
	}
	if err := validateManifestDelivery(valid, valid.Goal.RequirementsSHA256); err != nil {
		t.Fatalf("validateManifestDelivery(valid) error = %v", err)
	}

	mutations := map[string]func(*ManifestDelivery){
		"wrong goal id": func(value *ManifestDelivery) { value.Goal.ID = "other" },
		"wrong requirements": func(value *ManifestDelivery) {
			value.Goal.RequirementsSHA256 = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
		},
		"wrong derived status":          func(value *ManifestDelivery) { value.Goal.Status = GoalBlocked },
		"invalid dimension state":       func(value *ManifestDelivery) { value.Release = "invalid" },
		"terminal missing evidence":     func(value *ManifestDelivery) { value.Evidence.Implementation = nil },
		"nonterminal contains evidence": func(value *ManifestDelivery) { value.Evidence.Release = []EvidenceReference{reference} },
		"unsafe receipt path":           func(value *ManifestDelivery) { value.Evidence.Implementation[0].ReceiptPath = "../evidence.json" },
		"invalid receipt digest":        func(value *ManifestDelivery) { value.Evidence.Implementation[0].ReceiptSHA256 = "aaaaaaaa" },
		"empty entry identity":          func(value *ManifestDelivery) { value.Evidence.Implementation[0].EntryID = "" },
		"oversized receipt path": func(value *ManifestDelivery) {
			value.Evidence.Implementation[0].ReceiptPath = strings.Repeat("a", 4097)
		},
		"oversized entry identity":  func(value *ManifestDelivery) { value.Evidence.Implementation[0].EntryID = strings.Repeat("a", 129) },
		"duplicate cross dimension": func(value *ManifestDelivery) { value.Evidence.Hardening = []EvidenceReference{reference} },
		"duplicate within dimension": func(value *ManifestDelivery) {
			value.Evidence.Implementation = []EvidenceReference{reference, reference}
		},
		"unsorted evidence": func(value *ManifestDelivery) {
			value.Evidence.Implementation = []EvidenceReference{
				{ReceiptPath: "z", ReceiptSHA256: reference.ReceiptSHA256, EntryID: "z"},
				{ReceiptPath: "a", ReceiptSHA256: reference.ReceiptSHA256, EntryID: "a"},
			}
		},
	}
	for name, mutate := range mutations {
		candidate := cloneManifestDelivery(valid)
		mutate(&candidate)
		if err := validateManifestDelivery(candidate, valid.Goal.RequirementsSHA256); err == nil {
			t.Fatalf("validateManifestDelivery(%s) error = nil", name)
		}
	}
}

func TestDeliveryIdentityAndPathBoundaries(t *testing.T) {
	t.Parallel()

	for value, want := range map[string]bool{
		"implementation": true,
		"hardening":      true,
		"release":        true,
		"":               false,
		"other":          false,
	} {
		if got := validDimension(value); got != want {
			t.Fatalf("validDimension(%q) = %t, want %t", value, got, want)
		}
	}
	for value, want := range map[string]bool{
		strings.Repeat("a", 128): true,
		strings.Repeat("a", 129): false,
		" a":                     false,
		"a\n":                    false,
	} {
		if got := canonicalIdentity(value); got != want {
			t.Fatalf("canonicalIdentity(len %d) = %t, want %t", len(value), got, want)
		}
	}
	for name, path := range map[string]string{
		"one below":         strings.Repeat("a", 4095),
		"exact ASCII":       strings.Repeat("a", 4096),
		"exact multibyte":   strings.Repeat("ä", 2048),
		"safe characters":   "evidence/ääni report@v1+final.json",
		"safe leading dots": "evidence/..report.json",
	} {
		reference := EvidenceReference{
			ReceiptPath:   path,
			ReceiptSHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			EntryID:       strings.Repeat("a", 128),
		}
		if err := validateEvidenceReferences("implementation", DimensionVerified, []EvidenceReference{reference}, map[string]string{}); err != nil {
			t.Fatalf("validateEvidenceReferences(%s path boundary) error = %v", name, err)
		}
	}
	for name, path := range map[string]string{
		"one above":       strings.Repeat("a", 4097),
		"multibyte above": strings.Repeat("ä", 2049),
		"invalid UTF-8":   string([]byte{'a', 0xff}),
	} {
		reference := EvidenceReference{
			ReceiptPath:   path,
			ReceiptSHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			EntryID:       "entry",
		}
		if err := validateEvidenceReferences("implementation", DimensionVerified, []EvidenceReference{reference}, map[string]string{}); err == nil {
			t.Fatalf("validateEvidenceReferences(%s path) error = nil", name)
		}
	}
}

func TestValidateManifestDeliveryJSONRunsStrictSchemaAndSemanticValidation(t *testing.T) {
	t.Parallel()

	requirements := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	valid := ManifestDelivery{
		Goal:           DeliveryGoal{ID: cohesionGoalID, RequirementsSHA256: requirements, Status: GoalComplete},
		Implementation: DimensionVerified,
		Hardening:      DimensionNotApplicable,
		Release:        DimensionInProgress,
		Evidence: ManifestEvidence{
			Implementation: []EvidenceReference{{ReceiptPath: "evidence/implementation.json", ReceiptSHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", EntryID: "implementation"}},
			Hardening:      []EvidenceReference{{ReceiptPath: "evidence/hardening.json", ReceiptSHA256: "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", EntryID: "hardening"}},
			Release:        []EvidenceReference{},
		},
	}
	encode := func(value ManifestDelivery) []byte {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	if err := validateManifestDeliveryJSON(encode(valid), requirements); err != nil {
		t.Fatalf("validateManifestDeliveryJSON(valid) error = %v", err)
	}

	for name, path := range map[string]string{
		"one below":       strings.Repeat("a", 4095),
		"exact ASCII":     strings.Repeat("a", 4096),
		"exact multibyte": strings.Repeat("ä", 2048),
	} {
		candidate := cloneManifestDelivery(valid)
		candidate.Evidence.Implementation[0].ReceiptPath = path
		if err := validateManifestDeliveryJSON(encode(candidate), requirements); err != nil {
			t.Fatalf("validateManifestDeliveryJSON(%s path) error = %v", name, err)
		}
	}
	for name, path := range map[string]string{
		"one above":       strings.Repeat("a", 4097),
		"multibyte above": strings.Repeat("ä", 2049),
	} {
		candidate := cloneManifestDelivery(valid)
		candidate.Evidence.Implementation[0].ReceiptPath = path
		if err := validateManifestDeliveryJSON(encode(candidate), requirements); err == nil {
			t.Fatalf("validateManifestDeliveryJSON(%s path) error = nil", name)
		}
	}

	wrongShape := []byte(`{"goal":null}`)
	if err := validateManifestDeliveryJSON(wrongShape, requirements); err == nil {
		t.Fatal("validateManifestDeliveryJSON(wrong schema shape) error = nil")
	}
	duplicate := strings.Replace(string(encode(valid)), `"release":"in-progress"`, `"release":"in-progress","release":"verified"`, 1)
	if err := validateManifestDeliveryJSON([]byte(duplicate), requirements); err == nil {
		t.Fatal("validateManifestDeliveryJSON(duplicate key) error = nil")
	}
	invalidUTF8 := append([]byte(`{"receipt_path":"`), 0xff)
	invalidUTF8 = append(invalidUTF8, []byte(`"}`)...)
	if err := validateManifestDeliveryJSON(invalidUTF8, requirements); err == nil {
		t.Fatal("validateManifestDeliveryJSON(invalid UTF-8) error = nil")
	}
}

func TestCompareEvidenceReferenceUsesTheCompleteTuple(t *testing.T) {
	t.Parallel()

	base := EvidenceReference{ReceiptPath: "a", ReceiptSHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", EntryID: "a"}
	if compareEvidenceReference(base, EvidenceReference{ReceiptPath: "b"}) >= 0 {
		t.Fatal("receipt path did not sort first")
	}
	if compareEvidenceReference(base, EvidenceReference{ReceiptPath: "a", ReceiptSHA256: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}) >= 0 {
		t.Fatal("receipt digest did not sort second")
	}
	if compareEvidenceReference(base, EvidenceReference{ReceiptPath: base.ReceiptPath, ReceiptSHA256: base.ReceiptSHA256, EntryID: "b"}) >= 0 {
		t.Fatal("entry identity did not sort third")
	}
	if compareEvidenceReference(base, base) != 0 {
		t.Fatal("equal references did not compare equal")
	}
}

func independentlyDeriveGoalStatus(implementation, hardening DimensionState) GoalStatus {
	if (implementation == DimensionVerified || implementation == DimensionNotApplicable) &&
		(hardening == DimensionVerified || hardening == DimensionNotApplicable) {
		if implementation == DimensionNotApplicable && hardening == DimensionNotApplicable {
			return GoalNotApplicable
		}
		return GoalComplete
	}
	if implementation == DimensionBlocked || hardening == DimensionBlocked {
		return GoalBlocked
	}
	if (implementation == DimensionNotStarted || implementation == DimensionNotApplicable) &&
		(hardening == DimensionNotStarted || hardening == DimensionNotApplicable) {
		return GoalNotStarted
	}
	return GoalInProgress
}

func cloneManifestDelivery(value ManifestDelivery) ManifestDelivery {
	value.Evidence.Implementation = append([]EvidenceReference{}, value.Evidence.Implementation...)
	value.Evidence.Hardening = append([]EvidenceReference{}, value.Evidence.Hardening...)
	value.Evidence.Release = append([]EvidenceReference{}, value.Evidence.Release...)
	return value
}
