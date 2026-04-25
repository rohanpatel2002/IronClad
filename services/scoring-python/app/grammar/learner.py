import json
import os
from typing import List, Dict, Any, Set
import re

class FailureGrammarLearner:
    """
    Failure Grammar Learner maintains a set of known failure signatures.
    A signature represents a coding pattern (e.g., regex or AST structure)
    that has historically caused incidents.
    """
    def __init__(self, grammar_file: str = "failure_grammar.json"):
        self.grammar_file = grammar_file
        self.signatures: List[Dict[str, Any]] = []
        self._load_grammar()

    def _load_grammar(self):
        """Loads the failure grammar from disk if it exists."""
        if os.path.exists(self.grammar_file):
            try:
                with open(self.grammar_file, "r") as f:
                    self.signatures = json.load(f)
            except Exception as e:
                print(f"Error loading grammar file: {e}")
                self.signatures = self._default_grammar()
        else:
            self.signatures = self._default_grammar()
            self._save_grammar()

    def _save_grammar(self):
        """Saves the current failure grammar to disk."""
        try:
            with open(self.grammar_file, "w") as f:
                json.dump(self.signatures, f, indent=4)
        except Exception as e:
            print(f"Error saving grammar file: {e}")

    def _default_grammar(self) -> List[Dict[str, Any]]:
        """Returns a baseline set of failure signatures."""
        return [
            {
                "id": "FG-001",
                "name": "Unbounded Query",
                "pattern": r"SELECT\s+\*\s+FROM\s+\w+(?!\s+LIMIT|\s+WHERE)",
                "type": "regex",
                "severity": "HIGH",
                "description": "SQL query without LIMIT or WHERE clause"
            },
            {
                "id": "FG-002",
                "name": "Missing Timeout",
                "pattern": r"requests\.(get|post|put|delete)\([^,]+(?!,\s*timeout=).*\)",
                "type": "regex",
                "severity": "MEDIUM",
                "description": "HTTP request without a timeout specified"
            },
            {
                "id": "FG-003",
                "name": "Hardcoded Secret",
                "pattern": r"(?i)(password|secret|key|token)\s*=\s*['\"][a-zA-Z0-9]{10,}['\"]",
                "type": "regex",
                "severity": "CRITICAL",
                "description": "Potential hardcoded secret or credential"
            }
        ]

    def add_signature(self, signature: Dict[str, Any]):
        """Adds a new failure signature to the grammar."""
        self.signatures.append(signature)
        self._save_grammar()

    def get_signatures(self) -> List[Dict[str, Any]]:
        """Returns all loaded failure signatures."""
        return self.signatures
