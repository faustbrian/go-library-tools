# frozen_string_literal: true

require "fileutils"
require "minitest/autorun"
require "open3"
require "set"
require "tmpdir"

load File.join(__dir__, "validate.rb")

class IdentityPlatformInitialPreflightPopulationIntegrationTest < Minitest::Test
  PREFLIGHT = ".ai/identity-platform/PREFLIGHT_EVIDENCE.md"
  SEMANTIC_SOURCE = ".ai/identity-platform/PROGRAM.md"

  def setup
    @repository = Dir.mktmpdir("identity-initial-preflight-")
    git("init", "--quiet", "--initial-branch=main")
    write(PREFLIGHT, preflight_source("pending"))
    write(SEMANTIC_SOURCE, "baseline\n")
    write("CHANGELOG.md", "baseline\n")
    git("add", PREFLIGHT, SEMANTIC_SOURCE, "CHANGELOG.md")
    git("commit", "--quiet", "-m", "baseline")
    @base = git("rev-parse", "HEAD").strip
    @original_repository_root = REPOSITORY_ROOT
    Object.send(:remove_const, :REPOSITORY_ROOT)
    Object.const_set(:REPOSITORY_ROOT, @repository)
  end

  def teardown
    Object.send(:remove_const, :REPOSITORY_ROOT)
    Object.const_set(:REPOSITORY_ROOT, @original_repository_root)
    FileUtils.remove_entry(@repository) if @repository && File.exist?(@repository)
  end

  def test_accepts_pure_initial_population_commit
    commit = population_commit({})

    assert_empty custody_errors(commit)
  end

  def test_rejects_initial_population_commit_mixed_with_non_control_path
    commit = population_commit("CHANGELOG.md" => "changed\n")

    errors = custody_errors(commit)

    refute_empty errors
    assert_includes errors.first, "CHANGELOG.md"
  end

  def test_rejects_initial_population_commit_mixed_with_authorized_semantic_source
    commit = population_commit(SEMANTIC_SOURCE => "authorized change\n")

    errors = custody_errors(commit, semantic_application_sources: {commit => Set[SEMANTIC_SOURCE]})

    refute_empty errors
    assert_includes errors.first, SEMANTIC_SOURCE
  end

  private

  def population_commit(changes)
    write(PREFLIGHT, preflight_source(@base))
    changes.each { |path, content| write(path, content) }
    git("add", PREFLIGHT, *changes.keys)
    git("commit", "--quiet", "-m", "populate preflight")
    git("rev-parse", "HEAD").strip
  end

  def custody_errors(commit, semantic_application_sources: {})
    global_control_commit_class_errors(
      recorded_base: @base,
      integration_commits: Set.new,
      semantic_application_sources: semantic_application_sources,
      artifact_commit_paths: {}
    )
  end

  def preflight_source(base)
    <<~MARKDOWN
      # Identity Platform Preflight Evidence

      ## Execution identity

      | Field | Value |
      | --- | --- |
      | Recorded control execution baseline | #{base} |
      | Preflight input revision before the record commit | #{base} |

      ## Tool and environment lanes

      | Requirement/profile | Classification |
      | --- | --- |
      | tools | available |

      ## Conflict-recovery baselines

      | Recovery epoch | Status |
      | --- | --- |
    MARKDOWN
  end

  def write(path, content)
    absolute = File.join(@repository, path)
    FileUtils.mkdir_p(File.dirname(absolute))
    File.binwrite(absolute, content)
  end

  def git(*arguments)
    environment = {
      "GIT_AUTHOR_EMAIL" => "identity-platform@example.invalid",
      "GIT_AUTHOR_NAME" => "Identity Platform Test",
      "GIT_COMMITTER_EMAIL" => "identity-platform@example.invalid",
      "GIT_COMMITTER_NAME" => "Identity Platform Test"
    }
    output, error, status = Open3.capture3(environment, "git", "-C", @repository, *arguments)
    raise "git #{arguments.join(' ')} failed: #{error}" unless status.success?

    output
  end
end
