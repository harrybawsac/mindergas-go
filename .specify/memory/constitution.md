<!--
Sync Impact Report
- Version change: <none> -> 1.0.0
- Modified principles: none (template filled)
- Added sections: none
- Removed sections: none
- Templates checked: .specify/templates/plan-template.md (✅), .specify/templates/spec-template.md (✅), .specify/templates/tasks-template.md (✅), .specify/templates/agent-file-template.md (✅)
- Files requiring manual follow-up: none identified
- Deferred placeholders: RATIFICATION_DATE (TODO)
-->

# Mindergas Constitution
<!-- Project identity derived from repository folder name `mindergas`. -->

## Core Principles

### I. Library-First
Every new capability MUST start as a standalone, well-scoped library or package. Libraries
MUST be self-contained, independently testable, and documented with a clear purpose. The
team MUST avoid creating organizational-only libraries that do not have a stable public
contract or independent test coverage.

Rationale: Small, focused libraries encourage reuse, clear ownership, and make testing
and review simpler. This reduces coupling and eases distribution across services.

### II. CLI Interface
Libraries and developer tools SHOULD expose a clear command-line interface (CLI) for
automation and local iteration. When applicable, text-based protocols MUST follow the
convention: structured input via stdin/args, primary output via stdout, and errors via
stderr. CLIs MUST support a machine-readable format (JSON) and a human-readable mode
for debugging.

Rationale: A predictable CLI surface simplifies automation, CI integration, and manual
debugging across environments.

### III. Test-First (NON-NEGOTIABLE)
Test-First development is REQUIRED. For any new feature or change, tests MUST be
written before implementation. The Red-Green-Refactor cycle is the default workflow:
write failing tests, implement minimal code to pass, then refactor with tests green. All
changes to public contracts MUST include contract tests that fail before implementation.

Rationale: Test-First enforces clarity in requirements, prevents regressions, and makes
design decisions explicit through tests.

### IV. Integration Testing
Integration tests MUST cover cross-component interactions, public contract changes,
inter-service communication, and shared schemas. Integration tests SHOULD be run in CI
and exercised locally with lightweight fixtures where possible. Any change that affects
contracts MUST include updated integration tests and a migration plan if required.

Rationale: Integration tests protect system-level behavior and capture issues that unit
tests alone cannot reveal.

### V. Observability, Versioning & Simplicity
Systems and libraries MUST prioritize observability: structured logging, clear error
messages, and measurable metrics for critical flows. Versioning MUST follow semantic
versioning for public libraries (MAJOR.MINOR.PATCH). Breaking changes MUST increment
MAJOR; new principles or policy additions SHOULD increment MINOR; clarifications or
typo fixes SHOULD increment PATCH.

Keep designs as simple as possible (YAGNI): do not add capabilities without a clear
justification and a documented rationale.

Rationale: Observability reduces time-to-detect and time-to-fix incidents. Clear
versioning communicates intent and manages compatibility expectations.

## Additional Constraints & Requirements

Security: Security reviews MUST be performed for changes affecting authentication,
authorization, data storage, or external communication. Sensitive data MUST be handled
according to applicable laws and encrypted in transit and at rest when required.

Performance: Add performance targets on a per-feature basis; default targets MUST be
documented in the feature spec when latency or throughput is material to user value.

Technology: The team SHOULD prefer widely-supported, well-maintained frameworks and
avoid experimental stacks for production-critical components unless justified in the
complexity tracking section.

## Development Workflow & Quality Gates

Code Review: All changes MUST be reviewed by at least one other maintainer. Reviews
MUST include checks for tests, security implications, and compatibility with public
contracts.

Continuous Integration: CI MUST run linting, unit tests, contract tests, and
integration smoke tests for every PR. A PR MUST NOT be merged if CI failures are
present unless a documented exception is recorded.

Releases: Release notes MUST include a summary of breaking changes, migration steps,
and affected components for MAJOR and MINOR releases.

## Governance

This Constitution supersedes informal guidance. Amendments to the Constitution MUST be
documented in a PR with: a clear rationale, a migration plan for affected systems, and
an approval record (at least two maintainers). Amendments that remove or materially
redefine core principles are MAJOR and MUST include a stakeholder review.

Compliance: All plans and specs MUST include a Constitution Check and document any
accepted deviations in Complexity Tracking with justification and mitigations.

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): provide ratification date | **Last Amended**: 2025-10-06