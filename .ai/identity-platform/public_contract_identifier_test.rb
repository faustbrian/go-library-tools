# frozen_string_literal: true

require "json"
require "minitest/autorun"

class IdentityPlatformIdentifierContractTest < Minitest::Test
  def test_versioned_opaque_values_and_reachable_construction_match_the_goal
    row = metadata.fetch("external_contracts").find { |entry| entry.fetch("owner") == "identifier" }
    types = row.fetch("types").to_h { |entry| [entry.fetch("name"), entry] }
    assert_equal "opaque immutable value; unexported 1..4096 valid UTF-8 bytes and exact profile ID/version; returned by Normalize or validated RestoreCanonical; zero value invalid",
      types.fetch("identifier.Canonical").fetch("declaration")
    assert_equal "opaque immutable value; unexported validated NormalizationProfileSpec; constructed only by NewNormalizationProfile; zero value invalid",
      types.fetch("identifier.NormalizationProfile").fetch("declaration")
    assert_equal "struct{ID identifier.NormalizationProfileID; Version uint32; Unicode identifier.UnicodeNormalization; CaseFold bool; TrimSpace bool; IDNA bool; MaximumBytes uint32}",
      types.fetch("identifier.NormalizationProfileSpec").fetch("declaration")
    assert_equal "struct{Value string; ProfileID identifier.NormalizationProfileID; ProfileVersion uint32}",
      types.fetch("identifier.CanonicalRecord").fetch("declaration")
    [
      "identifier.NewNormalizationProfile(identifier.NormalizationProfileSpec) (identifier.NormalizationProfile, error)",
      "identifier.RestoreCanonical(identifier.CanonicalRecord, identifier.NormalizationProfile) (identifier.Canonical, error)",
      "identifier.Canonical.Record() identifier.CanonicalRecord",
      "identifier.Canonical.Equal(identifier.Canonical) bool"
    ].each { |method| assert_includes row.fetch("methods"), method }
    assert_includes types.fetch("identifier.Canonical").fetch("semantics"), "fixed point"
    assert_includes types.fetch("identifier.Canonical").fetch("semantics"), "Equal compares profile ID, version and bytes"
  end

  private

  def metadata
    JSON.parse(File.binread(File.join(__dir__, "fragments/public_contracts_meta.json")))
  end
end
