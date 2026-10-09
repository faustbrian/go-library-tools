package mutation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOrdinaryEquivalentInventoryTypedDecodeIsCategorical(t *testing.T) {
	got, err := ParseEquivalentInventory(strings.NewReader("false"))
	if !errors.Is(err, ErrInvalid) || got.SchemaVersion != 0 || got.Packages != nil {
		t.Fatalf("scalar inventory = %#v, %v; want zero and ErrInvalid", got, err)
	}
	if err.Error() != "invalid mutation evidence: decode failure" {
		t.Fatalf("typed decode diagnostic = %q; want categorical decode failure", err.Error())
	}
}

func TestOrdinaryPackageInputsRefusePostListingSourceLink(t *testing.T) {
	campaign, _ := campaignFixture(t)
	listingCalls := 0
	campaign.Process = func(_ context.Context, name string, args []string, directory string, _ map[string]string, stdout, _ io.Writer) error {
		if name != "go" || len(args) == 0 || args[0] != "list" {
			t.Fatal("unexpected callback; no executable invocation is permitted")
		}
		listingCalls++
		if _, err := fmt.Fprintf(stdout, `{"Dir":%q,"ImportPath":"example","GoFiles":["source.go"],"Module":{"Path":"example","Main":true,"GoVersion":"1.27.0"}}`, directory); err != nil {
			return err
		}
		return os.Symlink("source.go", filepath.Join(campaign.Root, "metadata.link"))
	}
	got, err := campaign.packageInputsForVerifiers(context.Background(), ".", LegacyVerifierDigest())
	if got != nil || !errors.Is(err, ErrInvalid) || listingCalls != 1 {
		t.Fatalf("post-list source admission = %#v, %v, callbacks %d; want nil and ErrInvalid", got, err, listingCalls)
	}
	if err.Error() != "invalid mutation evidence: symlink in mutation source: metadata.link" {
		t.Fatalf("final source diagnostic = %q; want source-link refusal", err.Error())
	}
}

func TestOrdinaryMutationOutputCapacityPreservesCompleteness(t *testing.T) {
	const capacity = 64 * 1024
	suffix := "\nNo results to report.\n"
	complete := strings.Repeat("a", capacity-len(suffix)) + suffix
	var output boundedMutationOutput
	written, err := output.Write([]byte(complete))
	if err != nil || written != capacity || output.buffer.Len() != capacity || output.buffer.String() != complete || output.overflow || !output.confirmsNoResults() {
		t.Fatal("complete output at capacity must retain all bytes and accept its no-results signal")
	}
	written, err = output.Write([]byte("x"))
	if err != nil || written != 1 || output.buffer.Len() != capacity || output.buffer.String() != complete || !output.overflow || output.confirmsNoResults() {
		t.Fatal("one extra acknowledged byte must preserve retained bytes but invalidate completeness")
	}
}

func TestOrdinaryMutationOutputZeroWriteAtCapacityCannotCertifyCompleteness(t *testing.T) {
	var output boundedMutationOutput
	if written, err := output.Write(bytes.Repeat([]byte("a"), maximumMutationToolOutput)); err != nil || written != maximumMutationToolOutput {
		t.Fatalf("capacity write = %d, %v", written, err)
	}
	if written, err := output.Write(nil); err != nil || written != 0 || !output.overflow || output.confirmsNoResults() {
		t.Fatalf("zero write at capacity = %d, %v, overflow %v; want retained incomplete state", written, err, output.overflow)
	}
}
