package gates

import (
	"regexp"
	"strings"
)

// #nosec G101 -- this public hash-only report template is not a credential.
const secretMetadataTemplate = `{{range .}}{{sha256sum .File}} {{sha256sum .RuleID}} {{.StartLine}} {{if .Commit}}{{.Commit}}{{else}}-{{end}}
{{end}}`

const maximumSecretLocations = 32

// The owned template emits only identities and positions, never finding text.
var secretLocationPattern = regexp.MustCompile(`^[0-9a-f]{64} [0-9a-f]{64} [1-9][0-9]{0,6} (-|[0-9a-f]{40}|[0-9a-f]{64})$`)

// Access is serialized by the containing boundedProcessOutput mutex.
type secretMetadata struct {
	line      [256]byte
	length    int
	invalid   bool
	omitted   bool
	locations []string
}

func (metadata *secretMetadata) write(value []byte) {
	for _, character := range value {
		if metadata.invalid {
			return
		}
		if character != '\n' {
			if metadata.length == len(metadata.line) {
				metadata.invalid = true
				metadata.locations = nil
				clear(metadata.line[:])
				metadata.length = 0
				return
			}
			metadata.line[metadata.length] = character
			metadata.length++
			continue
		}
		line := metadata.line[:metadata.length]
		if !secretLocationPattern.Match(line) {
			metadata.invalid = true
			metadata.locations = nil
			clear(metadata.line[:])
			metadata.length = 0
			return
		}
		if len(metadata.locations) < maximumSecretLocations {
			metadata.locations = append(metadata.locations, string(line))
		} else {
			metadata.omitted = true
		}
		clear(metadata.line[:metadata.length])
		metadata.length = 0
	}
}

func (metadata *secretMetadata) summary() string {
	if metadata.invalid || metadata.length != 0 || len(metadata.locations) == 0 {
		clear(metadata.line[:])
		metadata.length = 0
		return ""
	}
	result := "\nsecret-location " + strings.Join(metadata.locations, "\nsecret-location ")
	if metadata.omitted {
		result += "\nadditional secret locations omitted"
	}
	return result
}
