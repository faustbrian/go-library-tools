package assets

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faustbrian/go-library-tools/internal/jcs"
)

type Generator struct {
	Repository         string
	SchemaDir          string
	OutputDir          string
	TaskRoot           string
	GoBinary           string
	GitBinary          string
	ModuleProxy        string
	WorkspaceRoot      string
	RunnerTemplateDirs []string
}

type observation struct {
	CaseID             string `json:"case_id"`
	EntryPoint         string `json:"entry_point"`
	InputBase64        string `json:"input_base64"`
	Outcome            string `json:"outcome"`
	DecodedValueKind   string `json:"decoded_value_kind,omitempty"`
	DecodedValueBase64 string `json:"decoded_value_base64,omitempty"`
	EmittedValueKind   string `json:"emitted_value_kind,omitempty"`
	EmittedValueBase64 string `json:"emitted_value_base64,omitempty"`
	ErrorClass         string `json:"error_class,omitempty"`
}

type corpusAccepted struct {
	CaseID             string `json:"case_id"`
	Entrypoint         string `json:"entry_point"`
	Input              string `json:"input_base64"`
	DecodedValueSHA256 string `json:"decoded_value_sha256"`
	EmittedValueSHA256 string `json:"emitted_value_sha256"`
}

type corpusRejected struct {
	CaseID     string `json:"case_id"`
	Entrypoint string `json:"entry_point"`
	Input      string `json:"input_base64"`
	ErrorClass string `json:"error_class"`
}

type exportRow struct {
	Identifier string `json:"identifier"`
	Signature  string `json:"signature"`
}

type baseline struct {
	HistoricalSourceCommit string           `json:"historical_source_commit"`
	LoaderPaths            []string         `json:"loader_paths"`
	ParserPaths            []string         `json:"parser_paths"`
	ExportCount            int              `json:"export_count"`
	Exports                []exportRow      `json:"exports"`
	AcceptedCaseCount      int              `json:"accepted_case_count"`
	AcceptedCases          []corpusAccepted `json:"accepted_cases"`
	RejectedCaseCount      int              `json:"rejected_case_count"`
	RejectedCases          []corpusRejected `json:"rejected_cases"`
	AcceptedCorpusSHA256   string           `json:"accepted_corpus_sha256"`
	RejectedCorpusSHA256   string           `json:"rejected_corpus_sha256"`
}

func (generator Generator) Generate() (digests map[string]string, resultErr error) {
	if generator.Repository == "" || generator.SchemaDir == "" || generator.OutputDir == "" || generator.TaskRoot == "" || generator.GoBinary == "" || generator.GitBinary == "" || generator.ModuleProxy == "" || generator.WorkspaceRoot == "" || len(generator.RunnerTemplateDirs) == 0 {
		return nil, errors.New("generator paths and Go binary are required")
	}
	if err := validateSpecifications(); err != nil {
		return nil, err
	}
	if err := requireWithin(generator.WorkspaceRoot, generator.TaskRoot); err != nil {
		return nil, fmt.Errorf("task confinement: %w", err)
	}
	if err := requireResolvedWithin(generator.WorkspaceRoot, generator.TaskRoot); err != nil {
		return nil, fmt.Errorf("resolved task confinement: %w", err)
	}
	if err := requireWithin(generator.WorkspaceRoot, generator.OutputDir); err != nil {
		return nil, fmt.Errorf("output confinement: %w", err)
	}
	if err := requireDisjointTrees(generator.TaskRoot, generator.OutputDir); err != nil {
		return nil, fmt.Errorf("task/output isolation: %w", err)
	}
	if err := requireEmptyNonSymlinkDirectory(generator.TaskRoot); err != nil {
		return nil, fmt.Errorf("task directory: %w", err)
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedTree(generator.TaskRoot)) }()
	if err := os.MkdirAll(generator.OutputDir, 0o755); err != nil {
		return nil, err
	}
	if err := requireResolvedWithin(generator.WorkspaceRoot, generator.OutputDir); err != nil {
		return nil, fmt.Errorf("resolved output confinement: %w", err)
	}
	if err := requireResolvedDisjointTrees(generator.TaskRoot, generator.OutputDir); err != nil {
		return nil, fmt.Errorf("resolved task/output isolation: %w", err)
	}
	if err := requireEmptyNonSymlinkDirectory(generator.OutputDir); err != nil {
		return nil, fmt.Errorf("output directory: %w", err)
	}
	if err := requireWithin(generator.Repository, generator.SchemaDir); err != nil {
		return nil, fmt.Errorf("schema confinement: %w", err)
	}
	if err := requireResolvedWithin(generator.Repository, generator.SchemaDir); err != nil {
		return nil, fmt.Errorf("resolved schema confinement: %w", err)
	}
	snapshotRoot, inputDigests, err := generator.snapshotInputs()
	if err != nil {
		return nil, fmt.Errorf("snapshot inputs: %w", err)
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedTree(snapshotRoot)) }()
	generator.SchemaDir = filepath.Join(snapshotRoot, "schema")
	generator.RunnerTemplateDirs = []string{filepath.Join(snapshotRoot, "runners")}
	generator.ModuleProxy = filepath.Join(snapshotRoot, "proxy")
	manifestBytes, err := canonicalJSON(map[string]any{"inputs": inputDigests})
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(generator.TaskRoot, "run-manifest.json"), manifestBytes, 0o600); err != nil {
		return nil, err
	}
	defer func() {
		joinCleanup(&resultErr, cleanupOwnedFile(filepath.Join(generator.TaskRoot, "run-manifest.json")))
	}()
	contents := map[string][]byte{}
	for _, spec := range Specs {
		generated, err := generator.generateOne(spec)
		if err != nil {
			return nil, err
		}
		for name, value := range generated {
			if _, exists := contents[name]; exists {
				return nil, fmt.Errorf("duplicate asset %s", name)
			}
			contents[name] = value
		}
	}
	if len(contents) != 20 {
		return nil, fmt.Errorf("generated %d assets, want 20", len(contents))
	}
	names := make([]string, 0, len(contents))
	for name := range contents {
		names = append(names, name)
	}
	slices.Sort(names)
	return writeAssetBundle(generator.OutputDir, names, contents)
}

func validateSpecifications() error {
	if len(Specs) != 5 {
		return fmt.Errorf("specification count is %d, want 5", len(Specs))
	}
	for _, spec := range Specs {
		if len(spec.Commit) != 40 || spec.Base == "" || spec.SchemaPath == "" || spec.Runner == "" || spec.RunnerDirectory == "" || spec.RunnerTest == "" {
			return fmt.Errorf("incomplete specification %q", spec.Base)
		}
		paths := append(append([]string{}, spec.LoaderPaths...), spec.ParserPaths...)
		if len(paths) == 0 {
			return fmt.Errorf("specification %s has no entry points", spec.Base)
		}
		for index := 1; index < len(paths); index++ {
			if paths[index-1] >= paths[index] {
				return fmt.Errorf("specification %s entry points are not sorted unique", spec.Base)
			}
		}
		for _, runner := range spec.RuntimeRunners {
			if runner.Runner == "" || runner.Directory == "" || runner.Test == "" {
				return fmt.Errorf("specification %s has incomplete runtime runner", spec.Base)
			}
		}
	}
	return nil
}

