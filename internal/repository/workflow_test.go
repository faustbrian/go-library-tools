package repository_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
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
		"golib check --module",
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

func TestToolingWorkflowUploadsVerificationEvidenceOnEveryOutcome(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/ci.yml")
	start := strings.Index(content, "- name: Upload verification evidence")
	if start < 0 {
		t.Fatal("tooling workflow does not upload verification evidence")
	}
	remainder := content[start:]
	step, _, found := strings.Cut(remainder, "\n  codeql:")
	if !found {
		t.Fatal("tooling evidence upload is outside the quality job")
	}
	for _, required := range []string{
		"if: always()",
		"uses: actions/upload-artifact@",
		"path: .verification",
		"include-hidden-files: true",
	} {
		if !strings.Contains(step, required) {
			t.Errorf("tooling evidence upload lacks %q", required)
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
