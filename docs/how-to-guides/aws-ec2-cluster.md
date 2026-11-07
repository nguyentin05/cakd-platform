# Provision an AWS EC2 K3s Cluster

CAKD Platform can provision a single-node K3s cluster on AWS EC2 using Terraform and configure Traefik Ingress with automatic `<public_ip>.nip.io` domains.

## 1. Declare `spec.cluster` in `platform.yaml`

```yaml
apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: order-system
  owner: my-org
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
  services:
    - name: order-api
      language: java-spring-boot
```

## 2. Run `cakd init`

```bash
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="..."
./bin/cakd init -f platform.yaml
```

CAKD provisions the EC2 instance in `.platform/terraform/cluster/`, configures a 2GB swap file, installs K3s with `--tls-san <public_ip>`, writes `~/.kube/cakd-config`, and saves the cluster state (`~/.cakd/cluster-state.json`) for subsequent `cakd create` runs.
