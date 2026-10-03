// Package mutation validates and executes exact mutation-testing campaigns.
package mutation

import (
	"archive/zip"
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
)

const (
	// MaximumArchiveSize bounds one imported checkpoint archive.
	MaximumArchiveSize      = 16 << 20
	maximumExpandedSize     = 64 << 20
	maximumCheckpointSize   = 1 << 20
	maximumCheckpointCount  = 2048
	maximumCompressionRatio = 200
)

var (
	// ErrInvalid identifies malformed mutation evidence or configuration.
	ErrInvalid = errors.New("invalid mutation evidence")
	// ErrUnapproved identifies evidence lacking an exact migration approval.
	ErrUnapproved = errors.New("unapproved mutation evidence migration")
	// ErrInputChanged identifies an otherwise valid checkpoint whose observed
	// package input no longer matches an approved identity.
	ErrInputChanged       = errors.New("mutation checkpoint input changed")
	digestRE              = regexp.MustCompile(`^[0-9a-f]{64}$`)
	revisionRE            = regexp.MustCompile(`^[0-9a-f]{40}$`)
	versionRE             = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	errDuplicateObjectKey = errors.New("duplicate object key")
)

// Checkpoint is validated legacy evidence suitable for content-identity
// migration. Revision fields are intentionally discarded.
type Checkpoint struct {
	Module            string
	Package           string
	ExecutionRevision string
	InputDigest       string
	InputLineage      []string
	VerifierDigest    string
	BinaryDigest      string
	VerifierSource    string
	Gremlins          string
	Environment       map[string]string
	ReportDigest      string
	Report            json.RawMessage
	Mutants           int
}

type legacyCheckpoint struct {
	SchemaVersion               int               `json:"schema_version"`
	Module                      string            `json:"module"`
	Package                     string            `json:"package"`
	ExecutionRevision           string            `json:"execution_revision,omitempty"`
	ValidatedRevision           string            `json:"validated_revision,omitempty"`
	GateInputDigest             string            `json:"gate_input_digest"`
	IdentityLineage             []string          `json:"identity_lineage,omitempty"`
	IdentityMigration           json.RawMessage   `json:"identity_migration,omitempty"`
	LegacyModuleGateInputDigest string            `json:"legacy_module_gate_input_digest,omitempty"`
	GremlinsVersion             string            `json:"gremlins_version"`
	GremlinsVerifierSHA256      string            `json:"gremlins_verifier_sha256,omitempty"`
	VerifierIdentitySource      string            `json:"verifier_identity_source,omitempty"`
	GremlinsBinarySHA256        string            `json:"gremlins_binary_sha256,omitempty"`
	Environment                 map[string]string `json:"environment"`
	Report                      json.RawMessage   `json:"report"`
}

type report struct {
	GoModule          string          `json:"go_module,omitempty"`
	ElapsedTime       json.RawMessage `json:"elapsed_time,omitempty"`
	Files             []reportFile    `json:"files"`
	MutantsKilled     *int            `json:"mutants_killed,omitempty"`
	MutantsLived      *int            `json:"mutants_lived,omitempty"`
	MutantsNotCovered *int            `json:"mutants_not_covered,omitempty"`
	MutantsNotViable  *int            `json:"mutants_not_viable,omitempty"`
	MutantsTotal      *int            `json:"mutants_total,omitempty"`
	MutationCoverage  *float64        `json:"mutations_coverage,omitempty"`
	MutatorStatistics json.RawMessage `json:"mutator_statistics,omitempty"`
	TestEfficacy      *float64        `json:"test_efficacy,omitempty"`
}

type reportFile struct {
	FileName  string     `json:"file_name"`
	Mutations []mutation `json:"mutations"`
}

type mutation struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// ReportResult summarizes a strictly validated Gremlins report.
type ReportResult struct {
	Digest     string
	Mutants    int
	Killed     int
	Equivalent int
}

// ValidateReport requires every viable mutant to be killed and validates any
// aggregate counters against the individual mutation records.
func ValidateReport(reader io.Reader) (ReportResult, error) {
	return ValidateReportWithReview(reader, nil)
}

// ValidateReportWithReview retains native statuses while allowing only exact
// reviewed LIVED coordinates; nil retains the original all-KILLED policy.
func ValidateReportWithReview(reader io.Reader, review *EquivalentReview) (ReportResult, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumCheckpointSize+1))
	if err != nil {
		return ReportResult{}, fmt.Errorf("%w: read mutation report", ErrInvalid)
	}
	if len(data) > maximumCheckpointSize {
		return ReportResult{}, fmt.Errorf("%w: mutation report exceeds %d bytes", ErrInvalid, maximumCheckpointSize)
	}
	return validateReportDataWithReview(data, review)
}

