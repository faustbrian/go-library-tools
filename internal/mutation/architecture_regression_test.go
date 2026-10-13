package mutation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/evidence"
)

func TestCampaignEquivalentArchitectureAdmission(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64", "386"} {
		t.Run(arch, func(t *testing.T) {
			campaign, process := campaignFixture(t)
			campaign.RuntimeIdentity.GOARCH = arch
			source, err := SourceDigest(campaign.Root, ".", ".")
			if err != nil {
				t.Fatal(err)
			}
			data := fmt.Sprintf(`{"schema_version":1,"packages":[{"module_directory":".","package_directory":".","source_digest":%q,"gremlins_version":"v0.6.0","gremlins_verifier_sha256":%q,"mutations":[{"file_name":"source.go","type":"A","line":3,"column":1,"goarchs":["amd64","arm64"],"contract_domain":"Independently bounded 64-bit scalar equality domain","reason":"Only the supported 64-bit scalar domain preserves this exact outcome."}]}]}`, source, LegacyVerifierDigest())
			campaign.EquivalentReviews, err = ParseEquivalentInventory(strings.NewReader(data))
			if err != nil {
				t.Fatalf("optional bounded architecture admission rejected: %v", err)
			}
			report := `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}],"mutants_killed":0,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":1,"mutations_coverage":100,"test_efficacy":0}`
			process.report, process.fail = &report, "efficacy"
			err = campaign.Run(t.Context())
			if arch == "386" {
				if err == nil {
					t.Fatal("32-bit viable survivor accepted by 64-bit review")
				}
			} else if err != nil {
				t.Fatalf("admitted 64-bit survivor rejected: %v", err)
			}
		})
	}
}

