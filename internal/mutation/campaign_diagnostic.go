package mutation

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
)

const maximumFailedCoordinates = 16

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
