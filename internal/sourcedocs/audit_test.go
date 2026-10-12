package sourcedocs

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func trackedFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	args := make([]string, 3, 3+len(files))
	copy(args, []string{"add", "-f", "--"})
	for name, content := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		args = append(args, name)
	}
	for _, arguments := range [][]string{{"init", "-q"}, args} {
		command := exec.CommandContext(t.Context(), "git", arguments...)
		command.Dir = root
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func codes(report Report) []string {
	result := make([]string, 0, len(report.Findings))
	for _, finding := range report.Findings {
		result = append(result, finding.Code)
	}
	return result
}

func TestAuditCoversNestedModulesAllVariantsAndPublicMembers(t *testing.T) {
	root := trackedFixture(t, map[string]string{
		"go.mod":                  "module example\n",
		"doc.go":                  "// Package example owns the fixture contract.\npackage example\n",
		"api.go":                  "package example\n// API owns one value.\ntype API struct {\n Value int\n}\n// Reader returns values.\ntype Reader interface {\n Read() int\n}\n// Values names grouped constants.\nconst (One = 1; Two = 2)\n// Left and Right are paired values.\nvar Left, Right int\n",
		"api_windows.go":          "//go:build windows\n\npackage example\nfunc Platform() {}\n",
		"api_test.go":             "package example_test\nfunc TestIgnored() {}\n",
		"nested/go.mod":           "module example/nested\n",
		"nested/doc.go":           "// Package nested owns another module.\npackage nested\n",
		"z/go.mod":                "module example/z\n",
		"z/doc.go":                "// Package z owns a short directory name.\npackage z\n",
		"testdata/fixture/go.mod": "module fixture\n",
		"testdata/fixture/doc.go": "package fixture\nfunc Fixture() {}\n",
		"vendor/external.go":      "invalid vendor code is deliberately not parsed",
	})
	first, err := Audit(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Audit(t.Context(), root)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatal("audit is not deterministic", err)
	}
	if len(first.Modules) != 4 || len(first.Files) != 7 || first.ExcludedVendorFiles != 1 {
		t.Fatalf("scope = %#v", first)
	}
	if got := codes(first); !reflect.DeepEqual(got, []string{"exported-comment", "exported-comment", "exported-comment"}) {
		t.Fatalf("public members and platform declarations = %v", got)
	}
	for _, file := range first.Files {
		if file.Path == "z/doc.go" && file.Module != "z" {
			t.Fatal("single-character nested module was attributed to root")
		}
		if file.Path == "nested/doc.go" && file.Module != "nested" {
			t.Fatal("nested module was attributed to root")
		}
		if file.Path == "testdata/fixture/doc.go" && (file.Kind != "fixture" || file.Module != "testdata/fixture") {
			t.Fatal("fixture was certified as production")
		}
		if file.Path == "api_windows.go" && !file.BuildConstrained {
			t.Fatal("inactive platform was omitted")
		}
	}
}

func TestAuditMarkerDeprecationLinksAndSecretsRemainObjectiveAndRedacted(t *testing.T) {
	secret := "ghp_" + strings.Repeat("Z", 36)
	root := trackedFixture(t, map[string]string{
		"go.mod":          "module example\n",
		"doc.go":          "// Package example checks comments.\npackage example\n",
		"api.go":          "package example\n// Old returns a value.\n// deprecated: use New.\nfunc Old() {}\n// Good returns a value.\n//\n// Deprecated: Use New instead.\nfunc Good() {}\n// TODO: missing behavior\n// FIXME: https://github.com/example/example/issues/12 incorrect result.\n// NOTE: allocation ownership is intentional.\n// HACK: owner=example trigger=upstream-fix temporary bridge.\n// WARNING\n// Link [missing](not-there.md) and [spec](https://example.org/spec).\n// [local]: doc.go\n// " + secret + "\n",
		"example_test.go": "package example\nfunc ExampleOld() { println(\"" + secret + "\") }\n",
	})
	report, err := Audit(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, code := range codes(report) {
		counts[code]++
	}
	for code, count := range map[string]int{"deprecated-syntax": 1, "marker-context": 2, "marker-review": 3, "comment-link": 1, "secret-pattern": 2} {
		if counts[code] != count {
			t.Errorf("%s = %d want %d", code, counts[code], count)
		}
	}
	encoded, err := json.Marshal(report)
	if err != nil || bytes.Contains(encoded, []byte(secret)) {
		t.Fatal("report disclosed comment/example contents", err)
	}
	if !report.HasErrors() {
		t.Fatal("strict check ignored objective defects")
	}
}

func TestAuditPublicSurfaceExcludesPrivateImplementationAndRetainsReachableMembers(t *testing.T) {
	root := trackedFixture(t, map[string]string{
		"go.mod": "module example\n",
		"doc.go": "// Package example checks consumer-visible members.\npackage example\n",
		"api.go": `package example
// Public exposes nested and promoted members.
type Public struct {
 // Anonymous owns a nested value.
 Anonymous struct { Missing int }
 promoted
 private wire
}
// Exposed returns an unnamed public value.
func Exposed() struct { Value int } { panic("not executed") }
// Alias preserves reachable concrete members.
type Alias = hiddenResult
// Result returns a concrete public capability.
var Result = hiddenResult{}
// Count is an integer result, not the private factory input.
var Count = factory(factoryInput{})
`,
		"private.go": `package example
type wire struct { PrivateWire int }
func (wire) Error() string { return "private" }
type privateIO interface { Read() int }
func hidden() { _ = struct { Local int }{} }
type promoted struct { Promoted int }
type hiddenResult struct { Visible int }
func (hiddenResult) Reachable() {}
type factoryInput struct { Argument int }
func factory(factoryInput) int { return 0 }
`,
	})
	report, err := Audit(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	if got := codes(report); !reflect.DeepEqual(got, []string{"exported-comment", "exported-comment", "exported-comment", "exported-comment", "exported-comment"}) {
		t.Fatalf("consumer-visible members = %#v", report.Findings)
	}
	for _, finding := range report.Findings {
		if finding.Path == "private.go" && (finding.Line < 6 || finding.Line > 8) {
			t.Fatalf("private implementation got a public documentation requirement: %#v", finding)
		}
	}
}

func TestAuditAcceptsTrackedCanonicalHeaderGeneratorWithoutDirective(t *testing.T) {
	for _, scenario := range []string{"available", "missing", "untracked", "symlink"} {
		t.Run(scenario, func(t *testing.T) {
			root := trackedFixture(t, map[string]string{
				"go.mod":               "module example\n",
				"doc.go":               "// Package example checks an owned generator.\npackage example\n",
				"generated.go":         "// Code generated by scripts/generate.php; DO NOT EDIT.\npackage example\nfunc Generated() {}\n",
				"scripts/generate.php": "<?php throw new Exception('never execute');\n",
			})
			generator := filepath.Join(root, "scripts", "generate.php")
			switch scenario {
			case "missing":
				if err := os.Remove(generator); err != nil {
					t.Fatal(err)
				}
			case "untracked":
				command := exec.CommandContext(t.Context(), "git", "rm", "--cached", "scripts/generate.php")
				command.Dir = root
				if err := command.Run(); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Remove(generator); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(t.TempDir(), "outside.php"), generator); err != nil {
					t.Fatal(err)
				}
			}
			report, err := Audit(t.Context(), root)
			if err != nil {
				t.Fatal(err)
			}
			if report.HasErrors() != (scenario != "available") {
				t.Fatalf("header-only generator %s: %#v", scenario, report.Findings)
			}
		})
	}
}

func TestAuditCanonicalHeaderGoCommandDirectory(t *testing.T) {
	for _, scenario := range []string{"available", "shell directive", "build-constrained", "non-command", "no main", "invalid main", "missing", "test-only", "fixture-only", "generated-only", "untracked", "outside", "symlink", "non-Go"} {
		t.Run(scenario, func(t *testing.T) {
			files := map[string]string{
				"go.mod":                   "module example\n",
				"doc.go":                   "// Package example checks an owned generator.\npackage example\n",
				"generated.go":             "// Code generated by internal/cmd/gen; DO NOT EDIT.\npackage example\nfunc Generated() {}\n",
				"internal/cmd/gen/main.go": "// Command gen produces the owned table.\npackage main\nfunc main() { panic(\"generator must never execute\") }\n",
			}
			switch scenario {
			case "shell directive":
				files["doc.go"] += "//go:generate ./scripts/generate.sh\n"
				files["scripts/generate.sh"] = "#!/bin/sh\nexit 1\n"
			case "build-constrained":
				files["internal/cmd/gen/main.go"] = "//go:build ignore\n\n" + files["internal/cmd/gen/main.go"]
			case "no main":
				files["internal/cmd/gen/main.go"] = "// Command gen has no entry point.\npackage main\n"
			case "invalid main":
				files["internal/cmd/gen/main.go"] = "// Command gen has an invalid entry point.\npackage main\nfunc main(value int) {}\n"
			case "non-command":
				files["internal/cmd/gen/main.go"] = "// Package gen is not a command.\npackage gen\n"
			case "test-only", "fixture-only", "generated-only", "non-Go", "missing":
				source := files["internal/cmd/gen/main.go"]
				delete(files, "internal/cmd/gen/main.go")
				switch scenario {
				case "test-only":
					files["internal/cmd/gen/main_test.go"] = source
				case "fixture-only":
					files["internal/cmd/gen/testdata/main.go"] = source
					files["generated.go"] = "// Code generated by internal/cmd/gen/testdata; DO NOT EDIT.\npackage example\n"
				case "generated-only":
					files["internal/cmd/gen/main.go"] = "// Code generated by externaltool; DO NOT EDIT.\n" + source
				case "non-Go":
					files["internal/cmd/gen/README.md"] = "Not a command.\n"
				}
			case "outside":
				files["generated.go"] = "// Code generated by ../outside; DO NOT EDIT.\npackage example\n"
			}
			root := trackedFixture(t, files)
			if scenario == "untracked" || scenario == "symlink" {
				command := exec.CommandContext(t.Context(), "git", "rm", "--cached", "internal/cmd/gen/main.go")
				command.Dir = root
				if err := command.Run(); err != nil {
					t.Fatal(err)
				}
				if scenario == "symlink" {
					if err := os.RemoveAll(filepath.Join(root, "internal/cmd/gen")); err != nil {
						t.Fatal(err)
					}
					outside := t.TempDir()
					if err := os.WriteFile(filepath.Join(outside, "main.go"), []byte(files["internal/cmd/gen/main.go"]), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(outside, filepath.Join(root, "internal/cmd/gen")); err != nil {
						t.Fatal(err)
					}
				}
			}
			report, err := Audit(t.Context(), root)
			if err != nil {
				t.Fatal(err)
			}
			available := scenario == "available" || scenario == "shell directive" || scenario == "build-constrained"
			if slices.Contains(codes(report), "generator-unavailable") == available {
				t.Fatalf("directory producer %s: %#v", scenario, report.Findings)
			}
			if scenario == "shell directive" && !slices.Contains(codes(report), "generator-review") {
				t.Fatalf("unsupported shell regeneration lost manual review: %#v", report.Findings)
			}
		})
	}
}

func TestAuditRetainsExplicitGenericAndFunctionValueSignatures(t *testing.T) {
	root := trackedFixture(t, map[string]string{
		"go.mod": "module example\n",
		"doc.go": "// Package example checks explicit public signatures.\npackage example\n",
		"api.go": `package example
// Public constrains callers through an anonymous interface.
type Public[T interface { Missing() }] struct{}
// Function returns a public anonymous value.
var Function = func() struct { Missing int } { return struct { Missing int }{} }
// Called holds an explicit anonymous return value.
var Called = (func() struct { CallResult int } { return struct { CallResult int }{} })()
`,
	})
	report, err := Audit(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	if got := codes(report); !reflect.DeepEqual(got, []string{"exported-comment", "exported-comment", "exported-comment"}) {
		t.Fatalf("explicit generic/function signatures = %#v", report.Findings)
	}
}

func TestAuditDoesNotMistakeGeneratorInputsForProducerIdentity(t *testing.T) {
	for _, identity := range []string{"by externaltool from input.go;", "by externaltool;"} {
		root := trackedFixture(t, map[string]string{
			"go.mod":       "module example\n",
			"doc.go":       "// Package example checks generator ownership.\npackage example\n",
			"input.go":     "package example\n",
			"generated.go": "// Code generated " + identity + " DO NOT EDIT.\npackage example\n",
		})
		report, err := Audit(t.Context(), root)
		if err != nil {
			t.Fatal(err)
		}
		if report.HasErrors() || !slices.Contains(codes(report), "generator-review") {
			t.Fatalf("ambiguous generator identity %q = %#v", identity, report.Findings)
		}
	}
}

func TestAuditGeneratedRegenerationDeclarationIsInspectedWithoutExecution(t *testing.T) {
	root := trackedFixture(t, map[string]string{
		"go.mod":                "module example\n",
		"doc.go":                "// Package example checks generated source.\npackage example\n//go:generate go run ./gen.go\n",
		"gen.go":                "//go:build ignore\n\npackage main\nfunc main() { panic(\"generator must never execute\") }\n",
		"generated.go":          "// Code generated by gen.go; DO NOT EDIT.\npackage example\nfunc GeneratedUndocumented() {}\n",
		"identity_generated.go": "// Code generated ; DO NOT EDIT.\npackage example\n",
		"missing/generated.go":  "// Code generated by missing generator; DO NOT EDIT.\npackage missing\n",
		"stale.go":              "// Code generated by stale tool.\npackage example\n",
	})
	report, err := Audit(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, code := range codes(report) {
		counts[code]++
	}
	if counts["generator-review"] != 1 || counts["generated-header"] != 1 || counts["generated-identity"] != 1 || counts["exported-comment"] != 0 {
		t.Fatalf("generated accounting = %v", counts)
	}
	if err := os.Remove(filepath.Join(root, "gen.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := Audit(t.Context(), root); err == nil {
		t.Fatal("missing tracked generator was silently accepted")
	}
}

func TestAuditRejectsFilesystemEscapesAndUnsafeAcquisition(t *testing.T) {
	for _, scenario := range []string{"source symlink", "directory symlink", "oversized", "malformed", "nested root", "canceled"} {
		t.Run(scenario, func(t *testing.T) {
			root := trackedFixture(t, map[string]string{"go.mod": "module example\n", "nested/api.go": "// Package example owns input.\npackage example\n"})
			ctx := t.Context()
			name := filepath.Join(root, "nested", "api.go")
			switch scenario {
			case "source symlink":
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(t.TempDir(), "outside.go"), name); err != nil {
					t.Fatal(err)
				}
			case "directory symlink":
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(filepath.Dir(name)); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(t.TempDir(), filepath.Dir(name)); err != nil {
					t.Fatal(err)
				}
			case "oversized":
				if err := os.WriteFile(name, []byte(strings.Repeat(" ", maximumFileBytes+1)), 0o600); err != nil {
					t.Fatal(err)
				}
			case "malformed":
				if err := os.WriteFile(name, []byte("package example\nfunc Broken("), 0o600); err != nil {
					t.Fatal(err)
				}
			case "nested root":
				root = filepath.Dir(name)
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if _, err := Audit(ctx, root); err == nil {
				t.Fatal("unsafe acquisition succeeded")
			}
		})
	}
}

func TestPathInventoryAdmission(t *testing.T) {
	for _, data := range [][]byte{[]byte("../outside.go\x00"), []byte("/absolute.go\x00"), []byte("a\nb.go\x00"), []byte("api.go"), []byte(strings.Repeat("a", maximumPathBytes+1))} {
		if _, err := parsePaths(data); err == nil {
			t.Fatal("unsafe tracked input admitted")
		}
	}
	paths, err := parsePaths([]byte("z.go\x00a.go\x00z.go\x00"))
	if err != nil || !reflect.DeepEqual(paths, []string{"a.go", "z.go"}) {
		t.Fatal("paths are not canonical", paths, err)
	}
}

func TestMalformedFixtureIsVisibleButNotCertifiedAsProduction(t *testing.T) {
	root := trackedFixture(t, map[string]string{"go.mod": "module example\n", "doc.go": "// Package example owns fixtures.\npackage example\n", "testdata/broken.go": "package broken\nfunc Broken("})
	report, err := Audit(t.Context(), root)
	if err != nil || !reflect.DeepEqual(codes(report), []string{"fixture-parse-review"}) || report.HasErrors() || len(report.Files) != 2 {
		t.Fatalf("negative fixture accounting = %#v, %v", report, err)
	}
	if err := os.WriteFile(filepath.Join(root, "untracked.go"), []byte("not parsed"), 0o600); err != nil {
		t.Fatal(err)
	}
	again, err := Audit(t.Context(), root)
	if err != nil || !reflect.DeepEqual(report, again) {
		t.Fatal("untracked source entered tracked inventory", err)
	}
}

func TestFreeFormDebtContextRequiresReviewNotSyntaxRewrite(t *testing.T) {
	root := trackedFixture(t, map[string]string{"go.mod": "module example\n", "doc.go": "// Package example owns policy.\npackage example\n// TODO(Brian): enable the adapter when upstream fixes cancellation.\n"})
	report, err := Audit(t.Context(), root)
	if err != nil || report.HasErrors() || len(report.Findings) != 1 || report.Findings[0].Code != "marker-context" || report.Findings[0].Severity != "review" {
		t.Fatalf("free-form actionable context was given an unsupported fatal syntax requirement: %#v, %v", report, err)
	}
}

func TestDeprecatedAPINameIsNotADeactivationMarker(t *testing.T) {
	root := trackedFixture(t, map[string]string{"go.mod": "module example\n", "doc.go": "// Package example owns naming.\npackage example\n", "api.go": "package example\n// Deprecated returns a historical state name, not an obsolete API.\nfunc Deprecated() {}\n"})
	report, err := Audit(t.Context(), root)
	if err != nil || report.HasErrors() || len(report.Findings) != 0 {
		t.Fatalf("ordinary documentation of API named Deprecated fails strict check: %#v, %v", report.Findings, err)
	}
}

func TestGeneratorValidatesAllLocalInputsAndGoQuotedArguments(t *testing.T) {
	for _, scenario := range []struct {
		name, directive, extra string
		wantError, wantReview  bool
	}{
		{"missing later input", "go run present.go missing.go", "", true, false},
		{"quoted existing input", "go run \"generator with spaces.go\"", "generator with spaces.go", false, false},
		{"quoted missing later input", "go run present.go \"missing with spaces.go\"", "", true, false},
		{"multiple existing inputs", "go run present.go other.go", "other.go", false, false},
		{"program arguments are not source files", "go run present.go -- missing.go", "", false, false},
		{"external pinned tool", "go run example.invalid/generator@v1.0.0", "", false, true},
		{"unsupported flags", "go run -tags=generator present.go", "", false, true},
		{"environment input", "go run $GENERATOR", "", false, true},
		{"ambiguous quote", "go run \"present.go", "", false, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			files := map[string]string{"go.mod": "module example\n", "doc.go": "// Package example owns generation.\npackage example\n//go:generate " + scenario.directive + "\n", "present.go": "package example\n", "generated.go": "// Code generated by owned generator; DO NOT EDIT.\npackage example\n"}
			if scenario.extra != "" {
				files[scenario.extra] = "package example\n"
			}
			root := trackedFixture(t, files)
			report, err := Audit(t.Context(), root)
			if err != nil {
				t.Fatal(err)
			}
			gotError, gotReview := false, false
			for _, finding := range report.Findings {
				gotError = gotError || finding.Code == "generator-unavailable"
				gotReview = gotReview || finding.Code == "generator-review"
			}
			if gotError != scenario.wantError || gotReview != scenario.wantReview || report.HasErrors() != scenario.wantError {
				t.Fatalf("generator arguments have incorrect strict/review disposition: %v", codes(report))
			}
		})
	}
}

// FuzzGeneratorArguments caps directive text at 8 KiB, asserting deterministic
// lexical admission and at most 256 argv entries, without executing generators.
func FuzzGeneratorArguments(f *testing.F) {
	f.Add("go run present.go missing.go")
	f.Add("go run \"generator with spaces.go\"")
	f.Add("go run \"unterminated")
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 8192 {
			t.Skip()
		}
		first, understood := generatorArguments(input)
		second, again := generatorArguments(input)
		if !reflect.DeepEqual(first, second) || understood != again {
			t.Fatal("generator lexical admission is not deterministic")
		}
		if len(first) > 256 {
			t.Fatal("generator argv exceeded declared bound")
		}
	})
}

func TestLimitsRefusePartialSourceInventories(t *testing.T) {
	if _, err := parsePaths([]byte(strings.Repeat("a.go\x00", maximumPaths+1))); err == nil {
		t.Fatal("path count limit admitted excess source")
	}
	var output boundedBuffer
	if _, err := output.Write(make([]byte, maximumPathBytes+1)); err == nil || output.Len() != 0 {
		t.Fatal("oversized Git output retained partial inventory")
	}
	report := Report{}
	file := parsedFile{file: File{Path: "doc.go"}, fset: token.NewFileSet()}
	for range maximumFindings + 100 {
		addFinding(&report, file, token.NoPos, "limit", "error", "bounded")
	}
	if len(report.Findings) != maximumFindings+1 {
		t.Fatal("finding overflow was unbounded or silently discarded")
	}
}

// FuzzTrackedPaths caps Git path bytes at 8 KiB, asserting deterministic
// admission and refusal of unsafe paths.
func FuzzTrackedPaths(f *testing.F) {
	f.Add([]byte("doc.go\x00nested/api.go\x00"))
	f.Add([]byte("../escape.go\x00"))
	f.Add([]byte("unclosed"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 8192 {
			t.Skip()
		}
		first, firstErr := parsePaths(data)
		second, secondErr := parsePaths(data)
		if !reflect.DeepEqual(first, second) || (firstErr == nil) != (secondErr == nil) {
			t.Fatal("path admission is not deterministic")
		}
		for _, name := range first {
			if pathIsUnsafe(name) {
				t.Fatal("unsafe admitted path")
			}
		}
	})
}

// FuzzCommentStructure caps source at 32 KiB, parses accepted ASTs and asserts
// deterministic consumer-surface/comment findings without building package code.
func FuzzCommentStructure(f *testing.F) {
	root := f.TempDir()
	f.Add([]byte("// Package example owns parsing.\npackage example\n// Value is immutable.\ntype Value struct { Field int }\n"))
	f.Add([]byte("package example\n// TODO\n// Deprecated:\n// [link]: ../../outside\nfunc Exported() {}\n"))
	f.Fuzz(func(t *testing.T, content []byte) {
		if len(content) > 32768 {
			t.Skip()
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "source.go", content, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return
		}
		parsed := parsedFile{file: File{Path: "source.go", Kind: "source", Generated: ast.IsGenerated(file)}, relative: "source.go", ast: file, fset: fset}
		first, second := Report{}, Report{}
		surface := consumerSurface([]parsedFile{parsed})
		inspectFile(root, parsed, []parsedFile{parsed}, nil, surface, &first)
		inspectFile(root, parsed, []parsedFile{parsed}, nil, surface, &second)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("comment checks are not deterministic")
		}
	})
}

func pathIsUnsafe(name string) bool {
	return name == "" || strings.HasPrefix(name, "../") || filepath.IsAbs(name) || strings.ContainsAny(name, "\x00\r\n")
}
