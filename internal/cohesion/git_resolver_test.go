package cohesion

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveGitSourceFileReadsPinnedCleanObjectAndRejectsDirtyTree(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "init", "-q")
	runGitResolverTestCommand(t, root, "config", "user.name", "Test")
	runGitResolverTestCommand(t, root, "config", "user.email", "test@example.invalid")
	runGitResolverTestCommand(t, root, "remote", "add", "origin", "https://github.com/faustbrian/example.git")
	if err := os.WriteFile(filepath.Join(root, "input.json"), []byte("pinned"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "add", "input.json")
	runGitResolverTestCommand(t, root, "commit", "-q", "-m", "test")
	revision := strings.TrimSpace(runGitResolverTestCommand(t, root, "rev-parse", "HEAD"))

	got, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6)
	if err != nil || string(got) != "pinned" {
		t.Fatalf("ResolveGitSourceFile(clean) = %q, %v", got, err)
	}
	alternates := filepath.Join(root, ".git", "objects", "info", "alternates")
	if err := os.WriteFile(alternates, []byte(t.TempDir()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6); err == nil {
		t.Fatal("ResolveGitSourceFile(alternates) error = nil")
	}
	if err := os.Remove(alternates); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "missing"), alternates); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6); err == nil {
		t.Fatal("ResolveGitSourceFile(symlinked alternates) error = nil")
	}
	if err := os.Remove(alternates); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "input.json"), []byte("dirty"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6); err == nil {
		t.Fatal("ResolveGitSourceFile(dirty) error = nil")
	}
	if err := os.WriteFile(filepath.Join(root, "input.json"), []byte("pinned"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "gc", "--prune=now")
	got, err = ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6)
	if err != nil || string(got) != "pinned" {
		t.Fatalf("ResolveGitSourceFile(packed clean) = %q, %v", got, err)
	}
}

