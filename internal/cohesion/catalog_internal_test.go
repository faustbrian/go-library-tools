package cohesion

import "testing"

func TestDocumentationURLPreservesAbsoluteReferences(t *testing.T) {
	const reference = "https://docs.example.test/library"
	if got := documentationURL("github.com/example/library", ".", "v1.0.0", reference); got != reference {
		t.Fatalf("documentationURL() = %q, want %q", got, reference)
	}
}
