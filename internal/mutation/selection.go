package mutation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// productionSelection is the exact direct-file target admitted by native
// go-list output. It carries names, not permission to waive selected mutants.
type productionSelection struct {
	// Files contains sorted native GoFiles and CgoFiles without test variants.
	Files []string `json:"files"`
	// Excluded contains sorted inactive direct Go filenames, matched exactly.
	Excluded []string `json:"excluded"`
}

// selectProduction admits the primary native go-list target, not its augmented
// test variant, and excludes every other direct non-test Go file from mutation.
func selectProduction(root, module, pkg, modulePath string, listing io.Reader, aliases ...string) (productionSelection, error) {
	packages, err := parseListing(listing)
	if err != nil {
		return productionSelection{}, err
	}
	target := modulePath
	if pkg != "." {
		target += "/" + filepath.ToSlash(pkg)
	}
	directory := filepath.Join(root, filepath.FromSlash(module), filepath.FromSlash(pkg))
	var selected []string
	found := false
	for _, listed := range packages {
		if listed.ImportPath != target || listed.ForTest != "" {
			continue
		}
		admittedDirectory := listed.Dir == directory
		for _, alias := range aliases {
			if listed.Dir == filepath.Join(alias, filepath.FromSlash(module), filepath.FromSlash(pkg)) {
				admittedDirectory = true
			}
		}
		if !admittedDirectory {
			return productionSelection{}, fmt.Errorf("%w: native mutation target is outside selected directory", ErrInvalid)
		}
		actual, err := filepath.EvalSymlinks(listed.Dir)
		if err != nil || actual != directory || listed.Module == nil || listed.Module.Path != modulePath || found {
			return productionSelection{}, fmt.Errorf("%w: conflicting native mutation target %s", ErrInvalid, target)
		}
		found = true
		if len(listed.GoFiles) > mutationSourceLimits().entries || len(listed.CgoFiles) > mutationSourceLimits().entries-len(listed.GoFiles) {
			return productionSelection{}, fmt.Errorf("%w: native mutation filename selection exceeds bounds", ErrInvalid)
		}
		selected = append(selected, listed.GoFiles...)
		selected = append(selected, listed.CgoFiles...)
	}
	if !found || len(selected) == 0 {
		return productionSelection{}, fmt.Errorf("%w: missing native production target %s", ErrInvalid, target)
	}
	entries, err := readSourceEntries(directory, mutationSourceLimits().entries)
	if err != nil {
		return productionSelection{}, err
	}
	available := make(map[string]bool)
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			if strings.ContainsAny(entry.Name(), "\\\x00\r\n") {
				return productionSelection{}, fmt.Errorf("%w: unsafe mutation source filename", ErrInvalid)
			}
			available[entry.Name()] = true
		}
	}
	seen := make(map[string]bool, len(selected))
	for _, name := range selected {
		if !available[name] || filepath.Base(name) != name || strings.ContainsAny(name, "\\\x00\r\n") || seen[name] {
			return productionSelection{}, fmt.Errorf("%w: invalid native mutation source filename", ErrInvalid)
		}
		seen[name] = true
	}
	selection := productionSelection{Files: selected, Excluded: []string{}}
	for name := range available {
		if !seen[name] {
			selection.Excluded = append(selection.Excluded, name)
		}
	}
	sort.Strings(selection.Files)
	sort.Strings(selection.Excluded)
	return selection, nil
}

// exclusion preserves direct-package scope and anchors inactive filenames.
func (selection productionSelection) exclusion() string {
	if len(selection.Excluded) == 0 {
		return "^.+/"
	}
	names := make([]string, len(selection.Excluded))
	for i, name := range selection.Excluded {
		// pflag's StringSlice parses commas as CSV separators. Hex escaping
		// keeps a literal comma inside the single anchored filepath expression.
		names[i] = strings.NewReplacer(",", `\x2c`, `"`, `\x22`).Replace(regexp.QuoteMeta(name))
	}
	return "^.+/|^(?:" + strings.Join(names, "|") + ")$"
}

// selectedInputDigest binds native execution selection to the existing source,
// dependency and policy digest without relabeling historical checkpoint input.
func selectedInputDigest(input string, selection productionSelection, runtime RuntimeIdentity) string {
	data, _ := json.Marshal(struct {
		Input     string              `json:"input"`
		Selection productionSelection `json:"selection"`
		Runtime   RuntimeIdentity     `json:"runtime"`
		Exclusion string              `json:"exclusion"`
	}{input, selection, runtime, selection.exclusion()})
	digest := sha256.Sum256(append([]byte("golib/native-mutation-selection/v1\n"), data...))
	return "sha256:" + hex.EncodeToString(digest[:])
}
