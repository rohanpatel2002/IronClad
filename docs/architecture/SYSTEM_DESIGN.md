# IRONCLAD Architecture

## High-Level Overview

IRONCLAD is a semantic deployment risk assessment platform designed to understand the intent and impact of code changes before they reach production.

## System Components

### 1. Deployment Gate (Go)
- **Service**: `services/gate-go`
- **Port**: 8080
- **Responsibility**: 
  - Intercepts CI/CD promotion requests
  - Coordinates scoring and decision flow
  - Exposes decision API to CI systems

### 2. Topology Engine (Go)
- **Service**: `services/topology-go`
- **Port**: 8081
- **Responsibility**:
  - Crawls live service dependency graph
  - Calculates blast radius for code changes
  - Maintains service topology cache

### 3. Semantic Intent Classifier (Python)
- **Service**: `services/semantic-python`
- **Port**: 8082
- **Responsibility**:
  - Analyzes code diffs semantically
  - Classifies deployment intent (feature, hotfix, migration, etc.)
  - Uses Claude API for intent understanding

### 4. Risk Scoring Engine (Python)
- **Service**: `services/scoring-python`
- **Port**: 8083
- **Responsibility**:
  - Computes multi-axis risk scores
  - Learns failure patterns from incidents
  - Generates explainable decision justifications

### 5. Dashboard (TypeScript/React)
- **App**: `apps/dashboard`
- **Port**: 3000 (dev) / 3001 (prod)
- **Responsibility**:
  - Visualizes deployment risk assessments
  - Shows failure grammar patterns
  - Provides timeline and historical analysis

## Data Model

See `infra/postgres/schema.sql` for the full schema.

### Core Entities

- **Deployments**: All release attempts and decisions
- **Incidents**: Production incidents correlated with deployments
- **Risk Scores**: Immutable decision snapshots
- **Service Dependencies**: Graph edges for impact analysis
- **Failure Grammar**: Learned patterns that precede incidents
- **Decision Explanations**: Audit trail of gate decisions

## Request Flow

```
CI/CD Event
    ↓
[Gate API] → Validate request
    ↓
[Topology] → Calculate blast radius
    ↓
[Semantic] → Classify intent
    ↓
[Scoring] → Compute risk axes
    ↓
[Database] → Store decision + history
    ↓
[Gate API] → Return ALLOW | WARN | BLOCK + explanation
    ↓
CI/CD → Promote or hold
```

## Deployment Decisions

### ALLOW
- All risk axes are within acceptable bounds
- No historical pattern match
- Safe deployment window

### WARN
- One or more risk axes elevated
- Deployment is allowed but logged for monitoring
- Recommendation for safer window provided

### BLOCK
- Multiple risk axes in red zone
- Strong historical pattern match with incident correlation
- Actionable mitigation steps provided

## Learning Loop

1. Deploy occurs (ALLOW, WARN, or BLOCK)
2. Deployment executes in production
3. Monitor for incidents in post-deploy window (e.g., 1 hour)
4. If incident occurs:
   - Correlate with deployment
   - Extract failure patterns
   - Update failure grammar
   - Adjust risk weights

## API Contracts

### Gate API: POST /api/v1/decision

**Request**:
```json
{
  "commit_hash": "abc123def456",
  "diff": "...",
  "service": "user-service",
  "branch": "main",
  "author": "engineer@company.com"
}
```

**Response**:
```json
{
  "decision": "ALLOW|WARN|BLOCK",
  "risk_scores": {
    "blast_radius": 0.65,
    "reversibility": 0.85,
    "timing_risk": 0.45
  },
  "explanation": "..."
}
```

## Running Locally

```bash
# Start postgres
cd infra/docker && docker-compose up -d

# In separate terminals:
cd services/gate-go && go run main.go
cd services/semantic-python && python semantic_server.py
cd services/scoring-python && python scoring_server.py
cd apps/dashboard && npm run dev
```

Access:
- Dashboard: http://localhost:3000
- Gate API: http://localhost:8080
- Postgres: localhost:5432
