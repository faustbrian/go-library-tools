# frozen_string_literal: true

require "digest"
require "json"
require "minitest/autorun"

load File.join(__dir__, "validate.rb")

class IdentityPlatformEnvironmentInputIdentityTest < Minitest::Test
  def test_reads_committed_git_blobs_as_raw_bytes
    path = ".ai/identity-platform/INVENTORY.md"

    readers = [
      method(:git_blob_bytes), method(:git_worker_blob_bytes), method(:git_repository_blob_bytes)
    ]

    readers.each do |reader|
      actual = reader.call("HEAD", path)

      assert_equal Encoding::BINARY, actual.encoding
      assert_equal File.binread(path), actual
    end
  end

  def test_matches_previous_snapshot_fixtures_by_raw_bytes
    paths = %w[
      .ai/identity-platform/INVENTORY.md
      .ai/identity-platform/EXECUTION_LEDGER.md
      .ai/identity-platform/PREFLIGHT_EVIDENCE.md
      .ai/identity-platform/GOAL_MANIFEST.json
    ]
    previous_sources = paths.to_h { |path| [path, "fixture — #{path}\n"] }

    errors = previous_snapshot_binding_errors(
      previous_sources: previous_sources,
      history_required: true,
      candidate_fixture_mode: true,
      read_blob: ->(_revision, path) { previous_sources.fetch(path).b }
    )

    assert_empty errors
  end

  def test_accepts_an_exact_inventory_and_ledger_transition_envelope
    changed = %w[
      .ai/identity-platform/EXECUTION_LEDGER.md
      .ai/identity-platform/INVENTORY.md
    ]

    assert coordinator_control_envelope_allowed?(
      changed,
      parent: "49fbf1d1928dfa97c6b8a929e0c649bb49332953",
      commit: "31550aef19720d448cdcef11ce686c0a986f207d",
      complete_changed_paths: changed
    )
  end

  def test_accepts_an_exact_validator_repair_envelope_without_product_authority
    changed = %w[
      .ai/identity-platform/environment_input_identity_test.rb
      .ai/identity-platform/validate.rb
    ]

    assert coordinator_control_envelope_allowed?(
      changed,
      parent: "31550aef19720d448cdcef11ce686c0a986f207d",
      commit: "53b716bf0a48fcfaeeb95498ed33b2769505c992",
      complete_changed_paths: changed
    )
    changed.each do |path|
      assert coordinator_validation_implementation_path?(path)
      refute revision_product_semantic_source?(path, nil, nil)
    end
    assert revision_product_semantic_source?(".ai/identity-platform/PUBLIC_CONTRACTS.json", nil, nil)
    refute coordinator_control_envelope_allowed?(
      changed + [".ai/identity-platform/evidence/unbound.json"],
      parent: "31550aef19720d448cdcef11ce686c0a986f207d",
      commit: "53b716bf0a48fcfaeeb95498ed33b2769505c992",
      attributed_artifact_paths: Set[".ai/identity-platform/evidence/unbound.json"],
      complete_changed_paths: changed + [".ai/identity-platform/evidence/unbound.json"]
    )
    refute coordinator_control_envelope_allowed?(
      changed,
      parent: "31550aef19720d448cdcef11ce686c0a986f207d",
      commit: "53b716bf0a48fcfaeeb95498ed33b2769505c992",
      complete_changed_paths: changed + ["README.md"]
    )
  end

  def test_canonicalizes_symbol_keyed_environment_lanes
    tool_lanes = [{
      requirement: "PostgreSQL profile",
      required_by: "PostgreSQL adapters",
      classification: "available",
      identity: "postgres:18.4-alpine"
    }]
    expected = [{
      "classification" => "available",
      "identity" => "postgres:18.4-alpine",
      "required_by" => "PostgreSQL adapters",
      "requirement" => "PostgreSQL profile"
    }]

    external_lanes = [{
      profile: "captcha-turnstile",
      consumers: ["identity/risk/captcha/turnstile"],
      claims: ["journey.abuse-controls.v1"],
      classification: "not-yet-needed"
    }]
    expected_external = [{
      "claims" => ["journey.abuse-controls.v1"],
      "classification" => "not-yet-needed",
      "consumers" => ["identity/risk/captcha/turnstile"],
      "profile" => "captcha-turnstile"
    }]

    identities = expected_environment_input_identities(
      external_lanes: external_lanes,
      tool_lanes: tool_lanes,
      environment: {}
    )

    assert_equal(
      "sha256:#{Digest::SHA256.hexdigest(JSON.generate(expected))}",
      identities.fetch("environment:service-images")
    )
    assert_equal(
      "sha256:#{Digest::SHA256.hexdigest(JSON.generate(expected_external))}",
      identities.fetch("environment:external-profiles")
    )
  end
end
