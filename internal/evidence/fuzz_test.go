package evidence_test

import (
	"bytes"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/evidence"
)

// FuzzEvidence caps candidate JSON at MaximumSize+1 and exercises strict record
// parsing, not authenticated result reuse or persistence.
func FuzzEvidence(f *testing.F) {
	f.Add([]byte(`{"schema_version":1}`))
	f.Add([]byte(`{"schema_version":1,"repository":"example","module":".","gate":"tests","input_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","verifier_digest":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","report_digest":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","result":"passed","completed_at":"2026-01-01T00:00:00Z","environment":{}}`))
	f.Add([]byte(`{"schema_version":1}{"schema_version":1}`))
	f.Fuzz(func(_ *testing.T, input []byte) {
		if len(input) > evidence.MaximumSize+1 {
			return
		}
		_, _ = evidence.Parse(bytes.NewReader(input))
	})
}
