from scoring_server import app

def test_scoring_server_health():
    client = app.test_client()
    res = client.get('/health')
    assert res.status_code == 200
    assert res.json['status'] == 'healthy'

def test_scoring_server_metrics():
    client = app.test_client()
    res = client.get('/metrics')
    assert res.status_code == 200
    assert b"scoring_requests_total" in res.data

def test_scoring_server_post_score():
    client = app.test_client()
    payload = {
        "service": "payment-api",
        "commit_hash": "def456",
        "blast_radius": 0.5,
        "changed_files": ["services/payment.go"],
        "environment": "staging"
    }
    res = client.post('/api/v1/score', json=payload)
    assert res.status_code == 200
    assert "blast_radius_score" in res.json
