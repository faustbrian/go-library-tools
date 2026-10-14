#!/usr/bin/env ruby
# frozen_string_literal: true

require "fileutils"
require "minitest/autorun"
require "open3"
require "tmpdir"
require_relative "coordinator_custody"

class IdentityPlatformCoordinatorCustodyTest < Minitest::Test
  Custody = IdentityPlatformCoordinatorCustody

  def setup
    @root = Dir.mktmpdir("identity-platform-coordinator-custody-")
    @control = create_repository("go-library-tools", "github.com/faustbrian/go-library-tools")
    @package = create_repository("go-identity", "github.com/faustbrian/go-identity")
  end

  def teardown
    FileUtils.remove_entry(@root) if File.exist?(@root)
  end

  def test_resolves_control_and_package_commits_only_in_their_own_repositories
    control_revision = commit(@control, ".ai/identity-platform/EXECUTION_LEDGER.md", "ledger\n")
    package_revision = commit(@package, "identity.go", "package identity\n")

    control = Custody.resolve_commit(
      repository_root: @control,
      repository_identity: "github.com/faustbrian/go-library-tools",
      revision: control_revision,
      input_roots: [input_root(".ai/identity-platform/EXECUTION_LEDGER.md", "coordinator")]
    )
    package = Custody.resolve_commit(
      repository_root: @package,
      repository_identity: "github.com/faustbrian/go-identity",
      revision: package_revision,
      input_roots: [input_root("identity.go", "worker")]
    )

    assert_equal control_revision, control.fetch("revision")
    assert_equal package_revision, package.fetch("revision")
    assert_match(/\Agit-sha1:[0-9a-f]{40}\z/, control.fetch("tree"))
    assert_match(/\Agit-sha1:[0-9a-f]{40}\z/, package.fetch("tree"))
  end

  def test_rejects_a_control_commit_presented_as_a_package_commit
    control_revision = commit(@control, ".ai/identity-platform/EXECUTION_LEDGER.md", "ledger\n")
    commit(@package, "identity.go", "package identity\n")

    error = assert_raises(Custody::InvalidCustody) do
      Custody.resolve_commit(
        repository_root: @package,
        repository_identity: "github.com/faustbrian/go-identity",
        revision: control_revision,
        input_roots: [input_root("identity.go", "worker")]
      )
    end

    assert_match(/not an exact commit/, error.message)
  end

  def test_rejects_wrong_or_missing_repository_identity
    revision = commit(@package, "identity.go", "package identity\n")

    ["github.com/faustbrian/go-sso", "", nil].each do |identity|
      error = assert_raises(Custody::InvalidCustody) do
        Custody.resolve_commit(
          repository_root: @package,
          repository_identity: identity,
          revision: revision,
          input_roots: [input_root("identity.go", "worker")]
        )
      end
      assert_match(/repository identity/, error.message)
    end
  end

  def test_rejects_cross_repository_ancestry_queries
    error = assert_raises(Custody::InvalidCustody) do
      Custody.ancestor?(
        ancestor: {repository_identity: "github.com/faustbrian/go-library-tools", revision: "a" * 40},
        descendant: {repository_identity: "github.com/faustbrian/go-identity", revision: "b" * 40},
        repository_root: @package
      )
    end

    assert_match(/cross-repository ancestry is forbidden/, error.message)
  end

  def test_validates_control_evidence_and_package_execution_without_cross_repository_ancestry
    control_revision = commit(@control, ".ai/identity-platform/contract.json", "contract\n")
    package_revision = commit(@package, "identity.go", "package identity\n")
    evidence_commit = commit(@control, ".ai/identity-platform/evidence/result.json", "{}\n")
    control_binding = binding(@control, "github.com/faustbrian/go-library-tools", control_revision)
    package_binding = binding(@package, "github.com/faustbrian/go-identity", package_revision)

    errors = Custody.evidence_link_errors(
      control_binding: control_binding,
      package_binding: package_binding,
      control_repository_root: @control,
      package_repository_root: @package,
      evidence_commit: evidence_commit
    )

    assert_empty errors
    refute git_object_exists?(@control, package_revision)
  end

  def test_rejects_repository_binding_tree_substitution
    control_revision = commit(@control, ".ai/identity-platform/contract.json", "contract\n")
    package_revision = commit(@package, "identity.go", "package identity\n")
    evidence_commit = commit(@control, ".ai/identity-platform/evidence/result.json", "{}\n")
    control_binding = binding(@control, "github.com/faustbrian/go-library-tools", control_revision)
    package_binding = binding(@package, "github.com/faustbrian/go-identity", package_revision)
    package_binding["tree"] = control_binding.fetch("tree")

    errors = Custody.evidence_link_errors(
      control_binding: control_binding,
      package_binding: package_binding,
      control_repository_root: @control,
      package_repository_root: @package,
      evidence_commit: evidence_commit
    )

    assert_includes errors, "package repository tree drifted"
  end

  private

  def create_repository(name, identity)
    path = File.join(@root, name)
    FileUtils.mkdir_p(path)
    system("git", "-C", path, "init", "--quiet", exception: true)
    system("git", "-C", path, "remote", "add", "origin", "git@#{identity.sub('/', ':')}.git", exception: true)
    path
  end

  def commit(repository, path, bytes)
    absolute = File.join(repository, path)
    FileUtils.mkdir_p(File.dirname(absolute))
    File.binwrite(absolute, bytes)
    system("git", "-C", repository, "add", "--", path, exception: true)
    system(
      "git", "-C", repository, "-c", "user.name=Identity Platform", "-c",
      "user.email=identity@example.invalid", "commit", "--quiet", "-m", "fixture", exception: true
    )
    output, error, status = Open3.capture3("git", "-C", repository, "rev-parse", "HEAD")
    raise error unless status.success?

    output.strip
  end

  def input_root(path, owner)
    {path: path, owner: owner, reason: "custody fixture"}
  end

  def binding(repository, identity, revision)
    object_format = git(repository, "rev-parse", "--show-object-format")
    {
      "repository_identity" => identity,
      "object_format" => "git-#{object_format}",
      "revision" => revision,
      "tree" => "git-#{object_format}:#{git(repository, 'rev-parse', "#{revision}^{tree}")}"
    }
  end

  def git_object_exists?(repository, revision)
    system("git", "-C", repository, "cat-file", "-e", "#{revision}^{commit}", out: File::NULL, err: File::NULL)
  end

  def git(repository, *arguments)
    output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    raise error unless status.success?

    output.strip
  end
end
