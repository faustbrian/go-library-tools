package releasecheck

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"runtime/debug"
	"strings"
	"testing"
	"testing/fstest"
)

func compiledLicenseFixture(t *testing.T) (*debug.BuildInfo, fstest.MapFS) {
	t.Helper()
	corpus := fstest.MapFS{}
	if err := fs.WalkDir(binaryLicenseFiles, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(binaryLicenseFiles, name)
		corpus[name] = &fstest.MapFile{Data: bytes.Clone(data)}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var manifest binaryLicenseManifest
	if err := json.Unmarshal(corpus["licenses.json"].Data, &manifest); err != nil {
		t.Fatal(err)
	}
	info := &debug.BuildInfo{Path: manifest.BinaryPath, GoVersion: manifest.Runtime.Version, Settings: []debug.BuildSetting{{Key: "CGO_ENABLED", Value: "0"}}}
	for _, module := range manifest.Modules {
		info.Deps = append(info.Deps, &debug.Module{Path: module.Path, Version: module.Version, Sum: module.Sum})
	}
	return info, corpus
}

func TestBinaryLicensePayloadRetainsExactMaterialForCompiledGraphOnly(t *testing.T) {
	info, corpus := compiledLicenseFixture(t)
	license, notices, err := binaryLicensePayload(info, corpus)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(license, corpus["licenses/golib/LICENSE"].Data) {
		t.Fatal("product license bytes changed")
	}
	for name, file := range corpus {
		if name != "licenses.json" && name != "licenses/golib/LICENSE" && !bytes.Contains(notices, file.Data) {
			t.Fatalf("omitted source license material %s", name)
		}
	}
	// Manifest-only dependencies are not attributed as compiled binary inputs.
	info.Deps = info.Deps[1:]
	_, subset, err := binaryLicensePayload(info, corpus)
	if err != nil || bytes.Contains(subset, []byte("github.com/santhosh-tekuri/jsonschema/v6@")) {
		t.Fatalf("unused module retained or subset refused: %v", err)
	}
	info, corpus = compiledLicenseFixture(t)
	for left, right := 0, len(info.Deps)-1; left < right; left, right = left+1, right-1 {
		info.Deps[left], info.Deps[right] = info.Deps[right], info.Deps[left]
	}
	_, reordered, err := binaryLicensePayload(info, corpus)
	if err != nil || !bytes.Equal(notices, reordered) {
		t.Fatal("compiled dependency order changed deterministic notices")
	}
}

func TestBinaryLicensePayloadRefusesUnsupportedCompiledIdentities(t *testing.T) {
	for name, change := range map[string]func(*debug.BuildInfo){
		"unknown module":       func(i *debug.BuildInfo) { i.Deps[0].Path = "example.org/unknown" },
		"changed version":      func(i *debug.BuildInfo) { i.Deps[0].Version = "v6.0.4" },
		"changed checksum":     func(i *debug.BuildInfo) { i.Deps[0].Sum = "h1:changed" },
		"replacement":          func(i *debug.BuildInfo) { i.Deps[0].Replace = &debug.Module{Path: "replacement"} },
		"duplicate dependency": func(i *debug.BuildInfo) { i.Deps = append(i.Deps, i.Deps[0]) },
		"runtime drift":        func(i *debug.BuildInfo) { i.GoVersion = "go1.27.3" },
		"different binary":     func(i *debug.BuildInfo) { i.Path = "example.org/cmd/tool" },
		"cgo build":            func(i *debug.BuildInfo) { i.Settings[0].Value = "1" },
		"unknown cgo mode":     func(i *debug.BuildInfo) { i.Settings = nil },
		"ambiguous cgo mode":   func(i *debug.BuildInfo) { i.Settings = append(i.Settings, i.Settings[0]) },
		"missing dependency":   func(i *debug.BuildInfo) { i.Deps = append(i.Deps, nil) },
		"experimental runtime": func(i *debug.BuildInfo) {
			i.Settings = append(i.Settings, debug.BuildSetting{Key: "GOEXPERIMENT", Value: "boringcrypto"})
		},
	} {
		t.Run(name, func(t *testing.T) {
			info, corpus := compiledLicenseFixture(t)
			change(info)
			if _, _, err := binaryLicensePayload(info, corpus); err == nil {
				t.Fatal("unsupported compiled identity accepted")
			}
		})
	}
}

func TestBinaryLicensePayloadRefusesMissingChangedAndAmbiguousMaterial(t *testing.T) {
	for name, change := range map[string]func(fstest.MapFS){
		"missing manifest":   func(f fstest.MapFS) { delete(f, "licenses.json") },
		"malformed manifest": func(f fstest.MapFS) { f["licenses.json"].Data = []byte("{") },
		"unknown manifest field": func(f fstest.MapFS) {
			f["licenses.json"].Data = bytes.Replace(f["licenses.json"].Data, []byte(`"schema_version": 1`), []byte(`"schema_version": 1, "extra": true`), 1)
		},
		"trailing manifest":       func(f fstest.MapFS) { f["licenses.json"].Data = append(f["licenses.json"].Data, []byte(`{}`)...) },
		"missing NOTICE":          func(f fstest.MapFS) { delete(f, "licenses/yaml/NOTICE") },
		"changed NOTICE":          func(f fstest.MapFS) { f["licenses/yaml/NOTICE"].Data = []byte("replacement") },
		"changed runtime license": func(f fstest.MapFS) { f["licenses/go/LICENSE"].Data[0] ^= 1 },
		"changed product license": func(f fstest.MapFS) { f["licenses/golib/LICENSE"].Data[0] ^= 1 },
		"missing file mapping": func(f fstest.MapFS) {
			changeLicenseManifest(t, f, func(m *binaryLicenseManifest) { m.Modules[0].Files = nil })
		},
		"missing NOTICE mapping": func(f fstest.MapFS) {
			changeLicenseManifest(t, f, func(m *binaryLicenseManifest) { m.Modules[2].Files = m.Modules[2].Files[:1] })
		},
		"unsupported license": func(f fstest.MapFS) {
			changeLicenseManifest(t, f, func(m *binaryLicenseManifest) { m.Modules[0].License = "unsupported" })
		},
		"duplicate mapping": func(f fstest.MapFS) {
			changeLicenseManifest(t, f, func(m *binaryLicenseManifest) { m.Modules = append(m.Modules, m.Modules[0]) })
		},
		"duplicate file": func(f fstest.MapFS) {
			changeLicenseManifest(t, f, func(m *binaryLicenseManifest) { m.Modules[0].Files = append(m.Modules[0].Files, m.Modules[0].Files[0]) })
		},
		"escaping source path": func(f fstest.MapFS) {
			changeLicenseManifest(t, f, func(m *binaryLicenseManifest) { m.Modules[0].Files[0].Path = "../LICENSE" })
		},
		"insecure source": func(f fstest.MapFS) {
			changeLicenseManifest(t, f, func(m *binaryLicenseManifest) { m.Modules[0].Files[0].Source = "http://example.org/LICENSE" })
		},
	} {
		t.Run(name, func(t *testing.T) {
			info, corpus := compiledLicenseFixture(t)
			change(corpus)
			if _, _, err := binaryLicensePayload(info, corpus); err == nil {
				t.Fatal("unreviewed or incomplete license material accepted")
			}
		})
	}
}

func changeLicenseManifest(t *testing.T, corpus fstest.MapFS, change func(*binaryLicenseManifest)) {
	t.Helper()
	var manifest binaryLicenseManifest
	if err := json.Unmarshal(corpus["licenses.json"].Data, &manifest); err != nil {
		t.Fatal(err)
	}
	change(&manifest)
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	corpus["licenses.json"].Data = data
}

// FuzzVerifyBinaryArchiveRefusesHostileContentWithoutPanic caps untrusted
// archive content at 1 MiB; accepted output must at least be gzip and satisfy
// deeper owned archive checks, without executing its binary.
func FuzzVerifyBinaryArchiveRefusesHostileContentWithoutPanic(f *testing.F) {
	f.Add([]byte("not an archive"))
	f.Add([]byte{0x1f, 0x8b, 0x08})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		if err := verifyBinaryArchive(data, binaryLicenseFiles); err == nil && !strings.HasPrefix(string(data), "\x1f\x8b") {
			t.Fatal("non-gzip content accepted")
		}
	})
}
