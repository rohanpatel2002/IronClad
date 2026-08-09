import pytest
from scorer.risk_scorer import RiskScorer, ScoringRequest

def test_benchmark_risk_scorer_performance(benchmark=None):
    scorer = RiskScorer()
    req = ScoringRequest(
        service="payment-api",
        commit_hash="abc1234",
        blast_radius=0.75,
        changed_files=["migrations/001.sql", "main.go", "config.yaml"],
        environment="production",
        service_criticality=0.95
    )
    if benchmark:
        benchmark(scorer.score, req)
    else:
        for _ in range(100):
            res = scorer.score(req)
            assert res.blast_radius_score > 0
