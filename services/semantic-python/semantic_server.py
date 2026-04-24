"""
IRONCLAD Semantic Intent Classifier
Analyzes code diffs and classifies deployment intent.
"""

import os
from flask import Flask, jsonify, request
from dotenv import load_dotenv
from classifier import SemanticClassifier, IntentClassificationRequest

load_dotenv()

app = Flask(__name__)
classifier = SemanticClassifier()

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "healthy", "service": "semantic-python"}), 200

@app.route('/api/v1/classify', methods=['POST'])
def classify_intent():
    """
    Classify the intent of a deployment diff.
    """
    data = request.get_json()
    if not data:
        return jsonify({"error": "invalid_request", "message": "JSON body required"}), 400

    required_fields = ["service", "commit_hash", "branch", "changed_files"]
    for field in required_fields:
        if field not in data:
            return jsonify({"error": "missing_field", "field": field}), 400

    req = IntentClassificationRequest(
        service=data["service"],
        commit_hash=data["commit_hash"],
        branch=data["branch"],
        changed_files=data["changed_files"],
        diff_summary=data.get("diff_summary")
    )

    try:
        response = classifier.classify(req)
        return jsonify({
            "intent": response.intent,
            "confidence": response.confidence,
            "reasoning": response.reasoning
        }), 200
    except Exception as e:
        return jsonify({
            "intent": "unknown",
            "confidence": 0.0,
            "reasoning": f"Classification failed: {e}"
        }), 500

if __name__ == '__main__':
    port = os.getenv('SEMANTIC_PORT', '8082')
    print(f"🧠 IRONCLAD Semantic service starting on port {port}")
    app.run(host='0.0.0.0', port=int(port), debug=True)
