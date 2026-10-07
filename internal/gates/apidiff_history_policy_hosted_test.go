//go:build scannerintegration

package gates

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoricalAPIDiffPolicyPinnedScannerHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("pinned historical tool-identity verification runs only in hosted CI")
	}
	executor, cleanup, err := NewProcessExecutor(t.TempDir(), io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	runner := Runner{Executor: executor}
	configPath, cleanupConfig, err := runner.createGitleaksConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanupConfig(); err != nil {
			t.Error(err)
		}
	})
	ruleDigest := sha256.Sum256([]byte("generic-api-key"))
	ruleHash := hex.EncodeToString(ruleDigest[:])
	version := strings.Join([]string{"v0.0.0", "20260713181234", "a1b2c3d4e5f6"}, "-")
	for _, test := range []struct {
		name, path, content string
		findings            int
	}{
		{"root tool identity", "Makefile", "APIDIFF_VERSION := " + version + "\n", 0},
		{"legacy tool identity", ".golib/package.mk", "APIDIFF_VERSION := " + version + " \t\n", 0},
		{"root default tool identity", "Makefile", "APIDIFF_VERSION ?= " + version + "\n", 0},
		{"legacy default tool identity", ".golib/package.mk", "APIDIFF_VERSION ?= " + version + " \t\n", 0},
		{"other assignment", "Makefile", "ACCESS_TOKEN := " + version + "\n", 1},
		{"legacy other assignment", ".golib/package.mk", "ACCESS_TOKEN := " + version + "\n", 1},
		{"default other assignment", "Makefile", "ACCESS_TOKEN ?= " + version + "\n", 1},
		{"default other path", "nested/Makefile", "APIDIFF_VERSION ?= " + version + "\n", 1},
		{"default mixed assignments", "Makefile", "APIDIFF_VERSION ?= " + version + "\nACCESS_TOKEN := " + version + "\n", 1},
		{"other path", "nested/Makefile", "APIDIFF_VERSION := " + version + "\n", 1},
		{"mixed assignments", "Makefile", "APIDIFF_VERSION := " + version + "\nACCESS_TOKEN := " + version + "\n", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, test.path)
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(test.content), 0o600); err != nil {
				t.Fatal(err)
			}
			for _, arguments := range [][]string{
				{"init", "-q"}, {"add", "--", test.path},
				{"-c", "user.name=golib-test", "-c", "user.email=golib-test@example.invalid", "commit", "-qm", "public tool identity fixture"},
			} {
				// #nosec G204 -- executable and argument arrays are fixed task-owned fixture Git operations
				command := exec.CommandContext(t.Context(), "git", arguments...)
				command.Dir = root
				if err := command.Run(); err != nil {
					t.Fatal("initialize public identity fixture history")
				}
			}
			for _, mode := range []string{"git", "dir"} {
				arguments := []string{mode, ".", "--config", configPath, "--ignore-gitleaks-allow", "--no-banner", "--redact"}
				if mode == "git" {
					arguments = append(arguments, "--log-opts=--all")
				}
				err := runner.securityTool(t.Context(), io.Discard, ".", "secrets-"+mode, root,
					"github.com/zricethezav/gitleaks/v8@"+gitleaksVersion, arguments...)
				if test.findings == 0 {
					if err != nil {
						t.Fatal("exact public tool identity remains a scanner finding or scanner failed")
					}
				} else if err == nil || !strings.Contains(err.Error(), "secret-findings") || strings.Count(err.Error(), "secret-location ") != test.findings || strings.Count(err.Error(), " "+ruleHash+" ") != test.findings {
					t.Fatal("non-allowlisted assignment did not retain the exact expected finding count")
				}
				if test.findings != 0 {
					continue
				}
				// Default-rule counterfactual proves these public identity fixtures
				// really trigger generic scanning without the centrally owned policy.
				counterfactual := filepath.Join(t.TempDir(), "default-gitleaks.toml")
				if err := os.WriteFile(counterfactual, []byte("[extend]\nuseDefault = true\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				arguments[3] = counterfactual
				err = runner.securityTool(t.Context(), io.Discard, ".", "secrets-"+mode, root,
					"github.com/zricethezav/gitleaks/v8@"+gitleaksVersion, arguments...)
				if err == nil || !strings.Contains(err.Error(), "secret-findings") || strings.Count(err.Error(), "secret-location ") != 1 || strings.Count(err.Error(), " "+ruleHash+" ") != 1 {
					t.Fatal("default-rule counterfactual did not detect the public identity fixture")
				}
			}
		})
	}
}
