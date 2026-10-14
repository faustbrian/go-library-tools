# frozen_string_literal: true

require "pathname"

module IdentityPlatformRepositoryTopology
  class InvalidTopology < StandardError; end

  Entry = Struct.new(
    :family, :repository_root, :module_prefix, :repository_identity,
    keyword_init: true
  )

  Resolution = Struct.new(
    :logical_module, :family, :repository_root, :repository_identity,
    :module_directory, :module_path,
    keyword_init: true
  )

  class Map
    attr_reader :coordination_root, :control_repository_root, :entries

    def initialize(entries:, coordination_root:, control_repository_root:)
      @coordination_root = canonical_absolute_path(coordination_root, "coordination root")
      @control_repository_root = canonical_absolute_path(control_repository_root, "control repository root")
      @resolved_coordination_root = resolve_projected_path(@coordination_root, "coordination root")
      @resolved_control_repository_root = resolve_projected_path(@control_repository_root, "control repository root")
      @entries = validate_entries(entries).freeze
    end

    def resolve(logical_module)
      match = logical_module.to_s.match(%r{\Apkg/([a-z0-9]+(?:-[a-z0-9]+)*)(?:/([a-z0-9]+(?:[._-][a-z0-9]+)*(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*))?\z})
      raise InvalidTopology, "invalid logical module #{logical_module.inspect}" unless match

      family = match[1]
      entry = entries.fetch(family) do
        raise InvalidTopology, "unknown logical family #{family}"
      end
      remainder = match[2]
      module_directory = remainder || "."
      module_path = remainder ? "#{entry.module_prefix}/#{remainder}" : entry.module_prefix

      Resolution.new(
        logical_module: logical_module,
        family: family,
        repository_root: entry.repository_root,
        repository_identity: entry.repository_identity,
        module_directory: module_directory,
        module_path: module_path
      ).freeze
    end

    private

    def validate_entries(source_entries)
      raise InvalidTopology, "repository mapping is empty" if source_entries.empty?

      by_family = {}
      identities = {}
      roots = {}
      source_entries.each do |source_entry|
        family = source_entry.fetch(:family)
        repository_root = canonical_absolute_path(source_entry.fetch(:repository_root), "repository root for #{family}")
        module_prefix = source_entry.fetch(:module_prefix)
        expected_repository_name = "go-#{family}"
        repository_name = File.basename(repository_root)
        expected_prefix = "github.com/faustbrian/#{repository_name}"

        raise InvalidTopology, "duplicate logical family #{family}" if by_family.key?(family)
        unless descendant?(repository_root, coordination_root)
          raise InvalidTopology, "repository root for #{family} escapes coordination root"
        end
        if repository_root == control_repository_root
          raise InvalidTopology, "control repository cannot own package family #{family}"
        end
        resolved_repository_root = resolve_projected_path(repository_root, "repository root for #{family}")
        unless descendant?(resolved_repository_root, @resolved_coordination_root)
          raise InvalidTopology, "repository root for #{family} resolves outside coordination root"
        end
        if resolved_repository_root == @resolved_control_repository_root ||
           descendant?(resolved_repository_root, @resolved_control_repository_root)
          raise InvalidTopology, "repository root for #{family} resolves within control repository"
        end
        unless repository_name == expected_repository_name
          raise InvalidTopology, "repository name for #{family} must be #{expected_repository_name}"
        end
        if identities.key?(module_prefix)
          raise InvalidTopology, "duplicate repository identity #{module_prefix}"
        end
        if roots.key?(repository_root)
          raise InvalidTopology, "duplicate package repository root #{repository_root}"
        end
        unless module_prefix == expected_prefix
          raise InvalidTopology, "invalid module prefix for #{family}: expected #{expected_prefix}"
        end

        entry = Entry.new(
          family: family,
          repository_root: repository_root,
          module_prefix: module_prefix,
          repository_identity: module_prefix
        ).freeze
        by_family[family] = entry
        identities[module_prefix] = family
        roots[repository_root] = family
      end
      by_family
    end

    def descendant?(path, root)
      path.start_with?("#{root}#{File::SEPARATOR}")
    end

    def canonical_absolute_path(value, label)
      value = value.to_s
      path = Pathname.new(value)
      raise InvalidTopology, "#{label} must be absolute" unless path.absolute?

      clean = path.cleanpath.to_s
      raise InvalidTopology, "#{label} is not canonical" unless clean == value

      clean
    end

    def resolve_projected_path(path, label)
      unresolved = []
      existing = path
      until File.exist?(existing) || File.symlink?(existing)
        parent = File.dirname(existing)
        raise InvalidTopology, "#{label} has no resolvable ancestor" if parent == existing

        unresolved.unshift(File.basename(existing))
        existing = parent
      end

      resolved_existing = File.realpath(existing)
      unless File.directory?(resolved_existing)
        raise InvalidTopology, "#{label} does not resolve through a directory"
      end

      File.join(resolved_existing, *unresolved)
    rescue Errno::ENOENT, Errno::EACCES, Errno::ELOOP => e
      raise InvalidTopology, "#{label} cannot be resolved: #{e.class}"
    end
  end

  module_function

  def load(path, coordination_root:, control_repository_root:)
    parse(
      File.binread(path),
      coordination_root: coordination_root,
      control_repository_root: control_repository_root
    )
  end

  def load_declared(path)
    parse_declared(File.binread(path))
  end

  def parse_declared(source)
    roots = source.scan(%r{`(/[^`\n]+/go-library-tools)/\.ai/identity-platform`}).flatten.uniq
    raise InvalidTopology, "control repository declaration is missing" if roots.empty?
    raise InvalidTopology, "control repository declaration is ambiguous" unless roots.one?

    control_repository_root = roots.first
    parse(
      source,
      coordination_root: File.dirname(control_repository_root),
      control_repository_root: control_repository_root
    )
  end

  def parse(source, coordination_root:, control_repository_root:)
    lines = source.lines
    header = "| Logical family | Standalone repository | Module prefix |\n"
    header_indexes = lines.each_index.select { |index| lines[index] == header }
    raise InvalidTopology, "mapping table is missing" if header_indexes.empty?
    raise InvalidTopology, "mapping table is ambiguous" unless header_indexes.one?

    header_index = header_indexes.first
    unless lines[header_index + 1]&.match?(%r{\A\|\s*---\s*\|\s*---\s*\|\s*---\s*\|\s*\z})
      raise InvalidTopology, "mapping table delimiter is malformed"
    end

    table_rows = lines.drop(header_index + 2).take_while { |line| line.start_with?("|") }
    table_end_index = header_index + 2 + table_rows.length
    outside_lines = lines.each_with_index.reject do |_line, index|
      index.between?(header_index, table_end_index - 1)
    end.map(&:first)
    outside_mapping_line = outside_lines.any? { |line| line.count("|") >= 2 }
    if outside_mapping_line
      raise InvalidTopology, "mapping row appears outside canonical table"
    end

    rows = []
    table_rows.each do |line|
      match = line.match(%r{\A\| `([a-z0-9]+(?:-[a-z0-9]+)*)` \| `([^`]+)` \| `(github\.com/[^`]+)` \|\s*\z})
      raise InvalidTopology, "malformed mapping row #{line.strip.inspect}" unless match

      rows << {family: match[1], repository_root: match[2], module_prefix: match[3]}
    end

    Map.new(
      entries: rows,
      coordination_root: coordination_root,
      control_repository_root: control_repository_root
    )
  end
end
