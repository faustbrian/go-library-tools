# frozen_string_literal: true

require "minitest/autorun"
if File.file?(File.join(__dir__, "external_evidence_v3_validator.rb"))
  require_relative "external_evidence_v3_validator"
else
  module IdentityPlatformExternalEvidenceV3Validator
    class InvalidEvidence < StandardError; end

    def self.validate!(*); end
  end
end

class IdentityPlatformExternalEvidenceV3ValidatorTest < Minitest::Test
  Validator = IdentityPlatformExternalEvidenceV3Validator

  CONTROL = "github.com/faustbrian/go-library-tools"
  PACKAGE = "github.com/faustbrian/go-identity"
  CONTROL_REVISION = "a" * 40
  PACKAGE_REVISION = "b" * 40
  GATE_REVISION = "c" * 40
  REVALIDATION_REVISION = "d" * 40
  PACKAGE_TREE = "git-sha1:#{"4" * 40}"
  GATE_TREE = "git-sha1:#{"5" * 40}"
  REVALIDATION_TREE = "git-sha1:#{"6" * 40}"
  RECORD_REVISION = "e" * 40
  RECEIPT_REVISION = "f" * 40

  def test_accepts_two_exact_repository_snapshots_and_repository_scoped_custody_and_ancestry
    calls = []

    Validator.validate!(
      evidence: valid_evidence,
      expected_repositories: repository_bindings,
      ancestor: lambda { |repository_identity:, ancestor_revision:, descendant_revision:|
        calls << [repository_identity, ancestor_revision, descendant_revision]
        true
      }
    )

    assert_equal [
      [PACKAGE, PACKAGE_REVISION, GATE_REVISION],
      [PACKAGE, GATE_REVISION, REVALIDATION_REVISION]
    ], calls
  end

  def test_rejects_schema_v2_and_legacy_scalar_or_array_input_manifests
    legacy_documents = [
      valid_evidence.merge("schema_version" => 2),
      valid_evidence.merge("input_manifest" => {
        "schema_version" => 2,
        "repository_revision" => PACKAGE_REVISION,
        "repository_tree" => package_binding.fetch("tree")
      }),
      valid_evidence.merge("input_manifest" => repository_snapshots)
    ]

    legacy_documents.each do |evidence|
      assert_invalid do
        validate(evidence)
      end
    end
  end

  def test_rejects_omitted_duplicated_or_substituted_repository_bindings
    omitted = deep_copy(valid_evidence)
    omitted.fetch("repositories").reject! { |row| row.fetch("role") == "package" }

    duplicated = deep_copy(valid_evidence)
    duplicated.fetch("repositories") << deep_copy(duplicated.fetch("repositories").last)

    substituted = deep_copy(valid_evidence)
    substituted.fetch("repositories").last["repository_identity"] =
      "github.com/faustbrian/go-substitute"

    [omitted, duplicated, substituted].each do |evidence|
      assert_invalid do
        validate(evidence)
      end
    end
  end

  def test_rejects_omitted_duplicated_or_substituted_input_manifest_snapshots
    omitted = deep_copy(valid_evidence)
    omitted.fetch("input_manifest").fetch("repositories").reject! do |row|
      row.fetch("repository_identity") == PACKAGE
    end

    duplicated = deep_copy(valid_evidence)
    duplicated.fetch("input_manifest").fetch("repositories") <<
      deep_copy(duplicated.fetch("input_manifest").fetch("repositories").last)

    substituted = deep_copy(valid_evidence)
    substituted.fetch("input_manifest").fetch("repositories").last["repository_identity"] =
      "github.com/faustbrian/go-substitute"

    [omitted, duplicated, substituted].each do |evidence|
      assert_invalid do
        validate(evidence)
      end
    end
  end

  def test_rejects_wrong_package_revision_tree_or_module_workdir
    mutations = {
      "revision" => "1" * 40,
      "tree" => "git-sha1:#{"2" * 40}",
      "module_workdir" => "/coord/.worktrees/go-library-tools-identity-platform"
    }

    mutations.each do |field, replacement|
      evidence = deep_copy(valid_evidence)
      evidence.fetch("repositories").find { |row| row.fetch("role") == "package" }[field] = replacement

      assert_invalid(field) do
        validate(evidence)
      end
    end
  end

  def test_rejects_input_manifest_snapshot_revision_or_tree_drift
    %w[revision tree].each do |field|
      evidence = deep_copy(valid_evidence)
      snapshot = evidence.fetch("input_manifest").fetch("repositories").find do |row|
        row.fetch("repository_identity") == PACKAGE
      end
      snapshot[field] = field == "revision" ? "1" * 40 : "git-sha1:#{"2" * 40}"

      assert_invalid(field) do
        validate(evidence)
      end
    end
  end

  def test_rejects_package_ancestry_tree_drift
    evidence = deep_copy(valid_evidence)
    evidence.fetch("ancestry").last["ancestor_tree"] = "git-sha1:#{"7" * 40}"

    assert_invalid do
      validate(evidence)
    end
  end

  def test_rejects_record_or_receipt_outside_control_custody_and_record_tested_in_control
    mutations = [
      ["record", "repository_identity", PACKAGE],
      ["receipt", "repository_identity", PACKAGE],
      ["record", "tested_repository", CONTROL]
    ]

    mutations.each do |artifact, field, replacement|
      evidence = deep_copy(valid_evidence)
      evidence.fetch("custody").fetch(artifact)[field] = replacement

      assert_invalid("#{artifact} #{field}") do
        validate(evidence)
      end
    end
  end

  def test_rejects_any_cross_repository_ancestry_without_querying_it
    %w[control substitute].each do |kind|
      evidence = deep_copy(valid_evidence)
      evidence.fetch("ancestry").first["repository_identity"] =
        kind == "control" ? CONTROL : "github.com/faustbrian/go-substitute"
      calls = []

      assert_invalid do
        Validator.validate!(
          evidence: evidence,
          expected_repositories: repository_bindings,
          ancestor: ->(**arguments) { calls << arguments; true }
        )
      end
      assert_empty calls
    end
  end

  private

  def validate(evidence)
    Validator.validate!(
      evidence: evidence,
      expected_repositories: repository_bindings,
      ancestor: ->(**) { true }
    )
  end

  def valid_evidence
    {
      "schema_version" => 3,
      "repositories" => repository_bindings,
      "input_manifest" => {
        "environment_inputs" => [],
        "repositories" => repository_snapshots,
        "schema_version" => 2
      },
      "custody" => {
        "record" => custody(RECORD_REVISION),
        "receipt" => custody(RECEIPT_REVISION)
      },
      "ancestry" => [
        ancestry("tested", PACKAGE_REVISION, PACKAGE_TREE, "gate", GATE_REVISION, GATE_TREE),
        ancestry("gate", GATE_REVISION, GATE_TREE, "revalidation", REVALIDATION_REVISION, REVALIDATION_TREE)
      ]
    }
  end

  def repository_bindings
    [
      {
        "role" => "control",
        "repository_identity" => CONTROL,
        "object_format" => "git-sha1",
        "revision" => CONTROL_REVISION,
        "tree" => "git-sha1:#{"3" * 40}",
        "worktree_root" => "/coord/.worktrees/go-library-tools-identity-platform",
        "module_relative_path" => ".ai/identity-platform",
        "module_workdir" => "/coord/.worktrees/go-library-tools-identity-platform/.ai/identity-platform"
      },
      package_binding
    ]
  end

  def package_binding
    {
      "role" => "package",
      "repository_identity" => PACKAGE,
      "object_format" => "git-sha1",
      "revision" => PACKAGE_REVISION,
      "tree" => PACKAGE_TREE,
      "worktree_root" => "/coord/.worktrees/go-identity-identity-platform",
      "module_relative_path" => ".",
      "module_workdir" => "/coord/.worktrees/go-identity-identity-platform"
    }
  end

  def repository_snapshots
    repository_bindings.map do |binding|
      {
        "inputs" => [{
          "content_identity" => binding.fetch("tree"),
          "kind" => "tree",
          "mode" => "040000",
          "owner" => binding.fetch("role"),
          "path" => binding.fetch("role") == "control" ? ".ai/identity-platform" : ".",
          "reason" => "external evidence v3 input"
        }],
        "object_format" => binding.fetch("object_format"),
        "repository_identity" => binding.fetch("repository_identity"),
        "revision" => binding.fetch("revision"),
        "tree" => binding.fetch("tree")
      }
    end
  end

  def custody(revision)
    {
      "repository_identity" => CONTROL,
      "revision" => revision,
      "tested_repository" => PACKAGE,
      "tested_revision" => PACKAGE_REVISION
    }
  end

  def ancestry(ancestor_role, ancestor_revision, ancestor_tree, descendant_role, descendant_revision, descendant_tree)
    {
      "repository_identity" => PACKAGE,
      "ancestor_role" => ancestor_role,
      "ancestor_revision" => ancestor_revision,
      "ancestor_tree" => ancestor_tree,
      "descendant_role" => descendant_role,
      "descendant_revision" => descendant_revision,
      "descendant_tree" => descendant_tree
    }
  end

  def assert_invalid(message = nil, &block)
    assert_raises(Validator::InvalidEvidence, message, &block)
  end

  def deep_copy(value)
    Marshal.load(Marshal.dump(value))
  end
end
