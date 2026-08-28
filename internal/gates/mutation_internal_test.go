package gates

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestMutationImportLoadsApprovedRepositoryArtifacts(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	writeMutationImportFixtures(t, root)
	var observed mutation.Campaign
	var checkpoints []mutation.Checkpoint
	var ledger mutation.MigrationLedger
	var output bytes.Buffer
	runner := Runner{
		Root: root,
		Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{{
			Directory: ".", ModulePath: "example", GoVersion: "1.27.0",
			Gates: map[string]bool{"mutation": true}, Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}},
		}}},
		Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
		Executor: runtimeExecutor(filepath.Join(root, ".task")), Output: &output,
		mutationImport: func(_ context.Context, campaign mutation.Campaign, imported []mutation.Checkpoint, approvals mutation.MigrationLedger) error {
			observed, checkpoints, ledger = campaign, imported, approvals
			return nil
		},
	}
	if err := runner.MutationImport(context.Background(), ".", "legacy.zip", "ledger.json"); err != nil {
		t.Fatalf("MutationImport() error = %v", err)
	}
	if observed.Policy.ModulePath != "example" || len(checkpoints) != 1 || ledger.SchemaVersion != 3 || !strings.Contains(output.String(), "mutation-import") {
		t.Fatalf("MutationImport() observed %#v, %d, %#v, %q", observed.Policy, len(checkpoints), ledger, output.String())
	}
}

func TestMutationMaterializesConfiguredCheckpointBeforeVerification(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	writeMutationImportFixtures(t, root)
	var order []string
	runner := Runner{
		Root: root,
		Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{{
			Directory: ".", ModulePath: "example", GoVersion: "1.27.0",
			Gates: map[string]bool{"mutation": true}, Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}},
		}}},
		Policy: config.Config{
			Evidence: config.Evidence{Root: ".verification"},
			Mutation: config.Mutation{
				Root: ".verification/mutation",
				Imports: []config.MutationImport{{
					Module: ".", Archive: "legacy.zip", Ledger: "ledger.json",
				}},
			},
		},
		Executor: runtimeExecutor(filepath.Join(root, ".task")),
		mutationImport: func(context.Context, mutation.Campaign, []mutation.Checkpoint, mutation.MigrationLedger) error {
			order = append(order, "import")
			return nil
		},
		mutationCampaign: func(context.Context, mutation.Campaign) error {
			order = append(order, "verify")
			return nil
		},
	}
	if err := runner.Mutation(context.Background(), []string{"."}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"import", "verify"}) {
		t.Fatalf("mutation order = %v", order)
	}
	if _, exists := runner.configuredMutationImport("missing"); exists {
		t.Fatal("unexpected configured import for missing module")
	}
	runner.mutationImport = func(context.Context, mutation.Campaign, []mutation.Checkpoint, mutation.MigrationLedger) error {
		return errors.New("import failed")
	}
	runner.mutationCampaign = func(context.Context, mutation.Campaign) error {
		t.Fatal("verification ran after failed checkpoint import")
		return nil
	}
	if err := runner.Mutation(context.Background(), []string{"."}); err == nil || !strings.Contains(err.Error(), "import failed") {
		t.Fatalf("Mutation(failed import) error = %v", err)
	}
	runner.mutationImport = func(context.Context, mutation.Campaign, []mutation.Checkpoint, mutation.MigrationLedger) error {
		return mutation.ErrInputChanged
	}
	runs := 0
	runner.mutationCampaign = func(context.Context, mutation.Campaign) error {
		runs++
		return nil
	}
	if err := runner.Mutation(context.Background(), []string{"."}); err != nil || runs != 1 {
		t.Fatalf("Mutation(stale import) = %v, runs = %d", err, runs)
	}
}

