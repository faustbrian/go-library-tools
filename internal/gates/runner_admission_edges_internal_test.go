package gates

import (
	"bytes"
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestTerminalMarkerExtraCharacterDoesNotMatch(t *testing.T) {
	for _, chunks := range [][]string{{"okx\n"}, {"ok", "x", "\n"}} {
		output := &boundedProcessOutput{limit: 4, terminalLine: "ok"}
		for _, chunk := range chunks {
			if n, err := output.Write([]byte(chunk)); err != nil || n != len(chunk) {
				t.Fatalf("bounded marker write = %d/%v", n, err)
			}
		}
		if output.matchedTerminalLine() || output.didOverflow() {
			t.Fatal("extended marker matched or exceeded its inclusive byte allowance")
		}
	}
}

func TestSecretsPreservesSuppliedAnnouncementWriter(t *testing.T) {
	workspace := t.TempDir()
	stop := errors.New("inert scanner boundary")
	var output bytes.Buffer
	scans := 0
	runner := Runner{Root: t.TempDir(), Output: &output, Executor: workspaceExecutor{
		directory: workspace,
		run: func(_ context.Context, command Command) error {
			if command.Name == "go" {
				scans++
				return stop
			}
			return nil
		},
	}}
	if err := runner.Secrets(t.Context()); !errors.Is(err, stop) || scans != 1 {
		t.Fatalf("inert secret scan result = %v, scans = %d", err, scans)
	}
	if output.String() != "[.] secrets-history\n" {
		t.Fatalf("supplied announcement writer = %q", output.String())
	}
	if entries, err := os.ReadDir(workspace); err != nil || len(entries) != 0 {
		t.Fatalf("failed secret scan retained workspace files: %v", err)
	}
}

func TestDocsAdmitsInclusiveCatalogCardinality(t *testing.T) {
	for _, count := range []int{64, 65} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			modules := make([]inventory.Module, count)
			for index := range modules {
				modules[index] = inventory.Module{Directory: "module-" + strconv.Itoa(index)}
			}
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: t.TempDir(), Catalog: inventory.Inventory{Modules: modules}, Output: &output,
				Executor: executorFunction(func(context.Context, Command) error {
					calls++
					return nil
				}),
			}
			err := runner.Docs(t.Context(), []string{"module-0"})
			if count == 64 {
				if err != nil || output.String() != "[module-0] docs: not applicable\n" {
					t.Fatalf("inclusive catalog refused = %v, output = %q", err, output.String())
				}
			} else if err == nil || err.Error() != "module cardinality limit exceeded" || output.Len() != 0 {
				t.Fatalf("excess catalog admitted = %v, output = %q", err, output.String())
			}
			if calls != 0 {
				t.Fatal("catalog admission or disabled documentation dispatched an executor")
			}
		})
	}
}

func TestBlockCommentProseDoesNotHideLaterSuppression(t *testing.T) {
	const content = "package example\n/*\n// ordinary prose\n#nosec\n*/\nfunc value() {}\n"
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, "example.go", content, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	directives := nativeNosecGroupDirectives(parsed.Comments[0], fileSet)
	if len(directives) != 1 || directives[0].arguments != "" || directives[0].line != 4 {
		t.Fatalf("later native directive lost = %#v", directives)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "example.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkSecuritySuppressions(root); err == nil || !strings.HasSuffix(err.Error(), ":4: security suppression requires exact rule IDs") {
		t.Fatalf("later invalid suppression bypassed validation = %v", err)
	}
}
