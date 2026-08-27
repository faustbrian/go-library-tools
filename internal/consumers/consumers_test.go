package consumers_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/consumers"
)

func TestParseAndSummarizeConsumerInventory(t *testing.T) {
	manifest, err := consumers.Parse(encode(t, validManifest()))
	if err != nil {
		t.Fatal(err)
	}
	summary := manifest.Summary()
	if summary.SchemaVersion != 1 || summary.Owner != "faustbrian" || summary.Total != 3 || summary.Active != 1 || summary.Deferred != 1 || summary.Tooling != 1 {
		t.Fatalf("Summary() = %#v", summary)
	}
	summary.Repositories[0].Name = "changed"
	if manifest.Repositories[0].Name == "changed" {
		t.Fatal("Summary() aliases manifest repositories")
	}
}

func TestLoadReadsRepositoryRelativeConsumerInventory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "consumers.json"), encode(t, validManifest()), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := consumers.Load(root, "consumers.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := consumers.Load(root, "missing.json"); err == nil {
		t.Fatal("Load() missing error = nil")
	}
}

func TestParseRejectsInvalidConsumerInventory(t *testing.T) {
	tests := map[string]func(*manifestFixture){
		"schema":             func(value *manifestFixture) { value.SchemaVersion = 2 },
		"owner":              func(value *manifestFixture) { value.Owner = "-invalid" },
		"empty":              func(value *manifestFixture) { value.Repositories = nil },
		"too many":           func(value *manifestFixture) { value.Repositories = repeatedRepositories(257) },
		"name":               func(value *manifestFixture) { value.Repositories[0].Name = "invalid" },
		"unsorted":           func(value *manifestFixture) { value.Repositories[1].Name = "go-a" },
		"duplicate":          func(value *manifestFixture) { value.Repositories[1].Name = value.Repositories[0].Name },
		"branch empty":       func(value *manifestFixture) { value.Repositories[0].DefaultBranch = "" },
		"branch dot start":   func(value *manifestFixture) { value.Repositories[0].DefaultBranch = ".main" },
		"branch dot end":     func(value *manifestFixture) { value.Repositories[0].DefaultBranch = "main." },
		"branch slash start": func(value *manifestFixture) { value.Repositories[0].DefaultBranch = "/main" },
		"branch slash end":   func(value *manifestFixture) { value.Repositories[0].DefaultBranch = "main/" },
		"branch dots":        func(value *manifestFixture) { value.Repositories[0].DefaultBranch = "main..next" },
		"branch reflog":      func(value *manifestFixture) { value.Repositories[0].DefaultBranch = "main@{1}" },
		"branch forbidden":   func(value *manifestFixture) { value.Repositories[0].DefaultBranch = "main branch" },
		"classification":     func(value *manifestFixture) { value.Repositories[0].Classification = "unknown" },
		"active reason":      func(value *manifestFixture) { value.Repositories[0].Reason = "not allowed" },
		"deferred reason":    func(value *manifestFixture) { value.Repositories[1].Reason = " " },
		"tooling reason":     func(value *manifestFixture) { value.Repositories[2].Reason = "" },
		"no active": func(value *manifestFixture) {
			value.Repositories[0].Classification = "deferred"
			value.Repositories[0].Reason = "deferred"
		},
		"no tooling": func(value *manifestFixture) {
			value.Repositories[2].Classification = "deferred"
		},
		"two tooling": func(value *manifestFixture) {
			value.Repositories[1].Classification = "tooling"
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := validManifest()
			mutate(&value)
			if _, err := consumers.Parse(encode(t, value)); err == nil {
				t.Fatal("Parse() error = nil")
			}
		})
	}
}

func TestParseRejectsMalformedBoundaries(t *testing.T) {
	for name, data := range map[string][]byte{
		"malformed": []byte("{"),
		"unknown":   []byte(`{"schema_version":1,"owner":"faustbrian","repositories":[],"unknown":true}`),
		"multiple":  append(encode(t, validManifest()), []byte("{}")...),
		"oversized": []byte(strings.Repeat(" ", 1<<20+1)),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := consumers.Parse(data); err == nil {
				t.Fatal("Parse() error = nil")
			}
		})
	}
}

func TestParseAcceptsExactManifestBoundaries(t *testing.T) {
	value := validManifest()
	value.Repositories = boundaryRepositories(256)
	data := encode(t, value)
	data = append(data, []byte(strings.Repeat(" ", 1<<20-len(data)))...)
	manifest, err := consumers.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Repositories) != 256 {
		t.Fatalf("repository count = %d", len(manifest.Repositories))
	}
}

func TestParseRejectsRepositoryCountsOutsideBoundaries(t *testing.T) {
	for _, repositories := range [][]repositoryFixture{nil, boundaryRepositories(257)} {
		value := validManifest()
		value.Repositories = repositories
		_, err := consumers.Parse(encode(t, value))
		if err == nil || !strings.Contains(err.Error(), "between 1 and 256") {
			t.Fatalf("Parse() error = %v", err)
		}
	}
}

type manifestFixture struct {
	SchemaVersion int                 `json:"schema_version"`
	Owner         string              `json:"owner"`
	Repositories  []repositoryFixture `json:"repositories"`
}

type repositoryFixture struct {
	Name           string `json:"name"`
	Classification string `json:"classification"`
	DefaultBranch  string `json:"default_branch"`
	Reason         string `json:"reason,omitempty"`
}

func validManifest() manifestFixture {
	return manifestFixture{
		SchemaVersion: 1,
		Owner:         "faustbrian",
		Repositories: []repositoryFixture{
			{Name: "go-active", Classification: "active", DefaultBranch: "main"},
			{Name: "go-deferred", Classification: "deferred", DefaultBranch: "main", Reason: "empty"},
			{Name: "go-tooling", Classification: "tooling", DefaultBranch: "main", Reason: "owns tooling"},
		},
	}
}

func repeatedRepositories(count int) []repositoryFixture {
	result := make([]repositoryFixture, count)
	for index := range result {
		result[index] = repositoryFixture{Name: "go-repository", Classification: "active", DefaultBranch: "main"}
	}
	return result
}

func boundaryRepositories(count int) []repositoryFixture {
	result := make([]repositoryFixture, count)
	for index := range result {
		result[index] = repositoryFixture{
			Name:           fmt.Sprintf("go-repository-%03d", index),
			Classification: "active",
			DefaultBranch:  "main",
		}
	}
	result[count-1].Classification = "tooling"
	result[count-1].Reason = "owns tooling"
	return result
}

func encode(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
