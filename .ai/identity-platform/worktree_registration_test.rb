# frozen_string_literal: true

require "fileutils"
require "minitest/autorun"
require "open3"
require "tmpdir"

if File.file?(File.join(__dir__, "worktree_registration.rb"))
  require_relative "worktree_registration"
else
  module IdentityPlatformWorktreeRegistration
    def self.registered_paths(*) = []
  end
end

class IdentityPlatformWorktreeRegistrationTest < Minitest::Test
  def setup
    @root = Dir.mktmpdir("identity-worktree-registration-")
    @repositories = %w[control package].map { |name| create_repository(name) }
    @worktrees = @repositories.map.with_index do |repository, index|
      path = File.join(@root, "worktree-#{index}")
      git(repository, "worktree", "add", "--quiet", "-b", "feature/identity-platform", path)
      path
    end
  end

  def teardown
    FileUtils.remove_entry(@root) if @root && File.exist?(@root)
  end

  def test_collects_repository_qualified_worktrees_from_every_program_repository
    registered = IdentityPlatformWorktreeRegistration.registered_paths(@worktrees)

    @worktrees.each { |worktree| assert_includes registered, File.realpath(worktree) }
  end

  private

  def create_repository(name)
    path = File.join(@root, name)
    FileUtils.mkdir_p(path)
    git(path, "init", "--quiet", "--initial-branch=main")
    File.write(File.join(path, "README.md"), "#{name}\n")
    git(path, "add", "README.md")
    git(path, "commit", "--quiet", "-m", "initial")
    path
  end

  def git(repository, *arguments)
    environment = {
      "GIT_AUTHOR_EMAIL" => "identity-platform@example.invalid",
      "GIT_AUTHOR_NAME" => "Identity Platform Test",
      "GIT_COMMITTER_EMAIL" => "identity-platform@example.invalid",
      "GIT_COMMITTER_NAME" => "Identity Platform Test"
    }
    output, error, status = Open3.capture3(environment, "git", "-C", repository, *arguments)
    raise "git #{arguments.join(' ')} failed: #{error}" unless status.success?

    output
  end
end
