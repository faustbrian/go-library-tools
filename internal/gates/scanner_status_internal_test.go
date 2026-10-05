package gates

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

func TestSecretScannerFailureClassificationWithoutOutput(t *testing.T) {
	for _, test := range []struct {
		name           string
		chunks         []string
		tool           string
		cancel         bool
		overflow       bool
		stderrOverflow bool
		succeeded      bool
		findings       bool
	}{
		{name: "history findings", chunks: []string{"private-payload\nexit status 42\n"}, findings: true},
		{name: "split findings", chunks: []string{"private-payload\nex", "it status ", "42", "\n"}, findings: true},
		{name: "empty prefix", chunks: []string{"exit status 42\n"}, findings: true},
		{name: "ordinary failure", chunks: []string{"private-payload\nexit status 1\n"}},
		{name: "nonterminal marker", chunks: []string{"exit status 42\nexit status 1\n"}},
		{name: "embedded marker", chunks: []string{"private-payload exit status 42\n"}},
		{name: "unterminated marker", chunks: []string{"exit status 42"}},
		{name: "other scanner", tool: "example.invalid/scanner@v1", chunks: []string{"exit status 42\n"}},
		{name: "canceled findings", cancel: true, chunks: []string{"exit status 42\n"}},
		{name: "overflowed findings", overflow: true, chunks: []string{"exit status 42\n"}},
		{name: "stderr overflowed findings", stderrOverflow: true, chunks: []string{"exit status 42\n"}},
		{name: "successful command", succeeded: true, chunks: []string{"exit status 42\n"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			tool := test.tool
			if tool == "" {
				tool = "github.com/zricethezav/gitleaks/v8@" + gitleaksVersion
			}
			failure := errors.New("process failure")
			var announced strings.Builder
			runner := Runner{Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
				if test.tool == "" && !slices.Contains(command.Args, "--exit-code=42") {
					t.Fatal("Gitleaks findings still share the scanner failure status")
				}
				for _, chunk := range test.chunks {
					_, _ = io.WriteString(command.Stderr, chunk)
				}
				if test.overflow {
					output, ok := command.Stdout.(*boundedProcessOutput)
					if !ok {
						t.Fatal("stdout is not the bounded scanner output")
					}
					output.limit = 0
					_, _ = io.WriteString(output, "x")
				}
				if test.stderrOverflow {
					output, ok := command.Stderr.(*boundedProcessOutput)
					if !ok {
						t.Fatal("stderr is not the bounded scanner output")
					}
					output.limit = output.written
					_, _ = io.WriteString(output, "x")
				}
				if test.cancel {
					cancel()
				}
				if test.succeeded {
					return nil
				}
				return failure
			}}}
			err := runner.securityTool(ctx, &announced, ".", "secrets-history", ".", tool, "git", ".")
			if test.succeeded {
				if err != nil {
					t.Fatal("successful command became a failure")
				}
			} else if !errors.Is(err, failure) {
				t.Fatal("scanner failure was discarded")
			}
			text := announced.String()
			if err != nil {
				text += err.Error()
			}
			if strings.Contains(text, "private-payload") || strings.Contains(text, "exit status 42") {
				t.Fatal("scanner-controlled text leaked")
			}
			if strings.Contains(text, "secret-findings") != test.findings {
				t.Fatalf("findings classification = %v, want %v", strings.Contains(text, "secret-findings"), test.findings)
			}
		})
	}
}
