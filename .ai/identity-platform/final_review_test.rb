#!/usr/bin/env ruby
# frozen_string_literal: true

require "digest"
require "fileutils"
require "json"
require "minitest/autorun"
require "open3"
require "tmpdir"
require_relative "final_review_gate"
require_relative "final_review_verifier"

class IdentityPlatformFinalReviewTest < Minitest::Test
  Gate = IdentityPlatformFinalReviewGate
  Verifier = IdentityPlatformFinalReviewVerifier

  def setup
    @temporary_root = Dir.mktmpdir("identity-platform-final-review-")
  end

  def teardown
    FileUtils.remove_entry(@temporary_root) if File.exist?(@temporary_root)
  end

  def test_reviews_every_manifest_repository_in_its_own_worktree_and_tags_results
    first = create_repository("go-first", "github.com/example/go-first")
    second = create_repository("go-second", "github.com/example/go-second")
    manifest = manifest_bytes([snapshot(second), snapshot(first)])
    calls = []

    output = Gate.run(manifest, reviewer: lambda { |base:, chdir:, prompt:|
      calls << {base: base, chdir: chdir, prompt: prompt}
      '{"critical":[],"important":[]}'
    })
    document = JSON.parse(output)

    assert_equal [first, second], calls.map { |call| call.fetch(:chdir) }
    assert_equal JSON.parse(manifest).fetch("repositories").map { |row| row.fetch("base_revision") },
      calls.map { |call| call.fetch(:base) }
    assert calls.all? { |call| call.fetch(:prompt).include?("Critical or Important") }
    assert_equal "sha256:#{Digest::SHA256.hexdigest(manifest)}", document.fetch("manifest_sha256")
    assert_equal %w[github.com/example/go-first github.com/example/go-second],
      document.fetch("repositories").map { |row| row.fetch("repository_identity") }
    assert document.fetch("repositories").all? do |row|
      row.fetch("output_sha256") == "sha256:#{Digest::SHA256.hexdigest(JSON.generate(row.fetch('output')))}"
    end
    assert_equal output, JSON.generate(document)
  end

  def test_verifier_accepts_manifest_bound_clean_results
    repository = create_repository("go-first", "github.com/example/go-first")
    manifest = manifest_bytes([snapshot(repository)])
    output = Gate.run(manifest, reviewer: clean_reviewer)

    attestation = JSON.parse(Verifier.verify(output, manifest_bytes: manifest))

    assert_equal "identity-platform.final-review-verifier.v2", attestation.fetch("schema")
    assert_equal 1, attestation.fetch("repository_count")
    assert_equal "sha256:#{Digest::SHA256.hexdigest(output)}", attestation.fetch("review_sha256")
  end

  def test_verifier_rejects_manifest_that_omits_expected_repository_closure
    first = create_repository("go-first", "github.com/example/go-first")
    second = create_repository("go-second", "github.com/example/go-second")
    complete_repositories = [snapshot(first), snapshot(second)].sort_by do |row|
      row.fetch("repository_identity").b
    end
    incomplete_manifest = manifest_bytes([complete_repositories.first])
    incomplete_output = Gate.run(incomplete_manifest, reviewer: clean_reviewer)

    error = assert_raises(RuntimeError) do
      Verifier.verify(
        incomplete_output, manifest_bytes: incomplete_manifest,
        expected_repositories: complete_repositories
      )
    end

    assert_match(/repository closure drifted/, error.message)
  end

  def test_verifier_rejects_omitted_and_wrong_repository_rows
    first = create_repository("go-first", "github.com/example/go-first")
    second = create_repository("go-second", "github.com/example/go-second")
    manifest = manifest_bytes([snapshot(first), snapshot(second)])
    output = JSON.parse(Gate.run(manifest, reviewer: clean_reviewer))

    omitted = deep_copy(output)
    omitted.fetch("repositories").pop
    assert_raises(RuntimeError) { Verifier.verify(JSON.generate(omitted), manifest_bytes: manifest) }

    substituted = deep_copy(output)
    substituted.fetch("repositories").first["repository_identity"] = "github.com/example/go-wrong"
    assert_raises(RuntimeError) { Verifier.verify(JSON.generate(substituted), manifest_bytes: manifest) }
  end

  def test_verifier_rejects_output_digest_mismatch_and_findings
    repository = create_repository("go-first", "github.com/example/go-first")
    manifest = manifest_bytes([snapshot(repository)])
    output = JSON.parse(Gate.run(manifest, reviewer: clean_reviewer))

    mismatched = deep_copy(output)
    mismatched.fetch("repositories").first["output_sha256"] = "sha256:#{'0' * 64}"
    assert_raises(RuntimeError) { Verifier.verify(JSON.generate(mismatched), manifest_bytes: manifest) }

    %w[critical important].each do |severity|
      finding = deep_copy(output)
      review = finding.fetch("repositories").first.fetch("output")
      review.fetch(severity) << {"message" => "fixture"}
      finding.fetch("repositories").first["output_sha256"] =
        "sha256:#{Digest::SHA256.hexdigest(JSON.generate(review))}"
      assert_raises(RuntimeError, severity) do
        Verifier.verify(JSON.generate(finding), manifest_bytes: manifest)
      end
    end
  end

  def test_verifier_rejects_head_and_tree_drift_after_review
    repository = create_repository("go-first", "github.com/example/go-first")
    manifest = manifest_bytes([snapshot(repository)])
    output = Gate.run(manifest, reviewer: clean_reviewer)
    commit_file(repository, "drift.txt", "drift\n")

    error = assert_raises(RuntimeError) do
      Verifier.verify(output, manifest_bytes: manifest)
    end

    assert_match(/head or tree drift/, error.message)
  end

  def test_gate_rejects_noncanonical_manifest_and_repository_identity_substitution
    repository = create_repository("go-first", "github.com/example/go-first")
    document = JSON.parse(manifest_bytes([snapshot(repository)]))

    assert_raises(RuntimeError) do
      Gate.run(JSON.pretty_generate(document), reviewer: clean_reviewer)
    end

    document.fetch("repositories").first["repository_identity"] = "github.com/example/go-wrong"
    canonical = JSON.generate(document)
    assert_raises(RuntimeError) do
      Gate.run(canonical, reviewer: clean_reviewer)
    end
  end

  def test_gate_rechecks_all_repository_heads_after_the_last_review
    first = create_repository("go-first", "github.com/example/go-first")
    second = create_repository("go-second", "github.com/example/go-second")
    manifest = manifest_bytes([snapshot(first), snapshot(second)])
    calls = 0

    assert_raises(RuntimeError) do
      Gate.run(manifest, reviewer: lambda { |base:, chdir:, prompt:|
        calls += 1
        commit_file(first, "late-drift.txt", "drift\n") if calls == 2
        '{"critical":[],"important":[]}'
      })
    end
  end

  def test_gate_rejects_uncommitted_worktree_state_not_bound_by_head_tree
    repository = create_repository("go-first", "github.com/example/go-first")
    manifest = manifest_bytes([snapshot(repository)])
    File.binwrite(File.join(repository, "untracked.txt"), "not in the manifest\n")

    error = assert_raises(RuntimeError) do
      Gate.run(manifest, reviewer: clean_reviewer)
    end

    assert_match(/worktree is dirty/, error.message)
  end

  private

  def clean_reviewer
    lambda { |base:, chdir:, prompt:| '{"critical":[],"important":[]}' }
  end

  def deep_copy(value)
    JSON.parse(JSON.generate(value))
  end

  def create_repository(name, identity)
    repository = File.join(@temporary_root, name)
    FileUtils.mkdir_p(repository)
    git(repository, "init", "--quiet")
    git(repository, "config", "user.name", "Final Review Test")
    git(repository, "config", "user.email", "final-review@example.invalid")
    remote = name == "go-first" ? "git@#{identity.sub('/', ':')}.git" : "https://#{identity}.git"
    git(repository, "remote", "add", "origin", remote)
    commit_file(repository, "base.txt", "base\n")
    commit_file(repository, "head.txt", "head\n")
    File.realpath(repository)
  end

  def commit_file(repository, path, contents)
    full_path = File.join(repository, path)
    FileUtils.mkdir_p(File.dirname(full_path))
    File.binwrite(full_path, contents)
    git(repository, "add", path)
    git(repository, "commit", "--quiet", "-m", "fixture")
    git(repository, "rev-parse", "HEAD")
  end

  def snapshot(repository)
    {
      "base_revision" => git(repository, "rev-parse", "HEAD^"),
      "head_revision" => git(repository, "rev-parse", "HEAD"),
      "head_tree" => "git-sha1:#{git(repository, 'rev-parse', 'HEAD^{tree}')}",
      "object_format" => "git-#{git(repository, 'rev-parse', '--show-object-format')}",
      "repository_identity" => remote_identity(repository),
      "worktree" => repository
    }
  end

  def manifest_bytes(repositories)
    JSON.generate(
      "repositories" => repositories.sort_by { |row| row.fetch("repository_identity").b },
      "schema" => "identity-platform.final-review-manifest.v1",
      "schema_version" => 1
    )
  end

  def remote_identity(repository)
    remote = git(repository, "config", "--get", "remote.origin.url")
    return remote.delete_prefix("https://").delete_suffix(".git") if remote.start_with?("https://")

    remote.sub(/\Agit@([^:]+):/, '\\1/').delete_suffix(".git")
  end

  def git(repository, *arguments)
    output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    raise "git fixture failed: #{error}" unless status.success?

    output.strip
  end
end
