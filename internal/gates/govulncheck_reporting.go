package gates

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

const (
	maximumVulnerabilityFindings = 32
	maximumVulnerabilityVersions = 4
)

var (
	vulnerabilityHeading = regexp.MustCompile(`^Vulnerability #[1-9][0-9]{0,3}: (GO-[0-9]{4}-[0-9]{1,6})$`)
	vulnerabilityVersion = regexp.MustCompile(`^ {4}(Found|Fixed) in: ([A-Za-z0-9._~+/@-]{1,256})$`)
)

// govulncheckFailureMetadata names the advisories behind a failed scan. Only
// lines matching govulncheck's fixed heading and version shapes are kept, so
// traces, symbols and other scanner text never reach the gate output.
func govulncheckFailureMetadata(data []byte) string {
	const unusable = "govulncheck-tool-or-report-failure"
	if len(data) == 0 || len(data) > maximumSecurityProcessOutput {
		return unusable
	}
	var findings []string
	current := ""
	versions := 0
	omitted := false
	flush := func() {
		if current != "" {
			if len(findings) < maximumVulnerabilityFindings {
				findings = append(findings, current)
			} else {
				omitted = true
			}
		}
		current = ""
		versions = 0
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 4096), 4096)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Vulnerability #") {
			// An unrecognized heading still ends the previous advisory, so its
			// version lines are never attributed to the wrong advisory.
			flush()
			if match := vulnerabilityHeading.FindStringSubmatch(line); match != nil {
				current = "vulnerability " + match[1]
			}
			continue
		}
		if current == "" {
			continue
		}
		if match := vulnerabilityVersion.FindStringSubmatch(line); match != nil && versions < maximumVulnerabilityVersions {
			current += " " + strings.ToLower(match[1]) + "=" + match[2]
			versions++
		}
	}
	if scanner.Err() != nil {
		return unusable
	}
	flush()
	if len(findings) == 0 {
		return unusable
	}
	result := strings.Join(findings, "\n")
	if omitted {
		result += "\nadditional vulnerabilities omitted"
	}
	return result
}
