// Package sourcedocs inventories objective Go comment structure across tracked
// source variants without compiling packages, executing generators, or scoring prose.
package sourcedocs

import (
	"bytes"
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

const (
	maximumPaths      = 20000
	maximumPathBytes  = 1 << 20
	maximumFileBytes  = 4 << 20
	maximumTotalBytes = 128 << 20
	maximumFindings   = 20000
)

// Finding identifies a structural error or a manual-review item. Messages never
// contain source text, link destinations, generator commands, or credential values.
type Finding struct {
	// Path is repository-relative, or redacted when it resembles a credential.
	Path string `json:"path"`
	// Line is the one-based source line of the associated comment or declaration.
	Line int `json:"line"`
	// Code is the stable structural or manual-review diagnostic identifier.
	Code string `json:"code"`
	// Severity is error for strict failures or review for semantic investigation.
	Severity string `json:"severity"`
	// Message describes the check without copying repository-controlled text.
	Message string `json:"message"`
}

// File records scope without treating test fixtures or generated input as
// hand-maintained production code. Build constraints are inspected, never selected.
type File struct {
	// Path is the tracked repository-relative source path, with credential redaction.
	Path string `json:"path"`
	// Module is the nearest tracked go.mod directory, including fixture boundaries.
	Module string `json:"module"`
	// Package is the parsed package name; malformed negative fixtures leave it empty.
	Package string `json:"package"`
	// Kind is source, test, or fixture; fixture classification takes precedence.
	Kind string `json:"kind"`
	// Generated records Go's canonical generated header, not verified reproducibility.
	Generated bool `json:"generated"`
	// BuildConstrained records explicit build-tag text, not platform filename selection.
	BuildConstrained bool `json:"build_constrained"`
}

// Module identifies a tracked go.mod boundary, including fixture modules.
type Module struct {
	// Directory is the repository-relative go.mod owner, or a redacted path.
	Directory string `json:"directory"`
	// Kind distinguishes ordinary module boundaries from test fixture boundaries.
	Kind string `json:"kind"`
}

// Report is a deterministic inventory, not certification of comment meaning,
// supported build combinations, generator executability, or example compilation.
type Report struct {
	// SchemaVersion identifies the additive inventory representation, currently 1.
	SchemaVersion int `json:"schema_version"`
	// Modules lists every non-vendor tracked module boundary in directory order.
	Modules []Module `json:"modules"`
	// Files lists checked Go input in deterministic tracked-path order, never build-selected order.
	Files []File `json:"files"`
	// ExcludedVendorFiles counts tracked vendor Go files without claiming an audit.
	ExcludedVendorFiles int `json:"excluded_vendor_files"`
	// Findings is a non-nil path/line/code-sorted list of objective and review items.
	Findings []Finding `json:"findings"`
}

// HasErrors reports whether the explicit strict check must fail. Review items
// remain visible but cannot turn subjective prose judgments into a CI gate.
func (report Report) HasErrors() bool {
	return slices.ContainsFunc(report.Findings, func(f Finding) bool { return f.Severity == "error" })
}

type parsedFile struct {
	file     File
	relative string
	ast      *ast.File
	fset     *token.FileSet
}

// Audit inspects current contents of Git-tracked files across all modules,
// platforms, tags, and tests. Paths must remain stable during the bounded pass;
// symlinked source and oversized, missing, or malformed input fail closed.
// Untracked files and vendor source are not checked. Cancellation is cooperative
// during bounded parsing and filesystem reads; Git acquisition has a 30s deadline.
func Audit(ctx context.Context, root string) (Report, error) {
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil || !filepath.IsAbs(canonical) {
		return Report{}, errors.New("source documentation repository root is invalid")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	identity, err := gitOutput(ctx, canonical, "rev-parse", "--show-toplevel")
	if err != nil || strings.TrimSpace(string(identity)) != canonical {
		return Report{}, errors.New("source documentation requires the exact Git repository root")
	}
	data, err := gitOutput(ctx, canonical, "ls-files", "--cached", "--full-name", "-z")
	if err != nil {
		return Report{}, err
	}
	paths, err := parsePaths(data)
	if err != nil {
		return Report{}, err
	}
	return auditPaths(ctx, canonical, paths)
}

type boundedBuffer struct{ bytes.Buffer }

// Write refuses excess Git output before retaining any bytes from that write.
// Earlier bounded writes remain private and cannot be published as partial proof.
func (buffer *boundedBuffer) Write(data []byte) (int, error) {
	if len(data) > maximumPathBytes-buffer.Len() {
		return 0, errors.New("tracked source path output limit exceeded")
	}
	return buffer.Buffer.Write(data)
}

func gitOutput(ctx context.Context, root string, args ...string) ([]byte, error) {
	// Git index inspection runs no repository hooks, filters, generators or shell.
	// Ambient Git overrides cannot redirect the repository or inject configuration.
	command := exec.CommandContext(ctx, "git", append([]string{"-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-C", root}, args...)...)
	command.Dir = root
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if !strings.HasPrefix(name, "GIT_") {
			command.Env = append(command.Env, variable)
		}
	}
	command.Env = append(command.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0")
	var output boundedBuffer
	command.Stdout, command.Stderr = &output, io.Discard
	command.WaitDelay = time.Second
	if err := command.Run(); err != nil {
		return nil, errors.Join(errors.New("tracked source Git inventory failed or exceeded its bound"), ctx.Err())
	}
	return output.Bytes(), nil
}

func parsePaths(data []byte) ([]string, error) {
	if len(data) > maximumPathBytes || (len(data) > 0 && data[len(data)-1] != 0) {
		return nil, errors.New("tracked source path inventory is oversized or malformed")
	}
	paths := []string{}
	for _, name := range bytes.Split(bytes.TrimSuffix(data, []byte{0}), []byte{0}) {
		if len(data) == 0 {
			break
		}
		value := string(name)
		if len(paths) >= maximumPaths || len(value) > 1024 || value == "" || !utf8.ValidString(value) || path.IsAbs(value) || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") || strings.Contains(value, "\\") || strings.ContainsFunc(value, func(r rune) bool { return unicode.IsControl(r) || unicode.Is(unicode.Cf, r) }) {
			return nil, errors.New("tracked source path inventory contains an unsafe or excessive path")
		}
		paths = append(paths, value)
	}
	slices.Sort(paths)
	return slices.Compact(paths), nil
}

func fixture(name string) bool {
	for _, component := range strings.Split(name, "/") {
		if component == "testdata" || component == "fixtures" || component == "fixture" {
			return true
		}
	}
	return false
}

func vendor(name string) bool { return slices.Contains(strings.Split(name, "/"), "vendor") }

func safePath(name string) string {
	if secretPattern.MatchString(name) {
		return "[redacted-path]"
	}
	return name
}

func auditPaths(ctx context.Context, root string, paths []string) (Report, error) {
	report := Report{SchemaVersion: 1, Modules: []Module{}, Files: []File{}, Findings: []Finding{}}
	moduleDirectories := []string{}
	total := 0
	for _, name := range paths {
		if err := ctx.Err(); err != nil {
			return Report{}, err
		}
		if path.Base(name) == "go.mod" && !vendor(name) {
			content, err := repositoryfile.Read(root, filepath.FromSlash(name), maximumFileBytes)
			if err != nil {
				return Report{}, errors.New("tracked module boundary is unavailable, unsafe, or oversized")
			}
			if len(content) > maximumTotalBytes-total {
				return Report{}, errors.New("source documentation total byte limit exceeded")
			}
			total += len(content)
			directory := path.Dir(name)
			kind := "module"
			if fixture(name) {
				kind = "fixture"
			}
			report.Modules = append(report.Modules, Module{safePath(directory), kind})
			moduleDirectories = append(moduleDirectories, directory)
		}
	}
	parsed := []parsedFile{}
	for _, name := range paths {
		if err := ctx.Err(); err != nil {
			return Report{}, err
		}
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if vendor(name) {
			report.ExcludedVendorFiles++
			continue
		}
		content, err := repositoryfile.Read(root, filepath.FromSlash(name), maximumFileBytes)
		if err != nil {
			return Report{}, errors.New("source documentation file is unavailable, unsafe, or oversized")
		}
		if len(content) > maximumTotalBytes-total {
			return Report{}, errors.New("source documentation total byte limit exceeded")
		}
		total += len(content)
		module := ""
		for _, directory := range moduleDirectories {
			if directory == "." || strings.HasPrefix(name, directory+"/") {
				if module == "" || (directory != "." && (module == "." || len(directory) > len(module))) {
					module = directory
				}
			}
		}
		if module == "" {
			return Report{}, errors.New("tracked Go file has no tracked module boundary")
		}
		kind := "source"
		if strings.HasSuffix(name, "_test.go") {
			kind = "test"
		}
		if fixture(name) {
			kind = "fixture"
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "source.go", content, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			if kind != "fixture" {
				return Report{}, errors.New("source documentation requires parseable tracked Go files")
			}
			report.Files = append(report.Files, File{Path: safePath(name), Module: safePath(module), Kind: kind})
			report.Findings = append(report.Findings, Finding{safePath(name), 1, "fixture-parse-review", "review", "fixture is not parseable Go source; inspect its intended negative-input boundary manually"})
			continue
		}
		record := File{Path: safePath(name), Module: safePath(module), Package: safePath(file.Name.Name), Kind: kind, Generated: ast.IsGenerated(file), BuildConstrained: bytes.Contains(content, []byte("//go:build ")) || bytes.Contains(content, []byte("// +build "))}
		report.Files = append(report.Files, record)
		parsed = append(parsed, parsedFile{record, name, file, fset})
	}
	packages := map[string][]parsedFile{}
	directories := map[string][]parsedFile{}
	for _, file := range parsed {
		directories[path.Dir(file.relative)] = append(directories[path.Dir(file.relative)], file)
	}
	for _, file := range parsed {
		if err := ctx.Err(); err != nil {
			return Report{}, err
		}
		if err := inspectFile(root, file, directories[path.Dir(file.relative)], &report); err != nil {
			return Report{}, err
		}
		if file.file.Kind == "source" && !file.file.Generated {
			key := path.Dir(file.relative) + "/" + file.file.Package
			packages[key] = append(packages[key], file)
		}
	}
	for _, files := range packages {
		documented := slices.ContainsFunc(files, func(file parsedFile) bool { return validPackageComment(file.ast.Doc, file.file.Package) })
		if !documented {
			addFinding(&report, files[0], files[0].ast.Package, "package-comment", "error", "package needs an idiomatic canonical package comment")
		}
	}
	if len(report.Findings) > maximumFindings {
		return Report{}, errors.New("source documentation finding limit exceeded")
	}
	slices.SortFunc(report.Modules, func(a, b Module) int { return strings.Compare(a.Directory, b.Directory) })
	slices.SortFunc(report.Findings, func(a, b Finding) int {
		if comparison := strings.Compare(a.Path, b.Path); comparison != 0 {
			return comparison
		}
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return strings.Compare(a.Code, b.Code)
	})
	return report, ctx.Err()
}

func addFinding(report *Report, file parsedFile, pos token.Pos, code, severity, message string) {
	// Retain one overflow sentinel at most; the pass fails instead of truncating.
	if len(report.Findings) <= maximumFindings {
		report.Findings = append(report.Findings, Finding{file.file.Path, file.fset.Position(pos).Line, code, severity, message})
	}
}
