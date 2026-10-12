package cli

import (
	"context"
	"io"

	"github.com/faustbrian/go-library-tools/v2/internal/releasecheck"
)

func executeBinaryRelease(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if args[0] == "archive-check" {
		if len(args) != 3 || args[1] != "--file" || args[2] == "" {
			return usage(stderr, "usage: golib release archive-check --file <path>")
		}
		if err := releasecheck.VerifyBinaryArchive(ctx, args[2]); err != nil {
			return failure(stderr, err)
		}
		return publishValidationSuccess(ctx, stdout, stderr, "binary distribution licenses valid\n")
	}
	if len(args) != 7 || args[1] != "--binary" || args[2] == "" || args[3] != "--license" || args[4] == "" || args[5] != "--output" || args[6] == "" {
		return usage(stderr, "usage: golib release package --binary <path> --license <path> --output <path>")
	}
	if err := releasecheck.PackageBinary(ctx, args[2], args[4], args[6]); err != nil {
		return failure(stderr, err)
	}
	return publishValidationSuccess(ctx, stdout, stderr, "binary distribution packaged\n")
}
