import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))

import pytest
from app.grammar.learner import FailureGrammarLearner
from app.grammar.matcher import GrammarMatcher


@pytest.fixture
def matcher(tmp_path):
    """Create a fresh matcher backed by a temp grammar file."""
    grammar_file = str(tmp_path / "grammar.json")
    learner = FailureGrammarLearner(grammar_file=grammar_file)
    return GrammarMatcher(learner)


def test_matcher_finds_unbounded_sql(matcher):
    patch = "+SELECT * FROM users\n"
    matches = matcher.analyze_patch(patch)
    ids = [m.signature_id for m in matches]
    assert "FG-001" in ids, f"Expected FG-001, got {ids}"


def test_matcher_finds_hardcoded_secret(matcher):
    patch = "+password = 'supersecretvalue123'\n"
    matches = matcher.analyze_patch(patch)
    ids = [m.signature_id for m in matches]
    assert "FG-003" in ids, f"Expected FG-003 (hardcoded secret), got {ids}"


def test_matcher_ignores_deletions(matcher):
    # Lines starting with '-' are removed code and should not trigger
    patch = "-SELECT * FROM users\n"
    matches = matcher.analyze_patch(patch)
    assert len(matches) == 0, "Deleted lines should not be scanned"


def test_matcher_returns_empty_on_empty_patch(matcher):
    matches = matcher.analyze_patch("")
    assert matches == []


def test_matcher_returns_empty_on_safe_code(matcher):
    patch = "+def calculate_score(blast, reversibility, timing):\n+    return (blast + reversibility + timing) / 3\n"
    matches = matcher.analyze_patch(patch)
    assert matches == [], f"Expected no matches for safe code, got {matches}"


def test_match_has_correct_severity(matcher):
    patch = "+secret = 'hardcodedvalue999'\n"
    matches = matcher.analyze_patch(patch)
    critical = [m for m in matches if m.severity == "CRITICAL"]
    assert len(critical) >= 1, "Expected at least one CRITICAL match"


def test_matcher_respects_custom_signature(matcher):
    # Add a custom signature to the learner
    matcher.learner.add_signature({
        "id": "FG-TEST",
        "name": "Test Pattern",
        "pattern": r"test_danger_keyword",
        "type": "regex",
        "severity": "HIGH",
        "description": "Test pattern"
    })
    # Recompile
    from app.grammar.matcher import GrammarMatcher
    new_matcher = GrammarMatcher(matcher.learner)

    patch = "+some_code with test_danger_keyword here\n"
    matches = new_matcher.analyze_patch(patch)
    ids = [m.signature_id for m in matches]
    assert "FG-TEST" in ids
