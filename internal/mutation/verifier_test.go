package mutation_test

import (
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

func TestLegacyVerifierIdentityIsReproducible(t *testing.T) {
	if got := mutation.LegacyVerifierDigest(); got != "390cf4085a0063fe5dcaa29ce942ea4cb8c7876347da90376b65dc8c5c811104" {
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
	if mutation.LegacyVerifierDigest() != "390cf4085a0063fe5dcaa29ce942ea4cb8c7876347da90376b65dc8c5c811104" {
		t.Fatal("VerifierAssets() exposed mutable embedded state")
	}
}
