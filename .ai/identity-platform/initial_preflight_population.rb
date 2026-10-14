# frozen_string_literal: true

module IdentityPlatformInitialPreflightPopulation
  PREFLIGHT_PATH = ".ai/identity-platform/PREFLIGHT_EVIDENCE.md"

  module_function

  def allowed?(changed_paths:, parent:, recorded_base:, previous_source:, current_source:)
    return false unless changed_paths.is_a?(Array) && changed_paths == [PREFLIGHT_PATH]
    return false unless parent.is_a?(String) && parent == recorded_base
    return false unless previous_source.is_a?(String) && current_source.is_a?(String)
    return false unless table_row_only_change?(previous_source, current_source)

    previous_execution = execution_sections(previous_source)
    current_execution = execution_sections(current_source)
    return false unless pending_table_cell?(previous_execution)
    return false if pending_table_cell?(current_execution)

    identity = execution_identity(current_source)
    identity.fetch("Recorded control execution baseline", nil) == recorded_base &&
      identity.fetch("Preflight input revision before the record commit", nil) == recorded_base
  end

  def table_row_only_change?(previous_source, current_source)
    return false unless previous_source.is_a?(String) && current_source.is_a?(String)
    return false if previous_source == current_source

    markdown_structure(previous_source).b == markdown_structure(current_source).b
  end

  def execution_sections(source)
    source[/^## Execution identity\n.*?(?=^## Conflict-recovery baselines|\z)/m].to_s
  end
  private_class_method :execution_sections

  def markdown_structure(source)
    lines = source.lines
    data_rows = markdown_data_row_indices(lines)
    lines.each_with_index.filter_map do |line, index|
      next if data_rows[index]

      line
    end.join
  end
  private_class_method :markdown_structure

  def markdown_data_row_indices(lines)
    indices = {}
    index = 0
    while index + 1 < lines.length
      unless markdown_table_row?(lines.fetch(index)) && markdown_separator_row?(lines.fetch(index + 1))
        index += 1
        next
      end

      index += 2
      while index < lines.length && markdown_table_row?(lines.fetch(index)) &&
            !markdown_separator_row?(lines.fetch(index))
        indices[index] = true
        index += 1
      end
    end
    indices
  end
  private_class_method :markdown_data_row_indices

  def markdown_table_row?(line)
    line.start_with?("|") && line.chomp.end_with?("|")
  end
  private_class_method :markdown_table_row?

  def markdown_separator_row?(line)
    line.chomp.match?(/\A\|(?:\s*:?-+:?\s*\|)+\z/)
  end
  private_class_method :markdown_separator_row?

  def pending_table_cell?(source)
    source.lines.any? do |line|
      line.start_with?("|") && line.split("|").map { |cell| plain_cell(cell) }.include?("pending")
    end
  end
  private_class_method :pending_table_cell?

  def execution_identity(source)
    section = source[/^## Execution identity\n(.*?)(?=^## |\z)/m, 1].to_s
    section.lines.filter_map do |line|
      next unless line.start_with?("|")

      cells = line.split("|", -1).map(&:strip)[1...-1]
      next unless cells.length == 2
      next if cells.first == "Field" || cells.first == "---"

      cells.map { |cell| plain_cell(cell) }
    end.to_h
  end
  private_class_method :execution_identity

  def plain_cell(value)
    value.to_s.strip.delete_prefix("`").delete_suffix("`")
  end
  private_class_method :plain_cell
end
