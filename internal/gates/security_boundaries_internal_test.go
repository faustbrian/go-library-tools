package gates

import (
	"context"
	"errors"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryCommandsDoNotClaimScannerProcessContainment(t *testing.T) {
	prepared := false
	executor := &processExecutor{prepareBoundedProcess: func(*exec.Cmd) error { prepared = true; return errors.New("must not prepare") }}
	_ = executor.Run(t.Context(), Command{Name: "nonexistent-inert-repository-command", Stdout: &boundedProcessOutput{limit: 1}})
	if prepared {
		t.Fatal("repository command used scanner process-group path")
	}
}

func TestModuleCardinalityBound(t *testing.T) {
	selection := Runner{}
	for i := 0; i <= inventory.MaximumModules; i++ {
		selection.Catalog.Modules = append(selection.Catalog.Modules, inventory.Module{Directory: "."})
	}
	if _, err := selection.selectModules([]string{"."}); err == nil {
		t.Fatal("unbounded module inventory accepted")
	}
}

func TestSuppressionPreflightBoundsNonGoEntries(t *testing.T) {
	for _, directories := range []bool{false, true} {
		root := t.TempDir()
		for _, name := range []string{"one", "two", "three"} {
			path := filepath.Join(root, name)
			var err error
			if directories {
				err = os.Mkdir(path, 0o700)
			} else {
				err = os.WriteFile(path, nil, 0o600)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		runner := Runner{securitySourceLimits: &securitySourceLimits{entries: 2}}
		if err := runner.checkSecurityPolicy(t.Context(), io.Discard, root, "."); err == nil || !strings.Contains(err.Error(), "entry limit") {
			t.Fatalf("non-Go flood accepted: directories=%v, error=%v", directories, err)
		}
	}
}

func TestSuppressionPreflightHonorsOperationCancellation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "invalid.go"), []byte("not Go syntax"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	runner := Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: map[string]bool{"security": true}}}}}
	if err := runner.Check(ctx, []string{"."}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled preflight parsed source instead: %v", err)
	}
}

