# Schemas, descriptions and protocol alternatives

[Comparisons](index.md) · [Documentation map](../index.md)

## JSON Schema

[Golib JSON Schema](https://github.com/faustbrian/go-json-schema) and
[santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema)
both provide schema validation boundaries; compare the exact supported draft,
vocabularies, reference loading policy, regex semantics, annotations, diagnostic
limits and immutable compiled-runtime behavior. Santhosh-tekuri documents
multiple draft versions; do not infer the same draft support from a common
package name. Other peers can be candidates only after checking their current
maintenance and the same capability matrix.

Use pinned official fixtures and independently specified expected outcomes for
conformance claims. Passing a package's tests does not by itself demonstrate
all specification features. Runtime resource/remote-reference policy is a
separate concern from boolean validation correctness. Golib's specification
decisions and authority caveats remain visible through the package README and
[limitations](../limitations.md).

## OpenAPI

[Golib OpenAPI](https://github.com/faustbrian/go-openapi),
[kin-openapi](https://github.com/getkin/kin-openapi), and
[libopenapi](https://github.com/pb33f/libopenapi) operate in API-description
space. Inspect each pinned release for supported document versions, reference
resolution, validation, diff/analysis and generation capabilities; those are
adjacent features, not automatically equivalent. A parser does not replace
an HTTP runtime, and document validation does not prove a generated client's
compatibility with the running application.

Choose description tooling for required interoperability and diagnostics,
then test handler/client contract alignment separately. The supported portal
JSON:API recipe is not evidence of OpenAPI-generated handlers.

## JSON-RPC and JSON:API

[Golib JSON-RPC](https://github.com/faustbrian/go-jsonrpc),
[Sourcegraph jsonrpc2](https://github.com/sourcegraph/jsonrpc2), and
[go-ethereum rpc](https://github.com/ethereum/go-ethereum/tree/master/rpc)
are candidates with different transport and application expectations. Check
batch/notification handling, deadlines, cancellation, size limits, errors and
maintenance for the version you adopt. Sourcegraph's README explicitly records
unsupported batching at the observation date; it is not equivalent for a
batch-heavy service. An Ethereum-oriented RPC layer should be evaluated for
its domain/transport assumptions rather than assumed to be a neutral substitute.

[Golib JSON:API](https://github.com/faustbrian/go-jsonapi) and
[google/jsonapi](https://github.com/google/jsonapi) overlap at resource document
serialization. Serialization alone is not a complete resource server, query
planner, authorization policy or Atomic Operations implementation. Verify
relationship, include, sparse-field, extension/profile and error behavior.
Do not label an alternative maintained solely because a repository exists;
inspect its security/release activity before production selection.

No comparative throughput or superior-safety result is claimed here. Explicit
resource limits and rejected hidden initialization are intentional design
boundaries; unsupported protocol features are limitations and must stay named.

Next: [protocol decision guide](../api-protocols.md), [limitations](../limitations.md),
[benchmark interpretation](performance.md).
