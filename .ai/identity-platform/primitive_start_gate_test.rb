# frozen_string_literal: true

require "minitest/autorun"
require "set"

if File.file?(File.join(__dir__, "primitive_start_gate.rb"))
  require_relative "primitive_start_gate"
else
  module IdentityPlatformPrimitiveStartGate
    def self.consumer_state_errors(**)
      ["primitive start-gate reconciliation is unavailable"]
    end

    def self.reconciliation_errors(**)
      ["primitive start-gate reconciliation is unavailable"]
    end
  end
end

class IdentityPlatformPrimitiveStartGateTest < Minitest::Test
  StartGate = IdentityPlatformPrimitiveStartGate

  def test_permits_only_an_unassigned_initial_ready_row_during_preflight_reconciliation
    assert_empty StartGate.consumer_state_errors(
      row: inventory_row("ready"), entry: ledger_entry,
      primitives: %w[outbox telemetry], allow_initial_ready: true
    )

    refute_empty StartGate.consumer_state_errors(
      row: inventory_row("ready"), entry: ledger_entry,
      primitives: %w[outbox telemetry], allow_initial_ready: false
    )

    refute_empty StartGate.consumer_state_errors(
      row: inventory_row("ready"), entry: ledger_entry(transition: "v1 ready owner=— at=2026-08-30T00:00:00Z"),
      primitives: ["outbox"], allow_initial_ready: true
    )
    refute_empty StartGate.consumer_state_errors(
      row: inventory_row("ready"), entry: ledger_entry(task: "worker-identity-delivery"),
      primitives: ["outbox"], allow_initial_ready: true
    )
  end

  def test_accepts_proposed_and_exactly_blocked_consumers
    assert_empty StartGate.consumer_state_errors(
      row: inventory_row("proposed"), entry: ledger_entry,
      primitives: ["outbox"]
    )
    assert_empty StartGate.consumer_state_errors(
      row: inventory_row("blocked", owner: "blocker:primitive-outbox"), entry: ledger_entry,
      primitives: %w[outbox telemetry]
    )

    refute_empty StartGate.consumer_state_errors(
      row: inventory_row("blocked", owner: "blocker:primitive-workflow"), entry: ledger_entry,
      primitives: %w[outbox telemetry]
    )
    refute_empty StartGate.consumer_state_errors(
      row: inventory_row("in-progress", owner: "worker-identity-delivery"), entry: ledger_entry,
      primitives: ["outbox"]
    )
  end

  def test_noncurrent_primitive_allows_only_its_consumer_to_reconcile_to_proposed
    assert_empty StartGate.reconciliation_errors(
      unit: "identity/delivery", required_units: [], current_statuses: {},
      blocked_units: Set["identity/delivery"]
    )
    refute_empty StartGate.reconciliation_errors(
      unit: "primitive/authentication-identity-contracts",
      required_units: [], current_statuses: {},
      blocked_units: Set["identity/delivery"]
    )

    assert_empty StartGate.reconciliation_errors(
      unit: "dependent", required_units: ["foundation"],
      current_statuses: {"foundation" => "implemented-unverified"},
      blocked_units: Set.new
    )
  end

  private

  def inventory_row(status, owner: "—")
    {
      unit: "identity/delivery", module: "pkg/identity/delivery", requires: [],
      status: status, owner: owner, goal: "goals/identity-delivery.md"
    }
  end

  def ledger_entry(transition: "initial", task: "—")
    {
      generation: "0", task: task, branch: "—", worktree: "—",
      assignment: "—", assignment_tree: "—", worker_commit: "—",
      worker_tree: "—", checkpoint: "—", checkpoint_tree: "—",
      gate_revision: "—", gate_tree: "—", fingerprint: "—",
      external: "—", transition: transition
    }
  end
end
