package gates

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
)

const changedSecretBudget = 3 * time.Minute

type secretChangeRange struct {
	base, head string
	paths      map[string]bool
}

// SecretsChanges scans committed additions in the complete base-to-HEAD range
// and the current contents of changed paths. It does not certify dirty or
// untracked work; Secrets retains the full integrity-scan contract.
func (runner Runner) SecretsChanges(ctx context.Context, base string) error {
	ctx, cancel := context.WithTimeout(ctx, changedSecretBudget)
	defer cancel()
	changes, err := runner.committedChanges(ctx, base)
	if err != nil {
		return err
	}
	if err := runner.prepareSecretChanges(ctx, base, changes); err != nil {
		return err
	}
	return runner.Secrets(ctx)
}

func (runner *Runner) prepareSecretChanges(ctx context.Context, base string, changes []string) error {
	head, err := runner.changeGit(ctx, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return err
	}
	head = strings.TrimSuffix(head, "\n")
	if len(head) != 40 || strings.IndexFunc(head, func(c rune) bool { return !strings.ContainsRune("0123456789abcdef", c) }) >= 0 {
		return errors.New("changed secret scan requires an immutable HEAD commit")
	}
	if _, err := runner.changeGit(ctx, "merge-base", "--is-ancestor", base, head); err != nil {
		return err
	}
	boundChanges, err := runner.committedChangesBetween(ctx, base, head)
	if err != nil {
		return err
	}
	if !slices.Equal(changes, boundChanges) {
		return errors.New("changed secret scan source moved during selection")
	}
	paths := make(map[string]bool, len(changes))
	for _, name := range changes {
		paths[name] = true
	}
	runner.secretChanges = &secretChangeRange{base: base, head: head, paths: paths}
	return nil
}
