SHELL := /usr/bin/env bash

.PHONY: build check check-packages ci cohesion compatibility compatibility-candidate compatibility-rebase config consumers inventory local-check local-ci metadata-ci milestone-check repository-check security-differential source-docs-inventory source-docs-check workflows

PACKAGES ?=
COMPATIBILITY_SET_ID ?=
COMPATIBILITY_OBSERVED_AT ?=
COMPATIBILITY_MODULE_COUNT ?=
COMPATIBILITY_VERSION_OVERRIDES ?=
COMPATIBILITY_GO_MOD ?=

define with_go_task
	set -euo pipefail; \
	task="$$(mktemp -d "$${TMPDIR:-/tmp}/golib-cache-session.XXXXXX")"; \
	task="$$(cd "$$task" && pwd -P)"; \
	trap 'chmod -R u+w "$$task" 2>/dev/null || true; find "$$task" -depth -delete' EXIT; \
	mkdir -m 700 -p "$$task/cache" "$$task/mod" "$$task/tmp"; \
	export GOCACHE="$$task/cache" GOMODCACHE="$$task/mod" GOTMPDIR="$$task/tmp"; \
	$(1)
endef

run_go = $(call with_go_task,go $(1))
run_golib = $(call with_go_task,mkdir -m 700 "$$task/verifier-cache" "$$task/verifier-mod" "$$task/verifier-tmp"; export GOLIB_TASK_CACHE_ROOT="$$task"; go build -o "$$task/golib" ./cmd/golib; $(1))
metadata_contract = "$$task/golib" source-docs check --json; "$$task/golib" config validate; "$$task/golib" repository check; "$$task/golib" cohesion check; "$$task/golib" workflows check; "$$task/golib" security validate --directory docs/ecosystem/security

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

source-docs-inventory:
	$(call run_go,run ./cmd/golib source-docs inventory --json)

source-docs-check:
	$(call run_go,run ./cmd/golib source-docs check --json)

check-packages:
	@test -n "$(strip $(PACKAGES))" || { echo 'PACKAGES is required, for example PACKAGES=./internal/inventory' >&2; exit 2; }
	$(call with_go_task,go test $(PACKAGES); go vet $(PACKAGES))

security-differential:
	@set -euo pipefail; \
	task="$$(mktemp -d "$${TMPDIR:-/tmp}/go-library-tools-security-differential.XXXXXX")"; \
	trap 'chmod -R u+w "$$task" 2>/dev/null || true; find "$$task" -depth -delete' EXIT; \
	mkdir -p "$$task/cache" "$$task/mod" "$$task/tmp" "$$task/uv"; \
	GOCACHE="$$task/cache" GOMODCACHE="$$task/mod" GOTMPDIR="$$task/tmp" GOWORK=off go build -o "$$task/golib" ./cmd/golib; \
	UV_CACHE_DIR="$$task/uv" UV_NO_PROGRESS=1 UV_PYTHON_PREFERENCE=only-system TMPDIR="$$task/tmp" \
		uv run --no-project --with jsonschema==4.17.3 python tools/securitydocs/differential.py "$(CURDIR)" "$$task/golib"

local-check:
	$(call run_golib,"$$task/golib" source-docs check --json; "$$task/golib" config validate; "$$task/golib" repository check; "$$task/golib" cohesion check; "$$task/golib" workflows check)

local-ci:
	$(call run_golib,$(metadata_contract); arguments=(check --local); if [[ "$${GOLIB_CHANGE_EVENT:-}" == pull_request || "$${GOLIB_CHANGE_EVENT:-}" == push || -n "$${GOLIB_BASE_SHA:-}" ]]; then arguments+=(--base-sha "$${GOLIB_BASE_SHA:-}"); fi; "$$task/golib" "$${arguments[@]}")

metadata-ci:
	$(call run_golib,[[ "$${GOLIB_BASE_SHA:-}" =~ ^[0-9a-f]{40}$$ ]] || { echo 'metadata CI requires an immutable base commit' >&2; exit 1; }; $(metadata_contract); "$$task/golib" docs check --local; "$$task/golib" secrets check --base-sha "$${GOLIB_BASE_SHA}")

milestone-check: consumers compatibility

check:
	$(call with_go_task,go run ./cmd/golib source-docs check --json; go run ./cmd/golib check --all)

ci: local-ci milestone-check
