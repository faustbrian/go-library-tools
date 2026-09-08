package assets

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractTarArchiveAcceptsGitGlobalPAXMetadataWithoutCreatingAnEntry(t *testing.T) {
	t.Parallel()

	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	if err := writer.WriteHeader(&tar.Header{
		Name:       "pax_global_header",
		Typeflag:   tar.TypeXGlobalHeader,
		PAXRecords: map[string]string{"comment": strings.Repeat("a", 40)},
	}); err != nil {
		t.Fatal(err)
	}
	contents := []byte("ok\n")
	if err := writer.WriteHeader(&tar.Header{Name: "README.md", Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	destination := t.TempDir()
	if err := extractTarArchive(bytes.NewReader(archive.Bytes()), destination, strings.Repeat("a", 40)); err != nil {
		t.Fatalf("extractTarArchive(git global PAX header) error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(destination, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, contents) {
		t.Fatalf("README.md = %q, want %q", got, contents)
	}
	if _, err := os.Lstat(filepath.Join(destination, "pax_global_header")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("global PAX metadata created an entry: %v", err)
	}
	if err := extractTarArchive(bytes.NewReader(archive.Bytes()), t.TempDir(), strings.Repeat("b", 40)); err == nil {
		t.Fatal("extractTarArchive accepted global PAX metadata for another commit")
	}
}

func TestSpecificationsProduceExactlyTwentyUniqueAssets(t *testing.T) {
	t.Parallel()

	if len(Specs) != 5 {
		t.Fatalf("len(Specs) = %d, want 5", len(Specs))
	}
	seen := map[string]struct{}{}
	for _, spec := range Specs {
		if len(spec.Commit) != 40 || spec.Base == "" || spec.CorpusEntrypoint == "" || len(spec.LoaderPaths)+len(spec.ParserPaths) == 0 {
			t.Fatalf("incomplete spec: %#v", spec)
		}
		for _, name := range AssetNames(spec) {
			if _, exists := seen[name]; exists {
				t.Fatalf("duplicate asset %q", name)
			}
			seen[name] = struct{}{}
		}
	}
	if len(seen) != 20 {
		t.Fatalf("asset count = %d, want 20", len(seen))
	}
}

func TestHistoricalBaselinePathFamiliesAlwaysRenderAsArrays(t *testing.T) {
	t.Parallel()
	for name, paths := range map[string][]string{"nil": nil, "empty": {}} {
		if got := pathArray(paths); got == nil {
			t.Fatalf("pathArray(%s) returned nil", name)
		}
	}

	for _, spec := range Specs {
		encoded, err := canonicalJSON(baseline{
			HistoricalSourceCommit: spec.Commit,
			LoaderPaths:            pathArray(spec.LoaderPaths),
			ParserPaths:            pathArray(spec.ParserPaths),
			Exports:                []exportRow{},
			AcceptedCases:          []corpusAccepted{},
			RejectedCases:          []corpusRejected{},
		})
		if err != nil {
			t.Fatal(err)
		}
		var value map[string]any
		if err := json.Unmarshal(encoded, &value); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"loader_paths", "parser_paths"} {
			if _, ok := value[field].([]any); !ok {
				t.Fatalf("%s baseline %s = %T, want JSON array", spec.Base, field, value[field])
			}
		}
	}
}

func TestCatalogMarkdownCharacterizationIsOutsideJSONProvenance(t *testing.T) {
	t.Parallel()
	for _, spec := range Specs {
		if spec.Base != "cohesion-catalog-v1" {
			continue
		}
		if len(spec.ParserPaths) != 1 || !strings.Contains(spec.ParserPaths[0], "#Project(") {
			t.Fatalf("catalog provenance parser paths = %v", spec.ParserPaths)
		}
		for _, path := range append(append([]string{}, spec.LoaderPaths...), spec.ParserPaths...) {
			if strings.Contains(path, "RenderMarkdown") {
				t.Fatalf("RenderMarkdown appears in JSON provenance: %s", path)
			}
		}
		if len(spec.RuntimeRunners) != 1 || spec.RuntimeRunners[0].Test != "TestHistoricalCatalogV1MarkdownRuntimeAndCLI" {
			t.Fatalf("catalog Markdown runtime characterization = %#v", spec.RuntimeRunners)
		}
		return
	}
	t.Fatal("catalog specification missing")
}

func TestPackageExportsRetainsConstantValuesAndStructTags(t *testing.T) {
	t.Parallel()
	pkg := types.NewPackage("example.test/module", "module")
	pkg.Scope().Insert(types.NewConst(token.NoPos, pkg, "Answer", types.Typ[types.UntypedInt], constant.MakeInt64(42)))
	field := types.NewField(token.NoPos, pkg, "Name", types.Typ[types.String], false)
	structure := types.NewStruct([]*types.Var{field}, []string{`json:"name"`})
	name := types.NewTypeName(token.NoPos, pkg, "Record", nil)
	types.NewNamed(name, structure, nil)
	pkg.Scope().Insert(name)
	rows := packageExports(pkg)
	var joined strings.Builder
	for _, row := range rows {
		joined.WriteString(row.Identifier + "=" + row.Signature + "\n")
	}
	joinedText := joined.String()
	if !strings.Contains(joinedText, "Answer") || !strings.Contains(joinedText, "42") {
		t.Fatalf("constant value missing: %s", joinedText)
	}
	if !strings.Contains(joinedText, `tag "json:\"name\""`) {
		t.Fatalf("struct tag missing: %s", joinedText)
	}
}

func TestOutputDirectoryMustBeEmptyAndNonSymlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	nonempty := filepath.Join(root, "nonempty")
	if err := os.Mkdir(nonempty, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nonempty, "owned"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := requireEmptyNonSymlinkDirectory(nonempty); err == nil {
		t.Fatal("nonempty directory accepted")
	}
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := requireEmptyNonSymlinkDirectory(link); err == nil {
		t.Fatal("symlink directory accepted")
	}
}

func TestGenerateRejectsTaskAndOutputDirectoryOverlapAtAPIBoundary(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		paths func(string) (string, string)
	}{
		{name: "equal", paths: func(root string) (string, string) { return filepath.Join(root, "task"), filepath.Join(root, "task") }},
		{name: "output-under-task", paths: func(root string) (string, string) {
			return filepath.Join(root, "task"), filepath.Join(root, "task", "output")
		}},
		{name: "task-under-output", paths: func(root string) (string, string) {
			return filepath.Join(root, "output", "task"), filepath.Join(root, "output")
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			taskRoot, outputRoot := test.paths(workspace)
			if err := os.MkdirAll(taskRoot, 0o700); err != nil {
				t.Fatal(err)
			}
			generator := Generator{
				Repository:         workspace,
				SchemaDir:          workspace,
				OutputDir:          outputRoot,
				TaskRoot:           taskRoot,
				GoBinary:           "go",
				GitBinary:          "git",
				ModuleProxy:        workspace,
				WorkspaceRoot:      workspace,
				RunnerTemplateDirs: []string{workspace},
			}
			if _, err := generator.Generate(); err == nil || !strings.Contains(err.Error(), "overlap") {
				t.Fatalf("Generate() error = %v, want overlap rejection", err)
			}
			if _, err := os.Lstat(taskRoot); err != nil {
				t.Fatalf("rejected task root was mutated: %v", err)
			}
		})
	}
}

func TestGenerateRejectsAbsentOutputThroughSymlinkParentIntoTaskRoot(t *testing.T) {
	t.Parallel()
	workspace := t.TempDir()
	taskRoot := filepath.Join(workspace, "task")
	if err := os.Mkdir(taskRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(workspace, "task-alias")
	if err := os.Symlink(taskRoot, alias); err != nil {
		t.Fatal(err)
	}
	outputRoot := filepath.Join(alias, "absent-output")
	generator := Generator{
		Repository:         workspace,
		SchemaDir:          workspace,
		OutputDir:          outputRoot,
		TaskRoot:           taskRoot,
		GoBinary:           "go",
		GitBinary:          "git",
		ModuleProxy:        workspace,
		WorkspaceRoot:      workspace,
		RunnerTemplateDirs: []string{workspace},
	}
	if _, err := generator.Generate(); err == nil || !strings.Contains(err.Error(), "resolved task/output isolation") {
		t.Fatalf("Generate() error = %v, want resolved overlap rejection", err)
	}
	if _, err := os.Lstat(taskRoot); !os.IsNotExist(err) {
		t.Fatalf("rejected task-owned tree remains: %v", err)
	}
	if _, err := os.Lstat(outputRoot); !os.IsNotExist(err) {
		t.Fatalf("rejected aliased output remains: %v", err)
	}
}

func TestCanonicalJSONAllocationBoundaryUsesActualGeneratorRenderer(t *testing.T) {
	t.Parallel()
	value := baseline{
		HistoricalSourceCommit: strings.Repeat("a", 40),
		LoaderPaths:            []string{"example.Load"},
		ParserPaths:            []string{"example.Parse"},
		ExportCount:            1,
		Exports:                []exportRow{{Identifier: "example.Value", Signature: "type Value string"}},
		AcceptedCaseCount:      1,
		AcceptedCases:          []corpusAccepted{{CaseID: "accepted", Entrypoint: "example.Parse", Input: "e30=", DecodedValueSHA256: "sha256:decoded", EmittedValueSHA256: "sha256:emitted"}},
		RejectedCaseCount:      1,
		RejectedCases:          []corpusRejected{{CaseID: "rejected", Entrypoint: "example.Parse", Input: "ew==", ErrorClass: "syntax"}},
		AcceptedCorpusSHA256:   "sha256:accepted",
		RejectedCorpusSHA256:   "sha256:rejected",
	}
	charged := 0
	want, err := canonicalJSONWithCharge(value, func(size int) error {
		charged += size
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if charged == 0 {
		t.Fatal("renderer reported no capacity requests")
	}
	for _, test := range []struct {
		name     string
		limit    int
		accepted bool
	}{
		{name: "below", limit: charged - 1, accepted: false},
		{name: "exact", limit: charged, accepted: true},
		{name: "above", limit: charged + 1, accepted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			used := 0
			got, err := canonicalJSONWithCharge(value, func(size int) error {
				if used+size > test.limit {
					return errors.New("injected renderer allocation limit")
				}
				used += size
				return nil
			})
			if test.accepted != (err == nil) {
				t.Fatalf("limit=%d accepted=%v error=%v", test.limit, test.accepted, err)
			}
			if err == nil && !bytes.Equal(got, want) {
				t.Fatalf("canonical JSON = %s, want %s", got, want)
			}
		})
	}
}

func TestObservedValueDigestUsesHistoricalJSONSemantics(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "negative-zero", value: `-0`, want: `0`},
		{name: "fraction", value: `1.5`, want: `1.5`},
		{name: "exponent", value: `1e2`, want: `100`},
		{name: "escaped-noncharacter", value: `"\ufdd0"`, want: "\"﷐\""},
		{name: "raw-noncharacter", value: "\"﷐\"", want: "\"﷐\""},
	} {
		t.Run(test.name, func(t *testing.T) {
			encoded := base64.StdEncoding.EncodeToString([]byte(test.value))
			got, err := observedValueDigest("json", encoded)
			if err != nil {
				t.Fatal(err)
			}
			if got != digest([]byte(test.want)) {
				t.Fatalf("digest = %s, want semantic %q digest", got, test.want)
			}
		})
	}
}

func TestObservedValueDigestUsesReleasedHistoricalDepthBoundary(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		rawDepth int
		accepted bool
	}{
		{name: "below", rawDepth: 9996, accepted: true},
		{name: "exact", rawDepth: 9997, accepted: true},
		{name: "above", rawDepth: 9998, accepted: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := `{"modules":[{"provenance":` + strings.Repeat("[", test.rawDepth) + "0" + strings.Repeat("]", test.rawDepth) + `}]}`
			_, err := observedValueDigest("json", base64.StdEncoding.EncodeToString([]byte(value)))
			if test.accepted != (err == nil) {
				t.Fatalf("raw depth=%d accepted=%v error=%v", test.rawDepth, test.accepted, err)
			}
		})
	}
}

