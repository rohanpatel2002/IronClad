from semantic_server import app

def test_semantic_server_health():
    client = app.test_client()
    res = client.get('/health')
    assert res.status_code == 200
    assert res.json['status'] == 'healthy'

def test_semantic_server_metrics():
    client = app.test_client()
    res = client.get('/metrics')
    assert res.status_code == 200
    assert b"semantic_requests_total" in res.data
