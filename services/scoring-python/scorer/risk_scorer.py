"""
IRONCLAD Risk Scorer — 3-axis scoring engine

Computes blast_radius, reversibility, and timing_risk scores
for deployment decisions. Each axis produces a 0-1 score.
"""

import math
import time
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import List


@dataclass
class ScoringRequest:
    service: str
    commit_hash: str
    blast_radius: float          # pre-computed from topology service (0-1)
    changed_files: List[str] = field(default_factory=list)
    environment: str = "staging"
    service_criticality: float = 0.5  # 0-1


@dataclass
class ScoringResponse:
    blast_radius_score: float
    reversibility_score: float
    timing_risk_score: float
    confidence: float
    factors: List[str]
    computed_at: str


class BlastRadiusScorer:
    """
    Scales the pre-computed blast radius by service criticality.
    Acts as a pass-through with criticality weighting.
    """

    def score(self, req: ScoringRequest) -> tuple[float, List[str]]:
        factors = []
        score = req.blast_radius * req.service_criticality

        if score > 0.8:
            factors.append(
                f"Critical blast radius: {score:.0%} of system criticality at risk"
            )
        elif score > 0.5:
            factors.append(
                f"Moderate blast radius: {len(req.changed_files)} files changing across {score:.0%} system weight"
            )

        return min(score, 1.0), factors


class ReversibilityScorer:
    """
    Scores how difficult it would be to roll back this deployment.
    High score = hard to reverse = risky.
    """

    # File pattern → (score, label)
    PATTERNS = [
        (lambda f: "migration" in f.lower() or f.endswith(".sql"),  0.90, "DB migration — extremely hard to reverse"),
        (lambda f: f.endswith((".yaml", ".yml")) and "docker" not in f.lower(), 0.55, "Config change — moderate reversibility risk"),
        (lambda f: "dockerfile" in f.lower() or "docker-compose" in f.lower(), 0.65, "Docker config change — rebuild required"),
        (lambda f: f.endswith((".env", ".env.example")),             0.60, "Environment variable change"),
        (lambda f: "_test" in f.lower() or f.endswith((".test.ts", "_spec.py")), 0.10, "Test-only change — easily reversible"),
    ]

    def score(self, req: ScoringRequest) -> tuple[float, List[str]]:
        if not req.changed_files:
            return 0.35, ["No changed files specified — using conservative estimate"]

        factors = []
        max_score = 0.0
        file_scores = []

        for filepath in req.changed_files:
            file_score = 0.25  # baseline for unknown code files
            for pattern_fn, pattern_score, label in self.PATTERNS:
                if pattern_fn(filepath):
                    file_score = pattern_score
                    if label not in factors:
                        factors.append(label)
                    break
            file_scores.append(file_score)
            max_score = max(max_score, file_score)

        # Breadth penalty: many files increases risk even for low-risk changes
        if len(req.changed_files) > 10:
            breadth_penalty = min(0.2, len(req.changed_files) * 0.01)
            max_score = min(1.0, max_score + breadth_penalty)
            factors.append(f"{len(req.changed_files)} files changed — broad surface area increases risk")

        return min(max_score, 1.0), factors


class TimingRiskScorer:
    """
    Scores deployment risk based on time-of-day and day-of-week.
    High score = deploying at a risky time (Friday PM, nights, weekends).
    """

    def score(self, req: ScoringRequest) -> tuple[float, List[str]]:
        now = datetime.now(timezone.utc)
        hour = now.hour
        weekday = now.weekday()  # 0=Monday, 4=Friday, 5=Saturday, 6=Sunday
        factors = []
        score = 0.0

        # Day-of-week risk
        if weekday == 4 and hour >= 14:  # Friday afternoon
            score = 0.95
            factors.append("Friday afternoon deployment — highest risk window")
        elif weekday in (5, 6):  # Weekend
            score = 0.70
            factors.append(f"{'Saturday' if weekday == 5 else 'Sunday'} deployment — reduced on-call coverage")
        elif hour >= 22 or hour < 6:  # Night
            score = 0.75
            factors.append(f"Night deployment ({now.strftime('%H:%M UTC')}) — limited incident response capacity")
        elif hour >= 17:  # Evening
            score = 0.45
            factors.append("Evening deployment — team partially available")
        elif 10 <= hour < 16:  # Business hours
            score = 0.15
        else:  # Early morning
            score = 0.30

        # Production environments amplify timing risk
        if req.environment.lower() in ("production", "prod"):
            old_score = score
            score = min(1.0, score * 1.25)
            if score != old_score:
                factors.append("Production environment — timing risk amplified")

        return score, factors


class RiskScorer:
    """
    Combines all three scoring axes with configurable weights.
    Default weights: blast_radius=0.40, reversibility=0.35, timing=0.25
    """

    WEIGHTS = {
        "blast_radius": 0.40,
        "reversibility": 0.35,
        "timing": 0.25,
    }

    ENVIRONMENT_MULTIPLIERS = {
        "production": 1.30,
        "prod": 1.30,
        "staging": 1.00,
        "dev": 0.60,
        "development": 0.60,
    }

    def __init__(self):
        self.blast_scorer = BlastRadiusScorer()
        self.reversibility_scorer = ReversibilityScorer()
        self.timing_scorer = TimingRiskScorer()

    def score(self, req: ScoringRequest) -> ScoringResponse:
        start = time.time()

        blast_score, blast_factors = self.blast_scorer.score(req)
        rev_score, rev_factors = self.reversibility_scorer.score(req)
        timing_score, timing_factors = self.timing_scorer.score(req)

        env_multiplier = self.ENVIRONMENT_MULTIPLIERS.get(req.environment.lower(), 1.0)

        # Weighted combination
        combined = (
            blast_score * self.WEIGHTS["blast_radius"]
            + rev_score * self.WEIGHTS["reversibility"]
            + timing_score * self.WEIGHTS["timing"]
        ) * env_multiplier

        all_factors = blast_factors + rev_factors + timing_factors

        confidence = self._compute_confidence(blast_score, rev_score, timing_score)

        return ScoringResponse(
            blast_radius_score=round(blast_score, 4),
            reversibility_score=round(rev_score, 4),
            timing_risk_score=round(timing_score, 4),
            confidence=round(confidence, 4),
            factors=all_factors,
            computed_at=datetime.now(timezone.utc).isoformat(),
        )

    @staticmethod
    def _compute_confidence(blast: float, rev: float, timing: float) -> float:
        """
        Confidence is lower when scores cluster near decision boundaries (0.6, 0.8).
        Far from boundaries = high confidence in the decision.
        """
        avg = (blast + rev + timing) / 3.0
        dist_warn = abs(avg - 0.6)
        dist_block = abs(avg - 0.8)
        min_dist = min(dist_warn, dist_block)
        confidence = 0.50 + min_dist * 1.5
        return min(max(confidence, 0.0), 1.0)
