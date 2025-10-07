```markdown
# Feature Specification: Earliest-daily-meter CLI

**Feature Branch**: `001-a-cli-written`  
**Created**: 2025-10-06  
**Status**: Draft  
**Input**: User description: "a CLI written in Go. The database it talks to is Postgres. A config with database settings should be provided through a flag.

The goals of the CLI is to request the earliest gas meter reading of the day. No matter when it runs on the day, it should always return the earlierst read of the day.

Once that's done, the data should be send using a POST request in JSON format to a URL."

## Execution Flow (main)
```
1. Parse user description from Input
   → If empty: ERROR "No feature description provided"
2. Extract key concepts from description
   → Identify: actors, actions, data, constraints
3. For each unclear aspect:
   → Mark with [NEEDS CLARIFICATION: specific question]
4. Fill User Scenarios & Testing section
   → If no clear user flow: ERROR "Cannot determine user scenarios"
5. Generate Functional Requirements
   → Each requirement must be testable
   → Mark ambiguous requirements
6. Identify Key Entities (if data involved)
7. Run Review Checklist
   → If any [NEEDS CLARIFICATION]: WARN "Spec has uncertainties"
   → If implementation details found: ERROR "Remove tech details"
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure) — note: implementation details below are included to make this actionable for engineers; they should be condensed or moved to the plan if non-essential to the spec.

### Section Requirements
- **Mandatory sections**: Completed below

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story
As an operational or automation actor, I want a CLI that finds the earliest gas meter reading recorded for "today" and delivers that reading to a configured endpoint, so that downstream systems can process the daily baseline reading regardless of when the CLI runs.

### Acceptance Scenarios
1. **Given** the database contains multiple meter readings for today, **When** the CLI runs at any time on that day, **Then** it MUST return the reading with the earliest timestamp that is within the local day boundary (00:00:00 → 23:59:59) for the target meter and send it via POST to the configured URL.
2. **Given** there is a single reading for today, **When** the CLI runs, **Then** it MUST return and POST that reading.
3. **Given** there are no readings for today, **When** the CLI runs, **Then** it MUST return an explicit, machine-readable error status and log the absence; the POST MUST NOT be attempted (or must include a clear empty-result payload if that is an agreed behavior).
4. **Given** the POST endpoint returns a non-2xx response, **When** the CLI posts the JSON, **Then** it MUST retry according to a configurable retry policy and surface a non-zero exit code if final delivery fails.

