---
title: Dependency Decisions
description: Architecture Decision Records for major dependencies in CAKD Platform.
---

:::caution
ADRs record decisions made at a point in time. Existing entries are immutable — only add new entries for new dependencies.
:::

## ADR-001: Cobra

**Status:** Accepted

### Context

The `cakd` binary is a command-line interface (CLI) tool. A robust and idiomatic framework was required to define commands, subcommands, flags, and generate help messages efficiently. This framework needed to support a structured and extensible CLI architecture.

### Decision

Use `github.com/spf13/cobra v1.10.2`.

### Consequences

**Positive:**
- Provides a standardized and widely adopted structure for CLI applications in Go.
- Simplifies the definition of commands, subcommands, and flags.
- Automatically generates comprehensive help messages, improving user experience.
- Facilitates easy parsing of command-line arguments and flags.

**Negative/Trade-offs:**
- Introduces an external dependency to the project.
- New contributors unfamiliar with Cobra may require a brief learning period.

---

## ADR-002: golang.org/x/term

**Status:** Accepted

### Context

The `cakd` CLI may require interactive input from users, particularly for sensitive information such as passwords or API keys, during operations like `cakd create`. It is critical to handle such input securely, preventing it from being echoed to the terminal.

### Decision

Use `golang.org/x/term v0.29.0`.

### Consequences

**Positive:**
- Enables secure handling of sensitive user input by masking characters (e.g., for passwords).
- Provides cross-platform capabilities for interacting with the terminal, ensuring consistent behavior.
- Offers utilities for setting raw terminal mode, useful for advanced interactive prompts.

**Negative/Trade-offs:**
- Adds a dependency, although it is a standard Go "x" package.
- Its functionality is specific to terminal interaction and may not be frequently used across all `cakd` commands.

---

## ADR-003: YAML v3

**Status:** Accepted

### Context

The `cakd` platform relies heavily on YAML files for configuration, notably `platform.yaml`. A reliable and performant library is necessary to unmarshal (read) and marshal (write) YAML data structures into Go structs, ensuring accurate and consistent handling of platform configurations.

### Decision

Use `gopkg.in/yaml.v3 v3.0.1`.

### Consequences

**Positive:**
- Provides robust and efficient serialization and deserialization of YAML data.
- Widely used and actively maintained, ensuring stability and ongoing support.
- Supports the latest YAML specification features, allowing for complex configuration structures.

**Negative/Trade-offs:**
- Introduces an external dependency for YAML processing.
- Potential for minor API differences compared to `yaml.v2` if a migration were ever considered.

---