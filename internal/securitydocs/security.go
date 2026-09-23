// Package securitydocs validates the versioned ecosystem security record contract.
//
//go:generate go run ./cmd/genschemas
package securitydocs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const maximumSecurityDocument = 4 << 20

var (
	revisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	digestPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

var requiredScanners = []string{
	"codeql",
	"dependency-review",
	"go-vet",
	"gosec",
	"govulncheck",
	"license",
	"owned-analysis",
	"secret-current-tree",
	"secret-history",
	"staticcheck",
	"workflow-analysis",
}

type riskRegister struct {
	SchemaVersion json.Number `json:"schema_version"`
	Risks         []risk      `json:"risks"`
}
type risk struct {
	ID, Module, Severity, Status, Owner, Rationale, Mitigation, ReviewCondition string
	Evidence, ExpiresAt                                                         string
}

func (value *risk) UnmarshalJSON(data []byte) error {
	type wire struct {
		ID              string `json:"id"`
		Module          string `json:"module"`
		Severity        string `json:"severity"`
		Status          string `json:"status"`
		Owner           string `json:"owner"`
		Rationale       string `json:"rationale"`
		Mitigation      string `json:"mitigation"`
		ReviewCondition string `json:"review_condition"`
		Evidence        string `json:"evidence,omitempty"`
		ExpiresAt       string `json:"expires_at,omitempty"`
	}
	var decoded wire
	if err := strictBytes(data, &decoded); err != nil {
		return err
	}
	*value = risk(decoded)
	return nil
}

type matrix struct {
	SchemaVersion json.Number    `json:"schema_version"`
	Modules       []moduleRecord `json:"modules"`
}
type moduleRecord struct {
	Module   string    `json:"module"`
	Revision string    `json:"revision"`
	Scanners []scanner `json:"scanners"`
	Verdict  verdict   `json:"release_verdict"`
}
type scanner struct {
	Name         string `json:"name"`
	ToolVersion  string `json:"tool_version"`
	Command      string `json:"command"`
	Status       string `json:"status"`
	CompletedAt  string `json:"completed_at"`
	Result       string `json:"result"`
	ResultSHA256 string `json:"result_sha256"`
}
type verdict struct {
	Status        string   `json:"status"`
	Owner         string   `json:"owner"`
	DecidedAt     string   `json:"decided_at"`
	Rationale     string   `json:"rationale"`
	ResidualRisks []string `json:"residual_risks"`
}

// Validate checks the risk register and per-module scanner/release matrix.
func Validate(directory string) error {
	return validateAt(directory, time.Now())
}

func validateAt(directory string, now time.Time) error {
	return validateAtWithReader(directory, now, readSecurityDocument)
}

type securityDocumentReader func(*os.Root, string) ([]byte, error)

func validateAtWithReader(directory string, now time.Time, read securityDocumentReader) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return fmt.Errorf("open security document root: %w", err)
	}
	defer root.Close()
	riskData, err := read(root, "risk-register.json")
	if err != nil {
		return err
	}
	var risks riskRegister
	if err := decodeDocument(riskData, riskRegisterSchema, &risks); err != nil {
		return err
	}
	seenRisks := map[string]struct{}{}
	risksByID := map[string]risk{}
	for _, risk := range risks.Risks {
		for name, value := range map[string]string{"id": risk.ID, "module": risk.Module, "owner": risk.Owner, "rationale": risk.Rationale, "mitigation": risk.Mitigation, "review_condition": risk.ReviewCondition} {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("risk record requires %s", name)
			}
		}
		if _, exists := seenRisks[risk.ID]; exists {
			return errors.New("duplicate risk id")
		}
		seenRisks[risk.ID] = struct{}{}
		risksByID[risk.ID] = risk
		if !member(risk.Severity, "critical", "high", "medium", "low") || !member(risk.Status, "open", "mitigated", "accepted", "closed") {
			return errors.New("risk record has invalid severity or status")
		}
		if risk.Status == "accepted" {
			if strings.TrimSpace(risk.Evidence) == "" || strings.TrimSpace(risk.ExpiresAt) == "" {
				return errors.New("accepted risk requires evidence and expires_at")
			}
			expiresAt, err := time.Parse(time.RFC3339, risk.ExpiresAt)
			if err != nil {
				return errors.New("accepted risk has invalid expires_at")
			}
			if !expiresAt.After(now) {
				return errors.New("accepted risk has expired")
			}
		}
	}
	matrixData, err := read(root, "security-matrix.json")
	if err != nil {
		return err
	}
	matrixSum := sha256.Sum256(matrixData)
	matrixDigest := hex.EncodeToString(matrixSum[:])
	var modules matrix
	if err := decodeDocument(matrixData, securityMatrixSchema, &modules); err != nil {
		return err
	}
	seenModules := map[string]struct{}{}
	passingModules := map[string]struct{}{}
	for _, module := range modules.Modules {
		if strings.TrimSpace(module.Module) == "" || !revisionPattern.MatchString(module.Revision) {
			return errors.New("module record requires module and immutable revision")
		}
		if _, exists := seenModules[module.Module]; exists {
			return errors.New("duplicate module record")
		}
		seenModules[module.Module] = struct{}{}
		failed := false
		seenScanners := map[string]struct{}{}
		for _, scanner := range module.Scanners {
			if !member(scanner.Name, requiredScanners...) || strings.TrimSpace(scanner.ToolVersion) == "" ||
				strings.TrimSpace(scanner.Command) == "" || strings.TrimSpace(scanner.Result) == "" ||
				!digestPattern.MatchString(scanner.ResultSHA256) ||
				!member(scanner.Status, "passed", "failed", "not-applicable") {
				return errors.New("module record has invalid scanner result")
			}
			if _, exists := seenScanners[scanner.Name]; exists {
				return errors.New("module record has duplicate scanner")
			}
			seenScanners[scanner.Name] = struct{}{}
			if _, err := time.Parse(time.RFC3339, scanner.CompletedAt); err != nil {
				return errors.New("module record has invalid scanner completion")
			}
			failed = failed || scanner.Status == "failed"
		}
		if !member(module.Verdict.Status, "pass", "fail", "blocked") || strings.TrimSpace(module.Verdict.Owner) == "" || strings.TrimSpace(module.Verdict.Rationale) == "" {
			return errors.New("module record has invalid release verdict")
		}
		if _, err := time.Parse(time.RFC3339, module.Verdict.DecidedAt); err != nil {
			return errors.New("module record has invalid release verdict time")
		}
		if module.Verdict.Status == "pass" && failed {
			return errors.New("passing verdict has failed scanner")
		}
		if module.Verdict.Status == "pass" {
			for _, required := range requiredScanners {
				if _, exists := seenScanners[required]; !exists {
					return errors.New("passing verdict lacks a required scanner")
				}
			}
			for _, scanner := range module.Scanners {
				if scanner.Status != "passed" {
					return errors.New("passing verdict requires passed scanners")
				}
			}
		}
		residual := make(map[string]struct{}, len(module.Verdict.ResidualRisks))
		for _, riskID := range module.Verdict.ResidualRisks {
			registered, exists := risksByID[riskID]
			if !exists {
				return errors.New("module record references unknown residual risk")
			}
			if registered.Module != module.Module || registered.Status != "accepted" {
				return errors.New("residual risk must be accepted and module-owned")
			}
			if _, exists := residual[riskID]; exists {
				return errors.New("module record duplicates residual risk")
			}
			residual[riskID] = struct{}{}
		}
		if module.Verdict.Status == "pass" {
			passingModules[module.Module] = struct{}{}
			for id, registered := range risksByID {
				if registered.Module != module.Module {
					continue
				}
				if member(registered.Severity, "critical", "high") && !member(registered.Status, "mitigated", "closed") {
					return errors.New("passing verdict has unresolved critical or high risk")
				}
				if registered.Severity == "medium" && registered.Status == "open" {
					return errors.New("passing verdict has open medium risk")
				}
				if registered.Status == "accepted" {
					if _, exists := residual[id]; !exists {
						return errors.New("passing verdict omits an accepted risk")
					}
				}
			}
		}
	}
	for _, registered := range risksByID {
		if _, passes := passingModules[registered.Module]; !passes || !member(registered.Status, "accepted", "mitigated", "closed") {
			continue
		}
		if registered.Evidence != "sha256:"+matrixDigest+":security-matrix.json" {
			return errors.New("resolved risk requires exact security matrix evidence")
		}
	}
	return nil
}

