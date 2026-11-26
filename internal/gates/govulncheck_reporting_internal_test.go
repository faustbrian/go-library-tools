package gates

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestGovulncheckFailureMetadataNamesAdvisoriesOnly(t *testing.T) {
	report := strings.Join([]string{
		"=== Symbol Results ===",
		"",
		"Vulnerability #1: GO-2026-6859",
		"    HTTP/1 server crash due to MaxBytesReader header race in net/http",
		"  More info: https://pkg.go.dev/vuln/GO-2026-6859",
		"  Standard library",
		"    Found in: net/http@go1.27.3",
		"    Fixed in: net/http@go1.27.4",
		"    Example traces found:",
		"      #1: sequencehttp/handler.go:143:23: sequencehttp.Handler.inspect calls http.Header.Set",
		"",
		"Vulnerability #2: GO-2026-6865",
		"    Found in: github.com/klauspost/compress@v1.20.0",
		"    Fixed in: github.com/klauspost/compress@v1.20.2",
		"",
		"Your code is affected by 2 vulnerabilities.",
	}, "\n")
	got := govulncheckFailureMetadata([]byte(report))
	want := "vulnerability GO-2026-6859 found=net/http@go1.27.3 fixed=net/http@go1.27.4\n" +
		"vulnerability GO-2026-6865 found=github.com/klauspost/compress@v1.20.0 fixed=github.com/klauspost/compress@v1.20.2"
	if got != want {
		t.Fatalf("metadata = %q, want %q", got, want)
	}
	for _, leaked := range []string{"sequencehttp", "MaxBytesReader", "pkg.go.dev", "Standard library"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("metadata leaked scanner text %q: %q", leaked, got)
		}
	}
}

func TestGovulncheckFailureMetadataRejectsUnrecognizedOutput(t *testing.T) {
	for name, report := range map[string]string{
		"empty":                "",
		"tool error":           "govulncheck: loading packages: exit status 1",
		"version without head": "    Found in: net/http@go1.27.3",
		"malformed heading":    "Vulnerability #1: GO-2026-6859 extra",
	} {
		if got := govulncheckFailureMetadata([]byte(report)); got != "govulncheck-tool-or-report-failure" {
			t.Errorf("%s: metadata = %q", name, got)
		}
	}
}

func TestGovulncheckFailureMetadataBoundsFindings(t *testing.T) {
	var report strings.Builder
	for index := 1; index <= maximumVulnerabilityFindings+1; index++ {
		report.WriteString("Vulnerability #1: GO-2026-0001\n")
	}
	got := govulncheckFailureMetadata([]byte(report.String()))
	if strings.Count(got, "vulnerability GO-") != maximumVulnerabilityFindings || !strings.HasSuffix(got, "additional vulnerabilities omitted") {
		t.Fatalf("metadata was not bounded: %q", got)
	}
}

func TestVulnerabilityGateReportsAdvisoriesOnFindings(t *testing.T) {
	root := t.TempDir()
	runner := Runner{Root: root, Executor: executorFunction(func(_ context.Context, command Command) error {
		_, _ = io.WriteString(command.Stdout, "Vulnerability #1: GO-2026-6859\n    Found in: net/http@go1.27.3\n    Fixed in: net/http@go1.27.4\n      #1: secret/path.go:1:1: trace text\n")
		return compilerScannerExit{error: errors.New("exit status 1"), code: 1}
	})}
	err := runner.securityTool(t.Context(), io.Discard, ".", "vulnerability", root,
		"golang.org/x/vuln/cmd/govulncheck@"+govulncheckVersion, "./...")
	if err == nil {
		t.Fatal("vulnerability gate passed with findings")
	}
	message := err.Error()
	if !strings.Contains(message, "vulnerability GO-2026-6859 found=net/http@go1.27.3 fixed=net/http@go1.27.4") {
		t.Fatalf("gate error does not name the advisory: %q", message)
	}
	if strings.Contains(message, "trace text") || strings.Contains(message, "secret/path.go") {
		t.Fatalf("gate error leaked scanner text: %q", message)
	}
}

func TestVulnerabilityGatePassesWithoutFindings(t *testing.T) {
	root := t.TempDir()
	runner := Runner{Root: root, Executor: executorFunction(func(_ context.Context, command Command) error {
		_, _ = io.WriteString(command.Stdout, "No vulnerabilities found.\n")
		return nil
	})}
	if err := runner.securityTool(t.Context(), io.Discard, ".", "vulnerability", root,
		"golang.org/x/vuln/cmd/govulncheck@"+govulncheckVersion, "./..."); err != nil {
		t.Fatalf("vulnerability gate failed without findings: %v", err)
	}
}
