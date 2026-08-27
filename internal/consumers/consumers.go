// Package consumers loads the maintained repository rollout inventory.
package consumers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const maximumManifestSize = 1 << 20

var (
	namePattern  = regexp.MustCompile(`^go-[a-z0-9]+(?:-[a-z0-9]+)*$`)
	ownerPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?$`)
)

// Manifest is the explicit source of truth for repositories managed by the
// tooling release and rollout process.
type Manifest struct {
	SchemaVersion int          `json:"schema_version"`
	Owner         string       `json:"owner"`
	Repositories  []Repository `json:"repositories"`
}

// Repository classifies one repository for rollout decisions.
type Repository struct {
	Name           string `json:"name"`
	Classification string `json:"classification"`
	DefaultBranch  string `json:"default_branch"`
	Reason         string `json:"reason,omitempty"`
}

// Summary is a stable machine-readable inventory result.
type Summary struct {
	SchemaVersion int          `json:"schema_version"`
	Owner         string       `json:"owner"`
	Total         int          `json:"total"`
	Active        int          `json:"active"`
	Deferred      int          `json:"deferred"`
	Tooling       int          `json:"tooling"`
	Repositories  []Repository `json:"repositories"`
}

// Load reads and validates a bounded repository-relative manifest.
func Load(root, relative string) (Manifest, error) {
	data, err := repositoryfile.Read(root, relative, maximumManifestSize)
	if err != nil {
		return Manifest{}, fmt.Errorf("read consumer manifest: %w", err)
	}
	return Parse(data)
}

// Parse strictly decodes and validates a consumer manifest.
func Parse(data []byte) (Manifest, error) {
	if len(data) > maximumManifestSize {
		return Manifest{}, errors.New("consumer manifest exceeds maximum size")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode consumer manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Manifest{}, errors.New("consumer manifest must contain exactly one JSON document")
	}
	if err := manifest.validate(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (manifest Manifest) validate() error {
	if manifest.SchemaVersion != 1 {
		return errors.New("consumer manifest schema_version must be 1")
	}
	if !ownerPattern.MatchString(manifest.Owner) {
		return errors.New("consumer manifest owner is invalid")
	}
	if len(manifest.Repositories) == 0 || len(manifest.Repositories) > 256 {
		return errors.New("consumer manifest must contain between 1 and 256 repositories")
	}
	active, tooling := 0, 0
	previous := ""
	for index, repository := range manifest.Repositories {
		field := fmt.Sprintf("repositories[%d]", index)
		if !namePattern.MatchString(repository.Name) {
			return fmt.Errorf("%s.name is invalid", field)
		}
		if previous != "" && repository.Name <= previous {
			return fmt.Errorf("%s.name must be unique and sorted", field)
		}
		previous = repository.Name
		if !validBranch(repository.DefaultBranch) {
			return fmt.Errorf("%s.default_branch is invalid", field)
		}
		switch repository.Classification {
		case "active":
			active++
			if repository.Reason != "" {
				return fmt.Errorf("%s.reason must be omitted for an active repository", field)
			}
		case "deferred":
			if strings.TrimSpace(repository.Reason) == "" {
				return fmt.Errorf("%s.reason is required for a deferred repository", field)
			}
		case "tooling":
			tooling++
			if strings.TrimSpace(repository.Reason) == "" {
				return fmt.Errorf("%s.reason is required for a tooling repository", field)
			}
		default:
			return fmt.Errorf("%s.classification is invalid", field)
		}
	}
	if active < 1 {
		return errors.New("consumer manifest contains no active repositories")
	}
	if tooling != 1 {
		return errors.New("consumer manifest must identify exactly one tooling repository")
	}
	return nil
}

func validBranch(value string) bool {
	return value != "" && !strings.HasPrefix(value, ".") && !strings.HasSuffix(value, ".") &&
		!strings.HasPrefix(value, "/") && !strings.HasSuffix(value, "/") &&
		!strings.Contains(value, "..") && !strings.Contains(value, "@{") &&
		!strings.ContainsAny(value, "\\ ~^:?*[\t\r\n")
}

// Summary returns stable counts without changing repository order.
func (manifest Manifest) Summary() Summary {
	result := Summary{
		SchemaVersion: manifest.SchemaVersion,
		Owner:         manifest.Owner,
		Total:         len(manifest.Repositories),
		Repositories:  append([]Repository(nil), manifest.Repositories...),
	}
	for _, repository := range manifest.Repositories {
		switch repository.Classification {
		case "active":
			result.Active++
		case "deferred":
			result.Deferred++
		case "tooling":
			result.Tooling++
		}
	}
	return result
}
