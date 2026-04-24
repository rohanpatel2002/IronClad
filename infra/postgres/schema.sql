-- IRONCLAD Core Schema
-- This schema defines the data model for deployment tracking, incident correlation, and risk scoring.

-- Deployments table: tracks all deployment attempts and decisions
CREATE TABLE deployments (
    id VARCHAR(255) PRIMARY KEY,
    deploy_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    commit_hash VARCHAR(255) NOT NULL,
    branch VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    author_email VARCHAR(255),
    diff_summary TEXT,
    decision_status VARCHAR(50) NOT NULL, -- ALLOW, WARN, BLOCK
    decision_timestamp TIMESTAMP,
    explanation TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deploy_service ON deployments(service_name);
CREATE INDEX idx_deploy_timestamp ON deployments(deploy_timestamp);
CREATE INDEX idx_deploy_decision ON deployments(decision_status);

-- Incidents table: correlates production incidents with deployments
CREATE TABLE incidents (
    id VARCHAR(255) PRIMARY KEY,
    incident_timestamp TIMESTAMP NOT NULL,
    severity VARCHAR(50) NOT NULL, -- SEV1, SEV2, SEV3, etc.
    title VARCHAR(512) NOT NULL,
    description TEXT,
    impacted_services TEXT[], -- Array of service names
    root_cause TEXT,
    related_deployment_id VARCHAR(255) REFERENCES deployments(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP
);

CREATE INDEX idx_incidents_severity ON incidents(severity);
CREATE INDEX idx_incidents_timestamp ON incidents(incident_timestamp);

-- Risk scores: immutable snapshots of decision outcomes
CREATE TABLE risk_scores (
    id VARCHAR(255) PRIMARY KEY,
    deployment_id VARCHAR(255) NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    blast_radius NUMERIC(5,2) NOT NULL, -- 0-100 score
    reversibility NUMERIC(5,2) NOT NULL,
    timing_risk NUMERIC(5,2) NOT NULL,
    final_decision VARCHAR(50) NOT NULL,
    computed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_scores_deployment ON risk_scores(deployment_id);

-- Service dependencies: graph edges for blast radius calculation
CREATE TABLE service_dependencies (
    source_service VARCHAR(255) NOT NULL,
    target_service VARCHAR(255) NOT NULL,
    dependency_type VARCHAR(50) NOT NULL, -- http, db, queue, cache, etc.
    criticality NUMERIC(3,2) NOT NULL, -- 0-1 scale
    PRIMARY KEY (source_service, target_service, dependency_type)
);

CREATE INDEX idx_service_deps_source ON service_dependencies(source_service);
CREATE INDEX idx_service_deps_target ON service_dependencies(target_service);

-- Failure grammar patterns: learned risk motifs
CREATE TABLE failure_grammar_patterns (
    id VARCHAR(255) PRIMARY KEY,
    pattern_name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    code_signature VARCHAR(512), -- pattern to match in diffs
    confidence_score NUMERIC(5,4) NOT NULL, -- 0-1
    occurrence_count INT DEFAULT 0,
    average_blast_radius NUMERIC(5,2),
    average_reversibility NUMERIC(5,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_failure_grammar_confidence ON failure_grammar_patterns(confidence_score DESC);

-- Decision explanation logs: audit trail
CREATE TABLE decision_explanations (
    id VARCHAR(255) PRIMARY KEY,
    deployment_id VARCHAR(255) NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    decision VARCHAR(50) NOT NULL,
    reasoning TEXT NOT NULL,
    risk_factors TEXT[], -- Array of contributing factors
    suggested_safe_window TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
