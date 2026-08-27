package mutation

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCampaignRejectsMalformedPolicyAndEscapingEvidence(t *testing.T) {
	valid, _ := campaignFixture(t)
	for name, mutate := range map[string]func(*Campaign){
		"relative root":           func(value *Campaign) { value.Root = "." },
		"relative evidence":       func(value *Campaign) { value.EvidenceRoot = ".verification" },
		"relative mutation":       func(value *Campaign) { value.MutationRoot = ".verification/mutation" },
		"relative workspace":      func(value *Campaign) { value.Workspace = ".task" },
		"nil process":             func(value *Campaign) { value.Process = nil },
		"missing repository":      func(value *Campaign) { value.Policy.Repository = "" },
		"missing module path":     func(value *Campaign) { value.Policy.ModulePath = "" },
		"missing go version":      func(value *Campaign) { value.Policy.GoVersion = "" },
		"bad workers":             func(value *Campaign) { value.Policy.Workers = 0 },
		"excess workers":          func(value *Campaign) { value.Policy.Workers = 65 },
		"empty packages":          func(value *Campaign) { value.Policy.Packages = nil },
		"escaping evidence":       func(value *Campaign) { value.EvidenceRoot = filepath.Join(t.TempDir(), "outside") },
		"unsafe package":          func(value *Campaign) { value.Policy.Packages = []string{"../outside"} },
		"duplicate package":       func(value *Campaign) { value.Policy.Packages = []string{".", "."} },
		"unsafe tag":              func(value *Campaign) { value.Policy.TestTags = []string{"bad\ntag"} },
		"unsafe build tag":        func(value *Campaign) { value.Policy.BuildTags = []string{"bad\ntag"} },
		"unsafe environment":      func(value *Campaign) { value.Environment = map[string]string{"BAD=KEY": "value"} },
		"empty environment":       func(value *Campaign) { value.Environment = map[string]string{"": "value"} },
		"missing runtime Go":      func(value *Campaign) { value.RuntimeIdentity.GoVersion = "" },
		"missing runtime OS":      func(value *Campaign) { value.RuntimeIdentity.GOOS = "" },
		"missing runtime arch":    func(value *Campaign) { value.RuntimeIdentity.GOARCH = "" },
		"missing runtime CGO":     func(value *Campaign) { value.RuntimeIdentity.CGOEnabled = "" },
		"unsafe module directory": func(value *Campaign) { value.Policy.ModuleDirectory = "../outside" },
	} {
		t.Run(name, func(t *testing.T) {
			campaign := valid
			campaign.Policy.Packages = append([]string(nil), valid.Policy.Packages...)
			mutate(&campaign)
			if err := campaign.Run(context.Background()); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Run() error = %v", err)
			}
		})
	}
}

func TestCampaignAcceptsMaximumWorkers(t *testing.T) {
	campaign, _ := campaignFixture(t)
	campaign.Policy.Workers = 64
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run(maximum workers) error = %v", err)
	}
}

func TestCampaignImportsApprovedLegacyCheckpoint(t *testing.T) {
	campaign, process := campaignFixture(t)
	campaign.Output = nil
	campaign.Now = nil
	_, currentInput, err := campaign.packageInput(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, ledger := approvedImportFixture(t, currentInput)
	if err := campaign.Import(context.Background(), []Checkpoint{checkpoint}, ledger); err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	reused, result, err := Reuse(campaign.EvidenceRoot, campaign.MutationRoot, "example", ".", ".", currentInput)
	if err != nil || !reused || result.Mutants != 1 {
		t.Fatalf("Reuse() = %v, %#v, %v", reused, result, err)
	}
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run() after import error = %v", err)
	}
	if process.mutations != 0 {
		t.Fatalf("Run() executed %d mutation campaigns after import", process.mutations)
	}
}

