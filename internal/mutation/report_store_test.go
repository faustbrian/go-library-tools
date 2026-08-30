package mutation_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestStoreAndLoadReportByInputIdentity(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mutation")
	input := "sha256:" + strings.Repeat("a", 64)
	report := []byte(`{"files":[{"file_name":"source.go","mutations":[{"type":"ARITHMETIC_BASE","status":"KILLED","line":2,"column":3}]}]}`)
	path, reused, result, err := mutation.StoreReport(root, input, report)
	if err != nil || reused || result.Mutants != 1 {
		t.Fatalf("StoreReport() = %q, %v, %#v, %v", path, reused, result, err)
	}
	secondPath, reused, secondResult, err := mutation.StoreReport(root, input, report)
	if err != nil || !reused || secondPath != path || secondResult.Digest != result.Digest {
		t.Fatalf("StoreReport(reuse) = %q, %v, %#v, %v", secondPath, reused, secondResult, err)
	}
	data, loaded, err := mutation.LoadReport(root, input)
	if err != nil || string(data) != string(report) || loaded.Digest != result.Digest {
		t.Fatalf("LoadReport() = %q, %#v, %v", data, loaded, err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestStoreReportRejectsConflictingContent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mutation")
	input := "sha256:" + strings.Repeat("b", 64)
	first := []byte(`{"files":[{"file_name":"a.go","mutations":[{"type":"A","status":"KILLED","line":1,"column":1}]}]}`)
	second := []byte(`{"files":[{"file_name":"b.go","mutations":[{"type":"B","status":"KILLED","line":1,"column":1}]}]}`)
	if _, _, _, err := mutation.StoreReport(root, input, first); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := mutation.StoreReport(root, input, second); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("StoreReport(conflict) error = %v", err)
	}
}

func TestStoreReportReusesSemanticallyIdenticalGremlinsReports(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mutation")
	input := "sha256:" + strings.Repeat("c", 64)
	first := []byte(`{"go_module":"example","elapsed_time":1.25,"files":[{"file_name":"b.go","mutations":[{"type":"B","status":"KILLED","line":2,"column":3}]},{"file_name":"a.go","mutations":[{"type":"Z","status":"KILLED","line":4,"column":5},{"type":"A","status":"KILLED","line":1,"column":2}]}],"mutants_killed":3,"mutants_lived":0,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":3,"mutations_coverage":100,"mutator_statistics":{"B":{"killed":1},"A":{"killed":2}},"test_efficacy":100}`)
	second := []byte(`{"test_efficacy":100,"mutator_statistics":{"A":{"killed":2},"B":{"killed":1}},"mutations_coverage":100,"mutants_total":3,"mutants_not_viable":0,"mutants_not_covered":0,"mutants_lived":0,"mutants_killed":3,"files":[{"file_name":"a.go","mutations":[{"column":2,"line":1,"status":"KILLED","type":"A"},{"column":5,"line":4,"status":"KILLED","type":"Z"}]},{"file_name":"b.go","mutations":[{"column":3,"line":2,"status":"KILLED","type":"B"}]}],"elapsed_time":99,"go_module":"example"}`)
	path, reused, firstResult, err := mutation.StoreReport(root, input, first)
	if err != nil || reused {
		t.Fatalf("StoreReport(first) = %q, %v, %#v, %v", path, reused, firstResult, err)
	}
	secondPath, reused, secondResult, err := mutation.StoreReport(root, input, second)
	if err != nil || !reused || secondPath != path || secondResult.Digest != firstResult.Digest {
		t.Fatalf("StoreReport(semantic reuse) = %q, %v, %#v, %v", secondPath, reused, secondResult, err)
	}
	stored, _, err := mutation.LoadReport(root, input)
	if err != nil || string(stored) != string(first) {
		t.Fatalf("LoadReport() retained = %q, error = %v", stored, err)
	}

	changed := []byte(strings.Replace(string(second), `"line":1`, `"line":9`, 1))
	if _, _, _, err := mutation.StoreReport(root, input, changed); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("StoreReport(changed identity) error = %v", err)
	}
}

func TestStoreReportConcurrentlyReusesSemanticContent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mutation")
	input := "sha256:" + strings.Repeat("d", 64)
	reports := [][]byte{
		[]byte(`{"elapsed_time":1,"files":[{"file_name":"b.go","mutations":[{"type":"B","status":"KILLED","line":2,"column":3}]},{"file_name":"a.go","mutations":[{"type":"A","status":"KILLED","line":1,"column":2}]}]}`),
		[]byte(`{"files":[{"mutations":[{"column":2,"line":1,"status":"KILLED","type":"A"}],"file_name":"a.go"},{"mutations":[{"column":3,"line":2,"status":"KILLED","type":"B"}],"file_name":"b.go"}],"elapsed_time":42}`),
	}
	type outcome struct {
		reused bool
		err    error
	}
	outcomes := make(chan outcome, 8)
	var workers sync.WaitGroup
	for index := range cap(outcomes) {
		workers.Add(1)
		go func(report []byte) {
			defer workers.Done()
			_, reused, _, err := mutation.StoreReport(root, input, report)
			outcomes <- outcome{reused: reused, err: err}
		}(reports[index%len(reports)])
	}
	workers.Wait()
	close(outcomes)
	created := 0
	for result := range outcomes {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if !result.reused {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("new reports = %d, want 1", created)
	}
}
