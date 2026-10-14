# frozen_string_literal: true

require "digest"
require "json"
require "time"

module IdentityPlatformPrimitivePreflightEvidence
  TOOL_ENVIRONMENT_PATH = ".ai/identity-platform/evidence/preflight/tool-environment-inventory.json"
  ROOT_KEYS = %w[
    schema_version schema observed_control_revision recorded_at
    tool_environment_path tool_environment_sha256 api_input_scope
    gate_input_scope records
  ].freeze
  RECORD_KEYS = %w[
    primitive consumers registered repository_identity repository_revision
    repository_tree module_relative_path api_input_algorithm
    api_input_fingerprint gate_input_algorithm gate_input_fingerprint
    gate_command gate_working_directory gate_result gate_observation
  ].freeze
  OBSERVATION_KEYS = %w[
    classification exit_status log_path log_sha256 summary tested_revision
  ].freeze
  RESULTS = %w[pass failed blocked stale].freeze
  TOOL_ROOT_KEYS = %w[
    schema_version schema control_revision recorded_at platform tools
    service_profiles harnesses limitations
  ].freeze
  TOOL_IDS = %w[
    go ruby go-library-tools codex docker gitleaks golangci-lint staticcheck
    govulncheck gremlins
  ].freeze
  PINNED_TOOL_IDENTITIES = {
    "go" => "version:go1.27.0 darwin/arm64 path:/opt/homebrew/Cellar/go/1.27.0/libexec/bin/go binary-sha256:71c4991041d8e44975c882e4f72005719c958013d3340dc665a3808b72ddf702",
    "ruby" => "version:ruby 4.0.6 (2026-07-14 revision 03b6d3f889) +PRISM [arm64-darwin25] path:/opt/homebrew/Cellar/ruby/4.0.6_1/bin/ruby binary-sha256:341d69419d11b2706864868f08d987d1d0dfde226669f49f210f1d5e823e98e5",
    "go-library-tools" => "source:f448c203121b7edb90e7648a827ec5806208c723 binary-sha256:65fd0f74a7fbb235aa7c6a8bfdbcf5526fd1b0fd6799c75a6de3eccd255e447f",
    "codex" => "version:codex-cli 0.151.0 path:/opt/homebrew/Caskroom/codex/0.151.0/bin/codex binary-sha256:98491713ffb196061003ee148636e743997cc31d76144ba7c53462269896891d",
    "docker" => "version:client=29.7.2 server=29.7.2 path:/Applications/Docker.app/Contents/Resources/bin/docker binary-sha256:a6822407e207ff0b687419ccbe55700ffbf389d77de58ef625b3b1790d88747c",
    "gitleaks" => "version:version is set by build process path:/Users/brian/go/bin/gitleaks binary-sha256:366c5649485d51596d727080b43d05ccb9aec6ec36d94f72638c90cd41515bfd",
    "golangci-lint" => "version:2.12.2 go1.26.5 path:/Users/brian/go/bin/golangci-lint binary-sha256:7313eccd229be768bada337a73097ef2082ce08d08962ece4603409d8d28fe7d",
    "staticcheck" => "version:2026.1 (v0.7.0) path:/Users/brian/go/bin/staticcheck binary-sha256:47d5f8af165e49bc2d7785599bf5c7948fd33dd03f04871d23f9f09779ac7d5f",
    "govulncheck" => "version:govulncheck@v1.6.0 go1.27.0 path:/Users/brian/go/bin/govulncheck binary-sha256:5b781c87b5eccf29f7339672a64e9fdec47cdce0fc98c196911a3cbdf1609d4e",
    "gremlins" => "version:dev darwin/arm64 path:/Users/brian/go/bin/gremlins binary-sha256:d0ebb7c967c345818693785e3e8e4f87055cc0baff3d7b19b00429cd63399013"
  }.freeze
  SERVICE_IMAGES = {
    "postgresql" => {
      "image" => "postgres:18.4-alpine",
      "digest" => "sha256:9a8afca54e7861fd90fab5fdf4c42477a6b1cb7d293595148e674e0a3181de15"
    },
    "valkey" => {
      "image" => "valkey/valkey:9.1.0-alpine",
      "digest" => "sha256:c9b77919daeba2c02ad954d0c844cc4e7142069d177b89c5fd771f405daf9e02"
    }
  }.freeze
  LIMITATIONS = [
    "Tool presence and image availability are preflight facts, not passing package or interoperability evidence.",
    "Secret-bearing environment values were neither read nor recorded."
  ].freeze

  module_function

  def parse_source(source)
    document = JSON.parse(source)
    errors = []
    errors << "primitive preflight evidence is not canonical JSON" unless
      source == JSON.pretty_generate(document) + "\n"
    errors << "primitive preflight evidence schema drifted" unless
      document["schema"] == "identity-platform.preflight-primitive-contracts.v1"
    [document, errors]
  rescue JSON::ParserError
    [{}, ["primitive preflight evidence is invalid JSON"]]
  end

  def tool_environment_errors(document, expected_control_revision:, expected_recorded_at:,
                              expected_tool_identities:)
    return ["primitive preflight tool environment root schema is invalid"] unless
      document.is_a?(Hash) && document.keys == TOOL_ROOT_KEYS

    errors = []
    errors << "primitive preflight tool environment schema version drifted" unless document["schema_version"] == 1
    errors << "primitive preflight tool environment schema drifted" unless
      document["schema"] == "identity-platform.preflight-tool-environment.v1"
    errors << "primitive preflight tool environment control revision drifted" unless
      document["control_revision"] == expected_control_revision
    errors << "primitive preflight tool environment timestamp drifted" unless
      document["recorded_at"] == expected_recorded_at && rfc3339?(document["recorded_at"])
    errors.concat(platform_errors(document["platform"]))
    errors.concat(tool_errors(document["tools"], expected_control_revision, expected_tool_identities))
    errors.concat(service_profile_errors(document["service_profiles"]))
    errors.concat(harness_errors(document["harnesses"]))
    errors << "primitive preflight tool environment limitations drifted" unless document["limitations"] == LIMITATIONS
    errors
  end

  def errors(document:, expected_consumers:, expected_repositories:,
             tool_environment_path:, tool_environment_sha256:, artifact_sha256: nil)
    errors = []
    unless document.is_a?(Hash) && document.keys == ROOT_KEYS
      return ["primitive preflight evidence root schema is invalid"]
    end
    errors << "primitive preflight evidence schema version drifted" unless document["schema_version"] == 1
    errors << "primitive preflight evidence schema drifted" unless
      document["schema"] == "identity-platform.preflight-primitive-contracts.v1"
    errors << "primitive preflight observed control revision is invalid" unless
      full_sha1?(document["observed_control_revision"])
    errors << "primitive preflight recorded_at is invalid" unless rfc3339?(document["recorded_at"])
    errors << "primitive preflight tool path drifted" unless
      document["tool_environment_path"] == tool_environment_path
    errors << "primitive preflight tool digest drifted" unless
      document["tool_environment_sha256"] == tool_environment_sha256
    errors << "primitive preflight API input scope drifted" unless
      document["api_input_scope"] == "complete tracked repository tree at the recorded revision; conservative changes invalidate the fingerprint"
    errors << "primitive preflight gate input scope drifted" unless
      document["gate_input_scope"] == "complete tracked repository tree, module label, command, and tool-environment artifact digest"

    records = document["records"]
    return errors << "primitive preflight records are invalid" unless records.is_a?(Array)

    primitives = records.filter_map { |record| record["primitive"] if record.is_a?(Hash) }
    expected_primitives = expected_consumers.keys.sort_by(&:b)
    errors << "primitive preflight record closure or order drifted" unless primitives == expected_primitives
    errors << "primitive preflight records contain duplicates" unless primitives == primitives.uniq
    records.each do |record|
      errors.concat(record_errors(
        record, expected_consumers: expected_consumers,
        expected_repositories: expected_repositories,
        tool_environment_sha256: tool_environment_sha256,
        artifact_sha256: artifact_sha256
      ))
    end
    errors
  end

  def api_input_fingerprint(record)
    projection = {
      "repository_identity" => record.fetch("repository_identity"),
      "repository_revision" => record.fetch("repository_revision"),
      "repository_tree" => record.fetch("repository_tree"),
      "module_relative_path" => record.fetch("module_relative_path")
    }
    fingerprint("identity-platform-primitive-api-input-v1", projection)
  end

  def gate_input_fingerprint(record, tool_environment_sha256)
    projection = {
      "repository_identity" => record.fetch("repository_identity"),
      "repository_revision" => record.fetch("repository_revision"),
      "repository_tree" => record.fetch("repository_tree"),
      "module_relative_path" => record.fetch("module_relative_path"),
      "command" => record.fetch("gate_command"),
      "working_directory" => record.fetch("gate_working_directory"),
      "tool_environment_sha256" => tool_environment_sha256
    }
    fingerprint("identity-platform-primitive-gate-input-v1", projection)
  end

  def record_errors(record, expected_consumers:, expected_repositories:, tool_environment_sha256:, artifact_sha256:)
    return ["primitive preflight record schema is invalid"] unless
      record.is_a?(Hash) && record.keys == RECORD_KEYS

    primitive = record["primitive"]
    errors = []
    expected_repository = expected_repositories[primitive]
    errors << "primitive preflight record has an unknown primitive #{primitive.inspect}" unless
      expected_consumers.key?(primitive) && expected_repository
    return errors unless expected_consumers.key?(primitive) && expected_repository

    errors << "primitive #{primitive} consumers drifted" unless
      record["consumers"] == expected_consumers.fetch(primitive).sort_by(&:b)
    errors << "primitive #{primitive} registration drifted" unless record["registered"] == primitive
    expected_repository.each do |field, expected|
      errors << "primitive #{primitive} #{field.tr('_', ' ')} drifted" unless record[field] == expected
    end
    errors << "primitive #{primitive} API input algorithm drifted" unless
      record["api_input_algorithm"] == "identity-platform-primitive-api-input-v1"
    errors << "primitive #{primitive} API input fingerprint drifted" unless
      record["api_input_fingerprint"] == api_input_fingerprint(record)
    errors << "primitive #{primitive} gate input algorithm drifted" unless
      record["gate_input_algorithm"] == "identity-platform-primitive-gate-input-v1"
    errors << "primitive #{primitive} gate input fingerprint drifted" unless
      record["gate_input_fingerprint"] == gate_input_fingerprint(record, tool_environment_sha256)
    errors << "primitive #{primitive} gate command drifted" unless record["gate_command"] == %w[make check]
    errors << "primitive #{primitive} gate working directory drifted" unless record["gate_working_directory"] == "."
    errors << "primitive #{primitive} gate result is invalid" unless RESULTS.include?(record["gate_result"])
    errors.concat(observation_errors(primitive, record, artifact_sha256))
    errors
  rescue KeyError
    ["primitive preflight record schema is invalid"]
  end
  private_class_method :record_errors

  def observation_errors(primitive, record, artifact_sha256)
    observation = record["gate_observation"]
    return ["primitive #{primitive} gate observation schema is invalid"] unless
      observation.is_a?(Hash) && observation.keys == OBSERVATION_KEYS

    errors = []
    errors << "primitive #{primitive} gate observation summary is empty" unless
      observation["summary"].is_a?(String) && !observation["summary"].empty?
    case observation["classification"]
    when "not-executed"
      errors << "primitive #{primitive} non-executed gate is not stale" unless record["gate_result"] == "stale"
      errors << "primitive #{primitive} non-executed gate carries execution facts" unless
        observation.values_at("exit_status", "log_path", "log_sha256", "tested_revision").all?(&:nil?)
    when "executed"
      status = observation["exit_status"]
      errors << "primitive #{primitive} executed gate exit status is invalid" unless status.is_a?(Integer) && status >= 0
      expected_result = status == 0 ? "pass" : "failed"
      errors << "primitive #{primitive} executed gate result drifted" unless record["gate_result"] == expected_result
      errors << "primitive #{primitive} executed gate revision drifted" unless
        observation["tested_revision"] == record["repository_revision"]
      log_path = observation["log_path"]
      log_digest = observation["log_sha256"]
      errors << "primitive #{primitive} executed gate log path is invalid" unless
        log_path.is_a?(String) && log_path.match?(%r{\A\.ai/identity-platform/evidence/preflight/primitive-gates/[a-zA-Z0-9._-]+\.log\z})
      errors << "primitive #{primitive} executed gate log digest is invalid" unless
        sha256?(log_digest)
      if log_path.is_a?(String) && sha256?(log_digest)
        observed_digest = artifact_sha256.call(log_path) if artifact_sha256.respond_to?(:call)
        errors << "primitive #{primitive} executed gate log custody drifted" unless observed_digest == log_digest
      end
    else
      errors << "primitive #{primitive} gate observation classification is invalid"
    end
    errors
  end
  private_class_method :observation_errors

  def platform_errors(platform)
    return ["primitive preflight tool platform schema is invalid"] unless
      platform.is_a?(Hash) && platform.keys == %w[os architecture cgo_enabled go_toolchain_policy]

    errors = []
    errors << "primitive preflight tool platform OS is invalid" unless platform["os"].is_a?(String) && !platform["os"].empty?
    errors << "primitive preflight tool platform architecture is invalid" unless
      platform["architecture"].is_a?(String) && !platform["architecture"].empty?
    errors << "primitive preflight tool platform CGO identity is invalid" unless [true, false].include?(platform["cgo_enabled"])
    errors << "primitive preflight toolchain policy is invalid" unless
      platform["go_toolchain_policy"].is_a?(String) && !platform["go_toolchain_policy"].empty?
    errors
  end
  private_class_method :platform_errors

  def tool_errors(tools, expected_control_revision, expected_tool_identities)
    return ["primitive preflight tool identities are invalid"] unless tools.is_a?(Array)

    errors = []
    ids = tools.filter_map { |tool| tool["id"] if tool.is_a?(Hash) }
    errors << "primitive preflight tool identity closure drifted" unless ids == TOOL_IDS
    errors << "primitive preflight tool identities contain duplicates" unless ids == ids.uniq
    errors << "primitive preflight expected tool authority closure drifted" unless
      expected_tool_identities.is_a?(Hash) && expected_tool_identities.keys == TOOL_IDS
    tools.each do |tool|
      unless tool.is_a?(Hash) && tool.keys == %w[id classification identity]
        errors << "primitive preflight tool identity schema is invalid"
        next
      end
      errors << "primitive preflight tool #{tool['id']} classification drifted" unless tool["classification"] == "available"
      errors << "primitive preflight tool #{tool['id']} identity is empty" unless
        tool["identity"].is_a?(String) && !tool["identity"].empty?
      errors << "primitive preflight tool #{tool['id']} identity drifted" unless
        expected_tool_identities.is_a?(Hash) && tool["identity"] == expected_tool_identities[tool["id"]]
    end
    control = tools.find { |tool| tool.is_a?(Hash) && tool["id"] == "go-library-tools" }
    expected_pattern = /\Asource:#{Regexp.escape(expected_control_revision)} binary-sha256:[0-9a-f]{64}\z/
    errors << "primitive preflight go-library-tools identity is not control-bound" unless
      control && control["identity"].is_a?(String) && control["identity"].match?(expected_pattern)
    errors
  end
  private_class_method :tool_errors

  def service_profile_errors(profiles)
    return ["primitive preflight service profiles are invalid"] unless profiles.is_a?(Array)

    errors = []
    ids = profiles.filter_map { |profile| profile["id"] if profile.is_a?(Hash) }
    errors << "primitive preflight service profile closure drifted" unless ids == SERVICE_IMAGES.keys
    profiles.each do |profile|
      unless profile.is_a?(Hash) && profile.keys == %w[id classification image digest]
        errors << "primitive preflight service profile schema is invalid"
        next
      end
      expected_image = SERVICE_IMAGES[profile["id"]]
      errors << "primitive preflight service profile #{profile['id']} image drifted" unless
        expected_image && profile["classification"] == "available" &&
        profile["image"] == expected_image["image"]
      errors << "primitive preflight service profile #{profile['id']} digest drifted" unless
        expected_image && profile["digest"] == expected_image["digest"]
    end
    errors
  end
  private_class_method :service_profile_errors

  def harness_errors(harnesses)
    expected = [
      ["race-fuzz-leak-mutation", "available", "identity"],
      ["browser", "unavailable", "blocker"],
      ["interoperability", "not-yet-needed", "identity"]
    ]
    return ["primitive preflight harnesses are invalid"] unless harnesses.is_a?(Array)

    errors = []
    actual = harnesses.map do |harness|
      next unless harness.is_a?(Hash)

      value_key = harness.key?("identity") ? "identity" : harness.key?("blocker") ? "blocker" : nil
      [harness["id"], harness["classification"], value_key]
    end
    errors << "primitive preflight harness closure drifted" unless actual == expected
    harnesses.each do |harness|
      next unless harness.is_a?(Hash)

      value_key = harness.key?("identity") ? "identity" : "blocker"
      errors << "primitive preflight harness schema is invalid" unless
        harness.keys == ["id", "classification", value_key] &&
        harness[value_key].is_a?(String) && !harness[value_key].empty?
    end
    errors
  end
  private_class_method :harness_errors

  def fingerprint(domain, value)
    "sha256:#{Digest::SHA256.hexdigest("#{domain}\n#{JSON.generate(canonical(value))}")}"
  end
  private_class_method :fingerprint

  def canonical(value)
    case value
    when Hash then value.keys.sort.to_h { |key| [key, canonical(value.fetch(key))] }
    when Array then value.map { |entry| canonical(entry) }
    else value
    end
  end
  private_class_method :canonical

  def sha256?(value)
    value.is_a?(String) && value.match?(/\Asha256:[0-9a-f]{64}\z/)
  end
  private_class_method :sha256?

  def full_sha1?(value)
    value.is_a?(String) && value.match?(/\A[0-9a-f]{40}\z/)
  end
  private_class_method :full_sha1?

  def rfc3339?(value)
    Time.iso8601(value)
    value.match?(/\A\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z\z/)
  rescue ArgumentError, TypeError
    false
  end
  private_class_method :rfc3339?
end
