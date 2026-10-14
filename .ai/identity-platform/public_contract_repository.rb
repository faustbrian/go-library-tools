# frozen_string_literal: true

require "pathname"
require "open3"
require "digest"

require_relative "repository_input_manifest"

module IdentityPlatformPublicContractRepository
  class InvalidRepository < StandardError; end

  Resolution = Struct.new(
    :alias_name, :repository_root, :repository_identity, :package_directory,
    :package_path, :source_revision,
    keyword_init: true
  )

  PublicSourceFiles = Struct.new(:go_mod, :production_sources, :api_baselines, :contents, keyword_init: true) do
    def read(path)
      contents.fetch(path)
    end
  end

  class Resolver
    def initialize(entries:, forbidden_repository_roots:, revisions: {}, integration_baselines: {})
      unless revisions.is_a?(Hash) && revisions.all? do |identity, revision|
        identity.is_a?(String) && revision.is_a?(String) && revision.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/)
      end
        raise InvalidRepository, "public source revisions must name exact commits"
      end
      @revisions = revisions.to_h { |identity, revision| [identity.dup.freeze, revision.dup.freeze] }.freeze
      unless integration_baselines.is_a?(Hash) && integration_baselines.all? do |identity, revision|
        identity.is_a?(String) && revision.is_a?(String) && revision.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/)
      end
        raise InvalidRepository, "integration baselines must name exact commits"
      end
      unless (@revisions.keys & integration_baselines.keys).empty?
        raise InvalidRepository, "integration lanes cannot also use primitive revision overrides"
      end
      @integration_baselines = integration_baselines.to_h { |identity, revision| [identity.dup.freeze, revision.dup.freeze] }.freeze
      @captured_revisions = {}
      @captured_roots = {}
      @forbidden_repository_roots = forbidden_repository_roots.map do |root|
        canonical_real_directory(root, "forbidden repository root")
      end.freeze
      @entries = validate_entries(entries).freeze
    end

    def resolve(alias_name)
      @entries.fetch(alias_name.to_s) do
        raise InvalidRepository, "unknown public-contract package alias #{alias_name.inspect}"
      end
    end

    def public_source_files(alias_name)
      resolution = resolve(alias_name)
      entries = committed_entries(resolution)
      module_directory = resolution.package_directory
      module_directory = File.dirname(module_directory) until
        entries.key?(File.join(module_directory, "go.mod")) || module_directory == resolution.repository_root
      go_mod = File.join(module_directory, "go.mod")
      raise InvalidRepository, "external package #{alias_name} has no owning go.mod" unless entries.key?(go_mod)

      production_sources = entries.keys.select do |path|
        File.dirname(path) == resolution.package_directory && path.end_with?(".go") && !path.end_with?("_test.go")
      end.sort
      raise InvalidRepository, "package directory for #{alias_name} has no production Go source" if production_sources.empty?
      api_baselines = entries.keys.select { |path| path.start_with?("#{resolution.package_directory}/api/") }
        .sort
      selected = [go_mod, *production_sources, *api_baselines]
      raise InvalidRepository, "public source file count exceeds limit" if selected.length > 4096
      total = 0
      contents = selected.to_h do |path|
        mode, object, size = entries.fetch(path)
        label = path == go_mod ? "go.mod" : (api_baselines.include?(path) ? "API baseline" : "production source")
        raise InvalidRepository, "#{label} is a symbolic link" if mode == "120000"
        raise InvalidRepository, "#{label} is not a regular file" unless %w[100644 100755].include?(mode)
        total += size
        raise InvalidRepository, "public source byte limit exceeded" if size > 4 * 1024 * 1024 || total > 64 * 1024 * 1024
        bytes = git_bytes(resolution.repository_root, "cat-file", "blob", object, limit: size)
        raise InvalidRepository, "committed source size drifted" unless bytes.bytesize == size
        [path, bytes.freeze]
      end

      PublicSourceFiles.new(
        go_mod: go_mod,
        production_sources: production_sources.freeze,
        api_baselines: api_baselines.freeze,
        contents: contents.freeze
      ).freeze
    end

    private

    def validate_entries(source_entries)
      pairs = source_entries.respond_to?(:to_a) ? source_entries.to_a : []
      raise InvalidRepository, "public-contract repository mapping is empty" if pairs.empty?

      aliases = {}
      pairs.each do |alias_name, source|
        unless alias_name.to_s.match?(%r{\A[a-z][a-z0-9]*\z})
          raise InvalidRepository, "invalid public-contract package alias #{alias_name.inspect}"
        end
        raise InvalidRepository, "duplicate public-contract package alias #{alias_name}" if aliases.key?(alias_name)
        unless source.is_a?(Hash) && source.keys.sort == %i[package_directory package_path repository_identity repository_root]
          raise InvalidRepository, "invalid repository declaration for #{alias_name}"
        end

        aliases[alias_name] = resolve_entry(alias_name, source)
      end
      aliases
    end

    def resolve_entry(alias_name, source)
      identity = source.fetch(:repository_identity).to_s
      root = IdentityPlatformRepositoryInputManifest.verified_repository_root(
        source.fetch(:repository_root), identity
      )
      if @forbidden_repository_roots.any? { |forbidden| root == forbidden || descendant?(root, forbidden) }
        raise InvalidRepository, "repository for #{alias_name} resolves within a control repository"
      end

      relative = canonical_relative_path(source.fetch(:package_directory), alias_name)
      projected = relative == "." ? root : File.join(root, relative)
      directory = projected
      unless directory == root || descendant?(directory, root)
        raise InvalidRepository, "package directory for #{alias_name} escapes its standalone repository"
      end

      package_path = source.fetch(:package_path).to_s
      expected_package_path = relative == "." ? identity : "#{identity}/#{relative}"
      unless package_path == expected_package_path
        raise InvalidRepository,
          "package path for #{alias_name} must be #{expected_package_path}"
      end
      if @captured_roots.key?(identity) && @captured_roots.fetch(identity) != root
        raise InvalidRepository, "public source aliases disagree on their repository root"
      end
      @captured_roots[identity] = root
      revision = @captured_revisions[identity] ||= capture_revision(root, identity)

      resolution = Resolution.new(
        alias_name: alias_name,
        repository_root: root,
        repository_identity: identity,
        package_directory: directory,
        package_path: package_path,
        source_revision: revision
      ).freeze
      unless committed_entries(resolution).keys.any? do |path|
        File.dirname(path) == directory && path.end_with?(".go") && !path.end_with?("_test.go")
      end
        raise InvalidRepository, "package directory for #{alias_name} escapes or has no committed production source"
      end
      resolution
    rescue IdentityPlatformRepositoryInputManifest::InvalidManifest => e
      raise InvalidRepository, e.message
    end

    def canonical_relative_path(value, alias_name)
      path = value.to_s
      pathname = Pathname.new(path)
      unless !pathname.absolute? && pathname.cleanpath.to_s == path &&
          (path == "." || !path.start_with?("../"))
        raise InvalidRepository, "package directory for #{alias_name} escapes or is noncanonical"
      end
      path
    end

    def canonical_real_directory(value, label)
      path = value.to_s
      pathname = Pathname.new(path)
      raise InvalidRepository, "#{label} must be absolute" unless pathname.absolute?
      raise InvalidRepository, "#{label} is noncanonical" unless pathname.cleanpath.to_s == path

      resolved = File.realpath(path)
      raise InvalidRepository, "#{label} is not a directory" unless File.directory?(resolved)

      resolved
    rescue Errno::ENOENT, Errno::EACCES, Errno::ELOOP => e
      raise InvalidRepository, "#{label} cannot be resolved: #{e.class}"
    end

    def capture_revision(root, identity)
      requested = @revisions.fetch(identity, "HEAD")
      unless requested == "HEAD" || requested.to_s.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/)
        raise InvalidRepository, "public source revision must be an exact commit"
      end
      revision = git_bytes(root, "rev-parse", "--verify", "--end-of-options", "#{requested}^{commit}", limit: 128).strip
      unless revision.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/) && (requested == "HEAD" || requested == revision)
        raise InvalidRepository, "public source revision is not an exact commit"
      end
      if (baseline = @integration_baselines[identity])
        branch = git_bytes(root, "rev-parse", "--abbrev-ref", "HEAD", limit: 128).strip
        raise InvalidRepository, "unexpected actual integration branch" unless branch == "feature/identity-platform"
        begin
          ancestor = git_bytes(root, "merge-base", baseline, revision, limit: 128).strip
        rescue InvalidRepository
          raise InvalidRepository, "integration baseline ancestry is unavailable"
        end
        raise InvalidRepository, "HEAD does not descend from integration baseline" unless ancestor == baseline
        current = git_bytes(root, "rev-parse", "--verify", "HEAD", limit: 128).strip
        raise InvalidRepository, "integration HEAD changed during capture" unless current == revision
      end
      revision.freeze
    end

    def committed_entries(resolution)
      output = git_bytes(resolution.repository_root, "ls-tree", "-r", "-l", "-z",
                         resolution.source_revision, limit: 4 * 1024 * 1024)
      output.split("\0").to_h do |record|
        metadata, relative = record.split("\t", 2)
        mode, kind, object, size = metadata.to_s.split
        unless relative && %w[blob commit].include?(kind) && object.to_s.match?(/\A[0-9a-f]{40}(?:[0-9a-f]{24})?\z/)
          raise InvalidRepository, "invalid committed source inventory"
        end
        relative = relative.force_encoding(Encoding::UTF_8)
        unless relative.valid_encoding? && !relative.include?("\n") && Pathname.new(relative).cleanpath.to_s == relative &&
            !relative.start_with?("/", "../")
          raise InvalidRepository, "invalid committed source path"
        end
        [File.join(resolution.repository_root, relative), [mode, object, kind == "blob" ? Integer(size, 10) : 0]]
      end
    end

    def git_bytes(root, *arguments, limit:)
      # Replacement refs and legacy grafts must not rewrite the source objects
      # or parent graph named by the recorded repository authority.
      Open3.popen3({"GIT_GRAFT_FILE" => File::NULL}, "git", "--no-replace-objects",
                   "-c", "advice.graftFileDeprecated=false", "-C", root, *arguments) do |input, output, error, wait|
        input.close
        bytes = output.read(limit + 1).to_s.b
        output.close
        diagnostic = error.read(4097).to_s
        error.close
        unless wait.value.success? && diagnostic.empty? && bytes.bytesize <= limit
          raise InvalidRepository, "committed public source is unavailable or exceeds its limit"
        end
        bytes
      end
    end

    def descendant?(path, root)
      path.start_with?("#{root}#{File::SEPARATOR}")
    end
  end

  # Owns the generator's source-selection policy and the bytes used for both
  # symbol discovery and API pins. Planning metadata remains the authority;
  # this reader does not create a second snapshot or evidence registry.
  class Catalog
    def initialize(entries:, forbidden_repository_roots:, primitive_records:, integration_source:, source_revisions: {})
      unless source_revisions.is_a?(Hash) && source_revisions.all? do |identity, revision|
        identity.is_a?(String) && revision.is_a?(String) && revision.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/)
      end
        raise InvalidRepository, "catalog source revisions must name exact commits"
      end
      revisions = primitive_records.group_by { |row| row.fetch("repository_identity") }.to_h do |identity, records|
        candidates = records.map { |row| row.fetch("repository_revision") }.uniq
        raise InvalidRepository, "#{identity} has conflicting primitive source revisions" unless candidates.one?
        [identity, candidates.first]
      end
      section = integration_source[/^## Repository integration identities\n(.*?)(?=^## |\z)/m, 1]
      raise InvalidRepository, "repository integration identities are missing" unless section
      lanes = {}
      section.each_line do |line|
        cells = line.split("|", -1).map(&:strip)
        next unless cells[1]&.start_with?("github.com/")
        raise InvalidRepository, "invalid repository integration identity row" unless cells.length == 12
        raise InvalidRepository, "unexpected recorded integration branch" unless cells[8] == "feature/identity-platform"
        raise InvalidRepository, "duplicate repository integration identity" if lanes.key?(cells[1])
        lanes[cells[1]] = {root: cells[9], baseline: cells[5]}
      end
      selected = entries.to_h do |alias_name, declaration|
        identity = declaration.fetch(:repository_identity)
        [alias_name, lanes.key?(identity) ? declaration.merge(repository_root: lanes.fetch(identity).fetch(:root)) : declaration]
      end
      identities = selected.values.map { |declaration| declaration.fetch(:repository_identity) }.uniq
      baselines = lanes.select { |identity, _| identities.include?(identity) }.transform_values { |lane| lane.fetch(:baseline) }
      unmapped = source_revisions.keys - identities
      raise InvalidRepository, "unmapped catalog source revision: #{unmapped.sort.join(', ')}" unless unmapped.empty?
      overlap = source_revisions.keys & (revisions.keys | baselines.keys)
      raise InvalidRepository, "catalog source revision already recorded: #{overlap.sort.join(', ')}" unless overlap.empty?
      # Catalog-only authorities have no primitive-gate observation. Their
      # explicit source pins supplement, never override, existing authorities.
      revisions.merge!(source_revisions)
      missing = identities - baselines.keys - revisions.keys
      raise InvalidRepository, "missing recorded primitive revision for #{missing.sort.join(', ')}" unless missing.empty?
      @resolver = Resolver.new(
        entries: selected, forbidden_repository_roots: forbidden_repository_roots,
        revisions: revisions.reject { |identity, _| baselines.key?(identity) },
        integration_baselines: baselines
      )
    end

    def resolve(alias_name)
      @resolver.resolve(alias_name)
    end

    def contract(alias_name, logical_path)
      resolution = resolve(alias_name)
      files = @resolver.public_source_files(alias_name)
      repository = resolution.repository_root
      relative = Pathname.new(resolution.package_directory).relative_path_from(Pathname.new(repository)).to_s
      logical_repository = relative == "." ? logical_path : logical_path.delete_suffix("/#{relative}")
      labeled = [files.go_mod, *files.production_sources, *files.api_baselines].uniq.map do |path|
        [File.join(logical_repository, Pathname.new(path).relative_path_from(Pathname.new(repository)).to_s), path]
      end.sort_by(&:first)
      bytes = labeled.map do |label, path|
        contents = files.read(path)
        "#{label}\0#{contents.bytesize}\0#{contents}\0"
      end.join
      symbols = files.production_sources.flat_map do |path|
        files.read(path).scan(/^\s*(?:type|func|var|const)\s+([A-Z][A-Za-z0-9]*)\b/).flatten
      end.uniq.map { |symbol| "#{alias_name}.#{symbol}" }
      {
        "exported_symbols" => symbols,
        "api_baseline_files" => labeled.map(&:first),
        "api_baseline_sha256" => "sha256:#{Digest::SHA256.hexdigest(bytes)}"
      }
    end
  end

  module_function

  def new(...)
    Resolver.new(...)
  end
end
