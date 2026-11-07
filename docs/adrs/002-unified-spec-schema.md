# ADR 002: Unified `cakd.io/v1alpha1` Spec Schema & Isolated State

- **Status**: Accepted
- **Date**: 2026-11-07

## Context

Previously, `platform.yaml` placed `infra`, `providers`, `observability`, `backing`, and `services` at the root level alongside `metadata`, and shared a single `.platform/terraform/` directory between `cakd init` (AWS EC2) and `cakd create` (GitHub repository), causing Terraform state collisions.

## Decision

1. Standardize on Kubernetes-style `apiVersion: cakd.io/v1alpha1`, `kind: Project`, `metadata`, and top-level `spec:` block while maintaining a backward-compatible normalizer for legacy `platform.dev/v1alpha1` files.
2. Isolate Terraform states into `.platform/terraform/cluster/` and `.platform/terraform/vcs/`, passing cluster metadata via `ports.ClusterStateStore`.
