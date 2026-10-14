#!/usr/bin/env ruby
# frozen_string_literal: true

require "digest"
require "json"
require_relative "final_review_gate"

module IdentityPlatformFinalReviewVerifier
  OUTPUT_KEYS = %w[manifest_sha256 repositories schema schema_version].freeze
  OUTPUT_REPOSITORY_KEYS = %w[base_revision head_revision head_tree object_format output output_sha256 repository_identity worktree].freeze

  module_function

  def verify(bytes, manifest_bytes:, expected_repositories: nil,
             worktree_validator: IdentityPlatformFinalReviewGate.method(:validate_worktree!))
    review = IdentityPlatformFinalReviewGate.parse_canonical(bytes, "final review output")
    manifest = IdentityPlatformFinalReviewGate.parse_canonical(manifest_bytes, "final review manifest")
    IdentityPlatformFinalReviewGate.validate_manifest!(manifest)
    if expected_repositories && manifest.fetch("repositories") != expected_repositories
      raise "final review manifest repository closure drifted"
    end
    unless review.is_a?(Hash) && review.keys == OUTPUT_KEYS && review["schema_version"] == 2 &&
      review["schema"] == "identity-platform.final-review.v2"
      raise "final review output schema is invalid"
    end
    expected_manifest_digest = "sha256:#{Digest::SHA256.hexdigest(manifest_bytes)}"
    raise "final review manifest digest mismatch" unless review["manifest_sha256"] == expected_manifest_digest

    rows = review["repositories"]
    unless rows.is_a?(Array) && rows.all? { |row| row.is_a?(Hash) && row.keys == OUTPUT_REPOSITORY_KEYS }
      raise "final review repository result schema is invalid"
    end
    expected_rows = manifest.fetch("repositories")
    expected_identities = expected_rows.map { |row| row.fetch("repository_identity") }
    actual_identities = rows.map { |row| row.fetch("repository_identity") }
    unless actual_identities == expected_identities
      raise "final review repository rows are omitted, substituted, duplicated, or reordered"
    end

    rows.zip(expected_rows).each do |row, expected|
      identity = expected.fetch("repository_identity")
      unless row.slice(*IdentityPlatformFinalReviewGate::MANIFEST_REPOSITORY_KEYS) == expected
        raise "final review repository identity binding mismatch for #{identity}"
      end
      worktree_validator.call(expected)
      output = row.fetch("output")
      IdentityPlatformFinalReviewGate.validate_review!(output, identity)
      output_bytes = JSON.generate(output)
      expected_digest = "sha256:#{Digest::SHA256.hexdigest(output_bytes)}"
      raise "final review output digest mismatch for #{identity}" unless row["output_sha256"] == expected_digest
      unless output == {"critical" => [], "important" => []}
        raise "final review has a Critical or Important finding for #{identity}"
      end
    end

    JSON.generate({
      "critical_count" => 0,
      "important_count" => 0,
      "manifest_sha256" => expected_manifest_digest,
      "repository_count" => rows.length,
      "review_sha256" => "sha256:#{Digest::SHA256.hexdigest(bytes)}",
      "schema" => "identity-platform.final-review-verifier.v2",
      "schema_version" => 2
    })
  end
end

if $PROGRAM_NAME == __FILE__
  input_path = ENV.fetch("IDENTITY_ACCEPTANCE_VERIFIER_INPUT_PATH")
  output_path = ENV.fetch("IDENTITY_ACCEPTANCE_VERIFIER_OUTPUT_PATH")
  manifest_path = ENV.fetch("IDENTITY_FINAL_REVIEW_MANIFEST_PATH")
  File.binwrite(
    output_path,
    IdentityPlatformFinalReviewVerifier.verify(
      File.binread(input_path), manifest_bytes: File.binread(manifest_path)
    )
  )
end
