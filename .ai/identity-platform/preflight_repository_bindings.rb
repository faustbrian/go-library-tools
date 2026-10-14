# frozen_string_literal: true

require "time"

module IdentityPlatformPreflightRepositoryBindings
  class InvalidBindings < StandardError; end

  INTEGRATION_IDENTITY_KEYS = %w[
    repository_identity object_format base_revision base_tree baseline_revision baseline_tree foundation
    branch worktree recorded_at
  ].freeze
  FINAL_GATE_KEYS = %w[
    repository_identity gate_id worktree resolved_argv evidence_binding receipt_binding
    producer_capture verifier_capture tested_revision tested_tree repository_input_root
    program_input_root recorded_at
  ].freeze
  INTEGRATION_BRANCH = "feature/identity-platform"

  module_function

  def validate!(integration_identities:, final_gate_bindings:, expected_worktrees:,
                expected_foundations:, expected_gate_pairs:, current_revisions:, ancestor:, resolve_tree:,
                current_snapshot_valid:, require_final_gates: true)
    worktrees = validate_expected_worktrees(expected_worktrees)
    foundations = validate_expected_foundations(expected_foundations, worktrees)
    identities = validate_integration_identities(
      integration_identities, worktrees, foundations, ancestor, resolve_tree
    )
    identity_rows = integration_identities.to_h { |row| [row.fetch("repository_identity"), row] }
    validate_current_revisions(current_revisions, identity_rows, worktrees, ancestor)
    validate_final_gate_bindings(
      final_gate_bindings, expected_gate_pairs, identities,
      identity_rows,
      worktrees, current_revisions, ancestor, resolve_tree, current_snapshot_valid, require_final_gates
    )
    true
  end

  def validate_expected_worktrees(expected_worktrees)
    unless expected_worktrees.is_a?(Hash) && !expected_worktrees.empty? &&
      expected_worktrees.keys.all? { |identity| valid_repository_identity?(identity) } &&
      expected_worktrees.values.all? { |worktree| canonical_absolute_path?(worktree) } &&
      expected_worktrees.values.uniq.length == expected_worktrees.length
      raise InvalidBindings, "expected repository worktree map is invalid"
    end

    expected_worktrees
  end
  private_class_method :validate_expected_worktrees

  def validate_expected_foundations(source, expected_worktrees)
    unless source.is_a?(Hash) && source.keys == expected_worktrees.keys
      raise InvalidBindings, "expected repository foundation closure drifted"
    end

    source.each do |identity, row|
      unless row.is_a?(Hash) && row.keys == %w[foundation base_revision] &&
        %w[existing local-bootstrap].include?(row.fetch("foundation")) &&
        row.fetch("base_revision").is_a?(String)
        raise InvalidBindings, "#{identity}: expected repository foundation is invalid"
      end
    end

    source
  rescue KeyError, NoMethodError, TypeError
    raise InvalidBindings, "expected repository foundations are invalid"
  end
  private_class_method :validate_expected_foundations

  def validate_integration_identities(source, expected_worktrees, expected_foundations, ancestor, resolve_tree)
    unless source.is_a?(Array)
      raise InvalidBindings, "repository integration identities must be repository-qualified rows"
    end
    unless ancestor.respond_to?(:call) && resolve_tree.respond_to?(:call)
      raise InvalidBindings, "repository ancestry or tree resolver is absent"
    end

    rows = source.map.with_index do |row, index|
      unless row.is_a?(Hash) && row.keys == INTEGRATION_IDENTITY_KEYS
        raise InvalidBindings, "repository integration identity #{index} schema is invalid"
      end

      validate_integration_identity!(row)
      identity = row.fetch("repository_identity")
      expected_foundation = expected_foundations.fetch(identity)
      unless row.fetch("foundation") == expected_foundation.fetch("foundation") &&
        row.fetch("base_revision") == expected_foundation.fetch("base_revision")
        raise InvalidBindings, "#{identity}: repository foundation or base revision drifted"
      end
      row
    end
    identities = rows.map { |row| row.fetch("repository_identity") }
    expected_identities = expected_worktrees.keys
    unless identities.length == identities.uniq.length && identities == expected_identities
      raise InvalidBindings, "repository integration identity closure drifted"
    end

    ancestry_failures = rows.filter_map do |row|
      identity = row.fetch("repository_identity")
      unless row.fetch("worktree") == expected_worktrees.fetch(identity)
        raise InvalidBindings, "#{identity}: integration worktree drifted"
      end
      unless resolve_tree.call(repository_identity: identity, revision: row.fetch("base_revision")) == row.fetch("base_tree") &&
        resolve_tree.call(repository_identity: identity, revision: row.fetch("baseline_revision")) == row.fetch("baseline_tree")
        raise InvalidBindings, "#{identity}: integration repository tree drifted"
      end
      valid = ancestor.call(
        repository_identity: identity,
        base_revision: row.fetch("base_revision"),
        head_revision: row.fetch("baseline_revision")
      ) == true
      identity unless valid
    end
    unless ancestry_failures.empty?
      raise InvalidBindings,
        "#{ancestry_failures.min_by(&:b)}: preflight integration baseline does not descend from its repository base"
    end

    identities.to_h { |identity| [identity, true] }
  rescue KeyError
    raise InvalidBindings, "repository integration identity foundation closure drifted"
  end
  private_class_method :validate_integration_identities

  def validate_current_revisions(source, identity_rows, expected_worktrees, ancestor)
    unless source.is_a?(Hash) && source.keys == expected_worktrees.keys
      raise InvalidBindings, "current repository revision closure drifted"
    end

    source.each do |identity, revision|
      identity_row = identity_rows.fetch(identity)
      object_format = identity_row.fetch("object_format")
      expected_length = object_format == "git-sha1" ? 40 : object_format == "git-sha256" ? 64 : nil
      unless expected_length && revision.is_a?(String) && revision.match?(%r{\A[0-9a-f]{#{expected_length}}\z}) &&
        ancestor.call(
          repository_identity: identity,
          base_revision: identity_row.fetch("baseline_revision"),
          head_revision: revision
        ) == true
        raise InvalidBindings, "#{identity}: current integration revision predates or diverges from its preflight baseline"
      end
    end
  rescue KeyError, NoMethodError, TypeError
    raise InvalidBindings, "current repository revisions are invalid"
  end
  private_class_method :validate_current_revisions

  def validate_integration_identity!(row)
    identity = row.fetch("repository_identity")
    object_format = row.fetch("object_format")
    base_revision = row.fetch("base_revision")
    base_tree = row.fetch("base_tree")
    baseline_revision = row.fetch("baseline_revision")
    baseline_tree = row.fetch("baseline_tree")
    foundation = row.fetch("foundation")
    expected_length = object_format == "git-sha1" ? 40 : object_format == "git-sha256" ? 64 : nil

    raise InvalidBindings, "repository identity is invalid" unless valid_repository_identity?(identity)
    unless expected_length && base_revision.match?(%r{\A[0-9a-f]{#{expected_length}}\z}) &&
      baseline_revision.match?(%r{\A[0-9a-f]{#{expected_length}}\z}) &&
      base_tree.match?(%r{\A#{Regexp.escape(object_format)}:[0-9a-f]{#{expected_length}}\z}) &&
      baseline_tree.match?(%r{\A#{Regexp.escape(object_format)}:[0-9a-f]{#{expected_length}}\z})
      raise InvalidBindings, "#{identity}: Git object identity is invalid"
    end
    raise InvalidBindings, "#{identity}: integration branch drifted" unless row.fetch("branch") == INTEGRATION_BRANCH
    raise InvalidBindings, "#{identity}: repository foundation classification is invalid" unless
      %w[existing local-bootstrap].include?(foundation)
    raise InvalidBindings, "#{identity}: integration worktree is invalid" unless canonical_absolute_path?(row.fetch("worktree"))
    raise InvalidBindings, "#{identity}: recorded timestamp is invalid" unless rfc3339?(row.fetch("recorded_at"))
  rescue NoMethodError, TypeError
    raise InvalidBindings, "#{identity || 'repository'}: integration identity value is invalid"
  end
  private_class_method :validate_integration_identity!

  def validate_final_gate_bindings(source, expected_gate_pairs, identities, identity_rows, expected_worktrees,
                                   current_revisions, ancestor, resolve_tree, current_snapshot_valid,
                                   require_final_gates)
    unless source.is_a?(Array) && expected_gate_pairs.is_a?(Array)
      raise InvalidBindings, "final gate repository bindings are invalid"
    end
    expected = expected_gate_pairs.map do |pair|
      unless pair.is_a?(Array) && pair.length == 2 && valid_repository_identity?(pair[0]) &&
        pair[1].is_a?(String) && pair[1].match?(%r{\A[a-z0-9]+(?:-[a-z0-9]+)*\z}) && identities.key?(pair[0])
        raise InvalidBindings, "expected final gate pair is invalid"
      end
      pair
    end
    unless expected.length == expected.uniq.length
      raise InvalidBindings, "expected final gate pairs contain duplicates"
    end
    return if source.empty? && !require_final_gates

    actual = source.map.with_index do |row, index|
      unless row.is_a?(Hash) && row.keys == FINAL_GATE_KEYS &&
        identities.key?(row.fetch("repository_identity")) &&
        row.fetch("gate_id").is_a?(String) && row.fetch("gate_id").match?(%r{\A[a-z0-9]+(?:-[a-z0-9]+)*\z})
        raise InvalidBindings, "final gate repository binding #{index} is invalid"
      end
      validate_final_gate_row!(
        row, identity_rows.fetch(row.fetch("repository_identity")), expected_worktrees, ancestor, resolve_tree
      )
      [row.fetch("repository_identity"), row.fetch("gate_id")]
    end
    unless actual.length == actual.uniq.length && actual.sort_by { |pair| pair.map(&:b) } ==
      expected.sort_by { |pair| pair.map(&:b) }
      raise InvalidBindings, "final gate repository pair closure drifted"
    end
    program_roots = source.map { |row| row.fetch("program_input_root") }.uniq
    raise InvalidBindings, "final gate program input root drifted" unless program_roots.length == 1

    source.group_by { |row| row.fetch("repository_identity") }.each do |identity, rows|
      snapshots = rows.map { |row| row.values_at("tested_revision", "tested_tree") }.uniq
      raise InvalidBindings, "#{identity}: final gate repository snapshot drifted" unless snapshots.length == 1
      if require_final_gates && (!current_snapshot_valid.respond_to?(:call) ||
        current_snapshot_valid.call(
          repository_identity: identity,
          tested_revision: snapshots.first.first,
          current_revision: current_revisions.fetch(identity)
        ) != true)
        raise InvalidBindings, "#{identity}: final gate repository snapshot is stale against current integration state"
      end
      repository_roots = rows.map { |row| row.fetch("repository_input_root") }.uniq
      raise InvalidBindings, "#{identity}: final gate repository input root drifted" unless repository_roots.length == 1
    end
  end
  private_class_method :validate_final_gate_bindings

  def validate_final_gate_row!(row, identity, expected_worktrees, ancestor, resolve_tree)
    repository_identity = row.fetch("repository_identity")
    object_format = identity.fetch("object_format")
    expected_length = object_format == "git-sha1" ? 40 : object_format == "git-sha256" ? 64 : nil
    tested_revision = row.fetch("tested_revision")
    tested_tree = row.fetch("tested_tree")
    unless row.fetch("worktree") == expected_worktrees.fetch(repository_identity) && expected_length &&
      tested_revision.match?(%r{\A[0-9a-f]{#{expected_length}}\z}) &&
      tested_tree.match?(%r{\A#{Regexp.escape(object_format)}:[0-9a-f]{#{expected_length}}\z}) &&
      resolve_tree.call(repository_identity: repository_identity, revision: tested_revision) == tested_tree &&
      ancestor.call(
        repository_identity: repository_identity,
        base_revision: identity.fetch("baseline_revision"),
        head_revision: tested_revision
      ) == true
      raise InvalidBindings, "#{repository_identity}: final gate repository snapshot drifted"
    end
    argv = row.fetch("resolved_argv")
    unless argv.is_a?(Array) && !argv.empty? && argv.all? { |value| value.is_a?(String) && !value.empty? } &&
      canonical_absolute_path?(argv.first)
      raise InvalidBindings, "#{repository_identity}: final gate argv is invalid"
    end
    %w[evidence_binding receipt_binding producer_capture verifier_capture].each do |field|
      value = row.fetch(field)
      unless value.is_a?(String) && !value.empty? && !value.match?(/[\r\n]/)
        raise InvalidBindings, "#{repository_identity}: final gate #{field} is invalid"
      end
    end
    %w[repository_input_root program_input_root].each do |field|
      unless row.fetch(field).is_a?(String) && row.fetch(field).match?(%r{\Asha256:[0-9a-f]{64}\z})
        raise InvalidBindings, "#{repository_identity}: final gate #{field} is invalid"
      end
    end
    unless rfc3339?(row.fetch("recorded_at"))
      raise InvalidBindings, "#{repository_identity}: final gate timestamp is invalid"
    end
  rescue KeyError, NoMethodError, TypeError
    raise InvalidBindings, "#{repository_identity || 'repository'}: final gate repository binding is invalid"
  end
  private_class_method :validate_final_gate_row!

  def valid_repository_identity?(value)
    value.is_a?(String) && value.match?(%r{\A[a-z0-9.-]+/[A-Za-z0-9._-]+(?:/[A-Za-z0-9._-]+)+\z}) &&
      !value.split("/").include?("..")
  end
  private_class_method :valid_repository_identity?

  def canonical_absolute_path?(value)
    return false unless value.is_a?(String) && value.start_with?("/") && !value.include?("\0")

    File.expand_path(value) == value
  end
  private_class_method :canonical_absolute_path?

  def rfc3339?(value)
    value.is_a?(String) && Time.iso8601(value).iso8601 == value
  rescue ArgumentError
    false
  end
  private_class_method :rfc3339?
end
