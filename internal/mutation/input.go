package mutation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/evidence"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const (
	maximumListSize   = 64 << 20
	maximumInputFile  = 16 << 20
	maximumInputTotal = 256 << 20
	maximumInputFiles = 100000
)

// OwnedModule identifies a repository-local module whose source may observe a
// mutation campaign.
type OwnedModule struct {
	ModulePath string `json:"module_path"`
	Directory  string `json:"directory"`
}

// InputPolicy is the package-level semantic policy bound to mutation evidence.
type InputPolicy struct {
	ModuleDirectory   string            `json:"module_directory"`
	PackageDirectory  string            `json:"package_directory"`
	ModulePath        string            `json:"module_path"`
	GoVersion         string            `json:"go_version"`
	TestTags          []string          `json:"test_tags"`
	BuildTags         []string          `json:"build_tags"`
	RequiredServices  []string          `json:"required_services"`
	ServiceIdentities map[string]string `json:"service_identities"`
	OwnedModules      []OwnedModule     `json:"owned_modules"`
}

type listedPackage struct {
	Dir             string        `json:"Dir"`
	ImportPath      string        `json:"ImportPath"`
	ForTest         string        `json:"ForTest"`
	GoFiles         []string      `json:"GoFiles"`
	CgoFiles        []string      `json:"CgoFiles"`
	CFiles          []string      `json:"CFiles"`
	CXXFiles        []string      `json:"CXXFiles"`
	MFiles          []string      `json:"MFiles"`
	HFiles          []string      `json:"HFiles"`
	FFiles          []string      `json:"FFiles"`
	SFiles          []string      `json:"SFiles"`
	SwigFiles       []string      `json:"SwigFiles"`
	SwigCXXFiles    []string      `json:"SwigCXXFiles"`
	SysoFiles       []string      `json:"SysoFiles"`
	EmbedFiles      []string      `json:"EmbedFiles"`
	TestGoFiles     []string      `json:"TestGoFiles"`
	XTestGoFiles    []string      `json:"XTestGoFiles"`
	TestEmbedFiles  []string      `json:"TestEmbedFiles"`
	XTestEmbedFiles []string      `json:"XTestEmbedFiles"`
	Module          *listedModule `json:"Module"`
}

type listedModule struct {
	Path      string `json:"Path"`
	Version   string `json:"Version"`
	Sum       string `json:"Sum"`
	GoVersion string `json:"GoVersion"`
	Main      bool   `json:"Main"`
}

