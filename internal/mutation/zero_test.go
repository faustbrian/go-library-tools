package mutation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestParseZeroInventoryMatchesExactReviewedIdentity(t *testing.T) {
	digest := strings.Repeat("a", 64)
	verifier := strings.Repeat("b", 64)
	inventory, err := mutation.ParseZeroInventory(strings.NewReader(`{
  "schema_version": 1,
  "packages": [{
    "module_directory": ".",
    "package_directory": "adapter",
    "source_digest": "` + digest + `",
    "gremlins_version": "v0.6.0",
    "gremlins_verifier_sha256": "` + verifier + `",
    "reason": "The package contains declarations only and the complete operator set discovers no viable mutations."
  }]
}`))
	if err != nil {
		t.Fatalf("ParseZeroInventory() error = %v", err)
	}
	if !inventory.Reviewed(".", "adapter", digest, "v0.6.0", verifier) {
		t.Fatal("Reviewed() = false, want true")
	}
	review, reviewed := inventory.Review(".", "adapter", digest, "v0.6.0", verifier)
	if !reviewed || review == nil || review.Reason == "" {
		t.Fatalf("Review() = %#v, %v", review, reviewed)
	}
	if inventory.Reviewed(".", "adapter", strings.Repeat("c", 64), "v0.6.0", verifier) {
		t.Fatal("Reviewed() accepted changed source")
	}
	if review, reviewed := inventory.Review(".", "adapter", strings.Repeat("c", 64), "v0.6.0", verifier); reviewed || review != nil {
		t.Fatalf("Review() near match = %#v, %v", review, reviewed)
	}
}

func TestParseZeroInventoryRejectsInvalidPolicy(t *testing.T) {
	tests := map[string]string{
		"unknown field":       `{"schema_version":1,"packages":[],"unknown":true}`,
		"wrong schema":        `{"schema_version":2,"packages":[]}`,
		"unsafe path":         reviewJSON("../outside", ".", strings.Repeat("a", 64), strings.Repeat("b", 64), "v0.6.0", strings.Repeat("reason ", 8)),
		"bad source digest":   reviewJSON(".", ".", "bad", strings.Repeat("b", 64), "v0.6.0", strings.Repeat("reason ", 8)),
		"bad verifier digest": reviewJSON(".", ".", strings.Repeat("a", 64), "bad", "v0.6.0", strings.Repeat("reason ", 8)),
		"bad version":         reviewJSON(".", ".", strings.Repeat("a", 64), strings.Repeat("b", 64), "latest", strings.Repeat("reason ", 8)),
		"weak reason":         reviewJSON(".", ".", strings.Repeat("a", 64), strings.Repeat("b", 64), "v0.6.0", "none"),
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := mutation.ParseZeroInventory(strings.NewReader(body))
			if !errors.Is(err, mutation.ErrInvalid) {
				t.Fatalf("ParseZeroInventory() error = %v, want ErrInvalid", err)
			}
		})
	}
}

func reviewJSON(module, pkg, source, verifier, version, reason string) string {
	return `{"schema_version":1,"packages":[{"module_directory":"` + module + `","package_directory":"` + pkg + `","source_digest":"` + source + `","gremlins_version":"` + version + `","gremlins_verifier_sha256":"` + verifier + `","reason":"` + reason + `"}]}`
}
