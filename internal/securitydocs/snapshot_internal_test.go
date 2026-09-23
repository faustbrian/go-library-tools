package securitydocs

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidateUsesOneSecurityMatrixSnapshot(t *testing.T) {
	valid := testPassingMatrix()
	first := []byte(strings.Replace(string(valid), `"status":"passed"`, `"status":"failed"`, 1))
	digest := sha256.Sum256(first)
	risk := []byte(fmt.Sprintf(`{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"medium","status":"accepted","owner":"security@example.com","rationale":"risk","mitigation":"bounded","review_condition":"on release","evidence":"sha256:%x:security-matrix.json","expires_at":"2099-10-01T00:00:00Z"}]}`, digest))
	matrixReads := 0
	reader := func(_ *os.Root, path string) ([]byte, error) {
		if path == "risk-register.json" {
			return risk, nil
		}
		matrixReads++
		if matrixReads == 1 {
			return first, nil
		}
		return valid, nil
	}
	err := validateAtWithReader(t.TempDir(), time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC), reader)
	if err == nil || matrixReads != 1 {
		t.Fatalf("validateAtWithReader() = %v, matrix reads = %d", err, matrixReads)
	}
}

func testPassingMatrix() []byte {
	names := []string{"codeql", "dependency-review", "go-vet", "gosec", "govulncheck", "license", "owned-analysis", "secret-current-tree", "secret-history", "staticcheck", "workflow-analysis"}
	rows := make([]string, 0, len(names))
	for _, name := range names {
		rows = append(rows, fmt.Sprintf(`{"name":%q,"tool_version":"test","command":"test","status":"passed","completed_at":"2099-09-01T00:00:00Z","result":"passed","result_sha256":"%s"}`, name, strings.Repeat("a", 64)))
	}
	return []byte(fmt.Sprintf(`{"schema_version":1,"modules":[{"module":"github.com/acme/example","revision":"%s","scanners":[%s],"release_verdict":{"status":"pass","owner":"maintainers","decided_at":"2099-09-01T00:00:00Z","rationale":"verified","residual_risks":["SEC-1"]}}]}`, strings.Repeat("a", 40), strings.Join(rows, ",")))
}
