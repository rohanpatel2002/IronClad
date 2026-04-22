-- IRONCLAD Core Schema
-- This schema defines the data model for deployment tracking, incident correlation, and risk scoring.

-- Deployments table: tracks all deployment attempts and decisions
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_service (service_name),
    INDEX idx_timestamp (deploy_timestamp),
    INDEX idx_decision (decision_status)
);

-- Incidents table: correlates production incidents with deployments
CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_timestamp TIMESTAMP NOT NULL,
    severity VARCHAR(50) NOT NULL, -- SEV1, SEV2, SEV3, etc.
    title VARCHAR(512) NOT NULL,
    description TEXT,
    impacted_services TEXT[], -- Array of service names
    root_cause TEXT,
    related_deployment_id UUID REFERENCES deployments(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,
    INDEX idx_severity (severity),
    INDEX idx_timestamp (incident_timestamp)
);

-- Risk scores: immutable snapshots of decision outcomes
CREATE TABLE risk_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id),
    blast_radius NUMERIC(5,2) NOT NULL, -- 0-100 score
    reversibility NUMERIC(5,2) NOT NULL,
    timing_risk NUMERIC(5,2) NOT NULL,
    final_decision VARCHAR(50) NOT NULL,
    computed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_deployment (deployment_id),
    FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE
);

-- Service dependencies: graph edges for blast radius calculation
CREATE TABLE service_dependencies (
    source_service VARCHAR(255) NOT NULL,
    target_service VARCHAR(255) NOT NULL,
    dependency_type VARCHAR(50) NOT NULL, -- http, db, queue, cache, etc.
    criticality NUMERIC(3,2) NOT NULL, -- 0-1 scale
    PRIMARY KEY (source_service, target_service, dependency_type),
    INDEX idx_source (source_service),
    INDEX idx_target (target_service)
);

-- Failure grammar patterns: learned risk motifs
CREATE TABLE failure_grammar_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern_name VARCHAR(255) NOT NULL,
    description TEXT,
    code_signature VARCHAR(512), -- pattern to match in diffs
    confidence_score NUMERIC(5,4) NOT NULL, -- 0-1
    occurrence_count INT DEFAULT 0,
    average_blast_radius NUMERIC(5,2),
    average_reversibility NUMERIC(5,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_confidence (confidence_score DESC),
    UNIQUE(pattern_name)
);

-- Decision explanation logs: audit trail
CREATE TABLE decision_explanations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id),
    decision VARCHAR(50) NOT NULL,
    reasoning TEXT NOT NULL,
    risk_factors TEXT[], -- Array of contributing factors
    suggested_safe_window TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE
);
