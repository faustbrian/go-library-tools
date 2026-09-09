package repository_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"go.yaml.in/yaml/v3"
)

var remoteAction = regexp.MustCompile(`(?m)^\s*- uses: ([^./][^@\s]+)@([^\s#]+)`)
var immutableAction = regexp.MustCompile(`^[0-9a-f]{40}$`)

func TestRepositoryWorkflowsPinRemoteActions(t *testing.T) {
	root := projectRoot(t)
	entries, err := os.ReadDir(filepath.Join(root, ".github", "workflows"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		path := filepath.Join(root, ".github", "workflows", entry.Name())
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		matches := remoteAction.FindAllStringSubmatch(string(data), -1)
		if len(matches) == 0 {
			t.Fatalf("%s has no pinned remote actions", entry.Name())
		}
		for _, match := range matches {
			if !immutableAction.MatchString(match[2]) {
				t.Errorf("%s action %s uses mutable ref %s", entry.Name(), match[1], match[2])
			}
		}
	}
}

func TestReusableWorkflowPreservesConsumerContract(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/library-ci.yml")
	for _, required := range []string{
		"workflow_call:",
		"tooling_sha:",
		"security-events: write",
		"golib repository check",
		"golib workflows check",
		"golib specification check --online",
		"golib check --local --module",
		"github/codeql-action/init@",
		"github/codeql-action/analyze@",
		"name: Required",
		"if: always()",
		".verification",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("reusable workflow lacks %q", required)
		}
	}
	if strings.Contains(content, "curl |") || strings.Contains(content, "@main") {
		t.Fatal("reusable workflow contains an unsafe bootstrap reference")
	}
	if strings.Contains(content, "packages: read") {
		t.Fatal("reusable workflow requests package access that consumer callers do not grant")
	}
}

func TestReusableWorkflowBuildsReleaseModuleMatrix(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	input, exists := workflow.On.WorkflowCall.Inputs["release_module"]
	if !exists || input.Type != "string" || input.Default != "" {
		t.Fatalf("release_module workflow input = %#v, present = %v", input, exists)
	}
	var matrixScript string
	for _, step := range workflow.Jobs["prepare"].Steps {
		if step.ID == "modules" {
			if step.Env["RELEASE_MODULE"] != "${{ inputs.release_module }}" || step.Env["RELEASE_DRY_RUN"] != "${{ inputs.release_dry_run }}" {
				t.Fatalf("module-matrix environment = %#v", step.Env)
			}
			matrixScript = step.Run
			break
		}
	}
	if matrixScript == "" {
		t.Fatal("reusable workflow has no executable module-matrix step")
	}

	bin := t.TempDir()
	golib := filepath.Join(bin, "golib")
	stub := "#!/bin/sh\ncase \"$*\" in\n  'inventory --json') printf '%s\\n' \"$INVENTORY\" ;;\n  'config show --json') printf '%s\\n' '{\"runtimes\":{}}' ;;\n  *) exit 64 ;;\nesac\n"
	if err := os.WriteFile(golib, []byte(stub), 0o700); err != nil {
		t.Fatal(err)
	}
	inventory := `{"modules":[{"directory":".","releasable":true},{"directory":"nested","releasable":true},{"directory":"fixture","releasable":false}]}`
	tests := []struct {
		name      string
		dryRun    string
		module    string
		want      []string
		wantError string
	}{
		{name: "default", dryRun: "false", want: []string{".", "nested", "fixture"}},
		{name: "all releasable", dryRun: "true", want: []string{".", "nested"}},
		{name: "selected releasable", dryRun: "true", module: "nested", want: []string{"nested"}},
		{name: "unknown selection", dryRun: "true", module: "missing", wantError: "release module is unknown or not releasable: missing"},
		{name: "non-releasable selection", dryRun: "true", module: "fixture", wantError: "release module is unknown or not releasable: fixture"},
		{name: "selection outside rehearsal", dryRun: "false", module: "nested", wantError: "release_module requires release_dry_run: true"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "github-output")
			command := exec.CommandContext(t.Context(), "bash", "-c", matrixScript)
			command.Env = append(os.Environ(),
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVENTORY="+inventory,
				"GITHUB_OUTPUT="+output,
				"RELEASE_DRY_RUN="+test.dryRun,
				"RELEASE_MODULE="+test.module,
			)
			combined, err := command.CombinedOutput()
			if test.wantError != "" {
				if err == nil || !strings.Contains(string(combined), test.wantError) {
					t.Fatalf("matrix step error = %v, output = %q", err, combined)
				}
				return
			}
			if err != nil {
				t.Fatalf("matrix step error = %v, output = %q", err, combined)
			}
			contents, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			line, _, _ := strings.Cut(string(contents), "\n")
			encoded, found := strings.CutPrefix(line, "matrix=")
			if !found {
				t.Fatalf("matrix output = %q", contents)
			}
			var entries []struct {
				Directory string `json:"directory"`
			}
			if err := json.Unmarshal([]byte(encoded), &entries); err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(entries))
			for _, entry := range entries {
				got = append(got, entry.Directory)
			}
			if strings.Join(got, ",") != strings.Join(test.want, ",") {
				t.Fatalf("matrix directories = %v, want %v", got, test.want)
			}
		})
	}
}

