package securitydocs_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/securitydocs"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestValidateBindsResolvedRiskEvidenceToExactMatrixBytes(t *testing.T) {
	now := "2099-10-01T00:00:00Z"
	matrixDocument := matrix(scanners("passed"), `["SEC-1"]`)
	digest := sha256.Sum256([]byte(matrixDocument))
	validEvidence := fmt.Sprintf("sha256:%x:security-matrix.json", digest)
	for _, test := range []struct {
		name, evidence string
		valid          bool
	}{
		{"exact raw matrix digest", validEvidence, true},
		{"wrong digest", "sha256:" + strings.Repeat("0", 64) + ":security-matrix.json", false},
		{"mutable path", "sha256:" + strings.Repeat("0", 64) + ":./security-matrix.json", false},
		{"legacy pointer", "artifact://review", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			risk := fmt.Sprintf(`{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"medium","status":"accepted","owner":"security@example.com","rationale":"risk","mitigation":"bounded","review_condition":"on release","evidence":%q,"expires_at":%q}]}`, test.evidence, now)
			write(t, filepath.Join(root, "risk-register.json"), risk)
			write(t, filepath.Join(root, "security-matrix.json"), matrixDocument)
			if err := securitydocs.Validate(root); (err == nil) != test.valid {
				t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
			}
		})
	}
}

func TestValidateRequiresOwnedRisksAndConsistentReleaseVerdicts(t *testing.T) {
	validScanners := scanners("passed")
	validMatrix := matrix(validScanners, "[]")
	validRisk := fmt.Sprintf(`{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"medium","status":"mitigated","owner":"security@example.com","rationale":"bounded parser","mitigation":"entry limits","review_condition":"parser or limit changes","evidence":%q}]}`, matrixEvidence(validMatrix))
	tests := []struct{ name, risks, matrix, want string }{
		{"valid", validRisk, validMatrix, ""},
		{"missing risk review", `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"medium","status":"open","owner":"security@example.com","rationale":"risk","mitigation":"pending","review_condition":""}]}`, `{"schema_version":1,"modules":[]}`, "published schema"},
		{"passing verdict with failed scanner", `{"schema_version":1,"risks":[]}`, matrix(scanners("failed"), "[]"), "published schema"},
		{"passing verdict with inapplicable scanner", `{"schema_version":1,"risks":[]}`, matrix(scanners("not-applicable"), "[]"), "published schema"},
		{"passing verdict with incomplete scanners", `{"schema_version":1,"risks":[]}`, matrix(scanner("gosec", "passed"), "[]"), "published schema"},
		{"accepted risk without evidence", `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"medium","status":"accepted","owner":"security@example.com","rationale":"risk","mitigation":"bounded deployment","review_condition":"on boundary change"}]}`, `{"schema_version":1,"modules":[]}`, "published schema"},
		{"expired accepted risk", `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"medium","status":"accepted","owner":"security@example.com","rationale":"risk","mitigation":"bounded deployment","review_condition":"on boundary change","evidence":"artifact://review","expires_at":"2020-01-01T00:00:00Z"}]}`, `{"schema_version":1,"modules":[]}`, "has expired"},
		{"passing verdict with accepted high", `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"high","status":"accepted","owner":"security@example.com","rationale":"risk","mitigation":"bounded deployment","review_condition":"on boundary change","evidence":"artifact://review","expires_at":"2099-10-01T00:00:00Z"}]}`, matrix(validScanners, `["SEC-1"]`), "unresolved critical or high"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "risk-register.json"), test.risks)
			write(t, filepath.Join(root, "security-matrix.json"), test.matrix)
			err := securitydocs.Validate(root)
			if test.want == "" && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if test.want != "" && (err == nil || !strings.Contains(err.Error(), test.want)) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateRejectsNullRiskArray(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":null}`)
	write(t, filepath.Join(root, "security-matrix.json"), `{"schema_version":1,"modules":[]}`)
	if err := securitydocs.Validate(root); err == nil {
		t.Fatal("Validate accepted a null risks array rejected by the published schema")
	}
}

func TestValidateRejectsSchemaInvalidRecords(t *testing.T) {
	validRisks := `{"schema_version":1,"risks":[]}`
	validMatrix := `{"schema_version":1,"modules":[]}`
	tests := []struct {
		name, risks, matrix string
	}{
		{"missing risks", `{"schema_version":1}`, validMatrix},
		{"missing modules", validRisks, `{"schema_version":1}`},
		{"null modules", validRisks, `{"schema_version":1,"modules":null}`},
		{"invalid risk identifier", `{"schema_version":1,"risks":[{"id":"invalid","module":"github.com/acme/example","severity":"low","status":"open","owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release"}]}`, validMatrix},
		{"null residual risks", validRisks, matrix(scanners("passed"), "null")},
		{"missing residual risks", validRisks, strings.Replace(matrix(scanners("passed"), "[]"), `,"residual_risks":[]`, "", 1)},
		{"null scanners on blocked verdict", validRisks, strings.Replace(strings.Replace(matrix("", "[]"), `"scanners":[]`, `"scanners":null`, 1), `"status":"pass"`, `"status":"blocked"`, 1)},
		{"missing scanners on blocked verdict", validRisks, strings.Replace(strings.Replace(matrix("", "[]"), `,"scanners":[]`, "", 1), `"status":"pass"`, `"status":"blocked"`, 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "risk-register.json"), test.risks)
			write(t, filepath.Join(root, "security-matrix.json"), test.matrix)
			if err := securitydocs.Validate(root); err == nil {
				t.Fatal("Validate accepted a record rejected by the published schemas")
			}
		})
	}
}

