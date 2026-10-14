# frozen_string_literal: true

require "digest"
require "json"
require "open3"
require "pathname"
require "uri"

module IdentityPlatformRepositoryInputManifest
  class InvalidManifest < StandardError; end

  REPOSITORY_REQUEST_KEYS = %i[input_roots repository_identity repository_root revision].freeze
  INPUT_ROOT_KEYS = %i[owner path reason].freeze
  ENVIRONMENT_INPUT_KEYS = %i[content_identity environment_id owner reason].freeze
  DOCUMENT_KEYS = %w[environment_inputs repositories schema_version].freeze
  SNAPSHOT_KEYS = %w[inputs object_format repository_identity revision tree].freeze
  TRACKED_INPUT_KEYS = %w[content_identity kind mode owner path reason].freeze
  CANONICAL_ENVIRONMENT_KEYS = %w[content_identity environment_id owner reason].freeze

  module_function

  def build(repositories:, environment_inputs:)
    requests = validate_repository_requests(repositories)
    snapshots = requests.map { |request| build_snapshot(request) }
      .sort_by { |snapshot| snapshot.fetch("repository_identity").b }
    environments = validate_environment_inputs(environment_inputs)
    document = canonical_value(
      "schema_version" => 2,
      "repositories" => snapshots,
      "environment_inputs" => environments
    )
    validate_document!(document)
    deep_freeze(document)
  end

  def canonical_json(document)
    validate_document!(document)
    JSON.generate(canonical_value(document))
  end

  def fingerprint(document)
    validate_document!(document)
    projection = fingerprint_projection(document)
    bytes = "identity-platform-input-manifest-v2\n#{JSON.generate(canonical_value(projection))}"
    "sha256:#{Digest::SHA256.hexdigest(bytes)}"
  end

  def rebuild(document, repository_roots:, revision_overrides: {})
    validate_document!(document)
    unless repository_roots.is_a?(Hash) && repository_roots.keys.all? { |key| key.is_a?(String) } &&
      revision_overrides.is_a?(Hash) && revision_overrides.keys.all? { |key| key.is_a?(String) }
      raise InvalidManifest, "repository rebuild identities are invalid"
    end
    identities = document.fetch("repositories").map { |snapshot| snapshot.fetch("repository_identity") }
    unless repository_roots.keys.sort_by(&:b) == identities.sort_by(&:b)
      raise InvalidManifest, "repository rebuild roots do not close the manifest repository set"
    end
    unless (revision_overrides.keys - identities).empty?
      raise InvalidManifest, "repository rebuild revision overrides contain an unknown identity"
    end

    repositories = document.fetch("repositories").map do |snapshot|
      identity = snapshot.fetch("repository_identity")
      {
        repository_root: repository_roots.fetch(identity),
        repository_identity: identity,
        revision: revision_overrides.fetch(identity, snapshot.fetch("revision")),
        input_roots: snapshot.fetch("inputs").map do |input|
          {path: input.fetch("path"), owner: input.fetch("owner"), reason: input.fetch("reason")}
        end
      }
    end
    environment_inputs = document.fetch("environment_inputs").map do |input|
      {
        environment_id: input.fetch("environment_id"),
        content_identity: input.fetch("content_identity"),
        owner: input.fetch("owner"),
        reason: input.fetch("reason")
      }
    end
    build(repositories: repositories, environment_inputs: environment_inputs)
  end

  def verified_repository_root(repository_root, repository_identity)
    identity = required_string(repository_identity, "repository identity")
    raise InvalidManifest, "repository identity #{identity.inspect} is invalid" unless valid_repository_identity?(identity)

    root = resolve_repository_root(repository_root, identity)
    verify_repository_identity!(root, identity)
    root
  end

  def validate_repository_requests(repositories)
    unless repositories.is_a?(Array) && repositories.any?
      raise InvalidManifest, "repository snapshots are absent"
    end

    identities = {}
    roots = {}
    repositories.map.with_index do |request, index|
      unless exact_keys?(request, REPOSITORY_REQUEST_KEYS)
        raise InvalidManifest, "repository request #{index} schema is invalid"
      end

      identity = required_string(request.fetch(:repository_identity), "repository identity")
      raise InvalidManifest, "repository identity #{identity.inspect} is invalid" unless valid_repository_identity?(identity)
      raise InvalidManifest, "duplicate repository identity #{identity}" if identities.key?(identity)

      root = resolve_repository_root(request.fetch(:repository_root), identity)
      raise InvalidManifest, "duplicate repository root #{root}" if roots.key?(root)

      input_roots = validate_input_roots(request.fetch(:input_roots), identity)
      identities[identity] = true
      roots[root] = true
      {
        repository_identity: identity,
        repository_root: root,
        revision: request.fetch(:revision),
        input_roots: input_roots
      }
    end
  end
  private_class_method :validate_repository_requests

  def resolve_repository_root(value, identity)
    path = value.to_s
    unless Pathname.new(path).absolute?
      raise InvalidManifest, "repository root for #{identity} must be absolute"
    end

    resolved = File.realpath(path)
    unless File.directory?(resolved)
      raise InvalidManifest, "repository root for #{identity} is not a directory"
    end

    top_level, _error, status = Open3.capture3("git", "-C", resolved, "rev-parse", "--show-toplevel")
    unless status.success? && File.realpath(top_level.strip) == resolved
      raise InvalidManifest, "repository root for #{identity} is not an exact Git worktree root"
    end

    resolved
  rescue Errno::ENOENT, Errno::EACCES, Errno::ELOOP
    raise InvalidManifest, "repository root for #{identity} does not exist or cannot be resolved"
  end
  private_class_method :resolve_repository_root

  def validate_input_roots(input_roots, identity)
    unless input_roots.is_a?(Array) && input_roots.any?
      raise InvalidManifest, "input roots for #{identity} are absent"
    end

    input_roots.map.with_index do |input_root, index|
      unless exact_keys?(input_root, INPUT_ROOT_KEYS)
        raise InvalidManifest, "input root #{index} for #{identity} schema is invalid"
      end

      path = required_string(input_root.fetch(:path), "input root path")
      unless valid_relative_path?(path, allow_dot: true)
        raise InvalidManifest, "input root #{path.inspect} for #{identity} is absolute, traversing, or noncanonical"
      end

      {
        path: path,
        owner: required_string(input_root.fetch(:owner), "input root owner"),
        reason: required_string(input_root.fetch(:reason), "input root reason")
      }
    end
  end
  private_class_method :validate_input_roots

  def build_snapshot(request)
    root = request.fetch(:repository_root)
    identity = request.fetch(:repository_identity)
    object_format = git_output!(root, "rev-parse", "--show-object-format", label: "object format")
    unless %w[sha1 sha256].include?(object_format)
      raise InvalidManifest, "repository #{identity} has unsupported Git object format #{object_format.inspect}"
    end
    semantic_object_format = "git-#{object_format}"
    revision = exact_revision(root, request.fetch(:revision), object_format, identity)
    verify_repository_identity!(root, identity)
    tree = git_output!(root, "rev-parse", "--verify", "--end-of-options", "#{revision}^{tree}", label: "tree")
    inputs = tracked_inputs(root, revision, semantic_object_format, request.fetch(:input_roots), identity)

    canonical_value(
      "repository_identity" => identity,
      "object_format" => semantic_object_format,
      "revision" => revision,
      "tree" => "#{semantic_object_format}:#{tree}",
      "inputs" => inputs
    )
  end
  private_class_method :build_snapshot

  def exact_revision(root, value, object_format, identity)
    revision = value.to_s
    expected_length = object_format == "sha256" ? 64 : 40
    unless revision.match?(%r{\A[0-9a-f]{#{expected_length}}\z})
      raise InvalidManifest, "revision for #{identity} is missing, ambiguous, or not a full #{object_format} object ID"
    end

    resolved, error, status = Open3.capture3(
      "git", "-C", root, "rev-parse", "--verify", "--end-of-options", "#{revision}^{commit}"
    )
    unless status.success? && error.empty? && resolved.strip == revision
      raise InvalidManifest, "revision #{revision} for #{identity} is not an exact commit"
    end

    revision
  end
  private_class_method :exact_revision

  def verify_repository_identity!(root, expected)
    output, _error, status = Open3.capture3("git", "-C", root, "config", "--get-all", "remote.origin.url")
    remotes = status.success? ? output.lines.map(&:strip).reject(&:empty?) : []
    actual = remotes.one? ? repository_identity_from_remote(remotes.first) : nil
    return if actual == expected

    raise InvalidManifest,
      "repository identity substitution for #{expected}: origin identifies #{actual || "no unique canonical repository"}"
  end
  private_class_method :verify_repository_identity!

  def repository_identity_from_remote(remote)
    if (match = remote.match(%r{\A(?:[^@/]+@)?([^:/]+):([^/].*)\z})) && !remote.include?("://")
      host = match[1]
      path = match[2]
    else
      uri = URI.parse(remote)
      return unless %w[https ssh].include?(uri.scheme) && uri.host && !uri.query && !uri.fragment
      return if uri.port && uri.default_port != uri.port

      host = uri.host
      path = uri.path.sub(%r{\A/}, "")
    end
    path = path.sub(/\.git\z/, "").sub(%r{/\z}, "")
    identity = "#{host.downcase}/#{path}"
    valid_repository_identity?(identity) ? identity : nil
  rescue URI::InvalidURIError
    nil
  end
  private_class_method :repository_identity_from_remote

  def valid_repository_identity?(identity)
    identity.match?(%r{\A[a-z0-9.-]+/[A-Za-z0-9._-]+(?:/[A-Za-z0-9._-]+)+\z}) &&
      !identity.split("/").include?("..")
  end
  private_class_method :valid_repository_identity?

  def tracked_inputs(root, revision, object_format, input_roots, identity)
    inputs = input_roots.flat_map do |input_root|
      output = git_output!(
        root, "--literal-pathspecs", "ls-tree", "-r", "--full-tree", "-z", revision, "--",
        input_root.fetch(:path),
        label: "tracked inputs"
      )
      entries = output.split("\0", -1).reject(&:empty?).map do |record|
        match = record.match(%r{\A[0-7]{6} (\S+) ([0-9a-f]+)\t(.+)\z}m)
        unless match
          raise InvalidManifest, "repository #{identity} returned a malformed tracked input"
        end

        canonical_value(
          "path" => match[3],
          "kind" => match[1],
          "mode" => record[0, 6],
          "content_identity" => "#{object_format}:#{match[2]}",
          "owner" => input_root.fetch(:owner),
          "reason" => input_root.fetch(:reason)
        )
      end
      if entries.empty?
        raise InvalidManifest, "input root #{input_root.fetch(:path).inspect} for #{identity} has no tracked inputs"
      end
      entries
    end

    duplicates = inputs.group_by { |input| input.fetch("path") }.select { |_path, entries| entries.length > 1 }
    unless duplicates.empty?
      raise InvalidManifest, "duplicate input identity #{duplicates.keys.min_by(&:b)} in repository #{identity}"
    end

    inputs.sort_by { |input| input.fetch("path").b }
  end
  private_class_method :tracked_inputs

  def validate_environment_inputs(environment_inputs)
    unless environment_inputs.is_a?(Array)
      raise InvalidManifest, "environment inputs must be an array"
    end

    identities = {}
    environment_inputs.map.with_index do |input, index|
      unless exact_keys?(input, ENVIRONMENT_INPUT_KEYS)
        raise InvalidManifest, "environment input #{index} schema is invalid"
      end

      identity = required_string(input.fetch(:environment_id), "environment input identity")
      unless identity.match?(%r{\Aenvironment:[a-z0-9]+(?:[._-][a-z0-9]+)*\z})
        raise InvalidManifest, "environment input identity #{identity.inspect} is invalid"
      end
      raise InvalidManifest, "duplicate environment input identity #{identity}" if identities.key?(identity)

      content_identity = required_string(input.fetch(:content_identity), "environment input content identity")
      unless content_identity.match?(%r{\Asha256:[0-9a-f]{64}\z})
        raise InvalidManifest, "environment input #{identity} content identity is invalid"
      end
      identities[identity] = true
      canonical_value(
        "environment_id" => identity,
        "content_identity" => content_identity,
        "owner" => required_string(input.fetch(:owner), "environment input owner"),
        "reason" => required_string(input.fetch(:reason), "environment input reason")
      )
    end.sort_by { |input| input.fetch("environment_id").b }
  end
  private_class_method :validate_environment_inputs

  def validate_document!(document)
    unless document.is_a?(Hash) && document.keys == DOCUMENT_KEYS && document["schema_version"] == 2
      raise InvalidManifest, "repository input manifest document schema is invalid"
    end
    repositories = document["repositories"]
    environments = document["environment_inputs"]
    unless repositories.is_a?(Array) && repositories.any? && environments.is_a?(Array)
      raise InvalidManifest, "repository input manifest document collections are invalid"
    end
    unless repositories.all? { |entry| entry.is_a?(Hash) && entry.keys == SNAPSHOT_KEYS }
      raise InvalidManifest, "repository snapshot schema is invalid"
    end
    repository_identities = repositories.map do |snapshot|
      identity = required_string(snapshot["repository_identity"], "repository snapshot identity")
      raise InvalidManifest, "repository snapshot identity is invalid" unless valid_repository_identity?(identity)

      identity
    end
    unless repository_identities == repository_identities.sort_by(&:b) &&
      repository_identities.uniq.length == repositories.length
      raise InvalidManifest, "repository snapshots are not canonical and unique"
    end
    repositories.each do |snapshot|
      format = required_string(snapshot["object_format"], "repository snapshot object format")
      length = format == "git-sha256" ? 64 : 40
      revision = required_string(snapshot["revision"], "repository snapshot revision")
      tree = required_string(snapshot["tree"], "repository snapshot tree")
      unless %w[git-sha1 git-sha256].include?(format) &&
        revision.match?(%r{\A[0-9a-f]{#{length}}\z}) &&
        tree.match?(%r{\A#{format}:[0-9a-f]{#{length}}\z})
        raise InvalidManifest, "repository snapshot object identity is invalid"
      end
      inputs = snapshot.fetch("inputs")
      unless inputs.is_a?(Array) && inputs.any? &&
        inputs.all? { |input| input.is_a?(Hash) && input.keys == TRACKED_INPUT_KEYS }
        raise InvalidManifest, "tracked input schema is invalid"
      end
      paths = inputs.map do |input|
        path = required_string(input["path"], "tracked input path")
        kind = required_string(input["kind"], "tracked input kind")
        mode = required_string(input["mode"], "tracked input mode")
        content_identity = required_string(input["content_identity"], "tracked input content identity")
        valid_mode_kind = (kind == "blob" && %w[100644 100755 120000].include?(mode)) ||
          (kind == "commit" && mode == "160000")
        unless valid_relative_path?(path, allow_dot: false) && valid_mode_kind &&
          content_identity.match?(%r{\A#{format}:[0-9a-f]{#{length}}\z})
          raise InvalidManifest, "tracked input semantics are invalid"
        end
        required_string(input["owner"], "tracked input owner")
        required_string(input["reason"], "tracked input reason")
        path
      end
      unless paths == paths.sort_by(&:b) && paths.uniq.length == inputs.length
        raise InvalidManifest, "tracked inputs are not canonical and unique"
      end
    end
    unless environments.all? { |entry| entry.is_a?(Hash) && entry.keys == CANONICAL_ENVIRONMENT_KEYS } &&
      environments.all? { |entry| valid_canonical_environment_input?(entry) }
      raise InvalidManifest, "environment input semantics are invalid"
    end
    environment_identities = environments.map { |entry| entry.fetch("environment_id") }
    unless environment_identities == environment_identities.sort_by(&:b) &&
      environment_identities.uniq.length == environments.length
      raise InvalidManifest, "environment inputs are not canonical and unique"
    end
  end
  private_class_method :validate_document!

  def git_output!(root, *arguments, label:)
    output, error, status = Open3.capture3("git", "-C", root, *arguments)
    unless status.success? && error.empty?
      raise InvalidManifest, "#{label} is unavailable for repository root #{root}"
    end

    arguments.include?("-z") ? output : output.strip
  end
  private_class_method :git_output!

  def required_string(value, label)
    unless value.is_a?(String) && !value.empty? && value.valid_encoding? && !value.include?("\0")
      raise InvalidManifest, "#{label} is invalid"
    end

    value
  end
  private_class_method :required_string

  def exact_keys?(value, expected)
    value.is_a?(Hash) && value.keys.length == expected.length &&
      (value.keys - expected).empty? && (expected - value.keys).empty?
  end
  private_class_method :exact_keys?

  def valid_relative_path?(path, allow_dot:)
    pathname = Pathname.new(path)
    return false if pathname.absolute? || path.include?("\\") || path.split("/").include?("..")
    return false unless pathname.cleanpath.to_s == path

    (allow_dot && path == ".") || path.match?(%r{\A[^/\0]+(?:/[^/\0]+)*\z})
  end
  private_class_method :valid_relative_path?

  def valid_canonical_environment_input?(input)
    identity = required_string(input["environment_id"], "environment input identity")
    content_identity = required_string(input["content_identity"], "environment input content identity")
    required_string(input["owner"], "environment input owner")
    required_string(input["reason"], "environment input reason")
    identity.match?(%r{\Aenvironment:[a-z0-9]+(?:[._-][a-z0-9]+)*\z}) &&
      content_identity.match?(%r{\Asha256:[0-9a-f]{64}\z})
  end
  private_class_method :valid_canonical_environment_input?

  def fingerprint_projection(document)
    {
      "schema_version" => document.fetch("schema_version"),
      "repositories" => document.fetch("repositories").map do |snapshot|
        {
          "repository_identity" => snapshot.fetch("repository_identity"),
          "object_format" => snapshot.fetch("object_format"),
          "inputs" => snapshot.fetch("inputs").map do |input|
            input.slice("path", "mode", "kind", "content_identity")
          end
        }
      end,
      "environment_inputs" => document.fetch("environment_inputs").map do |input|
        input.slice("environment_id", "content_identity")
      end
    }
  end
  private_class_method :fingerprint_projection

  def canonical_value(value)
    case value
    when Hash
      value.keys.sort_by { |key| key.to_s.b }.to_h { |key| [key, canonical_value(value.fetch(key))] }
    when Array
      value.map { |item| canonical_value(item) }
    else
      value
    end
  end
  private_class_method :canonical_value

  def deep_freeze(value)
    case value
    when Hash
      value.each { |key, item| key.freeze; deep_freeze(item) }
    when Array
      value.each { |item| deep_freeze(item) }
    end
    value.freeze
  end
  private_class_method :deep_freeze
end