func TestSourceBudgetsRejectBeforeArtifactCreation(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) {
		command := exec.CommandContext(t.Context(), "git", args...)
		command.Dir = root
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q")
	if err := os.WriteFile(filepath.Join(root, "file"), []byte(strings.Repeat("x", 128)), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "file")
	git("-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "-qm", "fixture")
	git("tag", "extra")
	for _, limits := range []securitySourceLimits{
		{refs: 1, objects: 100, entries: 100, bytes: 10000},
		{refs: 10, objects: 1, entries: 100, bytes: 10000},
		{refs: 10, objects: 100, entries: 100, bytes: 64},
	} {
		workspace := t.TempDir()
		artifactCommand := false
		runner := Runner{Root: root, securitySourceLimits: &limits, Executor: workspaceExecutor{directory: workspace, run: func(ctx context.Context, command Command) error {
			if strings.Contains(strings.Join(command.Args, " "), "bundle create") || strings.Contains(strings.Join(command.Args, " "), "fetch") {
				artifactCommand = true
			}
			process := exec.CommandContext(ctx, command.Name, command.Args...)
			process.Dir = command.Dir
			process.Stdout = command.Stdout
			process.Stderr = command.Stderr
			return process.Run()
		}}}
		if _, cleanup, err := runner.createGitleaksSources(t.Context()); err == nil {
			if cleanup != nil {
				_ = cleanup()
			}
			t.Fatal("oversized history accepted")
		}
		if artifactCommand {
			t.Fatal("history artifact command ran before budget rejection")
		}
		entries, err := os.ReadDir(workspace)
		if err != nil || len(entries) != 0 {
			t.Fatalf("artifact created before budget rejection: %v, %d", err, len(entries))
		}
	}
	directoriesRoot := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		if err := os.Mkdir(filepath.Join(directoriesRoot, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	destination := filepath.Join(t.TempDir(), "copy")
	if err := copyGitleaksCurrentTreeBounded(t.Context(), directoriesRoot, destination, securitySourceLimits{entries: 2, bytes: 10000}); err == nil {
		t.Fatal("directory-only entry flood accepted")
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("tree copied before preflight")
	}
}

func TestRepositorySecretScanFailurePreventsEveryModuleExecution(t *testing.T) {
	for _, stage := range []string{"missing inventory", "unknown inventory", "failed history", "failed current"} {
		t.Run(stage, func(t *testing.T) {
			root, workspace := t.TempDir(), t.TempDir()
			executed := false
			runner := Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", ModulePath: "example", Gates: map[string]bool{"security": true, "tests": true}}}}, Executor: workspaceExecutor{directory: workspace, emptySourceInventory: true, run: func(_ context.Context, command Command) error {
				if command.Name == "git" && len(command.Args) > 2 && (command.Args[2] == "for-each-ref" || command.Args[2] == "cat-file") {
					if stage == "missing inventory" {
						return nil
					}
					if stage == "unknown inventory" {
						_, _ = io.WriteString(command.Stdout, "unknown\n")
						return nil
					}
					value := "refs/heads/main\n"
					if command.Args[2] == "cat-file" {
						value = "1\n"
					}
					_, _ = io.WriteString(command.Stdout, value)
				}
				joined := strings.Join(command.Args, " ")
				if strings.Contains(joined, "gitleaks") && strings.Contains(joined, " git ") && stage == "failed history" {
					return errors.New("owned scanner failure")
				}
				if strings.Contains(joined, "gitleaks") && strings.Contains(joined, " dir ") && stage == "failed current" {
					return errors.New("owned scanner failure")
				}
				if len(command.Args) > 0 && command.Args[0] == "test" {
					executed = true
				}
				return nil
			}}}
			if err := runner.Local(t.Context(), []string{"."}); err == nil {
				t.Fatal("invalid scan accepted")
			}
			if executed {
				t.Fatal("repository tests ran after failed source scans")
			}
			entries, err := os.ReadDir(workspace)
			if err != nil || len(entries) != 0 {
				t.Fatalf("source cleanup after failed scan: %v, %d", err, len(entries))
			}
		})
	}
}

func TestCancelledBundleCreationRemovesEverySourceArtifact(t *testing.T) {
	root, workspace := t.TempDir(), t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	runner := Runner{Root: root, Executor: workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
		if strings.Contains(strings.Join(command.Args, " "), "bundle create") {
			_, _ = io.WriteString(command.Stdout, "partial")
			cancel()
		}
		return nil
	}}}
	if _, cleanup, err := runner.createGitleaksSources(ctx); err == nil || !errors.Is(err, context.Canceled) {
		if cleanup != nil {
			_ = cleanup()
		}
		t.Fatalf("cancelled creation: %v", err)
	}
	entries, err := os.ReadDir(workspace)
	if err != nil || len(entries) != 0 {
		t.Fatalf("cancelled bundle cleanup: %v, %d", err, len(entries))
	}
}

func TestBoundedSourceFileDoesNotWriteBeyondLimit(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "bundle")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	writer := &boundedSourceFile{file: file, limit: 3}
	if _, err := io.WriteString(writer, "1234"); err == nil || !writer.didOverflow() {
		t.Fatal("bundle overflow accepted")
	}
	info, err := file.Stat()
	if err != nil || info.Size() != 3 {
		t.Fatalf("bundle grew beyond budget: %v, %v", info, err)
	}
}

func TestHistoryInventoryCannotBypassOutputBoundWithReaderFrom(t *testing.T) {
	output := &sourceInventoryOutput{boundedProcessOutput: boundedProcessOutput{limit: 3}}
	_, err := io.Copy(output, strings.NewReader("1234"))
	if err != nil || !output.didOverflow() || output.Len() > 3 {
		t.Fatalf("inventory output exceeded bound: %d, %v", output.Len(), err)
	}
	if _, ok := any(output).(io.ReaderFrom); ok {
		t.Fatal("inventory exposes unbounded buffer ReaderFrom")
	}
}
