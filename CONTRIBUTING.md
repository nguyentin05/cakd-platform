# Contributing to CAKD Platform

First off, thank you for taking the time to contribute. CAKD Platform is built on the belief that cloud-native development should be accessible to small teams — and that only happens with a community behind it.

This document covers everything you need to know to contribute effectively: from filing your first bug report to adding a new cloud provider.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [What We're Looking For](#what-were-looking-for)
- [What We're NOT Looking For](#what-were-not-looking-for)
- [Security Disclosures](#security-disclosures)
- [Your First Contribution](#your-first-contribution)
- [Getting Started](#getting-started)
- [How to Report a Bug](#how-to-report-a-bug)
- [How to Suggest a Feature](#how-to-suggest-a-feature)
- [Development Guide](#development-guide)
- [Adding a New Provider](#adding-a-new-provider)
- [Adding a New Language Template](#adding-a-new-language-template)
- [Pull Request Process](#pull-request-process)
- [Commit Message Convention](#commit-message-convention)
- [Code Style](#code-style)

---

## Code of Conduct

By participating, you are expected to uphold basic respect for everyone involved. We do not tolerate harassment of any kind. If you experience or witness unacceptable behavior, report it by opening a private issue or emailing the maintainer directly.

---

## What We're Looking For

We welcome all kinds of contributions:

- **Bug fixes** — especially around template generation and provider integrations
- **New provider implementations** — VCS, IaC, notification, LLM providers
- **New language templates** — NodeJS, Python, Go application templates
- **Documentation improvements** — the docs site is fully AI-generated; improving the prompt templates counts as a contribution
- **Test coverage** — we currently have low coverage; any test is welcome
- **Registry updates** — new supported versions for backing services or runtimes

---

## What We're NOT Looking For

To set expectations clearly:

- **Platform.yaml breaking changes** without a migration path — open an issue first
- **New features without tests** — all new code must include unit tests
- **Direct commits to `main`** — all changes go through Pull Requests
- **Support questions as issues** — use GitHub Discussions for questions

---

## Security Disclosures

**Do NOT open a public issue for security vulnerabilities.**

CAKD Platform handles GitHub tokens, API keys, and Kubernetes credentials. If you find a vulnerability — especially around secret handling, credential leakage, or template injection — please report it privately.

Email: **security@[maintainer-domain]** or use [GitHub's private security advisory](https://github.com/nguyentin05/cakd-platform/security/advisories/new).

When reporting, include:
- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and aim to release a fix within 7 days for critical issues.

---

## Your First Contribution

Not sure where to start? Look for issues labeled:

- `good first issue` — small, well-scoped, a few lines of Go
- `help wanted` — more involved but clearly defined
- `docs` — improvements to prompt templates or documentation content

Both lists are sorted by comment count as a proxy for community interest.

**Never contributed to open source before?**
Check out [How to Contribute to an Open Source Project on GitHub](https://egghead.io/series/how-to-contribute-to-an-open-source-project-on-github) — a free beginner-friendly series.

> At this point, you're ready to make your changes. Feel free to ask for help — everyone starts somewhere.

---

## Getting Started

### Prerequisites

```bash
go 1.24+
terraform 1.14+
task (Taskfile runner)
golangci-lint
```

### Setup

```bash
# Fork and clone the repo
git clone https://github.com/nguyentin05/cakd-platform.git
cd cakd-platform

# Install dependencies
go mod download

# Verify everything works
task ci
```

`task ci` runs: format → vet → lint → test → build → trivy scan. All must pass before submitting a PR.

### Environment for integration testing

Copy `.env.example` to `.env` and fill in:

```bash
CAKD_GITHUB_TOKEN=ghp_xxxx   # repo + delete_repo scopes
CAKD_GEMINI_API_KEY=AIzaxxxx
CAKD_DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/xxxx
```

Never commit `.env`. It is gitignored.

---

## How to Report a Bug

Before filing, search existing issues to avoid duplicates.

When filing a bug, include:

1. **`cakd version` output** — which version are you on?
2. **OS and architecture** — `uname -a`
3. **Your `platform.yaml`** — remove any sensitive values
4. **What you did** — exact command you ran
5. **What you expected**
6. **What happened instead** — include full error output

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.yml).

---

## How to Suggest a Feature

CAKD Platform follows a clear design philosophy:

> A single `platform.yaml` should be enough to bootstrap a production-ready cloud-native project. Everything else is an implementation detail.

Before suggesting a feature, ask: *does this reduce cognitive load for the developer, or does it add to it?*

To suggest a feature:

1. Check the [roadmap](docs/src/content/docs/adrs/) to see if it's already planned
2. Open a GitHub Issue using the [feature request template](.github/ISSUE_TEMPLATE/feature_request.yml)
3. Describe: what problem does this solve, who benefits, how should it work
4. Tag it with the relevant area: `provider`, `template`, `registry`, `cli`, `observability`

For large features (new provider, new schema field), open an issue **before** writing code. This saves everyone time.

---

## Development Guide

### Project structure

```
cmd/cakd/          CLI command definitions (thin layer, no business logic)
internal/
  registry/        Single source of truth for supported providers and defaults
  config/          Schema + parser (references registry, no hardcoded values)
  provider/        Provider implementations (vcs, iac, cd, llm, notify, app)
  bootstrap/       Orchestrators for create and init commands
  template/        Template engine + embedded .tmpl files
  observe/         Observability command logic
  agent/           AI webhook handler
```

### Key conventions

**Registry is the only source of truth:**
All supported values and defaults live in `internal/registry/`. The parser never hardcodes strings. If you add a new provider or default value, it goes in `registry/providers.go` or `registry/defaults.go`.

**Provider pattern:**
Every external integration is a provider implementing an interface in `internal/provider/interfaces.go`. The factory in `internal/factory/factory.go` maps string names to implementations.

**Template delimiters:**
All Go templates use `[[ ]]` instead of `{{ }}` to avoid conflicts with Helm and GitHub Actions syntax.

**Secrets never in platform.yaml:**
The platform.yaml schema must never contain credential fields. Credentials resolve via: env vars → `~/.cakd/credentials` → interactive prompt.

### Useful tasks

```bash
task build              # Build CLI binary
task test               # Run all tests with race detector
task lint               # Run golangci-lint
task generate           # Run bootstrap with sample platform.yaml
task validate           # Validate sample platform.yaml
task trivy:config       # Scan generated IaC/Helm output
task ci                 # Run full CI suite locally
```

---

## Adding a New Provider

Example: adding Slack as a notification provider.

**Step 1 — Register it:**

```go
// internal/registry/providers.go
Notification: []string{"discord", "slack"},  // add "slack"
```

**Step 2 — Add defaults if needed:**

```go
// internal/registry/defaults.go
// No change needed unless Slack has specific defaults
```

**Step 3 — Implement the interface:**

```go
// internal/provider/notify/slack/client.go
package slack

type Client struct { ... }

func (c *Client) Send(alert Alert) error { ... }
```

**Step 4 — Register in factory:**

```go
// internal/factory/factory.go
case "slack":
    return slack.New(webhookURL), nil
```

**Step 5 — Add credentials mapping:**

```go
// Resolution: CAKD_SLACK_WEBHOOK_URL → ~/.cakd/credentials slack.webhook_url
```

**Step 6 — Write tests:**

```
internal/provider/notify/slack/client_test.go
```

**Step 7 — Update registry docs:**
The docs will auto-regenerate on next CI run via the docs agent.

---

## Adding a New Language Template

Example: adding `nodejs-nestjs`.

**Step 1 — Register the language:**

```go
// internal/registry/providers.go
Languages: []string{"java-spring-boot", "nodejs-nestjs"},
```

**Step 2 — Add template files:**

```
internal/template/templates/nodejs-nestjs/
  Dockerfile.tmpl
  package.json.tmpl
  src/main.ts.tmpl
  src/app.module.ts.tmpl
```

All templates use `[[ ]]` delimiters. Available variables come from `*config.PlatformConfig`.

**Step 3 — Wire into template engine:**

```go
// internal/template/engine.go
case "nodejs-nestjs":
    return e.renderNodeJSNestJS(cfg, outputDir)
```

**Step 4 — Add default versions:**

```go
// internal/registry/defaults.go
NodeJS: "20",  // LTS
```

**Step 5 — Write a test fixture:**

```
test/fixtures/valid-nodejs-nestjs.yaml
```

And add an integration test verifying the generated output compiles.

---

## Pull Request Process

1. **Fork** the repo and create a branch from `main`
2. **Write tests** for any new behavior
3. **Run `task ci`** locally — all checks must pass
4. **Open a PR** against `main` with a clear description
5. **Reference the issue** your PR addresses: `Closes #123`
6. **Wait for review** — maintainers aim to review within 5 business days

PRs that break existing tests, skip tests for new code, or don't follow the commit convention will not be merged until fixed.

For changes to `internal/registry/` or `internal/config/schema.go` — these affect the entire platform and require extra scrutiny. Expect more back-and-forth.

### Obvious fixes

Small changes — typos, comment cleanup, formatting — can skip the issue step and go straight to a PR. Label it `fix: typo` or similar.

---

## Commit Message Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) **without scope**:

```
<type>: <description>

feat: add slack notification provider
fix: correct storage default for redis backing service
chore: update supported postgresql versions in registry
docs: improve quickstart tutorial steps
test: add unit tests for validateStructure
refactor: move argocd into provider/cd package
```

**Allowed types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

**Rules:**
- Subject line max 100 characters
- Lowercase — no `Feat:` or `FEAT:`
- No period at the end
- Present tense — "add" not "added"

Commit messages are enforced by `commitlint` on every PR. Non-compliant commits will block merge.

Semantic Release uses these commits to determine the next version:
- `feat:` → minor bump
- `fix:` → patch bump
- `feat!:` or `BREAKING CHANGE:` in footer → major bump

---

## Code Style

We use `golangci-lint` with the configuration in `.golangci.yml`. Run `task lint` before pushing.

Key rules enforced:
- `errcheck` — every error must be handled
- `gosec` — no hardcoded secrets, safe file operations
- `gocyclo` — max cyclomatic complexity 15
- `gofmt` + `goimports` — formatting is not negotiable

One rule explicitly disabled:
- `G304` (file path as taint input) — CLI tools must accept user-provided paths

For test files, `gosec` and `errcheck` are relaxed — test code has different conventions.

If `golangci-lint` flags something you believe is a false positive, add an inline `//nolint:rulename // reason` comment explaining why. Don't suppress entire files.

---

*Thank you for contributing to CAKD Platform. Every improvement — no matter how small — helps make cloud-native development more accessible.*