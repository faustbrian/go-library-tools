# HTTP service and logging choices

[Comparisons](index.md) · [Documentation map](../index.md)

## Server, routing and middleware

`net/http` is Golib Service's supported public HTTP transport. Compare plain
`net/http`, the owned core composition, Chi, Gin and Echo at matched service boundaries,
including routes, middleware, JSON work, probes, timeouts, errors and graceful
shutdown. The Service-owned [transport comparison](https://github.com/faustbrian/go-service/blob/ed47aa57048e8cd3780c87c5966542d4d1a91887/docs/transport-comparison.md)
records the methodology, measurements and their limitations; it is not a
universal framework ranking.

The measured owned composition is core Service + Correlation + standard
ServeMux. It does not time every optional Router, HTTP Middleware,
authentication or provider composition. The frozen recipes retain their own
composition proof, not performance qualification.

| Choice | Equivalent boundary | Tradeoff to evaluate |
| --- | --- | --- |
| [net/http](https://pkg.go.dev/net/http) | HTTP server, handler and ServeMux | Prefer when standard routing suffices. Your application still owns graceful shutdown, probes and policy order. |
| [Chi](https://github.com/go-chi/chi) | Composable routing over standard HTTP handlers | Evaluate its routing/middleware ergonomics without assuming it supplies a complete process lifecycle. |
| [Gin](https://gin-gonic.com/en/docs/middleware/) | HTTP routing and framework handler/middleware conventions | Binding and framework context can be useful; examine the conversion seam for existing `http.Handler` policies and tests. |
| [Echo](https://github.com/labstack/echo) | HTTP framework with its own context/handler conventions | Compare binding, error handling and middleware semantics, not only route dispatch. |
| Golib Service + Router + HTTP Middleware | Separate lifecycle, routing and policy owners | Select only the contracts needed. Explicit construction makes application decisions visible but requires application wiring; it is not an all-in-one framework. |

The framework-specific seams above are integration considerations, not measured
defects. A routing microbenchmark does not compare lifecycle, authentication,
serialization, telemetry or deployment. Golib's lifecycle and optional modules
can be valuable even when an application retains another router; test the actual
handler and shutdown boundary. See [ownership](../integration-map.md).

### Separate architecture track: Fiber and fasthttp

[Fiber](https://github.com/gofiber/fiber) uses fasthttp rather than the supported
`net/http` transport. It is a comparative architecture track, not a supported
Golib Service adapter. Keep its results separate from the matched `net/http`
comparison. For the Fiber v3.5.0/fasthttp v1.73.0 track,
[Fiber's context guidance](https://docs.gofiber.io/guide/go-context/) distinguishes
its reused request context from caller-supplied cancellation: `fiber.Ctx` does
not supply request-cancellation `Done`/`Err` behavior. Request/context and buffer
lifetimes, cancellation propagation, HTTP-version support, handler adapters and
middleware semantics are not interchangeable contracts. Native and adapted
streaming can have different writer, flush and body-ownership semantics;
comparison results do not establish streaming parity.

Do not duplicate router, middleware or lifecycle implementations without
measured production evidence that transport is the bottleneck. Supporting a
Fiber/fasthttp adapter requires a separately approved goal with an explicit
compatibility and maintenance commitment; comparison results alone do not
create that support obligation.

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