func TestValidateRequiresResidualRisksForNonpassingVerdicts(t *testing.T) {
	for _, status := range []string{"fail", "blocked"} {
		for _, test := range []struct {
			name, residual string
			valid          bool
		}{
			{"present empty", "[]", true},
			{"null", "null", false},
			{"missing", "missing", false},
		} {
			t.Run(status+"/"+test.name, func(t *testing.T) {
				root := t.TempDir()
				write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[]}`)
				value := strings.Replace(matrix("", "[]"), `"status":"pass"`, `"status":"`+status+`"`, 1)
				switch test.residual {
				case "null":
					value = strings.Replace(value, `"residual_risks":[]`, `"residual_risks":null`, 1)
				case "missing":
					value = strings.Replace(value, `,"residual_risks":[]`, "", 1)
				}
				write(t, filepath.Join(root, "security-matrix.json"), value)
				if err := securitydocs.Validate(root); (err == nil) != test.valid {
					t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
				}
			})
		}
	}
}

func TestValidateRequiresScannersForNonpassingVerdicts(t *testing.T) {
	for _, status := range []string{"fail", "blocked"} {
		for _, test := range []struct {
			name, scannerRows string
			valid             bool
		}{
			{"present empty", "[]", true},
			{"null", "null", false},
			{"missing", "missing", false},
			{"twelve", "[" + scanners("passed") + "," + scanner("codeql", "passed") + "]", false},
		} {
			t.Run(status+"/"+test.name, func(t *testing.T) {
				root := t.TempDir()
				write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[]}`)
				value := strings.Replace(matrix("", "[]"), `"status":"pass"`, `"status":"`+status+`"`, 1)
				switch test.scannerRows {
				case "null":
					value = strings.Replace(value, `"scanners":[]`, `"scanners":null`, 1)
				case "missing":
					value = strings.Replace(value, `,"scanners":[]`, "", 1)
				default:
					value = strings.Replace(value, `"scanners":[]`, `"scanners":`+test.scannerRows, 1)
				}
				write(t, filepath.Join(root, "security-matrix.json"), value)
				err := securitydocs.Validate(root)
				if (err == nil) != test.valid {
					t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
				}
				if !test.valid && !strings.Contains(err.Error(), "published schema") {
					t.Fatalf("Validate() error = %v, want published schema rejection", err)
				}
			})
		}
	}
}

func TestValidateRejectsDuplicateJSONKeys(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[],"risks":[]}`)
	write(t, filepath.Join(root, "security-matrix.json"), `{"schema_version":1,"modules":[]}`)
	if err := securitydocs.Validate(root); err == nil {
		t.Fatal("Validate accepted ambiguous duplicate JSON member keys")
	}
}

func TestValidateAcceptsEquivalentSchemaVersions(t *testing.T) {
	for _, test := range []struct {
		version string
		valid   bool
	}{
		{"1", true}, {"1.0", true}, {"1e0", true}, {"1E+0", true},
		{"0", false}, {"2", false}, {"1.1", false}, {`"1"`, false},
		{"true", false}, {"null", false},
	} {
		t.Run(test.version, func(t *testing.T) {
			root := t.TempDir()
			schemaValid := true
			for name, value := range map[string]string{
				"risk-register.json":   `{"schema_version":` + test.version + `,"risks":[]}`,
				"security-matrix.json": `{"schema_version":` + test.version + `,"modules":[]}`,
			} {
				write(t, filepath.Join(root, name), value)
				schemaName := "security-risk-register.schema.json"
				if name == "security-matrix.json" {
					schemaName = "security-matrix.schema.json"
				}
				schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "schema", schemaName))
				if err != nil {
					t.Fatal(err)
				}
				schemaDocument, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
				if err != nil {
					t.Fatal(err)
				}
				compiler := jsonschema.NewCompiler()
				if err := compiler.AddResource("schema.json", schemaDocument); err != nil {
					t.Fatal(err)
				}
				compiled, err := compiler.Compile("schema.json")
				if err != nil {
					t.Fatal(err)
				}
				instance, err := jsonschema.UnmarshalJSON(strings.NewReader(value))
				if err != nil {
					t.Fatal(err)
				}
				if err := compiled.Validate(instance); err != nil {
					schemaValid = false
				}
			}
			if schemaValid != test.valid {
				t.Fatalf("published schema classification = %t, want %t", schemaValid, test.valid)
			}
			if err := securitydocs.Validate(root); (err == nil) != test.valid {
				t.Fatalf("Validate(version=%s) error = %v, want valid=%t", test.version, err, test.valid)
			}
		})
	}
}

func TestValidateChecksEachSchemaVersionIndependently(t *testing.T) {
	for _, name := range []string{"risk-register.json", "security-matrix.json"} {
		for _, test := range []struct {
			version string
			valid   bool
		}{{"1.0", true}, {"1e0", true}, {"2", false}} {
			t.Run(name+"/"+test.version, func(t *testing.T) {
				root := t.TempDir()
				risks := `{"schema_version":1,"risks":[]}`
				modules := `{"schema_version":1,"modules":[]}`
				if name == "risk-register.json" {
					risks = strings.Replace(risks, `"schema_version":1`, `"schema_version":`+test.version, 1)
				} else {
					modules = strings.Replace(modules, `"schema_version":1`, `"schema_version":`+test.version, 1)
				}
				write(t, filepath.Join(root, "risk-register.json"), risks)
				write(t, filepath.Join(root, "security-matrix.json"), modules)
				if err := securitydocs.Validate(root); (err == nil) != test.valid {
					t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
				}
			})
		}
	}
}

func TestValidateAcceptsLargeExactNumericRepresentationOfVersionOne(t *testing.T) {
	root := t.TempDir()
	version := "1" + strings.Repeat("0", 400) + "e-400"
	write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":`+version+`,"risks":[]}`)
	write(t, filepath.Join(root, "security-matrix.json"), `{"schema_version":`+version+`,"modules":[]}`)
	if err := securitydocs.Validate(root); err != nil {
		t.Fatalf("Validate() rejected a finite JSON number mathematically equal to one: %v", err)
	}
}