func TestReleaseModuleSelectorFlowsThroughHostedReleasePaths(t *testing.T) {
	var reusable workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &reusable); err != nil {
		t.Fatal(err)
	}
	assertReleaseStep := func(job, name, expectedIf, expectedInput, expectedCommand string) string {
		t.Helper()
		for _, step := range reusable.Jobs[job].Steps {
			if step.Name == name {
				if step.If != expectedIf || step.Env["RELEASE_MODULE"] != expectedInput || !strings.Contains(step.Run, expectedCommand) {
					t.Fatalf("%s/%s does not preserve release routing: if=%q env=%#v run=%q", job, name, step.If, step.Env, step.Run)
				}
				return step.Run
			}
		}
		t.Fatalf("%s has no %q step", job, name)
		return ""
	}
	releaseCheckScript := assertReleaseStep("repository-contract", "Validate release contract", "inputs.release_dry_run == true", "${{ inputs.release_module }}", `golib release check --module "${RELEASE_MODULE}"`)
	qualityRehearsalScript := assertReleaseStep("quality", "Run release rehearsal", "inputs.release_dry_run == true", "${{ matrix.directory }}", `golib release dry-run --module "${RELEASE_MODULE}"`)
	ordinaryContract := false
	for _, step := range reusable.Jobs["quality"].Steps {
		if step.Name == "Run module contract" {
			ordinaryContract = true
			if step.If != "inputs.release_dry_run != true" {
				t.Fatalf("ordinary module contract guard = %q", step.If)
			}
			if step.Run != "golib check --local --module '${{ matrix.directory }}'" {
				t.Fatalf("ordinary module contract command = %q", step.Run)
			}
		}
	}
	if !ordinaryContract {
		t.Fatal("quality job has no ordinary module contract step")
	}

	golibBin := t.TempDir()
	golibInvocations := filepath.Join(t.TempDir(), "golib-invocations")
	golibStub := "#!/bin/sh\nprintf '%s\\n' \"$*\" >>\"$INVOCATIONS\"\n"
	if err := os.WriteFile(filepath.Join(golibBin, "golib"), []byte(golibStub), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		module string
		want   string
	}{
		{name: "blank", want: "release check"},
		{name: "selected", module: "nested", want: "release check --module nested"},
	} {
		t.Run("reusable release check "+test.name, func(t *testing.T) {
			if err := os.WriteFile(golibInvocations, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", releaseCheckScript)
			command.Env = append(os.Environ(),
				"PATH="+golibBin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVOCATIONS="+golibInvocations,
				"RELEASE_MODULE="+test.module,
			)
			if combined, err := command.CombinedOutput(); err != nil {
				t.Fatalf("release check error = %v, output = %q", err, combined)
			}
			contents, err := os.ReadFile(golibInvocations)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(contents)) != test.want {
				t.Fatalf("golib invocation = %q, want %q", contents, test.want)
			}
		})
	}
	if err := os.WriteFile(golibInvocations, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	qualityCommand := exec.CommandContext(t.Context(), "bash", "-c", qualityRehearsalScript)
	qualityCommand.Env = append(os.Environ(),
		"PATH="+golibBin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"INVOCATIONS="+golibInvocations,
		"RELEASE_MODULE=nested",
	)
	if combined, err := qualityCommand.CombinedOutput(); err != nil {
		t.Fatalf("quality release rehearsal error = %v, output = %q", err, combined)
	}
	qualityInvocation, err := os.ReadFile(golibInvocations)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(qualityInvocation)) != "release dry-run --module nested" {
		t.Fatalf("quality golib invocation = %q", qualityInvocation)
	}

	var caller workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/ci.yml")), &caller); err != nil {
		t.Fatal(err)
	}
	input, exists := caller.On.WorkflowDispatch.Inputs["release_module"]
	if !exists || input.Type != "string" || input.Default != "" {
		t.Fatalf("tooling release_module dispatch input = %#v, present = %v", input, exists)
	}
	rehearsal := caller.Jobs["release-rehearsal"]
	if rehearsal.If != "github.event_name == 'workflow_dispatch' && (inputs.release_rehearsal || inputs.release_module != '')" {
		t.Fatalf("tooling release rehearsal condition = %q", rehearsal.If)
	}
	var validationScript, rehearsalScript string
	for _, step := range rehearsal.Steps {
		switch step.Name {
		case "Validate release inputs":
			if step.Env["RELEASE_MODULE"] != "${{ inputs.release_module }}" || step.Env["RELEASE_REHEARSAL"] != "${{ inputs.release_rehearsal }}" {
				t.Fatalf("tooling input validation environment = %#v", step.Env)
			}
			validationScript = step.Run
		case "Run pre-tag release dry-run":
			if step.Env["RELEASE_MODULE"] != "${{ inputs.release_module }}" || !strings.Contains(step.Run, `release dry-run --module "${RELEASE_MODULE}"`) {
				t.Fatalf("tooling rehearsal does not propagate the release selector: env=%#v run=%q", step.Env, step.Run)
			}
			rehearsalScript = step.Run
		}
	}
	if validationScript == "" || rehearsalScript == "" {
		t.Fatalf("tooling workflow release scripts missing: validation=%v rehearsal=%v", validationScript != "", rehearsalScript != "")
	}

	for _, test := range []struct {
		name      string
		rehearsal string
		module    string
		wantError string
	}{
		{name: "blank", rehearsal: "true"},
		{name: "selected", rehearsal: "true", module: "nested"},
		{name: "selector without rehearsal", rehearsal: "false", module: "nested", wantError: "release_module requires release_rehearsal: true"},
	} {
		t.Run("input "+test.name, func(t *testing.T) {
			command := exec.CommandContext(t.Context(), "bash", "-c", validationScript)
			command.Env = append(os.Environ(), "RELEASE_REHEARSAL="+test.rehearsal, "RELEASE_MODULE="+test.module)
			combined, err := command.CombinedOutput()
			if test.wantError == "" && err != nil {
				t.Fatalf("input validation error = %v, output = %q", err, combined)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(string(combined), test.wantError)) {
				t.Fatalf("input validation error = %v, output = %q", err, combined)
			}
		})
	}

	bin := t.TempDir()
	invocations := filepath.Join(t.TempDir(), "go-invocations")
	stub := "#!/bin/sh\nprintf '%s\\n' \"$*\" >>\"$INVOCATIONS\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(stub), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		module string
		want   string
	}{
		{name: "blank", want: "run ./cmd/golib release dry-run"},
		{name: "selected", module: "nested", want: "run ./cmd/golib release dry-run --module nested"},
	} {
		t.Run("rehearsal "+test.name, func(t *testing.T) {
			if err := os.WriteFile(invocations, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", rehearsalScript)
			command.Env = append(os.Environ(),
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVOCATIONS="+invocations,
				"RELEASE_MODULE="+test.module,
				"RUNNER_TEMP="+t.TempDir(),
			)
			if combined, err := command.CombinedOutput(); err != nil {
				t.Fatalf("release rehearsal error = %v, output = %q", err, combined)
			}
			contents, err := os.ReadFile(invocations)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(contents)) != test.want {
				t.Fatalf("go invocation = %q, want %q", contents, test.want)
			}
		})
	}
}

