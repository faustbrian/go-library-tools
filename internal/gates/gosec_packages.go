package gates

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

const (
	maximumGosecPackages  = 4096
	maximumGosecArguments = 128 << 10
)

// Keep package discovery distinct from the Git source inventory. Both reuse
// the bounded capture and original-process-group cancellation owner.
type goPackageOutput struct{ sourceInventoryOutput }

// Ordinary Go selection excludes testdata, vendor, ignored directories and
// nested modules. Gosec's recursive filesystem walk does not, so pass only
// explicit directories selected by Go, including unregistered support code.
func (runner Runner) gosecPackages(ctx context.Context, directory string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return nil, errors.New("gosec package discovery requires an absolute module root")
	}
	stdout := &goPackageOutput{sourceInventoryOutput: sourceInventoryOutput{boundedProcessOutput: boundedProcessOutput{limit: maximumSecurityProcessOutput}}}
	stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	err = runner.Executor.Run(ctx, Command{
		Name: "go", Args: []string{"list", "-json=Dir", "./..."}, Dir: directory,
		Env: map[string]string{"GOWORK": "off"}, Stdout: stdout, Stderr: stderr, boundedScanner: true,
	})
	if err != nil || stdout.didOverflow() || stderr.didOverflow() || ctx.Err() != nil {
		return nil, errors.Join(sourceCommandError{class: "gosec package discovery failed or exceeded output limit", cause: err}, ctx.Err())
	}
	decoder := json.NewDecoder(&stdout.data)
	decoder.DisallowUnknownFields()
	packages := make([]string, 0)
	seen := make(map[string]bool)
	argumentBytes := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var selected struct{ Dir string }
		if err := decoder.Decode(&selected); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.New("gosec package discovery returned malformed output")
		}
		if !filepath.IsAbs(selected.Dir) || filepath.Clean(selected.Dir) != selected.Dir || strings.ContainsAny(selected.Dir, "\x00\r\n") {
			return nil, errors.New("gosec package discovery returned an invalid directory")
		}
		relative, err := filepath.Rel(root, selected.Dir)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || strings.HasSuffix(relative, "...") {
			return nil, errors.New("gosec package discovery returned an invalid directory")
		}
		argument := "./"
		if relative != "." {
			argument += filepath.ToSlash(relative)
		}
		argumentBytes += len(argument) + 1
		if seen[argument] || len(packages) >= maximumGosecPackages || len(argument) > 4096 || argumentBytes > maximumGosecArguments {
			return nil, errors.New("gosec package discovery returned duplicate or excessive directories")
		}
		seen[argument] = true
		packages = append(packages, argument)
	}
	if len(packages) == 0 {
		return nil, errors.New("gosec package discovery returned no packages")
	}
	slices.Sort(packages)
	return packages, nil
}