func TestValidateMatchesPublishedStructuralSchemas(t *testing.T) {
	validRisks := `{"schema_version":1,"risks":[]}`
	validMatrix := `{"schema_version":1,"modules":[]}`
	blocked := strings.Replace(matrix("", "[]"), `"status":"pass"`, `"status":"blocked"`, 1)
	validRisk := `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"low","status":"open","owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release"}]}`
	tests := []struct {
		name, risks, matrix string
		valid               bool
	}{
		{"empty arrays", validRisks, validMatrix, true},
		{"blocked without scanners", validRisks, blocked, true},
		{"missing risk array", `{"schema_version":1}`, validMatrix, false},
		{"null risk array", `{"schema_version":1,"risks":null}`, validMatrix, false},
		{"malformed risk id", `{"schema_version":1,"risks":[{"id":"invalid","module":"github.com/acme/example","severity":"low","status":"open","owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release"}]}`, validMatrix, false},
		{"missing risk owner", strings.Replace(validRisk, `,"owner":"security@example.com"`, "", 1), validMatrix, false},
		{"null risk owner", strings.Replace(validRisk, `"owner":"security@example.com"`, `"owner":null`, 1), validMatrix, false},
		{"wrong risk severity", strings.Replace(validRisk, `"severity":"low"`, `"severity":"urgent"`, 1), validMatrix, false},
		{"extra risk property", strings.Replace(validRisk, `"review_condition":"on release"`, `"review_condition":"on release","unknown":1`, 1), validMatrix, false},
		{"empty optional evidence", `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"low","status":"open","owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release","evidence":""}]}`, validMatrix, false},
		{"valid risk record", validRisk, validMatrix, true},
		{"missing matrix array", validRisks, `{"schema_version":1}`, false},
		{"null matrix array", validRisks, `{"schema_version":1,"modules":null}`, false},
		{"null scanners", validRisks, strings.Replace(blocked, `"scanners":[]`, `"scanners":null`, 1), false},
		{"missing scanners", validRisks, strings.Replace(blocked, `,"scanners":[]`, "", 1), false},
		{"null residual risks", validRisks, matrix(scanners("passed"), "null"), false},
		{"missing residual risks", validRisks, strings.Replace(matrix(scanners("passed"), "[]"), `,"residual_risks":[]`, "", 1), false},
		{"invalid revision", validRisks, strings.Replace(blocked, `"revision":"0123456789abcdef0123456789abcdef01234567"`, `"revision":"bad"`, 1), false},
		{"extra verdict property", validRisks, strings.Replace(blocked, `"residual_risks":[]`, `"residual_risks":[],"unknown":1`, 1), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			schemaValid := true
			for name, value := range map[string]string{"risk-register.json": test.risks, "security-matrix.json": test.matrix} {
				write(t, filepath.Join(root, name), value)
				schemaName := "security-risk-register.schema.json"
				if name == "security-matrix.json" {
					schemaName = "security-matrix.schema.json"
				}
				schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "schema", schemaName))
				if err != nil {
					t.Fatal(err)
				}
				document, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
				if err != nil {
					t.Fatal(err)
				}
				compiler := jsonschema.NewCompiler()
				if err := compiler.AddResource("schema.json", document); err != nil {
					t.Fatal(err)
				}
				compiled, err := compiler.Compile("schema.json")
				if err != nil {
					t.Fatal(err)
				}
				instance, err := jsonschema.UnmarshalJSON(strings.NewReader(value))
				if err != nil {
					t.Fatal(err)
				}
				if err := compiled.Validate(instance); err != nil {
					schemaValid = false
				}
			}
			if schemaValid != test.valid {
				t.Fatalf("published schema classification = %t, want %t", schemaValid, test.valid)
			}
			if err := securitydocs.Validate(root); (err == nil) != test.valid {
				t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
			}
		})
	}
}