func (generator Generator) snapshotInputs() (snapshotRoot string, digests map[string]string, resultErr error) {
	root := filepath.Join(generator.TaskRoot, "input-snapshot")
	defer func() {
		if resultErr != nil {
			joinCleanup(&resultErr, cleanupOwnedTree(root))
		}
	}()
	if err := os.Mkdir(root, 0o700); err != nil {
		return "", nil, err
	}
	if err := os.Mkdir(filepath.Join(root, "schema"), 0o700); err != nil {
		return "", nil, err
	}
	if err := os.Mkdir(filepath.Join(root, "runners"), 0o700); err != nil {
		return "", nil, err
	}
	proxyDigests, err := snapshotMinimalProxy(generator.ModuleProxy, filepath.Join(root, "proxy"))
	if err != nil {
		return "", nil, err
	}
	digests = map[string]string{}
	for path, value := range proxyDigests {
		digests["proxy/"+path] = value
	}
	entries, err := os.ReadDir(generator.SchemaDir)
	if err != nil {
		return "", nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}
		source := filepath.Join(generator.SchemaDir, entry.Name())
		data, err := readRegularNonSymlink(source, maximumSchemaBytes)
		if err != nil {
			return "", nil, err
		}
		if err := writeExclusiveFile(filepath.Join(root, "schema", entry.Name()), data); err != nil {
			return "", nil, err
		}
		digests["schema/"+entry.Name()] = digest(data)
	}
	for _, spec := range Specs {
		source, err := generator.runnerPath(spec.Runner)
		if err != nil {
			return "", nil, err
		}
		data, err := readRegularNonSymlink(source, maximumSchemaBytes)
		if err != nil {
			return "", nil, err
		}
		if err := writeExclusiveFile(filepath.Join(root, "runners", spec.Runner), data); err != nil {
			return "", nil, err
		}
		digests["runner/"+spec.Runner] = digest(data)
		for _, runtimeRunner := range spec.RuntimeRunners {
			source, err := generator.runnerPath(runtimeRunner.Runner)
			if err != nil {
				return "", nil, err
			}
			data, err := readRegularNonSymlink(source, maximumSchemaBytes)
			if err != nil {
				return "", nil, err
			}
			if err := writeExclusiveFile(filepath.Join(root, "runners", runtimeRunner.Runner), data); err != nil {
				return "", nil, err
			}
			digests["runner/"+runtimeRunner.Runner] = digest(data)
		}
	}
	if err := freezeSnapshot(root); err != nil {
		return "", nil, err
	}
	return root, digests, nil
}

func freezeSnapshot(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("snapshot contains symlink: %s", path)
		}
		if entry.IsDir() {
			return os.Chmod(path, 0o500)
		}
		return os.Chmod(path, 0o400)
	})
}

func cleanupOwnedTree(root string) error {
	return cleanupOwnedTreeWithRemover(root, os.RemoveAll)
}

func joinCleanup(resultErr *error, cleanupErr error) {
	*resultErr = errors.Join(*resultErr, cleanupErr)
}

func cleanupOwnedTreeWithRemover(root string, removeAll func(string) error) error {
	info, err := os.Lstat(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("cleanup root is not a non-symlink directory: %s", root)
	}
	if err := os.Chmod(root, 0o700); err != nil {
		return err
	}
	walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			return os.Chmod(path, 0o700)
		}
		return os.Chmod(path, 0o600)
	})
	removeErr := removeAll(root)
	_, absenceErr := os.Lstat(root)
	if errors.Is(absenceErr, os.ErrNotExist) {
		absenceErr = nil
	} else if absenceErr == nil {
		absenceErr = errors.New("cleanup target remains after removal")
	}
	return errors.Join(walkErr, removeErr, absenceErr)
}

func cleanupOwnedFile(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	_, absenceErr := os.Lstat(path)
	if errors.Is(absenceErr, os.ErrNotExist) {
		absenceErr = nil
	} else if absenceErr == nil {
		absenceErr = errors.New("cleanup file remains after removal")
	}
	return errors.Join(err, absenceErr)
}

type proxyModule struct {
	path, version, zipHash, modHash string
}

var allowedProxyModules = []proxyModule{
	{"github.com/santhosh-tekuri/jsonschema/v6", "v6.0.3", "h1:1EYB5IzjZawrrnELUi78f9fPu57HuXjmddZPjrls/28=", "h1:JXeL+ps8p7/KNMjDQk3TCwPpBy0wYklyWTfbkIzdIFU="},
	{"github.com/yuin/goldmark", "v1.8.5", "h1:r6N5afV5qj/5S4UTch8agZHJ8UxNCMwX7WjkkJam2NA=", "h1:ip/1k0VRfGynBgxOz0yCqHrbZXhcjxyuS66Brc7iBKg="},
	{"go.yaml.in/yaml/v3", "v3.0.4", "h1:tfq32ie2Jv2UxXFdLJdh3jXuOzWiL1fo0bu/FbuKpbc=", "h1:DhzuOOF2ATzADvBadXxruRBLzYTpT36CKvDb3+aBEFg="},
	{"golang.org/x/mod", "v0.40.0", "h1:hUv+3cXcdRHz08UmSiOob7sadHig73uo5bkXxQ/tvUs=", "h1:0/weTWkPWGBikyTWAX3dkjVztMmBA5hM0DH6BElSupE="},
	{"golang.org/x/text", "v0.14.0", "h1:ScX5w1eTa3QqT8oi6+ziP7dTV1S2+ALU0bI+0zXKWiQ=", "h1:18ZOQIKpY8NJVqYksKHtTdi31H5itFRjB5/qKTNYzSU="},
}

func snapshotMinimalProxy(source, destination string) (map[string]string, error) {
	return snapshotProxyModules(source, destination, allowedProxyModules)
}

func snapshotProxyModules(source, destination string, modules []proxyModule) (map[string]string, error) {
	info, err := os.Lstat(source)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("module proxy must be a non-symlink directory")
	}
	if err := os.Mkdir(destination, 0o700); err != nil {
		return nil, err
	}
	const maximumProxyFiles, maximumProxyBytes = 32, int64(256 << 20)
	digests := map[string]string{}
	var total int64
	for _, module := range modules {
		if module.path == "" || module.version == "" || module.zipHash == "" || module.modHash == "" {
			return nil, errors.New("incomplete proxy module allowlist entry")
		}
		for _, suffix := range []string{".info", ".mod", ".zip", ".ziphash"} {
			relative := filepath.ToSlash(filepath.Join(module.path, "@v", module.version+suffix))
			sourcePath := filepath.Join(source, filepath.FromSlash(relative))
			if err := requireResolvedWithin(source, sourcePath); err != nil {
				return nil, fmt.Errorf("proxy allowlist %s confinement: %w", relative, err)
			}
			data, err := readRegularNonSymlink(sourcePath, maximumProxyBytes)
			if err != nil {
				return nil, fmt.Errorf("proxy allowlist %s: %w", relative, err)
			}
			total += int64(len(data))
			if len(digests) == maximumProxyFiles || total > maximumProxyBytes {
				return nil, errors.New("minimal proxy exceeds aggregate bounds")
			}
			switch suffix {
			case ".ziphash":
				if strings.TrimSpace(string(data)) != module.zipHash {
					return nil, fmt.Errorf("proxy zip hash mismatch for %s %s", module.path, module.version)
				}
			case ".zip":
				got, err := hashModuleZip(data)
				if err != nil {
					return nil, fmt.Errorf("proxy zip invalid for %s %s: %w", module.path, module.version, err)
				}
				if got != module.zipHash {
					return nil, fmt.Errorf("proxy zip content hash mismatch for %s %s", module.path, module.version)
				}
			case ".mod":
				if !strings.HasPrefix(string(data), "module "+module.path+"\n") {
					return nil, fmt.Errorf("proxy go.mod identity mismatch for %s %s", module.path, module.version)
				}
				if got := hashModuleFiles([]moduleFile{{name: "go.mod", data: data}}); got != module.modHash {
					return nil, fmt.Errorf("proxy go.mod content hash mismatch for %s %s", module.path, module.version)
				}
			case ".info":
				var metadata struct{ Version string }
				if err := json.Unmarshal(data, &metadata); err != nil || metadata.Version != module.version {
					return nil, fmt.Errorf("proxy info identity mismatch for %s %s", module.path, module.version)
				}
			}
			target := filepath.Join(destination, filepath.FromSlash(relative))
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return nil, err
			}
			if err := writeExclusiveFile(target, data); err != nil {
				return nil, err
			}
			digests[relative] = digest(data)
		}
	}
	return digests, nil
}

type moduleFile struct {
	name string
	data []byte
}

