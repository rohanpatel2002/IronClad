import json
from events.consumer import EventProcessor

def test_event_processor():
    processor = EventProcessor()
    payload = json.dumps({
        "service": "order-service",
        "lines_changed": 150,
        "complexity": 8,
        "author_trust": 0.85,
        "blast_radius": 0.6
    })
    res = processor.process_event_payload(payload)
    assert res["service"] == "order-service"
    assert res["status"] == "processed"
    assert 0.0 <= res["risk_score"] <= 1.0
    assert processor.processed_count == 1