type workflowDocument struct {
	On struct {
		WorkflowCall struct {
			Inputs map[string]workflowInput `yaml:"inputs"`
		} `yaml:"workflow_call"`
		WorkflowDispatch struct {
			Inputs map[string]workflowInput `yaml:"inputs"`
		} `yaml:"workflow_dispatch"`
	} `yaml:"on"`
	Jobs map[string]workflowJob `yaml:"jobs"`
}

type workflowInput struct {
	Type    string `yaml:"type"`
	Default any    `yaml:"default"`
}

type workflowJob struct {
	If    string         `yaml:"if"`
	Steps []workflowStep `yaml:"steps"`
}

type workflowStep struct {
	ID   string            `yaml:"id"`
	Name string            `yaml:"name"`
	If   string            `yaml:"if"`
	Env  map[string]string `yaml:"env"`
	Run  string            `yaml:"run"`
}

func TestReusableWorkflowConfiguresBootstrapProxyForEveryGoBuild(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/library-ci.yml")
	for _, required := range []string{
		"uses: ./.golib-tooling/.github/actions/setup-bootstrap-proxy",
		"bootstrap_url: ${{ vars.GOLIB_BOOTSTRAP_PROXY_URL }}",
		"bootstrap_sha256: ${{ vars.GOLIB_BOOTSTRAP_PROXY_SHA256 }}",
	} {
		if count := strings.Count(content, required); count != 2 {
			t.Errorf("reusable workflow %q count = %d, want quality and CodeQL", required, count)
		}
	}
}

