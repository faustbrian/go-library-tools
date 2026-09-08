package cohesion

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseResolutionMapV1AcceptsCanonicalClosedMap(t *testing.T) {
	input := []byte(`{"sources":[{"repository":"github.com/faustbrian/example","source_revision":"0123456789abcdef0123456789abcdef01234567","root":"/srv/sources/example"}],"releases":[{"repository":"github.com/faustbrian/go-library-tools","release":"v1.5.6","tag_object_sha":"1111111111111111111111111111111111111111","peeled_commit":"2222222222222222222222222222222222222222","root":"/srv/releases/tooling"}],"toolchains":[{"version":"go1.26.6","goos":"linux","goarch":"amd64","distribution_sha256":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","tree_sha256":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","binary_sha256":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","archive":"/srv/toolchains/go1.26.6.tar.gz","root":"/srv/toolchains/go1.26.6"}],"module_proxy":{"tree_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","root":"/srv/proxy"}}`)

	got, err := ParseResolutionMapV1(input)
	if err != nil {
		t.Fatalf("ParseResolutionMapV1() error = %v", err)
	}
	if len(got.Sources) != 1 || got.Sources[0].Repository != "github.com/faustbrian/example" || got.Sources[0].Root != "/srv/sources/example" {
		t.Fatalf("sources = %#v", got.Sources)
	}
	if len(got.Releases) != 1 || len(got.Toolchains) != 1 || got.ModuleProxy.Root == nil || *got.ModuleProxy.Root != "/srv/proxy" {
		t.Fatalf("map = %#v", got)
	}
}

func TestParseResolutionMapV1UsesItsExactSourceCollectionLimit(t *testing.T) {
	build := func(count int) []byte {
		var input strings.Builder
		input.WriteString(`{"sources":[`)
		for index := range count {
			if index != 0 {
				input.WriteByte(',')
			}
			_, _ = fmt.Fprintf(&input, `{"repository":"github.com/faustbrian/r%05d","source_revision":"%040x","root":"/r/%05d"}`, index, index+1, index)
		}
		input.WriteString(`],"releases":[],"toolchains":[],"module_proxy":{"tree_sha256":null,"root":null}}`)
		return []byte(input.String())
	}

	if got, err := ParseResolutionMapV1(build(4097)); err != nil || len(got.Sources) != 4097 {
		t.Fatalf("ParseResolutionMapV1(4097 sources) count = %d, error = %v", len(got.Sources), err)
	}
	if _, err := ParseResolutionMapV1(build(16385)); err == nil || !strings.Contains(err.Error(), "more than 16384 elements") {
		t.Fatalf("ParseResolutionMapV1(16385 sources) error = %v", err)
	}
}

func TestParseResolutionMapV1RejectsUntrustedShapesBeforeResolution(t *testing.T) {
	valid := `{"sources":[],"releases":[],"toolchains":[],"module_proxy":{"tree_sha256":null,"root":null}}`
	tests := map[string]string{
		"unknown top-level member":  strings.Replace(valid, `"module_proxy"`, `"unexpected":true,"module_proxy"`, 1),
		"relative source root":      `{"sources":[{"repository":"github.com/faustbrian/example","source_revision":"0123456789abcdef0123456789abcdef01234567","root":"relative"}],"releases":[],"toolchains":[],"module_proxy":{"tree_sha256":null,"root":null}}`,
		"duplicate source identity": `{"sources":[{"repository":"github.com/faustbrian/example","source_revision":"0123456789abcdef0123456789abcdef01234567","root":"/a"},{"repository":"github.com/faustbrian/example","source_revision":"0123456789abcdef0123456789abcdef01234567","root":"/b"}],"releases":[],"toolchains":[],"module_proxy":{"tree_sha256":null,"root":null}}`,
		"unsorted source rows":      `{"sources":[{"repository":"github.com/faustbrian/z","source_revision":"0123456789abcdef0123456789abcdef01234567","root":"/z"},{"repository":"github.com/faustbrian/a","source_revision":"0123456789abcdef0123456789abcdef01234567","root":"/a"}],"releases":[],"toolchains":[],"module_proxy":{"tree_sha256":null,"root":null}}`,
		"half-null proxy":           fmt.Sprintf(`{"sources":[],"releases":[],"toolchains":[],"module_proxy":{"tree_sha256":null,"root":%q}}`, "/proxy"),
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseResolutionMapV1([]byte(input)); err == nil {
				t.Fatal("ParseResolutionMapV1() error = nil")
			}
		})
	}
}
