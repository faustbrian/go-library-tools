package repository_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
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
}

func TestParityWorkflowNormalizesReplacedDependencyChecksums(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	for _, required := range []string{
		"s#golib\\\\/#go-#g",
		"grep -Fq 'faustbrian\\/go-' .golib/scripts/internal/isolated-go.sh",
		"s/GOWORK=off go mod download/GOWORK=off GOFLAGS= go mod download/",
		"s/GOWORK=off go build/GOWORK=off GOFLAGS= go build/",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("parity workflow lacks %q", required)
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
		"gh release create",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("release workflow lacks %q", required)
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
