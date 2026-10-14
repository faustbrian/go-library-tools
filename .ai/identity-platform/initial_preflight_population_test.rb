# frozen_string_literal: true

require "minitest/autorun"
if File.file?(File.join(__dir__, "initial_preflight_population.rb"))
  require_relative "initial_preflight_population"
else
  module IdentityPlatformInitialPreflightPopulation
    def self.allowed?(**) = false
  end
end

class IdentityPlatformInitialPreflightPopulationTest < Minitest::Test
  Population = IdentityPlatformInitialPreflightPopulation
  PREFLIGHT = ".ai/identity-platform/PREFLIGHT_EVIDENCE.md"
  BASE = "a" * 40

  def test_accepts_only_the_complete_direct_initial_population
    assert Population.allowed?(
      changed_paths: [PREFLIGHT], parent: BASE, recorded_base: BASE,
      previous_source: preflight_source("pending"),
      current_source: preflight_source(BASE)
    )
  end

  def test_accepts_byte_identical_structure_across_string_encodings
    previous = preflight_source("pending").sub("Evidence", "Evidence — capture").b
    current = preflight_source(BASE).sub("Evidence", "Evidence — capture")

    assert Population.allowed?(
      changed_paths: [PREFLIGHT], parent: BASE, recorded_base: BASE,
      previous_source: previous,
      current_source: current
    )
  end

  def test_rejects_later_multirow_or_mixed_path_changes
    refute Population.allowed?(
      **valid_arguments,
      parent: "b" * 40
    )
    refute Population.allowed?(
      **valid_arguments,
      changed_paths: [PREFLIGHT, ".ai/identity-platform/INVENTORY.md"]
    )
  end

  def test_rejects_incomplete_or_wrongly_bound_population
    refute Population.allowed?(
      **valid_arguments,
      current_source: preflight_source(BASE, tool_value: "pending")
    )
    refute Population.allowed?(
      **valid_arguments,
      current_source: preflight_source("b" * 40)
    )
    refute Population.allowed?(
      **valid_arguments,
      previous_source: preflight_source(BASE)
    )
  end

  def test_rejects_prose_heading_and_table_header_changes
    refute Population.allowed?(
      **valid_arguments,
      current_source: preflight_source(BASE) + "\nunauthorized prose\n"
    )

    refute Population.allowed?(
      **valid_arguments,
      current_source: preflight_source(BASE).sub(
        "## Tool and environment lanes",
        "## Rewritten tool and environment lanes"
      )
    )

    refute Population.allowed?(
      **valid_arguments,
      current_source: preflight_source(BASE).sub(
        "| Requirement/profile | Classification |",
        "| Requirement/profile | Rewritten classification |"
      )
    )

    refute Population.allowed?(
      **valid_arguments,
      current_source: preflight_source(BASE) + "\n| pipe-prefixed non-table content |\n"
    )
  end

  private

  def valid_arguments
    {
      changed_paths: [PREFLIGHT], parent: BASE, recorded_base: BASE,
      previous_source: preflight_source("pending"),
      current_source: preflight_source(BASE)
    }
  end

  def preflight_source(base, tool_value: "available")
    <<~MARKDOWN
      # Identity Platform Preflight Evidence

      ## Execution identity

      | Field | Value |
      | --- | --- |
      | Recorded control execution baseline | #{base} |
      | Preflight input revision before the record commit | #{base} |

      ## Tool and environment lanes

      | Requirement/profile | Classification |
      | --- | --- |
      | tools | #{tool_value} |

      ## Conflict-recovery baselines

      | Recovery epoch | Status |
      | --- | --- |
    MARKDOWN
  end
end
