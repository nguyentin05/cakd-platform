---
trigger: always_on
description: Unified Project configuration schema (cakd.io/v1alpha1), backward-compatibility normalization, and isolated IaC/Cluster state storage rules.
---

# Rule 02: Configuration Schema & State Isolation

## 1. Canonical Project Configuration (`cakd.io/v1alpha1`)

All project configurations normalize into the single domain aggregate `project.Project` (`internal/domain/project`), structured under a top-level `spec:` block:

```yaml
apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: <project-name>
  owner: <github-owner>
spec:
  cluster:          # Optional cluster provisioning & addon settings (used by initcluster)
  providers:        # Selected driven adapter integrations (vcs, ci, cd, notification, llm, monitoring, logging)
  observability:    # Alerting & AI model configuration
  backing:          # Stateful dependencies (postgresql, redis, rabbitmq)
  services:         # Application microservices and their `uses` references
```

## 2. Backward Compatibility with `platform.dev/v1alpha1`

- `internal/infrastructure/configfile` MUST support reading legacy manifests (`apiVersion: platform.dev/v1alpha1` or manifests where `infra`, `providers`, `observability`, `backing`, `services` are placed at the YAML root) and automatically normalize them into `project.Project.Spec`.
- Domain and Use Case layers MUST ONLY operate on the normalized `project.Project` struct and never parse raw YAML or inspect legacy fields directly.
- Any structural change to `project.Project` MUST be followed by regenerating `schemas/platform.schema.json` via `go run ./cmd/cakd-tools/schema-gen`.

## 3. Strict State Directory Isolation

1. **Cluster Provisioning State**: Terraform modules that provision cluster hosts (such as `aws_ec2`) MUST execute in and store state under `<workDir>/.platform/terraform/cluster/`.
2. **VCS / Project Provisioning State**: Terraform modules that provision project repositories and secrets (such as `github`) MUST execute in and store state under `<workDir>/.platform/terraform/vcs/`.
3. **No Raw `terraform.tfstate` Coupling in Scaffold**: Template rendering (`internal/infrastructure/scaffold`) MUST NEVER read `terraform.tfstate` directly. Cluster public IP, domain, and kubeconfig paths MUST be persisted and retrieved via the `ports.ClusterStateStore` interface (`internal/infrastructure/statestore`).
