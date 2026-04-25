import re
from typing import List, Dict, Any, Optional
from .learner import FailureGrammarLearner

class SignatureMatch:
    def __init__(self, signature_id: str, name: str, severity: str, description: str, matched_text: str):
        self.signature_id = signature_id
        self.name = name
        self.severity = severity
        self.description = description
        self.matched_text = matched_text

    def to_dict(self) -> Dict[str, Any]:
        return {
            "signature_id": self.signature_id,
            "name": self.name,
            "severity": self.severity,
            "description": self.description,
            "matched_text": self.matched_text
        }

class GrammarMatcher:
    """
    Evaluates code patches against the known failure grammar.
    """
    def __init__(self, learner: FailureGrammarLearner):
        self.learner = learner
        self._compiled_regexes = {}
        self._compile_patterns()

    def _compile_patterns(self):
        """Pre-compiles regex patterns for performance."""
        for sig in self.learner.get_signatures():
            if sig.get("type") == "regex" and "pattern" in sig:
                try:
                    self._compiled_regexes[sig["id"]] = re.compile(sig["pattern"])
                except Exception as e:
                    print(f"Failed to compile pattern for {sig['id']}: {e}")

    def analyze_patch(self, patch_content: str) -> List[SignatureMatch]:
        """
        Analyzes a git patch/diff string to find known failure signatures.
        Currently focuses on added lines (+).
        """
        if not patch_content:
            return []

        # Extract additions (lines starting with + but not +++)
        added_lines = []
        for line in patch_content.split('\n'):
            if line.startswith('+') and not line.startswith('+++'):
                added_lines.append(line[1:]) # Remove the '+' prefix
        
        added_text = '\n'.join(added_lines)
        if not added_text.strip():
            return []

        matches = []
        for sig in self.learner.get_signatures():
            sig_id = sig["id"]
            if sig.get("type") == "regex" and sig_id in self._compiled_regexes:
                pattern = self._compiled_regexes[sig_id]
                found = pattern.finditer(added_text)
                for match in found:
                    matches.append(SignatureMatch(
                        signature_id=sig_id,
                        name=sig["name"],
                        severity=sig["severity"],
                        description=sig["description"],
                        matched_text=match.group(0)
                    ))
            
            # Placeholder for AST matching integration
            elif sig.get("type") == "ast":
                pass # AST matching would go here (e.g., using python's ast module)

        return matches