func TestWriteAssetBundleCleansFilesAfterPartialFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_, err := writeAssetBundle(root, []string{"first.json", "missing/second.json"}, map[string][]byte{"first.json": []byte(`{}`), "missing/second.json": []byte(`{}`)})
	if err == nil {
		t.Fatal("partial write unexpectedly succeeded")
	}
	entries, readErr := os.ReadDir(root)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("partial asset remains: %v", entries)
	}
}

func TestBuildCorporaKeepsDecodedAndEmittedDigestsIndependent(t *testing.T) {
	t.Parallel()
	spec := Spec{LoaderPaths: []string{"Load"}, DefaultValueKind: "json"}
	input := base64.StdEncoding.EncodeToString([]byte(`{"input":true}`))
	decoded := base64.StdEncoding.EncodeToString([]byte(`{"form":"decoded"}`))
	emitted := base64.StdEncoding.EncodeToString([]byte(`{"form":"emitted"}`))
	accepted, rejected, err := buildCorpora(spec, []observation{
		{CaseID: "different", EntryPoint: "Load", InputBase64: input, Outcome: "accepted", DecodedValueBase64: decoded, EmittedValueBase64: emitted},
		{CaseID: "rejected", EntryPoint: "Load", InputBase64: input, Outcome: "rejected", ErrorClass: "invalid"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rejected) != 1 || len(accepted) != 1 {
		t.Fatalf("accepted=%d rejected=%d", len(accepted), len(rejected))
	}
	if accepted[0].DecodedValueSHA256 == accepted[0].EmittedValueSHA256 {
		t.Fatal("distinct observations produced equal digests")
	}
}

func TestBuildCorporaRequiresAcceptedAndRejectedCoveragePerEntrypoint(t *testing.T) {
	t.Parallel()
	encoded := base64.StdEncoding.EncodeToString([]byte(`null`))
	spec := Spec{LoaderPaths: []string{"Load", "Verify"}, DefaultValueKind: "json"}
	_, _, err := buildCorpora(spec, []observation{
		{CaseID: "load-ok", EntryPoint: "Load", InputBase64: encoded, Outcome: "accepted", DecodedValueBase64: encoded, EmittedValueBase64: encoded},
		{CaseID: "load-bad", EntryPoint: "Load", InputBase64: encoded, Outcome: "rejected", ErrorClass: "invalid"},
	})
	if err == nil {
		t.Fatal("missing entrypoint coverage accepted")
	}
}

func TestResolvedConfinementRejectsSymlinkEscape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if err := requireResolvedWithin(root, link); err == nil {
		t.Fatal("resolved symlink escape accepted")
	}
}

func TestHashModuleFilesMatchesGoDirectoryHashVector(t *testing.T) {
	t.Parallel()
	got := hashModuleFiles([]moduleFile{{name: "a", data: []byte("x")}})
	if want := "h1:SkuHSh6XdqUuOP2CoFOCXzCfRhh6gDoPwMZnDGZw52g="; got != want {
		t.Fatalf("hashModuleFiles() = %q, want %q", got, want)
	}
}

func TestSnapshotProxyModulesCopiesOnlyVerifiedAllowlist(t *testing.T) {
	t.Parallel()
	source := t.TempDir()
	module := proxyModule{path: "example.com/module", version: "v1.2.3"}
	modData := []byte("module example.com/module\n")
	zipData := testModuleZip(t, "example.com/module@v1.2.3/go.mod", modData)
	module.modHash = hashModuleFiles([]moduleFile{{name: "go.mod", data: modData}})
	var err error
	module.zipHash, err = hashModuleZip(zipData)
	if err != nil {
		t.Fatal(err)
	}
	testWriteProxyModule(t, source, module, modData, zipData)
	if err := os.WriteFile(filepath.Join(source, "unrelated"), []byte("must not copy"), 0o600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "proxy")
	digests, err := snapshotProxyModules(source, destination, []proxyModule{module})
	if err != nil {
		t.Fatal(err)
	}
	if len(digests) != 4 {
		t.Fatalf("copied file count = %d, want 4", len(digests))
	}
	if _, err := os.Stat(filepath.Join(destination, "unrelated")); !os.IsNotExist(err) {
		t.Fatalf("unallowlisted file copied: %v", err)
	}

	corruptSource := t.TempDir()
	corruptZip := testModuleZip(t, "example.com/module@v1.2.3/go.mod", append(modData, []byte("// changed\n")...))
	testWriteProxyModule(t, corruptSource, module, modData, corruptZip)
	if _, err := snapshotProxyModules(corruptSource, filepath.Join(t.TempDir(), "proxy"), []proxyModule{module}); err == nil {
		t.Fatal("proxy with zip content divergent from frozen sum accepted")
	}
}

func TestSnapshotProxyModulesRejectsSymlinkedParent(t *testing.T) {
	t.Parallel()
	outside := t.TempDir()
	source := t.TempDir()
	module := proxyModule{path: "example.com/module", version: "v1.2.3"}
	modData := []byte("module example.com/module\n")
	zipData := testModuleZip(t, "example.com/module@v1.2.3/go.mod", modData)
	module.modHash = hashModuleFiles([]moduleFile{{name: "go.mod", data: modData}})
	var err error
	module.zipHash, err = hashModuleZip(zipData)
	if err != nil {
		t.Fatal(err)
	}
	testWriteProxyModule(t, outside, module, modData, zipData)
	if err := os.Symlink(filepath.Join(outside, "example.com"), filepath.Join(source, "example.com")); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshotProxyModules(source, filepath.Join(t.TempDir(), "proxy"), []proxyModule{module}); err == nil {
		t.Fatal("proxy path with symlinked parent accepted")
	}
}

func TestFrozenSnapshotIsReadOnlyAndRemovable(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "snapshot")
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "nested", "asset")
	if err := os.WriteFile(path, []byte("fixed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := freezeSnapshot(root); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o400 {
		t.Fatalf("snapshot file mode = %o, want 400", info.Mode().Perm())
	}
	if err := cleanupOwnedTree(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("frozen snapshot remains: %v", err)
	}
}

