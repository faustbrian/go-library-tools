package gates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

const maximumSBOMOutput = 16 << 20

type boundedSBOMBuffer struct {
	data     bytes.Buffer
	overflow bool
}

func (buffer *boundedSBOMBuffer) Write(value []byte) (int, error) {
	remaining := maximumSBOMOutput - buffer.data.Len()
	if len(value) <= remaining {
		return buffer.data.Write(value)
	}
	buffer.overflow = true
	_, _ = buffer.data.Write(value[:remaining])
	return len(value), nil
}

type cycloneDXDocument struct {
	BOMFormat   string `json:"bomFormat"`
	SpecVersion string `json:"specVersion"`
}

func (runner Runner) runSBOM(ctx context.Context, directory string, module inventory.Module) error {
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
	if err != nil {
		return fmt.Errorf("generate CycloneDX SBOM for %s: %w", module.ModulePath, err)
	}
	if document.overflow || diagnostics.overflow {
		return fmt.Errorf("CycloneDX SBOM output exceeded %d bytes for %s", maximumSBOMOutput, module.ModulePath)
	}
	if len(bytes.TrimSpace(document.data.Bytes())) == 0 {
		return fmt.Errorf("CycloneDX SBOM for %s is empty", module.ModulePath)
	}
	var decoded cycloneDXDocument
	if err := json.Unmarshal(document.data.Bytes(), &decoded); err != nil {
		return fmt.Errorf("CycloneDX SBOM for %s is not valid JSON: %w", module.ModulePath, err)
	}
	if decoded.BOMFormat != "CycloneDX" {
		return fmt.Errorf("CycloneDX SBOM for %s has bomFormat %q", module.ModulePath, decoded.BOMFormat)
	}
	if strings.TrimSpace(decoded.SpecVersion) == "" {
		return fmt.Errorf("CycloneDX SBOM for %s has no specVersion", module.ModulePath)
	}
	if decoded.SpecVersion != "1.6" {
		return fmt.Errorf("CycloneDX SBOM for %s has unsupported specVersion %q", module.ModulePath, decoded.SpecVersion)
	}
	return nil
}
