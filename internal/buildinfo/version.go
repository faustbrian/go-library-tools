// Package buildinfo exposes immutable release identity injected at build time.
package buildinfo

import (
	"errors"
	"fmt"
	"regexp"
)

var releaseRE = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

// Version is replaced with an exact release by the release build. Development
// builds remain explicit and may operate only by deliberate source execution.
var Version = "dev"

// ValidateRequired rejects a released binary that does not match repository
// policy. Development builds are accepted to avoid circular self-bootstrap.
func ValidateRequired(required string) error {
	if Version == "dev" {
		return nil
	}
	if !releaseRE.MatchString(Version) {
		return errors.New("golib binary has malformed release identity")
	}
	if Version != required {
		return fmt.Errorf("golib binary version %s does not match required %s", Version, required)
	}
	return nil
}
