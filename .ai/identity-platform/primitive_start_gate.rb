# frozen_string_literal: true

module IdentityPlatformPrimitiveStartGate
  ASSIGNMENT_FIELDS = %i[
    task branch worktree assignment assignment_tree worker_commit worker_tree
    checkpoint checkpoint_tree gate_revision gate_tree fingerprint external
  ].freeze

  module_function

  def consumer_state_errors(row:, entry:, primitives:, allow_initial_ready: false)
    return [] if primitives.empty?

    errors = []
    case row.fetch(:status)
    when "proposed"
      errors << "proposed primitive consumer has an owner" unless row.fetch(:owner) == "—"
    when "blocked"
      blockers = primitives.map { |primitive| "blocker:primitive-#{primitive.tr('/', '-')}" }
      errors << "blocked primitive consumer lacks an exact blocker claim" unless blockers.include?(row.fetch(:owner))
    when "ready"
      errors << "ready primitive consumer is not an initial preflight reconciliation row" unless
        allow_initial_ready && initial_unassigned?(row: row, entry: entry)
    else
      errors << "primitive consumer has forbidden status #{row.fetch(:status)}"
    end
    errors
  end

  def reconciliation_errors(unit:, required_units:, current_statuses:, blocked_units:)
    dag_prerequisites_current = required_units.all? do |required|
      current_statuses.fetch(required) == "verified"
    end
    return [] unless dag_prerequisites_current && !blocked_units.include?(unit)

    ["returned to proposed without an invalidated prerequisite or primitive start gate"]
  end

  def initial_unassigned?(row:, entry:)
    row.fetch(:owner) == "—" && entry.fetch(:generation) == "0" && entry.fetch(:transition) == "initial" &&
      ASSIGNMENT_FIELDS.all? { |field| entry.fetch(field) == "—" }
  end
  private_class_method :initial_unassigned?
end
