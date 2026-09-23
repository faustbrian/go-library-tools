package gates

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/docscheck"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestFormattingAndSafetyReportMalformedOrUnreadableTrees(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad.go"), []byte("package ["), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Executor: &processExecutor{stdout: io.Discard, stderr: io.Discard}}
	if err := runner.checkFormatting(context.Background(), root); err == nil || !strings.Contains(err.Error(), "run gofmt") {
		t.Fatalf("checkFormatting() error = %v", err)
	}
	if err := checkSafety(root); err == nil || !strings.Contains(err.Error(), "parse ") {
		t.Fatalf("checkSafety() error = %v", err)
	}
	if err := runner.checkFormatting(context.Background(), filepath.Join(root, "missing")); err == nil {
		t.Fatal("checkFormatting() missing root error = nil")
	}
}

func TestRunSecuritySuppressesAndBoundsScannerOutput(t *testing.T) {
	secret := []byte("credentials-AKIA1234567890-secret")
	failure := errors.New("scanner failed")
	for _, stream := range []string{"stdout", "stderr"} {
		t.Run(stream, func(t *testing.T) {
			var announcements bytes.Buffer
			runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
				if command.Stdout == nil || command.Stderr == nil {
					t.Fatal("security scanner streams are not owned")
				}
				writer := command.Stdout
				if stream == "stderr" {
					writer = command.Stderr
				}
				payload := bytes.Repeat(secret, (maximumSecurityProcessOutput/len(secret))+1)
				_, _ = writer.Write(payload[:maximumSecurityProcessOutput+1])
				return failure
			})}
			err := runner.runSecurity(context.Background(), &announcements, "/repo", inventory.Module{Directory: ".", ModulePath: "example"})
			if err == nil || !strings.Contains(err.Error(), "output exceeded") || !errors.Is(err, failure) {
				t.Fatalf("runSecurity() error = %v", err)
			}
			if bytes.Contains(announcements.Bytes(), secret) {
				t.Fatalf("runSecurity() disclosed scanner output: %q", announcements.String())
			}
		})
	}
}

func TestCheckModuleLocalPropagatesEachCommandFailure(t *testing.T) {
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":     "module example\n\ngo 1.27.0\n",
		"example.go": "package example\n",
		"README.md":  "# Example\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	module := inventory.Module{Directory: ".", Gates: map[string]bool{"lint": true, "tests": true, "documentation": true}}
	for failAt := 1; failAt <= 6; failAt++ {
		t.Run(fmt.Sprintf("command-%d", failAt), func(t *testing.T) {
			calls := 0
			runner := Runner{Root: root, Executor: executorFunction(func(context.Context, Command) error {
				calls++
				if calls == failAt {
					return errors.New("injected command failure")
				}
				return nil
			})}
			if err := runner.checkModuleLocal(context.Background(), io.Discard, module); err == nil {
				t.Fatal("checkModuleLocal() error = nil")
			}
		})
	}
}

func TestCoverageReportsOwnedFileAndExecutionFailures(t *testing.T) {
	failure := errors.New("injected failure")
	module := inventory.Module{Directory: ".", Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}}
	tests := []struct {
		name     string
		files    coverageFileSystem
		executor Executor
		want     string
	}{
		{"create", &fakeCoverageFiles{createErr: failure}, executorFunction(func(context.Context, Command) error { return nil }), "create coverage profile"},
		{"close", &fakeCoverageFiles{file: &fakeNamedFile{name: "profile", closeErr: failure}}, executorFunction(func(context.Context, Command) error { return nil }), "close coverage profile"},
		{"execute", &fakeCoverageFiles{file: &fakeNamedFile{name: "profile"}}, executorFunction(func(context.Context, Command) error { return failure }), "injected failure"},
		{"open", &fakeCoverageFiles{file: &fakeNamedFile{name: "profile"}, openErr: failure}, executorFunction(func(context.Context, Command) error { return nil }), "open coverage profile"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Executor: test.executor, coverageFiles: test.files}
			err := runner.runCoverage(context.Background(), io.Discard, t.TempDir(), module)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runCoverage() error = %v", err)
			}
		})
	}
}

func TestCoverageUsesTaskWorkspace(t *testing.T) {
	files := &fakeCoverageFiles{file: &fakeNamedFile{name: "profile"}}
	var observed Command
	executor := workspaceExecutor{directory: "/task", run: func(_ context.Context, command Command) error {
		observed = command
		return nil
	}}
	runner := Runner{Executor: executor, coverageFiles: files}
	module := inventory.Module{Directory: ".", Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}}
	if err := runner.runCoverage(context.Background(), io.Discard, t.TempDir(), module); err != nil {
		t.Fatalf("runCoverage() error = %v", err)
	}
	if files.directory != "/task" {
		t.Fatalf("coverage directory = %q", files.directory)
	}
	if got := strings.Join(observed.Args, " "); strings.Contains(got, "-tags=") {
		t.Fatalf("coverage command contains empty test tags: %q", got)
	}
}

func TestCoverageCountsModuleTestsForEachProductionPackage(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"go.mod":          "module example\n\ngo 1.26.0\n",
		"example.go":      "package example\n\nfunc Value() int { return 1 }\n",
		"example_test.go": "package example_test\n\nimport (\n\t\"testing\"\n\n\texample \"example\"\n\t\"example/child\"\n)\n\nfunc TestBehavior(t *testing.T) {\n\tif example.Value()+child.Value() != 3 {\n\t\tt.Fatal(\"unexpected value\")\n\t}\n}\n",
		"child/child.go":  "package child\n\nfunc Value() int { return 2 }\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("create process executor: %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Errorf("clean process executor: %v", err)
		}
	})
	module := inventory.Module{
		Directory: ".",
		Packages: []inventory.Package{
			{Directory: ".", ImportPath: "example", CoverageRequired: true},
			{Directory: "child", ImportPath: "example/child", CoverageRequired: true},
		},
	}
	if err := (Runner{Executor: executor}).runCoverage(context.Background(), io.Discard, root, module); err != nil {
		t.Fatalf("runCoverage() error = %v", err)
	}
}

func TestCoverageRejectsModulesWithoutRequiredPackages(t *testing.T) {
	runner := Runner{Executor: executorFunction(func(context.Context, Command) error {
		t.Fatal("coverage command ran without a required package")
		return nil
	})}
	err := runner.runCoverage(context.Background(), io.Discard, t.TempDir(), inventory.Module{})
	if err == nil || !strings.Contains(err.Error(), "no coverage-required packages") {
		t.Fatalf("runCoverage() error = %v", err)
	}
}

func TestGitleaksConfigRequiresTaskWorkspace(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".gitleaks.toml"), []byte("title = \"fixture\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name     string
		executor Executor
		root     string
		want     string
	}{
		{"no workspace", executorFunction(func(context.Context, Command) error { return nil }), root, "task-owned"},
		{"relative workspace", workspaceExecutor{directory: "relative", run: func(context.Context, Command) error { return nil }}, root, "task-owned"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := (Runner{Root: test.root, Executor: test.executor}).createGitleaksConfig()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("createGitleaksConfig() error = %v", err)
			}
		})
	}
}

