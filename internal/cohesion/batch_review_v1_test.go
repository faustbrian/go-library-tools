package cohesion

import "testing"

func TestVerifySchemaV3BatchReviewRequiresCompleteManifestExpectation(t *testing.T) {
	if err := verifySchemaV3BatchReview([]byte(`{}`), "sha256:"+string(make([]byte, 64)), nil); err == nil {
		t.Fatal("incomplete batch expectation accepted")
	}
}
