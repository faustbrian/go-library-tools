# Fixed identifier Unicode data decision

The user selected "Use fixed upstream tables; preserve newer Go support" in
response to the confirmed compiler-dependent normalization defect. This
authorizes the existing identifier worker to use the smallest private,
provenance-tracked upstream algorithm/data closure needed for compiler-independent
Unicode 15 behavior, preserving the pinned Go 1.26 baseline and public profile
contract on newer compilers. It does not authorize rejecting all newer compilers,
inventing Unicode algorithms, changing stored profile IDs or versions, weakening
verification gates, or treating copied upstream code as exempt from coverage.

Norm, casefold, IDNA, bidi and whitespace semantics must share the fixed data
authority; compiler build tags, standard-library Unicode changes and consumer
MVS upgrades must not silently change a profile's results. Source revisions,
checksums, licenses, extraction commands and security-update ownership remain
explicit. New versions of the normalization policy require deliberate profile
version/migration behavior, not reinterpretation of persisted records.

The existing owner, generation, branch and work are retained. A fresh bounded
mapper and QA input cover this material strategy change before the first copied
engine write. Their reports are implementation input, not an approval series.
