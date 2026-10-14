# frozen_string_literal: true

require "fileutils"
require "json"
require "minitest/autorun"
require "open3"
require "tmpdir"

require_relative "standalone_package_registry"

class IdentityPlatformStandalonePackageRegistryTest < Minitest::Test
  Registry = IdentityPlatformStandalonePackageRegistry

  def setup
    @coordination_root = Dir.mktmpdir("identity-package-registry-")
    @control_root = create_repository("go-library-tools")
  end

  def teardown
    FileUtils.remove_entry(@coordination_root)
  end

  def test_discovers_root_nested_module_and_nested_package_registrations
    repository = create_repository("go-authentication")
    write_catalogs(
      repository,
      modules: [
        module_entry(".", "github.com/faustbrian/go-authentication"),
        module_entry("jwt", "github.com/faustbrian/go-authentication/jwt")
      ],
      packages: [
        package_entry(".", "github.com/faustbrian/go-authentication"),
        package_entry("oidc", "github.com/faustbrian/go-authentication/oidc")
      ]
    )

    registered = build_registry.registered_consumables(%w[authentication authentication/jwt authentication/oidc])

    assert_equal Set.new(%w[authentication authentication/jwt authentication/oidc]), registered
  end

  def test_maps_the_outbox_logical_family_to_transactional_outbox
    repository = create_repository("go-transactional-outbox")
    write_catalogs(
      repository,
      modules: [module_entry(".", "github.com/faustbrian/go-transactional-outbox")],
      packages: [package_entry("postgres", "github.com/faustbrian/go-transactional-outbox/postgres")]
    )

    registered = build_registry.registered_consumables(%w[outbox outbox/postgres])

    assert_equal Set.new(%w[outbox outbox/postgres]), registered
    assert_equal File.realpath(repository), build_registry.repository_root_for("outbox/postgres")
  end

  def test_rejects_a_requested_consumable_absent_from_both_catalogs
    repository = create_repository("go-capability")
    write_catalogs(
      repository,
      modules: [module_entry(".", "github.com/faustbrian/go-capability")],
      packages: [package_entry(".", "github.com/faustbrian/go-capability")]
    )

    error = assert_raises(Registry::InvalidRegistry) do
      build_registry.registered_consumables(%w[capability capability/postgres])
    end

    assert_match(/capability\/postgres is not registered/, error.message)
  end

  def test_rejects_a_repository_with_a_substituted_origin
    repository = create_repository("go-audit", identity: "github.com/faustbrian/go-substitute")
    write_catalogs(
      repository,
      modules: [module_entry(".", "github.com/faustbrian/go-audit")],
      packages: [package_entry(".", "github.com/faustbrian/go-audit")]
    )

    error = assert_raises(Registry::InvalidRegistry) do
      build_registry.registered_consumables(%w[audit])
    end

    assert_match(/repository identity substitution/, error.message)
  end

  def test_rejects_a_catalog_that_claims_another_repository
    repository = create_repository("go-tenancy")
    write_catalogs(
      repository,
      repository_identity: "github.com/faustbrian/go-substitute",
      modules: [module_entry(".", "github.com/faustbrian/go-tenancy")],
      packages: [package_entry(".", "github.com/faustbrian/go-tenancy")]
    )

    error = assert_raises(Registry::InvalidRegistry) do
      build_registry.registered_consumables(%w[tenancy])
    end

    assert_match(/modules manifest repository drifted/, error.message)
  end

  def test_rejects_symlinked_catalogs
    repository = create_repository("go-identifier")
    external = File.join(@coordination_root, "external-modules.json")
    File.binwrite(external, JSON.generate("repository" => "github.com/faustbrian/go-identifier", "modules" => []))
    File.symlink(external, File.join(repository, "modules.json"))
    File.binwrite(File.join(repository, "packages.json"), JSON.generate("repository" => "github.com/faustbrian/go-identifier", "packages" => []))

    error = assert_raises(Registry::InvalidRegistry) do
      build_registry.registered_consumables(%w[identifier])
    end

    assert_match(/modules.json is a symbolic link/, error.message)
  end

  private

  def build_registry
    Registry.new(coordination_root: @coordination_root, control_repository_root: @control_root)
  end

  def create_repository(name, identity: "github.com/faustbrian/#{name}")
    path = File.join(@coordination_root, name)
    FileUtils.mkdir_p(path)
    git(path, "init", "--quiet")
    git(path, "remote", "add", "origin", "https://#{identity}.git")
    path
  end

  def write_catalogs(repository, repository_identity: "github.com/faustbrian/#{File.basename(repository)}", modules:, packages:)
    File.binwrite(
      File.join(repository, "modules.json"),
      JSON.generate("repository" => repository_identity, "modules" => modules)
    )
    File.binwrite(
      File.join(repository, "packages.json"),
      JSON.generate("repository" => repository_identity, "packages" => packages)
    )
  end

  def module_entry(directory, module_path)
    {"directory" => directory, "module_path" => module_path}
  end

  def package_entry(directory, import_path)
    {"directory" => directory, "import_path" => import_path}
  end

  def git(repository, *arguments)
    _output, error, status = Open3.capture3("git", "-C", repository, *arguments)
    raise error unless status.success?
  end
end