func TestGitleaksConfigReportsOwnedFileFailures(t *testing.T) {
	failure := errors.New("injected failure")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".gitleaks.toml"), []byte("title = \"fixture\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		files *fakeSecretConfigFiles
		want  string
	}{
		{"create", &fakeSecretConfigFiles{createErr: failure}, "create temporary"},
		{"write", &fakeSecretConfigFiles{file: &fakeNamedFile{name: "config", writeErr: failure}}, "write temporary"},
		{"close", &fakeSecretConfigFiles{file: &fakeNamedFile{name: "config", closeErr: failure}}, "close temporary"},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{
				Root: root,
				Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error {
					return nil
				}},
				secretConfigFiles: test.files,
			}
			_, _, err := runner.createGitleaksConfig()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("createGitleaksConfig() error = %v", err)
			}
			if test.name != "create" && test.files.removed != "config" {
				t.Fatalf("removed path = %q", test.files.removed)
			}
		})
	}
}

func TestCheckReportsAnalysisConfigCleanupFailure(t *testing.T) {
	failure := errors.New("injected cleanup failure")
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":         "module example\n\ngo 1.27.0\n",
		"example.go":     "package example\n",
		".gitleaks.toml": "title = \"fixture\"\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executor := workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return nil }}
	files := &fakeSecretConfigFiles{file: &fakeNamedFile{name: "config"}, removeErr: failure}
	runner := Runner{
		Root: root,
		Catalog: inventory.Inventory{Modules: []inventory.Module{{
			Directory: ".", ModulePath: "example", Gates: map[string]bool{"security": true},
		}}},
		Executor: executor, secretConfigFiles: files,
	}
	if err := runner.Check(context.Background(), []string{"."}); err == nil || !strings.Contains(err.Error(), "remove temporary analysis config") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestRunSecurityPreservesGitleaksScannerAndCleanupFailures(t *testing.T) {
	for _, stage := range []string{"git", "dir"} {
		t.Run(stage, func(t *testing.T) {
			scannerFailure := errors.New("injected scanner failure")
			cleanupFailure := errors.New("injected cleanup failure")
			files := &securityPolicyFiles{gitleaksRemoveErr: cleanupFailure}
			var commands []string
			runner := Runner{
				Root: t.TempDir(),
				Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
					joined := strings.Join(command.Args, " ")
					commands = append(commands, joined)
					if strings.Contains(joined, "gitleaks") && strings.Contains(joined, " "+stage+" ") {
						return scannerFailure
					}
					return nil
				}},
				secretConfigFiles: files,
			}
			err := runner.runSecurity(context.Background(), io.Discard, runner.Root, inventory.Module{Directory: ".", ModulePath: "example"})
			if !errors.Is(err, scannerFailure) || !errors.Is(err, cleanupFailure) {
				t.Fatalf("runSecurity() error = %v", err)
			}
			if files.gitleaksRemoved != 1 {
				t.Fatalf("gitleaks cleanup calls = %d", files.gitleaksRemoved)
			}
			for _, command := range commands {
				if strings.Contains(command, "go-licenses") || strings.Contains(command, "cyclonedx-gomod") {
					t.Fatalf("later gate ran after scanner failure: %q", command)
				}
			}
		})
	}
}

func TestRunSecurityStopsAfterGitleaksCleanupFailure(t *testing.T) {
	cleanupFailure := errors.New("injected cleanup failure")
	files := &securityPolicyFiles{gitleaksRemoveErr: cleanupFailure}
	var commands []string
	runner := Runner{
		Root: t.TempDir(),
		Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
			commands = append(commands, strings.Join(command.Args, " "))
			return nil
		}},
		secretConfigFiles: files,
	}
	err := runner.runSecurity(context.Background(), io.Discard, runner.Root, inventory.Module{Directory: ".", ModulePath: "example"})
	if !errors.Is(err, cleanupFailure) || !strings.Contains(err.Error(), "remove temporary gitleaks config") {
		t.Fatalf("runSecurity() error = %v", err)
	}
	for _, command := range commands {
		if strings.Contains(command, "go-licenses") || strings.Contains(command, "cyclonedx-gomod") {
			t.Fatalf("later gate ran after cleanup failure: %q", command)
		}
	}
}

func TestRunSecurityCleansGitleaksConfigBeforeLaterGates(t *testing.T) {
	stop := errors.New("stop at licenses")
	files := &securityPolicyFiles{}
	runner := Runner{
		Root: t.TempDir(),
		Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
			if strings.Contains(strings.Join(command.Args, " "), "go-licenses") {
				if files.gitleaksRemoved != 1 {
					t.Fatalf("gitleaks cleanup calls before licenses = %d", files.gitleaksRemoved)
				}
				return stop
			}
			return nil
		}},
		secretConfigFiles: files,
	}
	err := runner.runSecurity(context.Background(), io.Discard, runner.Root, inventory.Module{Directory: ".", ModulePath: "example"})
	if !errors.Is(err, stop) {
		t.Fatalf("runSecurity() error = %v", err)
	}
}

func TestGeneratedGitleaksPolicyIntegration(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("set GOLIB_GITLEAKS_INTEGRATION=1 to run the pinned scanner contract")
	}
	allowed := strings.Join([]string{"sk_", "test_", "0123456789abcdef", "ghijklmnopqrstuv"}, "")
	alternateStripe := strings.Join([]string{"sk_", "test_", "0123456789abcdef", "ghijklmnopqrstuw"}, "")
	githubPAT := strings.Join([]string{"gh", "p_", "A1b2C3d4E5f6", "G7h8I9j0K1l2", "M3n4O5p6Q7r8"}, "")

	t.Run("exact tuple in history and current tree", func(t *testing.T) {
		root, configPath := gitleaksRepository(t, map[string]string{
			"internal/inventory/inventory_test.go": allowed,
			".gitleaks.toml":                       "malformed = [",
		})
		for _, mode := range []string{"git", "dir"} {
			findings, output, err := runGitleaksIntegration(t, root, configPath, mode)
			if err != nil || len(findings) != 0 {
				t.Fatalf("gitleaks %s findings = %#v, error = %v, output = %q", mode, findings, err, output)
			}
		}
	})

	for _, test := range []struct {
		name              string
		files             map[string]string
		wantRules         []string
		assertExactAbsent bool
	}{
		{
			name: "value mismatch",
			files: map[string]string{
				"internal/inventory/inventory_test.go": alternateStripe,
			},
			wantRules: []string{"stripe-access-token"},
		},
		{
			name: "path mismatch",
			files: map[string]string{
				"internal/inventory/other_test.go": allowed,
			},
			wantRules: []string{"stripe-access-token"},
		},
		{
			name: "rule mismatch",
			files: map[string]string{
				"internal/inventory/inventory_test.go": githubPAT,
			},
			wantRules: []string{"github-pat"},
		},
		{
			name: "mixed findings",
			files: map[string]string{
				"internal/inventory/inventory_test.go": allowed,
				"internal/inventory/other_test.go":     allowed,
				"security/github.txt":                  githubPAT,
			},
			wantRules:         []string{"stripe-access-token", "github-pat"},
			assertExactAbsent: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, configPath := gitleaksRepository(t, test.files)
			findings, output, err := runGitleaksIntegration(t, root, configPath, "dir")
			if err == nil {
				t.Fatal("gitleaks error = nil")
			}
			if strings.Contains(output, allowed) || strings.Contains(output, alternateStripe) || strings.Contains(output, githubPAT) {
				t.Fatalf("gitleaks output disclosed a synthetic secret: %q", output)
			}
			for _, rule := range test.wantRules {
				if !slices.ContainsFunc(findings, func(finding gitleaksFinding) bool { return finding.RuleID == rule }) {
					t.Fatalf("gitleaks findings = %#v, missing rule %q", findings, rule)
				}
			}
			if test.assertExactAbsent && slices.ContainsFunc(findings, func(finding gitleaksFinding) bool {
				return finding.RuleID == "stripe-access-token" &&
					filepath.ToSlash(finding.File) == "internal/inventory/inventory_test.go"
			}) {
				t.Fatalf("exact allowlisted tuple was reported: %#v", findings)
			}
		})
	}
}