func readSecurityDocument(root *os.Root, path string) ([]byte, error) {
	file, err := root.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open security document: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maximumSecurityDocument+1))
	if err != nil {
		return nil, fmt.Errorf("read security document: %w", err)
	}
	if len(data) > maximumSecurityDocument {
		return nil, fmt.Errorf("security document exceeds %d bytes", maximumSecurityDocument)
	}
	return data, nil
}

func decodeDocument(data []byte, publishedSchema string, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	schemaDocument, err := jsonschema.UnmarshalJSON(strings.NewReader(publishedSchema))
	if err != nil {
		return fmt.Errorf("decode published security schema: %w", err)
	}
	if err := compiler.AddResource("schema.json", schemaDocument); err != nil {
		return fmt.Errorf("load published security schema: %w", err)
	}
	compiled, err := compiler.Compile("schema.json")
	if err != nil {
		return fmt.Errorf("compile published security schema: %w", err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return errors.New("security document must contain exactly one JSON value")
	}
	if err := compiled.Validate(instance); err != nil {
		return errors.New("security document violates published schema")
	}
	if err := strictBytes(data, target); err != nil {
		return errors.New("security document cannot be decoded")
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := walkJSONValue(decoder, 0); err != nil {
		return errors.New("security document has invalid or ambiguous JSON")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("security document must contain exactly one JSON value")
	}
	return nil
}

func walkJSONValue(decoder *json.Decoder, depth int) error {
	if depth > 128 {
		return errors.New("security document nesting exceeds limit")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	switch delimiter {
	case '{':
		keys := make(map[string]struct{})
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return errors.New("non-string JSON object key")
			}
			if _, exists := keys[name]; exists {
				return errors.New("duplicate JSON object key")
			}
			keys[name] = struct{}{}
			if err := walkJSONValue(decoder, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := walkJSONValue(decoder, depth+1); err != nil {
				return err
			}
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	_, err = decoder.Token()
	return err
}

func strictBytes(data []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func member(value string, allowed ...string) bool {
	return slices.Contains(allowed, value)
}
