"""
IRONCLAD Risk Scoring Service
Flask HTTP server wrapping the 3-axis risk scorer.
"""

import os
import sys

# Make scorer module importable whether run from repo root or service dir
sys.path.insert(0, os.path.dirname(__file__))

from flask import Flask, jsonify, request
from dotenv import load_dotenv
from scorer.risk_scorer import RiskScorer, ScoringRequest

load_dotenv()

app = Flask(__name__)
scorer = RiskScorer()


@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "scoring-python",
        "version": "0.1.0",
    }), 200


@app.route('/api/v1/score', methods=['POST'])
def score_deployment():
    """
    Score a deployment across blast radius, reversibility, and timing risk.

    Expected JSON body:
    {
        "service": "payment-api",
        "commit_hash": "abc123",
        "blast_radius": 0.6,           // pre-computed 0-1
        "changed_files": ["main.go"],
        "environment": "production",
        "service_criticality": 0.9
    }
    """
    data = request.get_json(force=True)
    if not data:
        return jsonify({"error": "invalid_request", "message": "JSON body required"}), 400

    required = ("service", "commit_hash")
    for field_name in required:
        if field_name not in data:
            return jsonify({"error": "missing_field", "field": field_name}), 400

    req = ScoringRequest(
        service=data["service"],
        commit_hash=data["commit_hash"],
        blast_radius=float(data.get("blast_radius", 0.5)),
        changed_files=data.get("changed_files", []),
        environment=data.get("environment", "staging"),
        service_criticality=float(data.get("service_criticality", 0.5)),
    )

    result = scorer.score(req)

    return jsonify({
        "blast_radius_score": result.blast_radius_score,
        "reversibility_score": result.reversibility_score,
        "timing_risk_score": result.timing_risk_score,
        "confidence": result.confidence,
        "factors": result.factors,
        "computed_at": result.computed_at,
    }), 200


@app.route('/api/v1/failure-grammar', methods=['GET'])
def get_failure_grammar():
    """Retrieve learned failure patterns (stub — Phase 3 will implement learning)."""
    return jsonify({
        "patterns": [],
        "total": 0,
        "message": "Failure grammar learner not yet implemented — coming in Phase 3",
    }), 200


if __name__ == '__main__':
    port = int(os.getenv('SCORING_PORT', '8083'))
    debug = os.getenv('DEBUG', 'true').lower() == 'true'
    print(f"🧮 IRONCLAD Scoring service starting on port {port}")
    app.run(host='0.0.0.0', port=port, debug=debug)