func hashModuleZip(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	files := make([]moduleFile, 0, len(reader.File))
	seen := map[string]struct{}{}
	var total int64
	for _, entry := range reader.File {
		if strings.HasSuffix(entry.Name, "/") {
			continue
		}
		if entry.Name == "" || strings.Contains(entry.Name, "\n") || strings.HasPrefix(entry.Name, "/") || strings.Contains(entry.Name, "\\") {
			return "", fmt.Errorf("unsafe module zip path %q", entry.Name)
		}
		if _, duplicate := seen[entry.Name]; duplicate {
			return "", fmt.Errorf("duplicate module zip path %q", entry.Name)
		}
		if len(files) == 65536 {
			return "", errors.New("module zip exceeds 65536 files")
		}
		if entry.UncompressedSize64 > 256<<20 || total > int64(256<<20)-int64(entry.UncompressedSize64) {
			return "", errors.New("module zip exceeds 256 MiB expanded bytes")
		}
		total += int64(entry.UncompressedSize64)
		seen[entry.Name] = struct{}{}
		file, err := entry.Open()
		if err != nil {
			return "", err
		}
		content, readErr := io.ReadAll(io.LimitReader(file, int64(maximumSchemaBytes)+1))
		closeErr := file.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			return "", err
		}
		if len(content) > maximumSchemaBytes {
			return "", fmt.Errorf("module zip member exceeds %d bytes", maximumSchemaBytes)
		}
		files = append(files, moduleFile{name: entry.Name, data: content})
	}
	return hashModuleFiles(files), nil
}

// hashModuleFiles implements the h1 directory hash used by Go module sums.
func hashModuleFiles(files []moduleFile) string {
	slices.SortFunc(files, func(left, right moduleFile) int { return strings.Compare(left.name, right.name) })
	outer := sha256.New()
	for _, file := range files {
		inner := sha256.Sum256(file.data)
		fmt.Fprintf(outer, "%x  %s\n", inner, file.name)
	}
	return "h1:" + base64.StdEncoding.EncodeToString(outer.Sum(nil))
}

func readRegularNonSymlink(path string, maximum int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("not a regular non-symlink file: %s", path)
	}
	if info.Size() > maximum {
		return nil, fmt.Errorf("file exceeds %d bytes: %s", maximum, path)
	}
	return os.ReadFile(path)
}

func writeExclusiveFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	resultErr := errors.Join(writeErr, closeErr)
	if resultErr != nil {
		joinCleanup(&resultErr, cleanupOwnedFile(path))
	}
	return resultErr
}

func (generator Generator) generateOne(spec Spec) (generated map[string][]byte, resultErr error) {
	checkout := filepath.Join(generator.TaskRoot, "checkout-"+spec.Base)
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		return nil, errors.Join(err, cleanupOwnedTree(checkout))
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedTree(checkout)) }()
	if err := archiveCommit(generator.GitBinary, generator.Repository, spec.Commit, checkout); err != nil {
		return nil, fmt.Errorf("archive %s: %w", spec.Base, err)
	}
	if err := validateHistoricalRequirements(checkout); err != nil {
		return nil, fmt.Errorf("historical requirements %s: %w", spec.Base, err)
	}
	if spec.OriginalPath != "" {
		data, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(spec.OriginalPath)))
		if err != nil {
			return nil, fmt.Errorf("historical schema %s: %w", spec.Base, err)
		}
		if got := digest(data); got != spec.OriginalSHA256 {
			return nil, fmt.Errorf("historical schema %s digest = %s, want %s", spec.Base, got, spec.OriginalSHA256)
		}
	}
	exports, err := generator.extractExports(checkout, spec.Base)
	if err != nil {
		return nil, fmt.Errorf("exports %s: %w", spec.Base, err)
	}
	if err := generator.runRuntimeCharacterizations(checkout, spec); err != nil {
		return nil, fmt.Errorf("runtime characterization %s: %w", spec.Base, err)
	}
	observed, err := generator.runHistorical(checkout, spec)
	if err != nil {
		return nil, fmt.Errorf("historical runner %s: %w", spec.Base, err)
	}
	accepted, rejected, err := buildCorpora(spec, observed)
	if err != nil {
		return nil, fmt.Errorf("corpora %s: %w", spec.Base, err)
	}
	acceptedBytes, err := canonicalJSON(accepted)
	if err != nil {
		return nil, err
	}
	rejectedBytes, err := canonicalJSON(rejected)
	if err != nil {
		return nil, err
	}
	base := baseline{HistoricalSourceCommit: spec.Commit, LoaderPaths: pathArray(spec.LoaderPaths), ParserPaths: pathArray(spec.ParserPaths), ExportCount: len(exports), Exports: exports, AcceptedCaseCount: len(accepted), AcceptedCases: accepted, RejectedCaseCount: len(rejected), RejectedCases: rejected, AcceptedCorpusSHA256: digest(acceptedBytes), RejectedCorpusSHA256: digest(rejectedBytes)}
	baselineBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, err
	}
	graphBytes, err := generator.verifiedReferenceGraph(checkout, spec)
	if err != nil {
		return nil, fmt.Errorf("reference graph %s: %w", spec.Base, err)
	}
	contents := [][]byte{baselineBytes, acceptedBytes, rejectedBytes, graphBytes}
	result := map[string][]byte{}
	for index, name := range AssetNames(spec) {
		result[name] = contents[index]
	}
	return result, nil
}

func pathArray(paths []string) []string {
	if len(paths) == 0 {
		return []string{}
	}
	return append([]string(nil), paths...)
}

func validateHistoricalRequirements(checkout string) error {
	data, err := os.ReadFile(filepath.Join(checkout, "go.mod"))
	if err != nil {
		return err
	}
	sum, err := os.ReadFile(filepath.Join(checkout, "go.sum"))
	if err != nil {
		return err
	}
	allowed := map[string]proxyModule{}
	for _, module := range allowedProxyModules {
		allowed[module.path] = module
	}
	requirements := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.SplitN(line, "//", 2)[0])
		var path, version string
		if len(fields) == 2 && strings.HasPrefix(fields[1], "v") {
			path, version = fields[0], fields[1]
		}
		if len(fields) == 3 && fields[0] == "require" {
			path, version = fields[1], fields[2]
		}
		if path == "" {
			continue
		}
		module, exists := allowed[path]
		if !exists || module.version != version {
			return fmt.Errorf("unallowlisted module requirement %s %s", path, version)
		}
		requirements[path] = version
		if !containsExactLine(sum, path+" "+version+" "+module.zipHash) || !containsExactLine(sum, path+" "+version+"/go.mod "+module.modHash) {
			return fmt.Errorf("go.sum does not bind %s %s", path, version)
		}
	}
	if len(requirements) == 0 {
		return errors.New("historical module has no allowlisted requirements")
	}
	return nil
}

func containsExactLine(data []byte, line string) bool {
	for _, candidate := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if candidate == line {
			return true
		}
	}
	return false
}

func (generator Generator) verifiedReferenceGraph(checkout string, spec Spec) (graph []byte, resultErr error) {
	candidate, err := resolvedReferenceGraph(generator.SchemaDir, spec.SchemaPath)
	if err != nil {
		return nil, err
	}
	if spec.OriginalPath == "" {
		return candidate, nil
	}
	differentialRoot := filepath.Join(generator.TaskRoot, "graph-differential-"+spec.Base)
	if err := os.MkdirAll(filepath.Join(differentialRoot, "schema"), 0o700); err != nil {
		return nil, errors.Join(err, cleanupOwnedTree(differentialRoot))
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedTree(differentialRoot)) }()
	rewrites := map[string]string{filepath.Base(spec.OriginalPath): filepath.Base(spec.SchemaPath)}
	if spec.Base == "cohesion-catalog-v1" {
		rewrites["modules.schema.json"] = "modules-v2.schema.json"
	}
	for oldName, newName := range rewrites {
		historical, err := os.ReadFile(filepath.Join(checkout, "schema", oldName))
		if err != nil {
			return nil, err
		}
		transformed, err := rewriteHistoricalSchema(historical, rewrites)
		if err != nil {
			return nil, err
		}
		candidateBytes, err := os.ReadFile(filepath.Join(generator.SchemaDir, newName))
		if err != nil {
			return nil, err
		}
		candidateCanonical, err := jcs.Canonicalize(candidateBytes)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(transformed, candidateCanonical) {
			return nil, fmt.Errorf("candidate %s differs beyond authorized schema filename and $id rewrites", newName)
		}
		if err := os.WriteFile(filepath.Join(differentialRoot, "schema", newName), transformed, 0o600); err != nil {
			return nil, err
		}
	}
	expected, err := resolvedReferenceGraph(filepath.Join(differentialRoot, "schema"), spec.SchemaPath)
	if err != nil {
		return nil, err
	}
	expectedShape, err := referenceGraphShape(expected)
	if err != nil {
		return nil, err
	}
	candidateShape, err := referenceGraphShape(candidate)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(expectedShape, candidateShape) {
		return nil, fmt.Errorf("candidate %s resolved graph differs from authorized historical rewrite", spec.SchemaPath)
	}
	return candidate, nil
}