func TestMutationImportFailsClosed(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	writeMutationImportFixtures(t, root)
	base := func() Runner {
		module := inventory.Module{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true}, Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}}}
		return Runner{
			Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{module}},
			Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
			Executor: runtimeExecutor(filepath.Join(root, ".task")),
		}
	}
	for name, run := range map[string]func(Runner) error{
		"unknown module": func(runner Runner) error {
			return runner.MutationImport(context.Background(), "missing", "legacy.zip", "ledger.json")
		},
		"disabled": func(runner Runner) error {
			runner.Catalog.Modules[0].Gates = map[string]bool{}
			return runner.MutationImport(context.Background(), ".", "legacy.zip", "ledger.json")
		},
		"missing archive": func(runner Runner) error {
			return runner.MutationImport(context.Background(), ".", "missing.zip", "ledger.json")
		},
		"missing ledger": func(runner Runner) error {
			return runner.MutationImport(context.Background(), ".", "legacy.zip", "missing.json")
		},
		"malformed archive": func(runner Runner) error {
			if err := os.WriteFile(filepath.Join(root, "bad.zip"), []byte("bad"), 0o600); err != nil {
				t.Fatal(err)
			}
			return runner.MutationImport(context.Background(), ".", "bad.zip", "ledger.json")
		},
		"malformed ledger": func(runner Runner) error {
			if err := os.WriteFile(filepath.Join(root, "bad.json"), []byte("{}"), 0o600); err != nil {
				t.Fatal(err)
			}
			return runner.MutationImport(context.Background(), ".", "legacy.zip", "bad.json")
		},
		"importer": func(runner Runner) error {
			runner.mutationImport = func(context.Context, mutation.Campaign, []mutation.Checkpoint, mutation.MigrationLedger) error {
				return errors.New("import failed")
			}
			return runner.MutationImport(context.Background(), ".", "legacy.zip", "ledger.json")
		},
		"default importer": func(runner Runner) error {
			return runner.MutationImport(context.Background(), ".", "legacy.zip", "ledger.json")
		},
		"campaign setup": func(runner Runner) error {
			runner.Executor = executorFunction(func(context.Context, Command) error { return nil })
			return runner.MutationImport(context.Background(), ".", "legacy.zip", "ledger.json")
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := run(base()); err == nil {
				t.Fatalf("MutationImport(%s) error = nil", name)
			}
		})
	}
	runner := base()
	runner.Output = nil
	runner.mutationImport = func(context.Context, mutation.Campaign, []mutation.Checkpoint, mutation.MigrationLedger) error {
		return nil
	}
	if err := runner.MutationImport(context.Background(), ".", "legacy.zip", "ledger.json"); err != nil {
		t.Fatalf("MutationImport(discarded output) error = %v", err)
	}
}

func TestSelectedMutationModuleFindsExactDirectory(t *testing.T) {
	module := inventory.Module{Directory: "."}
	selected, err := selectedMutationModule([]inventory.Module{module}, ".")
	if err != nil || selected.Directory != "." {
		t.Fatalf("selectedMutationModule() = %#v, %v", selected, err)
	}
	if _, err := selectedMutationModule(nil, "."); err == nil {
		t.Fatal("selectedMutationModule(unknown) error = nil")
	}
}

