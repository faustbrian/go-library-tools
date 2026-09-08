package cohesion

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestV2DurabilityUnknownErrorPreservesCause(t *testing.T) {
	cause := errors.New("sync failed")
	failure := V2DurabilityUnknownError{Cause: cause}
	if got := failure.Error(); got != "cohesion v2 publication committed but durability is unknown" {
		t.Fatalf("Error() = %q", got)
	}
	if !errors.Is(failure, cause) {
		t.Fatal("V2DurabilityUnknownError does not unwrap its cause")
	}
}

func TestPublishProjectV2CreatesImmutableTargetWithoutReplacement(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "catalog.json")
	if err := publishProjectV2([]byte("first"), target); err != nil {
		t.Fatalf("publishProjectV2() error = %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first" || info.Mode().Perm() != 0o444 {
		t.Fatalf("published project = %q mode %04o", got, info.Mode().Perm())
	}

	if err := publishProjectV2([]byte("replacement"), target); !errors.Is(err, ErrV2TargetExists) {
		t.Fatalf("replacement error = %v", err)
	}
	got, err = os.ReadFile(target)
	if err != nil || string(got) != "first" {
		t.Fatalf("target after replacement = %q, %v", got, err)
	}
	matches, err := filepath.Glob(filepath.Join(root, ".catalog.json.golib-stage-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("staging leftovers = %v, %v", matches, err)
	}
}

func TestPublishAggregateV2PublishesExactlyFourImmutableMembersAtomically(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "v2")
	t.Cleanup(func() { _ = os.Chmod(target, 0o700) })
	artifacts := map[string][]byte{
		"catalog-consumer.json":    []byte("consumer-json"),
		"catalog-consumer.md":      []byte("consumer-md"),
		"catalog-engineering.json": []byte("engineering-json"),
		"catalog-engineering.md":   []byte("engineering-md"),
	}
	if err := publishAggregateV2(artifacts, target); err != nil {
		t.Fatalf("publishAggregateV2() error = %v", err)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("aggregate member count = %d", len(entries))
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0o555 {
		t.Fatalf("aggregate mode = %v, %v", info.Mode(), err)
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o444 {
			t.Fatalf("aggregate member %s mode = %v, %v", entry.Name(), info.Mode(), err)
		}
	}
	if err := publishAggregateV2(artifacts, target); !errors.Is(err, ErrV2TargetExists) {
		t.Fatalf("replacement error = %v", err)
	}
}