type gitleaksFinding struct {
	RuleID string `json:"RuleID"`
	File   string `json:"File"`
}

func gitleaksRepository(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	root := t.TempDir()
	for path, content := range files {
		target := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	configPath := filepath.Join(root, "generated-gitleaks.toml")
	if err := os.WriteFile(configPath, []byte(gitleaksPolicy), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{
		{"init", "-q"},
		{"add", "."},
		{"-c", "user.name=golib-test", "-c", "user.email=golib-test@example.invalid", "commit", "-qm", "fixture"},
	} {
		// #nosec G204 -- the executable and arguments are fixed test-owned Git operations.
		command := exec.CommandContext(t.Context(), "git", arguments...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", arguments, err, output)
		}
	}
	return root, configPath
}

func runGitleaksIntegration(t *testing.T, root, configPath, mode string) ([]gitleaksFinding, string, error) {
	t.Helper()
	reportPath := filepath.Join(t.TempDir(), "gitleaks.json")
	arguments := []string{"run", "github.com/zricethezav/gitleaks/v8@v8.30.1", mode, ".", "--config", configPath, "--no-banner", "--redact", "--report-format", "json", "--report-path", reportPath}
	if mode == "git" {
		arguments = append(arguments, "--log-opts=--all")
	}
	// #nosec G204 -- the executable and scanner arguments are fixed; paths are test-owned.
	command := exec.CommandContext(t.Context(), "go", arguments...)
	command.Dir = root
	output, err := command.CombinedOutput()
	report, readErr := os.ReadFile(reportPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatal(readErr)
	}
	var findings []gitleaksFinding
	if len(report) > 0 {
		if decodeErr := json.Unmarshal(report, &findings); decodeErr != nil {
			t.Fatalf("decode gitleaks report: %v: %s", decodeErr, report)
		}
	}
	return findings, string(output), err
}

func TestSecuritySuppressionsRequireNarrowRulesAndReasons(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	for _, test := range []struct {
		name    string
		content string
		want    string
	}{
		{"broad nosec", "package example\n// #nosec -- broad\nfunc value() {}\n", "exact rule IDs"},
		{"unexplained nolint", "package example\nfunc value() int { return 1 } //nolint:gosec\n", "inline reason"},
		{"documented", "package example\n// #nosec G304 -- fixed repository-owned path\nfunc value() int { return 1 } //nolint:gosec // bounded conversion\n", ""},
		{"directive text in string", "package example\nconst guidance = \"#nosec requires exact rule IDs\"\n", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(test.content), 0o600); err != nil {
				t.Fatal(err)
			}
			err := checkSecuritySuppressions(root)
			if test.want == "" && err != nil {
				t.Fatalf("checkSecuritySuppressions() error = %v", err)
			}
			if test.want != "" && (err == nil || !strings.Contains(err.Error(), test.want)) {
				t.Fatalf("checkSecuritySuppressions() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestSecuritySuppressionsValidateEveryDirectiveInACommentGroup(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	content := "package example\n// #nosec G304 -- fixed repository-owned path\n// #nosec\nfunc value() {}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	err := checkSecuritySuppressions(root)
	if err == nil || !strings.Contains(err.Error(), "example.go:3:") || !strings.Contains(err.Error(), "exact rule IDs") {
		t.Fatalf("checkSecuritySuppressions() error = %v", err)
	}
}

func TestSecuritySuppressionsAcceptEveryNativeSpelling(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	for _, test := range []struct {
		name    string
		comment string
	}{
		{"spaced line", "// #nosec G304 -- fixed repository-owned path"},
		{"compact line", "//#nosec G304 -- fixed repository-owned path"},
		{"compact block", "/*#nosec G304 -- fixed repository-owned path*/"},
		{"spaced block", "/* #nosec G304 -- fixed repository-owned path */"},
		{"multiline block", "/*\n#nosec G304 -- fixed repository-owned path\n*/"},
		{"gosec disable", "//gosec:disable G304 -- fixed repository-owned path"},
		{"nosec continuation", "//#nosecG304 -- fixed repository-owned path"},
	} {
		t.Run(test.name, func(t *testing.T) {
			content := "package example\n" + test.comment + "\nfunc value() {}\n"
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := checkSecuritySuppressions(root); err != nil {
				t.Fatalf("checkSecuritySuppressions() error = %v", err)
			}
		})
	}
}

func TestNativeSecurityDirectiveMatchesGosecDisableTokenBoundary(t *testing.T) {
	for _, test := range []struct {
		name      string
		line      string
		wantFound bool
		wantArgs  string
	}{
		{"exact token", "//gosec:disable", true, ""},
		{"space delimiter", "//gosec:disable G304 -- bounded", true, "G304 -- bounded"},
		{"tab delimiter", "//gosec:disable\tG304 -- bounded", false, ""},
		{"identifier continuation", "//gosec:disablement", false, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			arguments, found := nativeSecurityDirective(test.line, true)
			if found != test.wantFound || arguments != test.wantArgs {
				t.Fatalf("nativeSecurityDirective() = (%q, %t), want (%q, %t)", arguments, found, test.wantArgs, test.wantFound)
			}
		})
	}
}

func TestNativeSecurityDirectiveMatchesGosecNosecPrefixBoundary(t *testing.T) {
	for _, test := range []struct {
		name      string
		line      string
		wantFound bool
		wantArgs  string
	}{
		{"compact token", "//#nosec", true, ""},
		{"continuation", "//#nosecG304 -- bounded", true, "G304 -- bounded"},
		{"prose", "// guidance about #nosec", false, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			arguments, found := nativeSecurityDirective(test.line, true)
			if found != test.wantFound || arguments != test.wantArgs {
				t.Fatalf("nativeSecurityDirective() = (%q, %t), want (%q, %t)", arguments, found, test.wantArgs, test.wantFound)
			}
		})
	}
}

func TestNativeNosecGroupDirectiveIgnoresBlockInteriorLineMarker(t *testing.T) {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, "example.go", "package example\n/*\n//#nosec\n*/\nfunc value() {}\n", parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if arguments, line, found := nativeNosecGroupDirective(parsed.Comments[0], fileSet); found {
		t.Fatalf("nativeNosecGroupDirective() = (%q, %d, true), want not found", arguments, line)
	}
}

func TestNolintGosecDirectiveNormalizesProviderWhitespaceAndCase(t *testing.T) {
	for _, test := range []struct {
		comment    string
		wantFound  bool
		wantReason string
	}{
		{"//nolint:govet, gosec", true, ""},
		{"//nolint: gosec // bounded", true, " bounded"},
		{"//nolint: GOSEC", true, ""},
		{"//nolint:govet, GoSeC", true, ""},
		{"//nolint:govet", false, ""},
	} {
		t.Run(test.comment, func(t *testing.T) {
			reason, found := nolintGosecDirective(test.comment)
			if found != test.wantFound || reason != test.wantReason {
				t.Fatalf("nolintGosecDirective() = (%q, %t), want (%q, %t)", reason, found, test.wantReason, test.wantFound)
			}
		})
	}
}

func TestSecuritySuppressionsRejectBroadAndMalformedNativeDirectives(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	for _, test := range []struct {
		name    string
		comment string
		want    string
	}{
		{"spaced line without rules", "// #nosec -- broad", "exact rule IDs"},
		{"compact line without rules", "//#nosec -- broad", "exact rule IDs"},
		{"bare disable", "//gosec:disable", "exact rule IDs"},
		{"disable without rules", "//gosec:disable -- broad", "exact rule IDs"},
		{"compact block without rules", "/*#nosec*/", "exact rule IDs"},
		{"compact block without reason", "/*#nosec G304*/", "reason after --"},
		{"lowercase rule", "//#nosec g304 -- broad", "exact rule IDs"},
		{"short rule", "//#nosec G30 -- broad", "exact rule IDs"},
		{"long rule", "//#nosec G0304 -- broad", "exact rule IDs"},
		{"nondigit rule", "//#nosec GABC -- broad", "exact rule IDs"},
		{"wildcard rule", "//#nosec G304,all -- broad", "exact rule IDs"},
		{"space separated rules", "//#nosec G304 G305 -- broad", "exact rule IDs"},
		{"trailing comma", "//#nosec G304, -- broad", "exact rule IDs"},
		{"missing delimiter", "//#nosec G304 reason", "reason after --"},
		{"blank reason", "//#nosec G304 --   ", "reason after --"},
		{"hyphen-only reason", "//#nosec G304 ---", "reason after --"},
		{"block terminator is not reason", "/*#nosec G304 -- */", "reason after --"},
	} {
		t.Run(test.name, func(t *testing.T) {
			content := "package example\n" + test.comment + "\nfunc value() {}\n"
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			err := checkSecuritySuppressions(root)
			if err == nil || !strings.Contains(err.Error(), "example.go:2:") || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("checkSecuritySuppressions() error = %v, want line and %q", err, test.want)
			}
		})
	}
}

func TestSecuritySuppressionsRecognizeMalformedSpacedAndMultilineBlocks(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	for _, test := range []struct {
		name    string
		comment string
		line    string
		want    string
	}{
		{"spaced block without rules", "/* #nosec */", "example.go:2:", "exact rule IDs"},
		{"spaced block without reason", "/* #nosec G304 */", "example.go:2:", "reason after --"},
		{"multiline block without rules", "/*\n#nosec\n*/", "example.go:3:", "exact rule IDs"},
		{"multiline block without reason", "/*\n#nosec G304\n*/", "example.go:3:", "reason after --"},
	} {
		t.Run(test.name, func(t *testing.T) {
			content := "package example\n" + test.comment + "\nfunc value() {}\n"
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			err := checkSecuritySuppressions(root)
			if err == nil || !strings.Contains(err.Error(), test.line) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("checkSecuritySuppressions() error = %v, want line %q and %q", err, test.line, test.want)
			}
		})
	}
}

func TestSecuritySuppressionsAcceptExactMultipleRules(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	for _, comment := range []string{
		"//#nosec G304,G305 -- both paths are bounded",
		"//gosec:disable G304,G305 -- both paths are bounded",
		"//#nosec G304, G305 -- both paths are bounded",
		"//gosec:disable G304, G305 -- both paths are bounded",
		"//#nosec G304,G304 -- repeated exact rule remains narrow",
		"//gosec:disable G304,G304 -- repeated exact rule remains narrow",
	} {
		if err := os.WriteFile(path, []byte("package example\n"+comment+"\nfunc value() {}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := checkSecuritySuppressions(root); err != nil {
			t.Fatalf("checkSecuritySuppressions(%q) error = %v", comment, err)
		}
	}
}

func TestSecuritySuppressionsKeepNolintGosecLineScopedAndReasoned(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	for _, test := range []struct {
		name    string
		comment string
		wantErr bool
	}{
		{"dedicated", "//nolint:gosec // bounded conversion", false},
		{"multiple linters", "//nolint:govet,gosec // bounded conversion", false},
		{"spaced multiple linters", "//nolint:govet, gosec // bounded conversion", false},
		{"spaced dedicated", "//nolint: gosec // bounded conversion", false},
		{"uppercase dedicated", "//nolint: GOSEC // bounded conversion", false},
		{"missing reason", "//nolint:gosec", true},
		{"blank inline reason", "//nolint:gosec //   ", true},
		{"multiple linters missing reason", "//nolint:govet,gosec", true},
		{"spaced multiple linters missing reason", "//nolint:govet, gosec", true},
		{"spaced dedicated missing reason", "//nolint: gosec", true},
		{"uppercase missing reason", "//nolint: GOSEC", true},
		{"mixed case missing reason", "//nolint:govet, GoSeC", true},
		{"following line reason", "//nolint:gosec\n// bounded conversion", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte("package example\nfunc value() int { return 1 } "+test.comment+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			err := checkSecuritySuppressions(root)
			if test.wantErr && (err == nil || !strings.Contains(err.Error(), "example.go:2:") || !strings.Contains(err.Error(), "inline reason")) {
				t.Fatalf("checkSecuritySuppressions() error = %v", err)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("checkSecuritySuppressions() error = %v", err)
			}
		})
	}
}

func TestSecuritySuppressionsIgnoreNonDirectiveTextAndRejectMixedBatches(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "example.go")
	for _, content := range []string{
		"package example\nconst guidance = \"//#nosec and //gosec:disable require reasons\"\n",
		"package example\nconst guidance = `//#nosec and //gosec:disable require reasons`\n",
		"package example\n// Guidance about #nosec and //gosec:disable syntax.\nfunc value() {}\n",
		"package example\n// Guidance about //nolint:gosec syntax.\nfunc value() {}\n",
		"package example\n// gosec:disable documents scanner behavior.\nfunc value() {}\n",
		"package example\n/* gosec:disable documents scanner behavior. */\nfunc value() {}\n",
		"package example\n/*\n//gosec:disable\n*/\nfunc value() {}\n",
		"package example\n/*\n//#nosec\n*/\nfunc value() {}\n",
		"package example\n/*\n * #nosec\n */\nfunc value() {}\n",
		"package example\n//gosec:disable\tG304 -- bounded\nfunc value() {}\n",
		"package example\n//gosec:disablement\nfunc value() {}\n",
		"package example\n//#nosec G304 -- bounded\nfunc first() {}\n//gosec:disable G305 -- bounded\nfunc second() {}\n",
	} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := checkSecuritySuppressions(root); err != nil {
			t.Fatalf("checkSecuritySuppressions() error = %v", err)
		}
	}
	mixed := "package example\n//#nosec G304 -- bounded\nfunc first() {}\n//gosec:disable\nfunc second() {}\n"
	if err := os.WriteFile(path, []byte(mixed), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkSecuritySuppressions(root); err == nil || !strings.Contains(err.Error(), "example.go:4:") {
		t.Fatalf("checkSecuritySuppressions() error = %v", err)
	}
}

func TestSecuritySuppressionPreflightStopsScanners(t *testing.T) {
	for _, test := range []struct {
		comment string
		want    string
	}{
		{"//#nosec", "exact rule IDs"},
		{"//gosec:disable G304", "reason after --"},
		{"//#nosecG304", "reason after --"},
		{"//#nosec G304 ---", "reason after --"},
		{"//nolint:govet, gosec", "inline reason"},
		{"//nolint: gosec", "inline reason"},
		{"//nolint: GOSEC", "inline reason"},
	} {
		t.Run(test.comment, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "example.go"), []byte("package example\n"+test.comment+"\nfunc value() {}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			commands := 0
			var output bytes.Buffer
			runner := Runner{Executor: executorFunction(func(context.Context, Command) error {
				commands++
				return nil
			})}
			err := runner.checkSecurity(context.Background(), &output, root, inventory.Module{Directory: "."})
			if err == nil || !strings.Contains(err.Error(), "example.go:2:") || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("checkSecurity() error = %v", err)
			}
			if commands != 0 {
				t.Fatalf("security commands = %d, want 0", commands)
			}
			if got := output.String(); got != "[.] security-suppressions\n" {
				t.Fatalf("security output = %q", got)
			}
		})
	}
}