func writeMutationImportFixtures(t *testing.T, root string) {
	t.Helper()
	report := json.RawMessage(`{"files":[{"file_name":"source.go","mutations":[{"type":"NEGATION","status":"KILLED","line":1,"column":1}]}]}`)
	result, err := mutation.ValidateReport(bytes.NewReader(report))
	if err != nil {
		t.Fatal(err)
	}
	oldInput := strings.Repeat("a", 64)
	revision := strings.Repeat("e", 40)
	verifier := mutation.LegacyVerifierDigest()
	checkpoint, err := json.Marshal(map[string]any{
		"schema_version": 3, "module": ".", "package": ".", "execution_revision": revision,
		"gate_input_digest": oldInput, "gremlins_version": mutation.GremlinsVersion,
		"gremlins_verifier_sha256": verifier, "environment": map[string]string{"GOVERSION": "go1.26.6"},
		"report": report,
	})
	if err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	entry, err := writer.Create("mutation-checkpoints/root.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write(checkpoint); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "legacy.zip"), archive.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	reportDigest := strings.TrimPrefix(result.Digest, "sha256:")
	ledger := mutation.MigrationLedger{
		SchemaVersion: 3,
		Reason:        "The exact legacy checkpoint is approved for content-addressed evidence migration.",
		VerifierMigrationReview: mutation.VerifierMigrationReview{
			GremlinsVerifierSHA256: verifier,
			Reason:                 "The complete legacy verifier semantics were independently reviewed for equivalence.",
			ReviewedAt:             "2026-08-27",
		},
		VerifierMigrations: []mutation.VerifierMigration{{
			ExecutionRevision: revision, GateInputDigest: oldInput,
			GremlinsVerifierSHA256: verifier, GremlinsVersion: mutation.GremlinsVersion,
			Module: ".", Package: ".", ReportSHA256: reportDigest,
		}},
		Entries: []mutation.InputMigration{},
	}
	data, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "ledger.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestMutationBuildsCanonicalCampaignAndRunsEnabledModules(t *testing.T) {
	root := t.TempDir()
	writeMutationFixture(t, root, ".", "module example\n\nrequire github.com/testcontainers/testcontainers-go v0.0.0\n")
	writeMutationFixture(t, root, "nested", "module example/nested\n")
	mutationRoot := filepath.Join(root, ".verification", "mutation")
	if err := os.MkdirAll(mutationRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mutationRoot, "zero-inventory.json"), []byte("{\"schema_version\":1,\"packages\":[]}"), 0o600); err != nil {
		t.Fatal(err)
	}

	processCalled := false
	executor := workspaceExecutor{directory: filepath.Join(root, ".task"), run: func(_ context.Context, command Command) error {
		if command.Name == "go" && reflect.DeepEqual(command.Args, []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}) {
			_, _ = io.WriteString(command.Stdout, `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
		}
		if command.Name == "probe" {
			processCalled = true
		}
		return nil
	}}
	var observed mutation.Campaign
	var output bytes.Buffer
	runner := Runner{
		Root: root,
		Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{
			{
				Directory: ".", ModulePath: "example", GoVersion: "1.27.0",
				Gates: map[string]bool{"mutation": true}, TestTags: []string{"integration", "interoperability"},
				BuildTags: []string{"custom"}, OwnedDependencies: []string{"example", "example/zeta", "example/nested", "example/fixture"},
				Packages: []inventory.Package{
					{Directory: ".", CoverageRequired: true},
					{Directory: "testsupport", CoverageRequired: false},
				},
			},
			{Directory: "nested", ModulePath: "example/nested", Kind: "public"},
			{Directory: "fixture", ModulePath: "example/fixture", Kind: "fixture"},
			{Directory: "other", ModulePath: "example/other", Kind: "public"},
			{Directory: "zeta", ModulePath: "example/zeta", Kind: "public"},
		}},
		Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
		Executor: executor,
		Output:   &output,
		mutationCampaign: func(_ context.Context, campaign mutation.Campaign) error {
			observed = campaign
			return nil
		},
	}

	if err := runner.Mutation(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Mutation() error = %v", err)
	}
	if observed.Policy.Repository != "example" || observed.Policy.ModulePath != "example" ||
		!reflect.DeepEqual(observed.Policy.Packages, []string{"."}) ||
		!reflect.DeepEqual(observed.Policy.TestTags, []string{"integration"}) ||
		!reflect.DeepEqual(observed.Policy.BuildTags, []string{"custom"}) || observed.Policy.Workers != 1 {
		t.Fatalf("campaign policy = %#v", observed.Policy)
	}
	if !reflect.DeepEqual(observed.Policy.OwnedModules, []mutation.OwnedModule{
		{ModulePath: "example/nested", Directory: "nested"},
		{ModulePath: "example/zeta", Directory: "zeta"},
	}) {
		t.Fatalf("owned modules = %#v", observed.Policy.OwnedModules)
	}
	wantWorkspace := filepath.Join(executor.directory, "mutation-campaigns", mutationModuleSlug("example"))
	if observed.RuntimeIdentity.GoVersion != "go1.27.0" || observed.Workspace != wantWorkspace || observed.Process == nil {
		t.Fatalf("campaign runtime/workspace/process = %#v, %q, %v", observed.RuntimeIdentity, observed.Workspace, observed.Process)
	}
	if err := observed.Process(context.Background(), "probe", []string{"arg"}, root, map[string]string{"A": "B"}, io.Discard, io.Discard); err != nil || !processCalled {
		t.Fatalf("campaign process = %v, called = %v", err, processCalled)
	}
	if !strings.Contains(output.String(), "mutation") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestMutationReportsDisabledModulesWithoutExecution(t *testing.T) {
	var output bytes.Buffer
	runner := Runner{
		Root: t.TempDir(), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: map[string]bool{}}}},
		Executor: executorFunction(func(context.Context, Command) error { return errors.New("must not execute") }), Output: &output,
	}
	if err := runner.Mutation(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Mutation() error = %v", err)
	}
	if !strings.Contains(output.String(), "not applicable") {
		t.Fatalf("output = %q", output.String())
	}
	runner.Output = nil
	if err := runner.Mutation(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Mutation() discarded output error = %v", err)
	}
	if err := runner.Mutation(context.Background(), []string{"missing"}); err == nil {
		t.Fatal("Mutation() unknown module error = nil")
	}
}

func TestMutationDisabledModuleDoesNotStopLaterEnabledModule(t *testing.T) {
	root := t.TempDir()
	writeMutationFixture(t, root, "z-enabled", "module example/enabled\n")
	mutationRoot := filepath.Join(root, ".verification", "mutation")
	if err := os.MkdirAll(mutationRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mutationRoot, "zero-inventory.json"), []byte("{\"schema_version\":1,\"packages\":[]}"), 0o600); err != nil {
		t.Fatal(err)
	}
	runs := 0
	runner := Runner{
		Root: root,
		Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{
			{Directory: "a-disabled"},
			{Directory: "z-enabled", ModulePath: "example/enabled", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true}},
		}},
		Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
		Executor: runtimeExecutor(filepath.Join(root, ".task")),
		mutationCampaign: func(context.Context, mutation.Campaign) error {
			runs++
			return nil
		},
	}
	if err := runner.Mutation(context.Background(), []string{"a-disabled", "z-enabled"}); err != nil {
		t.Fatalf("Mutation() error = %v", err)
	}
	if runs != 1 {
		t.Fatalf("enabled mutation runs = %d, want 1", runs)
	}
}

func TestMutationFailsClosedForInvalidSetup(t *testing.T) {
	failure := errors.New("campaign failed")
	tests := []struct {
		name     string
		prepare  func(*testing.T, string)
		executor Executor
		run      mutationCampaignRunner
		want     string
	}{
		{"workspace", writeValidMutationSetup, executorFunction(func(context.Context, Command) error { return nil }), nil, "task-owned workspace"},
		{"missing inventory", func(t *testing.T, root string) { writeMutationFixture(t, root, ".", "module example\n") }, workspaceExecutor{directory: "/task", run: func(context.Context, Command) error { return nil }}, nil, "zero-mutant inventory"},
		{"runtime", writeValidMutationSetup, workspaceExecutor{directory: "/task", run: func(_ context.Context, command Command) error {
			_, _ = io.WriteString(command.Stdout, "not-json")
			return nil
		}}, nil, "runtime identity"},
		{"runtime command", writeValidMutationSetup, workspaceExecutor{directory: "/task", run: func(context.Context, Command) error { return failure }}, nil, "read Go runtime identity"},
		{"services", writeValidMutationSetup, runtimeExecutor("/task"), nil, "service identities"},
		{"wrong services", writeValidMutationSetup, runtimeExecutor("/task"), nil, "identity for postgresql is missing"},
		{"campaign", writeValidMutationSetup, runtimeExecutor("/task"), func(context.Context, mutation.Campaign) error { return failure }, "campaign failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			test.prepare(t, root)
			module := inventory.Module{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true}, Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}}}
			if test.name == "services" || test.name == "wrong services" {
				module.RequiredServices = []string{"postgresql"}
			}
			runner := Runner{
				Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{module}},
				Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor: test.executor, mutationCampaign: test.run,
			}
			if test.name == "services" || test.name == "wrong services" {
				runner.startServices = func(context.Context, []string) (serviceLease, error) {
					lease := &fakeServiceLease{}
					if test.name == "wrong services" {
						lease.identities = map[string]string{"redis": "redis#sha256:identity"}
					}
					return lease, nil
				}
			}
			err := runner.Mutation(context.Background(), []string{"."})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Mutation() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestMutationRejectsMalformedAndSymlinkedReviewInventories(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, ".verification", "mutation")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "zero-inventory.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadZeroInventory(root, ".verification/mutation/zero-inventory.json"); err == nil || !strings.Contains(err.Error(), "parse zero-mutant inventory") {
		t.Fatalf("loadZeroInventory(malformed) error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "missing"), path); err != nil {
		t.Fatal(err)
	}
	if _, err := loadZeroInventory(root, ".verification/mutation/zero-inventory.json"); err == nil || !strings.Contains(err.Error(), "read zero-mutant inventory") {
		t.Fatalf("loadZeroInventory(symlink) error = %v", err)
	}
}

func TestMutationEnvironmentUsesIsolatedModfileAndHandlesFailures(t *testing.T) {
	root := t.TempDir()
	writeMutationFixture(t, root, ".", "module example\n")
	if err := os.WriteFile(filepath.Join(root, "go.sum"), []byte("sum\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	owned := []mutation.OwnedModule{{ModulePath: "example/nested", Directory: "nested"}}
	commands := 0
	runner := Runner{Root: root, Executor: executorFunction(func(_ context.Context, command Command) error {
		commands++
		if !strings.Contains(strings.Join(command.Args, " "), "-replace=example/nested=") {
			t.Fatalf("mod edit command = %#v", command.Args)
		}
		return nil
	})}
	environment, err := runner.mutationEnvironment(context.Background(), t.TempDir(), inventory.Module{Directory: ".", ModulePath: "example"}, owned)
	if err != nil || commands != 1 || !strings.Contains(environment["GOFLAGS"], "-modfile=") {
		t.Fatalf("mutationEnvironment() = %#v, commands = %d, error = %v", environment, commands, err)
	}

	failure := errors.New("injected failure")
	tests := []struct {
		name  string
		files mutationFileSystem
		exec  Executor
		root  func(*testing.T) string
		want  string
	}{
		{"module read", operatingMutationFiles{}, runner.Executor, func(t *testing.T) string { return t.TempDir() }, "read mutation go.mod"},
		{"mkdir", &failingMutationFiles{mkdirErr: failure}, runner.Executor, func(*testing.T) string { return root }, "create mutation modfile directory"},
		{"modfile write", &failingMutationFiles{writeErrAt: 1, failure: failure}, runner.Executor, func(*testing.T) string { return root }, "write mutation modfile"},
		{"sumfile write", &failingMutationFiles{writeErrAt: 2, failure: failure}, runner.Executor, func(*testing.T) string { return root }, "write mutation sumfile"},
		{"mod edit", operatingMutationFiles{}, executorFunction(func(context.Context, Command) error { return failure }), func(*testing.T) string { return root }, "configure mutation owned module"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := Runner{Root: test.root(t), Executor: test.exec, mutationFiles: test.files}
			_, err := candidate.mutationEnvironment(context.Background(), t.TempDir(), inventory.Module{Directory: ".", ModulePath: "example"}, owned)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("mutationEnvironment() error = %v, want %q", err, test.want)
			}
		})
	}
	unsafeRoot := t.TempDir()
	writeMutationFixture(t, unsafeRoot, ".", "module example\n")
	if err := os.Symlink(filepath.Join(unsafeRoot, "missing"), filepath.Join(unsafeRoot, "go.sum")); err != nil {
		t.Fatal(err)
	}
	unsafe := Runner{Root: unsafeRoot, Executor: runner.Executor}
	if _, err := unsafe.mutationEnvironment(context.Background(), t.TempDir(), inventory.Module{Directory: ".", ModulePath: "example"}, owned); err == nil || !strings.Contains(err.Error(), "read mutation go.sum") {
		t.Fatalf("mutationEnvironment() unsafe go.sum error = %v", err)
	}
}

func TestRunMutationPropagatesPreparationFailures(t *testing.T) {
	for _, test := range []struct {
		name    string
		catalog inventory.Inventory
		want    string
	}{
		{
			name: "environment",
			catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{
				{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", OwnedDependencies: []string{"example/nested"}, Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}}},
				{Directory: "nested", ModulePath: "example/nested"},
			}},
			want: "read mutation go.mod",
		},
		{
			name: "workers",
			catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{
				{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}}},
			}},
			want: "read mutation worker policy",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, ".verification", "mutation")
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "zero-inventory.json"), []byte("{\"schema_version\":1,\"packages\":[]}"), 0o600); err != nil {
				t.Fatal(err)
			}
			runner := Runner{
				Root: root, Catalog: test.catalog,
				Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor: runtimeExecutor(filepath.Join(root, ".task")),
			}
			err := runner.runMutation(context.Background(), io.Discard, test.catalog.Modules[0])
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runMutation() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestMutationWorkerPolicyAndDefaultCampaign(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Root: root}
	workers, err := runner.mutationWorkers(inventory.Module{Directory: "."})
	if err != nil || workers != 4 {
		t.Fatalf("mutationWorkers() = %d, %v", workers, err)
	}
	if _, err := (Runner{Root: t.TempDir()}).mutationWorkers(inventory.Module{Directory: "."}); err == nil {
		t.Fatal("mutationWorkers() missing go.mod error = nil")
	}

	task := filepath.Join(root, ".task")
	executor := workspaceExecutor{directory: task, run: func(_ context.Context, command Command) error {
		if reflect.DeepEqual(command.Args, []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}) {
			_, _ = io.WriteString(command.Stdout, `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
			return nil
		}
		if len(command.Args) > 0 && command.Args[0] == "list" {
			return errors.New("stop campaign")
		}
		return nil
	}}
	runner = Runner{
		Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{{
			Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true},
			Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}},
		}}},
		Policy: config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}}, Executor: executor,
	}
	if err := runner.Mutation(context.Background(), []string{"."}); err == nil || !strings.Contains(err.Error(), "stop campaign") {
		t.Fatalf("default campaign error = %v", err)
	}
}

