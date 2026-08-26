# Security Policy

Report vulnerabilities privately through GitHub Security Advisories for
`faustbrian/go-library-tools`. Do not include credentials, private repository
content, customer data, or production payloads in public issues.

Before `v1.0.0`, fixes are applied to `main`. After the first stable release,
the current major release receives security fixes. Reports are acknowledged,
triaged against affected versions, remediated, and disclosed in coordination
with the reporter.

The primary trust boundaries are untrusted repository content, external tool
output, downloaded release artifacts, verification evidence, service fixture
lifecycle, and pull-request workflows. See [docs/security.md](docs/security.md).
