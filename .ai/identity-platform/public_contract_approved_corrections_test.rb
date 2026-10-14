# frozen_string_literal: true

require "json"
require "minitest/autorun"

class IdentityPlatformApprovedCorrectionsTest < Minitest::Test
  def test_capability_preserves_wire_types_and_transaction_owner
    row = contract("capability")
    types = row.fetch("types").to_h { |type| [type.fetch("name"), type] }
    refute types.key?("capability.CaveatEntry")
    assert_includes types.fetch("capability.Payload").fetch("declaration"), "Caveats map[string]string"
    assert_includes types.fetch("capability.Proof").fetch("declaration"), "opaque"
    assert_includes types.fetch("capability.Proof").fetch("semantics"), "Limits.MaxTokenBytes"
    assert_includes row.fetch("methods"), "capability.ParseProof(string, capability.Limits) (capability.Proof, error)"
    assert_includes row.fetch("methods"), "capability.Proof.Transfer(context.Context, func(context.Context, []byte) error) error"
    assert_includes row.fetch("errors"), "capability.ErrProofTransferFailed"
    assert_includes types.fetch("capability.Proof").fetch("semantics"), "all copied aliases"
    service = row.fetch("interfaces").find { |entry| entry.fetch("name") == "capability.Service" }
    assert_includes service.fetch("semantics"), "does not consume"
    assert_includes service.fetch("semantics"), "protected domain transaction"
  end

  def test_password_separates_intrinsic_validation_from_policy_envelope
    type = contract("password").fetch("types").find { |entry| entry.fetch("name") == "golibpassword.Parameters" }
    assert_includes type.fetch("semantics"), "NewParameters validates intrinsic"
    assert_includes type.fetch("semantics"), "NewPolicy"
    assert_includes type.fetch("semantics"), "not universal maxima"
  end

  def test_capability_preserves_existing_timestamp_ordering
    type = contract("capability").fetch("types").find { |entry| entry.fetch("name") == "capability.Payload" }
    semantics = type.fetch("semantics")
    assert_includes semantics, "NotBefore may precede IssuedAt"
    assert_includes semantics, "ExpiresAt after both"
    assert_includes semantics, "earlier of IssuedAt and NotBefore"
    refute_includes semantics, "IssuedAt no later than NotBefore"
  end

  private

  def contract(owner)
    JSON.parse(File.binread(File.join(__dir__, "fragments/public_contracts_meta.json")))
      .fetch("external_contracts").find { |entry| entry.fetch("owner") == owner }
  end
end
