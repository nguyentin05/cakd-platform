# `cakd-agent` Daemon

`cakd-agent` is the in-cluster HTTP webhook receiver for Prometheus Alertmanager.

## Endpoints

- `GET /healthz` — Liveness and readiness check (`{"status":"ok"}`).
- `POST /api/v1/alerts` — Accepts Alertmanager webhook payloads (authenticated via `Authorization: Bearer <CAKD_AGENT_SECRET>` or `?secret=<CAKD_AGENT_SECRET>`).