func TestCheckAndLocalPreflightAllSecurityModulesBeforeServicesAndCommands(t *testing.T) {
	for _, caller := range []string{"check", "local"} {
		for _, scenario := range []struct {
			name    string
			modules []inventory.Module
			selects []string
		}{
			{
				name:    "service-backed invalid singleton",
				modules: []inventory.Module{{Directory: ".", ModulePath: "example", Gates: map[string]bool{"security": true}, RequiredServices: []string{"postgresql"}}},
				selects: []string{"."},
			},
			{
				name: "valid security module before invalid security module",
				modules: []inventory.Module{
					{Directory: "a-valid", ModulePath: "example/a", Gates: map[string]bool{"security": true}, RequiredServices: []string{"postgresql"}},
					{Directory: "z-invalid", ModulePath: "example/z", Gates: map[string]bool{"security": true}},
				},
				selects: []string{"z-invalid", "a-valid", "z-invalid"},
			},
			{
				name: "nonsecurity module before invalid security module",
				modules: []inventory.Module{
					{Directory: "a-plain", ModulePath: "example/a", RequiredServices: []string{"postgresql"}},
					{Directory: "z-invalid", ModulePath: "example/z", Gates: map[string]bool{"security": true}},
				},
				selects: []string{"z-invalid", "a-plain", "z-invalid"},
			},
		} {
			t.Run(caller+" "+scenario.name, func(t *testing.T) {
				root := t.TempDir()
				for _, module := range scenario.modules {
					directory := filepath.Join(root, module.Directory)
					if err := os.MkdirAll(directory, 0o700); err != nil {
						t.Fatal(err)
					}
					source := "package example\nfunc value() {}\n"
					if module.Directory == "." || module.Directory == "z-invalid" {
						source = "package example\n//#nosec\nfunc value() {}\n"
					}
					if err := os.WriteFile(filepath.Join(directory, "example.go"), []byte(source), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				starts := 0
				commands := 0
				var output bytes.Buffer
				runner := Runner{
					Root: root, Catalog: inventory.Inventory{Modules: scenario.modules}, Output: &output,
					Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error {
						commands++
						return nil
					}},
					startServices: func(context.Context, []string) (serviceLease, error) {
						starts++
						return &fakeServiceLease{}, nil
					},
				}
				var err error
				if caller == "check" {
					err = runner.Check(context.Background(), scenario.selects)
				} else {
					err = runner.Local(context.Background(), scenario.selects)
				}
				if err == nil || !strings.Contains(err.Error(), "exact rule IDs") {
					t.Fatalf("%s error = %v", caller, err)
				}
				wantOutput := "[.] security-suppressions\n"
				wantPath := "example.go:2:"
				switch scenario.name {
				case "valid security module before invalid security module":
					wantOutput = "[a-valid] security-suppressions\n[z-invalid] security-suppressions\n"
					wantPath = filepath.Join("z-invalid", "example.go") + ":2:"
				case "nonsecurity module before invalid security module":
					wantOutput = "[z-invalid] security-suppressions\n"
					wantPath = filepath.Join("z-invalid", "example.go") + ":2:"
				}
				if !strings.Contains(err.Error(), wantPath) || output.String() != wantOutput {
					t.Fatalf("%s error/output = %v/%q, want path %q and output %q", caller, err, output.String(), wantPath, wantOutput)
				}
				if starts != 0 || commands != 0 {
					t.Fatalf("%s starts/commands = %d/%d, want 0/0", caller, starts, commands)
				}
			})
		}
	}
}

func TestCheckAndLocalPreflightValidModulesExactlyOnceBeforeExecution(t *testing.T) {
	for _, caller := range []string{"check", "local"} {
		t.Run(caller, func(t *testing.T) {
			root := t.TempDir()
			modules := []inventory.Module{
				{Directory: "a-valid", ModulePath: "example/a", Gates: map[string]bool{"security": true}, RequiredServices: []string{"postgresql"}},
				{Directory: "z-valid", ModulePath: "example/z", Gates: map[string]bool{"security": true}, RequiredServices: []string{"valkey"}},
			}
			for _, module := range modules {
				directory := filepath.Join(root, module.Directory)
				if err := os.MkdirAll(directory, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(directory, "example.go"), []byte("package example\nfunc value() {}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var events []string
			output := &securityPreflightEventWriter{events: &events}
			var leases []*securityPreflightTestLease
			commands := 0
			moduleCommands := map[string]int{}
			runner := Runner{
				Root: root, Catalog: inventory.Inventory{Modules: modules}, Output: output,
				Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
					commands++
					module := filepath.Base(command.Dir)
					moduleCommands[module]++
					events = append(events, "exec:"+module)
					if command.Stdout != nil && strings.Contains(strings.Join(command.Args, " "), "cyclonedx-gomod") {
						_, _ = io.WriteString(command.Stdout, `{"bomFormat":"CycloneDX","specVersion":"1.6"}`)
					}
					return nil
				}},
				startServices: func(_ context.Context, names []string) (serviceLease, error) {
					name := strings.Join(names, ",")
					events = append(events, "start:"+name)
					lease := &securityPreflightTestLease{name: name, events: &events}
					leases = append(leases, lease)
					return lease, nil
				},
			}
			var err error
			if caller == "check" {
				err = runner.Check(context.Background(), []string{"z-valid", "a-valid", "z-valid", "a-valid"})
			} else {
				err = runner.Local(context.Background(), []string{"z-valid", "a-valid", "z-valid", "a-valid"})
			}
			if err != nil {
				t.Fatalf("%s error = %v", caller, err)
			}
			if strings.Count(output.String(), "[a-valid] security-suppressions\n") != 1 || strings.Count(output.String(), "[z-valid] security-suppressions\n") != 1 {
				t.Fatalf("%s suppression output = %q", caller, output.String())
			}
			if len(events) < 2 || events[0] != "output:[a-valid] security-suppressions\n" || events[1] != "output:[z-valid] security-suppressions\n" {
				t.Fatalf("%s phase trace = %#v", caller, events)
			}
			if len(leases) != 2 || leases[0].name != "postgresql" || leases[1].name != "valkey" || leases[0].closes != 1 || leases[1].closes != 1 ||
				commands == 0 || moduleCommands["a-valid"] == 0 || moduleCommands["z-valid"] == 0 {
				t.Fatalf("%s leases/commands = %#v/%d", caller, leases, commands)
			}
			for _, event := range events[2:] {
				if strings.Contains(event, "security-suppressions") {
					t.Fatalf("%s late suppression event in %#v", caller, events)
				}
			}
		})
	}
}

type securityPreflightEventWriter struct {
	bytes.Buffer
	events *[]string
}

func (writer *securityPreflightEventWriter) Write(value []byte) (int, error) {
	*writer.events = append(*writer.events, "output:"+string(value))
	return writer.Buffer.Write(value)
}

type securityPreflightTestLease struct {
	name   string
	events *[]string
	closes int
}

func (lease *securityPreflightTestLease) Environment() map[string]string { return nil }
func (lease *securityPreflightTestLease) Identities() map[string]string  { return nil }
func (lease *securityPreflightTestLease) Close(context.Context) error {
	lease.closes++
	*lease.events = append(*lease.events, "close:"+lease.name)
	return nil
}

func FuzzSecuritySuppressionsNeverPanic(f *testing.F) {
	for _, seed := range []string{
		"// #nosec G304 -- bounded",
		"//#nosec",
		"/*#nosec G304*/",
		"//gosec:disable G304 -- bounded",
		"//nolint:gosec // bounded",
		"/*\n//#nosec\n*/",
		"//#nosecG304",
		"//#nosec G304 ---",
		"//nolint:govet, gosec",
		"//nolint: gosec",
		"//nolint: GOSEC",
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, payload []byte) {
		if len(payload) > 4096 {
			t.Skip()
		}
		root := t.TempDir()
		content := append([]byte("package example\n"), payload...)
		if err := os.WriteFile(filepath.Join(root, "example.go"), content, 0o600); err != nil {
			t.Fatal(err)
		}
		_ = checkSecuritySuppressions(root)
	})
}

func TestStandaloneGatesContinuePastDisabledModules(t *testing.T) {
	root := t.TempDir()
	enabled := root
	if err := os.WriteFile(filepath.Join(enabled, "README.md"), []byte("# Enabled\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	modules := []inventory.Module{
		{Directory: "a-disabled"},
		{Directory: ".", Gates: map[string]bool{"coverage": true, "documentation": true}, Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}},
		{Directory: "z-enabled", Gates: map[string]bool{"coverage": true, "documentation": true}, Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}},
	}
	commands := 0
	spellingChecks := 0
	linkChecks := 0
	runner := Runner{
		Root: root, Catalog: inventory.Inventory{Modules: modules},
		Executor: executorFunction(func(context.Context, Command) error {
			commands++
			return nil
		}),
		coverageFiles: &fakeCoverageFiles{file: &fakeNamedFile{name: "profile"}},
		DocumentationSpelling: func(context.Context, string) error {
			spellingChecks++
			return nil
		},
		DocumentationLinks: func(context.Context, string) error {
			linkChecks++
			return nil
		},
	}
	selection := []string{"a-disabled", ".", "z-enabled"}
	if err := runner.Coverage(context.Background(), selection); err != nil {
		t.Fatalf("Coverage() error = %v", err)
	}
	if err := runner.Docs(context.Background(), selection); err != nil {
		t.Fatalf("Docs() error = %v", err)
	}
	if commands != 3 {
		t.Fatalf("enabled gate command runs = %d, want 3", commands)
	}
	if spellingChecks != 1 || linkChecks != 1 {
		t.Fatalf("enabled documentation checks = spelling %d, links %d", spellingChecks, linkChecks)
	}
}

