// Package coverage validates Go coverage profiles against canonical production
// package inventory.
package coverage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

type totals struct {
	covered int
	total   int
}

// Verify returns a deterministic package report only when every expected
// package has executable statements and exact statement coverage.
func Verify(profile io.Reader, expected []string) (string, error) {
	if len(expected) == 0 {
		return "", errors.New("coverage expected packages are empty")
	}
	scanner := bufio.NewScanner(profile)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	if !scanner.Scan() || !strings.HasPrefix(scanner.Text(), "mode: ") {
		return "", errors.New("coverage profile is missing mode header")
	}
	packages := make(map[string]totals)
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
		value := packages[packagePath]
		value.total += statements
		if count > 0 {
			value.covered += statements
		}
		packages[packagePath] = value
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read coverage profile: %w", err)
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
			return "", fmt.Errorf("%s is below exact 100%% coverage", packagePath)
		}
	}
	return report.String(), nil
}