func TestCampaignImportFailsClosedAtEachBoundary(t *testing.T) {
	tests := map[string]func(*Campaign, *campaignProcess, *Checkpoint, *MigrationLedger) []Checkpoint{
		"invalid campaign": func(campaign *Campaign, _ *campaignProcess, checkpoint *Checkpoint, _ *MigrationLedger) []Checkpoint {
			campaign.Root = "."
			return []Checkpoint{*checkpoint}
		},
		"directory preparation": func(campaign *Campaign, _ *campaignProcess, checkpoint *Checkpoint, _ *MigrationLedger) []Checkpoint {
			campaign.directoryFiles = failingCampaignDirectories{}
			return []Checkpoint{*checkpoint}
		},
		"duplicate checkpoint": func(_ *Campaign, _ *campaignProcess, checkpoint *Checkpoint, _ *MigrationLedger) []Checkpoint {
			return []Checkpoint{*checkpoint, *checkpoint}
		},
		"package input": func(_ *Campaign, process *campaignProcess, checkpoint *Checkpoint, _ *MigrationLedger) []Checkpoint {
			process.fail = "list"
			return []Checkpoint{*checkpoint}
		},
		"zero mutants without review": func(_ *Campaign, _ *campaignProcess, checkpoint *Checkpoint, ledger *MigrationLedger) []Checkpoint {
			result, err := ValidateReport(strings.NewReader(`{"files":[]}`))
			if err != nil {
				t.Fatal(err)
			}
			checkpoint.Report = []byte(`{"files":[]}`)
			checkpoint.ReportDigest = result.Digest
			checkpoint.Mutants = 0
			digest := strings.TrimPrefix(result.Digest, "sha256:")
			ledger.VerifierMigrations[0].ReportSHA256 = digest
			ledger.Entries[0].ReportSHA256 = digest
			return []Checkpoint{*checkpoint}
		},
		"unapproved current input": func(_ *Campaign, _ *campaignProcess, checkpoint *Checkpoint, ledger *MigrationLedger) []Checkpoint {
			ledger.Entries[0].ReplacementInputDigest = strings.Repeat("b", 64)
			return []Checkpoint{*checkpoint}
		},
		"report store": func(campaign *Campaign, _ *campaignProcess, checkpoint *Checkpoint, _ *MigrationLedger) []Checkpoint {
			if err := os.MkdirAll(campaign.MutationRoot, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(campaign.MutationRoot, "reports"), []byte("file"), 0o600); err != nil {
				t.Fatal(err)
			}
			return []Checkpoint{*checkpoint}
		},
		"report identity": func(_ *Campaign, _ *campaignProcess, checkpoint *Checkpoint, _ *MigrationLedger) []Checkpoint {
			checkpoint.Mutants++
			return []Checkpoint{*checkpoint}
		},
		"evidence store": func(campaign *Campaign, _ *campaignProcess, checkpoint *Checkpoint, _ *MigrationLedger) []Checkpoint {
			if err := os.MkdirAll(campaign.EvidenceRoot, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(campaign.EvidenceRoot, "by-input"), []byte("file"), 0o600); err != nil {
				t.Fatal(err)
			}
			return []Checkpoint{*checkpoint}
		},
	}
	for name, prepare := range tests {
		t.Run(name, func(t *testing.T) {
			campaign, process := campaignFixture(t)
			_, currentInput, err := campaign.packageInput(context.Background(), ".")
			if err != nil {
				t.Fatal(err)
			}
			checkpoint, ledger := approvedImportFixture(t, currentInput)
			checkpoints := prepare(&campaign, process, &checkpoint, &ledger)
			if err := campaign.Import(context.Background(), checkpoints, ledger); err == nil {
				t.Fatalf("Import(%s) error = nil", name)
			}
		})
	}
}

func TestCampaignRejectsUnapprovedOrMismatchedImports(t *testing.T) {
	campaign, _ := campaignFixture(t)
	_, currentInput, err := campaign.packageInput(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, ledger := approvedImportFixture(t, currentInput)
	tests := map[string]struct {
		checkpoint Checkpoint
		ledger     MigrationLedger
	}{
		"unknown module":  {checkpoint: checkpoint, ledger: ledger},
		"unknown package": {checkpoint: checkpoint, ledger: ledger},
		"unapproved":      {checkpoint: checkpoint, ledger: MigrationLedger{}},
	}
	module := tests["unknown module"]
	module.checkpoint.Module = "other"
	tests["unknown module"] = module
	pkg := tests["unknown package"]
	pkg.checkpoint.Package = "missing"
	tests["unknown package"] = pkg
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := campaign.Import(context.Background(), []Checkpoint{test.checkpoint}, test.ledger); err == nil {
				t.Fatalf("Import(%s) error = nil", name)
			}
		})
	}
	if err := campaign.Import(context.Background(), nil, ledger); err == nil {
		t.Fatal("Import(empty) error = nil")
	}
}

