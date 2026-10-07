package gates

import (
	"strings"
	"testing"
)

func TestSecretMetadataPartialTailSuppressesCompletedLocation(t *testing.T) {
	location := strings.Repeat("a", 64) + " " + strings.Repeat("b", 64) + " 12 -"
	var metadata secretMetadata
	metadata.write([]byte(location + "\n"))
	if got := metadata.summary(); got != "\nsecret-location "+location {
		t.Fatal("complete ordinary location was not admitted")
	}
	metadata.write([]byte("unfinished private marker"))
	if got := metadata.summary(); got != "" {
		t.Fatalf("completed location escaped an incomplete report: %q", got)
	}
	if metadata.length != 0 || metadata.line != [256]byte{} {
		t.Fatal("pending incomplete metadata was not cleared")
	}
}
