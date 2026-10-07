package mutation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

const maximumFailedCoordinates = 16

// failedDiagnostic retains native coordinates and numeric counter presence,
// but is never validation, complete accounting, or reusable passing evidence.
func (campaign Campaign) failedDiagnostic(output io.Writer, packageDirectory, input, reportPath string, failure error) error {
	path, err := campaign.captureFailedDiagnostic(packageDirectory, input, reportPath)
	if err != nil {
		_, _ = fmt.Fprintln(output, "mutation diagnostic: failed report could not be captured")
		return errors.Join(failure, fmt.Errorf("%w: failed mutation diagnostic capture", ErrInvalid))
	}
	reportFailedMutationCoordinates(output, path)
	return failure
}

func (campaign Campaign) captureFailedDiagnostic(packageDirectory, input, reportPath string) (string, error) {
	if !filepath.IsAbs(campaign.MutationRoot) || !strings.HasPrefix(input, "sha256:") || !digestRE.MatchString(strings.TrimPrefix(input, "sha256:")) {
		return "", ErrInvalid
	}
	data, err := repositoryfile.Read(campaign.Workspace, filepath.Base(reportPath), maximumCheckpointSize)
	if err != nil {
		return "", ErrInvalid
	}
	var projected report
	if err := decodeStrict(data, &projected); err != nil || projected.Files == nil {
		return "", ErrInvalid
	}
	packagePath := filepath.Join(filepath.FromSlash(campaign.Policy.ModuleDirectory), filepath.FromSlash(packageDirectory))
	if err := repositoryfile.ValidateDirectory(campaign.Root, filepath.ToSlash(packagePath)); err != nil {
		return "", ErrInvalid
	}
	directory := filepath.Join(campaign.Root, packagePath)
	entries, err := readSourceEntries(directory, maximumInputEntries)
	if err != nil {
		return "", ErrInvalid
	}
	admitted := make(map[string]bool)
	for _, entry := range entries {
		name := entry.Name()
		if !entry.Type().IsRegular() || filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") || strings.ContainsFunc(name, unicode.IsControl) {
			continue
		}
		if err := repositoryfile.ValidateRegularFile(directory, name); err != nil {
			return "", ErrInvalid
		}
		admitted[name] = true
		admitted[filepath.ToSlash(filepath.Join(filepath.FromSlash(packageDirectory), name))] = true
	}
	for _, file := range projected.Files {
		if !admitted[file.FileName] {
			return "", ErrInvalid
		}
		for _, candidate := range file.Mutations {
			if candidate.Line <= 0 || candidate.Column <= 0 || !nativeDiagnosticType(candidate.Type) || !nativeDiagnosticStatus(candidate.Status) {
				return "", ErrInvalid
			}
		}
	}
	projected.GoModule, projected.ElapsedTime, projected.MutatorStatistics = "", nil, nil
	data, err = json.Marshal(projected)
	if err != nil {
		return "", ErrInvalid
	}
	digest := sha256.Sum256(data)
	destinationDirectory := filepath.Join(campaign.MutationRoot, "failed-reports")
	destination := filepath.Join(destinationDirectory, strings.TrimPrefix(input, "sha256:")+"-"+hex.EncodeToString(digest[:])+".json")
	path, _, _, err := publishReport(operatingReportFiles{}, campaign.MutationRoot, destinationDirectory, destination, data, func(existing []byte) bool { return bytes.Equal(existing, data) })
	return path, err
}

func nativeDiagnosticType(value string) bool {
	switch value {
	case "ARITHMETIC_BASE", "CONDITIONALS_BOUNDARY", "CONDITIONALS_NEGATION", "INCREMENT_DECREMENT", "INVERT_ASSIGNMENTS", "INVERT_BITWISE", "INVERT_BWASSIGN", "INVERT_LOGICAL", "INVERT_LOOPCTRL", "INVERT_NEGATIVES", "REMOVE_SELF_ASSIGNMENTS":
		return true
	default:
		return false
	}
}

func nativeDiagnosticStatus(value string) bool {
	switch value {
	case "NOT COVERED", "RUNNABLE", "SKIPPED", "LIVED", "KILLED", "NOT VIABLE", "TIMED OUT":
		return true
	default:
		return false
	}
}

type exitCoder interface{ ExitCode() int }

func isEfficacyThresholdExit(err error) bool {
	var exit exitCoder
	return errors.As(err, &exit) && exit.ExitCode() == 10
}

func equivalentSelectionEqual(left, right *EquivalentReview) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if left.ModuleDirectory != right.ModuleDirectory || left.PackageDirectory != right.PackageDirectory ||
		left.SourceDigest != right.SourceDigest || left.GremlinsVersion != right.GremlinsVersion ||
		left.GremlinsVerifierSHA256 != right.GremlinsVerifierSHA256 {
		return false
	}
	selection := func(review *EquivalentReview) map[string]string {
		result := make(map[string]string, len(review.Mutations))
		for _, candidate := range review.Mutations {
			result[mutationIdentity(candidate.FileName, candidate.Type, candidate.Line, candidate.Column)] = candidate.ContractDomain
		}
		return result
	}
	return maps.Equal(selection(left), selection(right))
}

// reportFailedMutationCoordinates exposes only bounded native status and
// coordinate diagnostics. Failed reports are never installed as checkpoints.
func reportFailedMutationCoordinates(output io.Writer, path string) {
	file, err := os.Open(path) // #nosec G304 -- fixed report path in the task-owned campaign workspace
	if err != nil {
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maximumCheckpointSize+1))
	if err != nil || len(data) > maximumCheckpointSize {
		_, _ = fmt.Fprintln(output, "mutation diagnostic: native report unavailable or oversized")
		return
	}
	var native report
	if err := decodeStrict(data, &native); err != nil || native.Files == nil {
		_, _ = fmt.Fprintln(output, "mutation diagnostic: native report malformed")
		return
	}
	count := 0
	for _, file := range native.Files {
		for _, candidate := range file.Mutations {
			if candidate.Status == "KILLED" {
				continue
			}
			count++
			if count <= maximumFailedCoordinates {
				_, _ = fmt.Fprintf(output, "mutation diagnostic: %q %q %q %d:%d\n", candidate.Status, file.FileName, candidate.Type, candidate.Line, candidate.Column)
			}
		}
	}
	if count > maximumFailedCoordinates {
		_, _ = fmt.Fprintf(output, "mutation diagnostic: %d additional non-killed coordinates omitted\n", count-maximumFailedCoordinates)
	}
}
