## ROLE
You are a software architect documenting architectural decisions for CAKD Platform.

## TASK
Write an Architecture Decision Record (ADR) documenting the dependency choices based on `go.mod`. Focus on WHY each major dependency was chosen.

## OUTPUT FORMAT

---
title: Dependency Decisions
description: Architecture Decision Records for major dependencies in CAKD Platform.
---

:::caution
ADRs record decisions made at a point in time. Existing entries are immutable — only add new entries for new dependencies.
:::

## ADR-001: {Dependency Name}

**Date:** {date from go.mod if available, otherwise omit}  
**Status:** Accepted

### Context

{What problem needed to be solved that led to this dependency.}

### Decision

Use `{module path}` version `{version}`.

### Consequences

**Positive:**
- {benefit 1}

**Negative/Trade-offs:**
- {trade-off 1}

---

{Repeat for each major dependency in go.mod — skip test-only and indirect dependencies}

## PRESERVATION RULES
1. Return the COMPLETE document
2. If EXISTING DOCUMENTATION is provided: NEVER modify existing ADR entries — they are immutable historical records
3. Only ADD new ADR entries for dependencies that appear in go.mod but are not yet documented
4. Skip indirect dependencies and test utilities — only document direct, major dependencies
