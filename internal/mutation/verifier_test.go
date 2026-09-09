package mutation_test

import (
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestLegacyVerifierIdentityIsReproducible(t *testing.T) {
	if got := mutation.LegacyVerifierDigest(); got != "5eba124f842305aecef71fc439aea0f1056429b253fad3b2ed47468f614595a3" {
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
	if mutation.LegacyVerifierDigest() != "5eba124f842305aecef71fc439aea0f1056429b253fad3b2ed47468f614595a3" {
		t.Fatal("VerifierAssets() exposed mutable embedded state")
	}
}
