SHELL := /usr/bin/env bash

.PHONY: build check ci config inventory repository-check

define run_go
	task="$$(mktemp -d "$${TMPDIR:-/tmp}/go-library-tools-make.XXXXXX")"; \
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

check:
	$(call run_go,run ./cmd/golib check --all)

ci: repository-check check