func TestBootstrapProxyActionVerifiesArchiveBeforeExport(t *testing.T) {
	content := readProjectFile(t, ".github/actions/setup-bootstrap-proxy/action.yml")
	checksum := strings.Index(content, "sha256sum --check")
	extraction := strings.Index(content, "tar --extract")
	export := strings.Index(content, "GOPROXY=https://proxy.golang.org,file://")
	if checksum < 0 || extraction < 0 || export < 0 || checksum > extraction || extraction > export {
		t.Fatal("bootstrap proxy action must verify before extraction and export")
	}
	for _, required := range []string{"bootstrap_url:", "bootstrap_sha256:", "GONOSUMDB=github.com/faustbrian/go-*"} {
		if !strings.Contains(content, required) {
			t.Errorf("bootstrap proxy action lacks %q", required)
		}
	}
}

func TestToolingWorkflowSeparatesFastPullRequestAndAggregateMilestoneChecks(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/ci.yml")
	qualityStart := strings.Index(content, "  quality:\n")
	milestoneStart := strings.Index(content, "  milestone:\n")
	codeQLStart := strings.Index(content, "  codeql:\n")
	if qualityStart < 0 || milestoneStart < 0 || codeQLStart < 0 || qualityStart >= milestoneStart || milestoneStart >= codeQLStart {
		t.Fatal("tooling workflow does not define ordered quality and milestone jobs")
	}
	quality := content[qualityStart:milestoneStart]
	milestone := content[milestoneStart:codeQLStart]
	for _, required := range []string{"make local-ci", "Run mutation verifier behavior when affected", "-tags=verifierintegration"} {
		if !strings.Contains(quality, required) {
			t.Errorf("pull-request quality lacks %q", required)
		}
	}
	for _, forbidden := range []string{"make check", "golib check --all", "Upload verification evidence"} {
		if strings.Contains(quality, forbidden) {
			t.Errorf("pull-request quality includes aggregate work %q", forbidden)
		}
	}
	for _, required := range []string{"if: github.event_name != 'pull_request'", "needs: quality", "make milestone-check"} {
		if !strings.Contains(milestone, required) {
			t.Errorf("milestone job lacks %q", required)
		}
	}
	if strings.Contains(milestone, "make local-ci") {
		t.Fatal("milestone job repeats the successful local contract")
	}
	makefile := readProjectFile(t, "Makefile")
	if !strings.Contains(makefile, "local-ci:\n\t$(call run_go,run ./cmd/golib check --local)") ||
		!strings.Contains(makefile, "milestone-check: consumers compatibility") {
		t.Fatal("Makefile does not keep bounded local and aggregate checks separate")
	}
	for _, retired := range []string{"forward-oracle", "tools/provenance", "Upload verification evidence"} {
		if strings.Contains(content, retired) {
			t.Errorf("tooling workflow retains retired routine evidence machinery %q", retired)
		}
	}
}

