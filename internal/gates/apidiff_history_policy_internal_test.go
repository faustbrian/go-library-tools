package gates

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestHistoricalAPIDiffPolicyRequiresAssignmentIdentity(t *testing.T) {
	// Public Go tool identity, not a credential. Split the immutable version
	// components so the policy's own test source is not a scanner fixture.
	version := strings.Join([]string{"v0.0.0", "20260713181234", "a1b2c3d4e5f6"}, "-")
	for _, test := range []struct {
		name, path, rule, line string
		allowed                bool
	}{
		{"legacy version file", ".golib/versions.env", "generic-api-key", "APIDIFF_VERSION=" + version, true},
		{"legacy package makefile", ".golib/package.mk", "generic-api-key", "APIDIFF_VERSION := " + version, true},
		{"root makefile", "Makefile", "generic-api-key", "APIDIFF_VERSION := " + version, true},
		{"legacy default assignment", ".golib/package.mk", "generic-api-key", "APIDIFF_VERSION ?= " + version, true},
		{"root default assignment", "Makefile", "generic-api-key", "APIDIFF_VERSION ?= " + version, true},
		{"scanner line framing", "Makefile", "generic-api-key", "\nAPIDIFF_VERSION := " + version, true},
		{"other assignment", "Makefile", "generic-api-key", "ACCESS_TOKEN := " + version, false},
		{"legacy makefile other assignment", ".golib/package.mk", "generic-api-key", "ACCESS_TOKEN := " + version, false},
		{"other operator", "Makefile", "generic-api-key", "APIDIFF_VERSION += " + version, false},
		{"default other assignment", "Makefile", "generic-api-key", "ACCESS_TOKEN ?= " + version, false},
		{"default other path", "nested/Makefile", "generic-api-key", "APIDIFF_VERSION ?= " + version, false},
		{"default other rule", "Makefile", "another-rule", "APIDIFF_VERSION ?= " + version, false},
		{"default other value", "Makefile", "generic-api-key", "APIDIFF_VERSION ?= ordinary-value", false},
		{"default same line extra assignment", "Makefile", "generic-api-key", "APIDIFF_VERSION ?= " + version + "; ACCESS_TOKEN := " + version, false},
		{"trailing whitespace", "Makefile", "generic-api-key", "APIDIFF_VERSION := " + version + " \t", true},
		{"other path", "nested/Makefile", "generic-api-key", "APIDIFF_VERSION := " + version, false},
		{"other rule", "Makefile", "another-rule", "APIDIFF_VERSION := " + version, false},
		{"other value", "Makefile", "generic-api-key", "APIDIFF_VERSION := ordinary-value", false},
		{"prefix", "Makefile", "generic-api-key", "OTHER_APIDIFF_VERSION := " + version, false},
		{"same line extra assignment", "Makefile", "generic-api-key", "APIDIFF_VERSION := " + version + "; ACCESS_TOKEN := " + version, false},
		{"comment", "Makefile", "generic-api-key", "# APIDIFF_VERSION := " + version, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if allowed := historicalAPIDiffPolicyAllows(t, test.path, test.rule, test.line, version); allowed != test.allowed {
				t.Fatalf("historical tool-identity allowance = %v, want %v", allowed, test.allowed)
			}
		})
	}
}

// This inert test evaluates the rendered policy's path/content expressions at
// the documented Gitleaks match-target boundary. Hosted tests separately prove
// actual scanner configuration parsing and finding retention in both modes.
func historicalAPIDiffPolicyAllows(t *testing.T, path, rule, line, secret string) bool {
	t.Helper()
	var blocks []string
	for candidate := range strings.SplitSeq(gitleaksPolicy, "[[allowlists]]") {
		if strings.Contains(candidate, `description = "Historical APIDIFF_VERSION`) {
			blocks = append(blocks, candidate)
		}
	}
	if len(blocks) == 0 {
		t.Fatal("historical tool-identity policy is missing")
	}
	if rule != "generic-api-key" {
		return false
	}
	patterns := func(block, key string) []string {
		t.Helper()
		array := regexp.MustCompile(`(?s)` + key + ` = \[(.*?)\]\n`).FindStringSubmatch(block)
		if len(array) != 2 {
			t.Fatal("historical policy expression array is missing")
		}
		quoted := regexp.MustCompile(`'''(.*?)'''`).FindAllStringSubmatch(array[1], -1)
		result := make([]string, 0, len(quoted))
		for _, value := range quoted {
			result = append(result, value[1])
		}
		if len(result) == 0 {
			t.Fatal("historical policy expression array is empty")
		}
		return result
	}
	for _, block := range blocks {
		if !strings.Contains(block, `condition = "AND"`) || !strings.Contains(block, `targetRules = ["generic-api-key"]`) {
			t.Fatal("historical tool-identity policy is missing its rule/path conjunction")
		}
		content := secret
		if strings.Contains(block, `regexTarget = "line"`) {
			content = line
		}
		if slices.ContainsFunc(patterns(block, "paths"), func(pattern string) bool { return regexp.MustCompile(pattern).MatchString(path) }) &&
			slices.ContainsFunc(patterns(block, "regexes"), func(pattern string) bool { return regexp.MustCompile(pattern).MatchString(content) }) {
			return true
		}
	}
	return false
}
