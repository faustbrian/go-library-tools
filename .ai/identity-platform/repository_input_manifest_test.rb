#!/usr/bin/env ruby
# frozen_string_literal: true

require "fileutils"
require "json"
require "minitest/autorun"
require "open3"
require "tmpdir"
require_relative "repository_input_manifest"

class IdentityPlatformRepositoryInputManifestTest < Minitest::Test
  Manifest = IdentityPlatformRepositoryInputManifest

  def setup
    @temporary_root = Dir.mktmpdir("identity-platform-repository-manifest-")
  end

  def teardown
    FileUtils.remove_entry(@temporary_root) if File.exist?(@temporary_root)
  end

  def test_builds_canonical_v2_snapshots_for_distinct_repositories_with_the_same_commit
    first = create_repository("first", "github.com/example/first")
    second = create_repository("second", "github.com/example/second")
    first_revision = commit_files(first, "lib/value.txt" => "same\n")
    second_revision = commit_files(second, "lib/value.txt" => "same\n")
    assert_equal first_revision, second_revision

    document = Manifest.build(
      repositories: [
        repository_request(second, "github.com/example/second", second_revision),
        repository_request(first, "github.com/example/first", first_revision)
      ],
      environment_inputs: [
        environment_input("environment:toolchain", "b"),
        environment_input("environment:architecture", "a")
      ]
    )

    assert_equal 2, document.fetch("schema_version")
    assert_equal %w[environment_inputs repositories schema_version], document.keys
    assert_equal %w[environment:architecture environment:toolchain],
      document.fetch("environment_inputs").map { |entry| entry.fetch("environment_id") }
    assert_equal %w[github.com/example/first github.com/example/second],
      document.fetch("repositories").map { |entry| entry.fetch("repository_identity") }

    first_snapshot = document.fetch("repositories").first
    assert_equal "git-sha1", first_snapshot.fetch("object_format")
    assert_equal first_revision, first_snapshot.fetch("revision")
    assert_equal "git-sha1:#{git(first, "rev-parse", "#{first_revision}^{tree}")}", first_snapshot.fetch("tree")
    assert_equal [
      {
        "content_identity" => "git-sha1:#{git(first, "rev-parse", "#{first_revision}:lib/value.txt")}",
        "kind" => "blob",
        "mode" => "100644",
        "owner" => "library",
        "path" => "lib/value.txt",
        "reason" => "behavioral source"
      }
    ], first_snapshot.fetch("inputs")
    refute_includes JSON.generate(document), first
    refute_includes JSON.generate(document), second
    assert_match(/\Asha256:[0-9a-f]{64}\z/, Manifest.fingerprint(document))
  end

  def test_repository_and_input_order_do_not_change_document_or_fingerprint
    first = create_repository("first", "github.com/example/first")
    second = create_repository("second", "github.com/example/second")
    first_revision = commit_files(first, "a.txt" => "a\n", "z.txt" => "z\n")
    second_revision = commit_files(second, "value.txt" => "value\n")
    forward = [
      repository_request(first, "github.com/example/first", first_revision, input_roots: [
        {path: "z.txt", owner: "library", reason: "z input"},
        {path: "a.txt", owner: "library", reason: "a input"}
      ]),
      repository_request(second, "github.com/example/second", second_revision, input_roots: [
        {path: "value.txt", owner: "library", reason: "value input"}
      ])
    ]

    first_document = Manifest.build(repositories: forward, environment_inputs: [
      environment_input("environment:z", "z"), environment_input("environment:a", "a")
    ])
    second_document = Manifest.build(repositories: forward.reverse.map { |request|
      request.merge(input_roots: request.fetch(:input_roots).reverse)
    }, environment_inputs: [
      environment_input("environment:a", "a"), environment_input("environment:z", "z")
    ])

    assert_equal first_document, second_document
    assert_equal Manifest.canonical_json(first_document), Manifest.canonical_json(second_document)
    assert_equal Manifest.fingerprint(first_document), Manifest.fingerprint(second_document)
  end

  def test_one_repository_change_alters_only_its_snapshot_and_composite_fingerprint
    first = create_repository("first", "github.com/example/first")
    second = create_repository("second", "github.com/example/second")
    first_revision = commit_files(first, "lib/value.txt" => "before\n")
    second_revision = commit_files(second, "lib/value.txt" => "stable\n")
    before = Manifest.build(repositories: [
      repository_request(first, "github.com/example/first", first_revision),
      repository_request(second, "github.com/example/second", second_revision)
    ], environment_inputs: [])

    changed_revision = commit_files(first, "lib/value.txt" => "after\n")
    after = Manifest.build(repositories: [
      repository_request(first, "github.com/example/first", changed_revision),
      repository_request(second, "github.com/example/second", second_revision)
    ], environment_inputs: [])

    before_by_identity = before.fetch("repositories").to_h { |snapshot| [snapshot.fetch("repository_identity"), snapshot] }
    after_by_identity = after.fetch("repositories").to_h { |snapshot| [snapshot.fetch("repository_identity"), snapshot] }
    refute_equal before_by_identity.fetch("github.com/example/first"), after_by_identity.fetch("github.com/example/first")
    assert_equal before_by_identity.fetch("github.com/example/second"), after_by_identity.fetch("github.com/example/second")
    refute_equal Manifest.fingerprint(before), Manifest.fingerprint(after)
  end

  def test_commit_and_descriptive_metadata_do_not_change_behavioral_fingerprint
    repository = create_repository("repository", "github.com/example/repository")
    first_revision = commit_files(repository, "lib/value.txt" => "stable\n")
    before = Manifest.build(repositories: [
      repository_request(repository, "github.com/example/repository", first_revision)
    ], environment_inputs: [environment_input("environment:toolchain", "stable")])

    environment = {
      "GIT_AUTHOR_DATE" => "2000-01-02T00:00:00Z",
      "GIT_COMMITTER_DATE" => "2000-01-02T00:00:00Z"
    }
    git(repository, "-c", "user.name=Identity Platform", "-c", "user.email=identity@example.invalid",
      "commit", "--quiet", "--allow-empty", "-m", "metadata only", environment: environment)
    second_revision = git(repository, "rev-parse", "HEAD")
    after = Manifest.build(repositories: [
      repository_request(repository, "github.com/example/repository", second_revision)
    ], environment_inputs: [environment_input("environment:toolchain", "stable")])
    after = JSON.parse(JSON.generate(after))
    after.fetch("repositories").fetch(0).fetch("inputs").fetch(0)["owner"] = "renamed owner"
    after.fetch("repositories").fetch(0).fetch("inputs").fetch(0)["reason"] = "renamed reason"
    after.fetch("environment_inputs").fetch(0)["owner"] = "renamed owner"
    after.fetch("environment_inputs").fetch(0)["reason"] = "renamed reason"

    refute_equal before.fetch("repositories").fetch(0).fetch("revision"),
      after.fetch("repositories").fetch(0).fetch("revision")
    assert_equal before.fetch("repositories").fetch(0).fetch("tree"),
      after.fetch("repositories").fetch(0).fetch("tree")
    assert_equal Manifest.fingerprint(before), Manifest.fingerprint(after)
  end

  def test_rebuilds_a_committed_manifest_in_each_authoritative_repository
    control = create_repository("control", "github.com/example/control")
    package = create_repository("package", "github.com/example/package")
    control_revision = commit_files(control, "contract.json" => "{}\n")
    package_revision = commit_files(package, "lib/value.txt" => "stable\n")
    document = Manifest.build(repositories: [
      repository_request(control, "github.com/example/control", control_revision,
        input_roots: [{path: "contract.json", owner: "coordinator", reason: "contract"}]),
      repository_request(package, "github.com/example/package", package_revision)
    ], environment_inputs: [environment_input("environment:toolchain", "stable")])

    rebuilt = Manifest.rebuild(
      document,
      repository_roots: {
        "github.com/example/control" => control,
        "github.com/example/package" => package
      }
    )

    assert_equal document, rebuilt
  end

  def test_rebuilds_one_repository_at_a_revalidation_revision_without_cross_repository_ancestry
    control = create_repository("control", "github.com/example/control")
    package = create_repository("package", "github.com/example/package")
    control_revision = commit_files(control, "contract.json" => "{}\n")
    tested_revision = commit_files(package, "lib/value.txt" => "stable\n")
    document = Manifest.build(repositories: [
      repository_request(control, "github.com/example/control", control_revision,
        input_roots: [{path: "contract.json", owner: "coordinator", reason: "contract"}]),
      repository_request(package, "github.com/example/package", tested_revision)
    ], environment_inputs: [])
    git(package, "-c", "user.name=Identity Platform", "-c", "user.email=identity@example.invalid",
      "commit", "--quiet", "--allow-empty", "-m", "metadata only")
    revalidation_revision = git(package, "rev-parse", "HEAD")

    rebuilt = Manifest.rebuild(
      document,
      repository_roots: {
        "github.com/example/control" => control,
        "github.com/example/package" => package
      },
      revision_overrides: {"github.com/example/package" => revalidation_revision}
    )

    package_snapshot = rebuilt.fetch("repositories").find { |row| row.fetch("repository_identity") == "github.com/example/package" }
    assert_equal revalidation_revision, package_snapshot.fetch("revision")
    assert_equal Manifest.fingerprint(document), Manifest.fingerprint(rebuilt)
  end

  def test_executable_mode_changes_behavioral_fingerprint_without_blob_change
    repository = create_repository("repository", "github.com/example/repository")
    first_revision = commit_files(repository, "lib/script.sh" => "#!/bin/sh\nexit 0\n")
    before = Manifest.build(repositories: [
      repository_request(repository, "github.com/example/repository", first_revision)
    ], environment_inputs: [])

    git(repository, "update-index", "--chmod=+x", "lib/script.sh")
    environment = {
      "GIT_AUTHOR_DATE" => "2000-01-02T00:00:00Z",
      "GIT_COMMITTER_DATE" => "2000-01-02T00:00:00Z"
    }
    git(repository, "-c", "user.name=Identity Platform", "-c", "user.email=identity@example.invalid",
      "commit", "--quiet", "-m", "make executable", environment: environment)
    second_revision = git(repository, "rev-parse", "HEAD")
    after = Manifest.build(repositories: [
      repository_request(repository, "github.com/example/repository", second_revision)
    ], environment_inputs: [])

    before_input = before.fetch("repositories").fetch(0).fetch("inputs").fetch(0)
    after_input = after.fetch("repositories").fetch(0).fetch("inputs").fetch(0)
    assert_equal before_input.fetch("content_identity"), after_input.fetch("content_identity")
    assert_equal "100644", before_input.fetch("mode")
    assert_equal "100755", after_input.fetch("mode")
    refute_equal Manifest.fingerprint(before), Manifest.fingerprint(after)
  end

  def test_records_sha256_repositories_with_runtime_independent_object_identities
    repository = create_repository("repository", "github.com/example/repository", object_format: "sha256")
    revision = commit_files(repository, "lib/value.txt" => "value\n")

    document = Manifest.build(repositories: [
      repository_request(repository, "github.com/example/repository", revision)
    ], environment_inputs: [])
    snapshot = document.fetch("repositories").fetch(0)

    assert_equal 64, revision.length
    assert_equal "git-sha256", snapshot.fetch("object_format")
    assert_match(/\Agit-sha256:[0-9a-f]{64}\z/, snapshot.fetch("tree"))
    assert_match(/\Agit-sha256:[0-9a-f]{64}\z/, snapshot.fetch("inputs").fetch(0).fetch("content_identity"))
  end

  def test_rejects_duplicate_repository_identities_and_resolved_roots
    first = create_repository("first", "github.com/example/first")
    second = create_repository("second", "github.com/example/second")
    first_revision = commit_files(first, "lib/value.txt" => "first\n")
    second_revision = commit_files(second, "lib/value.txt" => "second\n")

    error = assert_raises(Manifest::InvalidManifest) do
      Manifest.build(repositories: [
        repository_request(first, "github.com/example/first", first_revision),
        repository_request(second, "github.com/example/first", second_revision)
      ], environment_inputs: [])
    end
    assert_match(/duplicate repository identity/, error.message)

    alias_root = File.join(@temporary_root, "first-alias")
    File.symlink(first, alias_root)
    error = assert_raises(Manifest::InvalidManifest) do
      Manifest.build(repositories: [
        repository_request(first, "github.com/example/first", first_revision),
        repository_request(alias_root, "github.com/example/first-alias", first_revision)
      ], environment_inputs: [])
    end
    assert_match(/duplicate repository root/, error.message)
  end

  def test_rejects_missing_non_git_and_nested_repository_roots
    missing = File.join(@temporary_root, "missing")
    non_git = File.join(@temporary_root, "non-git")
    FileUtils.mkdir_p(non_git)
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n")

    [missing, non_git, File.join(repository, "lib")].each do |root|
      error = assert_raises(Manifest::InvalidManifest) do
        Manifest.build(repositories: [
          repository_request(root, "github.com/example/repository", revision)
        ], environment_inputs: [])
      end
      assert_match(/repository root/, error.message)
    end
  end

  def test_rejects_missing_or_non_exact_revisions
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n")

    [nil, "", "HEAD", revision[0, 12], "f" * revision.length].each do |candidate|
      error = assert_raises(Manifest::InvalidManifest) do
        Manifest.build(repositories: [
          repository_request(repository, "github.com/example/repository", candidate)
        ], environment_inputs: [])
      end
      assert_match(/revision/, error.message)
    end
  end

  def test_rejects_repository_identity_substitution
    repository = create_repository("repository", "github.com/example/canonical")
    revision = commit_files(repository, "lib/value.txt" => "value\n")

    error = assert_raises(Manifest::InvalidManifest) do
      Manifest.build(repositories: [
        repository_request(repository, "github.com/example/substitute", revision)
      ], environment_inputs: [])
    end

    assert_match(/repository identity substitution/, error.message)
  end

  def test_rejects_repository_identity_substitution_through_a_noncanonical_origin_port
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n")
    git(repository, "remote", "set-url", "origin", "https://github.com:8443/example/repository.git")

    error = assert_raises(Manifest::InvalidManifest) do
      Manifest.build(repositories: [
        repository_request(repository, "github.com/example/repository", revision)
      ], environment_inputs: [])
    end

    assert_match(/repository identity substitution/, error.message)
  end

  def test_rejects_absolute_traversing_and_noncanonical_input_roots
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n")

    ["/lib", "../lib", "lib/../lib", "lib//value.txt"].each do |path|
      request = repository_request(repository, "github.com/example/repository", revision,
        input_roots: [{path: path, owner: "library", reason: "behavioral source"}])
      error = assert_raises(Manifest::InvalidManifest) do
        Manifest.build(repositories: [request], environment_inputs: [])
      end
      assert_match(/input root/, error.message)
    end
  end

  def test_treats_input_roots_as_literal_paths_instead_of_git_pathspecs
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n", "secret.txt" => "secret\n")

    [":(exclude)secret.txt", "*.txt"].each do |path|
      request = repository_request(repository, "github.com/example/repository", revision,
        input_roots: [{path: path, owner: "library", reason: "literal tracked path"}])
      error = assert_raises(Manifest::InvalidManifest) do
        Manifest.build(repositories: [request], environment_inputs: [])
      end
      assert_match(/has no tracked inputs/, error.message)
    end
  end

  def test_build_rejects_malformed_tracked_paths
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/bad\\name" => "value\n")

    error = assert_raises(Manifest::InvalidManifest) do
      Manifest.build(repositories: [
        repository_request(repository, "github.com/example/repository", revision)
      ], environment_inputs: [])
    end

    assert_match(/tracked input semantics are invalid/, error.message)
  end

  def test_rejects_duplicate_tracked_input_identities_from_overlapping_roots
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n")
    request = repository_request(repository, "github.com/example/repository", revision,
      input_roots: [
        {path: ".", owner: "coordinator", reason: "repository input"},
        {path: "lib", owner: "library", reason: "library input"}
      ])

    error = assert_raises(Manifest::InvalidManifest) do
      Manifest.build(repositories: [request], environment_inputs: [])
    end

    assert_match(/duplicate input identity lib\/value.txt/, error.message)
  end

  def test_rejects_malformed_environment_inputs
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n")
    repositories = [repository_request(repository, "github.com/example/repository", revision)]
    valid = environment_input("environment:toolchain", "toolchain")
    malformed = [
      valid.merge(environment_id: "toolchain"),
      valid.merge(content_identity: "not-a-digest"),
      valid.merge(owner: ""),
      valid.merge(reason: ""),
      valid.merge(extra: "unexpected"),
      valid.transform_keys(&:to_s),
      valid.merge("extra" => "unexpected")
    ]

    malformed.each do |input|
      error = assert_raises(Manifest::InvalidManifest) do
        Manifest.build(repositories: repositories, environment_inputs: [input])
      end
      assert_match(/environment input/, error.message)
    end

    error = assert_raises(Manifest::InvalidManifest) do
      Manifest.build(repositories: repositories, environment_inputs: [valid, valid.dup])
    end
    assert_match(/duplicate environment input identity/, error.message)
  end

  def test_fingerprint_rejects_semantically_malformed_canonical_documents
    repository = create_repository("repository", "github.com/example/repository")
    revision = commit_files(repository, "lib/value.txt" => "value\n")
    valid = Manifest.build(repositories: [
      repository_request(repository, "github.com/example/repository", revision)
    ], environment_inputs: [environment_input("environment:toolchain", "toolchain")])
    mutations = [
      ->(document) { document.fetch("repositories").fetch(0)["repository_identity"] = "not-a-repository" },
      ->(document) { document.fetch("repositories").fetch(0).fetch("inputs").fetch(0)["path"] = "../escape" },
      ->(document) { document.fetch("repositories").fetch(0).fetch("inputs").fetch(0)["kind"] = "not-a-git-kind" },
      ->(document) { document.fetch("repositories").fetch(0).fetch("inputs").fetch(0)["content_identity"] = "not-a-digest" },
      ->(document) { document.fetch("repositories").fetch(0).fetch("inputs").fetch(0)["owner"] = "" },
      ->(document) { document.fetch("environment_inputs").fetch(0)["environment_id"] = "not-an-environment" },
      ->(document) { document.fetch("environment_inputs").fetch(0)["content_identity"] = "not-a-digest" },
      ->(document) { document.fetch("environment_inputs").fetch(0)["reason"] = "" }
    ]

    mutations.each do |mutation|
      forged = JSON.parse(JSON.generate(valid))
      mutation.call(forged)
      assert_raises(Manifest::InvalidManifest) { Manifest.fingerprint(forged) }
    end
  end

  private

  def create_repository(name, repository_identity, object_format: nil)
    path = File.join(@temporary_root, name)
    FileUtils.mkdir_p(path)
    arguments = ["init", "--quiet"]
    arguments << "--object-format=#{object_format}" if object_format
    git(path, *arguments)
    git(path, "remote", "add", "origin", "https://#{repository_identity}.git")
    path
  end

  def commit_files(repository, files)
    files.each do |path, content|
      target = File.join(repository, path)
      FileUtils.mkdir_p(File.dirname(target))
      File.binwrite(target, content)
    end
    git(repository, "add", *files.keys.sort)
    environment = {
      "GIT_AUTHOR_DATE" => "2000-01-01T00:00:00Z",
      "GIT_COMMITTER_DATE" => "2000-01-01T00:00:00Z"
    }
    git(repository, "-c", "user.name=Identity Platform", "-c", "user.email=identity@example.invalid",
      "commit", "--quiet", "-m", "fixture", environment: environment)
    git(repository, "rev-parse", "HEAD")
  end

  def repository_request(root, identity, revision, input_roots: nil)
    {
      repository_identity: identity,
      repository_root: root,
      revision: revision,
      input_roots: input_roots || [
        {path: "lib", owner: "library", reason: "behavioral source"}
      ]
    }
  end

  def environment_input(identity, value)
    {
      environment_id: identity,
      content_identity: "sha256:#{Digest::SHA256.hexdigest(value)}",
      owner: "coordinator",
      reason: "execution environment"
    }
  end

  def git(repository, *arguments, environment: {})
    output, error, status = Open3.capture3(environment, "git", "-C", repository, *arguments)
    raise "git #{arguments.first} failed: #{error}" unless status.success?

    output.strip
  end
end