func approvedImportFixture(t *testing.T, currentInput string) (Checkpoint, MigrationLedger) {
	t.Helper()
	report := []byte(`{"files":[{"file_name":"source.go","mutations":[{"type":"NEGATION","status":"KILLED","line":3,"column":1}]}]}`)
	result, err := ValidateReport(strings.NewReader(string(report)))
	if err != nil {
		t.Fatal(err)
	}
	oldInput := strings.Repeat("a", 64)
	revision := strings.Repeat("e", 40)
	verifier := LegacyVerifierDigest()
	checkpoint := Checkpoint{
		Module: ".", Package: ".", ExecutionRevision: revision, InputDigest: oldInput,
		VerifierDigest: verifier, Gremlins: GremlinsVersion,
		Environment:  map[string]string{"GOVERSION": "go1.26.6"},
		ReportDigest: result.Digest, Report: report, Mutants: result.Mutants,
	}
	reportDigest := strings.TrimPrefix(result.Digest, "sha256:")
	ledger := MigrationLedger{
		SchemaVersion: 3,
		Reason:        "The exact legacy checkpoint is approved for content-addressed evidence migration.",
		VerifierMigrationReview: VerifierMigrationReview{
			GremlinsVerifierSHA256: verifier,
			Reason:                 "The complete legacy verifier semantics were independently reviewed for equivalence.",
			ReviewedAt:             "2026-08-27",
		},
		VerifierMigrations: []VerifierMigration{{
			ExecutionRevision: revision, GateInputDigest: oldInput,
			GremlinsVerifierSHA256: verifier, GremlinsVersion: GremlinsVersion,
			Module: ".", Package: ".", ReportSHA256: reportDigest,
		}},
		Entries: []InputMigration{{
			ExecutionRevision: revision, GateInputDigest: oldInput,
			ReplacementInputDigest: strings.TrimPrefix(currentInput, "sha256:"),
			GremlinsVerifierSHA256: verifier, GremlinsVersion: GremlinsVersion,
			Module: ".", Package: ".", ReportSHA256: reportDigest,
			MigrationReason: "The new digest removes Git identity while preserving all behavioral inputs.",
		}},
	}
	return checkpoint, ledger
}

func TestAppendTagArgument(t *testing.T) {
	if got := appendTagArgument([]string{"test"}, nil); len(got) != 1 {
		t.Fatalf("empty tags = %#v", got)
	}
	if got := appendTagArgument([]string{"test"}, []string{"a", "b"}); len(got) != 2 || got[1] != "-tags=a,b" {
		t.Fatalf("tags = %#v", got)
	}
}

func TestRemoveStaleReport(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing")
	if err := removeStaleReport(missing); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "report")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := removeStaleReport(file); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, "directory")
	if err := os.MkdirAll(filepath.Join(directory, "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := removeStaleReport(directory); err == nil {
		t.Fatal("removed non-empty directory")
	}
}

