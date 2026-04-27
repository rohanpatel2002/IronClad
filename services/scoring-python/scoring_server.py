"""
IRONCLAD Risk Scoring Service
Flask HTTP server wrapping the 3-axis risk scorer.
"""

import os
import sys

# Make scorer module importable whether run from repo root or service dir
sys.path.insert(0, os.path.dirname(__file__))

from flask import Flask, jsonify, request, abort
from functools import wraps
from dotenv import load_dotenv
from scorer.risk_scorer import RiskScorer, ScoringRequest

load_dotenv()

app = Flask(__name__)
scorer = RiskScorer()

def require_api_key(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        api_key = os.getenv('INTERNAL_API_KEY')
        if not api_key:
            # If no key is set in env, allow all (dev mode)
            return f(*args, **kwargs)
        
        request_key = request.headers.get('X-Internal-API-Key')
        if request_key != api_key:
            return jsonify({"error": "unauthorized", "message": "invalid internal api key"}), 401
        return f(*args, **kwargs)
    return decorated_function


@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "scoring-python",
        "version": "0.1.0",
    }), 200


@app.route('/api/v1/score', methods=['POST'])
@require_api_key
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
@require_api_key
def get_failure_grammar():
    """Retrieve learned failure patterns."""
    signatures = scorer.grammar_learner.get_signatures()
    return jsonify({
        "patterns": signatures,
        "total": len(signatures),
    }), 200


@app.route('/api/v1/failure-grammar', methods=['POST'])
@require_api_key
def add_failure_grammar():
    """Add a new failure pattern to the grammar."""
    data = request.get_json(force=True)
    if not data:
        return jsonify({"error": "invalid_request", "message": "JSON body required"}), 400

    required = ("id", "name", "pattern", "type", "severity", "description")
    for field_name in required:
        if field_name not in data:
            return jsonify({"error": "missing_field", "field": field_name}), 400

    scorer.grammar_learner.add_signature(data)

    return jsonify({
        "status": "success",
        "message": "Signature added successfully",
        "signature": data
    }), 201


if __name__ == '__main__':
    port = int(os.getenv('SCORING_PORT', '8083'))
    debug = os.getenv('DEBUG', 'true').lower() == 'true'
    print(f"🧮 IRONCLAD Scoring service starting on port {port}")
    app.run(host='0.0.0.0', port=port, debug=debug)
