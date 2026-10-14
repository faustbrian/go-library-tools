# frozen_string_literal: true

require "digest"
require "open3"
require "time"

module IdentityPlatformLocalGateBindingValidator
  module_function

  def errors(binding:, ledger_entry:, control_repository_root:, package_repository_root:)
    errors = []
    expected_path = ".ai/identity-platform/evidence/gates/#{ledger_entry[:unit].tr("/", "-")}.json"
    errors << "binding identity drifted" unless binding.values_at(:unit, :generation, :gate_revision) ==
      [ledger_entry[:unit], ledger_entry[:generation], ledger_entry[:gate_revision]]
    errors << "binding path drifted" unless binding[:path] == expected_path

    control_commit = binding[:commit]
    control_commit_valid = commit_exists?(control_repository_root, control_commit)
    errors << "binding commit is unreachable" unless control_commit_valid &&
      ancestor?(control_repository_root, control_commit, "HEAD")

    gate_revision = ledger_entry[:gate_revision]
    package_commit_valid = commit_exists?(package_repository_root, gate_revision)
    errors << "gate execution revision is unreachable" unless package_commit_valid
    if package_commit_valid && ledger_entry[:gate_tree]
      format = git_output(package_repository_root, "rev-parse", "--show-object-format")
      tree = git_output(package_repository_root, "rev-parse", "#{gate_revision}^{tree}")
      errors << "gate execution tree drifted" unless "git-#{format}:#{tree}" == ledger_entry[:gate_tree]
    end

    committed = control_commit_valid && git_output(
      control_repository_root, "show", "#{control_commit}:#{binding[:path]}", binary: true
    )
    absolute = File.expand_path(binding[:path].to_s, control_repository_root)
    exact = committed && absolute.start_with?(control_repository_root + File::SEPARATOR) &&
      File.file?(absolute) && File.binread(absolute) == committed &&
      binding[:digest] == "sha256:#{Digest::SHA256.hexdigest(committed)}"
    errors << "binding does not prove exact local gate bytes" unless exact
    errors << "binding timestamp is invalid" unless rfc3339?(binding[:bound_at])
    errors
  rescue KeyError, TypeError, NoMethodError
    ["local gate binding is malformed"]
  end

  def commit_exists?(repository, revision)
    return false unless revision.to_s.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/)

    _output, _error, status = Open3.capture3(
      "git", "-C", repository, "cat-file", "-e", "#{revision}^{commit}"
    )
    status.success?
  end
  private_class_method :commit_exists?

  def ancestor?(repository, ancestor, descendant)
    _output, _error, status = Open3.capture3(
      "git", "-C", repository, "merge-base", "--is-ancestor", ancestor, descendant
    )
    status.success?
  end
  private_class_method :ancestor?

  def git_output(repository, *arguments, binary: false)
    output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    return output if status.success? && error.empty? && binary
    return output.strip if status.success? && error.empty?

    nil
  end
  private_class_method :git_output

  def rfc3339?(value)
    value.is_a?(String) && Time.iso8601(value).utc.iso8601 == value
  rescue ArgumentError
    false
  end
  private_class_method :rfc3339?
end
