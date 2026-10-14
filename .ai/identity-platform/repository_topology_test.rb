#!/usr/bin/env ruby
# frozen_string_literal: true

require "minitest/autorun"
require "tmpdir"
require "fileutils"
require_relative "repository_topology"

class IdentityPlatformRepositoryTopologyTest < Minitest::Test
  CONTROL_ROOT = "/coord/go-library-tools"
  COORDINATION_ROOT = "/coord"

  def test_resolves_nested_logical_module
    topology = topology_for(
      row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")
    )

    resolution = topology.resolve("pkg/identity/postgres")

    assert_equal "identity", resolution.family
    assert_equal "/coord/go-identity", resolution.repository_root
    assert_equal "postgres", resolution.module_directory
    assert_equal "github.com/faustbrian/go-identity/postgres", resolution.module_path
    assert_equal "github.com/faustbrian/go-identity", resolution.repository_identity
  end

  def test_resolves_root_logical_module
    topology = topology_for(
      row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")
    )

    resolution = topology.resolve("pkg/identity")

    assert_equal ".", resolution.module_directory
    assert_equal "github.com/faustbrian/go-identity", resolution.module_path
  end

  def test_rejects_unknown_family
    topology = topology_for(
      row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")
    )

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      topology.resolve("pkg/sso")
    end

    assert_match(/unknown logical family sso/, error.message)
  end

  def test_rejects_duplicate_family
    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      topology_for(
        row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity"),
        row("identity", "/coord/go-other", "github.com/faustbrian/go-other")
      )
    end

    assert_match(/duplicate logical family identity/, error.message)
  end

  def test_rejects_repository_outside_coordination_root
    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      topology_for(
        row("identity", "/outside/go-identity", "github.com/faustbrian/go-identity")
      )
    end

    assert_match(/escapes coordination root/, error.message)
  end

  def test_rejects_control_repository_as_package_repository
    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      topology_for(
        row("identity", CONTROL_ROOT, "github.com/faustbrian/go-library-tools")
      )
    end

    assert_match(/control repository cannot own package family/, error.message)
  end

  def test_rejects_wrong_module_prefix
    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      topology_for(
        row("identity", "/coord/go-identity", "github.com/example/go-identity")
      )
    end

    assert_match(/invalid module prefix/, error.message)
  end

  def test_rejects_repository_owned_by_wrong_family
    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      topology_for(
        row("sso", "/coord/go-identity", "github.com/faustbrian/go-identity")
      )
    end

    assert_match(/repository name for sso must be go-sso/, error.message)
  end

  def test_rejects_logical_path_escape_and_empty_segments
    topology = topology_for(
      row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")
    )

    %w[pkg/identity/../sso pkg/identity//postgres /pkg/identity pkg/identity/].each do |logical_module|
      assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology, logical_module) do
        topology.resolve(logical_module)
      end
    end
  end

  def test_rejects_duplicate_repository_identity
    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      topology_for(
        row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity"),
        row("sso", "/coord/go-sso", "github.com/faustbrian/go-identity")
      )
    end

    assert_match(/duplicate repository identity/, error.message)
  end

  def test_loads_declared_control_and_coordination_roots
    source = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}

      The execution copy MUST live in
      `/coord/go-library-tools/.ai/identity-platform` on its coordinator branch.
    MARKDOWN

    topology = IdentityPlatformRepositoryTopology.parse_declared(source)

    assert_equal COORDINATION_ROOT, topology.coordination_root
    assert_equal CONTROL_ROOT, topology.control_repository_root
    assert_equal "/coord/go-identity", topology.resolve("pkg/identity").repository_root
  end

  def test_rejects_missing_declared_control_repository
    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse_declared(
        row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")
      )
    end

    assert_match(/control repository declaration is missing/, error.message)
  end

  def test_rejects_ambiguous_declared_control_repository
    source = <<~MARKDOWN
      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}
      `/coord/go-library-tools/.ai/identity-platform`
      `/other/go-library-tools/.ai/identity-platform`
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse_declared(source)
    end

    assert_match(/control repository declaration is ambiguous/, error.message)
  end

  def test_rejects_mapping_row_without_canonical_table
    source = <<~MARKDOWN
      # No mapping table

      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse(
        source,
        coordination_root: COORDINATION_ROOT,
        control_repository_root: CONTROL_ROOT
      )
    end

    assert_match(/mapping table is missing/, error.message)
  end

  def test_rejects_duplicate_mapping_tables
    table = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse(
        "#{table}\n#{table}",
        coordination_root: COORDINATION_ROOT,
        control_repository_root: CONTROL_ROOT
      )
    end

    assert_match(/mapping table is ambiguous/, error.message)
  end

  def test_rejects_malformed_row_inside_mapping_table
    source = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      | `identity` | `/coord/go-identity` | `not-a-module-prefix` |
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse(
        source,
        coordination_root: COORDINATION_ROOT,
        control_repository_root: CONTROL_ROOT
      )
    end

    assert_match(/malformed mapping row/, error.message)
  end

  def test_rejects_existing_repository_symlink_that_escapes_coordination_root
    Dir.mktmpdir("identity-topology-") do |directory|
      coordination_root = File.join(directory, "coord")
      outside_root = File.join(directory, "outside", "go-identity")
      control_root = File.join(coordination_root, "go-library-tools")
      FileUtils.mkdir_p(outside_root)
      FileUtils.mkdir_p(control_root)
      File.symlink(outside_root, File.join(coordination_root, "go-identity"))

      error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
        topology_for_paths(coordination_root, control_root)
      end

      assert_match(/resolves outside coordination root/, error.message)
    end
  end

  def test_rejects_existing_repository_symlink_to_control_repository
    Dir.mktmpdir("identity-topology-") do |directory|
      coordination_root = File.join(directory, "coord")
      control_root = File.join(coordination_root, "go-library-tools")
      FileUtils.mkdir_p(control_root)
      File.symlink(control_root, File.join(coordination_root, "go-identity"))

      error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
        topology_for_paths(coordination_root, control_root)
      end

      assert_match(/resolves within control repository/, error.message)
    end
  end

  def test_rejects_missing_repository_beneath_symlinked_escaping_ancestor
    Dir.mktmpdir("identity-topology-") do |directory|
      coordination_root = File.join(directory, "coord")
      outside_root = File.join(directory, "outside")
      control_root = File.join(coordination_root, "go-library-tools")
      FileUtils.mkdir_p(outside_root)
      FileUtils.mkdir_p(control_root)
      File.symlink(outside_root, File.join(coordination_root, "alias"))

      source = <<~MARKDOWN
        | Logical family | Standalone repository | Module prefix |
        | --- | --- | --- |
        #{row("identity", File.join(coordination_root, "alias", "go-identity"), "github.com/faustbrian/go-identity")}
      MARKDOWN
      error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
        IdentityPlatformRepositoryTopology.parse(
          source,
          coordination_root: coordination_root,
          control_repository_root: control_root
        )
      end

      assert_match(/resolves outside coordination root/, error.message)
    end
  end

  def test_rejects_row_like_content_outside_mapping_table
    source = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}

      | `sso` | broken |
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse(
        source,
        coordination_root: COORDINATION_ROOT,
        control_repository_root: CONTROL_ROOT
      )
    end

    assert_match(/mapping row appears outside canonical table/, error.message)
  end

  def test_rejects_indented_unbackticked_table_row_outside_mapping_table
    source = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}

        | sso | /coord/go-sso | github.com/faustbrian/go-sso |
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse(
        source,
        coordination_root: COORDINATION_ROOT,
        control_repository_root: CONTROL_ROOT
      )
    end

    assert_match(/mapping row appears outside canonical table/, error.message)
  end

  def test_rejects_missing_repository_beneath_control_repository_alias
    Dir.mktmpdir("identity-topology-") do |directory|
      coordination_root = File.join(directory, "coord")
      control_root = File.join(coordination_root, "go-library-tools")
      FileUtils.mkdir_p(control_root)
      File.symlink(control_root, File.join(coordination_root, "alias"))
      source = <<~MARKDOWN
        | Logical family | Standalone repository | Module prefix |
        | --- | --- | --- |
        #{row("identity", File.join(coordination_root, "alias", "go-identity"), "github.com/faustbrian/go-identity")}
      MARKDOWN

      error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
        IdentityPlatformRepositoryTopology.parse(
          source,
          coordination_root: coordination_root,
          control_repository_root: control_root
        )
      end

      assert_match(/resolves within control repository/, error.message)
    end
  end

  def test_rejects_existing_non_directory_repository_path
    Dir.mktmpdir("identity-topology-") do |directory|
      coordination_root = File.join(directory, "coord")
      control_root = File.join(coordination_root, "go-library-tools")
      FileUtils.mkdir_p(control_root)
      repository_root = File.join(coordination_root, "go-identity")
      File.write(repository_root, "not a repository")
      source = <<~MARKDOWN
        | Logical family | Standalone repository | Module prefix |
        | --- | --- | --- |
        #{row("identity", repository_root, "github.com/faustbrian/go-identity")}
      MARKDOWN

      error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
        IdentityPlatformRepositoryTopology.parse(
          source,
          coordination_root: coordination_root,
          control_repository_root: control_root
        )
      end

      assert_match(/does not resolve through a directory/, error.message)
    end
  end

  def test_rejects_pipe_optional_gfm_mapping_table_outside_canonical_table
    source = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}

      Logical family | Standalone repository | Module prefix
      --- | --- | ---
      sso | /coord/go-sso | github.com/faustbrian/go-sso
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse(
        source,
        coordination_root: COORDINATION_ROOT,
        control_repository_root: CONTROL_ROOT
      )
    end

    assert_match(/mapping row appears outside canonical table/, error.message)
  end

  def test_rejects_malformed_pipe_optional_row_outside_canonical_table
    source = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{row("identity", "/coord/go-identity", "github.com/faustbrian/go-identity")}

      `sso` | `/coord/go-sso` | `not-a-module-prefix`
    MARKDOWN

    error = assert_raises(IdentityPlatformRepositoryTopology::InvalidTopology) do
      IdentityPlatformRepositoryTopology.parse(
        source,
        coordination_root: COORDINATION_ROOT,
        control_repository_root: CONTROL_ROOT
      )
    end

    assert_match(/mapping row appears outside canonical table/, error.message)
  end

  private

  def topology_for(*rows)
    source = <<~MARKDOWN
      # Repository map

      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{rows.join("\n")}
    MARKDOWN

    IdentityPlatformRepositoryTopology.parse(
      source,
      coordination_root: COORDINATION_ROOT,
      control_repository_root: CONTROL_ROOT
    )
  end

  def topology_for_paths(coordination_root, control_root)
    source = <<~MARKDOWN
      | Logical family | Standalone repository | Module prefix |
      | --- | --- | --- |
      #{row("identity", File.join(coordination_root, "go-identity"), "github.com/faustbrian/go-identity")}
    MARKDOWN

    IdentityPlatformRepositoryTopology.parse(
      source,
      coordination_root: coordination_root,
      control_repository_root: control_root
    )
  end

  def row(family, repository, module_prefix)
    "| `#{family}` | `#{repository}` | `#{module_prefix}` |"
  end
end
