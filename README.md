## Cloud-agnostic Kubernetes Developer Platform

A CLI tool that scaffolds production-ready cloud-native microservices — from a single YAML file.

---

<div align="center">
  <img src="docs/src/assets/header-github.png" width="900" alt="CAKD Platform"/>
</div>

---

> **TL;DR:** CAKD eliminates the boilerplate of bootstrapping cloud-native projects. Declare your service in
`platform.yaml`, and CAKD generates a fully functional project with Helm charts, CI/CD pipelines, ArgoCD configuration,
> Dockerfile, and Terraform infrastructure. Plus, it features an AI-powered observability tool for instant Kubernetes
> debugging.

## Core Capabilities

- **Convention over Configuration** — A single file platform.yaml run the entire scaffold.
- **GitOps-Ready** — Generates ArgoCD Application manifests and GitHub Actions workflows out of the box.
- **Multi-Language Templates** — Currently supports Java Spring Boot, with more planned.
- **Helm Chart Generation** — Production-ready chart with Deployment and Service templates.
- **Terraform Integration** — Provisions GitHub repositories and related infrastructure automatically.
- **AI Diagnosis** — Collects metrics and logs, then queries AI Chatbot Provider for root cause analysis and actionable
  remediation steps.

---

## Documentation

Our modern documentation site includes:

- **Tutorials**: Step-by-step guides for beginners.
- **How-to Guides**: Goal-oriented recipes.
- **Explanation**: Deep dives into CAKD architecture.
- **Reference**: Full CLI and platform.yaml specifications.
- **ADRs**: Architectural Decision Records tracking the evolution of the platform.

## Quick Start

### Prerequisites

- Go 1.21+
- Terraform
- A Kubernetes cluster with Prometheus and Loki
- A Google Gemini API key (for AI diagnosis)

### 1. Installation

<details>
<summary><b>From Source</b></summary>

```bash
git clone https://github.com/nguyentin05/cakd-platform.git
cd cakd-platform
task build
```

</details>

<details open>
<summary><b>Quick Install (macOS / Linux)</b></summary>

Install the latest release directly to /usr/local/bin using our auto-install script:

```bash
curl -sSL https://raw.githubusercontent.com/nguyentin05/cakd-platform/main/scripts/install.sh | bash
```

</details>

<details>
<summary><b>From Release Binaries (Manual)</b></summary>

Download the latest binary for your platform from
the [Releases page](https://github.com/nguyentin05/cakd-platform/releases/latest).

```bash
chmod +x cakd
sudo mv cakd /usr/local/bin/
```

</details>

### 2. Scaffold a Project

Create a `platform.yaml` in your project root:

```yaml
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: my-app
  owner: your-github-username
spec:
  language: java-spring-boot
  version: "17"
  features:
    monitoring: true
    alerting: true
  dependencies:
    database:
      type: postgresql
      version: "15"
      storage: 5Gi
```

Run the scaffold command:

```bash
cakd create
```

## AI Observability

`cakd observe` connects to your cluster's Prometheus and Loki instances, collects current metrics and logs for a given
namespace, and sends them to Google Gemini for analysis.

```bash
export GEMINI_API_KEY=your_api_key
cakd observe --namespace my-app
```

<details>
<summary><b>View Example AI Output</b></summary>

```text
Fetching metrics from Prometheus...
Fetching logs from Loki...
Sending data to Google Gemini for AI diagnosis...

==========================================
CAKD AI DIAGNOSIS
==========================================
## Diagnosis

**Root Cause:** The `my-app` pod is in a CrashLoopBackOff state due to a
missing environment variable `DATABASE_URL`.

## Solution
1. Verify your Helm values include the `DATABASE_URL` secret reference.
2. Apply the updated chart: `helm upgrade my-app ./helm`
==========================================
```

</details>

## Development

This project uses Taskfile as the task runner.

```bash
task build     # Compile binary
task test      # Run tests
task lint      # Run golangci-lint
task --list    # List all available tasks
```

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a pull request.
This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification. All commits must
follow the format `type: description`.

## License

Licensed under the [Apache License 2.0](LICENSE).