func rewriteHistoricalSchema(data []byte, rewrites map[string]string) ([]byte, error) {
	value, err := decodeUniqueJSON(data)
	if err != nil {
		return nil, err
	}
	root, ok := value.(jcs.Object)
	if !ok {
		return nil, errors.New("historical schema root is not an object")
	}
	identityValue, exists := objectValue(root, "$id")
	identity, ok := identityValue.(string)
	if !exists || !ok {
		return nil, errors.New("historical schema has no string root $id")
	}
	changedIdentity := false
	oldNames := make([]string, 0, len(rewrites))
	for oldName := range rewrites {
		oldNames = append(oldNames, oldName)
	}
	slices.Sort(oldNames)
	for _, oldName := range oldNames {
		newName := rewrites[oldName]
		if strings.HasSuffix(identity, "/"+oldName) {
			identity = strings.TrimSuffix(identity, oldName) + newName
			changedIdentity = true
			break
		}
	}
	if !changedIdentity {
		return nil, errors.New("historical root $id has no authorized rewrite")
	}
	setObjectValue(root, "$id", identity)
	var walk func(any) error
	walk = func(current any) error {
		switch typed := current.(type) {
		case jcs.Object:
			for index := range typed {
				if typed[index].Name == "$ref" {
					ref, ok := typed[index].Value.(string)
					if !ok {
						return errors.New("$ref is not a string")
					}
					typed[index].Value = rewriteReferenceFilename(ref, oldNames, rewrites)
					continue
				}
				if err := walk(typed[index].Value); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range typed {
				if err := walk(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(root); err != nil {
		return nil, err
	}
	return jcs.CanonicalizeValue(root)
}

func rewriteReferenceFilename(reference string, oldNames []string, rewrites map[string]string) string {
	boundary := len(reference)
	if index := strings.IndexAny(reference, "?#"); index >= 0 {
		boundary = index
	}
	path, suffix := reference[:boundary], reference[boundary:]
	for _, oldName := range oldNames {
		if path == oldName || strings.HasSuffix(path, "/"+oldName) {
			return strings.TrimSuffix(path, oldName) + rewrites[oldName] + suffix
		}
	}
	return reference
}

func referenceGraphShape(data []byte) ([]byte, error) {
	value, err := jcs.Decode(data)
	if err != nil {
		return nil, err
	}
	nodeValues, ok := value.([]any)
	if !ok {
		return nil, errors.New("reference graph root is not an array")
	}
	for nodeIndex, nodeValue := range nodeValues {
		node, ok := nodeValue.(jcs.Object)
		if !ok {
			return nil, errors.New("reference graph node is not an object")
		}
		if _, err := requiredObjectString(node, "schema_path"); err != nil {
			return nil, err
		}
		referenceValue, exists := objectValue(node, "references")
		referenceValues, ok := referenceValue.([]any)
		if !exists || !ok {
			return nil, errors.New("reference graph references is not an array")
		}
		for refIndex, refValue := range referenceValues {
			ref, ok := refValue.(jcs.Object)
			if !ok {
				return nil, errors.New("reference graph reference is not an object")
			}
			if _, err := requiredObjectString(ref, "ref"); err != nil {
				return nil, err
			}
			if _, err := requiredObjectString(ref, "resolved_schema_path"); err != nil {
				return nil, err
			}
			referenceValues[refIndex] = withoutObjectMember(ref, "resolved_schema_bytes_sha256")
		}
		setObjectValue(node, "references", referenceValues)
		nodeValues[nodeIndex] = withoutObjectMember(node, "schema_bytes_sha256")
	}
	return jcs.CanonicalizeValue(nodeValues)
}

func withoutObjectMember(object jcs.Object, name string) jcs.Object {
	result := object[:0]
	for _, member := range object {
		if member.Name != name {
			result = append(result, member)
		}
	}
	return result
}

func requireEmptyNonSymlinkDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("must be a non-symlink directory")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return errors.New("must be empty")
	}
	return nil
}

func writeAssetBundle(output string, names []string, contents map[string][]byte) (digests map[string]string, resultErr error) {
	created := make([]string, 0, len(names))
	defer func() {
		if resultErr == nil {
			return
		}
		for _, path := range created {
			if err := cleanupOwnedFile(path); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("clean partial asset %s: %w", path, err))
			}
		}
	}()
	digests = make(map[string]string, len(names))
	for _, name := range names {
		if filepath.Base(name) != name || name == "." || name == ".." {
			return nil, fmt.Errorf("unsafe asset name %q", name)
		}
		path := filepath.Join(output, name)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return nil, err
		}
		created = append(created, path)
		_, writeErr := file.Write(contents[name])
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			return nil, errors.Join(writeErr, closeErr)
		}
		digests[name] = digest(contents[name])
	}
	return digests, nil
}

func archiveCommit(gitBinary, repository, commit, destination string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	identity := exec.CommandContext(ctx, gitBinary, "-C", repository, "rev-parse", commit+"^{commit}")
	identity.Env = []string{"PATH=" + filepath.Dir(gitBinary), "LANG=C.UTF-8", "LC_ALL=C.UTF-8"}
	resolved, err := identity.Output()
	if err != nil {
		return fmt.Errorf("resolve commit: %w", err)
	}
	if strings.TrimSpace(string(resolved)) != commit {
		return fmt.Errorf("resolved commit %s does not equal frozen %s", strings.TrimSpace(string(resolved)), commit)
	}
	command := exec.CommandContext(ctx, gitBinary, "-C", repository, "archive", "--format=tar", commit)
	command.Env = identity.Env
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return err
	}
	extractErr := extractTarArchive(stdout, destination, commit)
	if extractErr != nil {
		cancel()
	}
	waitErr := command.Wait()
	if extractErr != nil {
		return extractErr
	}
	if waitErr != nil {
		return fmt.Errorf("git archive: %w: %s", waitErr, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func extractTarArchive(input io.Reader, destination, commit string) error {
	reader := tar.NewReader(input)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			if header.Name != "pax_global_header" || len(header.PAXRecords) != 1 || header.PAXRecords["comment"] != commit {
				return errors.New("git archive contains unexpected global PAX metadata")
			}
			continue
		}
		clean := filepath.Clean(header.Name)
		if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unsafe archive path %q", header.Name)
		}
		path := filepath.Join(destination, clean)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.FileMode(header.Mode)&0o777)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, reader)
			closeErr := file.Close()
			if copyErr != nil || closeErr != nil {
				return errors.Join(copyErr, closeErr)
			}
		default:
			return fmt.Errorf("unsupported archive entry %q with type %d", header.Name, header.Typeflag)
		}
	}
	return nil
}

type listedPackage struct {
	ImportPath string
	Export     string
	Module     *struct {
		Path string
		Main bool
	}
}