func TestValidateBoundsBothSecurityDocuments(t *testing.T) {
	baseRisks := `{"schema_version":1,"risks":[]}`
	baseMatrix := `{"schema_version":1,"modules":[]}`
	for _, name := range []string{"risk-register.json", "security-matrix.json"} {
		for _, size := range []int{4<<20 - 1, 4 << 20, 4<<20 + 1} {
			t.Run(fmt.Sprintf("%s/%d", name, size), func(t *testing.T) {
				root := t.TempDir()
				write(t, filepath.Join(root, "risk-register.json"), baseRisks)
				write(t, filepath.Join(root, "security-matrix.json"), baseMatrix)
				base := baseRisks
				if name == "security-matrix.json" {
					base = baseMatrix
				}
				value := base + strings.Repeat(" ", size-len(base))
				write(t, filepath.Join(root, name), value)
				err := securitydocs.Validate(root)
				if (err == nil) != (size <= 4<<20) {
					t.Fatalf("Validate(size=%d) error = %v", size, err)
				}
				stored, readErr := os.ReadFile(filepath.Join(root, name))
				if readErr != nil || string(stored) != value {
					t.Fatalf("Validate modified %s: %v", name, readErr)
				}
			})
		}
	}
}

func TestValidateRejectsNestedDuplicateKeys(t *testing.T) {
	baseRisks := `{"schema_version":1,"risks":[]}`
	baseMatrix := matrix(scanners("passed"), "[]")
	risk := `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"low","status":"open","owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release"}]}`
	for _, test := range []struct{ name, risks, matrix string }{
		{"root version", strings.Replace(baseRisks, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1), baseMatrix},
		{"nested risk id", strings.Replace(risk, `"id":"SEC-1"`, `"id":"SEC-1","id":"SEC-1"`, 1), baseMatrix},
		{"nested risk module", strings.Replace(risk, `"module":"github.com/acme/example"`, `"module":"github.com/acme/example","module":"github.com/acme/example"`, 1), baseMatrix},
		{"nested risk severity", strings.Replace(risk, `"severity":"low"`, `"severity":"high","severity":"low"`, 1), baseMatrix},
		{"module identity", baseRisks, strings.Replace(baseMatrix, `"module":"github.com/acme/example"`, `"module":"github.com/acme/example","module":"github.com/acme/example"`, 1)},
		{"scanner identity", baseRisks, strings.Replace(baseMatrix, `"name":"codeql"`, `"name":"gosec","name":"codeql"`, 1)},
		{"scanner status", baseRisks, strings.Replace(baseMatrix, `"status":"passed"`, `"status":"passed","status":"passed"`, 1)},
		{"verdict status", baseRisks, strings.Replace(baseMatrix, `"status":"pass"`, `"status":"pass","status":"pass"`, 1)},
		{"verdict conflicting last pass", baseRisks, strings.Replace(baseMatrix, `"status":"pass"`, `"status":"fail","status":"pass"`, 1)},
		{"verdict conflicting first pass", baseRisks, strings.Replace(baseMatrix, `"status":"pass"`, `"status":"pass","status":"fail"`, 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "risk-register.json"), test.risks)
			write(t, filepath.Join(root, "security-matrix.json"), test.matrix)
			if err := securitydocs.Validate(root); err == nil {
				t.Fatal("Validate accepted duplicate JSON member keys")
			}
		})
	}
}