func TestDocumentationSpellingUsesPinnedTaskOwnedInstall(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	task, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	commands := make([]Command, 0, 2)
	executor := workspaceExecutor{directory: task, run: func(_ context.Context, command Command) error {
		commands = append(commands, command)
		return nil
	}}
	runner := Runner{Executor: executor}
	if err := runner.runDocumentationSpelling(context.Background(), root); err != nil {
		t.Fatalf("runDocumentationSpelling() error = %v", err)
	}
	if len(commands) != 2 || commands[0].Name != "npm" || !strings.HasSuffix(commands[1].Name, filepath.Join("node_modules", ".bin", "cspell")) {
		t.Fatalf("commands = %#v", commands)
	}
	if commands[0].Dir != filepath.Join(task, "documentation", "spelling") ||
		!strings.HasPrefix(commands[0].Env["NPM_CONFIG_CACHE"], task+string(filepath.Separator)) {
		t.Fatalf("npm command is not task-owned: %#v", commands[0])
	}
	for _, name := range []string{"package.json", "package-lock.json"} {
		if info, err := os.Stat(filepath.Join(commands[0].Dir, name)); err != nil || !info.Mode().IsRegular() {
			t.Fatalf("asset %s = %v, %v", name, info, err)
		}
	}
}

func TestDocumentationSpellingRequiresTaskWorkspace(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Executor: executorFunction(func(context.Context, Command) error { return nil })}
	if err := runner.runDocumentationSpelling(context.Background(), root); err == nil || !strings.Contains(err.Error(), "task-owned workspace") {
		t.Fatalf("runDocumentationSpelling() error = %v", err)
	}
}