// ReadBootstrap strictly reads one bounded legacy checkpoint archive.
func ReadBootstrap(reader io.ReaderAt, size int64) ([]Checkpoint, error) {
	if size <= 0 || size > MaximumArchiveSize {
		return nil, fmt.Errorf("%w: archive size must be between 1 and %d bytes", ErrInvalid, MaximumArchiveSize)
	}
	archive, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, fmt.Errorf("%w: open archive", ErrInvalid)
	}
	if len(archive.File) == 0 || len(archive.File) > maximumCheckpointCount {
		return nil, fmt.Errorf("%w: archive contains an invalid entry count", ErrInvalid)
	}
	files := append([]*zip.File(nil), archive.File...)
	slices.SortFunc(files, func(left, right *zip.File) int { return strings.Compare(left.Name, right.Name) })
	seenNames := make(map[string]struct{}, len(files))
	seenIdentities := make(map[string]struct{}, len(files))
	checkpoints := make([]Checkpoint, 0, len(files))
	var expanded uint64
	for _, file := range files {
		expanded, err = addExpanded(expanded, file.UncompressedSize64)
		if err != nil {
			return nil, err
		}
	}
	for _, file := range files {
		if err := validateArchiveEntry(file); err != nil {
			return nil, err
		}
		if _, exists := seenNames[file.Name]; exists {
			return nil, fmt.Errorf("%w: duplicate archive entry", ErrInvalid)
		}
		seenNames[file.Name] = struct{}{}
		checkpoint, err := readCheckpoint(file.UncompressedSize64, file.Open)
		if err != nil {
			return nil, fmt.Errorf("checkpoint: %w", err)
		}
		identity := checkpoint.Module + "\x00" + checkpoint.Package
		if _, exists := seenIdentities[identity]; exists {
			return nil, fmt.Errorf("%w: duplicate checkpoint identity", ErrInvalid)
		}
		seenIdentities[identity] = struct{}{}
		checkpoints = append(checkpoints, checkpoint)
	}
	return checkpoints, nil
}

func addExpanded(current, size uint64) (uint64, error) {
	if size > maximumExpandedSize || current > maximumExpandedSize-size {
		return 0, fmt.Errorf("%w: expanded archive exceeds %d bytes", ErrInvalid, maximumExpandedSize)
	}
	return current + size, nil
}

func validateArchiveEntry(file *zip.File) error {
	clean := path.Clean(file.Name)
	if clean != file.Name || strings.HasPrefix(clean, "/") || clean == "." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("%w: unsafe archive path", ErrInvalid)
	}
	if file.FileInfo().Mode()&os.ModeType != 0 || file.FileInfo().IsDir() {
		return fmt.Errorf("%w: archive entry is not a regular file", ErrInvalid)
	}
	if !strings.HasPrefix(clean, "mutation-checkpoints/") || path.Dir(clean) != "mutation-checkpoints" || path.Ext(clean) != ".json" {
		return fmt.Errorf("%w: unexpected archive entry", ErrInvalid)
	}
	if file.UncompressedSize64 == 0 || file.UncompressedSize64 > maximumCheckpointSize {
		return fmt.Errorf("%w: checkpoint size is invalid", ErrInvalid)
	}
	compressed := file.CompressedSize64
	if compressed == 0 || file.UncompressedSize64 > compressed*maximumCompressionRatio {
		return fmt.Errorf("%w: checkpoint compression ratio is excessive", ErrInvalid)
	}
	return nil
}

func readCheckpoint(declaredSize uint64, open func() (io.ReadCloser, error)) (Checkpoint, error) {
	opened, err := open()
	if err != nil {
		return Checkpoint{}, fmt.Errorf("%w: open checkpoint", ErrInvalid)
	}
	data, readErr := io.ReadAll(io.LimitReader(opened, maximumCheckpointSize+1))
	_ = opened.Close()
	if readErr != nil {
		return Checkpoint{}, fmt.Errorf("%w: read checkpoint", ErrInvalid)
	}
	if len(data) > maximumCheckpointSize || uint64(len(data)) != declaredSize {
		return Checkpoint{}, fmt.Errorf("%w: checkpoint size mismatch", ErrInvalid)
	}
	return parseCheckpoint(data)
}

func parseCheckpoint(data []byte) (Checkpoint, error) {
	var legacy legacyCheckpoint
	if err := decodeStrict(data, &legacy); err != nil {
		return Checkpoint{}, err
	}
	return validateCheckpoint(legacy)
}

