# Hexagonal Architecture (Ports & Adapters)

CAKD Platform follows a strict **Hexagonal Architecture (Ports & Adapters)** enforced continuously by an automated Go AST boundary test (`internal/platform/architecture_test.go`).

```mermaid
flowchart TB
    subgraph Driving ["Driving Adapters (Inbound)"]
        CLI["internal/adapter/cli"]
        TUI["internal/adapter/tui"]
        MCP["internal/adapter/mcpserver"]
        WH["internal/adapter/webhook"]
    end

    subgraph App ["Application Use Cases & Ports"]
        UC["internal/usecase/*"]
        PORTS["internal/usecase/ports"]
    end

    subgraph Domain ["Pure Domain Core"]
        DOM["internal/domain/{project,catalog,diagnostic,errcode}"]
    end

    subgraph Driven ["Driven Adapters (Outbound)"]
        CFG["internal/infrastructure/{configfile,credstore,statestore}"]
        IAC["internal/infrastructure/{terraform,k8s,scaffold}"]
        PROV["internal/infrastructure/providers/*"]
    end

    CLI --> UC
    TUI --> UC
    MCP --> UC
    WH --> UC
    UC --> DOM
    UC --> PORTS
    PORTS --> DOM
    CFG -.-> PORTS
    IAC -.-> PORTS
    PROV -.-> PORTS
```

## Core Architectural Invariants

1. **Inward-Only Dependency Direction**:
   - `internal/domain/*` imports only Go standard library packages or sibling `internal/domain/*` packages.
   - `internal/usecase/*` depends only on `internal/domain/*`, `internal/usecase/ports`, and `internal/platform/*`.
   - `internal/adapter/*` (Driving) never imports `internal/infrastructure/*` (Driven) directly.
2. **Zero Terminal I/O Coupling in Core Layers**:
   - Interactive credential prompting (`term.ReadPassword`, `os.Stdin`) exists exclusively in `internal/adapter/cli` and `internal/adapter/tui`.
   - Progress reporting uses `progress.Sink` on `context.Context`.
3. **Isolated Terraform State Directories**:
   - Cluster infrastructure state is stored at `out/<project>/.platform/terraform/cluster/`.
   - GitHub VCS repository state is stored at `out/<project>/.platform/terraform/vcs/`.
   - Both stages communicate via `ports.ClusterStateStore` (`~/.cakd/cluster-state.json`).
