# `platform.yaml` Schema Reference

CAKD Platform uses the canonical `cakd.io/v1alpha1` schema (with automatic backward compatibility for legacy `platform.dev/v1alpha1` files). The machine-readable Draft 2020-12 JSON Schema is located at `schemas/platform.schema.json` and can be exported at any time via `cakd-cli schema`.

## Full Example

```yaml
apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: order-system
  owner: my-org
  environment: production
spec:
  cluster:
    provider: aws-ec2
    region: ap-southeast-1
    instanceType: t3.medium
    diskSize: 30
    addons:
      argocd: true
      monitoring: true
      logging: true
      agent: true
  providers:
    versionControl: github
    ci: github-actions
    cd: argocd
    notification: discord
    llm: gemini
    monitoring: prometheus
    logging: loki
  observability:
    alerting: true
    ai:
      model: gemini-2.5-flash
  backing:
    - name: orders-db
      type: postgresql
      version: "16"
      storage: 5Gi
  services:
    - name: order-api
      language: java-spring-boot
      version: "21"
      projectBuild: maven-project
      packaging: jar
      configuration: yaml
      dependencies:
        - web
        - actuator
      replicas: 2
      resources:
        cpu: 500m
        memory: 512Mi
      uses:
        - orders-db
```