func validateCheckpoint(value legacyCheckpoint) (Checkpoint, error) {
	if value.SchemaVersion != 3 {
		return Checkpoint{}, fmt.Errorf("%w: schema_version must be 3", ErrInvalid)
	}
	if !validRelative(value.Module) || !validRelative(value.Package) {
		return Checkpoint{}, fmt.Errorf("%w: module and package must be safe relative paths", ErrInvalid)
	}
	if !digestRE.MatchString(value.GateInputDigest) || !versionRE.MatchString(value.GremlinsVersion) {
		return Checkpoint{}, fmt.Errorf("%w: checkpoint identity is malformed", ErrInvalid)
	}
	for name, digest := range map[string]string{
		"gremlins_verifier_sha256": value.GremlinsVerifierSHA256,
		"gremlins_binary_sha256":   value.GremlinsBinarySHA256,
	} {
		if digest != "" && !digestRE.MatchString(digest) {
			return Checkpoint{}, fmt.Errorf("%w: %s is malformed", ErrInvalid, name)
		}
	}
	if value.VerifierIdentitySource != "" && value.VerifierIdentitySource != "executed" && value.VerifierIdentitySource != "approved-semantic-migration" {
		return Checkpoint{}, fmt.Errorf("%w: invalid verifier_identity_source", ErrInvalid)
	}
	if len(value.Environment) == 0 {
		return Checkpoint{}, fmt.Errorf("%w: environment is required", ErrInvalid)
	}
	result, err := validateReportData(value.Report)
	if err != nil {
		return Checkpoint{}, err
	}
	return Checkpoint{
		Module: value.Module, Package: value.Package, ExecutionRevision: value.ExecutionRevision,
		InputDigest: value.GateInputDigest, InputLineage: append([]string(nil), value.IdentityLineage...),
		VerifierDigest: value.GremlinsVerifierSHA256, BinaryDigest: value.GremlinsBinarySHA256,
		VerifierSource: value.VerifierIdentitySource, Gremlins: value.GremlinsVersion,
		Environment: value.Environment, ReportDigest: result.Digest,
		Report: append(json.RawMessage(nil), value.Report...), Mutants: result.Mutants,
	}, nil
}

func validateReportData(data []byte) (ReportResult, error) {
	return validateReportDataWithReview(data, nil)
}

func validateReportDataWithReview(data []byte, review *EquivalentReview) (ReportResult, error) {
	var parsed report
	if err := decodeStrict(data, &parsed); err != nil {
		return ReportResult{}, fmt.Errorf("report: %w", err)
	}
	if parsed.Files == nil {
		return ReportResult{}, fmt.Errorf("%w: mutation report files must be an array", ErrInvalid)
	}
	mutants := 0
	killed := 0
	equivalent := 0
	reviewed := make(map[string]struct{})
	if review != nil {
		if err := review.validate(); err != nil {
			return ReportResult{}, fmt.Errorf("%w: invalid equivalent-mutant review", ErrInvalid)
		}
		for _, candidate := range review.Mutations {
			reviewed[mutationIdentity(candidate.FileName, candidate.Type, candidate.Line, candidate.Column)] = struct{}{}
		}
	}
	files := make(map[string]struct{}, len(parsed.Files))
	identities := make(map[string]struct{})
	for _, file := range parsed.Files {
		if file.FileName == "" {
			return ReportResult{}, fmt.Errorf("%w: mutation file name is required", ErrInvalid)
		}
		if _, exists := files[file.FileName]; exists {
			return ReportResult{}, fmt.Errorf("%w: duplicate mutation file", ErrInvalid)
		}
		files[file.FileName] = struct{}{}
		for _, candidate := range file.Mutations {
			if candidate.Type == "" || candidate.Line <= 0 || candidate.Column <= 0 {
				return ReportResult{}, fmt.Errorf("%w: mutation location is malformed", ErrInvalid)
			}
			identity := mutationIdentity(file.FileName, candidate.Type, candidate.Line, candidate.Column)
			if _, exists := identities[identity]; exists {
				return ReportResult{}, fmt.Errorf("%w: duplicate mutation identity", ErrInvalid)
			}
			identities[identity] = struct{}{}
			switch candidate.Status {
			case "KILLED":
				killed++
			case "LIVED":
				if _, exists := reviewed[identity]; !exists {
					return ReportResult{}, fmt.Errorf("%w: non-killed mutant lacks an exact equivalent review", ErrInvalid)
				}
				delete(reviewed, identity)
				equivalent++
			default:
				return ReportResult{}, fmt.Errorf("%w: non-killed mutant", ErrInvalid)
			}
			mutants++
		}
	}
	if len(reviewed) != 0 {
		return ReportResult{}, fmt.Errorf("%w: reviewed equivalent mutant is absent or no longer lived", ErrInvalid)
	}
	metrics := []bool{
		parsed.MutantsKilled != nil, parsed.MutantsLived != nil,
		parsed.MutantsNotCovered != nil, parsed.MutantsNotViable != nil,
		parsed.MutantsTotal != nil, parsed.MutationCoverage != nil,
		parsed.TestEfficacy != nil,
	}
	metricCount := 0
	for _, present := range metrics {
		if present {
			metricCount++
		}
	}
	if metricCount != 0 && metricCount != len(metrics) || equivalent > 0 && metricCount == 0 {
		return ReportResult{}, fmt.Errorf("%w: incomplete aggregate counters", ErrInvalid)
	}
	if metricCount > 0 {
		actual := [7]float64{float64(*parsed.MutantsKilled), float64(*parsed.MutantsLived), float64(*parsed.MutantsNotCovered), float64(*parsed.MutantsNotViable), float64(*parsed.MutantsTotal), *parsed.MutationCoverage, *parsed.TestEfficacy}
		efficacy := float64(100)
		if mutants > 0 {
			efficacy = float64(killed) * 100 / float64(mutants)
		}
		expected := [7]float64{float64(killed), float64(equivalent), 0, 0, float64(mutants), 100, efficacy}
		for index := range actual {
			if math.IsNaN(actual[index]) || math.Abs(actual[index]-expected[index]) > 1e-9 {
				return ReportResult{}, fmt.Errorf("%w: aggregate counters do not prove complete mutant accounting", ErrInvalid)
			}
		}
	}
	return ReportResult{Digest: canonicalReportDigest(data), Mutants: mutants, Killed: killed, Equivalent: equivalent}, nil
}