func TestDocumentationSpellingReportsConfigurationAndWorkspaceFailures(t *testing.T) {
	t.Run("configuration", func(t *testing.T) {
		runner := Runner{Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return nil }}}
		if err := runner.runDocumentationSpelling(context.Background(), t.TempDir()); err == nil || !strings.Contains(err.Error(), "read cspell configuration") {
			t.Fatalf("runDocumentationSpelling() error = %v", err)
		}
	})

	t.Run("tool root", func(t *testing.T) {
		root := spellingRepository(t)
		workspace := filepath.Join(t.TempDir(), "workspace")
		if err := os.WriteFile(workspace, []byte("occupied"), 0o600); err != nil {
			t.Fatal(err)
		}
		runner := Runner{Executor: workspaceExecutor{directory: workspace, run: func(context.Context, Command) error { return nil }}}
		if err := runner.runDocumentationSpelling(context.Background(), root); err == nil || !strings.Contains(err.Error(), "create spelling tool root") {
			t.Fatalf("runDocumentationSpelling() error = %v", err)
		}
	})

	t.Run("asset", func(t *testing.T) {
		root := spellingRepository(t)
		workspace := t.TempDir()
		packagePath := filepath.Join(workspace, "documentation", "spelling", "package.json")
		if err := os.MkdirAll(packagePath, 0o700); err != nil {
			t.Fatal(err)
		}
		runner := Runner{Executor: workspaceExecutor{directory: workspace, run: func(context.Context, Command) error { return nil }}}
		if err := runner.runDocumentationSpelling(context.Background(), root); err == nil || !strings.Contains(err.Error(), "write spelling tool package.json") {
			t.Fatalf("runDocumentationSpelling() error = %v", err)
		}
	})
}

func TestDocumentationSpellingReportsToolFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name      string
		failAt    int
		want      string
		wantCalls int
	}{
		{name: "install", failAt: 1, want: "install pinned spelling tool", wantCalls: 1},
		{name: "spellcheck", failAt: 2, want: "check documentation spelling", wantCalls: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			executor := workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error {
				calls++
				if calls == test.failAt {
					return failure
				}
				return nil
			}}
			runner := Runner{Executor: executor}
			if err := runner.runDocumentationSpelling(context.Background(), spellingRepository(t)); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runDocumentationSpelling() error = %v", err)
			}
			if calls != test.wantCalls {
				t.Fatalf("tool calls = %d, want %d", calls, test.wantCalls)
			}
		})
	}
}

func spellingRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCheckDocumentationReportsNativeAndSpellingFailures(t *testing.T) {
	failure := errors.New("injected failure")
	runner := Runner{DocumentationSpelling: func(context.Context, string) error { return failure }}
	rootModule := inventory.Module{Directory: "."}
	if err := runner.checkDocumentation(context.Background(), t.TempDir(), rootModule); err == nil || errors.Is(err, failure) {
		t.Fatalf("checkDocumentation() native error = %v", err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runner.checkDocumentation(context.Background(), root, rootModule); !errors.Is(err, failure) {
		t.Fatalf("checkDocumentation() spelling error = %v", err)
	}
	runner = Runner{
		DocumentationSpelling: func(context.Context, string) error { return nil },
		DocumentationLinks:    func(context.Context, string) error { return failure },
	}
	if err := runner.checkDocumentation(context.Background(), root, rootModule); !errors.Is(err, failure) {
		t.Fatalf("checkDocumentation() links error = %v", err)
	}
}

func TestNestedDocumentationRunsExamplesWithoutRepositoryChecks(t *testing.T) {
	root := t.TempDir()
	moduleRoot := filepath.Join(root, "integration", "contracts")
	if err := os.MkdirAll(moduleRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	var commands []Command
	runner := Runner{
		Root: root,
		Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
			commands = append(commands, command)
			return nil
		}},
		DocumentationSpelling: func(context.Context, string) error {
			t.Fatal("nested module invoked repository spelling check")
			return nil
		},
		DocumentationLinks: func(context.Context, string) error {
			t.Fatal("nested module invoked repository link check")
			return nil
		},
	}
	module := inventory.Module{Directory: "integration/contracts"}
	if err := runner.checkDocumentation(context.Background(), moduleRoot, module); err != nil {
		t.Fatalf("checkDocumentation() error = %v", err)
	}
	if len(commands) != 1 || commands[0].Name != "go" || commands[0].Dir != moduleRoot ||
		strings.Join(commands[0].Args, " ") != "test ./... -run=^Example -count=1 -timeout=20m" {
		t.Fatalf("nested documentation commands = %#v", commands)
	}

	commands = nil
	runner.Policy.Operations = []config.Operation{{
		Module: module.Directory,
		Gate:   "docs",
		Steps:  []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestDocs$", Count: 1, Timeout: "1m"}},
	}}
	if err := runner.checkDocumentation(context.Background(), moduleRoot, module); err != nil {
		t.Fatalf("checkDocumentation(typed) error = %v", err)
	}
	if len(commands) != 1 || strings.Join(commands[0].Args, " ") != "test . -count=1 -timeout=1m -run=^TestDocs$" {
		t.Fatalf("typed nested documentation commands = %#v", commands)
	}

	failure := errors.New("injected example failure")
	runner.Policy.Operations = nil
	runner.Executor = workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return failure }}
	if err := runner.checkDocumentation(context.Background(), moduleRoot, module); !errors.Is(err, failure) ||
		!strings.Contains(err.Error(), "check documentation examples") {
		t.Fatalf("checkDocumentation(failure) error = %v", err)
	}
}

func TestCheckDocumentationUsesPinnedSpellingByDefault(t *testing.T) {
	root := spellingRepository(t)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{
		Executor:           workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return nil }},
		DocumentationLinks: func(context.Context, string) error { return nil },
	}
	if err := runner.checkDocumentation(context.Background(), root, inventory.Module{Directory: "."}); err != nil {
		t.Fatalf("checkDocumentation() error = %v", err)
	}
}

func TestDocumentationLinksUsesVerifiedTaskOwnedBinary(t *testing.T) {
	root := t.TempDir()
	workspace := t.TempDir()
	commands := make([]Command, 0, 2)
	executor := workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
		commands = append(commands, command)
		if command.Name == "curl" {
			for index, argument := range command.Args {
				if argument == "--output" {
					return os.WriteFile(command.Args[index+1], []byte("archive"), 0o600)
				}
			}
		}
		return nil
	}}
	release := docscheck.LycheeRelease{Target: "test", URL: "https://example.test/lychee.tar.gz", SHA256: strings.Repeat("a", 64)}
	runner := Runner{
		Executor: executor,
		documentationRelease: func(string, string) (docscheck.LycheeRelease, error) {
			return release, nil
		},
		documentationExtract: func(path string, got docscheck.LycheeRelease) ([]byte, error) {
			if path != filepath.Join(workspace, "documentation", "links", "lychee.tar.gz") || got != release {
				t.Fatalf("extract input = %q, %#v", path, got)
			}
			return []byte("binary"), nil
		},
	}
	if err := runner.runDocumentationLinks(context.Background(), root); err != nil {
		t.Fatalf("runDocumentationLinks() error = %v", err)
	}
	if len(commands) != 2 || commands[0].Name != "curl" || !strings.HasSuffix(commands[1].Name, filepath.Join("documentation", "links", "lychee")) {
		t.Fatalf("commands = %#v", commands)
	}
	info, err := os.Stat(commands[1].Name)
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("lychee binary = %v, %v", info, err)
	}
}

func TestCheckDocumentationUsesPinnedLinksByDefault(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	executor := linkWorkspaceExecutor(workspace, 0, nil)
	runner := Runner{
		Executor:              executor,
		DocumentationSpelling: func(context.Context, string) error { return nil },
		documentationExtract:  func(string, docscheck.LycheeRelease) ([]byte, error) { return []byte("binary"), nil },
	}
	if err := runner.checkDocumentation(context.Background(), root, inventory.Module{Directory: "."}); err != nil {
		t.Fatalf("checkDocumentation() error = %v", err)
	}
}

func TestDocumentationLinksReportsSetupAndToolFailures(t *testing.T) {
	failure := errors.New("injected failure")
	release := docscheck.LycheeRelease{Target: "test", URL: "https://example.test/lychee.tar.gz", SHA256: strings.Repeat("a", 64)}
	releaseFor := func(string, string) (docscheck.LycheeRelease, error) { //nolint:unparam // Matches the injected production contract.
		return release, nil
	}
	extract := func(string, docscheck.LycheeRelease) ([]byte, error) { return []byte("binary"), nil }

	t.Run("workspace", func(t *testing.T) {
		runner := Runner{Executor: executorFunction(func(context.Context, Command) error { return nil })}
		if err := runner.runDocumentationLinks(context.Background(), t.TempDir()); err == nil || !strings.Contains(err.Error(), "task-owned workspace") {
			t.Fatalf("runDocumentationLinks() error = %v", err)
		}
	})

	t.Run("release", func(t *testing.T) {
		runner := Runner{
			Executor:             workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return nil }},
			documentationRelease: func(string, string) (docscheck.LycheeRelease, error) { return docscheck.LycheeRelease{}, failure },
		}
		if err := runner.runDocumentationLinks(context.Background(), t.TempDir()); !errors.Is(err, failure) {
			t.Fatalf("runDocumentationLinks() error = %v", err)
		}
	})

	t.Run("tool root", func(t *testing.T) {
		workspace := filepath.Join(t.TempDir(), "workspace")
		if err := os.WriteFile(workspace, []byte("occupied"), 0o600); err != nil {
			t.Fatal(err)
		}
		runner := Runner{Executor: workspaceExecutor{directory: workspace, run: func(context.Context, Command) error { return nil }}, documentationRelease: releaseFor}
		if err := runner.runDocumentationLinks(context.Background(), t.TempDir()); err == nil || !strings.Contains(err.Error(), "create link tool root") {
			t.Fatalf("runDocumentationLinks() error = %v", err)
		}
	})

	tests := []struct {
		name    string
		failAt  int
		extract func(string, docscheck.LycheeRelease) ([]byte, error)
		prepare func(string)
		want    string
	}{
		{name: "download", failAt: 1, extract: extract, want: "download pinned link checker"},
		{name: "extract", extract: func(string, docscheck.LycheeRelease) ([]byte, error) { return nil, failure }, want: "injected failure"},
		{name: "default extract", want: "checksum mismatch"},
		{name: "write", extract: extract, prepare: func(workspace string) {
			if err := os.MkdirAll(filepath.Join(workspace, "documentation", "links", "lychee"), 0o700); err != nil {
				t.Fatal(err)
			}
		}, want: "write link checker"},
		{name: "checker", failAt: 2, extract: extract, want: "check documentation links"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			if test.prepare != nil {
				test.prepare(workspace)
			}
			runner := Runner{
				Executor:             linkWorkspaceExecutor(workspace, test.failAt, failure),
				documentationRelease: releaseFor,
				documentationExtract: test.extract,
			}
			if err := runner.runDocumentationLinks(context.Background(), t.TempDir()); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runDocumentationLinks() error = %v", err)
			}
		})
	}
}

