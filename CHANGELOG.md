# Changelog

All notable changes to the IRONCLAD security platform are documented in this file.

## [1.2.0] - 2026-09-06

### Added
- **Topology Service**: Graph cycle detection, max depth analysis, and SLO error budget monitoring (`services/topology-go`).
- **Scoring Engine**: ML predictor fallback with confidence bounds, failure event queue consumer, and dynamic weight configuration (`services/scoring-python`).
- **Gate Security**: Leaky bucket & dynamic resource-aware rate limiters, SOAR quarantine engine, distributed mutex locking, and WAF input inspection (`services/gate-go`).
- **Semantic Service**: Diff embedding vector generator, intent classification endpoints, and test suite (`services/semantic-python`).
- **Dashboard UI**: Next.js React dashboard components for Risk Cards, Topology Graphs, Audit Streams, and Threat Maps (`apps/dashboard`).
- **Infrastructure & Docs**: Open Policy Agent Rego policies, Prometheus alerts, Grafana KPI dashboards, Kubernetes zero-trust NetworkPolicies, and production runbooks (`infra/`).
