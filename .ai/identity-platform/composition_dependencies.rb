# frozen_string_literal: true

require "set"

module IdentityPlatformCompositionDependencies
  HTTP_BASE = Set["identity", "primitive/authorization-identity-contracts"].freeze
  REFERENCE_FEATURES = Set[
    "primitive/authorization-identity-contracts",
    "identity", "identity/session", "identity/delivery", "identity/risk",
    "identity/risk/captcha", "identity/password", "identity/username",
    "identity/email", "identity/magiclink", "identity/otp", "identity/phone",
    "identity/anonymous", "identity/mfa", "webauthn", "passkey",
    "identity/oauth", "identity/oauth/providers", "identity/oauth/onetap",
    "identity/oauth/proxy", "identity/apikey", "identity/impersonation",
    "organization", "sso", "sso/oidc", "sso/oauth2", "sso/saml",
    "sso/domain-verification", "scim", "scim/organization", "oauth-server",
    "oauth-server/oidc", "oauth-server/device", "identity/i18n"
  ].freeze
  REFERENCE_ADAPTERS = Set[
    "identity/postgres", "identity/session/postgres", "identity/session/valkey",
    "identity/delivery/postgres", "identity/anonymous/postgres",
    "identity/risk/postgres", "identity/risk/valkey",
    "identity/risk/captcha/recaptcha", "identity/risk/captcha/turnstile",
    "identity/risk/captcha/hcaptcha", "identity/risk/captcha/captchafox",
    "identity/risk/hibp", "identity/password/postgres",
    "identity/otp/postgres", "identity/mfa/postgres", "passkey/postgres",
    "webauthn/postgres", "identity/oauth/postgres", "identity/apikey/postgres",
    "identity/apikey/valkey", "identity/impersonation/postgres",
    "organization/postgres", "sso/postgres", "scim/postgres", "oauth-server/postgres"
  ].freeze
  REFERENCE_REQUIRED = (REFERENCE_FEATURES | REFERENCE_ADAPTERS | Set[
    "identity/http", "primitive/capability-postgres-identity-contracts"
  ]).freeze

  module_function

  # HTTP bootstrap is deliberately independent of optional features. The
  # all-feature reference owns their eventual composed proof and dependencies.
  def errors(rows)
    errors = []
    http = rows.find { |row| row[:unit] == "identity/http" }
    reference = rows.find { |row| row[:unit] == "identity/reference" }
    return ["identity/http or identity/reference inventory row is missing"] unless http && reference

    unless http[:requires].to_set == HTTP_BASE
      errors << "identity/http base dependency set drifted"
    end
    actual = reference[:requires].to_set
    missing = REFERENCE_REQUIRED - actual
    extra = actual - REFERENCE_REQUIRED
    unless missing.empty? && extra.empty?
      errors << "identity/reference composition dependency set drifted: missing=#{missing.to_a.sort.join(',')} extra=#{extra.to_a.sort.join(',')}"
    end
    errors
  end
end
