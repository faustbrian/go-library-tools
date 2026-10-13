# Minimal HTTP service

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `01-minimal-http-service`. The [complete source](https://github.com/faustbrian/go-service/tree/61535b800e87ae68451a7134810cedfe0cd985f6/integration/reference-recipes) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Service, healthhttp and serverhttp**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Validate addresses and acquire two loopback listeners; construct readiness, handlers and supervised HTTP runtimes.

Serve a business request and readiness request; withdraw readiness during drain before joining both servers.

## Failure, ownership and cleanup

Caller owns listeners, contexts and shutdown deadlines. Close consumers before their listeners and release all successfully acquired resources on startup failure.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-service.git` |
| Source | `61535b800e87ae68451a7134810cedfe0cd985f6` |
| Module directory | `integration/reference-recipes` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^TestMinimalHTTPRecipeStartsServesReadinessAndShutsDown$' -count=1 .
```

## Copyable five-minute run

Run this in a shell with Git, Python 3 and Go 1.27.0 available. It provisions no
external backend; loopback listeners and all downloaded/build resources are
owned by this session and cleaned on exit.

```sh
(
  set -eu
  golib_recipe_root=$(mktemp -d "${TMPDIR:-/tmp}/golib-http-recipe.XXXXXX")
  golib_recipe_root=$(cd "$golib_recipe_root" && pwd -P)
  trap 'chmod -R u+w "$golib_recipe_root" 2>/dev/null || true; find "$golib_recipe_root" -depth -delete' EXIT
  git clone --depth 1 --branch v1.8.4 https://github.com/faustbrian/go-library-tools.git "$golib_recipe_root/tools"
  git clone --no-checkout https://github.com/faustbrian/go-service.git "$golib_recipe_root/source"
  git -C "$golib_recipe_root/source" checkout --detach 61535b800e87ae68451a7134810cedfe0cd985f6
  cd "$golib_recipe_root/source/integration/reference-recipes"
  python3 "$golib_recipe_root/tools/tools/compatibility/generate.py" rebase --set-id golib-compat-v1-20260912.2 --go-mod "$PWD/go.mod"
  mkdir -p "$golib_recipe_root/cache" "$golib_recipe_root/mod" "$golib_recipe_root/tmp"
  export GOCACHE="$golib_recipe_root/cache" GOMODCACHE="$golib_recipe_root/mod"
  export GOTMPDIR="$golib_recipe_root/tmp" GOWORK=off GOTOOLCHAIN=go1.27.0
  go mod tidy
  go test -mod=readonly -run '^TestMinimalHTTPRecipeStartsServesReadinessAndShutsDown$' -count=1 .
)
```

The test serves a real loopback business request and checks readiness withdrawal
and bounded shutdown. It is not a permanent server installation. Build an
application using the selected [Service README](https://github.com/faustbrian/go-service/blob/v1.1.0/README.md)
and [adoption guide](https://github.com/faustbrian/go-service/blob/v1.1.0/docs/adoption.md)
after this first composition succeeds.

## Evidence and deployment boundary

Loopback HTTP only; no deployed network, container or load proof.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
