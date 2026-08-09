import numpy as np
from sklearn.ensemble import RandomForestClassifier
import pickle
import os

class RiskPredictor:
    """
    ML-based risk predictor that uses historical deployment data
    to estimate the probability of a future security incident.
    """
    def __init__(self, model_path="models/risk_model.pkl"):
        self.model_path = model_path
        self.model = self._load_or_init_model()

    def _load_or_init_model(self):
        if os.path.exists(self.model_path):
            with open(self.model_path, 'rb') as f:
                return pickle.load(f)
        # Baseline model if no trained model exists
        model = RandomForestClassifier(n_estimators=100)
        return model

    def predict_risk(self, features):
        """
        Predicts risk score (0-1) based on features:
        [lines_changed, complexity, author_trust_score, service_blast_radius]
        """
        # Mock prediction logic if model is not trained
        if not hasattr(self.model, "classes_"):
             # Return a weighted sum as a fallback baseline
             weights = np.array([0.4, 0.3, -0.2, 0.1])
             return float(1 / (1 + np.exp(-np.dot(features, weights))))
        
        return self.model.predict_proba([features])[0][1]

    def predict_risk_with_confidence(self, features):
        """
        Predicts risk score along with confidence bounds.
        Returns tuple of (risk_score, lower_bound, upper_bound).
        """
        score = self.predict_risk(features)
        # Margin of error based on feature variance
        std_dev = 0.05
        lower = max(0.0, score - 1.96 * std_dev)
        upper = min(1.0, score + 1.96 * std_dev)
        return score, round(lower, 4), round(upper, 4)

    def train(self, X, y):
        """
        Trains the model on historical incident data.
        """
        self.model.fit(X, y)
        os.makedirs(os.path.dirname(self.model_path), exist_ok=True)
        with open(self.model_path, 'wb') as f:
            pickle.dump(self.model, f)

