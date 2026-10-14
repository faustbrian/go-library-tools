# frozen_string_literal: true

require "json"
require "pathname"
require "set"

require_relative "repository_input_manifest"

module IdentityPlatformStandalonePackageRegistry
  class InvalidRegistry < StandardError; end

  REPOSITORY_NAMES = {
    "outbox" => "go-transactional-outbox"
  }.freeze

  class Registry
    def initialize(coordination_root:, control_repository_root:)
      @coordination_root = canonical_real_directory(coordination_root, "coordination root")
      @control_repository_root = canonical_real_directory(control_repository_root, "control repository root")
    end

    def registered_consumables(requested_consumables)
      requested = validate_requested_consumables(requested_consumables)
      registered = requested.map { |name| name.split("/", 2).first }.uniq.flat_map do |family|
        repository_consumables(family)
      end.to_set
      missing = requested.reject { |name| registered.include?(name) }
      unless missing.empty?
        raise InvalidRegistry, "#{missing.min_by(&:b)} is not registered in its standalone repository catalogs"
      end

      registered.freeze
    end

    def repository_root_for(consumable)
      requested = validate_requested_consumables([consumable]).fetch(0)
      resolve_repository(requested.split("/", 2).first).fetch(:root)
    end

    private

    def validate_requested_consumables(source)
      unless source.respond_to?(:to_a)
        raise InvalidRegistry, "requested consumables must be enumerable"
      end

      requested = source.to_a.map(&:to_s)
      unless requested.all? { |name| name.match?(%r{\A[a-z0-9]+(?:-[a-z0-9]+)*(?:/[a-z0-9]+(?:-[a-z0-9]+)*)*\z}) }
        raise InvalidRegistry, "requested consumable name is invalid"
      end
      requested.uniq
    end

    def repository_consumables(family)
      repository = resolve_repository(family)
      repository_root = repository.fetch(:root)
      repository_identity = repository.fetch(:identity)

      modules = read_catalog(repository_root, "modules.json", "modules", repository_identity)
      packages = read_catalog(repository_root, "packages.json", "packages", repository_identity)
      module_paths = modules.map { |entry| required_path(entry, "module_path", "modules.json") }
      package_paths = packages.map { |entry| required_path(entry, "import_path", "packages.json") }
      (module_paths + package_paths).map do |path|
        logical_consumable(path, family, repository_identity)
      end
    rescue IdentityPlatformRepositoryInputManifest::InvalidManifest => e
      raise InvalidRegistry, e.message
    end


    def resolve_repository(family)
      repository_name = REPOSITORY_NAMES.fetch(family, "go-#{family}")
      repository_identity = "github.com/faustbrian/#{repository_name}"
      projected_root = File.join(@coordination_root, repository_name)
      repository_root = IdentityPlatformRepositoryInputManifest.verified_repository_root(
        projected_root, repository_identity
      )
      if repository_root == @control_repository_root || descendant?(repository_root, @control_repository_root)
        raise InvalidRegistry, "standalone package repository resolves within the control repository"
      end

      {root: repository_root, identity: repository_identity}
    rescue IdentityPlatformRepositoryInputManifest::InvalidManifest => e
      raise InvalidRegistry, e.message
    end

    def read_catalog(repository_root, filename, collection, repository_identity)
      path = File.join(repository_root, filename)
      metadata = File.lstat(path)
      raise InvalidRegistry, "#{filename} is a symbolic link" if metadata.symlink?
      raise InvalidRegistry, "#{filename} is not a regular file" unless metadata.file?
      unless File.realpath(path).start_with?("#{repository_root}#{File::SEPARATOR}")
        raise InvalidRegistry, "#{filename} resolves outside its standalone repository"
      end

      document = JSON.parse(File.binread(path))
      unless document.is_a?(Hash) && document.fetch("repository", nil) == repository_identity
        raise InvalidRegistry, "#{filename.delete_suffix('.json')} manifest repository drifted"
      end
      entries = document.fetch(collection, nil)
      unless entries.is_a?(Array)
        raise InvalidRegistry, "#{filename} #{collection} catalog is invalid"
      end
      entries
    rescue Errno::ENOENT, Errno::EACCES, Errno::ELOOP => e
      raise InvalidRegistry, "#{filename} cannot be resolved: #{e.class}"
    rescue JSON::ParserError
      raise InvalidRegistry, "#{filename} is not valid JSON"
    end

    def required_path(entry, key, filename)
      value = entry.is_a?(Hash) ? entry[key] : nil
      unless value.is_a?(String) && !value.empty?
        raise InvalidRegistry, "#{filename} contains an invalid #{key}"
      end
      value
    end

    def logical_consumable(path, family, repository_identity)
      unless path == repository_identity || path.start_with?("#{repository_identity}/")
        raise InvalidRegistry, "catalog path #{path.inspect} escapes #{repository_identity}"
      end
      suffix = path.delete_prefix(repository_identity)
      logical = "#{family}#{suffix}"
      unless logical.match?(%r{\A[a-z0-9]+(?:-[a-z0-9]+)*(?:/[A-Za-z0-9._-]+)*\z})
        raise InvalidRegistry, "catalog path #{path.inspect} is not a canonical Go import path"
      end
      logical
    end

    def canonical_real_directory(value, label)
      path = value.to_s
      pathname = Pathname.new(path)
      raise InvalidRegistry, "#{label} must be absolute" unless pathname.absolute?
      raise InvalidRegistry, "#{label} is noncanonical" unless pathname.cleanpath.to_s == path

      resolved = File.realpath(path)
      raise InvalidRegistry, "#{label} is not a directory" unless File.directory?(resolved)

      resolved
    rescue Errno::ENOENT, Errno::EACCES, Errno::ELOOP => e
      raise InvalidRegistry, "#{label} cannot be resolved: #{e.class}"
    end

    def descendant?(path, root)
      path.start_with?("#{root}#{File::SEPARATOR}")
    end
  end

  module_function

  def new(...)
    Registry.new(...)
  end
end
