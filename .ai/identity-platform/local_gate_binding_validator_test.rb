# frozen_string_literal: true

require "digest"
require "fileutils"
require "minitest/autorun"
require "open3"
require "tmpdir"
require_relative "local_gate_binding_validator"

class IdentityPlatformLocalGateBindingValidatorTest < Minitest::Test
  Validator = IdentityPlatformLocalGateBindingValidator

  def test_accepts_control_owned_record_for_disjoint_package_gate_revision
    with_repositories do |control, package|
      package_revision, package_tree = commit(package, "go.mod", "module example.test/package\n")
      record_path = ".ai/identity-platform/evidence/gates/identity.json"
      record = "{\"outcome\":\"pass\"}\n"
      control_revision, = commit(control, record_path, record)
      binding = binding_for(record_path, control_revision, package_revision, record)
      ledger = ledger_for(package_revision, package_tree)

      assert_empty Validator.errors(
        binding: binding, ledger_entry: ledger,
        control_repository_root: control, package_repository_root: package
      )
    end
  end

  def test_rejects_package_revision_substitution_and_tree_drift
    with_repositories do |control, package|
      package_revision, package_tree = commit(package, "go.mod", "module example.test/package\n")
      record_path = ".ai/identity-platform/evidence/gates/identity.json"
      record = "{\"outcome\":\"pass\"}\n"
      control_revision, = commit(control, record_path, record)
      binding = binding_for(record_path, control_revision, package_revision, record)

      substituted = ledger_for(control_revision, package_tree)
      refute_empty Validator.errors(
        binding: binding, ledger_entry: substituted,
        control_repository_root: control, package_repository_root: package
      )

      wrong_tree = ledger_for(package_revision, "git-sha1:#{"0" * 40}")
      refute_empty Validator.errors(
        binding: binding, ledger_entry: wrong_tree,
        control_repository_root: control, package_repository_root: package
      )
    end
  end

  private

  def with_repositories
    Dir.mktmpdir("identity-local-gate-binding-") do |directory|
      control = File.join(directory, "control")
      package = File.join(directory, "package")
      [control, package].each do |repository|
        FileUtils.mkdir_p(repository)
        git(repository, "init", "-q")
        git(repository, "config", "user.email", "fixture@example.test")
        git(repository, "config", "user.name", "Fixture")
      end
      yield control, package
    end
  end

  def commit(repository, path, bytes)
    absolute = File.join(repository, path)
    FileUtils.mkdir_p(File.dirname(absolute))
    File.binwrite(absolute, bytes)
    git(repository, "add", "--", path)
    git(repository, "commit", "-qm", "fixture")
    revision = git(repository, "rev-parse", "HEAD")
    tree = git(repository, "rev-parse", "#{revision}^{tree}")
    [revision, "git-sha1:#{tree}"]
  end

  def binding_for(path, control_revision, package_revision, bytes)
    {
      unit: "identity", generation: "1", gate_revision: package_revision,
      path: path, commit: control_revision,
      digest: "sha256:#{Digest::SHA256.hexdigest(bytes)}",
      bound_at: "2026-08-30T00:00:00Z"
    }
  end

  def ledger_for(revision, tree)
    {unit: "identity", generation: "1", gate_revision: revision, gate_tree: tree}
  end

  def git(repository, *arguments)
    output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    raise "git failed: #{error}" unless status.success?

    output.strip
  end
end
