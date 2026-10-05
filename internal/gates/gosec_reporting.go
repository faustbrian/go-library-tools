package gates

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// Only this projection of the pinned scanner's private report may leave the
// process boundary. Details, source, loading messages and path names never do.
func gosecFailureMetadata(data []byte, directory string) string {
	const unusable = "gosec-tool-or-report-failure"
	if len(data) == 0 || len(data) > maximumSecurityProcessOutput || !uniqueGosecJSON(data) {
		return unusable
	}
	var report struct {
		Errors json.RawMessage `json:"Golang errors"`
		Issues json.RawMessage `json:"Issues"`
		Stats  *struct {
			Files *int `json:"files"`
			Lines *int `json:"lines"`
			Nosec *int `json:"nosec"`
			Found *int `json:"found"`
		} `json:"Stats"`
	}
	if json.Unmarshal(data, &report) != nil || report.Errors == nil || report.Issues == nil || report.Stats == nil {
		return unusable
	}
	for _, value := range []*int{report.Stats.Files, report.Stats.Lines, report.Stats.Nosec, report.Stats.Found} {
		if value == nil || *value < 0 || *value > 100000000 {
			return unusable
		}
	}
	var loading map[string][]struct {
		Line   int     `json:"line"`
		Column int     `json:"column"`
		Error  *string `json:"error"`
	}
	var issues []struct {
		Rule         string          `json:"rule_id"`
		File         string          `json:"file"`
		Line         string          `json:"line"`
		Column       string          `json:"column"`
		Nosec        *bool           `json:"nosec"`
		Suppressions json.RawMessage `json:"suppressions"`
	}
	if json.Unmarshal(report.Errors, &loading) != nil || json.Unmarshal(report.Issues, &issues) != nil || len(loading) > 4096 || len(issues) > 4096 {
		return unusable
	}
	loadingCount := 0
	for _, entries := range loading {
		if len(entries) == 0 || len(entries) > 4096-loadingCount {
			return unusable
		}
		for _, entry := range entries {
			if entry.Error == nil || entry.Line < 0 || entry.Line > 100000000 || entry.Column < 0 || entry.Column > 100000000 {
				return unusable
			}
		}
		loadingCount += len(entries)
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return unusable
	}
	findings := 0
	var locations strings.Builder
	for _, issue := range issues {
		var suppressions []json.RawMessage
		if issue.Nosec == nil || issue.Suppressions == nil || json.Unmarshal(issue.Suppressions, &suppressions) != nil || *issue.Nosec || len(suppressions) != 0 {
			// Strict flags must not admit suppressed results into this projection.
			return unusable
		}
		if !pinnedGosecRule(issue.Rule) || !gosecCoordinate(issue.Line, true) || !gosecCoordinate(issue.Column, false) || issue.File == "" {
			return unusable
		}
		path := issue.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
			return unusable
		}
		findings++
		if findings <= 64 {
			fmt.Fprintf(&locations, "\ngosec-location %x %s %s %s", sha256.Sum256([]byte(filepath.ToSlash(relative))), issue.Rule, issue.Line, issue.Column)
		}
	}
	if findings != *report.Stats.Found || (loadingCount == 0 && (*report.Stats.Files == 0 || findings == 0)) {
		return unusable
	}
	class := "gosec-findings"
	if loadingCount > 0 {
		class = "gosec-loading-failure"
	}
	return fmt.Sprintf("%s findings=%d loading_packages=%d loading_errors=%d%s", class, findings, len(loading), loadingCount, locations.String())
}

// v2.29.0 rules/rulelist.go and analyzers/analyzerslist.go own this vocabulary.
// A tool upgrade must reconcile it before new rule identities are disclosed.
func pinnedGosecRule(rule string) bool {
	switch rule {
	case "G101", "G102", "G103", "G104", "G106", "G107", "G108", "G109", "G110", "G111", "G112", "G113", "G114", "G115", "G116", "G117", "G118", "G119", "G120", "G121", "G122", "G123", "G124",
		"G201", "G202", "G203", "G204", "G301", "G302", "G303", "G304", "G305", "G306", "G307",
		"G401", "G402", "G403", "G404", "G405", "G406", "G407", "G408", "G501", "G502", "G503", "G504", "G505", "G506", "G507", "G601", "G602",
		"G701", "G702", "G703", "G704", "G705", "G706", "G707", "G708", "G709", "G710":
		return true
	}
	return false
}

func decimalGosec(value string) bool {
	if value == "" {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func gosecCoordinate(value string, line bool) bool {
	parts := strings.Split(value, "-")
	if len(parts) > 2 || (!line && len(parts) != 1) {
		return false
	}
	previous := 0
	for _, part := range parts {
		if !decimalGosec(part) || len(part) > 9 {
			return false
		}
		number, err := strconv.Atoi(part)
		if err != nil || number < 1 || number > 100000000 || number < previous {
			return false
		}
		previous = number
	}
	return true
}

// Reject ambiguous duplicate keys, excessive nesting and trailing documents
// before decoding: a later duplicate must not hide loading errors or findings.
func uniqueGosecJSON(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var walk func(int) bool
	walk = func(depth int) bool {
		if depth > 64 {
			return false
		}
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delimiter, composite := token.(json.Delim)
		if !composite {
			return true
		}
		if delimiter != '{' && delimiter != '[' {
			return false
		}
		keys := map[string]bool{}
		for decoder.More() {
			if delimiter == '{' {
				key, err := decoder.Token()
				name, ok := key.(string)
				name = strings.ToLower(name)
				if err != nil || !ok || keys[name] {
					return false
				}
				keys[name] = true
			}
			if !walk(depth + 1) {
				return false
			}
		}
		end, err := decoder.Token()
		return err == nil && ((delimiter == '{' && end == json.Delim('}')) || (delimiter == '[' && end == json.Delim(']')))
	}
	if !walk(0) {
		return false
	}
	_, err := decoder.Token()
	return err == io.EOF
}
