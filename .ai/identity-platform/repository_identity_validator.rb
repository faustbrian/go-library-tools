# frozen_string_literal: true

require "open3"

module IdentityPlatformRepositoryIdentityValidator
  ASSIGNMENT_IDENTITY_FIELDS = %i[
    unit generation control_repository package_repository task
  ].freeze
  AUTHORIZATION_IDENTITY_FIELDS = %i[
    unit generation control_repository package_repository
  ].freeze

  module_function

  def assignment_identity(row)
    row.values_at(*ASSIGNMENT_IDENTITY_FIELDS)
  end

  def same_assignment?(left, right)
    assignment_identity(left) == assignment_identity(right)
  end

  def authorization_identity(row)
    row.values_at(*AUTHORIZATION_IDENTITY_FIELDS)
  end

  def same_authorization?(left, right)
    authorization_identity(left) == authorization_identity(right)
  end

  class Repositories
    def initialize(control_repository:, package_repository:)
      @control_repository = control_repository
      @package_repository = package_repository
    end

    def package_commit?(revision)
      commit?(@package_repository, revision)
    end

    def package_ancestor?(ancestor, descendant)
      ancestor?(@package_repository, ancestor, descendant)
    end

    def package_output(*arguments)
      output(@package_repository, *arguments)
    end

    def package_blob(revision, path)
      blob(@package_repository, revision, path)
    end

    def control_blob(revision, path)
      blob(@control_repository, revision, path)
    end

    def package_result_errors(baseline:, worker:, integration:, expected_merge:)
      errors = []
      errors << "package baseline commit is unreachable" unless package_commit?(baseline)
      errors << "package worker commit is unreachable" unless package_commit?(worker)
      errors << "package integration commit is unreachable" unless package_commit?(integration)
      errors << "package worker excludes baseline" unless package_ancestor?(baseline, worker)
      errors << "package integration excludes worker" unless package_ancestor?(worker, integration)
      if expected_merge
        parents = package_output("show", "-s", "--format=%P", integration).to_s.split
        errors << "package integration is not the exact baseline-worker merge" unless parents == [baseline, worker]
      end
      errors
    end

    private

    def commit?(repository, revision)
      return false unless revision.to_s.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/)

      _output, _error, status = Open3.capture3(
        "git", "-C", repository, "cat-file", "-e", "#{revision}^{commit}"
      )
      status.success?
    end

    def ancestor?(repository, ancestor, descendant)
      return false unless commit?(repository, ancestor) && commit?(repository, descendant)

      _output, _error, status = Open3.capture3(
        "git", "-C", repository, "merge-base", "--is-ancestor", ancestor, descendant
      )
      status.success?
    end

    def output(repository, *arguments)
      stdout, stderr, status = Open3.capture3("git", "-C", repository, *arguments)
      return stdout.strip if status.success? && stderr.empty?

      nil
    end

    def blob(repository, revision, path)
      return unless revision.to_s.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64}|HEAD)\z/)
      return unless path.to_s.match?(%r{\A[a-zA-Z0-9._/-]+\z}) &&
        !path.start_with?("/") && !path.split("/").include?("..")

      stdout, _stderr, status = Open3.capture3("git", "-C", repository, "show", "#{revision}:#{path}")
      status.success? ? stdout : nil
    end
  end
end
