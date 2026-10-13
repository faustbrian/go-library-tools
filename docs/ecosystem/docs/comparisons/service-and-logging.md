# HTTP service and logging choices

[Comparisons](index.md) · [Documentation map](../index.md)

## Server, routing and middleware

| Choice | Equivalent boundary | Tradeoff to evaluate |
| --- | --- | --- |
| [net/http](https://pkg.go.dev/net/http) | HTTP server, handler and ServeMux | Prefer when standard routing suffices. Your application still owns graceful shutdown, probes and policy order. |
| [Chi](https://github.com/go-chi/chi) | Composable routing over standard HTTP handlers | Evaluate its routing/middleware ergonomics without assuming it supplies a complete process lifecycle. |
| [Gin](https://gin-gonic.com/en/docs/middleware/) | HTTP routing and framework handler/middleware conventions | Binding and framework context can be useful; examine the conversion seam for existing `http.Handler` policies and tests. |
| [Echo](https://github.com/labstack/echo) | HTTP framework with its own context/handler conventions | Compare binding, error handling and middleware semantics, not only route dispatch. |
| [Fiber](https://github.com/gofiber/fiber) | Web framework based on fasthttp | Its request/context model is not the standard net/http contract; evaluate interoperability, buffer lifetime and required adapters explicitly. |
| Golib Service + Router + HTTP Middleware | Separate lifecycle, routing and policy owners | Select only the contracts needed. Explicit construction makes application decisions visible but requires application wiring; it is not an all-in-one framework. |

The framework-specific seams above are integration considerations, not measured
defects. A routing microbenchmark does not compare lifecycle, authentication,
serialization, telemetry or deployment. Golib's lifecycle and optional modules
can be valuable even when an application retains another router; test the actual
handler and shutdown boundary. See [ownership](../integration-map.md).

## Logging

| Choice | Capability | Adoption consideration |
| --- | --- | --- |
| [log/slog](https://pkg.go.dev/log/slog) | Standard structured logging API and handlers | Start here when standard fields and handler composition meet the need. Select redaction, output and cardinality deliberately. |
| [Zap](https://github.com/uber-go/zap) | Structured logging with typed and sugared APIs | Compare fields, caller/error encoding and handler integration on your workload; upstream benchmark results are not a Golib comparison. |
| [Zerolog](https://github.com/rs/zerolog) | Structured event-oriented logging | Check event construction, output format and disabled-level behavior against your application's logging contract. |
| [Golib Log](https://github.com/faustbrian/go-log) | Package-owned logging facilities and explicit composition | Read the actual supported handler/formatter boundary. Do not add it solely to replace adequate standard slog calls. |

Providers/handlers remain caller-owned. Logging a request body or credential is
not made safe by a different logger. Compare redaction, field collision behavior
and trace correlation separately from allocation throughput.

Next: [benchmark interpretation](performance.md), [operations](../operations/index.md),
[protocol comparisons](protocols.md).
