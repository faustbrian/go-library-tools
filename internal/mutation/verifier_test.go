package mutation_test

import (
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

func TestLegacyVerifierIdentityIsReproducible(t *testing.T) {
	if got := mutation.LegacyVerifierDigest(); got != "c28707fd4ce35dd228260de564107c1fb7726e42d4e9ea2bf9e988083e1f3b5b" {
		t.Fatalf("LegacyVerifierDigest() = %s", got)
	}
	assets := mutation.VerifierAssets()
	if len(assets) != 6 {
		t.Fatalf("VerifierAssets() count = %d, want 6", len(assets))
	}
	command, ok := assets["scripts/internal/mutation-command.sh"]
	if !ok || len(command) == 0 {
		t.Fatal("VerifierAssets() omitted the mutation command")
	}
	command[0] = 'x'
	if mutation.LegacyVerifierDigest() != "c28707fd4ce35dd228260de564107c1fb7726e42d4e9ea2bf9e988083e1f3b5b" {
		t.Fatal("VerifierAssets() exposed mutable embedded state")
	}
}
