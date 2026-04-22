# IRONCLAD

> **A deployment gate that understands intent, not just syntax.**

IRONCLAD is a semantic deployment risk engine that sits in front of CI/CD promotion and answers one critical question before every release:

**“Does what this code is trying to do match what this system can safely absorb right now?”**

Traditional pipelines validate _correctness of code_. IRONCLAD validates _safety of deployment intent_ against real-world production context.

---

## Table of Contents

- [Why IRONCLAD exists](#why-ironclad-exists)
- [What IRONCLAD does](#what-ironclad-does)
- [Core risk model](#core-risk-model)
- [System architecture](#system-architecture)
- [Tech stack](#tech-stack)
- [Planned repository layout](#planned-repository-layout)
- [Data model (high-level)](#data-model-high-level)
- [Deployment decision flow](#deployment-decision-flow)
- [Local setup](#local-setup)
- [Configuration](#configuration)
- [Security, compliance, and auditability](#security-compliance-and-auditability)
- [SLOs and operational goals](#slos-and-operational-goals)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Why IRONCLAD exists

Modern deploy pipelines answer:

- Does it compile?
- Do tests pass?
- Is style/lint clean?

They rarely answer:

- Is this change **semantically dangerous** for current production conditions?
- Is this blast radius acceptable right now?
- Can we reverse this safely within incident-response bounds?
- Is this a historically bad deployment window?

IRONCLAD fills that gap by combining code intent, dependency topology, historical incidents, and deployment timing intelligence.

---

## What IRONCLAD does

For each deployment candidate, IRONCLAD:

1. Reads the code diff and release metadata
2. Maps affected services to a live dependency graph
3. Correlates with incident/deploy history (last 90 days, configurable)
4. Computes a multi-axis risk score
5. Returns one of: **ALLOW**, **WARN**, or **BLOCK**
6. Explains the decision in plain English with actionable mitigation

It is **not** a linter, scanner, or test runner replacement. It is a **semantic production risk gate**.

---

## Core risk model

IRONCLAD computes deployment risk across three first-class axes:

### 1) Blast Radius

How many downstream services, data paths, and user journeys are exposed if this change fails.

### 2) Reversibility

Can this deployment be fully rolled back in under **60 seconds** (or organization-defined threshold)?

### 3) Timing Risk

Is the release occurring during a historically dangerous window (peak traffic, post-migration, low on-call depth, etc.)?

### Decision principle

If all three axes are red, IRONCLAD blocks deployment and provides:

- human-readable reasons,
- historical precedents,
- safer recommended deployment windows.

---

## System architecture

IRONCLAD is designed as a polyglot control-plane platform:

- **Go services**: low-latency deploy interception, dependency graph crawling, blast-radius computation
- **Python services**: intent classification, failure-grammar learning, risk scoring
- **TypeScript frontend**: risk dashboard, timeline, and failure-grammar explorer
- **PostgreSQL**: deploy history, incident correlations, risk snapshots, grammar registry

```text
CI/CD System ──> IRONCLAD Gate API (Go)
				  ├─ Diff Analyzer (Go/Python)
				  ├─ Dependency Graph Crawler (Go)
				  ├─ Risk Scoring Engine (Python)
				  ├─ Failure Grammar Learner (Python)
				  └─ Decision + Explanation
						   │
						   ├─ PostgreSQL (history + model metadata)
						   └─ Dashboard API (TypeScript clients)
```

---

## Tech stack

| Layer | Technology | Responsibility |
|---|---|---|
| Gate + Interceptors | Go | CI/CD webhook ingestion, request validation, policy enforcement |
| Topology Engine | Go | Live dependency graph retrieval and blast radius traversal |
| Semantic Engine | Python | Intent classification + deploy semantic interpretation |
| Learning Engine | Python | Failure grammar extraction and pattern evolution |
| Risk Service | Python | Multi-axis scoring and decision synthesis |
| Dashboard | TypeScript + React | Operator UX, timeline, explainability views |
| Data Store | PostgreSQL | Durable event history, risk snapshots, audit and lineage |

---

## Planned repository layout

> This structure is the target monorepo layout and will be scaffolded in upcoming commits.

```text
.
├─ apps/
│  └─ dashboard/                  # TypeScript/React risk UI
├─ services/
│  ├─ gate-go/                    # Deployment interceptor + decision API
│  ├─ topology-go/                # Dependency graph + blast radius engine
│  ├─ semantic-python/            # Intent classifier + semantic parser
│  └─ scoring-python/             # Risk scoring + failure grammar learner
├─ infra/
│  ├─ postgres/
│  │  └─ migrations/              # Schema and migrations
│  └─ docker/                     # Local dev orchestration
├─ docs/
│  ├─ architecture/
│  ├─ api/
│  ├─ runbooks/
│  └─ adrs/
├─ .github/
│  └─ workflows/                  # CI checks and release workflows
└─ README.md
```

---

## Data model (high-level)

Primary entities:

- `deployments`: release metadata, diff signatures, decision status
- `incidents`: severity, timeline, impacted services, root cause tags
- `risk_scores`: per-axis score snapshots and final decision outcome
- `service_dependencies`: graph edges for blast-radius traversal
- `failure_grammar_patterns`: learned risk motifs and confidence levels
- `decision_explanations`: immutable audit trail of why gate allowed/blocked

---

## Deployment decision flow

1. **Ingest** deploy request + diff metadata
2. **Classify intent** of change (functional, infra, migration, rollout, etc.)
3. **Resolve impact graph** from changed components
4. **Score risk axes** (blast, reversibility, timing)
5. **Consult historical grammar** and incident correlations
6. **Emit decision** with explanation and suggested safer window
7. **Log outcome** for audit + future model learning

---

## Local setup

The repo is currently in initial bootstrap stage; service scaffolding is planned next.

### Prerequisites

- `git`
- `go` (1.22+ recommended)
- `python` (3.11+ recommended)
- `node` (20+ recommended)
- `docker` (for local PostgreSQL and future service composition)

### Clone

```bash
git clone https://github.com/rohanpatel2002/IronClad.git
cd IronClad
```

### Current status

- Repository initialized
- License and baseline README present
- Monorepo services and CI are planned for next implementation milestones

---

## Configuration

Future services will use environment-based config with strong defaults.

Planned variables include:

- `IRONCLAD_ENV`
- `IRONCLAD_DATABASE_URL`
- `IRONCLAD_GATE_PORT`
- `IRONCLAD_CLAUDE_API_KEY`
- `IRONCLAD_POLICY_PROFILE`
- `IRONCLAD_ROLLBACK_SLO_SECONDS`

No secrets should be committed; all credentials will be managed via env vars and CI secrets.

---

## Security, compliance, and auditability

IRONCLAD is being built with enterprise controls in mind:

- Immutable decision logs for post-mortems and audits
- Explainable policy outputs (no black-box blocking)
- Principle-of-least-privilege service access
- Secret handling via runtime environment and vault-compatible patterns
- Backtestable policy changes before production rollout

---

## SLOs and operational goals

Target quality bar (initial goals):

- P95 gate decision latency: `< 2s`
- Decision availability: `99.9%`
- Explainability completeness: `100%` of blocked deployments include rationale + mitigation
- Rollback advisability accuracy: continuously improved through incident feedback loops

---

## Roadmap

### Phase 1 — Foundation

- Monorepo scaffolding (Go/Python/TypeScript/Postgres)
- Initial schema and migration pipeline
- Baseline gate API with stubbed scoring

### Phase 2 — Core scoring

- Live dependency graph crawler
- Blast radius traversal engine
- Timing risk model with historical windows

### Phase 3 — Learning system

- Incident correlation pipeline
- Failure grammar extraction + confidence scoring
- Continuous model updates from post-deploy outcomes

### Phase 4 — Operator UX

- Dashboard with timeline and explainability views
- Safer deployment window recommendations
- Grammar explorer and risk evolution analytics

---

## Contributing

Contributions are welcome. Formal contribution standards and development workflow will be published in `CONTRIBUTING.md`.

For now:

1. Fork the repo
2. Create a feature branch
3. Submit a focused PR with clear problem statement and tests

---

## License

This project is licensed under the **Apache License 2.0**. See `LICENSE` for details.
