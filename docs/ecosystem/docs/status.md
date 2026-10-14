# Capability status

[Documentation map](index.md)

Status is per module and version, not an ecosystem marketing label. The
[consumer catalog](../catalog-consumer.md) owns frozen active/deprecated module
facts; the [engineering catalog](../catalog-engineering.md) also records
planned and non-releasable work. A declared `go.mod`, source implementation or
passing local test is not a published installable release.

| State | Adoption meaning |
| --- | --- |
| Active, published | Install the exact released module for its documented guarantees. Check maturity and supported backend/specification scope separately. |
| Experimental | Do not infer stable API, persistence or production guarantees. Obtain the owner-specific compatibility contract. |
| Deprecated compatibility path | Existing consumers retain the stated support. New consumers prefer a released canonical successor; removal requires the authorized major/migration conditions. |
| Planned or unavailable | No installation route or runtime capability is promised. Identity-platform boundaries and Secret Store stay engineering-only. |
| Non-production reference | Executable composition evidence, not a deployable product or production dependency. |
| Blocked release or unresolved decision | Read the material caveat and owner; prose cannot supply a preferred API or waive a release gate. |

Fourteen recipes are bound to `golib-compat-v1-20260912.2`; the
[recipe index](recipes/index.md) makes each execution boundary explicit.
OpenAPI generated-contract behavior, Kafka broker interoperability, live
OpenSearch behavior and deployed Track/Postal/Location business behavior are
not inferred from those fixture results. New illustrative stacks do not join
the published cohort by appearing in documentation.

Golib Service supports `net/http`; the
[Fiber/fasthttp architecture comparison](comparisons/service-and-logging.md#separate-architecture-track-fiber-and-fasthttp)
does not promise a supported adapter or extend the frozen compatibility cohort.

**Verkle Tree remains experimental/non-production for every published v1
version and current source.** Do not use it to protect production integrity or
confidentiality. Stable API versions and the historical catalog's active/stable
labels are not security qualification; the [limitations register](limitations.md)
links the owner's unresolved risks and production-adoption revisit conditions.

The [limitations register](limitations.md) preserves residual migration,
upstream authority and external usage caveats. Newer independently published
majors remain distinct from the frozen cohort; see [versioning](versioning.md).

Next: [package catalog](packages.md), [limitations](limitations.md),
[versioning](versioning.md).