func TestCleanupOwnedTreeMakesCachesWritableAndVerifiesAbsence(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "cache")
	if err := os.MkdirAll(filepath.Join(root, "readonly"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "readonly", "entry"), []byte("cache"), 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "readonly"), 0o500); err != nil {
		t.Fatal(err)
	}
	if err := cleanupOwnedTree(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("cleanup target remains: %v", err)
	}
}

func TestCleanupOwnedTreeReturnsRemovalAndAbsenceFailures(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		remover func(string) error
		want    string
	}{
		{name: "removal-error", remover: func(string) error { return errors.New("injected removal failure") }, want: "injected removal failure"},
		{name: "false-success", remover: func(string) error { return nil }, want: "cleanup target remains"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "cache")
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
			err := cleanupOwnedTreeWithRemover(root, test.remover)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("cleanup error = %v, want %q", err, test.want)
			}
			if err := cleanupOwnedTree(root); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestJoinCleanupPropagatesFailureWithAndWithoutPrimaryError(t *testing.T) {
	t.Parallel()
	cleanupErr := errors.New("injected cleanup failure")
	for _, test := range []struct {
		name    string
		primary error
	}{
		{name: "successful-operation", primary: nil},
		{name: "failed-operation", primary: errors.New("injected operation failure")},
	} {
		t.Run(test.name, func(t *testing.T) {
			resultErr := test.primary
			joinCleanup(&resultErr, cleanupErr)
			if !errors.Is(resultErr, cleanupErr) {
				t.Fatalf("cleanup error was not propagated: %v", resultErr)
			}
			if test.primary != nil && !errors.Is(resultErr, test.primary) {
				t.Fatalf("primary error was not retained: %v", resultErr)
			}
		})
	}
}

func testModuleZip(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func testWriteProxyModule(t *testing.T, root string, module proxyModule, modData, zipData []byte) {
	t.Helper()
	directory := filepath.Join(root, filepath.FromSlash(module.path), "@v")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	info, err := json.Marshal(map[string]string{"Version": module.version})
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		module.version + ".info":    info,
		module.version + ".mod":     modData,
		module.version + ".zip":     zipData,
		module.version + ".ziphash": []byte(module.zipHash + "\n"),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(directory, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
