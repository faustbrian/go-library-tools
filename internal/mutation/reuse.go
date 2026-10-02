package mutation

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/faustbrian/go-library-tools/v2/internal/evidence"
)

// SemanticVerifierDigest is the content identity recorded for native evidence.
func SemanticVerifierDigest() string { return "sha256:" + LegacyVerifierDigest() }

// Reuse validates the exact evidence record and report for one package input.
// Missing evidence is a cache miss; malformed or incomplete evidence fails.
func Reuse(evidenceRoot, mutationRoot, repository, module, pkg, inputDigest string) (bool, ReportResult, error) {
	return ReuseWithReview(evidenceRoot, mutationRoot, repository, module, pkg, inputDigest, nil)
}

// ReuseWithReview checks the current exact classification before accepting an
// immutable report, independently of the evidence record's earlier verdict.
func ReuseWithReview(evidenceRoot, mutationRoot, repository, module, pkg, inputDigest string, review *EquivalentReview) (bool, ReportResult, error) {
	record, err := evidence.Load(evidenceRoot, "mutation", inputDigest)
	if errors.Is(err, os.ErrNotExist) {
		return false, ReportResult{}, nil
	}
	if err != nil {
		return false, ReportResult{}, err
	}
	if record.Repository != repository || record.Module != module || record.Package != pkg ||
		record.Result != "passed" || !applicableVerifierRecord(record) {
		return false, ReportResult{}, fmt.Errorf("%w: mutation evidence identity does not match requested package", ErrInvalid)
	}
	data, report, err := LoadReportWithReview(mutationRoot, inputDigest, review)
	if err != nil {
		return false, ReportResult{}, err
	}
	if report.Digest != record.ReportDigest && legacyCanonicalReportDigest(data) != record.ReportDigest {
		return false, ReportResult{}, fmt.Errorf("%w: mutation evidence report digest does not match", ErrInvalid)
	}
	return true, report, nil
}

func applicableVerifierRecord(record evidence.Record) bool {
	if record.VerifierDigest == SemanticVerifierDigest() {
		return true
	}
	// This is applicable historical execution, never corrected-verifier
	// execution. Import proves source-directory equivalence before persisting
	// the existing exact-input record; the digest alone grants no reuse.
	return historicalVerifierDigest(strings.TrimPrefix(record.VerifierDigest, "sha256:")) &&
		record.Environment["evidence_origin"] == "approved_legacy_checkpoint" &&
		record.Environment[legacyPackagePathCompatibility] == record.InputDigest
}
