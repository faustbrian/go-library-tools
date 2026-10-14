# frozen_string_literal: true

require "digest"
require "json"

module IdentityPlatformTerminalStateBindings
  class InvalidBinding < StandardError; end

  THRESHOLD_KEYS = %w[
    required_count observed_count history_event_ids repository_snapshots
    program_input_root blockers reconciliations exhausted_frontier
  ].freeze
  SNAPSHOT_KEYS = %w[repository_identity object_format revision tree].freeze
  NO_PUSH_KEYS = %w[repositories].freeze
  NO_PUSH_REPOSITORY_KEYS = %w[
    repository_identity program_base_revision final_input_revision remote_names
    refspecs observed_local_branch observed_local_head
  ].freeze

  module_function

  def execution_identity(repository_snapshots:, program_input_root:)
    payload = {
      "repository_snapshots" => repository_snapshots,
      "program_input_root" => program_input_root
    }
    "sha256:#{Digest::SHA256.hexdigest(JSON.generate(payload))}"
  end

  def validate!(blocked_threshold:, execution_identity:, no_push_scope:,
                expected_repository_snapshots:, expected_blockers:,
                expected_reconciliations:, expected_no_push_scope:, ancestor:)
    invalid!("blocked threshold schema is invalid") unless exact_keys?(blocked_threshold, THRESHOLD_KEYS)
    snapshots = blocked_threshold["repository_snapshots"]
    validate_snapshots!(snapshots)
    invalid!("repository snapshot projection drifted") unless snapshots == expected_repository_snapshots
    invalid!("blocked observation threshold is invalid") unless
      blocked_threshold["required_count"] == 3 &&
        blocked_threshold["observed_count"].is_a?(Integer) &&
        blocked_threshold["history_event_ids"].is_a?(Array) &&
        blocked_threshold["observed_count"] == blocked_threshold["history_event_ids"].length &&
        blocked_threshold["observed_count"] >= blocked_threshold["required_count"]
    invalid!("program input root is invalid") unless digest?(blocked_threshold["program_input_root"])
    invalid!("blocker projection drifted") unless blocked_threshold["blockers"] == expected_blockers
    invalid!("reconciliation projection drifted") unless
      blocked_threshold["reconciliations"] == expected_reconciliations
    invalid!("frontier is not exhausted") unless blocked_threshold["exhausted_frontier"] == true
    expected_identity = self.execution_identity(
      repository_snapshots: snapshots,
      program_input_root: blocked_threshold.fetch("program_input_root")
    )
    invalid!("execution identity does not bind terminal repositories") unless execution_identity == expected_identity

    validate_no_push_scope!(no_push_scope, snapshots)
    invalid!("no-push scope drifted") unless no_push_scope == expected_no_push_scope

    no_push_scope.fetch("repositories").each do |row|
      identity = row.fetch("repository_identity")
      invalid!("repository base does not reach final input") unless ancestor.call(
        repository_identity: identity,
        ancestor_revision: row.fetch("program_base_revision"),
        descendant_revision: row.fetch("final_input_revision")
      )
      invalid!("repository final input does not reach local head") unless ancestor.call(
        repository_identity: identity,
        ancestor_revision: row.fetch("final_input_revision"),
        descendant_revision: row.fetch("observed_local_head")
      )
    end

    true
  rescue KeyError, TypeError, NoMethodError => e
    raise InvalidBinding, "terminal state binding is malformed: #{e.message}"
  end

  def validate_snapshots!(snapshots)
    invalid!("repository snapshots are absent") unless snapshots.is_a?(Array) && snapshots.any?
    invalid!("repository snapshot schema is invalid") unless
      snapshots.all? { |snapshot| exact_keys?(snapshot, SNAPSHOT_KEYS) }
    identities = snapshots.map { |snapshot| snapshot["repository_identity"] }
    invalid!("repository snapshots are duplicated") unless identities.uniq.length == identities.length
    snapshots.each do |snapshot|
      format = snapshot["object_format"]
      invalid!("repository object format is invalid") unless %w[git-sha1 git-sha256].include?(format)
      length = format == "git-sha256" ? 64 : 40
      invalid!("repository revision is invalid") unless snapshot["revision"].to_s.match?(%r{\A[0-9a-f]{#{length}}\z})
      invalid!("repository tree is invalid") unless snapshot["tree"].to_s.match?(%r{\A#{format}:[0-9a-f]{#{length}}\z})
    end
  end
  private_class_method :validate_snapshots!

  def validate_no_push_scope!(scope, snapshots)
    invalid!("no-push scope schema is invalid") unless exact_keys?(scope, NO_PUSH_KEYS)
    repositories = scope["repositories"]
    invalid!("no-push repository scope is invalid") unless repositories.is_a?(Array) &&
      repositories.all? { |row| exact_keys?(row, NO_PUSH_REPOSITORY_KEYS) }
    expected_identities = snapshots.map { |snapshot| snapshot.fetch("repository_identity") }
    invalid!("no-push repository closure or order drifted") unless
      repositories.map { |row| row["repository_identity"] } == expected_identities
    repositories.zip(snapshots).each do |row, snapshot|
      invalid!("no-push final input revision drifted") unless row["final_input_revision"] == snapshot["revision"]
      invalid!("no-push repository revisions are invalid") unless
        %w[program_base_revision final_input_revision observed_local_head].all? do |key|
          row[key].is_a?(String) && row[key].match?(%r{\A[0-9a-f]{#{snapshot["object_format"] == "git-sha256" ? 64 : 40}}\z})
        end
      invalid!("no-push remotes are not canonical") unless string_set?(row["remote_names"])
      invalid!("no-push refspec scope drifted") unless row["refspecs"] == ["refs/*"]
      invalid!("no-push branch is invalid") unless row["observed_local_branch"].to_s.match?(%r{\A(?:feature|bugfix|hotfix|release|chore|refactor)/[A-Za-z0-9._/-]+\z})
    end
  end
  private_class_method :validate_no_push_scope!

  def string_set?(value)
    value.is_a?(Array) && value.any? && value.all? { |entry| entry.is_a?(String) && !entry.empty? } &&
      value == value.sort && value.uniq.length == value.length
  end
  private_class_method :string_set?

  def digest?(value)
    value.is_a?(String) && value.match?(%r{\Asha256:[0-9a-f]{64}\z})
  end
  private_class_method :digest?

  def exact_keys?(value, keys)
    value.is_a?(Hash) && value.keys == keys
  end
  private_class_method :exact_keys?

  def invalid!(message)
    raise InvalidBinding, message
  end
  private_class_method :invalid!
end