func TestValidateDistinguishesPassingAndNonpassingScannerRows(t *testing.T) {
	complete := scanners("passed")
	ten := strings.TrimSuffix(complete, ","+scanner("workflow-analysis", "passed"))
	for _, test := range []struct {
		name, status, scannerRows string
		valid                     bool
	}{
		{"pass complete", "pass", complete, true},
		{"pass empty", "pass", "", false},
		{"pass partial", "pass", scanner("gosec", "passed"), false},
		{"pass ten", "pass", ten, false},
		{"pass failed", "pass", scanners("failed"), false},
		{"pass failed middle", "pass", strings.Replace(complete, scanner("gosec", "passed"), scanner("gosec", "failed"), 1), false},
		{"pass failed last", "pass", strings.Replace(complete, scanner("workflow-analysis", "passed"), scanner("workflow-analysis", "failed"), 1), false},
		{"pass inapplicable", "pass", scanners("not-applicable"), false},
		{"pass inapplicable middle", "pass", strings.Replace(complete, scanner("gosec", "passed"), scanner("gosec", "not-applicable"), 1), false},
		{"pass inapplicable last", "pass", strings.Replace(complete, scanner("workflow-analysis", "passed"), scanner("workflow-analysis", "not-applicable"), 1), false},
		{"pass duplicate", "pass", complete + "," + scanner("gosec", "passed"), false},
		{"fail empty", "fail", "", true},
		{"fail partial", "fail", scanner("gosec", "failed"), true},
		{"fail duplicate scanner", "fail", scanner("gosec", "failed") + "," + scanner("gosec", "passed"), false},
		{"fail ten", "fail", ten, true},
		{"blocked empty", "blocked", "", true},
		{"blocked mixed", "blocked", scanners("not-applicable"), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[]}`)
			value := strings.Replace(matrix(test.scannerRows, "[]"), `"status":"pass"`, `"status":"`+test.status+`"`, 1)
			write(t, filepath.Join(root, "security-matrix.json"), value)
			if err := securitydocs.Validate(root); (err == nil) != test.valid {
				t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
			}
		})
	}
}

func TestValidateOptionalRiskEvidenceAndAcceptedExpiry(t *testing.T) {
	base := `{"id":"SEC-1","module":"github.com/acme/example","severity":"low","status":"open","owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release"`
	for _, test := range []struct {
		name, suffix string
		valid        bool
	}{
		{"optional absent", "", true},
		{"optional present", `,"evidence":"artifact://review","expires_at":"2099-10-01T00:00:00Z"`, true},
		{"optional blank evidence", `,"evidence":""`, false},
		{"optional null evidence", `,"evidence":null`, false},
		{"optional numeric evidence", `,"evidence":1`, false},
		{"optional null expiry", `,"expires_at":null`, false},
		{"accepted future", `,"evidence":"artifact://review","expires_at":"2099-10-01T00:00:00Z"`, true},
		{"accepted expired", `,"evidence":"artifact://review","expires_at":"2020-01-01T00:00:00Z"`, false},
		{"accepted missing expiry", `,"evidence":"artifact://review"`, false},
		{"accepted null evidence", `,"evidence":null,"expires_at":"2099-10-01T00:00:00Z"`, false},
		{"accepted null expiry", `,"evidence":"artifact://review","expires_at":null`, false},
		{"accepted malformed expiry", `,"evidence":"artifact://review","expires_at":"not a date"`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			risk := base + test.suffix + `}`
			if strings.HasPrefix(test.name, "accepted") {
				risk = strings.Replace(risk, `"status":"open"`, `"status":"accepted"`, 1)
			}
			write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[`+risk+`]}`)
			write(t, filepath.Join(root, "security-matrix.json"), `{"schema_version":1,"modules":[]}`)
			if err := securitydocs.Validate(root); (err == nil) != test.valid {
				t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
			}
		})
	}
}

func TestValidateIndependentRepeatedDirectories(t *testing.T) {
	for _, test := range []struct {
		name, risks string
		valid       bool
	}{
		{"valid", `{"schema_version":1,"risks":[]}`, true},
		{"invalid", `{"schema_version":1,"risks":null}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			matrix := `{"schema_version":1,"modules":[]}`
			write(t, filepath.Join(root, "risk-register.json"), test.risks)
			write(t, filepath.Join(root, "security-matrix.json"), matrix)
			for attempt := range 10 {
				if err := securitydocs.Validate(root); (err == nil) != test.valid {
					t.Fatalf("attempt %d: Validate() error = %v, want valid=%t", attempt, err, test.valid)
				}
			}
			stored, err := os.ReadFile(filepath.Join(root, "risk-register.json"))
			if err != nil || string(stored) != test.risks {
				t.Fatalf("risk register changed: %q, %v", stored, err)
			}
			stored, err = os.ReadFile(filepath.Join(root, "security-matrix.json"))
			if err != nil || string(stored) != matrix {
				t.Fatalf("security matrix changed: %q, %v", stored, err)
			}
		})
	}
}

func TestValidateRejectsMalformedAndTrailingDocuments(t *testing.T) {
	for _, name := range []string{"risk-register.json", "security-matrix.json"} {
		base := `{"schema_version":1,"risks":[]}`
		if name == "security-matrix.json" {
			base = `{"schema_version":1,"modules":[]}`
		}
		for _, test := range []struct{ name, value string }{
			{"null root", `null`},
			{"truncated", base[:len(base)-1]},
			{"second value", base + " " + base},
			{"trailing junk", base + " unexpected"},
		} {
			t.Run(name+"/"+test.name, func(t *testing.T) {
				root := t.TempDir()
				risks := `{"schema_version":1,"risks":[]}`
				matrix := `{"schema_version":1,"modules":[]}`
				if name == "risk-register.json" {
					risks = test.value
				} else {
					matrix = test.value
				}
				write(t, filepath.Join(root, "risk-register.json"), risks)
				write(t, filepath.Join(root, "security-matrix.json"), matrix)
				if err := securitydocs.Validate(root); err == nil {
					t.Fatal("Validate accepted a malformed or trailing security document")
				}
			})
		}
	}
}

func TestValidateRejectsMissingAndUnreadableDocuments(t *testing.T) {
	for _, name := range []string{"risk-register.json", "security-matrix.json"} {
		for _, unavailable := range []string{"missing", "directory instead of file"} {
			t.Run(name+"/"+unavailable, func(t *testing.T) {
				root := t.TempDir()
				if name != "risk-register.json" {
					write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[]}`)
				}
				if name != "security-matrix.json" {
					write(t, filepath.Join(root, "security-matrix.json"), `{"schema_version":1,"modules":[]}`)
				}
				if unavailable == "directory instead of file" {
					if err := os.Mkdir(filepath.Join(root, name), 0o700); err != nil {
						t.Fatal(err)
					}
				}
				if err := securitydocs.Validate(root); err == nil {
					t.Fatal("Validate accepted missing or unreadable document")
				}
			})
		}
	}
}