### Edge Cases
- Readings timestamp type: readings are stored in Postgres as `timestamptz` and look like `2025-10-06 19:47:02.028 +0200`. All servers and clients use the `Europe/Amsterdam` timezone. The day boundary MUST be interpreted in `Europe/Amsterdam` local time.
- Network outages: POST may fail — define retry/backoff policy (default to 3 attempts with exponential backoff).

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: The CLI MUST accept database connection settings via a flag (e.g., `--db-url` or `--db-host` + other flags) and allow overriding by environment variables.
- **FR-002**: The CLI MUST connect to a Postgres database and query the meter readings table for readings that fall within the current day (per configured timezone) for the specified meter(s).
- **FR-003**: The CLI MUST identify and return the earliest reading for the day (minimum timestamp within the day's boundaries).
- **FR-004**: The CLI MUST serialize the reading into JSON and send it in the body of a POST request to a configurable URL (`--post-url` flag).
- **FR-005**: The CLI MUST return an exit code of 0 on successful send (2xx response) and non-zero on unrecoverable errors (DB connection failed, no readings for the day (unless policy says otherwise), or POST delivery failure after retries).
- **FR-006**: The CLI MUST log structured events for: startup, DB query executed (including query time and row count), earliest reading selected (timestamp, value, id), POST attempt and response codes, retries, and final outcome.
- **FR-007**: The CLI SHOULD support a dry-run mode (`--dry-run`) where the JSON payload is constructed and emitted to stdout but not POSTed.
- **FR-008**: The CLI MUST be written in Go (user-specified). Implementation details such as language are recorded here for planning but not for high-level product sign-off.

*Ambiguities/assumptions made*: timezone = `Europe/Amsterdam` (Postgres `timestamptz`). Retry policy default: 3 attempts with exponential backoff unless overridden by flags.

### Key Entities
- **MeterReading**: represents a recorded meter read. Key attributes: id (UUID/int), meter_id, timestamp (ISO-8601, stored in DB column), value (numeric), metadata (optional JSON).
- **DeliveryEvent**: record of POST attempts (timestamp, payload, response code, attempt_count) — may be logged or recorded in DB depending on implementation choice.

---

## Review & Acceptance Checklist


### Content Quality
- [x] No implementation details that block stakeholder review (implementation notes included for planning)
- [x] Focused on user value and why this is required (daily earliest reading delivery)
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status

- [x] User description parsed
- [x] Key concepts extracted
- [ ] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [ ] Review checklist passed

---

``` 
# Feature Specification: [FEATURE NAME]

**Feature Branch**: `[###-feature-name]`  
**Created**: [DATE]  
**Status**: Draft  
**Input**: User description: "$ARGUMENTS"

## Execution Flow (main)
```
1. Parse user description from Input
   → If empty: ERROR "No feature description provided"
2. Extract key concepts from description
   → Identify: actors, actions, data, constraints
3. For each unclear aspect:
   → Mark with [NEEDS CLARIFICATION: specific question]
4. Fill User Scenarios & Testing section
   → If no clear user flow: ERROR "Cannot determine user scenarios"
5. Generate Functional Requirements
   → Each requirement must be testable
   → Mark ambiguous requirements
6. Identify Key Entities (if data involved)
7. Run Review Checklist
   → If any [NEEDS CLARIFICATION]: WARN "Spec has uncertainties"
   → If implementation details found: ERROR "Remove tech details"
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

### Section Requirements
- **Mandatory sections**: Must be completed for every feature
- **Optional sections**: Include only when relevant to the feature
- When a section doesn't apply, remove it entirely (don't leave as "N/A")

### For AI Generation
When creating this spec from a user prompt:
1. **Mark all ambiguities**: Use [NEEDS CLARIFICATION: specific question] for any assumption you'd need to make
2. **Don't guess**: If the prompt doesn't specify something (e.g., "login system" without auth method), mark it
3. **Think like a tester**: Every vague requirement should fail the "testable and unambiguous" checklist item
4. **Common underspecified areas**:
   - User types and permissions
   - Data retention/deletion policies  
   - Performance targets and scale
   - Error handling behaviors
   - Integration requirements
   - Security/compliance needs

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story
[Describe the main user journey in plain language]

### Acceptance Scenarios
1. **Given** [initial state], **When** [action], **Then** [expected outcome]
2. **Given** [initial state], **When** [action], **Then** [expected outcome]

### Edge Cases
- What happens when [boundary condition]?
- How does system handle [error scenario]?

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: System MUST [specific capability, e.g., "allow users to create accounts"]
- **FR-002**: System MUST [specific capability, e.g., "validate email addresses"]  
- **FR-003**: Users MUST be able to [key interaction, e.g., "reset their password"]
- **FR-004**: System MUST [data requirement, e.g., "persist user preferences"]
- **FR-005**: System MUST [behavior, e.g., "log all security events"]

*Example of marking unclear requirements:*
- **FR-006**: System MUST authenticate users via [NEEDS CLARIFICATION: auth method not specified - email/password, SSO, OAuth?]
- **FR-007**: System MUST retain user data for [NEEDS CLARIFICATION: retention period not specified]

### Key Entities *(include if feature involves data)*
- **[Entity 1]**: [What it represents, key attributes without implementation]
- **[Entity 2]**: [What it represents, relationships to other entities]

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [ ] No implementation details (languages, frameworks, APIs)
- [ ] Focused on user value and business needs
- [ ] Written for non-technical stakeholders
- [ ] All mandatory sections completed

### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain
- [ ] Requirements are testable and unambiguous  
- [ ] Success criteria are measurable
- [ ] Scope is clearly bounded
- [ ] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [ ] User description parsed
- [ ] Key concepts extracted
- [ ] Ambiguities marked
- [ ] User scenarios defined
- [ ] Requirements generated
- [ ] Entities identified
- [ ] Review checklist passed

---
