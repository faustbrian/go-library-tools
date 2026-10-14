# frozen_string_literal: true

require "minitest/autorun"
require "fileutils"
require "open3"
require "tmpdir"
require_relative "repository_identity_validator"

class IdentityPlatformRepositoryIdentityValidatorTest < Minitest::Test
  def setup
    @directory = Dir.mktmpdir("identity-platform-repository-identity-")
    @control = create_repository("control", "control.txt" => "control\n")
    @package = create_repository("package", "package.txt" => "package\n")
    @control_revision = git(@control, "rev-parse", "HEAD")
    @package_baseline = git(@package, "rev-parse", "HEAD")
    File.binwrite(File.join(@package, "package.txt"), "worker\n")
    git(@package, "add", "package.txt")
    git(@package, "commit", "-m", "worker")
    @package_worker = git(@package, "rev-parse", "HEAD")
    git(@package, "checkout", "-q", "-b", "integration", @package_baseline)
    git(@package, "merge", "--no-ff", "-m", "integrate worker", @package_worker)
    @package_integration = git(@package, "rev-parse", "HEAD")
  end

  def teardown
    FileUtils.remove_entry(@directory)
  end

  def test_reads_package_objects_only_from_the_mapped_package_repository
    repositories = IdentityPlatformRepositoryIdentityValidator::Repositories.new(
      control_repository: @control, package_repository: @package
    )

    assert_equal "worker\n", repositories.package_blob(@package_worker, "package.txt")
    assert_nil repositories.package_blob(@control_revision, "control.txt")
    assert_equal "control\n", repositories.control_blob(@control_revision, "control.txt")
    assert repositories.package_ancestor?(@package_baseline, @package_worker)
    refute repositories.package_commit?(@control_revision)
  end

  def test_package_result_binding_never_uses_control_commit_ancestry
    repositories = IdentityPlatformRepositoryIdentityValidator::Repositories.new(
      control_repository: @control, package_repository: @package
    )

    errors = repositories.package_result_errors(
      baseline: @package_baseline, worker: @package_worker,
      integration: @package_integration, expected_merge: true
    )

    assert_empty errors
    refute repositories.package_ancestor?(@control_revision, @package_worker)
  end

  def test_assignment_identity_includes_both_repository_identities
    expected = {
      unit: "identity", generation: "1",
      control_repository: "github.com/example/control",
      package_repository: "github.com/example/package",
      task: "worker:identity"
    }
    substituted = expected.merge(package_repository: "github.com/example/other-package")

    assert IdentityPlatformRepositoryIdentityValidator.same_assignment?(expected, expected.dup)
    refute IdentityPlatformRepositoryIdentityValidator.same_assignment?(expected, substituted)
  end

  def test_authorization_identity_matches_assignment_without_worker_task_column
    authorization = {
      unit: "identity", generation: "1",
      control_repository: "github.com/example/control",
      package_repository: "github.com/example/package"
    }
    assignment = authorization.merge(task: "worker:identity")

    assert IdentityPlatformRepositoryIdentityValidator.same_authorization?(authorization, assignment)
    refute IdentityPlatformRepositoryIdentityValidator.same_authorization?(
      authorization, assignment.merge(package_repository: "github.com/example/other-package")
    )
  end

  private

  def create_repository(name, files)
    path = File.join(@directory, name)
    Dir.mkdir(path)
    git(path, "init", "-q")
    git(path, "config", "user.email", "identity-platform@example.test")
    git(path, "config", "user.name", "Identity Platform Test")
    files.each { |relative, bytes| File.binwrite(File.join(path, relative), bytes) }
    git(path, "add", *files.keys)
    git(path, "commit", "-m", "initial")
    path
  end

  def git(repository, *arguments)
    output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    raise error unless status.success?

    output.strip
  end
end
