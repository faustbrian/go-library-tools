package gates

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Keep this finite direct-reader oracle ahead of copier tests: failfast must
// expose a bad cancellation guard before a copier can repeat a zero-byte read.
func TestContextSourceReaderStopsBeforeReading(t *testing.T) {
	source := strings.NewReader("abc")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	buffer := make([]byte, 3)
	if n, err := (contextSourceReader{ctx: ctx, reader: source}).Read(buffer); n != 0 || !errors.Is(err, context.Canceled) || source.Len() != 3 {
		t.Fatal("cancelled read consumed source bytes or lost cancellation")
	}
	if n, err := (contextSourceReader{ctx: t.Context(), reader: source}).Read(buffer); n != 3 || err != nil || string(buffer) != "abc" {
		t.Fatal("active read did not preserve source bytes")
	}
}
