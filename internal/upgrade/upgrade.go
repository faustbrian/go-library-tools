// Package upgrade plans and applies coordinated consumer tooling pins.
// Plan is read-only; Apply publishes individually atomic configuration/workflow
// replacements with attempted rollback. Callers own a stable serialized checkout;
// no cross-file crash transaction, remote release discovery, push or fleet
// completion is provided. Result metadata does not certify published adoption.
package upgrade

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

const maximumUpgradeFileSize = 1 << 20

var (
	versionPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	shaPattern     = regexp.MustCompile(`^[0-9a-f]{40}$`)
	digestPattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	configVersion  = regexp.MustCompile(`(?m)^(tool_version:[ \t]*)(v[0-9]+\.[0-9]+\.[0-9]+)[ \t]*$`)
	configDigest   = regexp.MustCompile(`(?m)^(tool_checksums_sha256:[ \t]*)([0-9a-f]{64})[ \t]*$`)
	workflowUse    = regexp.MustCompile(`(?m)^([ \t]*uses:[ \t]*faustbrian/go-library-tools/\.github/workflows/library-ci\.yml@)([0-9a-f]{40})(?:[ \t]+#[^\r\n]*)?[ \t]*$`)
	workflowInput  = regexp.MustCompile(`(?m)^([ \t]*tooling_sha:[ \t]*)([0-9a-f]{40})[ \t]*$`)
)

// Request identifies one immutable tooling release.
type Request struct {
	// Version is an exact vMAJOR.MINOR.PATCH CLI pin, not a release discovery query.
	Version string `json:"version"`
	// WorkflowSHA is the exact 40-character lowercase Git commit for both workflow
	// bindings.
	WorkflowSHA string `json:"workflow_sha"`
	// ChecksumsSHA256 is the lowercase 64-character release checksum-manifest
	// digest.
	ChecksumsSHA256 string `json:"checksums_sha256"`
}

// File describes whether one coordinated pin file changes.
type File struct {
	// Path is one of the coordinated repository-relative configuration/workflow
	// targets.
	Path string `json:"path"`
	// Changed compares planned bytes to the admitted original bytes, not remote
	// adoption.
	Changed bool `json:"changed"`
}

// Result is the deterministic upgrade plan or applied result.
type Result struct {
	// SchemaVersion is 1 for the stable JSON result shape.
	SchemaVersion int `json:"schema_version"`
	// Request is the validated requested identity; no remote publication is
	// inferred.
	Request Request `json:"request"`
	// Changed is true if either coordinated file differs.
	Changed bool `json:"changed"`
	// Files is a deterministic workflow/configuration listing; internal staged bytes
	// remain private.
	Files         []File `json:"files"`
	configuration []byte
	workflow      []byte
	original      map[string][]byte
}

// Plan validates current pins and returns the coordinated changes without
// modifying the repository.
func Plan(root string, request Request) (Result, error) {
	return buildPlan(root, request, repositoryfile.Read)
}

// Apply validates all changes before atomically replacing each affected file.
// If the second replacement fails, it attempts to restore the first file and
// joins rollback failure. Replacements are individually atomic, not one crash-
// atomic transaction; callers must serialize a stable checkout and inspect
// failure before assuming both pins agree. No publication or remote validation
// is performed.
func Apply(root string, request Request) (Result, error) {
	result, err := Plan(root, request)
	if err != nil {
		return Result{}, err
	}
	return apply(result, root, atomicWrite)
}

type reader func(string, string, int64) ([]byte, error)
type writer func(string, string, []byte) error

func buildPlan(root string, request Request, read reader) (Result, error) {
	if err := request.validate(); err != nil {
		return Result{}, err
	}
	configuration, err := read(root, ".golib.yaml", maximumUpgradeFileSize)
	if err != nil {
		return Result{}, fmt.Errorf("read .golib.yaml: %w", err)
	}
	workflow, err := read(root, ".github/workflows/ci.yml", maximumUpgradeFileSize)
	if err != nil {
		return Result{}, fmt.Errorf("read CI workflow: %w", err)
	}
	updatedConfiguration, err := rewriteConfiguration(configuration, request)
	if err != nil {
		return Result{}, err
	}
	updatedWorkflow, err := rewriteWorkflow(workflow, request)
	if err != nil {
		return Result{}, err
	}
	files := []File{
		{Path: ".github/workflows/ci.yml", Changed: !slices.Equal(workflow, updatedWorkflow)},
		{Path: ".golib.yaml", Changed: !slices.Equal(configuration, updatedConfiguration)},
	}
	return Result{
		SchemaVersion: 1,
		Request:       request,
		Changed:       files[0].Changed || files[1].Changed,
		Files:         files,
		configuration: updatedConfiguration,
		workflow:      updatedWorkflow,
		original: map[string][]byte{
			".golib.yaml":              configuration,
			".github/workflows/ci.yml": workflow,
		},
	}, nil
}

