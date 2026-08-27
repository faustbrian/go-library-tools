GO ?= go
APIDIFF_VERSION ?= v0.0.0-20260718201538-764159d718ef

.PHONY: api conformance

api:
	APIDIFF_VERSION=$(APIDIFF_VERSION) ./scripts/check-api.sh

conformance:
	$(GO) test ./internal/specification ./internal/modelgen -count=1
	$(GO) generate .
	git diff --exit-code -- oas30/model_generated.go oas30/model_generated_test.go oas31/model_generated.go oas31/model_generated_test.go oas32/model_generated.go oas32/model_generated_test.go swagger20/model_generated.go swagger20/model_generated_test.go specification/conformance/normative.tsv specification/conformance/object-fields.tsv
	$(GO) test ./internal/specification/cmd/provenance -count=1
	$(GO) run ./internal/specification/cmd/provenance -root .
