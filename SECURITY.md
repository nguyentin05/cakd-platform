# Security Policy

## Supported Versions

We actively support security updates for the following versions of CAKD Platform (including the `cakd` CLI and
`cakd-agent` daemon):

| Version | Supported          |
|---------|--------------------|
| 0.11.x  | :white_check_mark: |
| 0.10.x  | :white_check_mark: |
| < 0.10  | :x:                |

## Reporting Vulnerabilities

We take the security of our developer platform and in-cluster agents seriously. If you discover a security
vulnerability, please follow these steps:

### 1. Private Disclosure

Please **DO NOT** create a public GitHub issue for security vulnerabilities, as it exposes clusters running the agent.
Instead:

- Use GitHub's [Security Advisory](https://github.com/nguyentin05/cakd-platform/security/advisories) feature to report
  privately.
- Mention `@nguyentin05` in the advisory for immediate triage.

### 2. What to Include

When reporting a vulnerability, please include:

- **Component Affected**: Specify if it impacts the `cakd` CLI, the `cakd-agent` daemon, or the scaffolded templates.
- **Description**: Detailed explanation of the vulnerability.
- **Steps to Reproduce**: Step-by-step instructions, sample configuration YAML, or a Proof of Concept.
- **Impact**: Potential impact assessment (e.g., Privilege escalation inside the cluster, credential leaks from
  `auth login`, or prompt/token injection).
- **Mitigation**: Any suggested fixes or temporary workarounds.

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Critical cluster/agent exploits within 14 days, other issues within 60 days.

## Security Features

CAKD Platform integrates rigid security mechanisms natively into its architecture and development workflows:

- **Secure Credential Management**: The `cakd auth` command securely manages tokens without exposing plain-text keys in
  shell histories.
- **Automated Security Pipelines**: Our `task ci` workflow incorporates automated security scans on every change.
- **Static Application Security Testing**: Automated CodeQL scans analyze the Go-based CLI/agent and Java Spring Boot
  scaffolding templates for vulnerabilities.
- **Supply Chain Security**: Dependabot and automated dependency reviews ensure all Go modules, npm documentation
  packages, and generated manifests remain free of known CVEs.

## Security Best Practices for Users & Contributors

When working with or contributing to CAKD Platform:

1. **Least Privilege**: Ensure the service accounts and RBAC roles provisioned for `cakd-agent` have only the minimum
   permissions required to scrape Prometheus/Loki logs.
2. **Secrets Handling**: Never commit `platform.yaml` files containing production secrets or raw credentials. Always
   rely on CAKD's encrypted auth manager or environment variables.
3. **Input Sanitization**: All inputs parsed from configuration YAMLs or received by the `cakd-agent` webhook dispatcher
   must be strictly validated to avoid command injection into the host cluster.
4. **Error Handling**: The AI diagnostic engine `cakd observe` and Discord notification router must sanitize system
   internals, preventing raw cluster secrets from leaking into external webhook embeds.

## Security Updates & Fixes Log

We actively patch vulnerabilities and roll out fixes in our minor/patch updates.

### Recent Security Fixes

- **v0.11.0**: Enhanced security controls, client connection timeouts, and automatic git token redaction safeguards.
- **v0.10.0**: Introduced the secure 3-layer secret management fallback engine and interactive auth CLI.
- **v0.9.3**: Corrected Dependabot configuration and integrated automated patch dependency workflows.

## Contact

For security-related questions or architectural privacy concerns, please contact:

- GitHub Security Advisories: [Create Advisory](https://github.com/nguyentin05/cakd-platform/security/advisories)
- Project Maintainer: `@nguyentin05`

---

_This security policy is reviewed and updated regularly to ensure it meets cloud-native and Kubernetes security
standards._