func TestParityWorkflowDoesNotModifyRepresentativeSource(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	for _, forbidden := range []string{
		"Normalize legacy analyzer compatibility",
		"sed -i 's#golib",
		"sed -i 's/^STATICCHECK_VERSION=",
		".golib/scripts/build-golib-gremlins.sh",
	} {
		if strings.Contains(content, forbidden) {
			t.Errorf("parity workflow modifies representative source through %q", forbidden)
		}
	}
}

func TestSharedParityUsesRepresentativeGoVersionForConsumerGates(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	sharedStart := strings.Index(content, "  shared:\n")
	if sharedStart < 0 {
		t.Fatal("parity workflow has no shared job")
	}
	sharedRemainder := content[sharedStart:]
	shared, _, found := strings.Cut(sharedRemainder, "\n  performance:\n")
	if !found {
		t.Fatal("parity workflow has no performance job after shared parity")
	}
	if count := strings.Count(shared, "uses: actions/setup-go@"); count != 1 {
		t.Fatalf("shared parity setup-go steps = %d, want 1", count)
	}
	setup := strings.Index(shared, "go-version-file: source/.go-version")
	build := strings.Index(shared, "name: Build source CLI")
	if setup < 0 || build < 0 || setup > build {
		t.Fatal("shared parity must select the representative Go version before building the source CLI")
	}
	if !strings.Contains(shared[build:], "GOTOOLCHAIN=auto") {
		t.Fatal("source CLI build does not opt into its required automatic Go toolchain")
	}
}

func TestParityWorkflowUsesActionsPathChannelForGoWrapper(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	if count := strings.Count(content, `"${GITHUB_ENV}" "${GITHUB_PATH}"`); count != 3 {
		t.Fatalf("parity wrapper path exports = %d, want legacy, shared, and performance", count)
	}
}

func TestPerformanceRehearsalPublishesComparableRawMeasurements(t *testing.T) {
	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	for _, required := range []string{
		"name: Performance / ${{ matrix.name }}",
		"rehearsals/performance.sh",
		"performance-${{ matrix.artifact }}",
		"rehearsals/performance-compare.sh",
		"performance-results.json",
		"performance-services.status",
		"performance_source_run_id",
		"github-token: ${{ github.token }}",
		"Resolve measurement source",
		`/repos/${GITHUB_REPOSITORY}/actions/runs/${REQUESTED_RUN_ID}`,
		"run-id: ${{ steps.source.outputs.run_id }}",
		"rehearsals/repositories.json",
		`["core", "service"]`,
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("performance workflow lacks %q", required)
		}
	}

	harness := readProjectFile(t, "rehearsals/performance.sh")
	for _, required := range []string{
		"startup-diagnostic",
		"repository-inventory",
		"checkpoint-reuse",
		"module-scaling-sequential",
		"module-scaling-concurrent",
		"peak_rss_kib",
		"artifact_size_bytes",
		"isolated_cache_residue",
		"service-lifecycle",
		"mutation_package_count",
		"reused content-identical mutation evidence",
		"tooling_revision",
		"golib_sha256",
	} {
		if !strings.Contains(harness, required) {
			t.Errorf("performance harness lacks %q", required)
		}
	}

	documentation := readProjectFile(t, "docs/performance.md")
	for _, required := range []string{
		"content-identical",
		"Raw Results",
		"Runner variance",
		"No-op and checkpoint reuse",
		"Concurrent module scaling",
	} {
		if !strings.Contains(documentation, required) {
			t.Errorf("performance documentation lacks %q", required)
		}
	}
}

