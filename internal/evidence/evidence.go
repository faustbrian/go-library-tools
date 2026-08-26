// Package evidence owns the durable, content-addressed verification record.
package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	SchemaVersion = 1
	MaximumSize   = 1 << 20
)

var (
	ErrInvalid  = errors.New("invalid evidence")
	ErrConflict = errors.New("conflicting evidence")
	digestRE    = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	gateRE      = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

// Record binds a gate result to source content and verifier semantics. Git
// identity is deliberately absent because history does not change proof.
type Record struct {
	SchemaVersion  int               `json:"schema_version"`
	Repository     string            `json:"repository"`
	Module         string            `json:"module"`
	Package        string            `json:"package,omitempty"`
	Gate           string            `json:"gate"`
	InputDigest    string            `json:"input_digest"`
	VerifierDigest string            `json:"verifier_digest"`
	Result         string            `json:"result"`
	ReportDigest   string            `json:"report_digest"`
	Environment    map[string]string `json:"environment,omitempty"`
	CompletedAt    time.Time         `json:"completed_at"`
}

type durableFile interface {
	io.Writer
	Sync() error
	Close() error
	Name() string
}

type storeFileSystem interface {
	Mkdir(string, os.FileMode) error
	Lstat(string) (os.FileInfo, error)
	MkdirAll(string, os.FileMode) error
	CreateTemp(string, string) (durableFile, error)
	Link(string, string) error
	ReadFile(string) ([]byte, error)
	Remove(string) error
}

type operatingStoreFileSystem struct{}

func (operatingStoreFileSystem) Mkdir(path string, mode os.FileMode) error {
	return os.Mkdir(path, mode)
}
func (operatingStoreFileSystem) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }
func (operatingStoreFileSystem) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}
func (operatingStoreFileSystem) CreateTemp(directory, pattern string) (durableFile, error) {
	return os.CreateTemp(directory, pattern)
}
func (operatingStoreFileSystem) Link(oldPath, newPath string) error   { return os.Link(oldPath, newPath) }
func (operatingStoreFileSystem) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }
func (operatingStoreFileSystem) Remove(path string) error             { return os.Remove(path) }

// Load reads one content-addressed record and verifies its path identity.
func Load(root, gate, inputDigest string) (Record, error) {
	return load(operatingInspectFileSystem{}, root, gate, inputDigest)
}

type loadFileSystem interface {
	Lstat(string) (os.FileInfo, error)
	Open(string) (io.ReadCloser, error)
}

func load(files loadFileSystem, root, gate, inputDigest string) (Record, error) {
	if !filepath.IsAbs(root) || !gateRE.MatchString(gate) || !digestRE.MatchString(inputDigest) {
		return Record{}, fmt.Errorf("%w: evidence location is malformed", ErrInvalid)
	}
	path := filepath.Join(root, "by-input", gate, strings.TrimPrefix(inputDigest, "sha256:")+".json")
	info, err := files.Lstat(path)
	if err != nil {
		return Record{}, fmt.Errorf("inspect evidence: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Record{}, fmt.Errorf("%w: evidence is not a regular file", ErrInvalid)
	}
	file, err := files.Open(path)
	if err != nil {
		return Record{}, fmt.Errorf("open evidence: %w", err)
	}
	record, parseErr := Parse(file)
	closeErr := file.Close()
	if parseErr != nil {
		return Record{}, parseErr
	}
	if closeErr != nil {
		return Record{}, fmt.Errorf("close evidence: %w", closeErr)
	}
	if record.Gate != gate || record.InputDigest != inputDigest {
		return Record{}, fmt.Errorf("%w: evidence path identity does not match record", ErrInvalid)
	}
	return record, nil
}

// Digest computes an unambiguous identity from named content and semantics.
func Digest(semanticIdentity string, content map[string][]byte) string {
	hash := sha256.New()
	writePart(hash, []byte(semanticIdentity))
	names := make([]string, 0, len(content))
	for name := range content {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		writePart(hash, []byte(name))
		writePart(hash, content[name])
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

func writePart(writer io.Writer, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write(value)
}

// Parse strictly decodes and validates one bounded evidence record.
func Parse(reader io.Reader) (Record, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaximumSize+1))
	if err != nil {
		return Record{}, fmt.Errorf("%w: read: %v", ErrInvalid, err)
	}
	if len(data) > MaximumSize {
		return Record{}, fmt.Errorf("%w: record exceeds %d bytes", ErrInvalid, MaximumSize)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var record Record
	if err := decoder.Decode(&record); err != nil {
		return Record{}, fmt.Errorf("%w: decode: %v", ErrInvalid, err)
	}
	if err := requireEOF(decoder); err != nil {
		return Record{}, err
	}
	if err := record.Validate(); err != nil {
		return Record{}, err
	}
	return record, nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%w: multiple JSON values", ErrInvalid)
		}
		return fmt.Errorf("%w: trailing data: %v", ErrInvalid, err)
	}
	return nil
}

