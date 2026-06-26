<!--
Sync Impact Report
- Version change: 0.0.0 -> 1.0.0
- Modified principles:
  - Template Principle 1 -> I. Spec Before Build
  - Template Principle 2 -> II. Testable Delivery
  - Template Principle 3 -> III. Contract-First Integration
  - Template Principle 4 -> IV. Operability By Default
  - Template Principle 5 -> V. Keep Scope Honest
- Added sections:
  - Additional Constraints
  - Delivery Workflow
- Removed sections:
  - None
- Templates requiring updates:
  - ✅ .specify/templates/plan-template.md
  - ✅ .specify/templates/tasks-template.md
  - ✅ .specify/templates/spec-template.md
  - ⚠ .specify/templates/commands/ (directory not present in this repository)
- Follow-up TODOs:
  - None
-->
# FlowTask Constitution

## Core Principles

### I. Spec Before Build

Every implementable feature MUST have a current `spec.md`, `plan.md`, and
`tasks.md` before engineering work starts. Clarifications that materially affect
scope, behavior, contracts, or acceptance criteria MUST be written back to the
spec before implementation proceeds.

Rationale: the project uses spec-kit as the source of truth; stale or partial
artifacts create downstream rework.

### II. Testable Delivery

Every P1 and shared foundational change MUST include explicit validation work in
`tasks.md`. User-story tasks MUST define an independent test path, and
cross-cutting work MUST include regression or integration validation where it
changes behavior. Performance, concurrency, and error-handling requirements
with measurable success criteria MUST have a corresponding verification task.

Rationale: features are only complete when their acceptance path is executable
and measurable.

### III. Contract-First Integration

Changes that affect APIs, streaming payloads, state transitions, or shared data
models MUST update the relevant contract and design artifacts in the same change.
`contracts/`, `data-model.md`, `spec.md`, `plan.md`, and `tasks.md` MUST stay
mutually consistent for any externally visible behavior.

Rationale: this project spans a Go backend, a Next.js frontend, and AI-backed
streaming workflows; interface drift is one of the highest-cost failure modes.

### IV. Operability By Default

The system MUST define observable error handling, health checks, and startup
paths for every shared service dependency. Any feature introducing AI calls,
background-like aggregation, caching, or authentication MUST document failure
behavior and recovery expectations.

Rationale: FlowTask depends on PostgreSQL, Redis, JWT flows, and external AI
providers; operability cannot be deferred to polish.

### V. Keep Scope Honest

v1 scope MUST remain consistent across artifacts. Features declared out of scope
or deferred in the spec MUST NOT reappear as implied commitments in plans,
tasks, or contracts. When scope expands, the spec MUST be clarified first and
dependent artifacts MUST be regenerated or updated immediately.

Rationale: the fastest way to lose predictability is to let tasks silently grow
beyond the agreed product boundary.

## Additional Constraints

- The repository follows a monorepo split of `web/` and `server/`; document
  paths MUST match the chosen structure exactly.
- Authentication flows MUST use access + refresh token semantics consistently
  across spec, plan, contracts, and tasks.
- AI-facing features MUST define retry behavior, streaming behavior, and
  user-visible failure handling.
- Documentation changes that affect generation rules MUST update the relevant
  templates under `.specify/templates/`.

## Delivery Workflow

- `/speckit-clarify` MUST run before `/speckit-plan` when requirements still
  contain implementation-affecting ambiguity.
- `/speckit-analyze` MUST be treated as a release gate before
  `/speckit-implement` whenever `tasks.md` has been generated.
- Constitution violations are blocking until the impacted artifacts are brought
  back into compliance.

## Governance

This constitution supersedes conflicting planning or tasking guidance in the
repository. Amendments require:

1. An explicit update to this file.
2. A semantic version bump:
   - MAJOR for incompatible governance changes.
   - MINOR for new principles or materially expanded obligations.
   - PATCH for clarifications without changed obligations.
3. Review of impacted templates and active feature artifacts.

Every planning or implementation review MUST check compliance with these
principles. Any justified exception MUST be documented in the relevant `plan.md`
under a dedicated complexity or exception note.

**Version**: 1.0.0 | **Ratified**: 2026-06-17 | **Last Amended**: 2026-06-17
