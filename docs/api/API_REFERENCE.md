# IRONCLAD API Documentation

## Overview

IRONCLAD provides three primary API interfaces:

1. **Gate API** — Deployment decision requests
2. **Analysis API** — Risk scoring and pattern querying
3. **Admin API** — Configuration and model management

All endpoints require authentication via API key (header: `X-IRONCLAD-KEY`) unless noted otherwise.

## Gate API

### POST /api/v1/decision

Request a deployment decision before promotion.

**Request**:
```http
POST /api/v1/decision
X-IRONCLAD-KEY: your-api-key
Content-Type: application/json

{
  "commit_hash": "abc123def456789",
  "service": "user-auth-service",
  "branch": "main",
  "author_email": "engineer@company.com",
  "diff_size_bytes": 2048,
  "changed_files": ["src/auth/jwt.go", "src/auth/oauth.go"],
  "environment": "production",
  "force_check": false
}
```

**Response** (200 OK):
```json
{
  "decision_id": "dec-uuid-v4",
  "decision": "ALLOW",
  "risk_scores": {
    "blast_radius": 0.45,
    "reversibility": 0.92,
    "timing_risk": 0.20
  },
  "confidence": 0.87,
  "explanation": {
    "summary": "Safe to deploy: low blast radius, high reversibility, off-peak window",
    "risk_factors": [],
    "mitigations": []
  },
  "suggested_safe_windows": [
    {
      "start": "2026-04-23T10:00:00Z",
      "end": "2026-04-23T12:00:00Z",
      "confidence": 0.95
    }
  ],
  "decision_timestamp": "2026-04-22T18:30:00Z"
}
```

**Response** (429 Too Many Requests):
```json
{
  "error": "rate_limit_exceeded",
  "retry_after_seconds": 60
}
```

### GET /api/v1/decision/:decision_id

Retrieve details of a previous decision.

**Response** (200 OK):
```json
{
  "decision_id": "dec-uuid-v4",
  "deployment_id": "deploy-uuid-v4",
  "decision": "BLOCK",
  "risk_scores": { ... },
  "created_at": "2026-04-22T18:30:00Z",
  "explanation": { ... }
}
```

## Analysis API

### GET /api/v1/deployments

List recent deployments with filtering.

**Query Parameters**:
- `service`: (optional) Filter by service name
- `decision`: (optional) Filter by decision (ALLOW, WARN, BLOCK)
- `limit`: (optional) Max results (default: 20)
- `offset`: (optional) Pagination offset (default: 0)

**Response** (200 OK):
```json
{
  "total": 150,
  "deployments": [
    {
      "id": "deploy-uuid-v4",
      "service": "user-service",
      "commit_hash": "abc123",
      "decision": "ALLOW",
      "timestamp": "2026-04-22T18:30:00Z"
    }
  ]
}
```

### GET /api/v1/failure-grammar

Retrieve learned failure patterns.

**Query Parameters**:
- `min_confidence`: (optional) Minimum confidence score (0-1)
- `service`: (optional) Filter by affected service

**Response** (200 OK):
```json
{
  "patterns": [
    {
      "id": "pattern-uuid",
      "name": "Database connection pool exhaustion on cache invalidation",
      "code_signature": "SELECT.*JOIN.*ON.*WHERE.*NOT IN",
      "confidence": 0.89,
      "occurrence_count": 12,
      "average_blast_radius": 0.78,
      "related_incidents": 6
    }
  ]
}
```

### GET /api/v1/incidents

List correlated incidents.

**Query Parameters**:
- `severity`: (optional) SEV1, SEV2, SEV3
- `related_service`: (optional) Service name
- `days`: (optional) Look back window (default: 90)

**Response** (200 OK):
```json
{
  "incidents": [
    {
      "id": "incident-uuid",
      "timestamp": "2026-04-20T02:30:00Z",
      "severity": "SEV1",
      "title": "User auth service timeout cascade",
      "impacted_services": ["user-service", "api-gateway", "frontend"],
      "root_cause": "Database connection leak in JWT validation",
      "related_deployments": ["deploy-uuid-1", "deploy-uuid-2"]
    }
  ]
}
```

## Admin API

### POST /api/v1/admin/policy

Update deployment policy thresholds.

**Request** (requires `X-IRONCLAD-ADMIN-KEY` header):
```json
{
  "blast_radius_threshold": 0.8,
  "reversibility_minimum": 0.6,
  "timing_risk_window_start_hour": 16,
  "timing_risk_window_end_hour": 8,
  "decision_explanation_required": true
}
```

**Response** (200 OK):
```json
{
  "status": "updated",
  "effective_at": "2026-04-23T00:00:00Z"
}
```

### POST /api/v1/admin/incidents/correlate

Manually tag an incident as related to a deployment.

**Request** (requires `X-IRONCLAD-ADMIN-KEY`):
```json
{
  "incident_id": "incident-uuid",
  "deployment_id": "deploy-uuid",
  "root_cause_category": "database_leak"
}
```

**Response** (200 OK):
```json
{
  "status": "correlated",
  "grammar_updated": true
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid_request",
  "message": "commit_hash is required",
  "details": { ... }
}
```

### 401 Unauthorized
```json
{
  "error": "authentication_failed",
  "message": "Invalid or missing API key"
}
```

### 404 Not Found
```json
{
  "error": "not_found",
  "message": "Decision with ID 'dec-abc123' not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal_error",
  "message": "An unexpected error occurred",
  "request_id": "req-uuid-for-support"
}
```

## Rate Limiting

All endpoints (except /health) enforce rate limits:

- **Default**: 1000 requests/hour per API key
- **Headers returned**:
  - `X-RateLimit-Limit`: Total limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Unix timestamp of reset

## Authentication

API keys are managed via your IRONCLAD instance dashboard.

Provide in all requests:
```
X-IRONCLAD-KEY: your-api-key-here
```

Admin operations require:
```
X-IRONCLAD-ADMIN-KEY: your-admin-key-here
```

---

For implementation details and SDKs, see the service directories.