// Validate rejects incomplete, forged, or non-canonical record metadata.
func (record Record) Validate() error {
	if record.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: schema_version must be %d", ErrInvalid, SchemaVersion)
	}
	if record.Repository == "" || record.Module == "" {
		return fmt.Errorf("%w: repository and module are required", ErrInvalid)
	}
	if !gateRE.MatchString(record.Gate) {
		return fmt.Errorf("%w: invalid gate", ErrInvalid)
	}
	for name, digest := range map[string]string{
		"input_digest": record.InputDigest, "verifier_digest": record.VerifierDigest,
		"report_digest": record.ReportDigest,
	} {
		if !digestRE.MatchString(digest) {
			return fmt.Errorf("%w: invalid %s", ErrInvalid, name)
		}
	}
	if record.Result != "passed" && record.Result != "advisory" && record.Result != "not_applicable" && record.Result != "failed" {
		return fmt.Errorf("%w: invalid result", ErrInvalid)
	}
	if record.CompletedAt.IsZero() {
		return fmt.Errorf("%w: completed_at is required", ErrInvalid)
	}
	for key := range record.Environment {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("%w: environment keys must not be empty", ErrInvalid)
		}
	}
	return nil
}

// Marshal emits deterministic, newline-terminated JSON after validation.
func Marshal(record Record) ([]byte, error) {
	if err := record.Validate(); err != nil {
		return nil, err
	}
	data, _ := json.MarshalIndent(record, "", "  ")
	// Record contains only JSON-native values, so validation leaves no marshal
	// failure mode.
	return append(data, '\n'), nil
}

// Store writes a record once without replacing different evidence. The bool
// reports whether byte-identical evidence was already present.
func Store(root string, record Record) (string, bool, error) {
	if err := record.Validate(); err != nil {
		return "", false, err
	}
	files := operatingStoreFileSystem{}
	if err := prepareDirectories(files, root, record.Gate); err != nil {
		return "", false, err
	}
	return store(files, root, record)
}

func prepareDirectories(files storeFileSystem, root, gate string) error {
	if !filepath.IsAbs(root) {
		return fmt.Errorf("%w: evidence root must be absolute", ErrInvalid)
	}
	for _, directory := range []string{root, filepath.Join(root, "by-input"), filepath.Join(root, "by-input", gate)} {
		if err := files.Mkdir(directory, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("prepare evidence directory: %w", err)
		}
		info, err := files.Lstat(directory)
		if err != nil {
			return fmt.Errorf("inspect evidence directory: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: evidence path is not a real directory: %s", ErrInvalid, directory)
		}
	}
	return nil
}

func store(files storeFileSystem, root string, record Record) (string, bool, error) {
	data, err := Marshal(record)
	if err != nil {
		return "", false, err
	}
	digest := strings.TrimPrefix(record.InputDigest, "sha256:")
	directory := filepath.Join(root, "by-input", record.Gate)
	if err := files.MkdirAll(directory, 0o700); err != nil {
		return "", false, fmt.Errorf("create evidence directory: %w", err)
	}
	destination := filepath.Join(directory, digest+".json")
	temporary, err := files.CreateTemp(directory, ".evidence-*")
	if err != nil {
		return "", false, fmt.Errorf("create temporary evidence: %w", err)
	}
	temporaryPath := temporary.Name()
	defer files.Remove(temporaryPath)
	if _, err = temporary.Write(data); err == nil {
		err = temporary.Sync()
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", false, fmt.Errorf("write evidence: %w", err)
	}
	if err := files.Link(temporaryPath, destination); err == nil {
		return destination, false, nil
	} else if !errors.Is(err, os.ErrExist) {
		return "", false, fmt.Errorf("publish evidence: %w", err)
	}
	existing, err := files.ReadFile(destination)
	if err != nil {
		return "", false, fmt.Errorf("read existing evidence: %w", err)
	}
	existingRecord, err := Parse(bytes.NewReader(existing))
	if err != nil || !equivalent(existingRecord, record) {
		return "", false, fmt.Errorf("%w: %s", ErrConflict, destination)
	}
	return destination, true, nil
}

func equivalent(left, right Record) bool {
	return left.SchemaVersion == right.SchemaVersion &&
		left.Repository == right.Repository && left.Module == right.Module &&
		left.Package == right.Package && left.Gate == right.Gate &&
		left.InputDigest == right.InputDigest && left.VerifierDigest == right.VerifierDigest &&
		left.Result == right.Result && left.ReportDigest == right.ReportDigest &&
		maps.Equal(left.Environment, right.Environment)
}
