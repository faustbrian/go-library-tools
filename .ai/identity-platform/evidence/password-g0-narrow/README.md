# Password: focused external-input closure

The coordinator preserved the successful original configured `make ci`
execution at `a3dbd862340c48376fab7285cae33ee23cd09409`, native session
20865. Only two unbound external inputs needed fresh execution: host PHP and
the live vulnerability database. No coverage, mutation or package matrix was
repeated for a bookkeeping revision.

Before the focused executions the coordinator captured `input-manifest.json`
using the existing repository input builder. It contains 110 immutable package
inputs, selected control requirements/readers, and environment identities.
`environment-inputs.json` retains the canonical JSON preimages of those
identities. The builder independently rebuilt the manifest and every preimage
hash matched. The complete root is
`sha256:add88858a2e1462cab1aee682de07162fb4b8bff957e22e12faf03d4f1504eca`.
This reconstruction does not pretend a manifest existed before session 20865;
original aggregate provenance and new focused executions remain distinct.

At the unchanged, clean package integration revision, native coordinator
session 99821 executed with Go 1.26.6, GOWORK=off, GOFLAGS=-p=2,
GOMAXPROCS=2 and disposable build/module/temp caches:

- `/usr/bin/make -f verification/package.mk interoperability`: exit 0,
  2026-10-01T09:29:49Z through 09:29:53Z. This executes the unchanged
  repository's bidirectional PHP/Go Argon2id and bcrypt script. PHP 8.5.10,
  its exact executable checksum, static crypto version information, dynamic
  linkage and macOS build were captured before execution.
- Captured absolute `govulncheck -db file:///tmp/identity-password-narrow-gates.kqVk5A/db -json ./...`:
  exit 0, 09:29:53Z through 09:29:54Z. The official pinned scanner is
  v1.6.0. Its exact build/module checksums and executable hash were captured.
  The official database ZIP was downloaded once, hashed, extracted, and read
  through the scanner's supported file database mode; no live feed was used
  during scanning. Raw ZIP and all 208 native JSON messages are retained.

The scan has no affected imported-package or called-symbol finding. Four
required-module-only advisories remain: GO-2026-5932, GO-2026-6303,
GO-2026-6354 and GO-2026-6355. This is not zero dependency advisories.
The database index reports last modification 2026-09-28T16:43:40Z.
Raw artifact hashes were verified during custody transfer; ZIP identity is
`faa8c45538d259ae1fd3c34ae20632ee738d1190836c84ac95d5f2bc604d85bc`.

The default commit-hygiene secret scan reported 60 Sourcegraph-token matches
in the raw vulnerability stream. Each was independently classified as an
official public `go.googlesource.com/go/+/<Git revision>` or
`go.googlesource.com/net/+/<Git revision>` URL, not a credential. The stream
SHA-256 is `119103bde334c1bb1752eab1bc6771c00ad58e8bdaca92b4b17850c860a89c73`.
The default scanner's nonzero result is preserved as false positives, not
represented as an exit-zero scan; no rule or artifact content was changed.

Both checks succeeded; the enclosing shell initially exited 1 because its
cleanup trap encountered read-only downloaded Go module permissions.
The coordinator repaired permissions on the exact task-owned module cache,
removed all caches, binaries, database extraction and the temporary root, and
confirmed its absence. This cleanup failure is not reported as a check failure
or concealed as an initially successful disposal. Package source stayed clean.

All five existing v1 production self-consumers were exercised by the original
coordinator gate. The five identity consumers named by the goal—identity/otp,
identity/password, identity/phone, identity/risk/hibp and identity/username—are
proposed/unbuilt; their migrations belong to their own goals and are not
certified here. The existing main tools compatibility consumer imports
go-password/v2 v2.0.0, not this v1 module; it is not an affected v1 reverse gate.
Future consumers must use the integrated constructive Secret/Parameters API
when assigned; no raw-byte migration or downstream HIBP composition is claimed.

Native independent `/root/password_delivery_g0` approved this complete closure
for local unit verification and DAG prerequisite eligibility after inspecting
the original native QA, regression, correctness and Go reports, immutable
source, the manifest rebuild and native executions. The Go principal's
clerical correction identifies current root mutation report
`bb6cca6fadd4213f839c103ad4afee716cd1e010cdd9ed19b7fe3f86522cd18c`;
its earlier `7cadf9...` attribution belongs to a different historical input.
The original report remains preserved alongside its native correction.

Inventory and ledger record local verified status only. Neither these local
checks nor the original aggregate execution proves publication, hosted CI,
deployment, downstream migration, the shared-issuer milestone or full-program
completion. This status change unlocks only consumers whose other prerequisites
are also integrated and verified; it does not satisfy their remaining gates.
