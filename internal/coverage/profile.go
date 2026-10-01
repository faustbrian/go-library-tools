// Package coverage validates Go coverage profiles against canonical production
// package inventory.
package coverage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type totals struct {
	covered int
	total   int
}

type block struct {
	packagePath string
	statements  int
	covered     bool
}

// Verify returns a deterministic package report when every expected package
// has executable statements and exact statement coverage. On incomplete coverage,
// it returns bounded source-block diagnostics alongside the unchanged error.
func Verify(profile io.Reader, expected []string) (string, error) {
	if len(expected) == 0 {
		return "", errors.New("coverage expected packages are empty")
	}
	scanner := bufio.NewScanner(profile)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	if !scanner.Scan() || !strings.HasPrefix(scanner.Text(), "mode: ") {
		return "", errors.New("coverage profile is missing mode header")
	}
	blocks := make(map[string]block)
	line := 1
	for scanner.Scan() {
		line++
		fields := strings.Fields(scanner.Text())
		if len(fields) != 3 {
			return "", fmt.Errorf("invalid coverage profile line %d", line)
		}
		separator := strings.LastIndexByte(fields[0], ':')
		if separator <= 0 {
			return "", fmt.Errorf("invalid coverage profile line %d", line)
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil || statements < 0 {
			return "", fmt.Errorf("invalid statement count on coverage profile line %d", line)
		}
		count, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid execution count on coverage profile line %d", line)
		}
		packagePath := path.Dir(fields[0][:separator])
		identity := fields[0]
		value, exists := blocks[identity]
		if exists && (value.packagePath != packagePath || value.statements != statements) {
			return "", fmt.Errorf("inconsistent duplicate coverage block on line %d", line)
		}
		blocks[identity] = block{
			packagePath: packagePath,
			statements:  statements,
			covered:     value.covered || count > 0,
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read coverage profile: %w", err)
	}
	packages := make(map[string]totals)
	for _, coverageBlock := range blocks {
		value := packages[coverageBlock.packagePath]
		if coverageBlock.statements > int(^uint(0)>>1)-value.total {
			return "", fmt.Errorf("coverage statement total overflows for %s", coverageBlock.packagePath)
		}
		value.total += coverageBlock.statements
		if coverageBlock.covered {
			value.covered += coverageBlock.statements
		}
		packages[coverageBlock.packagePath] = value
	}

	unique := make(map[string]struct{}, len(expected))
	for _, packagePath := range expected {
		unique[packagePath] = struct{}{}
	}
	ordered := make([]string, 0, len(unique))
	for packagePath := range unique {
		ordered = append(ordered, packagePath)
	}
	sort.Strings(ordered)
	var report strings.Builder
	for _, packagePath := range ordered {
		value, exists := packages[packagePath]
		if !exists || value.total == 0 {
			return "", fmt.Errorf("%s missing executable coverage evidence", packagePath)
		}
		_, _ = fmt.Fprintf(&report, "%s %d/%d statements\n", packagePath, value.covered, value.total)
		if value.covered != value.total {
			return uncoveredReport(blocks, packagePath), fmt.Errorf("%s is below exact 100%% coverage", packagePath)
		}
	}
	return report.String(), nil
}

// Limit failure output without changing which blocks determine exact coverage.
const diagnosticBlockLimit = 512
const diagnosticLocationByteLimit = 160

var sourceCoordinates = regexp.MustCompile(`^[0-9]+\.[0-9]+,[0-9]+\.[0-9]+$`)

func uncoveredReport(blocks map[string]block, packagePath string) string {
	locations := make([]string, 0)
	uncovered := 0
	for identity, value := range blocks {
		if value.packagePath != packagePath || value.statements == 0 || value.covered {
			continue
		}
		uncovered++
		separator := strings.LastIndexByte(identity, ':')
		filename, coordinates := identity[:separator], identity[separator+1:]
		if path.IsAbs(filename) || strings.ContainsAny(filename, `\:`) ||
			filename != packagePath+"/"+path.Base(filename) || !strings.HasSuffix(filename, ".go") ||
			!sourceCoordinates.MatchString(coordinates) {
			continue
		}
		location := strconv.QuoteToASCII(path.Base(filename) + ":" + coordinates)
		location = location[1 : len(location)-1]
		if len(location) > diagnosticLocationByteLimit {
			continue
		}
		locations = append(locations, location)
	}
	sort.Strings(locations)
	if len(locations) > diagnosticBlockLimit {
		locations = locations[:diagnosticBlockLimit]
	}
	var report strings.Builder
	report.WriteString("uncovered production blocks:\n")
	for _, location := range locations {
		fmt.Fprintf(&report, "  %s\n", location)
	}
	if omitted := uncovered - len(locations); omitted > 0 {
		fmt.Fprintf(&report, "  (%d additional blocks omitted)\n", omitted)
	}
	return report.String()
}
