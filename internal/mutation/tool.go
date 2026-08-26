package mutation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maximumDownloadMetadataSize = 1 << 20

// Process executes one non-shell command for verifier construction.
type Process func(context.Context, string, []string, string, map[string]string, io.Writer, io.Writer) error

// Tool identifies one built, content-addressed verifier binary.
type Tool struct {
	Path   string
	Digest string
}

type downloadMetadata struct {
	Directory string `json:"Dir"`
	Sum       string `json:"Sum"`
	GoModSum  string `json:"GoModSum"`
}

// Arguments returns the complete pinned Gremlins campaign contract.
func Arguments(target, output, tags string, discover bool, workers int) ([]string, error) {
	path := strings.TrimPrefix(target, "./")
	if (target != "." && (!strings.HasPrefix(target, "./") || !validRelative(path))) ||
		!filepath.IsAbs(output) || workers < 1 || workers > 64 || strings.ContainsAny(tags, "\x00\r\n") {
		return nil, fmt.Errorf("%w: mutation command arguments are malformed", ErrInvalid)
	}
	arguments := []string{
		"unleash", target,
		"--integration", "--coverpkg", target,
		"--exclude-files", "^.+/",
		"--workers", strconv.Itoa(workers), "--test-cpu", "1", "--timeout-coefficient", "10",
		"--threshold-efficacy", "100", "--threshold-mcover", "100",
		"--arithmetic-base", "--conditionals-boundary", "--conditionals-negation",
		"--invert-assignments", "--invert-bitwise", "--invert-bwassign",
		"--increment-decrement", "--invert-logical", "--invert-loopctrl",
		"--invert-negatives", "--remove-self-assignments",
		"--output-statuses", "lctvsr", "--output", output,
	}
	if tags != "" {
		arguments = append(arguments, "--tags", tags)
	}
	if discover {
		arguments = append(arguments, "--dry-run")
	}
	return arguments, nil
}

// BuildVerifier downloads the checksum-pinned Gremlins source, applies the
// embedded semantic patches, and builds a task-owned binary.
func BuildVerifier(ctx context.Context, workspace string, process Process) (Tool, error) {
	if !filepath.IsAbs(workspace) || process == nil {
		return Tool{}, fmt.Errorf("%w: verifier workspace and process are required", ErrInvalid)
	}
	environment, err := toolEnvironment(workspace)
	if err != nil {
		return Tool{}, err
	}
	var metadataOutput boundedBuffer
	metadataOutput.maximum = maximumDownloadMetadataSize
	if err := process(ctx, "go", []string{"mod", "download", "-json", "github.com/go-gremlins/gremlins@" + GremlinsVersion}, workspace, environment, &metadataOutput, io.Discard); err != nil {
		return Tool{}, fmt.Errorf("download Gremlins: %w", err)
	}
	var metadata downloadMetadata
	decoder := json.NewDecoder(strings.NewReader(metadataOutput.String()))
	if err := decoder.Decode(&metadata); err != nil {
		return Tool{}, fmt.Errorf("decode Gremlins download metadata: %w", err)
	}
	if err := requireJSONEnd(decoder); err != nil {
		return Tool{}, err
	}
	if metadata.Directory == "" || metadata.Sum != gremlinsSum || metadata.GoModSum != gremlinsGoModSum {
		return Tool{}, fmt.Errorf("%w: Gremlins download checksum mismatch", ErrInvalid)
	}
	source := filepath.Join(workspace, "gremlins-source")
	if err := os.CopyFS(source, os.DirFS(metadata.Directory)); err != nil {
		return Tool{}, fmt.Errorf("copy Gremlins source: %w", err)
	}
	assets := VerifierAssets()
	for _, asset := range verifierAssetNames[2:] {
		patchPath := filepath.Join(workspace, filepath.Base(asset.embedded))
		if err := os.WriteFile(patchPath, assets[asset.identity], 0o600); err != nil {
			return Tool{}, fmt.Errorf("write verifier patch: %w", err)
		}
		if err := process(ctx, "git", []string{"apply", "--whitespace=nowarn", patchPath}, source, nil, io.Discard, io.Discard); err != nil {
			return Tool{}, fmt.Errorf("apply verifier patch %s: %w", filepath.Base(asset.embedded), err)
		}
	}
	binary := filepath.Join(workspace, "golib-gremlins")
	if err := process(ctx, "go", []string{"build", "-trimpath", "-buildvcs=false", "-o", binary, "./cmd/gremlins"}, source, environment, io.Discard, io.Discard); err != nil {
		return Tool{}, fmt.Errorf("build Gremlins: %w", err)
	}
	file, err := os.Open(binary)
	if err != nil {
		return Tool{}, fmt.Errorf("open Gremlins binary: %w", err)
	}
	digest, hashErr := hashVerifier(file)
	_ = file.Close()
	if hashErr != nil {
		return Tool{}, hashErr
	}
	return Tool{Path: binary, Digest: digest}, nil
}

func toolEnvironment(workspace string) (map[string]string, error) {
	environment := map[string]string{
		"GOWORK":     "off",
		"GOCACHE":    filepath.Join(workspace, "go-build"),
		"GOMODCACHE": filepath.Join(workspace, "go-mod"),
		"GOTMPDIR":   filepath.Join(workspace, "go-tmp"),
		"GOBIN":      filepath.Join(workspace, "go-bin"),
	}
	for _, directory := range []string{environment["GOCACHE"], environment["GOMODCACHE"], environment["GOTMPDIR"], environment["GOBIN"]} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return nil, fmt.Errorf("create verifier workspace: %w", err)
		}
	}
	return environment, nil
}

func requireJSONEnd(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("multiple JSON values")
		}
		return fmt.Errorf("decode Gremlins download metadata: %w", err)
	}
	return nil
}

func hashVerifier(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("hash Gremlins binary: %w", err)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

type boundedBuffer struct {
	data    strings.Builder
	maximum int
}

func (buffer *boundedBuffer) Write(value []byte) (int, error) {
	if buffer.data.Len()+len(value) > buffer.maximum {
		return 0, errors.New("output exceeds bound")
	}
	return buffer.data.Write(value)
}

func (buffer *boundedBuffer) String() string { return buffer.data.String() }
