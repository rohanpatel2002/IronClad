import pytest

@pytest.fixture
def sample_commit_diffs():
    return [
        {
            "service": "payment-api",
            "commit_hash": "commit-001",
            "branch": "feature/stripe-v2",
            "changed_files": ["services/payment.go", "go.mod"],
            "expected_intent": "feature"
        },
        {
            "service": "user-service",
            "commit_hash": "commit-002",
            "branch": "fix/login-leak",
            "changed_files": ["pkg/auth/login.go"],
            "expected_intent": "hotfix"
        }
    ]
