package cohesion

import (
	"encoding/json"
	"errors"
	"fmt"
)

const maximumSourceLockSize int64 = 1 << 20

// SourceLock binds every repository used for ecosystem catalog publication to
// an immutable commit or to the release source that supplies the generator.
type SourceLock struct {
	SchemaVersion   int                    `json:"schema_version"`
	RepositoryCount int                    `json:"repository_count"`
	Repositories    []SourceLockRepository `json:"repositories"`
}

// SourceLockRepository records one catalog input repository and its expected
// adopted tooling identity when the repository is a consumer.
type SourceLockRepository struct {
	Repository string             `json:"repository"`
	Source     SourceLockRevision `json:"source"`
	Tooling    *SourceLockTooling `json:"tooling,omitempty"`
}

// SourceLockRevision is either an immutable consumer commit or the one release
// source whose commit is resolved from the release tag.
type SourceLockRevision struct {
	Kind   string `json:"kind"`
	Commit string `json:"commit,omitempty"`
}

// SourceLockTooling is the released CLI identity already adopted by a consumer.
type SourceLockTooling struct {
	Version         string `json:"version"`
	ChecksumsSHA256 string `json:"checksums_sha256"`
}

// LoadSourceLock reads and validates the closed release-time source manifest.
func LoadSourceLock(path string) (SourceLock, error) {
	return loadSourceLockWithOperations(path, defaultSourceLockOperations())
}

type sourceLockOperations struct {
	read     func(string, int64, string) ([]byte, error)
	validate func([]byte) error
	decode   func([]byte, *SourceLock) error
	semantic func(SourceLock) error
}

func defaultSourceLockOperations() sourceLockOperations {
	return sourceLockOperations{
		read:     readBoundedAggregateFile,
		validate: validateSourcesSchema,
		decode:   func(data []byte, lock *SourceLock) error { return json.Unmarshal(data, lock) },
		semantic: validateSourceLockOrder,
	}
}

func loadSourceLockWithOperations(path string, operations sourceLockOperations) (SourceLock, error) {
	data, err := operations.read(path, maximumSourceLockSize, "cohesion source lock")
	if err != nil {
		return SourceLock{}, fmt.Errorf("read cohesion source lock: %w", err)
	}
	if err := operations.validate(data); err != nil {
		return SourceLock{}, fmt.Errorf("cohesion source lock schema is invalid: %w", err)
	}
	var lock SourceLock
	if err := operations.decode(data, &lock); err != nil {
		return SourceLock{}, fmt.Errorf("decode cohesion source lock: %w", err)
	}
	if err := operations.semantic(lock); err != nil {
		return SourceLock{}, err
	}
	return lock, nil
}

func validateSourceLockOrder(lock SourceLock) error {
	if lock.RepositoryCount != len(lock.Repositories) {
		return fmt.Errorf("cohesion source lock repository_count is %d but contains %d repositories", lock.RepositoryCount, len(lock.Repositories))
	}
	releaseSources := 0
	previous := ""
	for _, repository := range lock.Repositories {
		if previous != "" && previous >= repository.Repository {
			return errors.New("cohesion source repositories must be strictly sorted and unique")
		}
		previous = repository.Repository
		if repository.Source.Kind == "release-source" {
			releaseSources++
		}
	}
	if releaseSources != 1 {
		return fmt.Errorf("cohesion source lock must contain exactly one release source; found %d", releaseSources)
	}
	return nil
}

// VerifyPolicy confirms that a checked-out repository's typed policy matches
// the immutable identity reviewed in the source lock. The release source has
// no prior-release policy requirement because it produces the new identity.
func (lock SourceLock) VerifyPolicy(repository, version, checksumsSHA256 string) error {
	for _, entry := range lock.Repositories {
		if entry.Repository != repository {
			continue
		}
		if entry.Source.Kind == "release-source" {
			return nil
		}
		if entry.Tooling == nil || entry.Tooling.Version != version || entry.Tooling.ChecksumsSHA256 != checksumsSHA256 {
			return fmt.Errorf("cohesion source policy does not match lock for %s", repository)
		}
		return nil
	}
	return fmt.Errorf("cohesion source repository is not locked: %s", repository)
}
