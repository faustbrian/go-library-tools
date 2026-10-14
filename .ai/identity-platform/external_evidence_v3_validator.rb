# frozen_string_literal: true

module IdentityPlatformExternalEvidenceV3Validator
  class InvalidEvidence < StandardError; end

  EVIDENCE_KEYS = %w[schema_version repositories input_manifest custody ancestry].freeze
  REPOSITORY_KEYS = %w[
    role repository_identity object_format revision tree worktree_root
    module_relative_path module_workdir
  ].freeze
  MANIFEST_KEYS = %w[environment_inputs repositories schema_version].freeze
  SNAPSHOT_KEYS = %w[inputs object_format repository_identity revision tree].freeze
  CUSTODY_KEYS = %w[record receipt].freeze
  CUSTODY_RECORD_KEYS = %w[repository_identity revision tested_repository tested_revision].freeze
  ANCESTRY_KEYS = %w[
    repository_identity ancestor_role ancestor_revision ancestor_tree
    descendant_role descendant_revision descendant_tree
  ].freeze

  module_function

  def validate!(evidence:, expected_repositories:, ancestor:)
    invalid!("evidence schema is invalid") unless exact_keys?(evidence, EVIDENCE_KEYS)
    invalid!("external evidence schema version is not 3") unless evidence["schema_version"] == 3
    invalid!("expected repository bindings are invalid") unless expected_repositories.is_a?(Array)

    repositories = evidence["repositories"]
    validate_repositories!(repositories)
    invalid!("repository bindings drifted") unless repositories == expected_repositories

    control = repositories.find { |row| row["role"] == "control" }
    package = repositories.find { |row| row["role"] == "package" }
    invalid!("control and package repositories are not distinct") unless
      control && package && control["repository_identity"] != package["repository_identity"]

    validate_manifest!(evidence["input_manifest"], repositories)
    validate_custody!(evidence["custody"], control, package)
    ancestry_rows = validate_ancestry!(evidence["ancestry"], package)

    ancestry_rows.each do |row|
      valid = ancestor.call(
        repository_identity: row.fetch("repository_identity"),
        ancestor_revision: row.fetch("ancestor_revision"),
        descendant_revision: row.fetch("descendant_revision")
      )
      invalid!("repository-qualified ancestry is invalid") unless valid
    end

    true
  rescue KeyError, TypeError, NoMethodError => e
    raise InvalidEvidence, "external evidence is malformed: #{e.message}"
  end

  def validate_repositories!(repositories)
    invalid!("repository bindings are absent") unless repositories.is_a?(Array) && repositories.length == 2
    invalid!("repository binding schema is invalid") unless
      repositories.all? { |row| exact_keys?(row, REPOSITORY_KEYS) }
    invalid!("repository roles are invalid") unless repositories.map { |row| row["role"] } == %w[control package]

    identities = repositories.map { |row| row["repository_identity"] }
    invalid!("repository identities are invalid") unless
      identities.uniq.length == identities.length && identities.all? { |identity| repository_identity?(identity) }
    repositories.each do |row|
      validate_object_identity!(row)
      invalid!("repository worktree root is invalid") unless absolute_path?(row["worktree_root"])
      invalid!("repository module path is invalid") unless relative_path?(row["module_relative_path"], allow_dot: true)
      invalid!("repository module workdir is invalid") unless absolute_path?(row["module_workdir"])
      expected_workdir = row["module_relative_path"] == "." ? row["worktree_root"] :
        File.join(row["worktree_root"], row["module_relative_path"])
      invalid!("repository module workdir drifted") unless row["module_workdir"] == expected_workdir
    end
  end
  private_class_method :validate_repositories!

  def validate_manifest!(manifest, repositories)
    invalid!("input manifest schema is invalid") unless exact_keys?(manifest, MANIFEST_KEYS)
    invalid!("input manifest schema version is not 2") unless manifest["schema_version"] == 2
    invalid!("input manifest environment inputs are invalid") unless manifest["environment_inputs"].is_a?(Array)
    snapshots = manifest["repositories"]
    invalid!("input manifest repository snapshots are absent") unless snapshots.is_a?(Array)
    invalid!("input manifest repository snapshot schema is invalid") unless
      snapshots.all? { |snapshot| exact_keys?(snapshot, SNAPSHOT_KEYS) && snapshot["inputs"].is_a?(Array) }

    expected = repositories.to_h { |row| [row.fetch("repository_identity"), row] }
    actual_identities = snapshots.map { |snapshot| snapshot["repository_identity"] }
    invalid!("input manifest repository closure drifted") unless
      actual_identities.uniq.length == actual_identities.length && actual_identities.sort == expected.keys.sort
    snapshots.each do |snapshot|
      binding = expected.fetch(snapshot.fetch("repository_identity"))
      %w[object_format revision tree].each do |field|
        invalid!("input manifest repository #{field} drifted") unless snapshot[field] == binding[field]
      end
    end
  end
  private_class_method :validate_manifest!

  def validate_custody!(custody, control, package)
    invalid!("custody schema is invalid") unless exact_keys?(custody, CUSTODY_KEYS)
    CUSTODY_KEYS.each do |kind|
      record = custody[kind]
      invalid!("#{kind} custody schema is invalid") unless exact_keys?(record, CUSTODY_RECORD_KEYS)
      invalid!("#{kind} is outside control repository custody") unless
        record["repository_identity"] == control["repository_identity"]
      invalid!("#{kind} tested repository drifted") unless
        record["tested_repository"] == package["repository_identity"]
      invalid!("#{kind} tested revision drifted") unless record["tested_revision"] == package["revision"]
      invalid!("#{kind} custody revision is invalid") unless revision?(record["revision"], control["object_format"])
    end
  end
  private_class_method :validate_custody!

  def validate_ancestry!(rows, package)
    invalid!("ancestry chain is invalid") unless rows.is_a?(Array) && rows.length == 2
    invalid!("ancestry schema is invalid") unless rows.all? { |row| exact_keys?(row, ANCESTRY_KEYS) }
    invalid!("ancestry crosses repository identities") unless
      rows.all? { |row| row["repository_identity"] == package["repository_identity"] }
    invalid!("ancestry roles drifted") unless
      rows.map { |row| [row["ancestor_role"], row["descendant_role"]] } ==
        [["tested", "gate"], ["gate", "revalidation"]]
    invalid!("tested ancestry revision drifted") unless rows.first["ancestor_revision"] == package["revision"]
    invalid!("tested ancestry tree drifted") unless rows.first["ancestor_tree"] == package["tree"]
    invalid!("ancestry chain is discontinuous") unless
      rows.first["descendant_revision"] == rows.last["ancestor_revision"] &&
        rows.first["descendant_tree"] == rows.last["ancestor_tree"]
    rows.each do |row|
      invalid!("ancestry revision is invalid") unless
        revision?(row["ancestor_revision"], package["object_format"]) &&
          revision?(row["descendant_revision"], package["object_format"])
      invalid!("ancestry tree is invalid") unless
        tree?(row["ancestor_tree"], package["object_format"]) &&
          tree?(row["descendant_tree"], package["object_format"])
    end
    rows
  end
  private_class_method :validate_ancestry!

  def validate_object_identity!(row)
    format = row["object_format"]
    invalid!("repository object format is invalid") unless %w[git-sha1 git-sha256].include?(format)
    invalid!("repository revision is invalid") unless revision?(row["revision"], format)
    length = format == "git-sha256" ? 64 : 40
    invalid!("repository tree is invalid") unless row["tree"].to_s.match?(%r{\A#{format}:[0-9a-f]{#{length}}\z})
  end
  private_class_method :validate_object_identity!

  def revision?(value, format)
    length = format == "git-sha256" ? 64 : 40
    value.is_a?(String) && value.match?(%r{\A[0-9a-f]{#{length}}\z})
  end
  private_class_method :revision?

  def tree?(value, format)
    length = format == "git-sha256" ? 64 : 40
    value.is_a?(String) && value.match?(%r{\A#{format}:[0-9a-f]{#{length}}\z})
  end
  private_class_method :tree?

  def repository_identity?(value)
    value.is_a?(String) && value.match?(%r{\A[a-z0-9.-]+/[A-Za-z0-9._-]+(?:/[A-Za-z0-9._-]+)+\z})
  end
  private_class_method :repository_identity?

  def absolute_path?(value)
    value.is_a?(String) && value.start_with?("/") && !value.include?("\0") && !value.split("/").include?("..")
  end
  private_class_method :absolute_path?

  def relative_path?(value, allow_dot:)
    return true if allow_dot && value == "."
    value.is_a?(String) && value.match?(%r{\A[^/\0]+(?:/[^/\0]+)*\z}) && !value.split("/").include?("..")
  end
  private_class_method :relative_path?

  def exact_keys?(value, keys)
    value.is_a?(Hash) && value.keys == keys
  end
  private_class_method :exact_keys?

  def invalid!(message)
    raise InvalidEvidence, message
  end
  private_class_method :invalid!
end
