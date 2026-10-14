# frozen_string_literal: true

require "open3"

module IdentityPlatformWorktreeRegistration
  module_function

  def registered_paths(repository_worktrees)
    Array(repository_worktrees).flat_map do |repository_worktree|
      output, _error, status = Open3.capture3(
        "git", "-C", repository_worktree, "worktree", "list", "--porcelain"
      )
      next [] unless status.success?

      output.lines.filter_map { |line| line[/\Aworktree (.+)\n?\z/, 1] }
        .filter_map do |candidate|
          File.realpath(candidate)
        rescue Errno::ENOENT
          nil
        end
    end.uniq
  end
end
