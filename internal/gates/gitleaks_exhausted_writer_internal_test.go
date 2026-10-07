package gates

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestBoundedSourceFileExhaustedBudgetPrecedesDestinationFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	writer := &boundedSourceFile{file: file, limit: 1}
	if n, err := io.WriteString(writer, "a"); err != nil || n != 1 {
		t.Fatalf("inclusive first write = %d, %v", n, err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	n, err := io.WriteString(writer, "b")
	if n != 1 || err == nil || err.Error() != "history bundle byte limit exceeded" || !writer.didOverflow() || writer.written != 1 {
		t.Fatalf("exhausted budget = %d, %v, overflow %t, written %d", n, err, writer.didOverflow(), writer.written)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "a" {
		t.Fatalf("persisted exhausted prefix = %q, %v", data, err)
	}
}
