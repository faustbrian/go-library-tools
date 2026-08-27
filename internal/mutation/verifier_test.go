package mutation_test

import (
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestLegacyVerifierIdentityIsReproducible(t *testing.T) {
	if got := mutation.LegacyVerifierDigest(); got != "9a9499ff68a8dfd49a0be7995590297a8ee563a1aa226bfd8b9361dc53058108" {
		t.Fatalf("LegacyVerifierDigest() = %s", got)
	}
	assets := mutation.VerifierAssets()
	if len(assets) != 5 {
		t.Fatalf("VerifierAssets() count = %d, want 5", len(assets))
	}
	command, ok := assets["scripts/internal/mutation-command.sh"]
	if !ok || len(command) == 0 {
		t.Fatal("VerifierAssets() omitted the mutation command")
	}
	command[0] = 'x'
	if mutation.LegacyVerifierDigest() != "9a9499ff68a8dfd49a0be7995590297a8ee563a1aa226bfd8b9361dc53058108" {
		t.Fatal("VerifierAssets() exposed mutable embedded state")
	}
}
