# frozen_string_literal: true

require "digest"
require "json"
require "minitest/autorun"

if File.file?(File.join(__dir__, "primitive_preflight_evidence.rb"))
  require_relative "primitive_preflight_evidence"
else
  module IdentityPlatformPrimitivePreflightEvidence
    def self.errors(**)
      ["primitive preflight evidence validation is unavailable"]
    end
  end
end

class IdentityPlatformPrimitivePreflightEvidenceTest < Minitest::Test
  Evidence = IdentityPlatformPrimitivePreflightEvidence
  REVISION = "a" * 40
  TREE = "git-sha1:#{'b' * 40}"
  TOOL_DIGEST = "sha256:#{'c' * 64}"

  def test_accepts_exact_closed_stale_record
    assert_empty errors_for(document)
  end

  def test_rejects_substitution_and_unbound_fingerprints
    {
      "consumer substitution" => lambda { |copy| copy.dig("records", 0, "consumers") << "identity/session" },
      "repository substitution" => lambda { |copy| copy.dig("records", 0, "repository_revision").replace("d" * 40) },
      "API fingerprint substitution" => lambda { |copy| copy.dig("records", 0, "api_input_fingerprint").replace("sha256:#{'d' * 64}") },
      "gate fingerprint substitution" => lambda { |copy| copy.dig("records", 0, "gate_input_fingerprint").replace("sha256:#{'d' * 64}") },
      "tool digest substitution" => lambda { |copy| copy["tool_environment_sha256"] = "sha256:#{'d' * 64}" },
      "extra record field" => lambda { |copy| copy.dig("records", 0)["unexpected"] = true }
    }.each do |label, mutate|
      copy = JSON.parse(JSON.generate(document))
      mutate.call(copy)
      refute_empty errors_for(copy), label
    end
  end

  def test_requires_executed_result_to_match_exit_status_and_revision
    executed = document
    observation = executed.dig("records", 0, "gate_observation")
    observation.replace(
      "classification" => "executed", "exit_status" => 0,
      "log_path" => ".ai/identity-platform/evidence/preflight/primitive-gates/audit.log",
      "log_sha256" => "sha256:#{'e' * 64}", "summary" => "passed",
      "tested_revision" => REVISION
    )
    executed.dig("records", 0)["gate_result"] = "pass"
    assert_empty errors_for(executed)

    executed.dig("records", 0)["gate_result"] = "failed"
    refute_empty errors_for(executed)
    executed.dig("records", 0)["gate_result"] = "pass"
    observation["tested_revision"] = "f" * 40
    refute_empty errors_for(executed)
  end

  def test_rejects_executed_result_without_custodied_log
    executed = document
    observation = executed.dig("records", 0, "gate_observation")
    observation.replace(
      "classification" => "executed", "exit_status" => 1,
      "log_path" => nil, "log_sha256" => nil, "summary" => "failed",
      "tested_revision" => REVISION
    )
    executed.dig("records", 0)["gate_result"] = "failed"

    refute_empty errors_for(executed)
  end

  def test_source_loader_rejects_noncanonical_schema
    source = JSON.pretty_generate(document.merge("schema" => "substitute.v1")) + "\n"

    _parsed, errors = Evidence.parse_source(source)

    refute_empty errors
  end

  def test_tool_environment_authority_is_closed_and_control_bound
    assert_empty Evidence.tool_environment_errors(
      tool_environment_document,
      expected_control_revision: "1" * 40,
      expected_recorded_at: "2026-08-30T10:32:30Z",
      expected_tool_identities: tool_identities
    )

    substituted = JSON.parse(JSON.generate(tool_environment_document))
    substituted["control_revision"] = "2" * 40
    refute_empty Evidence.tool_environment_errors(
      substituted,
      expected_control_revision: "1" * 40,
      expected_recorded_at: "2026-08-30T10:32:30Z",
      expected_tool_identities: tool_identities
    )

    substituted = JSON.parse(JSON.generate(tool_environment_document))
    substituted.dig("service_profiles", 0)["digest"] = "sha256:#{'d' * 64}"
    refute_empty Evidence.tool_environment_errors(
      substituted,
      expected_control_revision: "1" * 40,
      expected_recorded_at: "2026-08-30T10:32:30Z",
      expected_tool_identities: tool_identities
    )

    substituted = JSON.parse(JSON.generate(tool_environment_document))
    substituted["untrusted"] = true
    refute_empty Evidence.tool_environment_errors(
      substituted,
      expected_control_revision: "1" * 40,
      expected_recorded_at: "2026-08-30T10:32:30Z",
      expected_tool_identities: tool_identities
    )

    substituted = JSON.parse(JSON.generate(tool_environment_document))
    substituted.dig("tools", 0)["identity"] = "substituted"
    refute_empty Evidence.tool_environment_errors(
      substituted,
      expected_control_revision: "1" * 40,
      expected_recorded_at: "2026-08-30T10:32:30Z",
      expected_tool_identities: tool_identities
    )
  end

  private

  def document
    record = {
      "primitive" => "audit",
      "consumers" => ["identity"],
      "registered" => "audit",
      "repository_identity" => "github.com/faustbrian/go-audit",
      "repository_revision" => REVISION,
      "repository_tree" => TREE,
      "module_relative_path" => ".",
      "api_input_algorithm" => "identity-platform-primitive-api-input-v1",
      "api_input_fingerprint" => "",
      "gate_input_algorithm" => "identity-platform-primitive-gate-input-v1",
      "gate_input_fingerprint" => "",
      "gate_command" => ["make", "check"],
      "gate_working_directory" => ".",
      "gate_result" => "stale",
      "gate_observation" => {
        "classification" => "not-executed", "exit_status" => nil,
        "log_path" => nil, "log_sha256" => nil, "summary" => "not run", "tested_revision" => nil
      }
    }
    record["api_input_fingerprint"] = Evidence.api_input_fingerprint(record) if Evidence.respond_to?(:api_input_fingerprint)
    record["gate_input_fingerprint"] = Evidence.gate_input_fingerprint(record, TOOL_DIGEST) if Evidence.respond_to?(:gate_input_fingerprint)
    {
      "schema_version" => 1,
      "schema" => "identity-platform.preflight-primitive-contracts.v1",
      "observed_control_revision" => "1" * 40,
      "recorded_at" => "2026-08-30T10:32:30Z",
      "tool_environment_path" => ".ai/identity-platform/evidence/preflight/tool-environment-inventory.json",
      "tool_environment_sha256" => TOOL_DIGEST,
      "api_input_scope" => "complete tracked repository tree at the recorded revision; conservative changes invalidate the fingerprint",
      "gate_input_scope" => "complete tracked repository tree, module label, command, and tool-environment artifact digest",
      "records" => [record]
    }
  end

  def errors_for(value)
    Evidence.errors(
      document: value,
      expected_consumers: {"audit" => ["identity"]},
      expected_repositories: {
        "audit" => {
          "repository_identity" => "github.com/faustbrian/go-audit",
          "repository_revision" => REVISION,
          "repository_tree" => TREE,
          "module_relative_path" => "."
        }
      },
      tool_environment_path: ".ai/identity-platform/evidence/preflight/tool-environment-inventory.json",
      tool_environment_sha256: TOOL_DIGEST,
      artifact_sha256: lambda do |path|
        path == ".ai/identity-platform/evidence/preflight/primitive-gates/audit.log" ? "sha256:#{'e' * 64}" : nil
      end
    )
  end

  def tool_environment_document
    {
      "schema_version" => 1,
      "schema" => "identity-platform.preflight-tool-environment.v1",
      "control_revision" => "1" * 40,
      "recorded_at" => "2026-08-30T10:32:30Z",
      "platform" => {
        "os" => "darwin", "architecture" => "arm64", "cgo_enabled" => true,
        "go_toolchain_policy" => "auto"
      },
      "tools" => [
        {"id" => "go", "classification" => "available", "identity" => "go1.27.0 darwin/arm64"},
        {"id" => "ruby", "classification" => "available", "identity" => "ruby 4.0.6 arm64-darwin25"},
        {"id" => "go-library-tools", "classification" => "available", "identity" => "source:#{'1' * 40} binary-sha256:#{'a' * 64}"},
        {"id" => "codex", "classification" => "available", "identity" => "codex-cli 0.151.0"},
        {"id" => "docker", "classification" => "available", "identity" => "client 29.7.2 server 29.7.2"},
        {"id" => "gitleaks", "classification" => "available", "identity" => "/tool/gitleaks"},
        {"id" => "golangci-lint", "classification" => "available", "identity" => "/tool/golangci-lint"},
        {"id" => "staticcheck", "classification" => "available", "identity" => "/tool/staticcheck"},
        {"id" => "govulncheck", "classification" => "available", "identity" => "/tool/govulncheck"},
        {"id" => "gremlins", "classification" => "available", "identity" => "/tool/gremlins"}
      ],
      "service_profiles" => [
        {"id" => "postgresql", "classification" => "available", "image" => "postgres:18.4-alpine", "digest" => "sha256:9a8afca54e7861fd90fab5fdf4c42477a6b1cb7d293595148e674e0a3181de15"},
        {"id" => "valkey", "classification" => "available", "image" => "valkey/valkey:9.1.0-alpine", "digest" => "sha256:c9b77919daeba2c02ad954d0c844cc4e7142069d177b89c5fd771f405daf9e02"}
      ],
      "harnesses" => [
        {"id" => "race-fuzz-leak-mutation", "classification" => "available", "identity" => "available"},
        {"id" => "browser", "classification" => "unavailable", "blocker" => "blocker:browser-harness"},
        {"id" => "interoperability", "classification" => "not-yet-needed", "identity" => "pinned"}
      ],
      "limitations" => [
        "Tool presence and image availability are preflight facts, not passing package or interoperability evidence.",
        "Secret-bearing environment values were neither read nor recorded."
      ]
    }
  end

  def tool_identities
    tool_environment_document.fetch("tools").to_h do |tool|
      [tool.fetch("id"), tool.fetch("identity")]
    end
  end
end