func TestCampaignReportsRepositoryAndProcessFailures(t *testing.T) {
	for _, failure := range []string{"list", "download", "build", "coverage", "mutation", "second-list"} {
		t.Run(failure, func(t *testing.T) {
			campaign, process := campaignFixture(t)
			process.fail = failure
			if err := campaign.Run(context.Background()); err == nil {
				t.Fatalf("Run(%s) error = nil", failure)
			}
		})
	}
	t.Run("evidence root", func(t *testing.T) {
		campaign, _ := campaignFixture(t)
		if err := os.WriteFile(campaign.EvidenceRoot, []byte("file"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := campaign.Run(context.Background()); err == nil {
			t.Fatal("Run(evidence root) error = nil")
		}
	})
	t.Run("mutation root", func(t *testing.T) {
		campaign, _ := campaignFixture(t)
		if err := os.Mkdir(campaign.EvidenceRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(campaign.MutationRoot, []byte("file"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := campaign.Run(context.Background()); err == nil {
			t.Fatal("Run(mutation root) error = nil")
		}
	})
	t.Run("missing source", func(t *testing.T) {
		campaign, _ := campaignFixture(t)
		if err := os.Remove(filepath.Join(campaign.Root, "source.go")); err != nil {
			t.Fatal(err)
		}
		if err := campaign.Run(context.Background()); err == nil {
			t.Fatal("Run(missing source) error = nil")
		}
	})
}

func TestCampaignRejectsSymlinkedRoots(t *testing.T) {
	for _, name := range []string{"evidence", "mutation", "workspace"} {
		t.Run(name, func(t *testing.T) {
			campaign, _ := campaignFixture(t)
			target := filepath.Join(campaign.Root, "target-"+name)
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			var path string
			switch name {
			case "evidence":
				path = campaign.EvidenceRoot
			case "mutation":
				if err := os.Mkdir(campaign.EvidenceRoot, 0o700); err != nil {
					t.Fatal(err)
				}
				path = campaign.MutationRoot
			default:
				path = campaign.Workspace
			}
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
			if err := campaign.Run(context.Background()); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Run(%s symlink) error = %v", name, err)
			}
		})
	}
}

func TestCampaignRejectsMissingMalformedZeroAndChangedReports(t *testing.T) {
	zero := `{"files":[]}`
	malformed := `{}`
	for _, test := range []struct {
		name    string
		report  *string
		skip    bool
		mutate  bool
		prepare func(*Campaign, *campaignProcess)
	}{
		{"missing", nil, true, false, nil},
		{"malformed", &malformed, false, false, nil},
		{"unreviewed zero", &zero, false, false, nil},
		{"changed input", nil, false, true, nil},
		{"package cache", nil, false, false, func(campaign *Campaign, _ *campaignProcess) {
			if err := os.MkdirAll(filepath.Join(campaign.Workspace, "mutation-cache"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(campaign.Workspace, "mutation-cache", "root"), []byte("file"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"stale report", nil, false, false, func(campaign *Campaign, _ *campaignProcess) {
			path := filepath.Join(campaign.Workspace, "report-root.json")
			if err := os.MkdirAll(filepath.Join(path, "child"), 0o700); err != nil {
				t.Fatal(err)
			}
		}},
		{"report store", nil, false, false, func(campaign *Campaign, process *campaignProcess) {
			process.afterMutation = func() error {
				return os.WriteFile(filepath.Join(campaign.MutationRoot, "reports"), []byte("file"), 0o600)
			}
		}},
		{"evidence store", nil, false, false, func(campaign *Campaign, process *campaignProcess) {
			process.afterMutation = func() error {
				return os.WriteFile(filepath.Join(campaign.EvidenceRoot, "by-input"), []byte("file"), 0o600)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			campaign, process := campaignFixture(t)
			process.report = test.report
			process.skipReport = test.skip
			process.mutateSource = test.mutate
			if test.prepare != nil {
				test.prepare(&campaign, process)
			}
			if err := campaign.Run(context.Background()); err == nil {
				t.Fatalf("Run(%s) error = nil", test.name)
			}
		})
	}
}

func TestCampaignAcceptsReviewedZeroReportAndDefaultOutputClock(t *testing.T) {
	campaign, process := campaignFixture(t)
	source, err := SourceDigest(campaign.Root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	campaign.Output = nil
	campaign.Now = nil
	campaign.ZeroReviews = ZeroInventory{SchemaVersion: 1, Packages: []ZeroReview{{
		ModuleDirectory: ".", PackageDirectory: ".", SourceDigest: source,
		GremlinsVersion: GremlinsVersion, GremlinsVerifierSHA256: LegacyVerifierDigest(),
		Reason: "The package has declarations only and no viable mutation points after complete manual review.",
	}}}
	process.skipReport = true
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run(reviewed zero) error = %v", err)
	}
}

func TestRunPackageReportsArgumentAndCorruptReuseFailures(t *testing.T) {
	campaign, _ := campaignFixture(t)
	campaign.Policy.Workers = 0
	state := campaignState{}
	if err := campaign.runPackage(context.Background(), io.Discard, ".", &state); err == nil {
		t.Fatal("runPackage(invalid arguments) error = nil")
	}
	campaign, _ = campaignFixture(t)
	_, input, err := campaign.packageInput(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(campaign.EvidenceRoot, "by-input", "mutation")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, strings.TrimPrefix(input, "sha256:")+".json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := campaign.runPackage(context.Background(), io.Discard, ".", &campaignState{}); err == nil {
		t.Fatal("runPackage(corrupt reuse) error = nil")
	}
}

func TestPackageSlugsDoNotAliasDifferentDirectories(t *testing.T) {
	if packageSlug("a/b") == packageSlug("a-b") || packageSlug(".") != "root" {
		t.Fatal("packageSlug() aliases distinct package directories")
	}
}

func TestCampaignDirectoryPreparationReportsInspectionFailure(t *testing.T) {
	campaign, _ := campaignFixture(t)
	campaign.directoryFiles = failingCampaignDirectories{}
	if err := campaign.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "inspect failed") {
		t.Fatalf("Run(directory inspection) error = %v", err)
	}
}

type failingCampaignDirectories struct{}

func (failingCampaignDirectories) MkdirAll(string, os.FileMode) error { return nil }
func (failingCampaignDirectories) Lstat(string) (os.FileInfo, error) {
	return nil, errors.New("inspect failed")
}

func campaignFixture(t *testing.T) (Campaign, *campaignProcess) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	verifier := t.TempDir()
	if err := os.WriteFile(filepath.Join(verifier, "go.mod"), []byte("module verifier\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	process := &campaignProcess{root: root, verifierSource: verifier}
	campaign := Campaign{
		Root: root, EvidenceRoot: filepath.Join(root, ".verification"),
		MutationRoot: filepath.Join(root, ".verification", "mutation"), Workspace: filepath.Join(root, ".task"),
		Policy: CampaignPolicy{
			Repository: "example", ModuleDirectory: ".", ModulePath: "example", GoVersion: "1.27.0",
			Packages: []string{"."}, ServiceIdentities: map[string]string{}, Workers: 1,
		},
		Environment:     map[string]string{},
		RuntimeIdentity: RuntimeIdentity{GoVersion: "go1.27.0", GOOS: "linux", GOARCH: "amd64", CGOEnabled: "0"},
		Process:         process.run,
	}
	return campaign, process
}
