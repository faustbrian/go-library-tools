# frozen_string_literal: true

require "minitest/autorun"
if File.file?(File.join(__dir__, "preflight_repository_bindings.rb"))
  require_relative "preflight_repository_bindings"
else
  module IdentityPlatformPreflightRepositoryBindings
    class InvalidBindings < StandardError; end

    def self.validate!(*) = nil
  end
end

class IdentityPlatformPreflightRepositoryBindingsTest < Minitest::Test
  Bindings = IdentityPlatformPreflightRepositoryBindings

  CONTROL = "github.com/faustbrian/go-library-tools"
  AUTHENTICATION = "github.com/faustbrian/go-authentication"
  IDENTITY = "github.com/faustbrian/go-identity"

  EXPECTED_WORKTREES = {
    CONTROL => "/coord/.worktrees/go-library-tools-identity-platform",
    AUTHENTICATION => "/coord/.worktrees/go-authentication-identity-platform",
    IDENTITY => "/coord/.worktrees/go-identity-identity-platform"
  }.freeze
  EXPECTED_FOUNDATIONS = {
    CONTROL => {"foundation" => "existing", "base_revision" => "a" * 40},
    AUTHENTICATION => {"foundation" => "existing", "base_revision" => "d" * 40},
    IDENTITY => {"foundation" => "local-bootstrap", "base_revision" => "1" * 40}
  }.freeze

  CONTROL_GATES = %w[
    inventory-validation pinned-upstream-validation program-final-input-acceptance
    structural-validation
  ].freeze
  PACKAGE_GATES = %w[
    affected-release-gates complete-repository-gate final-complete-diff-review
  ].freeze
  EXPECTED_GATE_PAIRS = (
    CONTROL_GATES.map { |gate_id| [CONTROL, gate_id] } +
    [AUTHENTICATION, IDENTITY].flat_map do |repository_identity|
      PACKAGE_GATES.map { |gate_id| [repository_identity, gate_id] }
    end
  ).freeze

  def test_accepts_exact_repository_qualified_integration_identities_and_gate_pairs
    calls = []

    Bindings.validate!(
      integration_identities: valid_identities,
      final_gate_bindings: valid_gate_bindings,
      expected_worktrees: EXPECTED_WORKTREES,
      expected_foundations: EXPECTED_FOUNDATIONS,
      expected_gate_pairs: EXPECTED_GATE_PAIRS,
      current_revisions: current_revisions,
      resolve_tree: method(:resolve_tree),
      current_snapshot_valid: ->(**) { true },
      ancestor: lambda { |repository_identity:, base_revision:, head_revision:|
        calls << [repository_identity, base_revision, head_revision]
        true
      }
    )

    assert_equal [CONTROL, AUTHENTICATION, IDENTITY].sort_by(&:b), calls.map(&:first).uniq.sort_by(&:b)
  end

  def test_accepts_final_gate_snapshot_descending_from_preflight_baseline
    bindings = deep_copy(valid_gate_bindings)
    bindings.select { |row| row.fetch("repository_identity") == AUTHENTICATION }.each do |row|
      row["tested_revision"] = "9" * 40
      row["tested_tree"] = "git-sha1:#{"8" * 40}"
    end

    Bindings.validate!(
      **validation_arguments,
      final_gate_bindings: bindings,
      require_final_gates: false,
      ancestor: ->(**) { true },
      resolve_tree: lambda { |repository_identity:, revision:|
        revision == "9" * 40 ? "git-sha1:#{"8" * 40}" : resolve_tree(repository_identity: repository_identity, revision: revision)
      }
    )
  end

  def test_rejects_final_gate_tree_not_resolved_from_tested_revision
    bindings = deep_copy(valid_gate_bindings)
    bindings.select { |row| row.fetch("repository_identity") == AUTHENTICATION }.each do |row|
      row["tested_tree"] = "git-sha1:#{"8" * 40}"
    end

    assert_invalid do
      Bindings.validate!(
        **validation_arguments,
        final_gate_bindings: bindings,
        resolve_tree: method(:resolve_tree)
      )
    end
  end

  def test_allows_empty_final_gate_bindings_before_terminal_phase
    Bindings.validate!(
      **validation_arguments,
      final_gate_bindings: [],
      require_final_gates: false
    )

    assert_invalid do
      Bindings.validate!(
        **validation_arguments,
        final_gate_bindings: [],
        require_final_gates: true
      )
    end
  end

  def test_accepts_current_heads_descending_from_preflight_baselines_and_rejects_divergence
    current_revisions = valid_identities.to_h do |row|
      [row.fetch("repository_identity"), row.fetch("baseline_revision")]
    end
    current_revisions[CONTROL] = "7" * 40

    Bindings.validate!(
      **validation_arguments,
      current_revisions: current_revisions,
      resolve_tree: method(:resolve_tree),
      ancestor: ->(**) { true }
    )

    divergent = current_revisions.merge(CONTROL => "8" * 40)
    assert_invalid do
      Bindings.validate!(
        **validation_arguments,
        current_revisions: divergent,
        ancestor: lambda { |base_revision:, head_revision:, **|
          !(base_revision == valid_identities.first.fetch("baseline_revision") && head_revision == "8" * 40)
        }
      )
    end
  end

  def test_rejects_terminal_gate_snapshots_stale_against_current_revisions
    current = current_revisions.merge(AUTHENTICATION => "9" * 40)

    assert_invalid do
      Bindings.validate!(
        **validation_arguments,
        current_revisions: current,
        current_snapshot_valid: lambda { |repository_identity:, tested_revision:, current_revision:|
          repository_identity != AUTHENTICATION || tested_revision == current_revision
        },
        ancestor: ->(**) { true }
      )
    end
  end

  def test_rejects_scalar_legacy_execution_identity_rows
    assert_invalid do
      Bindings.validate!(
        **validation_arguments,
        integration_identities: {"Integration worktree" => EXPECTED_WORKTREES.fetch(CONTROL)}
      )
    end
  end

  def test_rejects_missing_repository_identity_fields
    %w[object_format base_revision base_tree baseline_revision baseline_tree foundation branch worktree recorded_at].each do |field|
      identities = deep_copy(valid_identities)
      identities.fetch(1).delete(field)

      assert_invalid(field) do
        Bindings.validate!(**validation_arguments, integration_identities: identities)
      end
    end
  end

  def test_rejects_a_repository_bound_to_the_wrong_worktree
    identities = deep_copy(valid_identities)
    identities.fetch(1)["worktree"] = EXPECTED_WORKTREES.fetch(IDENTITY)

    assert_invalid do
      Bindings.validate!(**validation_arguments, integration_identities: identities)
    end
  end

  def test_rejects_unknown_repository_foundation_classification
    identities = deep_copy(valid_identities)
    identities.fetch(1)["foundation"] = "assumed-bootstrap"

    assert_invalid do
      Bindings.validate!(**validation_arguments, integration_identities: identities)
    end
  end

  def test_rejects_repository_foundation_mislabeled_as_existing
    identities = deep_copy(valid_identities)
    identities.find { |row| row.fetch("repository_identity") == IDENTITY }["foundation"] = "existing"

    assert_invalid do
      Bindings.validate!(**validation_arguments, integration_identities: identities)
    end
  end

  def test_rejects_repository_base_revision_substitution
    identities = deep_copy(valid_identities)
    identities.find { |row| row.fetch("repository_identity") == AUTHENTICATION }["base_revision"] = "9" * 40

    assert_invalid do
      Bindings.validate!(**validation_arguments, integration_identities: identities)
    end
  end

  def test_rejects_expected_foundation_omission_and_substitution
    omitted = deep_copy(EXPECTED_FOUNDATIONS).tap { |rows| rows.delete(IDENTITY) }
    substituted = deep_copy(EXPECTED_FOUNDATIONS)
    substituted["github.com/faustbrian/go-substitute"] = substituted.delete(IDENTITY)

    [omitted, substituted].each do |expected_foundations|
      assert_invalid do
        Bindings.validate!(**validation_arguments, expected_foundations: expected_foundations)
      end
    end
  end

  def test_rejects_repository_omission_duplication_and_substitution
    omitted = deep_copy(valid_identities).reject { |row| row.fetch("repository_identity") == IDENTITY }
    duplicated = deep_copy(valid_identities) << deep_copy(valid_identities.fetch(1))
    substituted = deep_copy(valid_identities)
    substituted.fetch(1)["repository_identity"] = "github.com/faustbrian/go-substitute"

    [omitted, duplicated, substituted].each do |identities|
      assert_invalid do
        Bindings.validate!(**validation_arguments, integration_identities: identities)
      end
    end
  end

  def test_rejects_final_gate_pair_omission_duplication_and_repository_substitution
    omitted = deep_copy(valid_gate_bindings).tap(&:pop)
    duplicated = deep_copy(valid_gate_bindings) << deep_copy(valid_gate_bindings.first)
    substituted = deep_copy(valid_gate_bindings)
    substituted.fetch(2)["repository_identity"] = IDENTITY

    [omitted, duplicated, substituted].each do |bindings|
      assert_invalid do
        Bindings.validate!(**validation_arguments, final_gate_bindings: bindings)
      end
    end
  end

  def test_rejects_legacy_scalar_final_gate_rows
    assert_invalid do
      Bindings.validate!(**validation_arguments, final_gate_bindings: EXPECTED_GATE_PAIRS.map(&:last))
    end
  end

  def test_rejects_missing_or_extra_final_gate_row_fields
    %w[
      repository_identity gate_id worktree resolved_argv evidence_binding receipt_binding
      producer_capture verifier_capture tested_revision tested_tree repository_input_root
      program_input_root recorded_at
    ].each do |field|
      bindings = deep_copy(valid_gate_bindings)
      bindings.fetch(1).delete(field)

      assert_invalid(field) do
        Bindings.validate!(**validation_arguments, final_gate_bindings: bindings)
      end
    end

    bindings = deep_copy(valid_gate_bindings)
    bindings.fetch(1)["legacy_execution_identity"] = "structural-validation"

    assert_invalid("extra field") do
      Bindings.validate!(**validation_arguments, final_gate_bindings: bindings)
    end
  end

  def test_rejects_final_gate_repository_snapshot_identity_drift
    mutations = {
      "worktree" => EXPECTED_WORKTREES.fetch(IDENTITY),
      "tested_revision" => "9" * 40,
      "tested_tree" => "git-sha1:#{"8" * 40}",
      "repository_input_root" => "sha256:#{"7" * 64}",
      "program_input_root" => "sha256:#{"6" * 64}"
    }

    mutations.each do |field, replacement|
      bindings = deep_copy(valid_gate_bindings)
      bindings.find { |row| row.fetch("repository_identity") == AUTHENTICATION }[field] = replacement

      assert_invalid(field) do
        Bindings.validate!(**validation_arguments, final_gate_bindings: bindings)
      end
    end
  end

  def test_rejects_cross_repository_ancestry_assumptions
    calls = []
    ancestor = lambda do |repository_identity:, base_revision:, head_revision:|
      calls << [repository_identity, base_revision, head_revision]
      revisions = repository_revisions.fetch(repository_identity)
      assert_includes revisions, base_revision, "ancestry base crossed a repository identity"
      assert_includes revisions, head_revision, "ancestry head crossed a repository identity"
      repository_identity == CONTROL
    end

    assert_invalid do
      Bindings.validate!(
        integration_identities: valid_identities,
        final_gate_bindings: valid_gate_bindings,
        expected_worktrees: EXPECTED_WORKTREES,
        expected_foundations: EXPECTED_FOUNDATIONS,
        expected_gate_pairs: EXPECTED_GATE_PAIRS,
        current_revisions: current_revisions,
        resolve_tree: method(:resolve_tree),
        current_snapshot_valid: ->(**) { true },
        ancestor: ancestor
      )
    end

    assert_equal [CONTROL, AUTHENTICATION, IDENTITY].sort_by(&:b), calls.map(&:first).uniq.sort_by(&:b)
  end

  private

  def validation_arguments
    {
      integration_identities: valid_identities,
      final_gate_bindings: valid_gate_bindings,
      expected_worktrees: EXPECTED_WORKTREES,
      expected_foundations: EXPECTED_FOUNDATIONS,
      expected_gate_pairs: EXPECTED_GATE_PAIRS,
      current_revisions: current_revisions,
      resolve_tree: method(:resolve_tree),
      current_snapshot_valid: ->(**) { true },
      ancestor: ->(**) { true }
    }
  end

  def current_revisions
    valid_identities.to_h { |row| [row.fetch("repository_identity"), row.fetch("baseline_revision")] }
  end

  def resolve_tree(repository_identity:, revision:)
    row = valid_identities.find { |candidate| candidate.fetch("repository_identity") == repository_identity }
    return unless row
    return row.fetch("base_tree") if revision == row.fetch("base_revision")
    return unless revision == row.fetch("baseline_revision")

    row.fetch("baseline_tree")
  end

  def valid_identities
    [
      identity(CONTROL, "a", "b", "c"),
      identity(AUTHENTICATION, "d", "e", "f"),
      identity(IDENTITY, "1", "2", "3")
    ]
  end

  def identity(repository_identity, base_digit, tree_digit, head_digit)
    {
      "repository_identity" => repository_identity,
      "object_format" => "git-sha1",
      "base_revision" => base_digit * 40,
      "base_tree" => "git-sha1:#{tree_digit * 40}",
      "baseline_revision" => head_digit * 40,
      "baseline_tree" => "git-sha1:#{head_digit * 40}",
      "foundation" => EXPECTED_FOUNDATIONS.fetch(repository_identity).fetch("foundation"),
      "branch" => "feature/identity-platform",
      "worktree" => EXPECTED_WORKTREES.fetch(repository_identity),
      "recorded_at" => "2026-08-30T10:15:30Z"
    }
  end

  def valid_gate_bindings
    EXPECTED_GATE_PAIRS.map.with_index do |(repository_identity, gate_id), index|
      identity_row = valid_identities.find do |row|
        row.fetch("repository_identity") == repository_identity
      end
      digest_digit = (index + 1).to_s(16)

      {
        "repository_identity" => repository_identity,
        "gate_id" => gate_id,
        "worktree" => EXPECTED_WORKTREES.fetch(repository_identity),
        "resolved_argv" => ["/usr/bin/ruby", "/coord/#{gate_id}.rb"],
        "evidence_binding" => ".ai/identity-platform/evidence/final-gates/#{gate_id}.json@#{"c" * 40}@sha256:#{digest_digit * 64}",
        "receipt_binding" => ".ai/identity-platform/evidence/final-gates/executions/#{gate_id}.json@#{"c" * 40}@sha256:#{digest_digit * 64}",
        "producer_capture" => "producer-#{gate_id}@#{"c" * 40}@sha256:#{digest_digit * 64}",
        "verifier_capture" => "verifier-#{gate_id}@#{"c" * 40}@sha256:#{digest_digit * 64}",
        "tested_revision" => identity_row.fetch("baseline_revision"),
        "tested_tree" => identity_row.fetch("baseline_tree"),
        "repository_input_root" => "sha256:#{repository_root_digit(repository_identity) * 64}",
        "program_input_root" => "sha256:#{"a" * 64}",
        "recorded_at" => "2026-08-30T10:15:30Z"
      }
    end
  end

  def repository_root_digit(repository_identity)
    {CONTROL => "b", AUTHENTICATION => "d", IDENTITY => "e"}.fetch(repository_identity)
  end

  def repository_revisions
    valid_identities.to_h do |row|
      [row.fetch("repository_identity"), [row.fetch("base_revision"), row.fetch("baseline_revision")]]
    end
  end

  def assert_invalid(message = nil, &block)
    assert_raises(Bindings::InvalidBindings, message, &block)
  end

  def deep_copy(value)
    Marshal.load(Marshal.dump(value))
  end
end
