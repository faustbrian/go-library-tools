package mutation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const maximumRuntimeIdentitySize = 16 << 10

// ParseRuntimeIdentity validates bounded output from `go env -json`.
func ParseRuntimeIdentity(reader io.Reader) (RuntimeIdentity, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumRuntimeIdentitySize+1))
	if err != nil {
		return RuntimeIdentity{}, fmt.Errorf("%w: read Go runtime identity: %v", ErrInvalid, err)
	}
	if len(data) > maximumRuntimeIdentitySize {
		return RuntimeIdentity{}, fmt.Errorf("%w: Go runtime identity exceeds %d bytes", ErrInvalid, maximumRuntimeIdentitySize)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var identity RuntimeIdentity
	if err := decoder.Decode(&identity); err != nil {
		return RuntimeIdentity{}, fmt.Errorf("%w: decode Go runtime identity: %v", ErrInvalid, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return RuntimeIdentity{}, fmt.Errorf("%w: trailing Go runtime identity data", ErrInvalid)
	}
	if identity.GoVersion == "" || identity.GOOS == "" || identity.GOARCH == "" || identity.CGOEnabled == "" {
		return RuntimeIdentity{}, fmt.Errorf("%w: Go runtime identity is incomplete", ErrInvalid)
	}
	return identity, nil
}
