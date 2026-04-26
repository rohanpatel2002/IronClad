# IRONCLAD Production Runbook

This guide covers operational best practices, deployment, and troubleshooting for the IRONCLAD security platform in a production environment.

## 1. System Architecture

IRONCLAD consists of 4 microservices:
1. **gate-go (Port 8080):** API Gateway and webhook receiver.
2. **topology-go (Port 8081):** Kubernetes dependency graph builder.
3. **semantic-python (Port 8082):** LLM-based intent analysis.
4. **scoring-python (Port 8083):** 3-axis risk scorer and failure grammar engine.

## 2. Observability

### Prometheus & Grafana
- **Prometheus** runs on port `9090` and scrapes metrics every 10s.
- **Grafana** runs on port `3000` (`admin` / `ironclad`).
- Pre-built dashboard `ironclad-prod.json` is automatically provisioned and tracks:
  - HTTP Request Rates
  - Latency (p95 / p99)
  - Blast Radius BFS Traversals
  - Decision Counters (ALLOW / WARN / BLOCK)

### Distributed Tracing
- All requests entering `gate-go` are assigned an `X-Request-ID`.
- This ID is propagated into logs using `slog` for structured JSON logging.

### Circuit Breakers
- `gate-go` uses `gobreaker` to wrap downstream calls.
- View real-time status at `GET /api/v1/circuit-breaker/status`.
- If a circuit breaker is `open`, downstream requests fail-fast to prevent cascading failure. It automatically enters `half-open` after a timeout.

## 3. GitHub Webhook Integration

1. Configure your GitHub App or Repo Webhook to point to `https://<your-domain>/api/v1/webhooks/github`.
2. Select **Pull requests** events.
3. Set a secure webhook secret and export it as `GITHUB_WEBHOOK_SECRET` on `gate-go`.
4. Ensure `GITHUB_TOKEN` is set on `gate-go` for it to fetch PR diffs and post decision comments.

## 4. Kubernetes Topology Discovery

- `topology-go` automatically discovers the dependency graph by reading K8s `Service` objects.
- Ensure the pod running `topology-go` has a ServiceAccount with RBAC permissions to `get` and `list` Services across all namespaces.
- Add annotations to your K8s services:
  - `ironclad.security/depends-on: "service-a,service-b"`
  - `ironclad.security/criticality: "0.95"`

## 5. Troubleshooting & Incident Response

### `gate-go` is rejecting webhooks with 401 Unauthorized
**Cause:** HMAC signature mismatch.
**Fix:** Verify `GITHUB_WEBHOOK_SECRET` matches exactly between GitHub and the `gate-go` environment variable.

### `semantic-python` is timing out
**Cause:** Anthropic API latency.
**Fix:** `gate-go` has exponential backoff retries. If it still fails, the circuit breaker will trip and fallback logic will return a conservative risk score. Ensure `ANTHROPIC_API_KEY` is valid.

### `topology-go` returns a stale graph
**Cause:** K8s API connectivity issue.
**Fix:** The `K8sGraphBuilder` caches the graph for 5 minutes. If it cannot reach K8s during a refresh, it emits a warning log and continues serving the stale cache. Check K8s network policies and ServiceAccount RBAC.
