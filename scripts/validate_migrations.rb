#!/usr/bin/env ruby
# frozen_string_literal: true

require "pathname"

ROOT = Pathname.new(__dir__).join("..").expand_path
MIGRATIONS_DIR = ROOT.join("infrastructure/postgres/migrations")

required_up_fragments = [
  "CREATE SCHEMA IF NOT EXISTS identity",
  "CREATE SCHEMA IF NOT EXISTS investment",
  "CREATE SCHEMA IF NOT EXISTS analytics",
  "CREATE SCHEMA IF NOT EXISTS audit",
  "CREATE TABLE identity.users",
  "CREATE TABLE identity.user_investment_links",
  "CREATE TABLE investment.subjects",
  "CREATE TABLE investment.assets",
  "CREATE TABLE investment.portfolios",
  "CREATE TABLE investment.transaction_entries",
  "CREATE TABLE investment.command_deduplication",
  "CREATE TABLE investment.outbox_events",
  "CREATE TABLE analytics.portfolio_snapshots",
  "CREATE TABLE analytics.snapshot_positions",
  "CREATE TABLE analytics.calculation_runs",
  "CREATE TABLE analytics.inbox_messages",
  "CREATE TABLE audit.actors",
  "CREATE TABLE audit.events"
]

forbidden_up_patterns = {
  /\bDROP\b/i => "DROP is forbidden in up migrations",
  /\bTRUNCATE\b/i => "TRUNCATE is forbidden in up migrations",
  /\bDELETE\s+FROM\b/i => "DELETE FROM is forbidden in up migrations",
  /\bUPDATE\s+[a-zA-Z0-9_.\"]+\s+SET\b/i => "UPDATE is forbidden in up migrations",
  /\bALTER\s+TABLE\b.*\bDROP\b/im => "ALTER TABLE ... DROP is forbidden in up migrations"
}

unless MIGRATIONS_DIR.directory?
  warn "Missing migrations directory: #{MIGRATIONS_DIR}"
  exit 1
end

up_files = MIGRATIONS_DIR.children.select { |path| path.basename.to_s.end_with?(".up.sql") }.sort
down_files = MIGRATIONS_DIR.children.select { |path| path.basename.to_s.end_with?(".down.sql") }.sort

if up_files.empty?
  warn "No up migrations found"
  exit 1
end

up_prefixes = up_files.map { |path| path.basename.to_s.delete_suffix(".up.sql") }
down_prefixes = down_files.map { |path| path.basename.to_s.delete_suffix(".down.sql") }

missing_down = up_prefixes - down_prefixes
missing_up = down_prefixes - up_prefixes

unless missing_down.empty? && missing_up.empty?
  warn "Migration up/down mismatch"
  warn "Missing down migrations for: #{missing_down.join(", ")}" unless missing_down.empty?
  warn "Missing up migrations for: #{missing_up.join(", ")}" unless missing_up.empty?
  exit 1
end

unless up_prefixes == up_prefixes.sort
  warn "Migration files must be lexicographically ordered"
  exit 1
end

up_files.each do |path|
  sql = path.read

  forbidden_up_patterns.each do |pattern, message|
    if sql.match?(pattern)
      warn "#{path.relative_path_from(ROOT)}: #{message}"
      exit 1
    end
  end

  next unless path.basename.to_s.start_with?("000001_stage_03_01_vertical_slice")

  required_up_fragments.each do |fragment|
    unless sql.include?(fragment)
      warn "#{path.relative_path_from(ROOT)}: missing required fragment: #{fragment}"
      exit 1
    end
  end

  if sql.match?(/\bFLOAT\b|\bDOUBLE\b|\bREAL\b/i)
    warn "#{path.relative_path_from(ROOT)}: binary floating-point types are forbidden for financial schema"
    exit 1
  end

  unless sql.include?("NUMERIC(28, 8)")
    warn "#{path.relative_path_from(ROOT)}: expected NUMERIC(28, 8) decimal financial columns"
    exit 1
  end
end

puts "Validated #{up_files.length} migration pair(s)"
