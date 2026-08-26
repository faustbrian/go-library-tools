package evidence

import (
	"strings"
	"testing"
	"time"
)

func TestEquivalentRequiresEverySemanticField(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	base := Record{
		SchemaVersion: SchemaVersion, Repository: "example", Module: ".", Package: "package", Gate: "tests",
		InputDigest: digest, VerifierDigest: digest, Result: "passed", ReportDigest: digest,
		Environment: map[string]string{"go": "1.27.0"}, CompletedAt: time.Unix(1, 0).UTC(),
	}
	if !equivalent(base, base) {
		t.Fatal("equivalent rejected identical records")
	}
	tests := map[string]func(*Record){
		"schema":      func(record *Record) { record.SchemaVersion++ },
		"repository":  func(record *Record) { record.Repository = "other" },
		"module":      func(record *Record) { record.Module = "nested" },
		"package":     func(record *Record) { record.Package = "other" },
		"gate":        func(record *Record) { record.Gate = "coverage" },
		"input":       func(record *Record) { record.InputDigest = "sha256:" + strings.Repeat("b", 64) },
		"verifier":    func(record *Record) { record.VerifierDigest = "sha256:" + strings.Repeat("b", 64) },
		"result":      func(record *Record) { record.Result = "failed" },
		"report":      func(record *Record) { record.ReportDigest = "sha256:" + strings.Repeat("b", 64) },
		"environment": func(record *Record) { record.Environment = map[string]string{"go": "1.27.1"} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			other := base
			mutate(&other)
			if equivalent(base, other) {
				t.Fatal("equivalent accepted a semantic mismatch")
			}
		})
	}
}
