# Security Policy

Report vulnerabilities privately through GitHub Security Advisories for
`faustbrian/go-library-tools`. Do not include credentials, private repository
content, customer data, or production payloads in public issues.

Include the affected module and version, impact, prerequisites, and the
smallest safe reproduction. Maintainers acknowledge and triage reports under
the ecosystem
[vulnerability-management process](docs/ecosystem/security/vulnerability-management.md),
coordinate embargo and disclosure with the reporter, and never publish private
reporter data or exploit-enabling secrets.

Before `v1.0.0`, fixes are applied to `main`. After the first stable release,
the current major release receives security fixes. Security fixes are scoped to
affected modules and do not force unrelated package releases. Advisories
identify exact affected and fixed versions and include upgrade guidance when
disclosure is safe.

The primary trust boundaries are untrusted repository content, external tool
output, downloaded release artifacts, verification evidence, service fixture
lifecycle, and pull-request workflows. See [docs/security.md](docs/security.md).
