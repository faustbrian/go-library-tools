# frozen_string_literal: true

require "minitest/autorun"
require "set"
require_relative "composition_dependencies"

class IdentityPlatformCompositionDependenciesTest < Minitest::Test
  # Independent inventory of the unchanged all-feature reference contract.
  FEATURES = %w[
    identity identity/session identity/delivery identity/risk identity/risk/captcha
    identity/password identity/username identity/email identity/magiclink identity/otp
    identity/phone identity/anonymous identity/mfa webauthn passkey identity/oauth
    identity/oauth/providers identity/oauth/onetap identity/oauth/proxy identity/apikey
    identity/impersonation organization sso sso/oidc sso/oauth2 sso/saml
    sso/domain-verification scim scim/organization oauth-server oauth-server/oidc
    oauth-server/device identity/i18n primitive/authorization-identity-contracts
  ].freeze
  ADAPTERS = %w[
    identity/postgres identity/session/postgres identity/session/valkey
    identity/delivery/postgres identity/anonymous/postgres identity/risk/postgres
    identity/risk/valkey identity/risk/captcha/recaptcha identity/risk/captcha/turnstile
    identity/risk/captcha/hcaptcha identity/risk/captcha/captchafox identity/risk/hibp
    identity/password/postgres identity/otp/postgres identity/mfa/postgres
    passkey/postgres webauthn/postgres identity/oauth/postgres identity/apikey/postgres
    identity/apikey/valkey identity/impersonation/postgres organization/postgres
    sso/postgres scim/postgres oauth-server/postgres
  ].freeze

  def test_base_only_http_and_all_feature_reference_are_accepted
    assert_empty policy_errors(complete_rows)
  end

  def test_each_required_http_base_dependency_is_enforced
    %w[identity primitive/authorization-identity-contracts].each do |dependency|
      rows = complete_rows
      rows.find { |row| row[:unit] == "identity/http" }[:requires].delete(dependency)
      assert policy_errors(rows).any? { |error| error.include?("identity/http") }, dependency
    end
  end

  def test_optional_features_cannot_become_http_bootstrap_requirements
    rows = complete_rows
    rows.find { |row| row[:unit] == "identity/http" }[:requires] << "identity/password"
    assert policy_errors(rows).any? { |error| error.include?("identity/http") }
  end

  def test_reference_cannot_lose_any_feature_or_adapter
    (FEATURES + ADAPTERS).uniq.each do |unit|
      rows = complete_rows
      rows.find { |row| row[:unit] == "identity/reference" }[:requires].delete(unit)
      assert policy_errors(rows).any? { |error| error.include?(unit) }, unit
    end
  end

  def test_actual_inventory_preserves_the_full_reference_proof_closure
    rows = File.readlines(File.join(__dir__, "INVENTORY.md")).filter_map do |line|
      cells = line.split("|", -1).map(&:strip)
      next unless cells.length == 8 && cells[1].start_with?("`")
      {unit: cells[1].delete("`"), requires: cells[3].scan(/`([^`]+)`/).flatten}
    end
    assert_empty policy_errors(rows)
  end

  def test_reference_rejects_prerequisites_outside_its_composition_contract
    rows = complete_rows
    rows.find { |row| row[:unit] == "identity/reference" }[:requires] << "primitive/authentication-identity-contracts"
    assert policy_errors(rows).any? { |error| error.include?("extra=primitive/authentication-identity-contracts") }
  end

  private

  def complete_rows
    requirements = (FEATURES + ADAPTERS + ["identity/http", "primitive/capability-postgres-identity-contracts"]).uniq
    rows = requirements.map { |unit| {unit: unit, requires: []} }
    rows.find { |row| row[:unit] == "identity/http" }[:requires] = %w[identity primitive/authorization-identity-contracts]
    rows << {unit: "identity/reference", requires: requirements.dup}
    rows
  end

  def policy_errors(rows)
    IdentityPlatformCompositionDependencies.errors(rows)
  end
end