type moduleIdentity struct {
	Path      string `json:"path"`
	Version   string `json:"version,omitempty"`
	Sum       string `json:"sum,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
	Main      bool   `json:"main,omitempty"`
}

// InputDigest binds a mutation campaign to the exact local source, observing
// tests, fixtures, dependency versions, package policy, and verifier semantics.
func InputDigest(root string, policy InputPolicy, listing io.Reader, review *ZeroReview) (string, error) {
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("%w: repository root must be absolute", ErrInvalid)
	}
	if err := policy.validate(); err != nil {
		return "", err
	}
	if review != nil {
		if err := review.validate(); err != nil {
			return "", fmt.Errorf("%w: zero-mutant review does not match input policy", ErrInvalid)
		}
		actual := [4]string{review.ModuleDirectory, review.PackageDirectory, review.GremlinsVersion, review.GremlinsVerifierSHA256}
		expected := [4]string{policy.ModuleDirectory, policy.PackageDirectory, GremlinsVersion, LegacyVerifierDigest()}
		if actual != expected {
			return "", fmt.Errorf("%w: zero-mutant review does not match input policy", ErrInvalid)
		}
	}
	packages, err := parseListing(listing)
	if err != nil {
		return "", err
	}
	policy = policy.canonical()
	moduleRoots := append([]OwnedModule{{ModulePath: policy.ModulePath, Directory: policy.ModuleDirectory}}, policy.OwnedModules...)
	target := policy.ModulePath
	if policy.PackageDirectory != "." {
		target += "/" + filepath.ToSlash(policy.PackageDirectory)
	}
	content := make(map[string][]byte)
	modules := make(map[string]moduleIdentity)
	observedTarget := false
	relevantDirectories := make(map[string]OwnedModule)
	total := 0
	for _, pkg := range packages {
		if pkg.Module != nil && pkg.Module.Path != "" {
			identity := moduleIdentity{Path: pkg.Module.Path, Version: pkg.Module.Version, Sum: pkg.Module.Sum, GoVersion: pkg.Module.GoVersion, Main: pkg.Module.Main}
			if existing, exists := modules[identity.Path]; exists && existing != identity {
				return "", fmt.Errorf("%w: conflicting module identity for %s", ErrInvalid, identity.Path)
			}
			modules[identity.Path] = identity
		}
		owned, local := resolveOwnedRoot(root, pkg, moduleRoots)
		if !local {
			continue
		}
		canonicalImport, _, _ := strings.Cut(pkg.ImportPath, " [")
		if canonicalImport == target+".test" {
			continue
		}
		observer := canonicalImport == target
		if pkg.ForTest == target {
			observer = true
		}
		if canonicalImport == target {
			observedTarget = true
		}
		files := productionFiles(pkg)
		if observer {
			files = append(files, pkg.TestGoFiles...)
			files = append(files, pkg.XTestGoFiles...)
			files = append(files, pkg.TestEmbedFiles...)
			files = append(files, pkg.XTestEmbedFiles...)
		}
		if len(files) == 0 {
			continue
		}
		relevantDirectories[pkg.Dir] = owned
		for _, name := range files {
			if err := addInputFile(root, owned, pkg.Dir, name, content, &total); err != nil {
				return "", err
			}
		}
	}
	if !observedTarget {
		return "", fmt.Errorf("%w: go list did not resolve target package %s", ErrInvalid, target)
	}
	for directory, owned := range relevantDirectories {
		for _, dataDirectory := range []string{"corpus", "fixtures", "testdata"} {
			if err := addDataDirectory(root, owned, filepath.Join(directory, dataDirectory), content, &total); err != nil {
				return "", err
			}
		}
	}
	if len(content) == 0 {
		return "", fmt.Errorf("%w: mutation input contains no local files", ErrInvalid)
	}
	moduleList := make([]moduleIdentity, 0, len(modules))
	for _, identity := range modules {
		moduleList = append(moduleList, identity)
	}
	slices.SortFunc(moduleList, func(left, right moduleIdentity) int { return strings.Compare(left.Path, right.Path) })
	semantic := struct {
		Policy   InputPolicy      `json:"policy"`
		Modules  []moduleIdentity `json:"modules"`
		Verifier string           `json:"verifier"`
		Review   *ZeroReview      `json:"zero_review,omitempty"`
	}{policy, moduleList, LegacyVerifierDigest(), review}
	encoded, _ := json.Marshal(semantic)
	return evidence.Digest("golib/mutation-input/v1\n"+string(encoded), content), nil
}

func (policy InputPolicy) validate() error {
	identity := [4]bool{policy.ModulePath != "", policy.GoVersion != "", validRelative(policy.ModuleDirectory), validRelative(policy.PackageDirectory)}
	if identity != [4]bool{true, true, true, true} {
		return fmt.Errorf("%w: mutation input policy identity is malformed", ErrInvalid)
	}
	seen := map[string]struct{}{policy.ModulePath: {}}
	services := make(map[string]struct{}, len(policy.RequiredServices))
	for _, service := range policy.RequiredServices {
		identity, identified := policy.ServiceIdentities[service]
		_, duplicate := services[service]
		valid := [5]bool{service != "", identity != "", identified, !duplicate, !strings.ContainsAny(service+identity, "\x00\r\n")}
		if valid != [5]bool{true, true, true, true, true} {
			return fmt.Errorf("%w: required service identity is malformed", ErrInvalid)
		}
		services[service] = struct{}{}
	}
	if len(services) != len(policy.ServiceIdentities) {
		return fmt.Errorf("%w: service identities must exactly match required services", ErrInvalid)
	}
	for _, owned := range policy.OwnedModules {
		if owned.ModulePath == "" {
			return fmt.Errorf("%w: owned module identity is malformed", ErrInvalid)
		}
		if !validRelative(owned.Directory) {
			return fmt.Errorf("%w: owned module identity is malformed", ErrInvalid)
		}
		if _, exists := seen[owned.ModulePath]; exists {
			return fmt.Errorf("%w: duplicate owned module %s", ErrInvalid, owned.ModulePath)
		}
		seen[owned.ModulePath] = struct{}{}
	}
	return nil
}

func (policy InputPolicy) canonical() InputPolicy {
	policy.TestTags = sortedCopy(policy.TestTags)
	policy.BuildTags = sortedCopy(policy.BuildTags)
	policy.RequiredServices = sortedCopy(policy.RequiredServices)
	policy.OwnedModules = append([]OwnedModule(nil), policy.OwnedModules...)
	slices.SortFunc(policy.OwnedModules, func(left, right OwnedModule) int { return strings.Compare(left.ModulePath, right.ModulePath) })
	return policy
}

func sortedCopy(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; !exists {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func parseListing(reader io.Reader) ([]listedPackage, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumListSize+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read go list output: %s", ErrInvalid, err.Error())
	}
	if len(data) > maximumListSize {
		return nil, fmt.Errorf("%w: go list output exceeds %d bytes", ErrInvalid, maximumListSize)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	packages := make([]listedPackage, 0)
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); errors.Is(err, io.EOF) {
			if len(packages) == 0 {
				return nil, fmt.Errorf("%w: go list output is empty", ErrInvalid)
			}
			return packages, nil
		} else if err != nil {
			return nil, fmt.Errorf("%w: decode go list output: %s", ErrInvalid, err.Error())
		}
		packages = append(packages, pkg)
	}
}

func productionFiles(pkg listedPackage) []string {
	var result []string
	for _, files := range [][]string{pkg.GoFiles, pkg.CgoFiles, pkg.CFiles, pkg.CXXFiles, pkg.MFiles, pkg.HFiles, pkg.FFiles, pkg.SFiles, pkg.SwigFiles, pkg.SwigCXXFiles, pkg.SysoFiles, pkg.EmbedFiles} {
		result = append(result, files...)
	}
	return result
}

func resolveOwnedRoot(root string, pkg listedPackage, owned []OwnedModule) (OwnedModule, bool) {
	if pkg.Module == nil {
		return OwnedModule{}, false
	}
	for _, candidate := range owned {
		if candidate.ModulePath == pkg.Module.Path {
			moduleRoot := filepath.Join(root, filepath.FromSlash(candidate.Directory))
			if isWithin(moduleRoot, pkg.Dir) {
				return candidate, true
			}
		}
	}
	return OwnedModule{}, false
}

func isWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	if relative == ".." {
		return false
	}
	return !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func addInputFile(root string, owned OwnedModule, directory, name string, content map[string][]byte, total *int) error {
	absolute := name
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(directory, filepath.FromSlash(name))
	}
	moduleRoot := filepath.Join(root, filepath.FromSlash(owned.Directory))
	if !isWithin(moduleRoot, absolute) {
		return fmt.Errorf("%w: listed file escapes owned module: %s", ErrInvalid, name)
	}
	relative, _ := filepath.Rel(root, absolute)
	data, err := repositoryfile.Read(root, relative, maximumInputFile)
	if err != nil {
		return fmt.Errorf("mutation input %s: %w", filepath.ToSlash(relative), err)
	}
	keyRelative, _ := filepath.Rel(moduleRoot, absolute)
	key := "module:" + owned.ModulePath + "/" + filepath.ToSlash(keyRelative)
	return addContent(content, key, data, total)
}

func addContent(content map[string][]byte, key string, data []byte, total *int) error {
	if _, exists := content[key]; exists {
		return nil
	}
	if len(content) >= maximumInputFiles {
		return fmt.Errorf("%w: mutation input exceeds content bounds", ErrInvalid)
	}
	*total += len(data)
	if *total > maximumInputTotal {
		return fmt.Errorf("%w: mutation input exceeds content bounds", ErrInvalid)
	}
	content[key] = data
	return nil
}

func addDataDirectory(root string, owned OwnedModule, directory string, content map[string][]byte, total *int) error {
	info, err := os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect mutation data directory: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: mutation data path is not a real directory: %s", ErrInvalid, directory)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: mutation data path is not a real directory: %s", ErrInvalid, directory)
	}
	return filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		return visitDataEntry(root, owned, path, entry, walkErr, content, total)
	})
}

func visitDataEntry(root string, owned OwnedModule, path string, entry fs.DirEntry, walkErr error, content map[string][]byte, total *int) error {
	if walkErr != nil {
		return walkErr
	}
	if entry.Type()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: symlink in mutation data: %s", ErrInvalid, path)
	}
	if entry.IsDir() {
		return nil
	}
	if !entry.Type().IsRegular() {
		return fmt.Errorf("%w: non-regular mutation data: %s", ErrInvalid, path)
	}
	return addInputFile(root, owned, filepath.Dir(path), filepath.Base(path), content, total)
}
