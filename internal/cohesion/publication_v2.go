package cohesion

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

// ErrV2TargetExists reports that a publication target already exists.
var ErrV2TargetExists = errors.New("cohesion v2 publication target already exists")

// V2DurabilityUnknownError means publication committed the immutable name but
// the containing-directory durability sync did not complete.
type V2DurabilityUnknownError struct{ Cause error }

func (failure V2DurabilityUnknownError) Error() string {
	return "cohesion v2 publication committed but durability is unknown"
}

func (failure V2DurabilityUnknownError) Unwrap() error { return failure.Cause }

func publishProjectV2(content []byte, target string) error {
	if !filepath.IsAbs(target) || filepath.Clean(target) != target {
		return errors.New("cohesion v2 publication target must be canonical and absolute")
	}
	parent, name := filepath.Dir(target), filepath.Base(target)
	stage, err := os.CreateTemp(parent, "."+name+".golib-stage-*")
	if err != nil {
		return errors.New("create cohesion v2 project stage")
	}
	stagePath := stage.Name()
	defer os.Remove(stagePath)
	closed := false
	defer func() {
		if !closed {
			_ = stage.Close()
		}
	}()
	if err := stage.Chmod(0o600); err != nil {
		return errors.New("set cohesion v2 project stage mode")
	}
	if _, err := io.Copy(stage, bytes.NewReader(content)); err != nil {
		return errors.New("write cohesion v2 project stage")
	}
	if err := stage.Chmod(0o444); err != nil {
		return errors.New("set cohesion v2 project immutable mode")
	}
	if err := stage.Sync(); err != nil {
		return errors.New("sync cohesion v2 project stage")
	}
	if err := stage.Close(); err != nil {
		return errors.New("close cohesion v2 project stage")
	}
	closed = true

	if err := os.Link(stagePath, target); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrV2TargetExists
		}
		return errors.New("commit cohesion v2 project publication")
	}
	if err := syncDirectory(parent); err != nil {
		return V2DurabilityUnknownError{Cause: err}
	}
	if err := os.Remove(stagePath); err != nil {
		return V2DurabilityUnknownError{Cause: fmt.Errorf("remove committed stage name: %w", err)}
	}
	if err := syncDirectory(parent); err != nil {
		return V2DurabilityUnknownError{Cause: err}
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return errors.New("open publication directory for sync")
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return errors.New("sync publication directory")
	}
	return nil
}

var aggregateV2MemberNames = []string{
	"catalog-consumer.json",
	"catalog-consumer.md",
	"catalog-engineering.json",
	"catalog-engineering.md",
}

func publishAggregateV2(artifacts map[string][]byte, target string) error {
	if !filepath.IsAbs(target) || filepath.Clean(target) != target {
		return errors.New("cohesion v2 aggregate target must be canonical and absolute")
	}
	if len(artifacts) != len(aggregateV2MemberNames) {
		return errors.New("cohesion v2 aggregate must contain exactly four members")
	}
	for _, name := range aggregateV2MemberNames {
		if _, exists := artifacts[name]; !exists {
			return errors.New("cohesion v2 aggregate member set is invalid")
		}
	}

	parent, name := filepath.Dir(target), filepath.Base(target)
	stage, err := os.MkdirTemp(parent, "."+name+".golib-stage-*")
	if err != nil {
		return errors.New("create cohesion v2 aggregate stage")
	}
	defer cleanupAggregateStage(stage)
	if err := os.Chmod(stage, 0o700); err != nil {
		return errors.New("set cohesion v2 aggregate stage mode")
	}
	names := slices.Clone(aggregateV2MemberNames)
	slices.Sort(names)
	for _, member := range names {
		path := filepath.Join(stage, member)
		file, createErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if createErr != nil {
			return errors.New("create cohesion v2 aggregate member")
		}
		if _, writeErr := file.Write(artifacts[member]); writeErr != nil {
			_ = file.Close()
			return errors.New("write cohesion v2 aggregate member")
		}
		if chmodErr := file.Chmod(0o444); chmodErr != nil {
			_ = file.Close()
			return errors.New("set cohesion v2 aggregate member mode")
		}
		if syncErr := file.Sync(); syncErr != nil {
			_ = file.Close()
			return errors.New("sync cohesion v2 aggregate member")
		}
		if closeErr := file.Close(); closeErr != nil {
			return errors.New("close cohesion v2 aggregate member")
		}
	}
	if err := os.Chmod(stage, 0o555); err != nil {
		return errors.New("set cohesion v2 aggregate immutable mode")
	}
	if err := syncDirectory(stage); err != nil {
		return err
	}
	if err := publishDirectoryNoReplace(stage, target); err != nil {
		return err
	}
	if err := syncDirectory(parent); err != nil {
		return V2DurabilityUnknownError{Cause: err}
	}
	return nil
}

func cleanupAggregateStage(path string) {
	_ = os.Chmod(path, 0o700)
	_ = os.RemoveAll(path)
}