func TestValidateRiskVerdictCrossProduct(t *testing.T) {
	for _, test := range []struct {
		name, severity, status, verdict, residual string
		valid                                     bool
	}{
		{"open low passes", "low", "open", "pass", "[]", true},
		{"open medium blocks pass", "medium", "open", "pass", "[]", false},
		{"open medium permits fail", "medium", "open", "fail", "[]", true},
		{"accepted medium listed passes", "medium", "accepted", "pass", `["SEC-1"]`, true},
		{"accepted medium omitted blocks pass", "medium", "accepted", "pass", "[]", false},
		{"accepted low omitted blocks pass", "low", "accepted", "pass", "[]", false},
		{"accepted low listed passes", "low", "accepted", "pass", `["SEC-1"]`, true},
		{"accepted high blocks pass", "high", "accepted", "pass", `["SEC-1"]`, false},
		{"accepted critical blocks pass", "critical", "accepted", "pass", `["SEC-1"]`, false},
		{"open critical blocks pass", "critical", "open", "pass", "[]", false},
		{"open critical permits blocked", "critical", "open", "blocked", "[]", true},
		{"mitigated high passes", "high", "mitigated", "pass", "[]", true},
		{"closed high passes", "high", "closed", "pass", "[]", true},
		{"wrong residual state fails", "low", "open", "blocked", `["SEC-1"]`, false},
		{"unknown residual fails", "low", "open", "blocked", `["SEC-2"]`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			value := matrix(scanners("passed"), test.residual)
			value = strings.Replace(value, `"status":"pass"`, `"status":"`+test.verdict+`"`, 1)
			risk := fmt.Sprintf(`{"id":"SEC-1","module":"github.com/acme/example","severity":%q,"status":%q,"owner":"security@example.com","rationale":"risk","mitigation":"planned","review_condition":"on release"`, test.severity, test.status)
			switch test.status {
			case "accepted":
				risk += fmt.Sprintf(`,"evidence":%q,"expires_at":"2099-10-01T00:00:00Z"`, matrixEvidence(value))
			case "mitigated", "closed":
				risk += fmt.Sprintf(`,"evidence":%q`, matrixEvidence(value))
			}
			write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[`+risk+`}]}`)
			write(t, filepath.Join(root, "security-matrix.json"), value)
			if err := securitydocs.Validate(root); (err == nil) != test.valid {
				t.Fatalf("Validate() error = %v, want valid=%t", err, test.valid)
			}
		})
	}
}

func TestValidateAcceptsMultipleOwnedModuleRows(t *testing.T) {
	root := t.TempDir()
	firstRow := strings.TrimSuffix(strings.TrimPrefix(matrix(scanners("passed"), "[]"), `{"schema_version":1,"modules":[`), `]}`)
	firstRow = strings.Replace(firstRow, "github.com/acme/example", "github.com/acme/first", 1)
	secondRow := strings.Replace(firstRow, "github.com/acme/first", "github.com/acme/second", 1)
	matrixDocument := `{"schema_version":1,"modules":[` + firstRow + `,` + secondRow + `]}`
	firstRisk := fmt.Sprintf(`{"id":"SEC-1","module":"github.com/acme/first","severity":"medium","status":"mitigated","owner":"security@example.com","rationale":"bounded input","mitigation":"limits","review_condition":"on parser change","evidence":%q}`, matrixEvidence(matrixDocument))
	secondRisk := strings.Replace(firstRisk, `"SEC-1"`, `"SEC-2"`, 1)
	secondRisk = strings.Replace(secondRisk, "github.com/acme/first", "github.com/acme/second", 1)
	write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[`+firstRisk+`,`+secondRisk+`]}`)
	write(t, filepath.Join(root, "security-matrix.json"), matrixDocument)
	if err := securitydocs.Validate(root); err != nil {
		t.Fatalf("Validate() rejected distinct owned rows: %v", err)
	}
}