func (generator Generator) extractExports(root, cacheName string) (exports []exportRow, resultErr error) {
	cacheRoot := filepath.Join(generator.TaskRoot, "types-cache-"+cacheName)
	if err := makeCacheRoots(cacheRoot); err != nil {
		return nil, errors.Join(err, cleanupOwnedTree(cacheRoot))
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedTree(cacheRoot)) }()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, generator.GoBinary, "list", "-deps", "-export", "-json", "./...")
	command.Dir = root
	command.Env = boundedGoEnvironment(cacheRoot, generator.ModuleProxy, generator.GoBinary)
	output, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, fmt.Errorf("go list: %w: %s", err, strings.TrimSpace(string(exit.Stderr)))
		}
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	all := map[string]listedPackage{}
	modulePath := ""
	for {
		var listed listedPackage
		if err := decoder.Decode(&listed); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		all[listed.ImportPath] = listed
		if listed.Module != nil && listed.Module.Main {
			modulePath = listed.Module.Path
		}
	}
	if modulePath == "" {
		return nil, errors.New("go list did not identify the historical module")
	}
	lookup := func(path string) (io.ReadCloser, error) {
		listed, exists := all[path]
		if !exists || listed.Export == "" {
			return nil, fmt.Errorf("no export data for %s", path)
		}
		return os.Open(listed.Export)
	}
	importer := typesImporter(token.NewFileSet(), lookup)
	rows := []exportRow{}
	paths := make([]string, 0)
	for path, listed := range all {
		if listed.Module != nil && listed.Module.Path == modulePath && (path == modulePath || strings.HasPrefix(path, modulePath+"/")) {
			paths = append(paths, path)
		}
	}
	slices.Sort(paths)
	for _, path := range paths {
		pkg, err := importer.Import(path)
		if err != nil {
			return nil, fmt.Errorf("import %s: %w", path, err)
		}
		rows = append(rows, packageExports(pkg)...)
	}
	slices.SortFunc(rows, func(left, right exportRow) int {
		if left.Identifier != right.Identifier {
			return strings.Compare(left.Identifier, right.Identifier)
		}
		return strings.Compare(left.Signature, right.Signature)
	})
	return slices.CompactFunc(rows, func(left, right exportRow) bool { return left == right }), nil
}

type packageImporter interface {
	Import(string) (*types.Package, error)
}

func typesImporter(set *token.FileSet, lookup func(string) (io.ReadCloser, error)) packageImporter {
	return importer.ForCompiler(set, "gc", lookup)
}

func packageExports(pkg *types.Package) []exportRow {
	qualifier := func(other *types.Package) string {
		if other == nil {
			return ""
		}
		return other.Path()
	}
	rows := []exportRow{}
	for _, name := range pkg.Scope().Names() {
		object := pkg.Scope().Lookup(name)
		if !object.Exported() {
			continue
		}
		signature := types.ObjectString(object, qualifier)
		if constant, ok := object.(*types.Const); ok {
			signature += " = " + constant.Val().ExactString()
		}
		rows = append(rows, exportRow{Identifier: pkg.Path() + "." + name, Signature: signature})
		named, ok := types.Unalias(object.Type()).(*types.Named)
		if !ok {
			continue
		}
		if structure, ok := named.Underlying().(*types.Struct); ok {
			for index := range structure.NumFields() {
				field := structure.Field(index)
				if field.Exported() {
					signature := types.ObjectString(field, qualifier)
					if tag := structure.Tag(index); tag != "" {
						signature += " tag " + strconv.Quote(tag)
					}
					rows = append(rows, exportRow{Identifier: pkg.Path() + "." + name + "." + field.Name(), Signature: signature})
				}
			}
		}
		if contract, ok := named.Underlying().(*types.Interface); ok {
			contract.Complete()
			for index := range contract.NumMethods() {
				method := contract.Method(index)
				if method.Exported() {
					rows = append(rows, exportRow{Identifier: pkg.Path() + "." + name + "." + method.Name(), Signature: types.ObjectString(method, qualifier)})
				}
			}
		}
		for _, typ := range []types.Type{named, types.NewPointer(named)} {
			methods := types.NewMethodSet(typ)
			for index := range methods.Len() {
				method := methods.At(index).Obj()
				if method.Exported() {
					rows = append(rows, exportRow{Identifier: pkg.Path() + "." + name + "." + method.Name(), Signature: types.ObjectString(method, qualifier)})
				}
			}
		}
	}
	return rows
}

func (generator Generator) runHistorical(checkout string, spec Spec) (observations []observation, resultErr error) {
	runnerPath, err := generator.runnerPath(spec.Runner)
	if err != nil {
		return nil, err
	}
	runner, err := os.ReadFile(runnerPath)
	if err != nil {
		return nil, fmt.Errorf("read runner %s: %w", spec.Runner, err)
	}
	directory := filepath.Join(checkout, filepath.FromSlash(spec.RunnerDirectory))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	target := filepath.Join(directory, "historical_asset_test.go")
	if err := writeExclusiveFile(target, runner); err != nil {
		return nil, err
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedFile(target)) }()
	resultPath := filepath.Join(generator.TaskRoot, spec.Base+"-observations.json")
	defer func() { joinCleanup(&resultErr, cleanupOwnedFile(resultPath)) }()
	cacheRoot := filepath.Join(generator.TaskRoot, "cache-"+spec.Base)
	if err := makeCacheRoots(cacheRoot); err != nil {
		return nil, errors.Join(err, cleanupOwnedTree(cacheRoot))
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedTree(cacheRoot)) }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, generator.GoBinary, "test", "./"+spec.RunnerDirectory, "-run", "^"+spec.RunnerTest+"$", "-count=1")
	command.Dir = checkout
	command.Env = append(boundedGoEnvironment(cacheRoot, generator.ModuleProxy, generator.GoBinary), "HISTORICAL_OBSERVATION_PATH="+resultPath, "HISTORICAL_ASSET_BASE="+spec.Base)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go test: %w: %s", err, strings.TrimSpace(string(output)))
	}
	data, err := readRegularNonSymlink(resultPath, maximumSchemaBytes)
	if err != nil {
		return nil, err
	}
	return decodeObservations(data)
}

func decodeObservations(data []byte) ([]observation, error) {
	value, err := jcs.Decode(data)
	if err != nil {
		return nil, err
	}
	rows, ok := value.([]any)
	if !ok {
		return nil, errors.New("historical observation corpus is not an array")
	}
	result := make([]observation, len(rows))
	allowed := []string{"case_id", "decoded_value_base64", "decoded_value_kind", "emitted_value_base64", "emitted_value_kind", "entry_point", "error_class", "input_base64", "outcome"}
	for index, rowValue := range rows {
		row, ok := rowValue.(jcs.Object)
		if !ok {
			return nil, fmt.Errorf("historical observation %d is not an object", index)
		}
		for _, member := range row {
			if !slices.Contains(allowed, member.Name) {
				return nil, fmt.Errorf("historical observation %d has unknown member %q", index, member.Name)
			}
			if _, ok := member.Value.(string); !ok {
				return nil, fmt.Errorf("historical observation %d member %q is not a string", index, member.Name)
			}
		}
		var requiredErr error
		result[index].CaseID, requiredErr = requiredObjectString(row, "case_id")
		if requiredErr != nil {
			return nil, requiredErr
		}
		result[index].EntryPoint, requiredErr = requiredObjectString(row, "entry_point")
		if requiredErr != nil {
			return nil, requiredErr
		}
		result[index].InputBase64, requiredErr = requiredObjectString(row, "input_base64")
		if requiredErr != nil {
			return nil, requiredErr
		}
		result[index].Outcome, requiredErr = requiredObjectString(row, "outcome")
		if requiredErr != nil {
			return nil, requiredErr
		}
		result[index].DecodedValueKind = optionalObjectString(row, "decoded_value_kind")
		result[index].DecodedValueBase64 = optionalObjectString(row, "decoded_value_base64")
		result[index].EmittedValueKind = optionalObjectString(row, "emitted_value_kind")
		result[index].EmittedValueBase64 = optionalObjectString(row, "emitted_value_base64")
		result[index].ErrorClass = optionalObjectString(row, "error_class")
	}
	return result, nil
}

func optionalObjectString(object jcs.Object, name string) string {
	value, _ := objectValue(object, name)
	text, _ := value.(string)
	return text
}

