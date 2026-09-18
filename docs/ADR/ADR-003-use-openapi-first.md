# ADR-003: Use an OpenAPI-first API contract

- Status: Accepted
- Date: 2026-06-19

## Context

Web, mobile, desktop, and future public clients must share one server-side business contract.

## Decision

Design and version HTTP contracts in OpenAPI before implementation. Clients consume server-calculated results and do not duplicate business logic.

## Consequences
