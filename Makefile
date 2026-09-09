SHELL := /usr/bin/env bash

.PHONY: build check check-packages ci cohesion compatibility config consumers inventory local-check local-ci milestone-check repository-check workflows

PACKAGES ?=

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

workflows:
	$(call run_go,run ./cmd/golib workflows check)

check-packages:
	@test -n "$(strip $(PACKAGES))" || { echo 'PACKAGES is required, for example PACKAGES=./internal/inventory' >&2; exit 2; }
	$(call run_go,test $(PACKAGES))
	$(call run_go,vet $(PACKAGES))

local-check: config repository-check cohesion workflows

local-ci:
	$(call run_go,run ./cmd/golib check --local)
	$(MAKE) local-check

milestone-check: consumers compatibility

check:
	$(call run_go,run ./cmd/golib check --all)

ci: local-ci milestone-check
