package repository_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParityComparisonAcceptsStrongerMutationInventory(t *testing.T) {
	root := t.TempDir()
	writeParityFixture(t, root, "legacy", []string{".\t./convert\t963"})
	writeParityFixture(t, root, "shared", []string{".\t./convert\t977"})

	command := exec.CommandContext(t.Context(), "bash", filepath.Join(projectRoot(t), "rehearsals", "compare.sh"), filepath.Join(root, "artifacts"), filepath.Join(root, "repositories.json"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compare stronger mutation inventory: %v: %s", err, output)
	}
}

func TestParityComparisonRejectsWeakerMutationInventory(t *testing.T) {
	root := t.TempDir()
	writeParityFixture(t, root, "legacy", []string{".\t./convert\t977"})
	writeParityFixture(t, root, "shared", []string{".\t./convert\t963"})

	command := exec.CommandContext(t.Context(), "bash", filepath.Join(projectRoot(t), "rehearsals", "compare.sh"), filepath.Join(root, "artifacts"), filepath.Join(root, "repositories.json"))
	if err := command.Run(); err == nil {
		t.Fatal("compare weaker mutation inventory error = nil")
	}
}

func TestParityComparisonRejectsDuplicateMutationRecords(t *testing.T) {
	root := t.TempDir()
	writeParityFixture(t, root, "legacy", []string{".\t./convert\t963"})
	writeParityFixture(t, root, "shared", []string{
		".\t./convert\t977",
		".\t./convert\t978",
	})

	command := exec.CommandContext(t.Context(), "bash", filepath.Join(projectRoot(t), "rehearsals", "compare.sh"), filepath.Join(root, "artifacts"), filepath.Join(root, "repositories.json"))
	if err := command.Run(); err == nil {
		t.Fatal("compare duplicate mutation records error = nil")
	}
}

func writeParityFixture(t *testing.T, root, mode string, mutations []string) {
	t.Helper()
	manifest := []byte(`{"repositories":[{"name":"go-example"}]}`)
	if err := os.WriteFile(filepath.Join(root, "repositories.json"), manifest, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	directory := filepath.Join(root, "artifacts", "parity-"+mode+"-go-example")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("create artifact directory: %v", err)
	}
	summary := map[string]any{
		"schema_version":         1,
		"mode":                   mode,
		"content_digest":         "content",
		"check_status":           0,
		"service_cleanup_status": 0,
		"selected_modules":       []string{".\texample"},
		"effective_gates":        []string{".\tmutation"},
		"required_services":      []string{},
		"release_units":          []string{".\tv"},
		"coverage":               []string{"example\t1\t1"},
		"mutation":               mutations,
		"nilaway_advisories":     []string{".\t0"},
		"release_decision":       "tag-collision",
	}
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, mode+"-summary"), data, 0o600); err != nil {
		t.Fatalf("write summary: %v", err)
	}
}
