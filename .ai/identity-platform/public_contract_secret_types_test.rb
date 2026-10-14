# frozen_string_literal: true

require "json"
require "minitest/autorun"

class IdentityPlatformPublicContractSecretTypesTest < Minitest::Test
  def test_catalog_preserves_the_goal_secret_contracts
    assert_opaque_codes(metadata)
  end

  def test_defined_string_counterfactuals_violate_the_contract
    %w[authentication.OneTimeCode authentication.RecoveryCode].each do |name|
      changed = metadata
      secret_type(changed, name)["declaration"] = "defined string"

      error = assert_raises(Minitest::Assertion) { assert_opaque_codes(changed) }
      assert_includes error.message, name
    end
  end

  def test_disclosure_semantics_counterfactuals_violate_the_contract
    %w[authentication.OneTimeCode authentication.RecoveryCode].each do |name|
      changed = metadata
      secret_type(changed, name)["semantics"] = "Secret may be formatted."

      error = assert_raises(Minitest::Assertion) { assert_opaque_codes(changed) }
      assert_includes error.message, name
    end
  end

  private

  def metadata
    JSON.parse(File.binread(File.join(__dir__, "fragments/public_contracts_meta.json")))
  end

  def contract(document)
    document.fetch("external_contracts").find { |row| row.fetch("owner") == "authentication" }
  end

  def secret_type(document, name)
    contract(document).fetch("types").find { |row| row.fetch("name") == name }
  end

  def assert_opaque_codes(document)
    {
      "authentication.OneTimeCode" => ["4..128 ASCII symbols", "NewOneTimeCode"],
      "authentication.RecoveryCode" => ["8..256 canonical ASCII symbols", "NewRecoveryCode"]
    }.each do |name, (bounds, constructor)|
      row = secret_type(document, name)
      expected = "opaque struct with unexported immutable secret storage; #{bounds}; " \
        "constructed only by authentication.#{constructor}; zero value invalid"
      assert_equal expected, row.fetch("declaration"), "#{name} must remain opaque and constructor-only"
      assert_includes row.fetch("semantics"), "no String, text, JSON, logging or formatting implementation",
        "#{name} must prohibit secret disclosure"
      signature = "authentication.#{constructor}(string) (#{name}, error)"
      assert_includes contract(document).fetch("methods"), signature, "#{name} needs a reachable constructor"
    end
  end
end
