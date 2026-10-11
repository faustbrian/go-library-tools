package gates

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

func mutationPackageRunner(t *testing.T) Runner {
	t.Helper()
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	return Runner{
		Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{{
			Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true},
			Packages: []inventory.Package{
				{Directory: ".", Production: true, CoverageRequired: true},
				{Directory: "selected", Production: true, CoverageRequired: true},
				{Directory: "excluded"},
				{Directory: "nonproduction", CoverageRequired: true},
			},
		}}},
		Policy: config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{
			Root: ".verification/mutation", MaxWorkers: 4,
			Imports: []config.MutationImport{{Module: ".", Archive: "missing.zip", Ledger: "missing.json"}},
		}},
		Executor: runtimeExecutor(filepath.Join(root, ".task")),
	}
}

func TestMutationPackageSelectsOnePackageWithoutImportingWholeArchive(t *testing.T) {
	for _, target := range []string{".", "selected"} {
		t.Run(target, func(t *testing.T) {
			runner := mutationPackageRunner(t)
			var output bytes.Buffer
			runner.Output = &output
			called := false
			runner.mutationCampaign = func(ctx context.Context, campaign mutation.Campaign) error {
				called = true
				if !reflect.DeepEqual(campaign.Policy.Packages, []string{target}) || campaign.Policy.Workers != 1 {
					t.Fatalf("selected policy = %#v", campaign.Policy)
				}
				deadline, ok := ctx.Deadline()
				if remaining := time.Until(deadline); !ok || remaining <= 0 || remaining > 20*time.Minute {
					t.Fatalf("selected deadline = %v, present %t", deadline, ok)
				}
				return nil
			}
			if err := runner.MutationPackage(t.Context(), ".", target); err != nil || !called {
				t.Fatalf("MutationPackage() = %v, called %t", err, called)
			}
			if len(runner.Catalog.Modules[0].Packages) != 4 || runner.Policy.Mutation.MaxWorkers != 4 {
				t.Fatal("single-package command mutated the caller's full-module policy")
			}
			if !strings.Contains(output.String(), "package "+target+" only") || strings.Contains(output.String(), "mutation-import") {
				t.Fatalf("scope output = %q", output.String())
			}
		})
	}
}

func TestMutationPackageRejectsInvalidSelectionsBeforeRuntime(t *testing.T) {
	for _, test := range []struct {
		name, module, directory, want string
		disabled                      bool
	}{
		{name: "unknown module", module: "missing", directory: ".", want: "unknown module"},
		{name: "disabled", module: ".", directory: ".", disabled: true, want: "does not enable mutation"},
		{name: "unknown package", module: ".", directory: "missing", want: "not mutation-enabled"},
		{name: "excluded package", module: ".", directory: "excluded", want: "not mutation-enabled"},
		{name: "nonproduction package", module: ".", directory: "nonproduction", want: "not mutation-enabled"},
		{name: "empty package", module: ".", want: "not mutation-enabled"},
		{name: "noncanonical", module: ".", directory: "./selected", want: "not mutation-enabled"},
		{name: "outside", module: ".", directory: "../outside", want: "not mutation-enabled"},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := mutationPackageRunner(t)
			runner.Catalog.Modules[0].RequiredServices = []string{"postgresql"}
			if test.disabled {
				runner.Catalog.Modules[0].Gates["mutation"] = false
			}
			runner.Executor = executorFunction(func(context.Context, Command) error {
				t.Fatal("invalid selection executed a runtime command")
				return nil
			})
			runner.startServices = func(context.Context, []string) (serviceLease, error) {
				t.Fatal("invalid selection started services")
				return nil, nil
			}
			if err := runner.MutationPackage(t.Context(), test.module, test.directory); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("MutationPackage() = %v, want %q", err, test.want)
			}
		})
	}
}

func TestMutationPackagePreservesCallerDeadlineAndCleansServicesAfterFailure(t *testing.T) {
	runner := mutationPackageRunner(t)
	runner.Catalog.Modules[0].RequiredServices = []string{"postgresql"}
	cleanupFailure := errors.New("cleanup failed")
	lease := &fakeServiceLease{identities: map[string]string{"postgresql": "fixture-identity"}, closeErr: cleanupFailure}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	wantDeadline, _ := ctx.Deadline()
	runner.startServices = func(ctx context.Context, names []string) (serviceLease, error) {
		if deadline, _ := ctx.Deadline(); deadline != wantDeadline || !reflect.DeepEqual(names, []string{"postgresql"}) {
			t.Fatalf("service deadline/names = %v, %v", deadline, names)
		}
		return lease, nil
	}
	campaignFailure := errors.New("viable survivor")
	runner.mutationCampaign = func(ctx context.Context, campaign mutation.Campaign) error {
		if deadline, _ := ctx.Deadline(); deadline != wantDeadline || campaign.Policy.ServiceIdentities["postgresql"] != "fixture-identity" {
			t.Fatalf("campaign deadline/services = %v, %v", deadline, campaign.Policy.ServiceIdentities)
		}
		return campaignFailure
	}
	if err := runner.MutationPackage(ctx, ".", "selected"); !errors.Is(err, campaignFailure) || !errors.Is(err, cleanupFailure) || lease.closes != 1 {
		t.Fatalf("campaign failure/cleanup = %v, closes %d", err, lease.closes)
	}
}

func TestMutationPackageCancellationCannotBecomePassingEvidence(t *testing.T) {
	for _, phase := range []string{"before setup", "during campaign"} {
		t.Run(phase, func(t *testing.T) {
			runner := mutationPackageRunner(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			called := false
			runner.mutationCampaign = func(context.Context, mutation.Campaign) error {
				called = true
				cancel()
				return nil
			}
			if phase == "before setup" {
				cancel()
			}
			if err := runner.MutationPackage(ctx, ".", "."); !errors.Is(err, context.Canceled) || called != (phase == "during campaign") {
				t.Fatalf("canceled command = %v, campaign called %t", err, called)
			}
		})
	}
	ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	defer cancel()
	if err := mutationPackageRunner(t).MutationPackage(ctx, ".", "."); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("exhausted caller budget = %v", err)
	}
}
