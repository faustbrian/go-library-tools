SHELL := /usr/bin/env bash

.PHONY: build check check-packages ci cohesion compatibility compatibility-candidate compatibility-rebase config consumers inventory local-check local-ci milestone-check repository-check security-differential workflows

PACKAGES ?=
COMPATIBILITY_SET_ID ?=
COMPATIBILITY_OBSERVED_AT ?=
COMPATIBILITY_MODULE_COUNT ?=
COMPATIBILITY_VERSION_OVERRIDES ?=
COMPATIBILITY_GO_MOD ?=

define run_go
	set -euo pipefail; \
	task="$$(mktemp -d "$${TMPDIR:-/tmp}/go-library-tools-make.XXXXXX")"; \
	task="$$(cd "$$task" && pwd -P)"; \
	trap 'chmod -R u+w "$$task" 2>/dev/null || true; find "$$task" -depth -delete' EXIT; \
	mkdir -p "$$task/cache" "$$task/mod" "$$task/tmp"; \
	GOCACHE="$$task/cache" GOMODCACHE="$$task/mod" GOTMPDIR="$$task/tmp" go $(1)
endef

build:
	$(call run_go,build ./cmd/golib)

config:
	$(call run_go,run ./cmd/golib config validate)

inventory repository-check:
	$(call run_go,run ./cmd/golib repository check)

consumers:
	$(call run_go,run ./cmd/golib consumers validate)

cohesion:
	$(call run_go,run ./cmd/golib cohesion check)

compatibility:
	python3 tools/compatibility/generate.py

compatibility-candidate:
	@test -n "$(strip $(COMPATIBILITY_SET_ID))" || { echo 'COMPATIBILITY_SET_ID is required' >&2; exit 2; }
	@test -n "$(strip $(COMPATIBILITY_OBSERVED_AT))" || { echo 'COMPATIBILITY_OBSERVED_AT is required' >&2; exit 2; }
	@test -n "$(strip $(COMPATIBILITY_MODULE_COUNT))" || { echo 'COMPATIBILITY_MODULE_COUNT is required' >&2; exit 2; }
	python3 tools/compatibility/generate.py candidate --set-id "$(COMPATIBILITY_SET_ID)" --observed-at "$(COMPATIBILITY_OBSERVED_AT)" --module-count "$(COMPATIBILITY_MODULE_COUNT)" $(foreach binding,$(COMPATIBILITY_VERSION_OVERRIDES),--version "$(binding)") --write

compatibility-rebase:
	@test -n "$(strip $(COMPATIBILITY_SET_ID))" || { echo 'COMPATIBILITY_SET_ID is required' >&2; exit 2; }
	@test -n "$(strip $(COMPATIBILITY_GO_MOD))" || { echo 'COMPATIBILITY_GO_MOD is required' >&2; exit 2; }
	python3 tools/compatibility/generate.py rebase --set-id "$(COMPATIBILITY_SET_ID)" --go-mod "$(COMPATIBILITY_GO_MOD)"

workflows:
	$(call run_go,run ./cmd/golib workflows check)

check-packages:
	@test -n "$(strip $(PACKAGES))" || { echo 'PACKAGES is required, for example PACKAGES=./internal/inventory' >&2; exit 2; }
	$(call run_go,test $(PACKAGES))
	$(call run_go,vet $(PACKAGES))

security-differential:
	@set -euo pipefail; \
	task="$$(mktemp -d "$${TMPDIR:-/tmp}/go-library-tools-security-differential.XXXXXX")"; \
	trap 'chmod -R u+w "$$task" 2>/dev/null || true; find "$$task" -depth -delete' EXIT; \
	mkdir -p "$$task/cache" "$$task/mod" "$$task/tmp" "$$task/uv"; \
	GOCACHE="$$task/cache" GOMODCACHE="$$task/mod" GOTMPDIR="$$task/tmp" GOWORK=off go build -o "$$task/golib" ./cmd/golib; \
	UV_CACHE_DIR="$$task/uv" UV_NO_PROGRESS=1 UV_PYTHON_PREFERENCE=only-system TMPDIR="$$task/tmp" \
		uv run --no-project --with jsonschema==4.17.3 python tools/securitydocs/differential.py "$(CURDIR)" "$$task/golib"

local-check: config repository-check cohesion workflows

local-ci:
	$(call run_go,run ./cmd/golib check --local)
	$(call run_go,run ./cmd/golib security validate --directory docs/ecosystem/security)
	$(MAKE) local-check

milestone-check: consumers compatibility

check:
	$(call run_go,run ./cmd/golib check --all)

ci: local-ci milestone-check
