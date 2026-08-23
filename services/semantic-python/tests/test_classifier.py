from classifier import SemanticClassifier, IntentClassificationRequest

def test_semantic_classifier_heuristic():
    classifier = SemanticClassifier()
    req = IntentClassificationRequest(
        service="payment-api",
        commit_hash="abc1234",
        branch="feature/payment-v2",
        changed_files=["migrations/001_create_table.sql"]
    )
    res = classifier.classify(req)
    assert res.intent == "migration"
    assert res.confidence >= 0.8

def test_semantic_classifier_batch():
    classifier = SemanticClassifier()
    reqs = [
        IntentClassificationRequest("svc1", "hash1", "main", ["test.py"]),
        IntentClassificationRequest("svc2", "hash2", "fix/bug", ["app.go"])
    ]
    results = classifier.batch_classify(reqs)
    assert len(results) == 2