func linkWorkspaceExecutor(workspace string, failAt int, failure error) workspaceExecutor {
	calls := 0
	return workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
		calls++
		if calls == failAt {
			return failure
		}
		if command.Name == "curl" {
			for index, argument := range command.Args {
				if argument == "--output" {
					return os.WriteFile(command.Args[index+1], []byte("archive"), 0o600)
				}
			}
		}
		return nil
	}}
}

func TestRunOperationReportsTimeoutCommandAndExecutionFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name     string
		step     config.Step
		executor Executor
		want     string
	}{
		{"timeout", config.Step{Type: "go-test", Timeout: "forever"}, executorFunction(func(context.Context, Command) error { return nil }), "timeout"},
		{"execution", config.Step{Type: "go-test", Packages: []string{"."}, Count: 1, Timeout: "1m"}, executorFunction(func(context.Context, Command) error { return failure }), "injected failure"},
		{"makefile", config.Step{Type: "make", Makefile: "missing.mk", Target: "verify", Timeout: "1m"}, executorFunction(func(context.Context, Command) error { return nil }), "read makefile"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Executor: test.executor}
			err := runner.runOperation(context.Background(), t.TempDir(), inventory.Module{Directory: "."}, config.Operation{Gate: "docs", Steps: []config.Step{test.step}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runOperation() error = %v", err)
			}
		})
	}
}

type executorFunction func(context.Context, Command) error

func (function executorFunction) Run(ctx context.Context, command Command) error {
	return function(ctx, command)
}

type fakeCoverageFiles struct {
	file      namedWriteCloser
	createErr error
	openErr   error
	directory string
}

func (fake *fakeCoverageFiles) CreateTemp(directory string) (namedWriteCloser, error) {
	fake.directory = directory
	return fake.file, fake.createErr
}

func (fake *fakeCoverageFiles) Open(string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewBufferString("mode: atomic\nexample/file.go:1.1,2.1 1 1\n")), fake.openErr
}

func (*fakeCoverageFiles) Remove(string) error { return nil }

type fakeSecretConfigFiles struct {
	file             namedWriteCloser
	createErr        error
	removeErr        error
	createdDirectory string
	pattern          string
	removed          string
}

type securityPolicyFiles struct {
	gitleaksRemoveErr error
	gitleaksRemoved   int
}

func (*securityPolicyFiles) CreateTemp(_ string, pattern string) (namedWriteCloser, error) {
	return &fakeNamedFile{name: pattern}, nil
}

func (files *securityPolicyFiles) Remove(path string) error {
	if strings.HasPrefix(path, "gitleaks-config-") {
		files.gitleaksRemoved++
		return files.gitleaksRemoveErr
	}
	return nil
}

func (files *fakeSecretConfigFiles) CreateTemp(directory, pattern string) (namedWriteCloser, error) {
	files.createdDirectory = directory
	files.pattern = pattern
	return files.file, files.createErr
}

func (files *fakeSecretConfigFiles) Remove(path string) error {
	files.removed = path
	return files.removeErr
}

type fakeNamedFile struct {
	name     string
	closeErr error
	writeErr error
	writeN   int
}

func (file *fakeNamedFile) Write(data []byte) (int, error) {
	if file.writeN > 0 {
		return file.writeN, file.writeErr
	}
	return len(data), file.writeErr
}
func (file *fakeNamedFile) Close() error { return file.closeErr }
func (file *fakeNamedFile) Name() string { return file.name }

func TestOperationCommandSupportsSelectorsAndRejectsUnknownTypes(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "verification"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "verification", "package.mk"), []byte("generated:\n\t@true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	module := inventory.Module{TestTags: []string{"integration"}}
	for name, test := range map[string]struct {
		step config.Step
		want string
	}{
		"benchmark": {
			step: config.Step{Type: "go-test", Packages: []string{"."}, Benchmark: ".", Budget: "250ms", Count: 1, Timeout: "1m"},
			want: "go test -tags=integration . -count=1 -timeout=1m -run=^$ -bench=. -benchmem -benchtime=250ms",
		},
		"fuzz": {
			step: config.Step{Type: "go-test", Packages: []string{"."}, Fuzz: "FuzzInput", Budget: "100x", Count: 1, Timeout: "1m"},
			want: "go test -tags=integration . -count=1 -timeout=1m -run=^$ -fuzz=FuzzInput -fuzztime=100x",
		},
		"make": {
			step: config.Step{Type: "make", Makefile: "verification/package.mk", Target: "generated", Timeout: "1m"},
			want: "make --no-print-directory -f - generated",
		},
	} {
		t.Run(name, func(t *testing.T) {
			command, err := operationCommand(root, module, test.step)
			if err != nil {
				t.Fatalf("operation command = %#v, %v", command, err)
			}
			if got := strings.Join(append([]string{command.Name}, command.Args...), " "); got != test.want {
				t.Fatalf("operation command = %q, want %q", got, test.want)
			}
			if name == "make" {
				content, readErr := io.ReadAll(command.Stdin)
				if readErr != nil || string(content) != "generated:\n\t@true\n" {
					t.Fatalf("makefile input = %q, %v", content, readErr)
				}
			}
		})
	}
	if _, err := operationCommand(root, module, config.Step{Type: "unsupported"}); err == nil {
		t.Fatal("operationCommand() error = nil")
	}
	command, err := operationCommand(root, inventory.Module{}, config.Step{Type: "go-test", Packages: []string{"."}, Count: 1, Timeout: "1m"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(command.Args, " "); strings.Contains(got, "-tags=") {
		t.Fatalf("operation command contains empty test tags: %q", got)
	}
}

func TestSafetyIgnoresTestOnlyUnsafeImports(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "unsafe_test.go"), []byte("package example\n\nimport \"unsafe\"\n\nvar _ = unsafe.Sizeof(0)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkSafety(root); err != nil {
		t.Fatalf("checkSafety() error = %v", err)
	}
}

func TestWalkModuleFilesSkipsVendorAndGitDirectories(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"vendor", ".git"} {
		path := filepath.Join(root, directory)
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "bad.go"), []byte("package ["), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := (Runner{}).checkFormatting(context.Background(), root); err != nil {
		t.Fatalf("checkFormatting() error = %v", err)
	}
}

func TestFormattingReportsUnreadableSource(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink("missing", filepath.Join(root, "missing.go")); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Executor: &processExecutor{stdout: io.Discard, stderr: io.Discard}}
	if err := runner.checkFormatting(context.Background(), root); err == nil {
		t.Fatal("checkFormatting() error = nil")
	}
}

func TestFormattingRejectsSourcePathsWithLineBreaks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows rejects control characters in file names")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad\nname.go"), []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := (Runner{}).checkFormatting(context.Background(), root); err == nil || !strings.Contains(err.Error(), "line break") {
		t.Fatalf("checkFormatting() error = %v", err)
	}
}

func TestWalkModuleFilesReportsNestedModuleInspectionFailure(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("go.mod", filepath.Join(nested, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if err := (Runner{}).checkFormatting(context.Background(), root); err == nil {
		t.Fatal("checkFormatting() error = nil")
	}
}