func TestValidateRejectsCrossModuleAndDuplicateRiskClaims(t *testing.T) {
	accepted := `{"id":"SEC-1","module":"github.com/acme/first","severity":"low","status":"accepted","owner":"security@example.com","rationale":"risk","mitigation":"limits","review_condition":"on release","evidence":"artifact://review","expires_at":"2099-10-01T00:00:00Z"}`
	firstRow := strings.TrimSuffix(strings.TrimPrefix(matrix(scanners("passed"), `["SEC-1"]`), `{"schema_version":1,"modules":[`), `]}`)
	firstRow = strings.Replace(firstRow, "github.com/acme/example", "github.com/acme/first", 1)
	secondRow := strings.Replace(firstRow, "github.com/acme/first", "github.com/acme/second", 1)
	secondRow = strings.Replace(secondRow, `["SEC-1"]`, `[]`, 1)
	for _, test := range []struct{ name, risks, rows, want string }{
		{"wrong module residual", `{"schema_version":1,"risks":[` + accepted + `]}`, firstRow + "," + strings.Replace(secondRow, `"residual_risks":[]`, `"residual_risks":["SEC-1"]`, 1), "module-owned"},
		{"mixed valid and unknown residuals", `{"schema_version":1,"risks":[` + accepted + `]}`, strings.Replace(firstRow, `["SEC-1"]`, `["SEC-1","SEC-2"]`, 1), "unknown residual risk"},
		{"duplicate residual", `{"schema_version":1,"risks":[` + accepted + `]}`, strings.Replace(firstRow, `["SEC-1"]`, `["SEC-1","SEC-1"]`, 1), "published schema"},
		{"duplicate risk identifier", `{"schema_version":1,"risks":[` + accepted + `,` + accepted + `]}`, firstRow, "duplicate risk id"},
		{"duplicate module row", `{"schema_version":1,"risks":[` + accepted + `]}`, firstRow + "," + firstRow, "duplicate module record"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			matrixDocument := `{"schema_version":1,"modules":[` + test.rows + `]}`
			risks := strings.ReplaceAll(test.risks, "artifact://review", matrixEvidence(matrixDocument))
			write(t, filepath.Join(root, "risk-register.json"), risks)
			write(t, filepath.Join(root, "security-matrix.json"), matrixDocument)
			if err := securitydocs.Validate(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func matrixEvidence(value string) string {
	digest := sha256.Sum256([]byte(value))
	return fmt.Sprintf("sha256:%x:security-matrix.json", digest)
}

func TestValidateRereadsCorrectedRecords(t *testing.T) {
	root := t.TempDir()
	matrix := `{"schema_version":1,"modules":[]}`
	write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":null}`)
	write(t, filepath.Join(root, "security-matrix.json"), matrix)
	if err := securitydocs.Validate(root); err == nil {
		t.Fatal("Validate accepted the initial invalid record")
	}
	corrected := `{"schema_version":1,"risks":[]}`
	write(t, filepath.Join(root, "risk-register.json"), corrected)
	if err := securitydocs.Validate(root); err != nil {
		t.Fatalf("Validate rejected corrected records: %v", err)
	}
	for name, expected := range map[string]string{"risk-register.json": corrected, "security-matrix.json": matrix} {
		stored, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(stored) != expected {
			t.Fatalf("%s changed during retry: %v", name, err)
		}
	}
}

func TestValidateRejectsMissingNestedRequiredFields(t *testing.T) {
	baseRisk := `{"schema_version":1,"risks":[{"id":"SEC-1","module":"github.com/acme/example","severity":"medium","status":"mitigated","owner":"security@example.com","rationale":"risk","mitigation":"limits","review_condition":"on release"}]}`
	baseMatrix := matrix(scanners("passed"), "[]")
	sections := []struct {
		name   string
		fields []string
	}{
		{"risk", []string{"id", "module", "severity", "status", "owner", "rationale", "mitigation", "review_condition"}},
		{"module", []string{"module", "revision", "scanners", "release_verdict"}},
		{"scanner", []string{"name", "tool_version", "command", "status", "completed_at", "result", "result_sha256"}},
		{"verdict", []string{"status", "owner", "decided_at", "rationale", "residual_risks"}},
	}
	for _, section := range sections {
		for _, field := range section.fields {
			for _, mutation := range []string{"missing", "null"} {
				t.Run(section.name+"/"+field+"/"+mutation, func(t *testing.T) {
					var riskDoc, matrixDoc map[string]any
					if err := json.Unmarshal([]byte(baseRisk), &riskDoc); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal([]byte(baseMatrix), &matrixDoc); err != nil {
						t.Fatal(err)
					}
					module := objectArrayElement(t, matrixDoc, "modules", 0)
					var target map[string]any
					switch section.name {
					case "risk":
						target = objectArrayElement(t, riskDoc, "risks", 0)
					case "module":
						target = module
					case "scanner":
						target = objectArrayElement(t, module, "scanners", 0)
					case "verdict":
						target = objectField(t, module, "release_verdict")
					}
					if mutation == "missing" {
						delete(target, field)
					} else {
						target[field] = nil
					}
					riskBytes, err := json.Marshal(riskDoc)
					if err != nil {
						t.Fatal(err)
					}
					matrixBytes, err := json.Marshal(matrixDoc)
					if err != nil {
						t.Fatal(err)
					}
					root := t.TempDir()
					writeBytes(t, filepath.Join(root, "risk-register.json"), riskBytes)
					writeBytes(t, filepath.Join(root, "security-matrix.json"), matrixBytes)
					if err := securitydocs.Validate(root); err == nil {
						t.Fatal("Validate accepted an absent or null required field")
					}
				})
			}
		}
	}
}

func TestValidateRejectsMalformedScannerAndVerdictProvenance(t *testing.T) {
	base := matrix(scanners("passed"), "[]")
	for _, test := range []struct{ name, old, replacement string }{
		{"short digest", `"result_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`, `"result_sha256":"bad"`},
		{"uppercase digest", `"result_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`, `"result_sha256":"0123456789ABCDEF0123456789abcdef0123456789abcdef0123456789abcdef"`},
		{"scanner timestamp", `"completed_at":"2026-09-13T00:00:00Z"`, `"completed_at":"not a timestamp"`},
		{"verdict timestamp", `"decided_at":"2026-09-13T00:00:00Z"`, `"decided_at":"not a timestamp"`},
		{"blank tool version", `"tool_version":"v1.0.0"`, `"tool_version":" "`},
		{"blank command", `"command":"tool ./..."`, `"command":" "`},
		{"blank result", `"result":"artifact://result"`, `"result":""`},
		{"blank verdict owner", `"owner":"security@example.com"`, `"owner":" "`},
		{"blank verdict rationale", `"rationale":"reviewed exact evidence"`, `"rationale":" "`},
		{"wrong scanner name", `"name":"codeql"`, `"name":"unknown"`},
		{"wrong command type", `"command":"tool ./..."`, `"command":4`},
		{"extra scanner property", `"name":"codeql"`, `"name":"codeql","unknown":1`},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "risk-register.json"), `{"schema_version":1,"risks":[]}`)
			value := strings.Replace(base, test.old, test.replacement, 1)
			if value == base {
				t.Fatal("fixture mutation did not apply")
			}
			write(t, filepath.Join(root, "security-matrix.json"), value)
			if err := securitydocs.Validate(root); err == nil {
				t.Fatal("Validate accepted malformed provenance")
			}
		})
	}
}

