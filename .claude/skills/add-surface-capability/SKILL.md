---
name: add-surface-capability
description: Use when adding or modifying an Application Use Case and exposing it consistently across all 3 Driving Adapters (CLI with --json, Interactive TUI, and MCP Server).
---

# Adding or Updating a Multi-Surface Capability

Every user-facing capability in CAKD Platform must be implemented once in `internal/usecase/` and exposed consistently across **CLI**, **TUI**, and **MCP Server**.

## Step 1: Implement or Update the Use Case (`internal/usecase/<capability>`)
1. Define typed `Request` and `Result` structs in `internal/usecase/<capability>/service.go` with JSON tags.
2. Inject required Outbound Ports (`internal/usecase/ports`) via the `NewService(...)` constructor.
3. Emit step-by-step progress using `progress.Emit(ctx, progress.Event{...})`.
4. Write unit tests in `service_test.go` using in-memory fakes/mocks of the ports.

## Step 2: Expose in the MCP Tool Catalog (`internal/usecase/agenttools`)
1. Register the capability in `internal/usecase/agenttools/catalog.go`:
   - Define the tool name (`cakd_<action>`), clear description, and JSON Schema `InputSchema`.
   - Wire the dispatcher to unmarshal JSON arguments, call the use case, and return the serialized JSON `Result`.
2. If the capability provides read-only reference data, also consider exposing an MCP Resource (`cakd://...`).

## Step 3: Expose in the CLI Adapter (`internal/adapter/cli`)
1. Create or update the Cobra command in `internal/adapter/cli/<command>.go`.
2. Support the `--json` flag:
   - When `--json` is false: attach a terminal `progress.Sink` to `ctx` and print a concise human-readable summary.
   - When `--json` is true: do not print intermediate text to `stdout`; marshal the use case's `Result` (or structured error) as JSON to `stdout`.
3. If the use case requires credentials, check `CredentialStore` in the CLI command first and prompt interactively only when `term.IsTerminal(os.Stdin)` and `--json` is false.

## Step 4: Expose in the TUI Adapter (`internal/adapter/tui`)
1. Add or update the corresponding form step or action screen in `internal/adapter/tui/`.
2. Connect the use case execution to the Bubble Tea program using a `progress.Sink` that forwards `progress.Event` messages to update the TUI progress view.

## Step 5: Verify All Surfaces
1. Run unit tests across `internal/usecase/...` and `internal/adapter/...`.
2. Verify CLI (`--json`), TUI model transitions, and MCP `tools/list` + `tools/call` tests pass.
