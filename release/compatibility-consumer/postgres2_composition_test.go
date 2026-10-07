package compatibilityconsumer_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	postgres "github.com/faustbrian/go-postgres/v2"
	postgresservice "github.com/faustbrian/go-postgres/v2/adapters/service"
	"github.com/faustbrian/go-service"
)

func TestPublicPostgresExplicitResolution(t *testing.T) {
	_, err := postgres.PrepareConfig(context.Background(), postgres.Config{DSN: "fixture"})
	var configErr *postgres.ConfigError
	if !errors.As(err, &configErr) || configErr.Field != "resolver" || configErr.Problem != "is required" {
		t.Fatalf("missing resolver: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	_, err = postgres.PrepareConfig(ctx, postgres.Config{
		DSN: "fixture",
		ResolveDSN: func(context.Context, string) (*postgres.PoolConfig, error) {
			calls++
			return nil, nil
		},
	})
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("pre-cancellation: error=%v resolver calls=%d", err, calls)
	}

	privateCause := errors.New("fixture-private-resolver-marker")
	_, err = postgres.PrepareConfig(context.Background(), postgres.Config{
		DSN: "fixture",
		ResolveDSN: func(context.Context, string) (*postgres.PoolConfig, error) {
			return nil, privateCause
		},
	})
	configErr = nil
	if !errors.As(err, &configErr) || configErr.Field != "resolver" ||
		configErr.Problem != "could not resolve configuration" || configErr.Cause != nil ||
		errors.Is(err, privateCause) || strings.Contains(err.Error(), privateCause.Error()) {
		t.Fatal("resolver failure did not retain its private public classification")
	}
}

func TestPublicPostgresBorrowedServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	pool := &borrowedPostgresResource{}
	adapter, err := postgresservice.New(postgresservice.Options{
		Name: "fixture-postgres", Pool: pool, StartupPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var component service.Component = adapter.Component()
	var readiness service.ReadinessCheck = adapter.Readiness()
	if !errors.Is(readiness.Run(ctx), postgresservice.ErrUnavailable) || pool.pings != 0 {
		t.Fatal("inactive borrowed pool was probed")
	}
	if err := component.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if pool.pings != 1 {
		t.Fatal("startup did not validate the borrowed pool")
	}
	if err := readiness.Run(ctx); err != nil || pool.pings != 2 {
		t.Fatalf("active readiness: %v; pings=%d", err, pool.pings)
	}
	for range 2 {
		if err := component.Stop(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if pool.closes != 0 {
		t.Fatal("adapter closed a caller-owned pool")
	}
	if !errors.Is(readiness.Run(ctx), postgresservice.ErrUnavailable) || pool.pings != 2 {
		t.Fatal("stopped borrowed pool remained available")
	}
}

type borrowedPostgresResource struct {
	pings  int
	closes int
}

func (pool *borrowedPostgresResource) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	pool.pings++
	return nil
}

func (pool *borrowedPostgresResource) Close(context.Context) error {
	pool.closes++
	return nil
}