func canonicalReportDigest(data []byte) string {
	var value report
	// validateCheckpoint has already strictly decoded this report.
	_ = json.Unmarshal(data, &value)
	value.ElapsedTime = nil
	for index := range value.Files {
		slices.SortFunc(value.Files[index].Mutations, func(left, right mutation) int {
			if order := strings.Compare(left.Type, right.Type); order != 0 {
				return order
			}
			if order := cmp.Compare(left.Line, right.Line); order != 0 {
				return order
			}
			return cmp.Compare(left.Column, right.Column)
		})
	}
	slices.SortFunc(value.Files, func(left, right reportFile) int {
		return strings.Compare(left.FileName, right.FileName)
	})
	if len(value.MutatorStatistics) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(value.MutatorStatistics))
		decoder.UseNumber()
		var statistics any
		_ = decoder.Decode(&statistics)
		value.MutatorStatistics = canonicalJSON(statistics)
	}
	serialized, _ := json.Marshal(value)
	decoder := json.NewDecoder(bytes.NewReader(serialized))
	decoder.UseNumber()
	var canonicalValue any
	_ = decoder.Decode(&canonicalValue)
	canonical := canonicalJSON(canonicalValue)
	hash := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func legacyCanonicalReportDigest(data []byte) string {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	_ = decoder.Decode(&value)
	canonical := canonicalJSON(value)
	hash := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func canonicalJSON(value any) json.RawMessage {
	var canonical bytes.Buffer
	encoder := json.NewEncoder(&canonical)
	encoder.SetEscapeHTML(false)
	// Validated report values contain only types supported by json.Encoder.
	_ = encoder.Encode(value)
	return canonical.Bytes()
}

func decodeStrict(data []byte, destination any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		// Keep the static unknown-field category without its JSON field name.
		if strings.HasPrefix(err.Error(), "json: unknown field ") {
			return fmt.Errorf("%w: decode: unknown field", ErrInvalid)
		}
		return fmt.Errorf("%w: decode failure", ErrInvalid)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%w: multiple JSON values", ErrInvalid)
		}
		return fmt.Errorf("%w: trailing data", ErrInvalid)
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		if errors.Is(err, errDuplicateObjectKey) {
			return fmt.Errorf("%w: decode: duplicate object key", ErrInvalid)
		}
		return fmt.Errorf("%w: decode failure", ErrInvalid)
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	if delimiter == '{' {
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key := objectKey(keyToken)
			if _, exists := seen[key]; exists {
				return errDuplicateObjectKey
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	} else {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	}
	_, err = decoder.Token()
	return err
}

func objectKey(token any) string {
	key, ok := token.(string)
	if !ok {
		panic("JSON decoder returned a non-string object key")
	}
	return key
}

func validRelative(value string) bool {
	if value == "." {
		return true
	}
	return value != "" && !strings.ContainsAny(value, "\\\x00") && path.Clean(value) == value && !strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "../")
}