func TestCheckRunsAndStopsAtMutationGate(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("mutation failed")
	for _, test := range []struct {
		name string
		run  mutationCampaignRunner
		want error
	}{
		{"pass", func(context.Context, mutation.Campaign) error { return nil }, nil},
		{"fail", func(context.Context, mutation.Campaign) error { return failure }, failure},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{
				Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{{
					Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true},
					Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}},
				}}},
				Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor: runtimeExecutor(filepath.Join(root, ".task")), mutationCampaign: test.run,
			}
			err := runner.Check(context.Background(), []string{"."})
			if !errors.Is(err, test.want) {
				t.Fatalf("Check() error = %v, want %v", err, test.want)
			}
		})
	}
}

type failingMutationFiles struct {
	mkdirErr   error
	writeErrAt int
	writes     int
	failure    error
}

func (files *failingMutationFiles) MkdirAll(string, os.FileMode) error { return files.mkdirErr }

func (files *failingMutationFiles) WriteFile(string, []byte, os.FileMode) error {
	files.writes++
	if files.writes == files.writeErrAt {
		return files.failure
	}
	return nil
}

func writeValidMutationSetup(t *testing.T, root string) {
	t.Helper()
	writeMutationFixture(t, root, ".", "module example\n")
	directory := filepath.Join(root, ".verification", "mutation")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "zero-inventory.json"), []byte("{\"schema_version\":1,\"packages\":[]}"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeMutationFixture(t *testing.T, root, directory, mod string) {
	t.Helper()
	path := filepath.Join(root, directory)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "go.mod"), []byte(mod), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runtimeExecutor(directory string) Executor {
	return workspaceExecutor{directory: directory, run: func(_ context.Context, command Command) error {
		if reflect.DeepEqual(command.Args, []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}) {
			_, _ = io.WriteString(command.Stdout, `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
		}
		return nil
	}}
}