func architectureInventory(t *testing.T, scope string) EquivalentInventory {
	t.Helper()
	data := fmt.Sprintf(`{"schema_version":1,"packages":[{"module_directory":".","package_directory":".","source_digest":%q,"gremlins_version":"v0.6.0","gremlins_verifier_sha256":%q,"mutations":[{"file_name":"source.go","type":"A","line":3,"column":1,%s"contract_domain":"Independently bounded scalar equality admission domain","reason":"The documented bounded scalar domain preserves this exact outcome."}]}]}`, strings.Repeat("a", 64), LegacyVerifierDigest(), scope)
	inventory, err := ParseEquivalentInventory(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return inventory
}

func TestArchitectureScopeStrictAdmissionAndLegacyCompatibility(t *testing.T) {
	inventory := architectureInventory(t, `"goarchs":["amd64","arm64"],`)
	for _, arch := range []string{"", "unknown", "AMD64"} {
		if _, err := inventory.Review(".", ".", strings.Repeat("a", 64), GremlinsVersion, LegacyVerifierDigest(), arch); !errors.Is(err, ErrInvalid) {
			t.Fatalf("unknown native target admitted: %q %v", arch, err)
		}
	}
	if _, err := inventory.Review(".", ".", strings.Repeat("a", 64), GremlinsVersion, LegacyVerifierDigest()); !errors.Is(err, ErrInvalid) {
		t.Fatal("missing runtime admitted scoped review")
	}
	if _, err := ValidateReportWithReview(strings.NewReader(`{"files":[]}`), &inventory.Packages[0]); !errors.Is(err, ErrInvalid) {
		t.Fatal("unbound parsed review admitted directly")
	}
	legacy := architectureInventory(t, "")
	for _, arch := range []string{"amd64", "arm64", "386"} {
		if review, err := legacy.Review(".", ".", strings.Repeat("a", 64), GremlinsVersion, LegacyVerifierDigest(), arch); err != nil || review == nil {
			t.Fatalf("legacy global disposition changed: %q %v", arch, err)
		}
	}
	encoded := fmt.Sprintf(`{"schema_version":1,"packages":[{"module_directory":".","package_directory":".","source_digest":%q,"gremlins_version":"v0.6.0","gremlins_verifier_sha256":%q,"mutations":[{"file_name":"source.go","type":"A","line":3,"column":1,"goarchs":%%s,"contract_domain":"Independently bounded scalar equality admission domain","reason":"The documented bounded scalar domain preserves this exact outcome."}]}]}`, strings.Repeat("a", 64), LegacyVerifierDigest())
	for _, scope := range []string{`null`, `[]`, `["amd64","amd64"]`, `["arm64","amd64"]`, `["unknown"]`, `"amd64"`, `[32]`} {
		if _, err := ParseEquivalentInventory(strings.NewReader(fmt.Sprintf(encoded, scope))); !errors.Is(err, ErrInvalid) {
			t.Fatalf("malformed scope accepted %s: %v", scope, err)
		}
	}
	selected, err := inventory.Review(".", ".", strings.Repeat("a", 64), GremlinsVersion, LegacyVerifierDigest(), "amd64")
	if err != nil {
		t.Fatal(err)
	}
	selected.Mutations[0].Goarchs[0] = "386"
	if inventory.Packages[0].Mutations[0].Goarchs[0] != "amd64" {
		t.Fatal("review shares mutable architecture scope with inventory")
	}
}

func TestScopedReuseRequiresMatchingNativeRecordArchitecture(t *testing.T) {
	inventory := architectureInventory(t, `"goarchs":["amd64","arm64"],`)
	review, err := inventory.Review(".", ".", strings.Repeat("a", 64), GremlinsVersion, LegacyVerifierDigest(), "amd64")
	if err != nil {
		t.Fatal(err)
	}
	report := []byte(`{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}],"mutants_killed":0,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":1,"mutations_coverage":100,"test_efficacy":0}`)
	for _, arch := range []string{"amd64", "386", ""} {
		t.Run(arch, func(t *testing.T) {
			root := t.TempDir()
			mutationRoot := filepath.Join(root, "mutation")
			input := "sha256:" + strings.Repeat("c", 64)
			_, _, result, err := StoreReportWithReview(mutationRoot, input, report, review)
			if err != nil {
				t.Fatal(err)
			}
			record := evidence.Record{SchemaVersion: evidence.SchemaVersion, Repository: "example", Module: ".", Package: ".", Gate: "mutation", InputDigest: input, VerifierDigest: SemanticVerifierDigest(), Result: "passed", ReportDigest: result.Digest, CompletedAt: time.Unix(10, 0).UTC(), Environment: map[string]string{"GOARCH": arch}}
			if _, _, err := evidence.Store(root, record); err != nil {
				t.Fatal(err)
			}
			reused, _, err := ReuseWithReview(root, mutationRoot, "example", ".", ".", input, review)
			if arch == "amd64" {
				if err != nil || !reused {
					t.Fatalf("matching native record refused: %v", err)
				}
			} else if !errors.Is(err, ErrInvalid) || reused {
				t.Fatalf("wrong/missing native record target accepted: %v", err)
			}
		})
	}
}

func TestScopedCampaignCannotReuseEvidenceAfterArchitectureChange(t *testing.T) {
	campaign, process := campaignFixture(t)
	source, err := SourceDigest(campaign.Root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	inventory := architectureInventory(t, `"goarchs":["amd64","arm64"],`)
	inventory.Packages[0].SourceDigest = source
	campaign.EquivalentReviews = inventory
	report := `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}],"mutants_killed":0,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":1,"mutations_coverage":100,"test_efficacy":0}`
	process.report, process.fail = &report, "efficacy"
	if err := campaign.Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := campaign.Run(t.Context()); err != nil || process.mutations != 1 {
		t.Fatalf("matching target did not reuse: %v", err)
	}
	campaign.RuntimeIdentity.GOARCH = "386"
	campaign.Workspace = filepath.Join(campaign.Root, ".other-task")
	if err := campaign.Run(t.Context()); err == nil || process.mutations != 2 {
		t.Fatalf("386 reused 64-bit survivor proof: %v executions=%d", err, process.mutations)
	}
	// No new passed record may appear for the refused 32-bit campaign.
	_, input, err := campaign.packageInput(t.Context(), ".")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := evidence.Load(campaign.EvidenceRoot, "mutation", input); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("refused native target published evidence: %v", err)
	}
}