func TestGitResolverPublicInputValidation(t *testing.T) {
	validSHA := strings.Repeat("a", 40)
	tests := []struct {
		name string
		call func() error
	}{
		{"repository", func() error {
			_, err := ResolveGitSourceFile("/tmp/root", "invalid", validSHA, "input.json", 1)
			return err
		}},
		{"revision", func() error {
			_, err := ResolveGitSourceFile("/tmp/root", "github.com/example/repo", "bad", "input.json", 1)
			return err
		}},
		{"path", func() error {
			_, err := ResolveGitSourceFile("/tmp/root", "github.com/example/repo", validSHA, "../input.json", 1)
			return err
		}},
		{"maximum", func() error {
			_, err := ResolveGitSourceFile("/tmp/root", "github.com/example/repo", validSHA, "input.json", -1)
			return err
		}},
		{"root-relative", func() error {
			_, err := ResolveGitSourceFile("tmp/root", "github.com/example/repo", validSHA, "input.json", 1)
			return err
		}},
		{"root-unclean", func() error {
			_, err := ResolveGitSourceFile("/tmp/./root", "github.com/example/repo", validSHA, "input.json", 1)
			return err
		}},
		{"release", func() error {
			return VerifyGitReleaseTag("/tmp/root", "github.com/example/repo", "release", validSHA, validSHA)
		}},
		{"tag-object", func() error {
			return VerifyGitReleaseTag("/tmp/root", "github.com/example/repo", "v1.0.0", "bad", validSHA)
		}},
		{"peeled-commit", func() error {
			return VerifyGitReleaseTag("/tmp/root", "github.com/example/repo", "v1.0.0", validSHA, "bad")
		}},
		{"release-root-relative", func() error {
			return VerifyGitReleaseTag("tmp/root", "github.com/example/repo", "v1.0.0", validSHA, validSHA)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestGitResolverPureParsingAndSafety(t *testing.T) {
	validSHA := strings.Repeat("a", 40)
	for _, test := range []struct {
		name  string
		input []byte
		want  string
	}{
		{"commit-valid", []byte("tree " + validSHA + "\n"), validSHA},
		{"commit-missing-line", []byte("author Test\n"), ""},
		{"commit-invalid-id", []byte("tree not-a-sha\n"), ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := commitTreeID(test.input)
			if test.want == "" {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("commitTreeID() = %q, %v", got, err)
			}
		})
	}

	for _, test := range []struct {
		name, input string
		wantErr     bool
	}{
		{"valid", "object " + validSHA + "\ntype commit\ntag v1.0.0\n\nmessage", false},
		{"missing-separator", "object " + validSHA, true},
		{"duplicate", "object " + validSHA + "\nobject " + validSHA + "\ntype commit\ntag v1.0.0\n\n", true},
		{"incomplete", "object " + validSHA + "\ntype commit\n\n", true},
	} {
		t.Run("tag-"+test.name, func(t *testing.T) {
			_, _, _, err := annotatedTagIdentity([]byte(test.input))
			if (err != nil) != test.wantErr {
				t.Fatalf("annotatedTagIdentity() error = %v", err)
			}
		})
	}

	for _, test := range []struct {
		name, input string
		want        string
		wantErr     bool
	}{
		{"valid", "[remote \"origin\"]\n url = https://github.com/example/repo.git\n", "https://github.com/example/repo.git", false},
		{"comments", "; comment\n[remote \"upstream\"]\nurl=x\n", "", true},
		{"duplicate", "[remote \"origin\"]\nurl=x\nurl=y\n", "", true},
		{"missing", "[core]\nurl=x\n", "", true},
	} {
		t.Run("config-"+test.name, func(t *testing.T) {
			got, err := originURLFromGitConfig([]byte(test.input))
			if (err != nil) != test.wantErr || got != test.want {
				t.Fatalf("originURLFromGitConfig() = %q, %v", got, err)
			}
		})
	}

	for _, test := range []struct {
		remote, repo string
		want         bool
	}{
		{"https://github.com/example/repo", "github.com/example/repo", true},
		{"https://github.com/example/repo.git", "github.com/example/repo", true},
		{"git@github.com:example/repo.git", "github.com/example/repo", true},
		{"ssh://git@github.com/example/repo.git", "github.com/example/repo", true},
		{"https://github.com/example/other", "github.com/example/repo", false},
	} {
		t.Run("remote", func(t *testing.T) {
			if got := gitRemoteMatchesRepository(test.remote, test.repo); got != test.want {
				t.Fatalf("match = %v", got)
			}
		})
	}
	for _, test := range []struct {
		name, ref string
		want      bool
	}{
		{"valid", "refs/heads/main", true}, {"tag", "refs/tags/v1", false}, {"dot", "refs/heads/a..b", false}, {"lock", "refs/heads/main.lock", false},
	} {
		t.Run("ref-"+test.name, func(t *testing.T) {
			if got := safeGitRef(test.ref); got != test.want {
				t.Fatalf("safeGitRef() = %v", got)
			}
		})
	}
}

func TestParseGitIndexV2ValidAndMalformed(t *testing.T) {
	object, _ := hex.DecodeString(strings.Repeat("b", 40))
	entry := make([]byte, 0, 80)
	entry = append(entry, make([]byte, 24)...)
	mode := make([]byte, 4)
	binary.BigEndian.PutUint32(mode, 0o100644)
	entry = append(entry, mode...)
	entry = append(entry, make([]byte, 12)...)
	entry = append(entry, object...)
	flags := make([]byte, 2)
	binary.BigEndian.PutUint16(flags, uint16(len("input.json")))
	entry = append(entry, flags...)
	entry = append(entry, []byte("input.json\x00")...)
	for len(entry)%8 != 0 {
		entry = append(entry, 0)
	}
	data := make([]byte, 12)
	copy(data, []byte("DIRC"))
	binary.BigEndian.PutUint32(data[4:8], 2)
	binary.BigEndian.PutUint32(data[8:12], 1)
	data = append(data, entry...)
	digest := sha1.Sum(data)
	data = append(data, digest[:]...)
	got, err := parseGitIndexV2(data)
	if err != nil || got["input.json"].object != strings.Repeat("b", 40) {
		t.Fatalf("valid index = %#v, %v", got, err)
	}
	for _, malformed := range [][]byte{[]byte("bad"), append([]byte(nil), data[:len(data)-1]...)} {
		if _, err := parseGitIndexV2(malformed); err == nil {
			t.Fatal("malformed index accepted")
		}
	}
	if _, err := parseGitIndexV2(bytes.Replace(data, []byte("input.json"), []byte("../x"), 1)); err == nil {
		t.Fatal("unsafe path accepted")
	}
}

func runGitResolverTestCommand(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("/usr/bin/git", append([]string{"-C", directory}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return string(output)
}