func TestParityWorkflowCanReuseCompletedContractArtifacts(t *testing.T) {
	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	for _, required := range []string{
		"compatibility_source_run_id",
		"Resolve compatibility source",
		`/repos/${GITHUB_REPOSITORY}/actions/runs/${REQUESTED_RUN_ID}`,
		`/repos/${GITHUB_REPOSITORY}/actions/runs/${REQUESTED_RUN_ID}/jobs?per_page=100`,
		`git fetch --no-tags --depth=1 origin "${source_sha}"`,
		`git diff --name-only "${source_sha}" "${GITHUB_SHA}" --`,
		`grep -Ev '(^rehearsals/compare\.sh$|_test\.go$)'`,
		"run-id: ${{ steps.contract-source.outputs.run_id }}",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("compatibility workflow lacks %q", required)
		}
	}
	if count := strings.Count(workflow, "if: inputs.compatibility_source_run_id == ''"); count != 2 {
		t.Fatalf("compatibility artifact-reuse guards = %d, want legacy and shared", count)
	}
}

func TestParityRehearsalReusesRepresentativeMutationEvidence(t *testing.T) {
	t.Parallel()

	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	sharedStart := strings.Index(workflow, "  shared:\n")
	performanceStart := strings.Index(workflow, "  performance:\n")
	performanceReportStart := strings.Index(workflow, "  performance-report:\n")
	if sharedStart < 0 || performanceStart < 0 || performanceReportStart < 0 ||
		sharedStart >= performanceStart || performanceStart >= performanceReportStart {
		t.Fatal("parity workflow job order is invalid")
	}
	shared := workflow[sharedStart:performanceStart]
	performance := workflow[performanceStart:performanceReportStart]
	for _, required := range []string{
		"Restore mutation checkpoints",
		"restore-ci-mutation-evidence.sh",
		"if: inputs.performance_source_run_id == ''",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("parity workflow lacks %q", required)
		}
	}
	for name, job := range map[string]string{"shared": shared, "performance": performance} {
		for _, required := range []string{
			"cp .golib/mutation-bootstrap/*.zip",
			"cp .golib/mutation-history-migrations.json",
		} {
			if !strings.Contains(job, required) {
				t.Errorf("%s parity job lacks %q", name, required)
			}
		}
	}

	wantImports := map[string]int{
		"go-authorization":        1,
		"go-cloudevents":          2,
		"go-knapsack":             2,
		"go-openapi":              1,
		"go-transactional-outbox": 5,
	}
	for repository, want := range wantImports {
		configuration, err := config.Load(filepath.Join(projectRoot(t), "rehearsals", repository))
		if err != nil {
			t.Errorf("load %s rehearsal configuration: %v", repository, err)
			continue
		}
		if got := len(configuration.Mutation.Imports); got != want {
			t.Errorf("%s mutation imports = %d, want %d", repository, got, want)
		}
	}
}

func TestPerformanceRehearsalExcludesInjectedPolicyFromSourceChecks(t *testing.T) {
	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	performanceStart := strings.Index(workflow, "  performance:\n")
	if performanceStart < 0 {
		t.Fatal("parity workflow has no performance job")
	}
	performance, _, found := strings.Cut(workflow[performanceStart:], "\n  performance-report:\n")
	if !found {
		t.Fatal("parity workflow has no performance report after performance jobs")
	}
	for _, required := range []string{
		`printf '%s\n' '/.golib.yaml' '/.verification/'`,
		`>>.git/info/exclude`,
	} {
		if !strings.Contains(performance, required) {
			t.Errorf("performance policy setup lacks %q", required)
		}
	}
}

func TestRehearsalFuzzTargetsAreExact(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join(projectRoot(t), "rehearsals", "go-*", ".golib.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		content, readErr := os.ReadFile(fixture)
		if readErr != nil {
			t.Fatal(readErr)
		}
		for line := range strings.SplitSeq(string(content), "\n") {
			selector := strings.TrimSpace(line)
			if !strings.HasPrefix(selector, "fuzz:") {
				continue
			}
			value := strings.TrimSpace(strings.TrimPrefix(selector, "fuzz:"))
			if !strings.HasPrefix(value, "'^Fuzz") || !strings.HasSuffix(value, "$'") {
				t.Errorf("%s has non-exact fuzz selector %q", filepath.Base(filepath.Dir(fixture)), value)
			}
		}
	}
}

