# Setup Discord Alerts & AI Diagnostics

CAKD Platform integrates Prometheus Alertmanager, `cakd-agent`, Google Gemini, and Discord to deliver two-stage incident notifications:

1. **Immediate Alert Notification** — Raw firing/resolved alerts are posted immediately to the project's Discord channel.
2. **AI Root-Cause Diagnosis** — Firing alerts are analyzed by Gemini and posted as a follow-up `🤖 CAKD AI Diagnosis` message.

## 1. Configure Credentials

```bash
cakd-cli auth set \
  --discord-token "YOUR_DISCORD_BOT_TOKEN" \
  --discord-guild "YOUR_DISCORD_GUILD_ID" \
  --gemini-key "YOUR_GEMINI_API_KEY"
```

## 2. Enable Observability in `platform.yaml`

```yaml
apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: order-system
  owner: my-org
spec:
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
  services:
    - name: order-api
      language: java-spring-boot
```
