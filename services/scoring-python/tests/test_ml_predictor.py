from scorer.ml_predictor import RiskPredictor

def test_risk_predictor_fallback():
    predictor = RiskPredictor(model_path="non_existent.pkl")
    features = [10, 2.5, 0.9, 0.5]
    score, lower, upper = predictor.predict_risk_with_confidence(features)
    assert 0.0 <= score <= 1.0
    assert 0.0 <= lower <= score
    assert score <= upper <= 1.0