func FuzzValidateNeverPanics(f *testing.F) {
	emptyRisks := `{"schema_version":1,"risks":[]}`
	emptyMatrix := `{"schema_version":1,"modules":[]}`
	f.Add([]byte(emptyRisks), []byte(emptyMatrix))
	f.Add([]byte(`{"schema_version":1,"risks":[]}`), []byte(matrix(scanners("passed"), "[]")))
	f.Add([]byte(emptyRisks+strings.Repeat(" ", (4<<20)-len(emptyRisks))), []byte(emptyMatrix))
	f.Add([]byte(emptyRisks), []byte(emptyMatrix+strings.Repeat(" ", (4<<20)+1-len(emptyMatrix))))
	f.Add([]byte(`{"schema_version":1e0,"risks":[]}`), []byte(`{"schema_version":1.0,"modules":[]}`))
	f.Add([]byte(`{"schema_version":1,"risks":[],"risks":[]}`), []byte(`{"schema_version":1,"modules":[]}`))
	f.Add([]byte(`{"schema_version":1,"risks":[]}`), []byte(`{"schema_version":1,"modules":null}`))
	f.Add([]byte(`{"schema_version":1,"risks":[]} unexpected`), []byte(`{"schema_version":1,"modules":[]}`))
	f.Add([]byte("null"), []byte("null"))
	f.Fuzz(func(t *testing.T, risks, matrix []byte) {
		root := t.TempDir()
		writeBytes(t, filepath.Join(root, "risk-register.json"), risks)
		writeBytes(t, filepath.Join(root, "security-matrix.json"), matrix)
		result := securitydocs.Validate(root)
		if (!json.Valid(risks) || !json.Valid(matrix)) && result == nil {
			t.Fatal("validation accepted malformed JSON")
		}
		for name, expected := range map[string][]byte{"risk-register.json": risks, "security-matrix.json": matrix} {
			stored, err := os.ReadFile(filepath.Join(root, name))
			if err != nil || !bytes.Equal(stored, expected) {
				t.Fatalf("validation modified %s: %v", name, err)
			}
		}
	})
}

func scanners(status string) string {
	names := []string{"codeql", "dependency-review", "go-vet", "gosec", "govulncheck", "license", "owned-analysis", "secret-current-tree", "secret-history", "staticcheck", "workflow-analysis"}
	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, scanner(name, status))
		status = "passed"
	}
	return strings.Join(values, ",")
}

func scanner(name, status string) string {
	return fmt.Sprintf(`{"name":%q,"tool_version":"v1.0.0","command":"tool ./...","status":%q,"completed_at":"2026-09-13T00:00:00Z","result":"artifact://result","result_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`, name, status)
}

func matrix(scannerValues, residual string) string {
	return fmt.Sprintf(`{"schema_version":1,"modules":[{"module":"github.com/acme/example","revision":"0123456789abcdef0123456789abcdef01234567","scanners":[%s],"release_verdict":{"status":"pass","owner":"security@example.com","decided_at":"2026-09-13T00:00:00Z","rationale":"reviewed exact evidence","residual_risks":%s}}]}`, scannerValues, residual)
}

func objectArrayElement(t *testing.T, document map[string]any, field string, index int) map[string]any {
	t.Helper()
	values, ok := document[field].([]any)
	if !ok || index < 0 || index >= len(values) {
		t.Fatalf("%s = %#v, want object array with index %d", field, document[field], index)
	}
	value, ok := values[index].(map[string]any)
	if !ok {
		t.Fatalf("%s[%d] = %#v, want object", field, index, values[index])
	}
	return value
}

func objectField(t *testing.T, document map[string]any, field string) map[string]any {
	t.Helper()
	value, ok := document[field].(map[string]any)
	if !ok {
		t.Fatalf("%s = %#v, want object", field, document[field])
	}
	return value
}

func write(t *testing.T, path, value string) {
	t.Helper()
	writeBytes(t, path, []byte(value))
}

func writeBytes(t testing.TB, path string, value []byte) {
	t.Helper()
	if err := os.WriteFile(path, value, 0o600); err != nil {
		t.Fatal(err)
	}
}
