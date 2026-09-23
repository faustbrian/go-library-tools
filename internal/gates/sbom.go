package gates

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

const maximumSBOMOutput = 16 << 20

type boundedSBOMBuffer struct {
	mutex    sync.Mutex
	data     bytes.Buffer
	overflow bool
	onLimit  func()
}

func (buffer *boundedSBOMBuffer) Write(value []byte) (int, error) {
	buffer.mutex.Lock()
	remaining := maximumSBOMOutput - buffer.data.Len()
	if len(value) <= remaining {
		written, err := buffer.data.Write(value)
		buffer.mutex.Unlock()
		return written, err
	}
	first := !buffer.overflow
	buffer.overflow = true
	_, _ = buffer.data.Write(value[:remaining])
	onLimit := buffer.onLimit
	buffer.mutex.Unlock()
	if first && onLimit != nil {
		onLimit()
	}
	return len(value), nil
}

func (buffer *boundedSBOMBuffer) setOverflowCallback(callback func()) {
	buffer.mutex.Lock()
	buffer.onLimit = callback
	alreadyOverflowed := buffer.overflow
	buffer.mutex.Unlock()
	if alreadyOverflowed && callback != nil {
		callback()
	}
}

func (buffer *boundedSBOMBuffer) didOverflow() bool {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	return buffer.overflow
}

type cycloneDXDocument struct {
	BOMFormat   string `json:"bomFormat"`
	SpecVersion string `json:"specVersion"`
}

func (runner Runner) runSBOM(ctx context.Context, directory string) error {
	var document, diagnostics boundedSBOMBuffer
	err := runner.Executor.Run(ctx, Command{
		Name: "go", Dir: directory, Env: map[string]string{"GOWORK": "off"},
		Args: []string{
			"run", "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@" + cycloneDXVersion,
			"mod", "-json", "-licenses", "-type", "library", "-noserial", "-notimestamp", "-output", "-", ".",
		},
		Stdout: &document,
		Stderr: &diagnostics,
	})
	var overflow error
	if document.didOverflow() || diagnostics.didOverflow() {
		overflow = fmt.Errorf("CycloneDX SBOM output exceeded %d bytes", maximumSBOMOutput)
	}
	if err != nil || overflow != nil {
		return fmt.Errorf("generate CycloneDX SBOM: %w", errors.Join(overflow, err))
	}
	if len(bytes.TrimSpace(document.data.Bytes())) == 0 {
		return errors.New("CycloneDX SBOM is empty")
	}
	var decoded cycloneDXDocument
	if err := json.Unmarshal(document.data.Bytes(), &decoded); err != nil {
		return errors.New("CycloneDX SBOM is not valid JSON")
	}
	if decoded.BOMFormat != "CycloneDX" {
		return errors.New("CycloneDX SBOM has invalid bomFormat")
	}
	if strings.TrimSpace(decoded.SpecVersion) == "" {
		return errors.New("CycloneDX SBOM has no specVersion")
	}
	if decoded.SpecVersion != "1.6" {
		return errors.New("CycloneDX SBOM has unsupported specVersion")
	}
	return nil
}
