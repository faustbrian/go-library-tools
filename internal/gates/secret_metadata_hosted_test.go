package gates

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSecretScannerMetadataHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("scanner metadata verification runs only in hosted CI")
	}
	location := strings.Repeat("a", 64) + " " + strings.Repeat("b", 64) + " 12 " + strings.Repeat("c", 40)
	for _, test := range []struct {
		name     string
		chunks   []string
		status   string
		cancel   bool
		overflow bool
		want     int
		omitted  bool
	}{
		{name: "history", chunks: []string{location + "\n"}, want: 1},
		{name: "split", chunks: []string{location[:20], location[20:] + "\n"}, want: 1},
		{name: "current tree", chunks: []string{strings.Repeat("a", 64) + " " + strings.Repeat("b", 64) + " 1 -\n"}, want: 1},
		{name: "untrusted text", chunks: []string{"private-payload\n"}},
		{name: "incomplete", chunks: []string{location}},
		{name: "invalid later line", chunks: []string{location + "\nprivate-payload\n"}},
		{name: "oversized line", chunks: []string{strings.Repeat("x", 257) + "\n"}},
		{name: "zero line", chunks: []string{strings.Repeat("a", 64) + " " + strings.Repeat("b", 64) + " 0 -\n"}},
		{name: "negative line", chunks: []string{strings.Repeat("a", 64) + " " + strings.Repeat("b", 64) + " -1 -\n"}},
		{name: "extra field", chunks: []string{location + " private-payload\n"}},
		{name: "partial scan", chunks: []string{location + "\n"}, status: "exit status 1\n"},
		{name: "cancel", chunks: []string{location + "\n"}, cancel: true},
		{name: "overflow", chunks: []string{location + "\n"}, overflow: true},
		{name: "location cap", chunks: []string{strings.Repeat(location+"\n", 33)}, want: 32, omitted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			workspace := t.TempDir()
			failure := errors.New("process failure")
			runner := Runner{Executor: workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
				for _, chunk := range test.chunks {
					_, _ = io.WriteString(command.Stdout, chunk)
				}
				status := test.status
				if status == "" {
					status = "exit status 42\n"
				}
				_, _ = io.WriteString(command.Stderr, "private-payload\n"+status)
				if test.cancel {
					cancel()
				}
				if test.overflow {
					output, ok := command.Stdout.(*boundedProcessOutput)
					if !ok {
						t.Fatal("stdout is not the bounded scanner output")
					}
					output.limit = output.written
					_, _ = io.WriteString(output, "x")
				}
				return failure
			}}}
			err := runner.securityTool(ctx, io.Discard, ".", "secrets-history", ".", "github.com/zricethezav/gitleaks/v8@"+gitleaksVersion, "git", ".")
			if !errors.Is(err, failure) {
				t.Fatal("scanner failure was discarded")
			}
			if strings.Count(err.Error(), "secret-location ") != test.want {
				t.Fatalf("missing safe finding location: got %d, want %d", strings.Count(err.Error(), "secret-location "), test.want)
			}
			if strings.Contains(err.Error(), "additional secret locations omitted") != test.omitted {
				t.Fatal("truncated metadata presented as complete")
			}
			if strings.Contains(err.Error(), "private-payload") {
				t.Fatal("scanner text leaked")
			}
			files, readErr := os.ReadDir(workspace)
			if readErr != nil || len(files) != 0 {
				t.Fatal("metadata template was not cleaned up")
			}
		})
	}
}
