package mutation

import (
	"errors"
	"fmt"
	"os"

	"github.com/faustbrian/go-library-tools/internal/evidence"
)

// SemanticVerifierDigest is the content identity recorded for native evidence.
func SemanticVerifierDigest() string { return "sha256:" + LegacyVerifierDigest() }

// Reuse validates the exact evidence record and report for one package input.
// Missing evidence is a cache miss; malformed or incomplete evidence fails.
func Reuse(evidenceRoot, mutationRoot, repository, module, pkg, inputDigest string) (bool, ReportResult, error) {
	record, err := evidence.Load(evidenceRoot, "mutation", inputDigest)
	if errors.Is(err, os.ErrNotExist) {
		return false, ReportResult{}, nil
	}
	if err != nil {
		return false, ReportResult{}, err
	}
	if record.Repository != repository || record.Module != module || record.Package != pkg ||
		record.Result != "passed" || record.VerifierDigest != SemanticVerifierDigest() {
		return false, ReportResult{}, fmt.Errorf("%w: mutation evidence identity does not match requested package", ErrInvalid)
	}
	data, report, err := LoadReport(mutationRoot, inputDigest)
	if err != nil {
		return false, ReportResult{}, err
	}
	if report.Digest != record.ReportDigest && legacyCanonicalReportDigest(data) != record.ReportDigest {
		return false, ReportResult{}, fmt.Errorf("%w: mutation evidence report digest does not match", ErrInvalid)
	}
	return true, report, nil
}
