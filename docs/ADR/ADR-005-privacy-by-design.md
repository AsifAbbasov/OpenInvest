# ADR-005: Apply Privacy by Design

- Status: Accepted
- Date: 2026-06-19

## Context

OpenInvest handles sensitive financial and optional identity or tax data. The user owns the data, and unnecessary collection creates avoidable risk.

## Decision

Collect only required data, keep Privacy Mode enabled by default, segregate identity, investment, tax, audit, and notification data, and encrypt sensitive fields. Passport, INN, address, and tax-profile storage remain optional.

## Consequences

Features must justify every collected field and support export and deletion. Any privacy-model change requires a superseding ADR.