func TestSetupActionVerifiesReleasedArtifactBeforeExtraction(t *testing.T) {
	content := readProjectFile(t, ".github/actions/setup-golib/action.yml")
	checksumSet := strings.Contains(content, "tool_checksums_sha256:")
	checksumSetVerification := strings.Index(content, "checksums.txt\" | sha256sum --check")
	checksum := strings.LastIndex(content, "sha256sum --check")
	attestation := strings.Index(content, "gh attestation verify")
	extraction := strings.Index(content, "tar --extract")
	if !checksumSet || checksumSetVerification < 0 || checksumSetVerification > checksum ||
		checksum < 0 || attestation < 0 || extraction < 0 || checksum > extraction || attestation > extraction {
		t.Fatal("setup action must verify checksum and provenance before extraction")
	}
	for _, platform := range []string{"linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64"} {
		if !strings.Contains(content, platform) {
			t.Errorf("setup action lacks %s", platform)
		}
	}
}

func TestReleaseWorkflowBuildsAndAttestsEverySupportedPlatform(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/release.yml")
	for _, required := range []string{
		"{goos: linux, goarch: amd64}",
		"{goos: linux, goarch: arm64}",
		"{goos: darwin, goarch: amd64}",
		"{goos: darwin, goarch: arm64}",
		"checksums.txt",
		"release-manifest.json",
		"anchore/sbom-action@",
		"actions/attest@",
		"actions: read",
		"actions/workflows/ci.yml/runs",
		"gh release create",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("release workflow lacks %q", required)
		}
	}
	if strings.Contains(content, `[[ "${declared}" == "${GITHUB_REF_NAME}" ]]`) {
		t.Fatal("release workflow requires the unpublished release to bootstrap itself")
	}
	if strings.Contains(content, "go run ./cmd/golib check --all") {
		t.Fatal("release workflow repeats the exact-commit repository contract")
	}
}

