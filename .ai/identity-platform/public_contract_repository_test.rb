# frozen_string_literal: true

require "fileutils"
require "minitest/autorun"
require "open3"
require "tmpdir"

require_relative "public_contract_repository"

class IdentityPlatformPublicContractRepositoryTest < Minitest::Test
  Resolver = IdentityPlatformPublicContractRepository

  def setup
    @temporary_root = Dir.mktmpdir("identity-public-contract-repositories-")
    @control_root = create_repository("go-library-tools", "github.com/faustbrian/go-library-tools")
    @audit_root = create_repository("go-audit", "github.com/faustbrian/go-audit")
    write_module(@audit_root, "github.com/faustbrian/go-audit")
    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Builder struct{}\n")
    FileUtils.mkdir_p(File.join(@audit_root, "postgres"))
    File.binwrite(File.join(@audit_root, "postgres", "store.go"), "package postgres\n\ntype Store struct{}\n")
    FileUtils.mkdir_p(File.join(@audit_root, "api"))
    File.binwrite(File.join(@audit_root, "api", "baseline.txt"), "type Builder struct\n")
    commit_audit
  end

  def teardown
    FileUtils.remove_entry(@temporary_root)
  end

  def test_resolves_root_and_nested_packages_from_the_authoritative_standalone_repository
    resolver = build_resolver(
      "audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"),
      "auditpostgres" => entry("go-audit", "postgres", "github.com/faustbrian/go-audit/postgres")
    )

    audit = resolver.resolve("audit")
    postgres = resolver.resolve("auditpostgres")

    assert_equal File.realpath(@audit_root), audit.repository_root
    assert_equal File.realpath(@audit_root), audit.package_directory
    assert_equal "github.com/faustbrian/go-audit", audit.package_path
    assert_equal File.realpath(File.join(@audit_root, "postgres")), postgres.package_directory
    assert_equal "github.com/faustbrian/go-audit/postgres", postgres.package_path
  end

  def test_rejects_a_repository_with_the_wrong_remote_identity
    error = assert_raises(Resolver::InvalidRepository) do
      build_resolver("audit" => {
        repository_root: @audit_root,
        repository_identity: "github.com/faustbrian/go-substitute",
        package_directory: ".",
        package_path: "github.com/faustbrian/go-substitute"
      })
    end

    assert_match(/repository identity substitution/, error.message)
  end

  def test_rejects_package_paths_that_escape_the_standalone_repository
    error = assert_raises(Resolver::InvalidRepository) do
      build_resolver("audit" => entry("go-audit", "../go-library-tools", "github.com/faustbrian/go-audit"))
    end

    assert_match(/package directory.*escapes/, error.message)
  end

  def test_rejects_package_paths_that_escape_through_a_symlink
    escaped = File.join(@temporary_root, "escaped-package")
    FileUtils.mkdir_p(escaped)
    File.binwrite(File.join(escaped, "escape.go"), "package escape\n")
    File.symlink(escaped, File.join(@audit_root, "escape"))

    error = assert_raises(Resolver::InvalidRepository) do
      build_resolver("audit" => entry("go-audit", "escape", "github.com/faustbrian/go-audit/escape"))
    end

    assert_match(/package directory.*escapes/, error.message)
  end

  def test_rejects_archived_monorepo_packages_even_when_their_source_exists
    FileUtils.mkdir_p(File.join(@control_root, "pkg", "audit"))
    File.binwrite(File.join(@control_root, "pkg", "audit", "audit.go"), "package audit\n")
    write_module(@control_root, "github.com/faustbrian/go-library-tools")

    error = assert_raises(Resolver::InvalidRepository) do
      Resolver.new(
        entries: {"audit" => {
          repository_root: @control_root,
          repository_identity: "github.com/faustbrian/go-library-tools",
          package_directory: "pkg/audit",
          package_path: "github.com/faustbrian/golib/pkg/audit"
        }},
        forbidden_repository_roots: [@control_root]
      )
    end

    assert_match(/control repository/, error.message)
  end

  def test_rejects_a_symlinked_production_source_file
    external = File.join(@temporary_root, "external.go")
    File.binwrite(external, "package audit\n\ntype External struct{}\n")
    FileUtils.rm(File.join(@audit_root, "record.go"))
    File.symlink(external, File.join(@audit_root, "record.go"))
    commit_audit
    resolver = build_resolver("audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"))

    error = assert_raises(Resolver::InvalidRepository) { resolver.public_source_files("audit") }

    assert_match(/production source.*symbolic link/, error.message)
  end

  def test_rejects_a_symlinked_go_mod_file
    external = File.join(@temporary_root, "external.mod")
    File.binwrite(external, "module github.com/faustbrian/go-audit\n")
    FileUtils.rm(File.join(@audit_root, "go.mod"))
    File.symlink(external, File.join(@audit_root, "go.mod"))
    commit_audit
    resolver = build_resolver("audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"))

    error = assert_raises(Resolver::InvalidRepository) { resolver.public_source_files("audit") }

    assert_match(/go.mod.*symbolic link/, error.message)
  end

  def test_rejects_a_symlinked_api_baseline_file
    external = File.join(@temporary_root, "external-baseline.txt")
    File.binwrite(external, "external bytes\n")
    FileUtils.mkdir_p(File.join(@audit_root, "api"))
    FileUtils.rm(File.join(@audit_root, "api", "baseline.txt"))
    File.symlink(external, File.join(@audit_root, "api", "baseline.txt"))
    commit_audit
    resolver = build_resolver("audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"))

    error = assert_raises(Resolver::InvalidRepository) { resolver.public_source_files("audit") }

    assert_match(/API baseline.*symbolic link/, error.message)
  end

  def test_rejects_missing_aliases_and_duplicate_alias_declarations
    resolver = build_resolver("audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"))

    assert_raises(Resolver::InvalidRepository) { resolver.resolve("missing") }
    assert_raises(Resolver::InvalidRepository) do
      Resolver.new(entries: [["audit", entry("go-audit", ".", "github.com/faustbrian/go-audit")], ["audit", entry("go-audit", ".", "github.com/faustbrian/go-audit")]], forbidden_repository_roots: [@control_root])
    end
  end

  def test_source_inventory_ignores_index_and_worktree_changes
    FileUtils.mkdir_p(File.join(@audit_root, "api"))
    baseline = File.join(@audit_root, "api", "baseline.txt")
    File.binwrite(baseline, "type Builder struct\n")
    resolver = build_resolver("audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"))

    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Changed struct{}\n")
    git(@audit_root, "add", "record.go")
    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Unstaged struct{}\n")
    FileUtils.mv(baseline, File.join(@audit_root, "api", "replacement.txt"))
    File.binwrite(File.join(@audit_root, "untracked.go"), "package audit\n\ntype Untracked struct{}\n")

    files = resolver.public_source_files("audit")
    source_root = File.realpath(@audit_root)
    assert_equal [File.join(source_root, "record.go")], files.production_sources
    assert_equal [File.join(source_root, "api", "baseline.txt")], files.api_baselines
    assert_equal "package audit\n\ntype Builder struct{}\n", files.read(File.join(source_root, "record.go"))
    assert_equal "type Builder struct\n", files.read(File.join(source_root, "api", "baseline.txt"))
  end

  def test_root_and_nested_reads_keep_the_captured_revision_until_a_new_resolver
    declarations = {
      "audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"),
      "auditpostgres" => entry("go-audit", "postgres", "github.com/faustbrian/go-audit/postgres")
    }
    resolver = build_resolver(declarations)
    root = File.realpath(@audit_root)
    record = File.join(root, "record.go")
    store = File.join(root, "postgres", "store.go")
    File.binwrite(record, "package audit\n\ntype Changed struct{}\n")
    git(@audit_root, "add", "record.go")
    git(@audit_root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
        "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "advance")
    FileUtils.remove_entry(File.join(@audit_root, "postgres"))

    assert_equal "package audit\n\ntype Builder struct{}\n", resolver.public_source_files("audit").read(record)
    assert_equal "package postgres\n\ntype Store struct{}\n", resolver.public_source_files("auditpostgres").read(store)
    fresh = build_resolver(declarations)
    assert_equal "package audit\n\ntype Changed struct{}\n", fresh.public_source_files("audit").read(record)
    refute_equal resolver.resolve("audit").source_revision, fresh.resolve("audit").source_revision
    assert_equal resolver.resolve("audit").source_revision, resolver.resolve("auditpostgres").source_revision
  end

  def test_rejects_a_symbolic_revision_override
    error = assert_raises(Resolver::InvalidRepository) do
      Resolver.new(
        entries: {"audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit")},
        forbidden_repository_roots: [@control_root],
        revisions: {"github.com/faustbrian/go-audit" => "main"}
      )
    end
    assert_match(/exact commit/, error.message)
  end

  def test_explicit_revision_keeps_the_recorded_baseline_and_copies_its_configuration
    declarations = {"audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit")}
    baseline = build_resolver(declarations).resolve("audit").source_revision
    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Changed struct{}\n")
    commit_audit
    revisions = {"github.com/faustbrian/go-audit" => baseline.dup}
    resolver = Resolver.new(entries: declarations, forbidden_repository_roots: [@control_root], revisions: revisions)
    revisions.values.first.replace("0" * 40)

    path = File.join(File.realpath(@audit_root), "record.go")
    assert_equal baseline, resolver.resolve("audit").source_revision
    assert_equal "package audit\n\ntype Builder struct{}\n", resolver.public_source_files("audit").read(path)
    assert_equal "package audit\n\ntype Changed struct{}\n", build_resolver(declarations).public_source_files("audit").read(path)
  end

  def test_rejects_oversized_committed_source_before_reading_blob_contents
    File.binwrite(File.join(@audit_root, "record.go"), "x" * (4 * 1024 * 1024 + 1))
    commit_audit
    resolver = build_resolver("audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"))

    error = assert_raises(Resolver::InvalidRepository) { resolver.public_source_files("audit") }
    assert_match(/byte limit/, error.message)
  end

  def test_catalog_uses_integration_descendants_and_preserves_primitive_revisions
    baseline = git_output(@audit_root, "rev-parse", "HEAD").strip
    git(@audit_root, "switch", "--quiet", "-c", "feature/identity-platform")
    primitive = create_repository("go-clock", "github.com/faustbrian/go-clock")
    write_module(primitive, "github.com/faustbrian/go-clock")
    File.binwrite(File.join(primitive, "clock.go"), "package clock\n\ntype Original struct{}\n")
    commit_files(primitive, "go.mod", "clock.go")
    primitive_baseline = git_output(primitive, "rev-parse", "HEAD").strip

    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Advanced struct{}\n")
    commit_audit
    File.binwrite(File.join(primitive, "clock.go"), "package clock\n\ntype Later struct{}\n")
    commit_files(primitive, "clock.go")
    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Dirty struct{}\n")
    git(@audit_root, "add", "record.go")
    File.binwrite(File.join(primitive, "clock.go"), "package clock\n\ntype Dirty struct{}\n")
    FileUtils.rm(File.join(@audit_root, "api", "baseline.txt"))

    catalog = Resolver::Catalog.new(
      entries: {
        "audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"),
        "clock" => entry("go-clock", ".", "github.com/faustbrian/go-clock")
      },
      forbidden_repository_roots: [@control_root],
      primitive_records: [
        {"repository_identity" => "github.com/faustbrian/go-audit", "repository_revision" => baseline},
        {"repository_identity" => "github.com/faustbrian/go-clock", "repository_revision" => primitive_baseline}
      ],
      integration_source: integration_source(baseline)
    )
    audit = catalog.contract("audit", "pkg/audit")
    clock = catalog.contract("clock", "pkg/clock")
    assert_equal ["audit.Advanced"], audit.fetch("exported_symbols")
    assert_equal ["clock.Original"], clock.fetch("exported_symbols")
    expected_files = {
      "pkg/audit/api/baseline.txt" => "type Builder struct\n",
      "pkg/audit/go.mod" => "module github.com/faustbrian/go-audit\n\ngo 1.26.0\n",
      "pkg/audit/record.go" => "package audit\n\ntype Advanced struct{}\n"
    }
    encoded = expected_files.map { |path, bytes| "#{path}\0#{bytes.bytesize}\0#{bytes}\0" }.join
    assert_equal expected_files.keys, audit.fetch("api_baseline_files")
    assert_equal "sha256:#{Digest::SHA256.hexdigest(encoded)}", audit.fetch("api_baseline_sha256")
    assert_equal primitive_baseline, catalog.resolve("clock").source_revision
    assert_equal git_output(@audit_root, "rev-parse", "HEAD").strip, catalog.resolve("audit").source_revision
  end

  def test_catalog_rejects_detached_and_non_descendant_integration_heads
    baseline = git_output(@audit_root, "rev-parse", "HEAD").strip
    git(@audit_root, "switch", "--quiet", "--detach")
    assert_match(/integration branch/, assert_raises(Resolver::InvalidRepository) { build_catalog(baseline) }.message)
    git(@audit_root, "switch", "--quiet", "-c", "feature/identity-platform")
    tree = git_output(@audit_root, "rev-parse", "HEAD^{tree}").strip
    unrelated = git_output(@audit_root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
                           "-c", "commit.gpgsign=false", "commit-tree", tree, "-m", "unrelated").strip
    # A fixture-only ref update creates the same-tree non-descendant case.
    git(@audit_root, "update-ref", "refs/heads/feature/identity-platform", unrelated)
    assert_match(/integration baseline/, assert_raises(Resolver::InvalidRepository) { build_catalog(baseline) }.message)
  end

  def test_catalog_rejects_an_omitted_non_integration_primitive_revision
    error = assert_raises(Resolver::InvalidRepository) do
      Resolver::Catalog.new(
        entries: {"audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit")},
        forbidden_repository_roots: [@control_root], primitive_records: [],
        integration_source: "## Repository integration identities\n## Next section\n"
      )
    end
    assert_match(/missing recorded primitive revision/, error.message)
  end

  def test_catalog_can_pin_an_external_source_without_inventing_a_primitive_gate
    baseline = git_output(@audit_root, "rev-parse", "HEAD").strip
    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Later struct{}\n")
    commit_audit
    File.binwrite(File.join(@audit_root, "record.go"), "package audit\n\ntype Dirty struct{}\n")
    revisions = {"github.com/faustbrian/go-audit" => baseline.dup}
    catalog = standalone_catalog(revisions)
    revisions.values.first.replace("0" * 40)

    assert_equal baseline, catalog.resolve("audit").source_revision
    assert_equal ["audit.Builder"], catalog.contract("audit", "pkg/audit").fetch("exported_symbols")
    files = {
      "pkg/audit/api/baseline.txt" => "type Builder struct\n",
      "pkg/audit/go.mod" => "module github.com/faustbrian/go-audit\n\ngo 1.26.0\n",
      "pkg/audit/record.go" => "package audit\n\ntype Builder struct{}\n"
    }
    bytes = files.map { |path, contents| "#{path}\0#{contents.bytesize}\0#{contents}\0" }.join
    assert_equal "sha256:#{Digest::SHA256.hexdigest(bytes)}", catalog.contract("audit", "pkg/audit").fetch("api_baseline_sha256")
  end

  def test_catalog_source_pins_cannot_override_existing_authorities
    baseline = git_output(@audit_root, "rev-parse", "HEAD").strip
    revisions = {"github.com/faustbrian/go-audit" => baseline}
    records = [{"repository_identity" => "github.com/faustbrian/go-audit", "repository_revision" => baseline}]
    assert_match(/already recorded/, assert_raises(Resolver::InvalidRepository) do
      standalone_catalog(revisions, primitive_records: records)
    end.message)
    git(@audit_root, "switch", "--quiet", "-c", "feature/identity-platform")
    assert_match(/already recorded/, assert_raises(Resolver::InvalidRepository) do
      standalone_catalog(revisions, integration: integration_source(baseline))
    end.message)
  end

  def test_catalog_rejects_malformed_unknown_and_unavailable_source_pins
    [nil, [], {"github.com/faustbrian/go-audit" => "HEAD"},
     {"github.com/faustbrian/go-audit" => "f" * 39}].each do |revisions|
      assert_raises(Resolver::InvalidRepository) { standalone_catalog(revisions) }
    end
    assert_match(/unmapped/, assert_raises(Resolver::InvalidRepository) do
      standalone_catalog({"github.com/faustbrian/go-unmapped" => "f" * 40})
    end.message)
    assert_match(/unavailable/, assert_raises(Resolver::InvalidRepository) do
      standalone_catalog({"github.com/faustbrian/go-audit" => "0" * 40})
    end.message)
  end

  def test_committed_source_bytes_ignore_same_size_blob_replacements
    object = git_output(@audit_root, "rev-parse", "HEAD:record.go").strip
    replacement_bytes = "package audit\n\ntype EvilOne struct{}\n"
    replacement, error, status = Open3.capture3("git", "-C", @audit_root, "hash-object", "-w", "--stdin",
                                               stdin_data: replacement_bytes)
    raise error unless status.success?
    git(@audit_root, "replace", object, replacement.strip)
    resolver = build_resolver("audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit"))

    assert_equal "package audit\n\ntype Builder struct{}\n",
                 resolver.public_source_files("audit").read(File.join(File.realpath(@audit_root), "record.go"))
  end

  def test_integration_ancestry_ignores_commit_replacements
    baseline = git_output(@audit_root, "rev-parse", "HEAD").strip
    git(@audit_root, "switch", "--quiet", "-c", "feature/identity-platform")
    tree = git_output(@audit_root, "rev-parse", "HEAD^{tree}").strip
    unrelated = git_output(@audit_root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
                           "-c", "commit.gpgsign=false", "commit-tree", tree, "-m", "unrelated").strip
    git(@audit_root, "update-ref", "refs/heads/feature/identity-platform", unrelated)
    descendant = git_output(@audit_root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
                            "-c", "commit.gpgsign=false", "commit-tree", tree, "-p", baseline, "-m", "descendant").strip
    git(@audit_root, "replace", unrelated, descendant)

    assert_match(/integration baseline/, assert_raises(Resolver::InvalidRepository) { build_catalog(baseline) }.message)
  end

  def test_integration_ancestry_ignores_legacy_grafts
    baseline = git_output(@audit_root, "rev-parse", "HEAD").strip
    git(@audit_root, "switch", "--quiet", "-c", "feature/identity-platform")
    tree = git_output(@audit_root, "rev-parse", "HEAD^{tree}").strip
    unrelated = git_output(@audit_root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
                           "-c", "commit.gpgsign=false", "commit-tree", tree, "-m", "unrelated").strip
    git(@audit_root, "update-ref", "refs/heads/feature/identity-platform", unrelated)
    git(@audit_root, "config", "advice.graftFileDeprecated", "false")
    File.binwrite(File.join(@audit_root, ".git", "info", "grafts"), "#{unrelated} #{baseline}\n")

    assert_match(/integration baseline/, assert_raises(Resolver::InvalidRepository) { build_catalog(baseline) }.message)
  end

  private

  def standalone_catalog(revisions, primitive_records: [], integration: "## Repository integration identities\n## Next section\n")
    Resolver::Catalog.new(
      entries: {"audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit")},
      forbidden_repository_roots: [@control_root], primitive_records: primitive_records,
      integration_source: integration, source_revisions: revisions
    )
  end

  def integration_source(baseline)
    "## Repository integration identities\n" \
      "| github.com/faustbrian/go-audit | git-sha1 | #{baseline} | git-sha1:#{'0' * 40} | #{baseline} | git-sha1:#{'0' * 40} | existing | feature/identity-platform | #{@audit_root} | fixture |\n" \
      "## Next section\n"
  end

  def build_catalog(baseline)
    Resolver::Catalog.new(
      entries: {"audit" => entry("go-audit", ".", "github.com/faustbrian/go-audit")},
      forbidden_repository_roots: [@control_root], primitive_records: [],
      integration_source: integration_source(baseline)
    )
  end

  def commit_files(repository, *paths)
    git(repository, "add", *paths)
    git(repository, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
        "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture")
  end

  def git_output(repository, *arguments)
    output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    raise error unless status.success?
    output
  end

  def commit_audit
    git(@audit_root, "add", "go.mod", "record.go", "postgres", "api")
    git(@audit_root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
        "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture")
  end

  def build_resolver(entries)
    Resolver.new(entries: entries, forbidden_repository_roots: [@control_root])
  end

  def entry(repository_name, package_directory, package_path)
    {
      repository_root: File.join(@temporary_root, repository_name),
      repository_identity: "github.com/faustbrian/#{repository_name}",
      package_directory: package_directory,
      package_path: package_path
    }
  end

  def create_repository(name, identity)
    path = File.join(@temporary_root, name)
    FileUtils.mkdir_p(path)
    git(path, "init", "--quiet")
    git(path, "remote", "add", "origin", "https://#{identity}.git")
    path
  end

  def write_module(repository, module_path)
    File.binwrite(File.join(repository, "go.mod"), "module #{module_path}\n\ngo 1.26.0\n")
  end

  def git(repository, *arguments)
    _output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    raise error unless status.success?
  end
end
