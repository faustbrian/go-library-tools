package cohesion

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSourceLockPropagatesBoundaryFailures(t *testing.T) {
	failure := errors.New("failure")
	for name, mutate := range map[string]func(*sourceLockOperations){
		"read": func(operations *sourceLockOperations) {
			operations.read = func(string, int64, string) ([]byte, error) { return nil, failure }
		},
		"schema": func(operations *sourceLockOperations) {
			operations.validate = func([]byte) error { return failure }
		},
		"decode": func(operations *sourceLockOperations) {
			operations.decode = func([]byte, *SourceLock) error { return failure }
		},
		"semantic": func(operations *sourceLockOperations) {
			operations.semantic = func(SourceLock) error { return failure }
		},
	} {
		t.Run(name, func(t *testing.T) {
			operations := sourceLockOperations{
				read:     func(string, int64, string) ([]byte, error) { return []byte(`{}`), nil },
				validate: func([]byte) error { return nil },
				decode:   func([]byte, *SourceLock) error { return nil },
				semantic: func(SourceLock) error { return nil },
			}
			mutate(&operations)
			if _, err := loadSourceLockWithOperations("sources.json", operations); !errors.Is(err, failure) {
				t.Fatalf("loadSourceLockWithOperations() error = %v", err)
			}
		})
	}
}

func TestLoadSourceLockAcceptsReviewedManifest(t *testing.T) {
	lock, err := LoadSourceLock("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatalf("LoadSourceLock() error = %v", err)
	}
	if lock.SchemaVersion != 1 || lock.RepositoryCount != 92 || len(lock.Repositories) != 92 {
		t.Fatalf("LoadSourceLock() = schema %d, count %d, %d repositories", lock.SchemaVersion, lock.RepositoryCount, len(lock.Repositories))
	}
	if lock.Repositories[0].Repository != "github.com/faustbrian/go-adaptive-throttle" {
		t.Fatalf("first repository = %q", lock.Repositories[0].Repository)
	}
}

func TestLoadSourceLockRejectsRepositoryCountMismatch(t *testing.T) {
	data, err := os.ReadFile("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	var candidate map[string]any
	if err := json.Unmarshal(data, &candidate); err != nil {
		t.Fatal(err)
	}
	candidate["repository_count"] = 91
	candidateData, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "sources.json")
	if err := os.WriteFile(path, candidateData, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSourceLock(path); err == nil || !strings.Contains(err.Error(), "repository_count") {
		t.Fatalf("LoadSourceLock() error = %v", err)
	}
}

func TestValidateSourceLockOrderRejectsMultipleReleaseSources(t *testing.T) {
	lock := SourceLock{
		RepositoryCount: 2,
		Repositories: []SourceLockRepository{
			{Repository: "github.com/faustbrian/go-a", Source: SourceLockRevision{Kind: "release-source"}},
			{Repository: "github.com/faustbrian/go-b", Source: SourceLockRevision{Kind: "release-source"}},
		},
	}
	if err := validateSourceLockOrder(lock); err == nil || !strings.Contains(err.Error(), "found 2") {
		t.Fatalf("validateSourceLockOrder() error = %v", err)
	}
}

func TestReviewedSourceLockHasExpectedReleasePinDistribution(t *testing.T) {
	lock, err := LoadSourceLock("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, repository := range lock.Repositories {
		if repository.Source.Kind == "release-source" {
			counts["release-source"]++
			continue
		}
		if repository.Tooling == nil {
			t.Fatalf("consumer %s has no tooling identity", repository.Repository)
		}
		counts[repository.Tooling.Version+"/"+repository.Tooling.ChecksumsSHA256]++
	}
	if counts["release-source"] != 1 ||
		counts["v1.4.0/ba1b71d41dc9b58d5bfb411bc89ae2d45b6f3f778152714bfd1afeab0bef2f33"] != 90 ||
		counts["v1.5.0/fb648d7a23d6b845f8cc475d43769c74bbdcb6d9296f0b0e70fc1b7e381909fd"] != 1 {
		t.Fatalf("reviewed source-lock distribution = %#v", counts)
	}
}

func TestLoadSourceLockRejectsSemanticAmbiguity(t *testing.T) {
	data, err := os.ReadFile("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	var valid map[string]any
	if err := json.Unmarshal(data, &valid); err != nil {
		t.Fatal(err)
	}

	for name, test := range map[string]struct {
		mutate func([]any)
		want   string
	}{
		"unsorted": {
			mutate: func(repositories []any) { repositories[0], repositories[1] = repositories[1], repositories[0] },
			want:   "strictly sorted",
		},
		"duplicate": {
			mutate: func(repositories []any) { repositories[1] = repositories[0] },
			want:   "strictly sorted",
		},
		"missing release source": {
			mutate: func(repositories []any) {
				for _, raw := range repositories {
					repository := jsonObjectValue(raw)
					source := jsonObjectValue(repository["source"])
					if source["kind"] == "release-source" {
						repository["source"] = map[string]any{"kind": "commit", "commit": strings.Repeat("a", 40)}
						repository["tooling"] = map[string]any{"version": "v1.5.0", "checksums_sha256": strings.Repeat("b", 64)}
						return
					}
				}
			},
			want: "exactly one release source",
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := cloneJSONMap(t, valid)
			repositories := jsonArrayValue(candidate["repositories"])
			test.mutate(repositories)
			path := filepath.Join(t.TempDir(), "sources.json")
			candidateData, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, candidateData, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadSourceLock(path); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("LoadSourceLock() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadSourceLockRejectsOversizedInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sources.json")
	if err := os.WriteFile(path, make([]byte, maximumSourceLockSize+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSourceLock(path); err == nil || !strings.Contains(err.Error(), "exceed maximum size") {
		t.Fatalf("LoadSourceLock() error = %v", err)
	}
}

func TestSourceLockVerifyPolicyMatchesExactConsumerIdentity(t *testing.T) {
	lock, err := LoadSourceLock("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	consumer := lock.Repositories[0]
	if consumer.Tooling == nil {
		t.Fatal("first source-lock repository has no tooling identity")
	}
	if err := lock.VerifyPolicy(consumer.Repository, consumer.Tooling.Version, consumer.Tooling.ChecksumsSHA256); err != nil {
		t.Fatalf("VerifyPolicy(exact) error = %v", err)
	}
	for name, values := range map[string][3]string{
		"unknown repository": {"github.com/faustbrian/go-unknown", consumer.Tooling.Version, consumer.Tooling.ChecksumsSHA256},
		"wrong version":      {consumer.Repository, "v9.9.9", consumer.Tooling.ChecksumsSHA256},
		"wrong digest":       {consumer.Repository, consumer.Tooling.Version, strings.Repeat("c", 64)},
	} {
		t.Run(name, func(t *testing.T) {
			if err := lock.VerifyPolicy(values[0], values[1], values[2]); err == nil {
				t.Fatal("VerifyPolicy() error = nil")
			}
		})
	}
	if err := lock.VerifyPolicy("github.com/faustbrian/go-library-tools", "v1.0.0", ""); err != nil {
		t.Fatalf("VerifyPolicy(release source) error = %v", err)
	}
}