func TestReleaseWorkflowPublishesVerifiedCohesionCatalogs(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/release.yml")
	for _, required := range []string{
		"  prepare-catalog:\n",
		"  project-catalog:\n",
		"  assemble-catalog:\n",
		"  verify-projection:\n",
		"  verify-catalog:\n",
		"  prepare-publication:\n",
		"  verify-publication:\n",
		"  attest-publication:\n",
		"max-parallel: 8",
		"persist-credentials: false",
		"submodules: false",
		"lfs: false",
		"gh attestation verify",
		"golib cohesion sources check",
		"golib cohesion sources verify",
		"golib cohesion catalog engineering --json",
		"golib cohesion aggregate generate",
		"golib cohesion aggregate check",
		"cmp --silent",
		"cohesion-sources.json",
		"cohesion-inputs.json",
		"cohesion-projections.tar.gz",
		"catalog-consumer.json",
		"catalog-consumer.md",
		"catalog-engineering.json",
		"catalog-engineering.md",
		"gzip -n",
		"needs: [build, verify-catalog]",
		"needs: prepare-publication",
		"needs: verify-publication",
		"needs: attest-publication",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("release workflow lacks catalog publication contract %q", required)
		}
	}
	verifyProjectionStart := strings.Index(content, "  verify-projection:\n")
	verifyCatalogStart := strings.Index(content, "  verify-catalog:\n")
	if verifyProjectionStart < 0 || verifyCatalogStart < 0 || verifyProjectionStart >= verifyCatalogStart {
		t.Fatal("release workflow must independently verify projections before catalog verification")
	}
	verifyProjection := content[verifyProjectionStart:verifyCatalogStart]
	for _, required := range []string{
		"needs: [prepare-catalog, assemble-catalog]",
		"matrix: ${{ fromJSON(needs.prepare-catalog.outputs.matrix) }}",
		"ref: ${{ matrix.commit }}",
		"golib cohesion sources verify",
		"golib cohesion catalog engineering --json",
		"cmp --silent",
		"cohesion-projections.tar.gz",
	} {
		if !strings.Contains(verifyProjection, required) {
			t.Errorf("projection verifier lacks %q", required)
		}
	}
	verifyCatalog, _, found := strings.Cut(content[verifyCatalogStart:], "\n  prepare-publication:\n")
	if !found || !strings.Contains(verifyCatalog, "needs: [prepare-catalog, assemble-catalog, verify-projection]") {
		t.Fatal("catalog verification must wait for independent projection verification")
	}
	for _, required := range []string{
		"expected-members",
		"source_repository=",
		"input_repository=",
		"input_projection=",
		"sha256sum \"${projection}\"",
	} {
		if !strings.Contains(verifyCatalog, required) {
			t.Errorf("catalog verifier lacks source-lock binding %q", required)
		}
	}
	prepareStart := strings.Index(content, "  prepare-publication:\n")
	verifyStart := strings.Index(content, "  verify-publication:\n")
	attestStart := strings.Index(content, "  attest-publication:\n")
	publishStart := strings.Index(content, "  publish:\n")
	if prepareStart < 0 || verifyStart < 0 || attestStart < 0 || publishStart < 0 ||
		prepareStart >= verifyStart || verifyStart >= attestStart || attestStart >= publishStart {
		t.Fatal("release workflow must prepare, verify, attest, then publish immutable assets")
	}
	prepare := content[prepareStart:verifyStart]
	for _, required := range []string{
		"release-manifest.json",
		"checksums.txt",
		"name: release-publication",
	} {
		if !strings.Contains(prepare, required) {
			t.Errorf("publication preparation lacks %q", required)
		}
	}
	verify := content[verifyStart:attestStart]
	for _, required := range []string{
		"needs: prepare-publication",
		"name: release-publication",
		"sha256sum --check checksums.txt",
		"release-manifest.json",
		"checksums.txt",
	} {
		if !strings.Contains(verify, required) {
			t.Errorf("publication verification lacks %q", required)
		}
	}
	attest := content[attestStart:publishStart]
	if !strings.Contains(attest, "needs: verify-publication") ||
		!strings.Contains(attest, "actions/attest@") ||
		!strings.Contains(attest, "name: release-publication") {
		t.Error("attestation must consume the independently verified publication set")
	}
	publish := content[publishStart:]
	for _, required := range []string{
		"needs: attest-publication",
		"name: release-publication",
		"gh attestation verify",
		"gh release create",
	} {
		if !strings.Contains(publish, required) {
			t.Errorf("static publisher lacks %q", required)
		}
	}
	_, releaseCreate, found := strings.Cut(publish, "gh release create")
	if !found {
		t.Fatal("static publisher lacks release creation command")
	}
	if !strings.Contains(releaseCreate, `--repo "${GITHUB_REPOSITORY}"`) {
		t.Error("static publisher must bind release creation to the workflow repository")
	}
	for _, forbidden := range []string{
		"golib cohesion",
		"go run",
		"make ",
		"--clobber",
		">dist/release-manifest.json",
		">dist/checksums.txt",
		"actions/attest@",
	} {
		if strings.Contains(publish, forbidden) {
			t.Errorf("write-capable publish job contains forbidden execution %q", forbidden)
		}
	}
}

func TestConsumerUpgradeWorkflowIsBoundedAndReviewable(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/update-consumers.yml")
	for _, required := range []string{
		"workflow_dispatch:",
		"version:",
		"workflow_sha:",
		"checksums_sha256:",
		"cohort:",
		"dry_run:",
		"golib consumers validate --json",
		"cohort must contain 1-10 repositories",
		"git fetch --no-tags --depth=1 origin",
		"gh release download",
		"actual_checksums_sha256",
		"ref: ${{ inputs.workflow_sha }}",
		"max-parallel: 5",
		"if: inputs.dry_run == false",
		"GOLIB_ROLLOUT_TOKEN",
		"git add .golib.yaml .github/workflows/ci.yml",
		"git ls-files --others --exclude-standard",
		"gh pr create",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("consumer upgrade workflow lacks %q", required)
		}
	}
	for _, forbidden := range []string{"--force", "git add --all", "git add -A", "@main", "cancel-in-progress: true"} {
		if strings.Contains(content, forbidden) {
			t.Errorf("consumer upgrade workflow contains forbidden %q", forbidden)
		}
	}
	if regexp.MustCompile(`(?m)^\s*git add \.\s*$`).MatchString(content) {
		t.Error("consumer upgrade workflow contains wholesale staging")
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate repository test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
}

func readProjectFile(t *testing.T, relative string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectRoot(t), filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
