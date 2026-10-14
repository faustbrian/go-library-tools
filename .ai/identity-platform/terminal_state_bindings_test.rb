# frozen_string_literal: true

require "digest"
require "json"
require "minitest/autorun"

if File.file?(File.join(__dir__, "terminal_state_bindings.rb"))
  require_relative "terminal_state_bindings"
else
  module IdentityPlatformTerminalStateBindings
    class InvalidBinding < StandardError; end

    def self.execution_identity(repository_snapshots:, program_input_root:)
      "sha256:#{"0" * 64}"
    end

    def self.validate!(**_arguments)
      true
    end
  end
end

class IdentityPlatformTerminalStateBindingsTest < Minitest::Test
  Bindings = IdentityPlatformTerminalStateBindings

  CONTROL = "github.com/faustbrian/go-library-tools"
  PACKAGE = "github.com/faustbrian/go-identity"
  PROGRAM_ROOT = "sha256:#{"a" * 64}"

  def test_accepts_exact_repository_qualified_terminal_state
    assert Bindings.validate!(**validation_arguments)
  end

  def test_execution_identity_hashes_canonical_repository_snapshots_and_program_root
    canonical_payload = {
      "repository_snapshots" => repository_snapshots,
      "program_input_root" => PROGRAM_ROOT
    }
    expected = "sha256:#{Digest::SHA256.hexdigest(JSON.generate(canonical_payload))}"

    assert_equal expected, execution_identity

    {
      "repository identity" => ->(rows) { rows.fetch(1)["repository_identity"] = "github.com/faustbrian/go-sso" },
      "object format" => ->(rows) { rows.fetch(1)["object_format"] = "git-sha256" },
      "revision" => ->(rows) { rows.fetch(1)["revision"] = "9" * 40 },
      "tree" => ->(rows) { rows.fetch(1)["tree"] = "git-sha1:#{"8" * 40}" }
    }.each do |label, mutation|
      rows = deep_copy(repository_snapshots)
      mutation.call(rows)
      refute_equal execution_identity,
        Bindings.execution_identity(repository_snapshots: rows, program_input_root: PROGRAM_ROOT), label
    end

    refute_equal execution_identity,
      Bindings.execution_identity(
        repository_snapshots: repository_snapshots,
        program_input_root: "sha256:#{"b" * 64}"
      )
  end

  def test_rejects_legacy_scalar_final_input_revision_and_input_root
    threshold = deep_copy(blocked_threshold)
    threshold["final_input_revision"] = threshold.fetch("repository_snapshots").fetch(0).fetch("revision")
    threshold["input_root"] = threshold.delete("program_input_root")

    assert_invalid do
      Bindings.validate!(**validation_arguments.merge(blocked_threshold: threshold))
    end

    scope = deep_copy(no_push_scope)
    scope["final_input_revision"] = scope.fetch("repositories").fetch(0).fetch("final_input_revision")
    scope["input_root"] = PROGRAM_ROOT

    assert_invalid do
      Bindings.validate!(**validation_arguments.merge(no_push_scope: scope))
    end
  end

  def test_rejects_repository_omission_duplication_substitution_and_reordering
    threshold_mutations = [
      ->(rows) { rows.pop },
      ->(rows) { rows << deep_copy(rows.first) },
      ->(rows) { rows.fetch(1)["repository_identity"] = "github.com/faustbrian/go-sso" },
      ->(rows) { rows.reverse! }
    ]
    threshold_mutations.each do |mutation|
      threshold = deep_copy(blocked_threshold)
      mutation.call(threshold.fetch("repository_snapshots"))
      assert_invalid do
        Bindings.validate!(**validation_arguments.merge(blocked_threshold: threshold))
      end
    end

    scope_mutations = [
      ->(rows) { rows.pop },
      ->(rows) { rows << deep_copy(rows.first) },
      ->(rows) { rows.fetch(1)["repository_identity"] = "github.com/faustbrian/go-sso" },
      ->(rows) { rows.reverse! }
    ]
    scope_mutations.each do |mutation|
      scope = deep_copy(no_push_scope)
      mutation.call(scope.fetch("repositories"))
      assert_invalid do
        Bindings.validate!(**validation_arguments.merge(no_push_scope: scope))
      end
    end
  end

  def test_rejects_snapshot_scope_and_terminal_projection_mismatch
    threshold = deep_copy(blocked_threshold)
    threshold.fetch("repository_snapshots").fetch(1)["tree"] = "git-sha1:#{"7" * 40}"
    assert_invalid do
      Bindings.validate!(**validation_arguments.merge(blocked_threshold: threshold))
    end

    {
      "blockers" => [],
      "reconciliations" => [],
      "exhausted_frontier" => false,
      "program_input_root" => "sha256:#{"6" * 64}"
    }.each do |field, value|
      threshold = deep_copy(blocked_threshold)
      threshold[field] = value
      assert_invalid(field) do
        Bindings.validate!(**validation_arguments.merge(blocked_threshold: threshold))
      end
    end

    scope = deep_copy(no_push_scope)
    scope.fetch("repositories").fetch(1)["remote_names"] = %w[origin]
    assert_invalid do
      Bindings.validate!(**validation_arguments.merge(no_push_scope: scope))
    end

    scope = deep_copy(no_push_scope)
    scope.fetch("repositories").fetch(1)["remote_names"] = []
    assert_invalid do
      Bindings.validate!(**validation_arguments.merge(no_push_scope: scope))
    end
  end

  def test_rejects_missing_extra_or_misordered_closed_fields
    snapshot = deep_copy(blocked_threshold)
    snapshot.fetch("repository_snapshots").fetch(1).delete("tree")
    assert_invalid { Bindings.validate!(**validation_arguments.merge(blocked_threshold: snapshot)) }

    snapshot = deep_copy(blocked_threshold)
    snapshot.fetch("repository_snapshots").fetch(1)["legacy"] = "value"
    assert_invalid { Bindings.validate!(**validation_arguments.merge(blocked_threshold: snapshot)) }

    scope = deep_copy(no_push_scope)
    row = scope.fetch("repositories").fetch(1)
    row.replace(row.to_a.reverse.to_h)
    assert_invalid { Bindings.validate!(**validation_arguments.merge(no_push_scope: scope)) }

    scope = deep_copy(no_push_scope)
    scope.fetch("repositories").fetch(1).delete("program_base_revision")
    assert_invalid { Bindings.validate!(**validation_arguments.merge(no_push_scope: scope)) }
  end

  def test_ancestry_is_checked_only_within_each_repository
    revisions = repository_snapshots.to_h do |snapshot|
      scope = no_push_scope.fetch("repositories").find do |row|
        row.fetch("repository_identity") == snapshot.fetch("repository_identity")
      end
      [snapshot.fetch("repository_identity"),
        [scope.fetch("program_base_revision"), scope.fetch("final_input_revision"), scope.fetch("observed_local_head")]]
    end
    calls = []
    ancestor = lambda do |repository_identity:, ancestor_revision:, descendant_revision:|
      calls << [repository_identity, ancestor_revision, descendant_revision]
      owned = revisions.fetch(repository_identity)
      assert_includes owned, ancestor_revision
      assert_includes owned, descendant_revision
      true
    end

    Bindings.validate!(**validation_arguments.merge(ancestor: ancestor))

    assert_equal [CONTROL, CONTROL, PACKAGE, PACKAGE], calls.map(&:first)

    assert_invalid do
      Bindings.validate!(**validation_arguments.merge(
        ancestor: ->(repository_identity:, ancestor_revision:, descendant_revision:) {
          repository_identity == CONTROL && ancestor_revision != descendant_revision
        }
      ))
    end
  end

  private

  def validation_arguments
    {
      blocked_threshold: blocked_threshold,
      execution_identity: execution_identity,
      no_push_scope: no_push_scope,
      expected_repository_snapshots: repository_snapshots,
      expected_blockers: blockers,
      expected_reconciliations: reconciliations,
      expected_no_push_scope: no_push_scope,
      ancestor: ->(**) { true }
    }
  end

  def repository_snapshots
    [
      repository_snapshot(CONTROL, "a", "b"),
      repository_snapshot(PACKAGE, "c", "d")
    ]
  end

  def repository_snapshot(identity, revision_digit, tree_digit)
    {
      "repository_identity" => identity,
      "object_format" => "git-sha1",
      "revision" => revision_digit * 40,
      "tree" => "git-sha1:#{tree_digit * 40}"
    }
  end

  def blockers
    [{
      "repository_identity" => PACKAGE,
      "id" => "provider-credential",
      "category" => "credential",
      "unit" => "identity/oauth",
      "evidence" => "provider credential is absent",
      "required_user_action" => "provide the provider credential"
    }]
  end

  def reconciliations
    [
      reconciliation(CONTROL, "1", [], ["integration-worktree"]),
      reconciliation(PACKAGE, "2", ["identity/oauth:g1"], [])
    ]
  end

  def reconciliation(identity, digit, assignments, resources)
    {
      "repository_identity" => identity,
      "assignment_ids" => assignments,
      "resource_ids" => resources,
      "assignments_root" => "sha256:#{digit * 64}",
      "resources_root" => "sha256:#{digit.succ * 64}"
    }
  end

  def blocked_threshold
    {
      "required_count" => 3,
      "observed_count" => 3,
      "history_event_ids" => %w[blocked-1 blocked-2 blocked-3],
      "repository_snapshots" => repository_snapshots,
      "program_input_root" => PROGRAM_ROOT,
      "blockers" => blockers,
      "reconciliations" => reconciliations,
      "exhausted_frontier" => true
    }
  end

  def no_push_scope
    {
      "repositories" => [
        no_push_repository(CONTROL, "1", "a", "e", ["origin"]),
        no_push_repository(PACKAGE, "2", "c", "f", ["mirror", "origin"])
      ]
    }
  end

  def no_push_repository(identity, base_digit, final_digit, head_digit, remotes)
    {
      "repository_identity" => identity,
      "program_base_revision" => base_digit * 40,
      "final_input_revision" => final_digit * 40,
      "remote_names" => remotes,
      "refspecs" => ["refs/*"],
      "observed_local_branch" => "feature/identity-platform",
      "observed_local_head" => head_digit * 40
    }
  end

  def execution_identity
    Bindings.execution_identity(
      repository_snapshots: repository_snapshots,
      program_input_root: PROGRAM_ROOT
    )
  end

  def assert_invalid(message = nil, &block)
    assert_raises(Bindings::InvalidBinding, message, &block)
  end

  def deep_copy(value)
    Marshal.load(Marshal.dump(value))
  end
end
