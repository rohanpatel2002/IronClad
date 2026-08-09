import pytest
from scorer.risk_scorer import RiskScorer, ScoringRequest

def test_risk_scorer_basic():
    scorer = RiskScorer()
    req = ScoringRequest(
        service="payment-api",
        commit_hash="abc1234",
        blast_radius=0.8,
        changed_files=["migrations/001.sql"],
        environment="production",
        service_criticality=0.9
    )
    res = scorer.score(req)
    assert res.blast_radius_score > 0.5
    assert res.reversibility_score >= 0.9
    assert res.confidence > 0.0

def test_risk_scorer_weight_tuning():
    scorer = RiskScorer()
    scorer.set_weights(blast=0.5, reversibility=0.3, timing=0.2)
    assert pytest.approx(scorer.WEIGHTS["blast_radius"], 0.01) == 0.5
    scorer.reset_weights()
    assert pytest.approx(scorer.WEIGHTS["blast_radius"], 0.01) == 0.4