func (generator Generator) runRuntimeCharacterizations(checkout string, spec Spec) error {
	for index, runtimeRunner := range spec.RuntimeRunners {
		if err := generator.runRuntimeCharacterization(checkout, spec, runtimeRunner, index); err != nil {
			return err
		}
	}
	return nil
}

func (generator Generator) runRuntimeCharacterization(checkout string, spec Spec, runtimeRunner RuntimeRunner, index int) (resultErr error) {
	runnerPath, err := generator.runnerPath(runtimeRunner.Runner)
	if err != nil {
		return err
	}
	runner, err := readRegularNonSymlink(runnerPath, maximumSchemaBytes)
	if err != nil {
		return err
	}
	directory := filepath.Join(checkout, filepath.FromSlash(runtimeRunner.Directory))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	target := filepath.Join(directory, fmt.Sprintf("historical_runtime_asset_%d_test.go", index))
	if err := writeExclusiveFile(target, runner); err != nil {
		return err
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedFile(target)) }()
	cacheRoot := filepath.Join(generator.TaskRoot, fmt.Sprintf("runtime-cache-%s-%d", spec.Base, index))
	if err := makeCacheRoots(cacheRoot); err != nil {
		return errors.Join(err, cleanupOwnedTree(cacheRoot))
	}
	defer func() { joinCleanup(&resultErr, cleanupOwnedTree(cacheRoot)) }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	command := exec.CommandContext(ctx, generator.GoBinary, "test", "./"+runtimeRunner.Directory, "-run", "^"+runtimeRunner.Test+"$", "-count=1")
	command.Dir = checkout
	command.Env = boundedGoEnvironment(cacheRoot, generator.ModuleProxy, generator.GoBinary)
	output, commandErr := command.CombinedOutput()
	cancel()
	if commandErr != nil {
		return fmt.Errorf("go test %s: %w: %s", runtimeRunner.Test, commandErr, strings.TrimSpace(string(output)))
	}
	return nil
}

func (generator Generator) runnerPath(name string) (string, error) {
	var matches []string
	for _, root := range generator.RunnerTemplateDirs {
		candidate := filepath.Join(root, name)
		info, err := os.Lstat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("runner %s is not a regular non-symlink file", candidate)
		}
		matches = append(matches, candidate)
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("runner %s matched %d template roots, want exactly one", name, len(matches))
	}
	return matches[0], nil
}