func (request Request) validate() error {
	if !versionPattern.MatchString(request.Version) {
		return errors.New("upgrade version must be an exact vMAJOR.MINOR.PATCH release")
	}
	if !shaPattern.MatchString(request.WorkflowSHA) {
		return errors.New("upgrade workflow SHA must be 40 lowercase hexadecimal characters")
	}
	if !digestPattern.MatchString(request.ChecksumsSHA256) {
		return errors.New("upgrade checksums SHA-256 must be 64 lowercase hexadecimal characters")
	}
	return nil
}

func rewriteConfiguration(content []byte, request Request) ([]byte, error) {
	versionMatches := configVersion.FindAllSubmatchIndex(content, len(content))
	if len(versionMatches) != 1 {
		return nil, errors.New(".golib.yaml must contain exactly one canonical tool_version")
	}
	updated := configVersion.ReplaceAll(content, []byte("${1}"+request.Version))
	digestMatches := configDigest.FindAllSubmatchIndex(updated, len(updated))
	switch len(digestMatches) {
	case 0:
		location := configVersion.FindSubmatchIndex(updated)
		lineEnd := location[1]
		updated = slices.Concat(updated[:lineEnd], []byte("\ntool_checksums_sha256: "+request.ChecksumsSHA256), updated[lineEnd:])
	case 1:
		updated = configDigest.ReplaceAll(updated, []byte("${1}"+request.ChecksumsSHA256))
	default:
		return nil, errors.New(".golib.yaml must contain at most one canonical tool_checksums_sha256")
	}
	return updated, nil
}

func rewriteWorkflow(content []byte, request Request) ([]byte, error) {
	useMatches := workflowUse.FindAllSubmatchIndex(content, len(content))
	if len(useMatches) != 1 {
		return nil, errors.New("CI workflow must contain exactly one immutable library-ci reference")
	}
	inputMatches := workflowInput.FindAllSubmatchIndex(content, len(content))
	if len(inputMatches) != 1 {
		return nil, errors.New("CI workflow must contain exactly one tooling_sha input")
	}
	useSHA := string(content[useMatches[0][4]:useMatches[0][5]])
	inputSHA := string(content[inputMatches[0][4]:inputMatches[0][5]])
	if useSHA != inputSHA {
		return nil, errors.New("CI workflow reference and tooling_sha do not match")
	}
	updated := workflowUse.ReplaceAll(content, []byte("${1}"+request.WorkflowSHA+" # "+request.Version))
	updated = workflowInput.ReplaceAll(updated, []byte("${1}"+request.WorkflowSHA))
	return updated, nil
}

func apply(result Result, root string, write writer) (Result, error) {
	configurationChanged := result.Files[1].Changed
	workflowChanged := result.Files[0].Changed
	if configurationChanged {
		if err := write(root, ".golib.yaml", result.configuration); err != nil {
			return Result{}, fmt.Errorf("write .golib.yaml: %w", err)
		}
	}
	if workflowChanged {
		if err := write(root, ".github/workflows/ci.yml", result.workflow); err != nil {
			if configurationChanged {
				rollbackErr := write(root, ".golib.yaml", result.original[".golib.yaml"])
				return Result{}, errors.Join(fmt.Errorf("write CI workflow: %w", err), rollbackErr)
			}
			return Result{}, fmt.Errorf("write CI workflow: %w", err)
		}
	}
	return result, nil
}

func atomicWrite(root, relative string, content []byte) error {
	return atomicWriteWithFiles(root, relative, content, operatingAtomicFiles{})
}

type atomicFile interface {
	Name() string
	Chmod(os.FileMode) error
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

type atomicFiles interface {
	Lstat(string) (os.FileInfo, error)
	CreateTemp(string, string) (atomicFile, error)
	Rename(string, string) error
	Remove(string) error
}

type operatingAtomicFiles struct{}

func (operatingAtomicFiles) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }
func (operatingAtomicFiles) CreateTemp(directory, pattern string) (atomicFile, error) {
	return os.CreateTemp(directory, pattern)
}
func (operatingAtomicFiles) Rename(oldPath, newPath string) error { return os.Rename(oldPath, newPath) }
func (operatingAtomicFiles) Remove(path string) error             { return os.Remove(path) }

func atomicWriteWithFiles(root, relative string, content []byte, files atomicFiles) error {
	target := filepath.Join(root, filepath.FromSlash(relative))
	info, err := files.Lstat(target)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return repositoryfile.ErrNotRegular
	}
	temporary, err := files.CreateTemp(filepath.Dir(target), ".golib-upgrade-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = files.Remove(temporaryPath) }()
	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		_ = temporary.Close()
		return err
	}
	written, err := temporary.Write(content)
	if err != nil {
		_ = temporary.Close()
		return err
	}
	if written != len(content) {
		_ = temporary.Close()
		return io.ErrShortWrite
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := files.Rename(temporaryPath, target); err != nil {
		return err
	}
	return nil
}

// Human renders a concise deterministic result.
func (result Result) Human(applied bool) string {
	verb := "planned"
	if applied {
		verb = "updated"
	}
	if !result.Changed {
		return "tooling pins already current\n"
	}
	var changed []string
	for _, file := range result.Files {
		if file.Changed {
			changed = append(changed, file.Path)
		}
	}
	return fmt.Sprintf("tooling pins %s: %s\n", verb, strings.Join(changed, ", "))
}
