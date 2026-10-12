package consumers_test

import (
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/consumers"
)

// FuzzConsumerManifest varies repository classification JSON through Parse's 1
// MiB boundary; no rollout or remote mutation is performed.
func FuzzConsumerManifest(f *testing.F) {
	f.Add([]byte(`{"schema_version":1,"owner":"faustbrian","repositories":[{"name":"go-library","classification":"active","default_branch":"main"},{"name":"go-tooling","classification":"tooling","default_branch":"main","reason":"tooling"}]}`))
	f.Add([]byte("{"))
	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = consumers.Parse(data)
	})
}