func makeCacheRoots(root string) error {
	for _, child := range []string{"go-build", "go-mod", "gopath", "tmp"} {
		if err := os.MkdirAll(filepath.Join(root, child), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func boundedGoEnvironment(cacheRoot, moduleProxy, goBinary string) []string {
	proxyURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(moduleProxy)}).String()
	environment := []string{
		"PATH=" + filepath.Dir(goBinary),
		"LANG=C.UTF-8", "LC_ALL=C.UTF-8", "GOTOOLCHAIN=local", "GOWORK=off", "GOENV=off", "CGO_ENABLED=0",
		"GOPROXY=" + proxyURL, "GONOPROXY=none", "GOSUMDB=off", "GONOSUMDB=none", "GOVCS=*:off", "GOAUTH=off", "GOFLAGS=",
		"GOPATH=" + filepath.Join(cacheRoot, "gopath"), "TMPDIR=" + filepath.Join(cacheRoot, "tmp"),
		"GOCACHE=" + filepath.Join(cacheRoot, "go-build"),
		"GOMODCACHE=" + filepath.Join(cacheRoot, "go-mod"),
		"GOTMPDIR=" + filepath.Join(cacheRoot, "tmp"),
	}
	return environment
}

func buildCorpora(spec Spec, observed []observation) ([]corpusAccepted, []corpusRejected, error) {
	accepted := make([]corpusAccepted, 0, len(observed))
	rejected := make([]corpusRejected, 0, len(observed))
	allowed := append(append([]string{}, spec.LoaderPaths...), spec.ParserPaths...)
	seen := map[string]struct{}{}
	type outcomeCount struct{ accepted, rejected int }
	coverage := make(map[string]outcomeCount, len(allowed))
	for _, row := range observed {
		if row.CaseID == "" {
			return nil, nil, errors.New("empty case ID")
		}
		if _, exists := seen[row.CaseID]; exists {
			return nil, nil, fmt.Errorf("duplicate case ID %q", row.CaseID)
		}
		seen[row.CaseID] = struct{}{}
		if !slices.Contains(allowed, row.EntryPoint) {
			return nil, nil, fmt.Errorf("case %s uses undeclared entry point %q", row.CaseID, row.EntryPoint)
		}
		decodedInput, err := base64.StdEncoding.DecodeString(row.InputBase64)
		if err != nil {
			return nil, nil, fmt.Errorf("accepted %s input: %w", row.CaseID, err)
		}
		if base64.StdEncoding.EncodeToString(decodedInput) != row.InputBase64 {
			return nil, nil, fmt.Errorf("case %s input is not canonical padded RFC 4648 base64", row.CaseID)
		}
		if row.Outcome == "rejected" {
			if row.ErrorClass == "" || row.DecodedValueBase64 != "" || row.DecodedValueKind != "" || row.EmittedValueBase64 != "" || row.EmittedValueKind != "" {
				return nil, nil, fmt.Errorf("rejected %s has inconsistent fields", row.CaseID)
			}
			if strings.HasPrefix(row.ErrorClass, "unexpected") || strings.HasPrefix(row.ErrorClass, "other:") {
				return nil, nil, fmt.Errorf("rejected %s has unstable error class %q", row.CaseID, row.ErrorClass)
			}
			rejected = append(rejected, corpusRejected{row.CaseID, row.EntryPoint, row.InputBase64, row.ErrorClass})
			count := coverage[row.EntryPoint]
			count.rejected++
			coverage[row.EntryPoint] = count
			continue
		}
		if row.Outcome != "accepted" || row.ErrorClass != "" {
			return nil, nil, fmt.Errorf("case %s has invalid outcome %q", row.CaseID, row.Outcome)
		}
		decodedKind, emittedKind := row.DecodedValueKind, row.EmittedValueKind
		if decodedKind == "" {
			decodedKind = spec.DefaultValueKind
		}
		if emittedKind == "" {
			emittedKind = spec.DefaultValueKind
		}
		decodedDigest, err := observedValueDigest(decodedKind, row.DecodedValueBase64)
		if err != nil {
			return nil, nil, fmt.Errorf("accepted %s decoded value: %w", row.CaseID, err)
		}
		emittedDigest, err := observedValueDigest(emittedKind, row.EmittedValueBase64)
		if err != nil {
			return nil, nil, fmt.Errorf("accepted %s emitted value: %w", row.CaseID, err)
		}
		accepted = append(accepted, corpusAccepted{row.CaseID, row.EntryPoint, row.InputBase64, decodedDigest, emittedDigest})
		count := coverage[row.EntryPoint]
		count.accepted++
		coverage[row.EntryPoint] = count
	}
	if len(accepted) == 0 || len(rejected) == 0 {
		return nil, nil, errors.New("corpus must contain accepted and rejected cases")
	}
	for _, entryPoint := range allowed {
		count := coverage[entryPoint]
		if count.accepted == 0 || count.rejected == 0 {
			return nil, nil, fmt.Errorf("entry point %q coverage accepted=%d rejected=%d, want both", entryPoint, count.accepted, count.rejected)
		}
	}
	slices.SortFunc(accepted, func(left, right corpusAccepted) int { return strings.Compare(left.CaseID, right.CaseID) })
	slices.SortFunc(rejected, func(left, right corpusRejected) int { return strings.Compare(left.CaseID, right.CaseID) })
	return accepted, rejected, nil
}

func observedValueDigest(kind, encoded string) (string, error) {
	value, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if base64.StdEncoding.EncodeToString(value) != encoded {
		return "", errors.New("value is not canonical padded RFC 4648 base64")
	}
	var semantic []byte
	switch kind {
	case "json":
		semantic, err = jcs.CanonicalizeHistorical(value)
	default:
		return "", fmt.Errorf("unsupported value kind %q", kind)
	}
	if err != nil {
		return "", err
	}
	return digest(semantic), nil
}

func canonicalJSON(value any) ([]byte, error) {
	return canonicalJSONWithCharge(value, func(int) error { return nil })
}

func canonicalJSONWithCharge(value any, charge func(int) error) ([]byte, error) {
	return jcs.MarshalWithCharge(value, charge)
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

type referenceGraphNode struct {
	SchemaPath        string              `json:"schema_path"`
	SchemaBytesSHA256 string              `json:"schema_bytes_sha256"`
	References        []referenceGraphRef `json:"references"`
}

type referenceGraphRef struct {
	Ref                       string `json:"ref"`
	ResolvedSchemaPath        string `json:"resolved_schema_path"`
	ResolvedSchemaBytesSHA256 string `json:"resolved_schema_bytes_sha256"`
}

const (
	maximumSchemaBytes           = 32 << 20
	maximumGraphSchemaBytes      = 64 << 20
	maximumGraphAllocationBytes  = 2*maximumSchemaBytes + (8 << 20)
	maximumGraphNodes            = 4096
	maximumGraphUniqueReferences = 4096
)

type graphBudget struct {
	schemaBytes     int64
	allocationBytes int64
	schemaLimit     int64
	allocationLimit int64
}

func (budget *graphBudget) chargeSchema(size int64) error {
	limit := budget.schemaLimit
	if limit == 0 {
		limit = maximumGraphSchemaBytes
	}
	if size < 0 || budget.schemaBytes > limit-size {
		return errors.New("resolved graph exceeds cumulative schema byte budget")
	}
	budget.schemaBytes += size
	return budget.chargeAllocation(size)
}

func (budget *graphBudget) chargeAllocation(size int64) error {
	limit := budget.allocationLimit
	if limit == 0 {
		limit = maximumGraphAllocationBytes
	}
	if size < 0 || budget.allocationBytes > limit-size {
		return errors.New("resolved graph exceeds cumulative allocation budget")
	}
	budget.allocationBytes += size
	return nil
}

type schemaDocument struct {
	bytes    []byte
	identity string
	path     string
	value    any
}

type graphResolver struct {
	repositoryRoot string
	schemaRoot     string
	loaded         []*schemaDocument
	budget         graphBudget
}

func resolvedReferenceGraph(schemaRoot, entryPath string) ([]byte, error) {
	return resolvedReferenceGraphWithBudget(schemaRoot, entryPath, &graphBudget{})
}

func resolvedReferenceGraphWithBudget(schemaRoot, entryPath string, budget *graphBudget) ([]byte, error) {
	root, err := filepath.Abs(filepath.Dir(schemaRoot))
	if err != nil {
		return nil, err
	}
	resolver := &graphResolver{repositoryRoot: root, schemaRoot: filepath.Clean(schemaRoot), budget: *budget}
	pointerBytes := int64(strconv.IntSize / 8)
	if err := resolver.budget.chargeAllocation(int64(maximumGraphNodes) * pointerBytes); err != nil {
		return nil, err
	}
	resolver.loaded = make([]*schemaDocument, 0, maximumGraphNodes)
	entry, err := resolver.load(entryPath)
	if err != nil {
		return nil, err
	}
	if err := resolver.budget.chargeAllocation(int64(maximumGraphNodes * 16)); err != nil {
		return nil, err
	}
	pending := make([]string, 0, maximumGraphNodes)
	pending = append(pending, entry.path)
	if err := resolver.budget.chargeAllocation(int64(maximumGraphNodes * 16)); err != nil {
		return nil, err
	}
	scheduled := make([]string, 0, maximumGraphNodes)
	scheduled = append(scheduled, entry.path)
	if err := resolver.budget.chargeAllocation(int64(maximumGraphNodes) * int64(reflect.TypeOf(referenceGraphNode{}).Size())); err != nil {
		return nil, err
	}
	nodes := make([]referenceGraphNode, 0, maximumGraphNodes)
	for len(pending) != 0 {
		path := pending[0]
		pending = pending[1:]
		document, err := resolver.load(path)
		if err != nil {
			return nil, err
		}
		refs, err := collectReferences(document.value, &resolver.budget)
		if err != nil {
			return nil, fmt.Errorf("inspect references in %s: %w", path, err)
		}
		if err := resolver.budget.chargeAllocation(int64(len(refs) * 24)); err != nil {
			return nil, err
		}
		resolved := make([]referenceGraphRef, 0, len(refs))
		for _, ref := range refs {
			target, err := resolver.resolve(document, ref)
			if err != nil {
				return nil, fmt.Errorf("resolve %q from %s: %w", ref, path, err)
			}
			if err := resolver.budget.chargeAllocation(71); err != nil {
				return nil, err
			}
			resolved = append(resolved, referenceGraphRef{ref, target.path, digest(target.bytes)})
			if !slices.Contains(scheduled, target.path) {
				scheduled = append(scheduled, target.path)
				pending = append(pending, target.path)
			}
		}
		slices.SortFunc(resolved, func(left, right referenceGraphRef) int {
			if left.Ref != right.Ref {
				return strings.Compare(left.Ref, right.Ref)
			}
			if left.ResolvedSchemaPath != right.ResolvedSchemaPath {
				return strings.Compare(left.ResolvedSchemaPath, right.ResolvedSchemaPath)
			}
			return strings.Compare(left.ResolvedSchemaBytesSHA256, right.ResolvedSchemaBytesSHA256)
		})
		resolved = slices.CompactFunc(resolved, func(left, right referenceGraphRef) bool { return left == right })
		if err := resolver.budget.chargeAllocation(71); err != nil {
			return nil, err
		}
		nodes = append(nodes, referenceGraphNode{path, digest(document.bytes), resolved})
	}
	slices.SortFunc(nodes, func(left, right referenceGraphNode) int { return strings.Compare(left.SchemaPath, right.SchemaPath) })
	result, err := canonicalReferenceGraph(nodes, &resolver.budget)
	*budget = resolver.budget
	return result, err
}

func canonicalReferenceGraph(nodes []referenceGraphNode, budget *graphBudget) ([]byte, error) {
	if err := budget.chargeAllocation(int64(len(nodes) * 16)); err != nil {
		return nil, err
	}
	values := make([]any, len(nodes))
	memberBytes := int64(reflect.TypeOf(jcs.Member{}).Size())
	for nodeIndex, node := range nodes {
		if err := budget.chargeAllocation(3 * memberBytes); err != nil {
			return nil, err
		}
		if err := budget.chargeAllocation(int64(len(node.References) * 16)); err != nil {
			return nil, err
		}
		references := make([]any, len(node.References))
		for refIndex, reference := range node.References {
			if err := budget.chargeAllocation(3 * memberBytes); err != nil {
				return nil, err
			}
			references[refIndex] = jcs.Object{
				{Name: "ref", Value: reference.Ref},
				{Name: "resolved_schema_path", Value: reference.ResolvedSchemaPath},
				{Name: "resolved_schema_bytes_sha256", Value: reference.ResolvedSchemaBytesSHA256},
			}
		}
		values[nodeIndex] = jcs.Object{
			{Name: "schema_path", Value: node.SchemaPath},
			{Name: "schema_bytes_sha256", Value: node.SchemaBytesSHA256},
			{Name: "references", Value: references},
		}
	}
	return jcs.CanonicalizeValueWithCharge(values, func(size int) error {
		return budget.chargeAllocation(int64(size))
	})
}

func (resolver *graphResolver) resolve(source *schemaDocument, rawReference string) (*schemaDocument, error) {
	reference, err := url.Parse(rawReference)
	if err != nil {
		return nil, err
	}
	if reference.User != nil || reference.RawQuery != "" {
		return nil, errors.New("userinfo and queries are forbidden")
	}
	base, err := url.Parse(source.identity)
	if err != nil {
		return nil, err
	}
	resolvedURI := base.ResolveReference(reference)
	if resolvedURI.Scheme != "https" || resolvedURI.Host != "github.com" {
		return nil, fmt.Errorf("non-local schema identity %q", resolvedURI.String())
	}
	const identityPrefix = "/faustbrian/go-library-tools/schema/"
	if !strings.HasPrefix(resolvedURI.Path, identityPrefix) {
		return nil, fmt.Errorf("schema identity escapes repository: %q", resolvedURI.String())
	}
	filename, err := url.PathUnescape(strings.TrimPrefix(resolvedURI.Path, identityPrefix))
	if err != nil {
		return nil, err
	}
	if filename == "" || filepath.Base(filename) != filename {
		return nil, fmt.Errorf("unsafe schema filename %q", filename)
	}
	target, err := resolver.load(filepath.Join("schema", filename))
	if err != nil {
		return nil, err
	}
	if err := verifyFragment(target.value, resolvedURI.EscapedFragment()); err != nil {
		return nil, err
	}
	return target, nil
}

func (resolver *graphResolver) load(relativePath string) (*schemaDocument, error) {
	clean := filepath.Clean(filepath.FromSlash(relativePath))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("unsafe schema path %q", relativePath)
	}
	path := filepath.Join(resolver.repositoryRoot, clean)
	within, err := filepath.Rel(resolver.schemaRoot, path)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("schema path is outside schema directory: %q", relativePath)
	}
	canonicalPath := filepath.ToSlash(clean)
	if !strings.HasSuffix(canonicalPath, ".schema.json") {
		return nil, fmt.Errorf("schema path lacks .schema.json suffix: %q", relativePath)
	}
	for _, existing := range resolver.loaded {
		if existing.path == canonicalPath {
			return existing, nil
		}
	}
	if len(resolver.loaded) == maximumGraphNodes {
		return nil, fmt.Errorf("resolved graph exceeds %d schema nodes", maximumGraphNodes)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("schema is not a regular non-symlink file: %s", canonicalPath)
	}
	if info.Size() > maximumSchemaBytes {
		return nil, fmt.Errorf("schema exceeds %d bytes: %s", maximumSchemaBytes, canonicalPath)
	}
	if err := resolver.budget.chargeSchema(info.Size()); err != nil {
		return nil, fmt.Errorf("schema %s: %w", canonicalPath, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, errors.New("schema is not valid UTF-8")
	}
	value, err := jcs.DecodeWithCharge(data, func(size int) error {
		return resolver.budget.chargeAllocation(int64(size))
	})
	if err != nil {
		return nil, err
	}
	object, ok := value.(jcs.Object)
	if !ok {
		return nil, errors.New("schema root is not an object")
	}
	identityValue, exists := objectValue(object, "$id")
	identity, ok := identityValue.(string)
	if !exists || !ok || identity == "" {
		return nil, errors.New("schema has no string $id")
	}
	wantIdentity := "https://github.com/faustbrian/go-library-tools/" + canonicalPath
	if identity != wantIdentity {
		return nil, fmt.Errorf("schema $id %q does not equal path identity %q", identity, wantIdentity)
	}
	if err := resolver.budget.chargeAllocation(int64(reflect.TypeOf(schemaDocument{}).Size())); err != nil {
		return nil, err
	}
	document := &schemaDocument{bytes: data, identity: identity, path: canonicalPath, value: value}
	resolver.loaded = append(resolver.loaded, document)
	return document, nil
}

func collectReferences(value any, budget *graphBudget) ([]string, error) {
	refs := []string{}
	var walk func(any) error
	walk = func(current any) error {
		switch typed := current.(type) {
		case jcs.Object:
			for _, member := range typed {
				if member.Name == "$ref" {
					ref, ok := member.Value.(string)
					if !ok {
						return errors.New("$ref value is not a string")
					}
					if !slices.Contains(refs, ref) {
						if len(refs) == maximumGraphUniqueReferences {
							return fmt.Errorf("schema contains more than %d unique references", maximumGraphUniqueReferences)
						}
						if len(refs) == cap(refs) {
							capacity := nextGraphCapacity(cap(refs), len(refs)+1, maximumGraphUniqueReferences)
							if err := budget.chargeAllocation(int64(capacity * 16)); err != nil {
								return err
							}
							grown := make([]string, len(refs), capacity)
							copy(grown, refs)
							refs = grown
						}
						refs = append(refs, ref)
					}
				}
				if err := walk(member.Value); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range typed {
				if err := walk(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(value); err != nil {
		return nil, err
	}
	slices.Sort(refs)
	return refs, nil
}

func nextGraphCapacity(current, needed, maximum int) int {
	capacity := current
	if capacity == 0 {
		capacity = 8
	}
	for capacity < needed {
		capacity *= 2
	}
	if capacity > maximum {
		return maximum
	}
	return capacity
}

func decodeUniqueJSON(data []byte) (any, error) {
	return jcs.Decode(data)
}

func verifyFragment(document any, fragment string) error {
	decodedFragment, err := url.PathUnescape(fragment)
	if err != nil {
		return err
	}
	if decodedFragment == "" {
		return nil
	}
	if !strings.HasPrefix(decodedFragment, "/") {
		return fmt.Errorf("unsupported non-pointer fragment %q", decodedFragment)
	}
	current := document
	for _, token := range strings.Split(strings.TrimPrefix(decodedFragment, "/"), "/") {
		for index := 0; index < len(token); index++ {
			if token[index] == '~' && (index+1 == len(token) || (token[index+1] != '0' && token[index+1] != '1')) {
				return fmt.Errorf("invalid JSON Pointer escape in fragment %q", decodedFragment)
			}
		}
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch container := current.(type) {
		case jcs.Object:
			next, exists := objectValue(container, token)
			if !exists {
				return fmt.Errorf("fragment %q does not exist", decodedFragment)
			}
			current = next
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(container) || (len(token) > 1 && token[0] == '0') {
				return fmt.Errorf("fragment %q has invalid array index", decodedFragment)
			}
			current = container[index]
		default:
			return fmt.Errorf("fragment %q traverses a scalar", decodedFragment)
		}
	}
	return nil
}

func objectValue(object jcs.Object, name string) (any, bool) {
	for _, member := range object {
		if member.Name == name {
			return member.Value, true
		}
	}
	return nil, false
}

func setObjectValue(object jcs.Object, name string, value any) bool {
	for index := range object {
		if object[index].Name == name {
			object[index].Value = value
			return true
		}
	}
	return false
}

func requiredObjectString(object jcs.Object, name string) (string, error) {
	value, exists := objectValue(object, name)
	text, ok := value.(string)
	if !exists || !ok {
		return "", fmt.Errorf("required string member %q is absent", name)
	}
	return text, nil
}

func requireWithin(root, target string) error {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return fmt.Errorf("path %s escapes root %s", target, root)
	}
	return nil
}

func requireDisjointTrees(left, right string) error {
	leftAbsolute, err := filepath.Abs(left)
	if err != nil {
		return err
	}
	rightAbsolute, err := filepath.Abs(right)
	if err != nil {
		return err
	}
	if pathContains(leftAbsolute, rightAbsolute) || pathContains(rightAbsolute, leftAbsolute) {
		return errors.New("task and output directories overlap")
	}
	leftResolved, leftErr := filepath.EvalSymlinks(leftAbsolute)
	rightResolved, rightErr := filepath.EvalSymlinks(rightAbsolute)
	if leftErr == nil && rightErr == nil && (pathContains(leftResolved, rightResolved) || pathContains(rightResolved, leftResolved)) {
		return errors.New("resolved task and output directories overlap")
	}
	return nil
}

func requireResolvedDisjointTrees(left, right string) error {
	leftResolved, err := filepath.EvalSymlinks(left)
	if err != nil {
		return err
	}
	rightResolved, err := filepath.EvalSymlinks(right)
	if err != nil {
		return err
	}
	if pathContains(leftResolved, rightResolved) || pathContains(rightResolved, leftResolved) {
		return errors.New("resolved task and output directories overlap")
	}
	return nil
}

func pathContains(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	return err == nil && (relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative))
}

func requireResolvedWithin(root, target string) error {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return err
	}
	return requireWithin(resolvedRoot, resolvedTarget)
}
