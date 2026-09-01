package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/buildinfo"
	"github.com/faustbrian/go-library-tools/internal/cohesion"
	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/gates"
)

func TestExecuteCohesionCoversOutputAndFailureContracts(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	legacyRoot := internalFixture(t)
	legacyPolicy, err := config.Load(legacyRoot)
	if err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{{}, {"catalog"}, {"catalog", "unknown"}, {"catalog", "consumer", "--yaml"}, {"unknown"}} {
		var stdout, stderr bytes.Buffer
		if code := executeCohesion(args, root, policy, &stdout, &stderr); code != 2 {
			t.Fatalf("executeCohesion(%v) = %d, %q", args, code, stderr.String())
		}
	}
	var catalogUsage bytes.Buffer
	if code := executeCohesion([]string{"catalog"}, root, policy, &bytes.Buffer{}, &catalogUsage); code != 2 || !strings.HasPrefix(catalogUsage.String(), "usage: golib cohesion catalog ") {
		t.Fatalf("executeCohesion(catalog) = %d, %q", code, catalogUsage.String())
	}
	for _, args := range [][]string{{"check"}, {"check", "--json"}, {"catalog", "engineering"}} {
		var stdout, stderr bytes.Buffer
		if code := executeCohesion(args, root, policy, &stdout, &stderr); code != 0 || stderr.Len() != 0 || stdout.Len() == 0 {
			t.Fatalf("executeCohesion(%v) = %d, %q, %q", args, code, stdout.String(), stderr.String())
		}
	}
	var invalidStdout, invalidStderr bytes.Buffer
	if code := executeCohesion([]string{"check"}, legacyRoot, legacyPolicy, &invalidStdout, &invalidStderr); code != 1 || invalidStdout.Len() == 0 || invalidStderr.Len() != 0 {
		t.Fatalf("executeCohesion(legacy check) = %d, %q, %q", code, invalidStdout.String(), invalidStderr.String())
	}
	var catalogStdout, catalogStderr bytes.Buffer
	if code := executeCohesion([]string{"catalog", "engineering"}, legacyRoot, legacyPolicy, &catalogStdout, &catalogStderr); code != 1 || catalogStderr.Len() == 0 {
		t.Fatalf("executeCohesion(legacy catalog) = %d, %q", code, catalogStderr.String())
	}

	for _, args := range [][]string{{"check", "--json"}, {"catalog", "engineering", "--json"}, {"catalog", "engineering"}} {
		var stderr bytes.Buffer
		if code := executeCohesion(args, root, policy, errorWriter{}, &stderr); code != 1 || !strings.Contains(stderr.String(), "write cohesion") {
			t.Fatalf("executeCohesion(writer %v) = %d, %q", args, code, stderr.String())
		}
	}

	originalDigest := buildinfo.DesignLanguageSHA256
	buildinfo.DesignLanguageSHA256 = "invalid"
	t.Cleanup(func() { buildinfo.DesignLanguageSHA256 = originalDigest })
	var stdout, stderr bytes.Buffer
	if code := executeCohesion([]string{"catalog", "engineering"}, root, policy, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "design-language identity") {
		t.Fatalf("executeCohesion(invalid identity) = %d, %q", code, stderr.String())
	}
	buildinfo.DesignLanguageSHA256 = originalDigest

	originalVersion := buildinfo.Version
	originalSource := buildinfo.DesignLanguageSourceIdentity
	buildinfo.Version = "v1.2.3"
	buildinfo.DesignLanguageSourceIdentity = "v1.2.3"
	t.Cleanup(func() {
		buildinfo.Version = originalVersion
		buildinfo.DesignLanguageSourceIdentity = originalSource
	})
	stdout.Reset()
	stderr.Reset()
	if code := executeCohesion([]string{"catalog", "engineering", "--json"}, root, policy, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), `"publication_status": "published"`) {
		t.Fatalf("executeCohesion(published) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	buildinfo.Version = originalVersion
	buildinfo.DesignLanguageSourceIdentity = originalSource

	originalRenderer := renderCohesionCatalog
	renderCohesionCatalog = func(cohesion.Envelope) ([]byte, error) { return nil, errors.New("render failure") }
	t.Cleanup(func() { renderCohesionCatalog = originalRenderer })
	stdout.Reset()
	stderr.Reset()
	if code := executeCohesion([]string{"catalog", "engineering"}, root, policy, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "render failure") {
		t.Fatalf("executeCohesion(render failure) = %d, %q", code, stderr.String())
	}
}

func TestExecuteCohesionRejectsSourceBuildPublishingAggregateCatalogs(t *testing.T) {
	root := internalFixture(t)
	policy, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := executeCohesion([]string{
		"aggregate", "generate",
		"--inputs", filepath.Join(root, "inputs.json"),
		"--output", filepath.Join(root, "docs", "ecosystem"),
	}, root, policy, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "source build cannot publish cohesion catalogs") {
		t.Fatalf("executeCohesion(aggregate generate) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "ecosystem")); !os.IsNotExist(err) {
		t.Fatalf("public output exists after rejection: %v", err)
	}

	publicOutput := filepath.Join(root, "docs", "ecosystem")
	if err := os.MkdirAll(publicOutput, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "catalog-alias")
	if err := os.Symlink(publicOutput, alias); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = executeCohesion([]string{
		"aggregate", "generate",
		"--inputs", filepath.Join(root, "inputs.json"),
		"--output", alias,
	}, root, policy, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "source build cannot publish cohesion catalogs") {
		t.Fatalf("executeCohesion(aggregate generate through symlink) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteCohesionAllowsSourceBuildToCheckPublishedAggregateCatalogs(t *testing.T) {
	root := t.TempDir()
	originalVersion := buildinfo.Version
	buildinfo.Version = "dev"
	t.Cleanup(func() { buildinfo.Version = originalVersion })
	var stdout, stderr bytes.Buffer
	code := executeCohesion([]string{
		"aggregate", "check",
		"--inputs", filepath.Join(root, "missing-inputs.json"),
		"--output", filepath.Join(root, "docs", "ecosystem"),
	}, root, config.Config{}, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "read cohesion aggregation inputs") || strings.Contains(stderr.String(), "source build cannot publish") {
		t.Fatalf("executeCohesion(aggregate check) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteCohesionRoutesSourceBuildGenerationThroughProtectedOutputBoundary(t *testing.T) {
	root := t.TempDir()
	originalVersion := buildinfo.Version
	buildinfo.Version = "dev"
	t.Cleanup(func() { buildinfo.Version = originalVersion })
	var stdout, stderr bytes.Buffer
	code := executeCohesion([]string{
		"aggregate", "generate",
		"--inputs", filepath.Join(root, "missing-inputs.json"),
		"--output", filepath.Join(root, "preview"),
	}, root, config.Config{}, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "read cohesion aggregation inputs") {
		t.Fatalf("executeCohesion(source aggregate generate) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteCohesionGeneratesAndChecksAggregateCatalogs(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	originalVersion := buildinfo.Version
	originalSource := buildinfo.DesignLanguageSourceIdentity
	buildinfo.Version = "v1.3.0"
	buildinfo.DesignLanguageSourceIdentity = "v1.3.0"
	t.Cleanup(func() {
		buildinfo.Version = originalVersion
		buildinfo.DesignLanguageSourceIdentity = originalSource
	})

	var projection, projectionStderr bytes.Buffer
	if code := executeCohesion([]string{"catalog", "engineering", "--json"}, root, policy, &projection, &projectionStderr); code != 0 {
		t.Fatalf("executeCohesion(catalog) = %d, %q", code, projectionStderr.String())
	}
	var envelope struct {
		Repository string `json:"repository"`
	}
	if err := json.Unmarshal(projection.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	task := t.TempDir()
	projectionPath := filepath.Join(task, "repository.json")
	if err := os.WriteFile(projectionPath, projection.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(projection.Bytes())
	inputsPath := filepath.Join(task, "inputs.json")
	inputs := `{"schema_version":1,"design_language":{"version":"` + buildinfo.DesignLanguageVersion + `","sha256":"` + buildinfo.DesignLanguageSHA256 + `"},"repositories":[{"repository":"` + envelope.Repository + `","projection":"repository.json","sha256":"` + hex.EncodeToString(digest[:]) + `"}]}`
	if err := os.WriteFile(inputsPath, []byte(inputs), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(task, "preview")

	for _, action := range []string{"generate", "check"} {
		var stdout, stderr bytes.Buffer
		code := executeCohesion([]string{"aggregate", action, "--inputs", inputsPath, "--output", output}, root, policy, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 || stdout.String() != "cohesion aggregate "+action+" passed\n" {
			t.Fatalf("executeCohesion(aggregate %s) = %d, %q, %q", action, code, stdout.String(), stderr.String())
		}
	}
}

func TestCohesionAggregateArgumentsAndPathBoundaries(t *testing.T) {
	for name, args := range map[string][]string{
		"length":       {"generate"},
		"action":       {"publish", "--inputs", "a", "--output", "b"},
		"flag":         {"generate", "--unknown", "a", "--output", "b"},
		"empty":        {"generate", "--inputs", "", "--output", "b"},
		"empty output": {"generate", "--inputs", "a", "--output", ""},
		"duplicate":    {"generate", "--inputs", "a", "--inputs", "b"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, _, err := cohesionAggregateArguments(args); err == nil {
				t.Fatal("cohesionAggregateArguments() error = nil")
			}
		})
	}
	action, inputs, output, err := cohesionAggregateArguments([]string{"check", "--output", "out", "--inputs", "in"})
	if err != nil || action != "check" || inputs != "in" || output != "out" {
		t.Fatalf("cohesionAggregateArguments() = %q, %q, %q, %v", action, inputs, output, err)
	}
	if got := resolveCohesionPath("/repo", "relative"); got != "/repo/relative" {
		t.Fatalf("resolveCohesionPath(relative) = %q", got)
	}
	if got := resolveCohesionPath("/repo", "/absolute/../catalog"); got != "/catalog" {
		t.Fatalf("resolveCohesionPath(absolute) = %q", got)
	}
}

func TestCanonicalCohesionPathReportsEveryResolutionBoundary(t *testing.T) {
	failure := errors.New("injected failure")
	if _, err := canonicalCohesionPathWithFunctions("value", func(string) (string, error) {
		return "", failure
	}, filepath.EvalSymlinks); !errors.Is(err, failure) {
		t.Fatalf("canonicalCohesionPathWithFunctions(abs) error = %v", err)
	}
	if _, err := canonicalCohesionPathWithFunctions("value", func(string) (string, error) {
		return "/value", nil
	}, func(string) (string, error) {
		return "", failure
	}); !errors.Is(err, failure) {
		t.Fatalf("canonicalCohesionPathWithFunctions(eval) error = %v", err)
	}
	missing := &os.PathError{Op: "eval", Path: "/", Err: os.ErrNotExist}
	if _, err := canonicalCohesionPathWithFunctions("value", func(string) (string, error) {
		return "/", nil
	}, func(string) (string, error) {
		return "", missing
	}); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canonicalCohesionPathWithFunctions(root) error = %v", err)
	}
	resolved, err := canonicalCohesionPathWithFunctions("value", func(string) (string, error) {
		return "/root/a/b", nil
	}, func(value string) (string, error) {
		if value == "/root" {
			return "/resolved", nil
		}
		return "", missing
	})
	if err != nil || resolved != "/resolved/a/b" {
		t.Fatalf("canonicalCohesionPathWithFunctions(missing suffix) = %q, %v", resolved, err)
	}

	calls := 0
	canonicalize := func(string) (string, error) {
		calls++
		if calls == 1 {
			return "", failure
		}
		return "/value", nil
	}
	if _, err := sameCohesionPathWithCanonicalizer("left", "right", canonicalize); !errors.Is(err, failure) {
		t.Fatalf("sameCohesionPathWithCanonicalizer(left) error = %v", err)
	}
	calls = 0
	canonicalize = func(string) (string, error) {
		calls++
		if calls == 2 {
			return "", failure
		}
		return "/value", nil
	}
	if _, err := sameCohesionPathWithCanonicalizer("left", "right", canonicalize); !errors.Is(err, failure) {
		t.Fatalf("sameCohesionPathWithCanonicalizer(right) error = %v", err)
	}
}

func TestExecuteCohesionRejectsUnresolvablePublicationPath(t *testing.T) {
	root := internalFixture(t)
	policy, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	originalVersion := buildinfo.Version
	buildinfo.Version = "dev"
	t.Cleanup(func() { buildinfo.Version = originalVersion })
	var stdout, stderr bytes.Buffer
	code := executeCohesion([]string{
		"aggregate", "generate", "--inputs", "inputs.json", "--output", "\x00",
	}, root, policy, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "resolve cohesion publication path") {
		t.Fatalf("executeCohesion(unresolvable path) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteCohesionReportsAggregateUsageAndGenerationFailures(t *testing.T) {
	root := internalFixture(t)
	policy, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := executeCohesion([]string{"aggregate"}, root, policy, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage: golib cohesion aggregate") {
		t.Fatalf("executeCohesion(aggregate usage) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	originalVersion := buildinfo.Version
	originalSource := buildinfo.DesignLanguageSourceIdentity
	buildinfo.Version = "v1.3.0"
	buildinfo.DesignLanguageSourceIdentity = "v1.3.0"
	t.Cleanup(func() {
		buildinfo.Version = originalVersion
		buildinfo.DesignLanguageSourceIdentity = originalSource
	})
	stdout.Reset()
	stderr.Reset()
	code := executeCohesion([]string{
		"aggregate", "generate", "--inputs", "missing-inputs.json", "--output", "preview",
	}, root, policy, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "read cohesion aggregation inputs") {
		t.Fatalf("executeCohesion(aggregate failure) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("write failure") }

func TestExecuteReportsExecutorCreationAndCleanupFailures(t *testing.T) {
	root := internalFixture(t)
	failure := errors.New("injected failure")
	tests := []struct {
		name    string
		factory executorFactory
	}{
		{"creation", func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
			return nil, nil, failure
		}},
		{"cleanup", func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
			return successfulExecutor{}, func() error { return failure }, nil
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := execute([]string{"check"}, root, &stdout, &stderr, test.factory)
			if code != 1 || !strings.Contains(stderr.String(), failure.Error()) {
				t.Fatalf("execute() code = %d, stderr = %q", code, stderr.String())
			}
		})
	}
}

func TestExecuteContextPropagatesCancellationToCommands(t *testing.T) {
	root := internalFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return cliExecutorFunction(func(commandContext context.Context, _ gates.Command) error {
			return commandContext.Err()
		}), func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := executeContext(ctx, []string{"check"}, root, &stdout, &stderr, factory); code != 1 || !strings.Contains(stderr.String(), context.Canceled.Error()) {
		t.Fatalf("executeContext() = %d, %q", code, stderr.String())
	}
}

func TestExecuteRoutesWorkflowChecks(t *testing.T) {
	root := internalFixture(t)
	var command gates.Command
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return cliExecutorFunction(func(_ context.Context, value gates.Command) error {
			command = value
			return nil
		}), func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"workflows", "check"}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute(workflows check) = %d, %q", code, stderr.String())
	}
	rootInfo, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	commandInfo, err := os.Stat(command.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if command.Name != "go" || !os.SameFile(rootInfo, commandInfo) || !strings.Contains(strings.Join(command.Args, " "), "actionlint@v1.7.12") {
		t.Fatalf("workflow command = %#v", command)
	}
	for _, args := range [][]string{{"workflows"}, {"workflows", "lint"}, {"workflows", "check", "extra"}} {
		stderr.Reset()
		if code := execute(args, root, &stdout, &stderr, factory); code != 2 || !strings.Contains(stderr.String(), "usage: golib workflows") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteRoutesAPICheckAndUpdate(t *testing.T) {
	root := internalFixture(t)
	if err := os.Mkdir(filepath.Join(root, "api"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "api", "baseline.txt"), []byte("baseline"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"api_compatibility":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return apiExecutor{}, func() error { return nil }, nil
	}
	tests := []struct {
		args       []string
		wantOutput string
		wantBase   string
	}{
		{args: []string{"api", "check"}, wantOutput: "API compatibility passed", wantBase: "baseline"},
		{args: []string{"api", "update", "--module", "."}, wantOutput: "API baseline updated", wantBase: "snapshot"},
	}
	for _, test := range tests {
		if err := os.WriteFile(filepath.Join(root, "api", "baseline.txt"), []byte("baseline"), 0o600); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		if code := execute(test.args, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), test.wantOutput) {
			t.Fatalf("execute(%v) = %d, %q, %q", test.args, code, stdout.String(), stderr.String())
		}
		baseline, err := os.ReadFile(filepath.Join(root, "api", "baseline.txt"))
		if err != nil || string(baseline) != test.wantBase {
			t.Fatalf("execute(%v) baseline = %q, %v", test.args, baseline, err)
		}
	}
}

func TestExecuteRejectsInvalidAPIArguments(t *testing.T) {
	for _, args := range [][]string{{"api"}, {"api", "remove"}, {"api", "check", "--all", "extra"}} {
		var stdout, stderr bytes.Buffer
		if code := execute(args, internalFixture(t), &stdout, &stderr, nil); code != 2 || !strings.Contains(stderr.String(), "usage: golib api") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteRoutesStandaloneCoverage(t *testing.T) {
	root := internalFixture(t)
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"coverage":true},"packages":[{"module_directory":".","directory":".","name":"example","import_path":"example","kind":"public","production":true,"executable":true,"coverage_required":true,"build_required":true,"build_tags":[]}]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	packages := `{"schema_version":1,"repository":"example","packages":[{"module_directory":".","directory":".","name":"example","import_path":"example","kind":"public","production":true,"executable":true,"coverage_required":true,"build_required":true,"build_tags":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "packages.json"), []byte(packages), 0o600); err != nil {
		t.Fatal(err)
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return coverageExecutor{}, func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"coverage", "--module", "."}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute() = %d, %q", code, stderr.String())
	}
	if code := execute([]string{"coverage", "--bad"}, root, &stdout, &stderr, factory); code != 2 {
		t.Fatalf("execute() invalid coverage = %d", code)
	}
}

func TestExecuteRoutesStandaloneMutation(t *testing.T) {
	root := internalFixture(t)
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return coverageExecutor{}, func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"mutation"}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute() default mutation = %d, %q", code, stderr.String())
	}
	if code := execute([]string{"mutation", "--module", "."}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute() mutation = %d, %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "not applicable") {
		t.Fatalf("mutation output = %q", stdout.String())
	}
	if code := execute([]string{"mutation", "--bad"}, root, &stdout, &stderr, factory); code != 2 {
		t.Fatalf("execute() invalid mutation = %d", code)
	}
	if code := execute([]string{"mutation", "import"}, root, &stdout, &stderr, factory); code != 2 {
		t.Fatalf("execute() invalid mutation import = %d", code)
	}
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"mutation":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	stderr.Reset()
	if code := execute([]string{"mutation", "import", "--module", ".", "--archive", "legacy.zip", "--ledger", "ledger.json"}, root, &stdout, &stderr, factory); code != 1 || !strings.Contains(stderr.String(), "checkpoint archive") {
		t.Fatalf("execute() mutation import = %d, %q", code, stderr.String())
	}
}

func TestMutationImportArguments(t *testing.T) {
	options, err := mutationImportArguments([]string{"--ledger", "ledger.json", "--module", ".", "--archive", "legacy.zip"})
	if err != nil || options.module != "." || options.archive != "legacy.zip" || options.ledger != "ledger.json" {
		t.Fatalf("mutationImportArguments() = %#v, %v", options, err)
	}
	for _, args := range [][]string{
		nil,
		{"--module", ".", "--archive", "legacy.zip", "--unknown", "ledger.json"},
		{"--module", ".", "--module", ".", "--ledger", "ledger.json"},
		{"--module", ".", "--archive", "", "--ledger", "ledger.json"},
		{"--module", ".", "--archive", "legacy.zip", "--archive", "other.zip"},
	} {
		if _, err := mutationImportArguments(args); err == nil {
			t.Fatalf("mutationImportArguments(%v) error = nil", args)
		}
	}
}

func TestExecuteRoutesDocumentationCheck(t *testing.T) {
	root := internalFixture(t)
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return cliWorkspaceExecutor{directory: t.TempDir()}, func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"docs", "check"}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "docs: not applicable") {
		t.Fatalf("execute() = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	for _, args := range [][]string{{"docs"}, {"docs", "update"}, {"docs", "check", "--bad"}} {
		stderr.Reset()
		if code := execute(args, root, &stdout, &stderr, factory); code != 2 || !strings.Contains(stderr.String(), "usage: golib docs") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteRoutesReleaseChecks(t *testing.T) {
	root := internalFixture(t)
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"version":"1.0.0","tag_prefix":"v","gates":{"coverage":true,"documentation":true,"lint":true,"mutation":true,"race":true,"security":true,"tests":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"release", "check"}, root, &stdout, &stderr, nil); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "release contract passed") {
		t.Fatalf("execute(release check) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return successfulExecutor{}, func() error { return nil }, nil
	}
	stderr.Reset()
	if code := execute([]string{"release", "dry-run"}, root, &stdout, &stderr, factory); code != 1 || stderr.Len() == 0 {
		t.Fatalf("execute(release dry-run) = %d, %q", code, stderr.String())
	}
	for _, args := range [][]string{{"release"}, {"release", "publish"}, {"release", "check", "extra"}} {
		stderr.Reset()
		if code := execute(args, root, &stdout, &stderr, nil); code != 2 || !strings.Contains(stderr.String(), "usage: golib release") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteReportsReleaseContractFailures(t *testing.T) {
	root := internalFixture(t)
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"release", "check"}, root, &stdout, &stderr, nil); code != 1 || !strings.Contains(stderr.String(), "stable version") {
		t.Fatalf("execute(release invalid) = %d, %q", code, stderr.String())
	}
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"version":"1.0.0","tag_prefix":"v","gates":{"coverage":true,"documentation":true,"lint":true,"mutation":true,"race":true,"security":true,"tests":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".go-version"), []byte("1.26.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stderr.Reset()
	if code := execute([]string{"release", "check"}, root, &stdout, &stderr, nil); code != 1 || !strings.Contains(stderr.String(), ".go-version") {
		t.Fatalf("execute(release repository) = %d, %q", code, stderr.String())
	}
}

type successfulExecutor struct{}

func (successfulExecutor) Run(context.Context, gates.Command) error { return nil }

type cliExecutorFunction func(context.Context, gates.Command) error

func (function cliExecutorFunction) Run(ctx context.Context, command gates.Command) error {
	return function(ctx, command)
}

type cliWorkspaceExecutor struct {
	directory string
}

func (cliWorkspaceExecutor) Run(context.Context, gates.Command) error { return nil }
func (executor cliWorkspaceExecutor) TemporaryDirectory() string      { return executor.directory }

type apiExecutor struct{}

func (apiExecutor) Run(_ context.Context, command gates.Command) error {
	for index, argument := range command.Args {
		if argument == "-w" {
			return os.WriteFile(command.Args[index+1], []byte("snapshot"), 0o600)
		}
	}
	return nil
}

type coverageExecutor struct{}

func (coverageExecutor) Run(_ context.Context, command gates.Command) error {
	for _, argument := range command.Args {
		if profile, found := strings.CutPrefix(argument, "-coverprofile="); found {
			return os.WriteFile(profile, []byte("mode: atomic\nexample/file.go:1.1,2.1 1 1\n"), 0o600)
		}
	}
	return nil
}

func internalFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		".golib.yaml":   "schema_version: 1\ntool_version: v1.0.0\n",
		".go-version":   "1.27.0\n",
		"go.mod":        "module example\n\ngo 1.27.0\n",
		"example.go":    "package example\n",
		"modules.json":  `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`,
		"packages.json": `{"schema_version":1,"repository":"example","packages":[]}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
