#!/usr/bin/env ruby
# frozen_string_literal: true

require "json"
require "open3"
require "digest"
require "pathname"
require "uri"

module IdentityPlatformFinalReviewGate
  MANIFEST_KEYS = %w[repositories schema schema_version].freeze
  MANIFEST_REPOSITORY_KEYS = %w[base_revision head_revision head_tree object_format repository_identity worktree].freeze
  REVIEW_KEYS = %w[critical important].freeze
  OUTPUT_KEYS = %w[manifest_sha256 repositories schema schema_version].freeze
  OUTPUT_REPOSITORY_KEYS = %w[base_revision head_revision head_tree object_format output output_sha256 repository_identity worktree].freeze
  PROMPT = 'Review the complete identity-platform diff for unresolved Critical or Important correctness, security, contract, dependency, evidence, or orchestration findings. Output only canonical compact JSON with exactly the ordered keys critical and important, each an array of finding objects; use {"critical":[],"important":[]} only when neither class has a finding.'

  module_function

  def run(manifest_bytes, reviewer: method(:codex_review))
    manifest = parse_canonical(manifest_bytes, "final review manifest")
    validate_manifest!(manifest)
    rows = manifest.fetch("repositories").map do |repository|
      validate_worktree!(repository)
      review_bytes = reviewer.call(
        base: repository.fetch("base_revision"),
        chdir: repository.fetch("worktree"),
        prompt: PROMPT
      )
      review = parse_canonical(review_bytes, "review output for #{repository.fetch('repository_identity')}")
      validate_review!(review, repository.fetch("repository_identity"))
      validate_worktree!(repository)
      canonical_review = JSON.generate(review)
      repository.merge(
        "output" => review,
        "output_sha256" => "sha256:#{Digest::SHA256.hexdigest(canonical_review)}"
      ).sort.to_h
    end
    manifest.fetch("repositories").each { |repository| validate_worktree!(repository) }
    JSON.generate(
      "manifest_sha256" => "sha256:#{Digest::SHA256.hexdigest(manifest_bytes)}",
      "repositories" => rows,
      "schema" => "identity-platform.final-review.v2",
      "schema_version" => 2
    ).tap do |bytes|
      raise "final review output schema is not canonical" unless JSON.parse(bytes).keys == OUTPUT_KEYS
      raise "final review repository output schema is invalid" unless rows.all? { |row| row.keys == OUTPUT_REPOSITORY_KEYS }
    end
  end

  def codex_review(base:, chdir:, prompt:)
    stdout, stderr, status = Open3.capture3("codex", "review", "--base", base, prompt, chdir: chdir)
    raise "independent final review failed in #{chdir}: #{stderr}" unless status.success?

    stdout.strip
  end

  def parse_canonical(bytes, label)
    value = JSON.parse(bytes)
    raise "#{label} is not canonical compact JSON" unless bytes == JSON.generate(value)

    value
  rescue JSON::ParserError => e
    raise "#{label} is invalid JSON: #{e.message}"
  end

  def validate_manifest!(manifest)
    unless manifest.is_a?(Hash) && manifest.keys == MANIFEST_KEYS &&
      manifest["schema_version"] == 1 && manifest["schema"] == "identity-platform.final-review-manifest.v1"
      raise "final review manifest schema is invalid"
    end
    repositories = manifest["repositories"]
    unless repositories.is_a?(Array) && repositories.any? &&
      repositories.all? { |row| row.is_a?(Hash) && row.keys == MANIFEST_REPOSITORY_KEYS }
      raise "final review manifest repository schema is invalid"
    end
    identities = repositories.map { |row| row.fetch("repository_identity") }
    unless identities == identities.sort_by(&:b) && identities.uniq.length == identities.length
      raise "final review manifest repositories are not canonical and unique"
    end
    repositories.each { |repository| validate_repository_fields!(repository) }
  end

  def validate_repository_fields!(repository)
    format = repository.fetch("object_format")
    length = format == "git-sha256" ? 64 : 40
    unless %w[git-sha1 git-sha256].include?(format) &&
      repository.fetch("base_revision").match?(/\A[0-9a-f]{#{length}}\z/) &&
      repository.fetch("head_revision").match?(/\A[0-9a-f]{#{length}}\z/) &&
      repository.fetch("head_tree").match?(/\A#{format}:[0-9a-f]{#{length}}\z/) &&
      valid_repository_identity?(repository.fetch("repository_identity")) &&
      Pathname.new(repository.fetch("worktree")).absolute?
      raise "final review manifest repository identity is invalid"
    end
  end

  def validate_worktree!(repository)
    worktree = File.realpath(repository.fetch("worktree"))
    raise "final review worktree path is not canonical" unless worktree == repository.fetch("worktree")
    top = git!(worktree, "rev-parse", "--show-toplevel")
    raise "final review repository is not an exact worktree root" unless File.realpath(top) == worktree
    format = "git-#{git!(worktree, 'rev-parse', '--show-object-format')}"
    identity = repository_identity_from_remote(git!(worktree, "config", "--get", "remote.origin.url"))
    raise "final review repository identity substitution" unless identity == repository.fetch("repository_identity")
    raise "final review repository object format drift" unless format == repository.fetch("object_format")
    head = git!(worktree, "rev-parse", "--verify", "HEAD^{commit}")
    tree = "#{format}:#{git!(worktree, 'rev-parse', '--verify', 'HEAD^{tree}')}"
    unless head == repository.fetch("head_revision") && tree == repository.fetch("head_tree")
      raise "final review repository head or tree drift"
    end
    raise "final review repository worktree is dirty" unless git!(worktree, "status", "--porcelain=v1").empty?
    base = git!(worktree, "rev-parse", "--verify", "#{repository.fetch('base_revision')}^{commit}")
    raise "final review repository base revision drift" unless base == repository.fetch("base_revision")
    _out, _error, status = Open3.capture3(
      "git", "-C", worktree, "merge-base", "--is-ancestor",
      repository.fetch("base_revision"), repository.fetch("head_revision")
    )
    raise "final review repository base is not an ancestor of head" unless status.success?
  rescue Errno::ENOENT, Errno::EACCES, Errno::ELOOP
    raise "final review worktree is absent or cannot be resolved"
  end

  def validate_review!(review, identity)
    unless review.is_a?(Hash) && review.keys == REVIEW_KEYS &&
      review.values.all? { |findings| findings.is_a?(Array) && findings.all? { |finding| finding.is_a?(Hash) } }
      raise "review output schema for #{identity} is invalid"
    end
  end

  def git!(worktree, *arguments)
    output, error, status = Open3.capture3("git", "-C", worktree, *arguments)
    raise "final review Git inspection failed: #{error}" unless status.success?

    output.strip
  end

  def valid_repository_identity?(identity)
    identity.is_a?(String) && identity.match?(%r{\A[a-z0-9.-]+/[A-Za-z0-9._-]+(?:/[A-Za-z0-9._-]+)+\z}) &&
      !identity.split("/").include?("..")
  end

  def repository_identity_from_remote(remote)
    if (match = remote.match(%r{\A(?:[^@/]+@)?([^:/]+):([^/].*)\z})) && !remote.include?("://")
      host = match[1]
      path = match[2]
    else
      uri = URI.parse(remote)
      return unless %w[https ssh].include?(uri.scheme) && uri.host && !uri.query && !uri.fragment

      host = uri.host
      path = uri.path.delete_prefix("/")
    end
    path = path.delete_suffix(".git").delete_suffix("/")
    identity = "#{host.downcase}/#{path}"
    valid_repository_identity?(identity) ? identity : nil
  rescue URI::InvalidURIError
    nil
  end
end

if $PROGRAM_NAME == __FILE__
  raise "final review gate does not accept positional arguments" unless ARGV.empty?

  manifest_path = ENV.fetch("IDENTITY_FINAL_REVIEW_MANIFEST_PATH")
  output = IdentityPlatformFinalReviewGate.run(File.binread(manifest_path))
  File.binwrite(ENV.fetch("IDENTITY_ACCEPTANCE_RAW_OUTPUT_PATH"), output)
  print output
end
