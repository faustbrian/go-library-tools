package securitydocs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedSecuritySchemasMatchPublishedFiles(t *testing.T) {
	for _, test := range []struct{ name, embedded string }{
		{"security-risk-register.schema.json", riskRegisterSchema},
		{"security-matrix.schema.json", securityMatrixSchema},
	} {
		published, err := os.ReadFile(filepath.Join("..", "..", "schema", test.name))
		if err != nil {
			t.Fatal(err)
		}
		if string(published) != test.embedded {
			t.Fatalf("embedded %s differs from the published schema; run go generate ./internal/securitydocs", test.name)
		}
	}
}

func TestJSONWalkHasExplicitNestingLimit(t *testing.T) {
	for _, test := range []struct {
		depth int
		valid bool
	}{{128, true}, {129, false}} {
		value := strings.Repeat("[", test.depth) + "0" + strings.Repeat("]", test.depth)
		decoder := json.NewDecoder(strings.NewReader(value))
		if err := walkJSONValue(decoder, 0); (err == nil) != test.valid {
			t.Fatalf("depth %d: walkJSONValue() error = %v, want valid=%t", test.depth, err, test.valid)
		}
	}
}

func TestAcceptedRiskExpiryMustFollowDecisionTime(t *testing.T) {
	now := time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name, expiry string
		valid        bool
	}{
		{"one second before", "2026-09-22T11:59:59Z", false},
		{"exactly now", "2026-09-22T12:00:00Z", false},
		{"one second after", "2026-09-22T12:00:01Z", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			risks := `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"low","status":"accepted","owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release","evidence":"artifact://review","expires_at":"` + test.expiry + `"}]}`
			if err := os.WriteFile(filepath.Join(root, "risk-register.json"), []byte(risks), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "security-matrix.json"), []byte(`{"schema_version":1,"modules":[]}`), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := validateAt(root, now); (err == nil) != test.valid {
				t.Fatalf("validateAt() error = %v, want valid=%t", err, test.valid)
			}
		})
	}
}
