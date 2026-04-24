import os
import json
from dataclasses import dataclass
from typing import List, Optional
try:
    from anthropic import Anthropic
except ImportError:
    Anthropic = None

@dataclass
class IntentClassificationRequest:
    service: str
    commit_hash: str
    branch: str
    changed_files: List[str]
    diff_summary: Optional[str] = None

@dataclass
class IntentClassificationResponse:
    intent: str  # e.g., 'feature', 'hotfix', 'migration', 'rollout'
    confidence: float
    reasoning: str

class SemanticClassifier:
    def __init__(self):
        self.api_key = os.environ.get("ANTHROPIC_API_KEY")
        self.client = None
        if self.api_key and Anthropic:
            self.client = Anthropic(api_key=self.api_key)

    def classify(self, req: IntentClassificationRequest) -> IntentClassificationResponse:
        # Fallback if no API key or anthropic package
        if not self.client:
            return self._mock_classification(req)
        
        try:
            return self._call_claude(req)
        except Exception as e:
            print(f"Claude API failed: {e}")
            return self._mock_classification(req)

    def _call_claude(self, req: IntentClassificationRequest) -> IntentClassificationResponse:
        system_prompt = """You are an AI code intent classifier for a continuous deployment pipeline. 
Given the service name, branch, changed files, and optional diff summary, classify the deployment intent.
You MUST respond with a JSON object with three keys:
- 'intent': one of ['feature', 'hotfix', 'migration', 'rollout', 'refactor', 'config_update', 'unknown']
- 'confidence': float between 0.0 and 1.0
- 'reasoning': a brief explanation of why you classified it this way.
Output ONLY valid JSON."""

        prompt = f"""Service: {req.service}
Branch: {req.branch}
Commit: {req.commit_hash}
Changed Files: {', '.join(req.changed_files)}
Diff Summary: {req.diff_summary or 'None'}

Classify the intent."""

        response = self.client.messages.create(
            model="claude-3-haiku-20240307",
            max_tokens=300,
            temperature=0.0,
            system=system_prompt,
            messages=[
                {"role": "user", "content": prompt}
            ]
        )

        try:
            text = response.content[0].text
            # Simple heuristic to extract JSON if Claude adds markdown formatting
            if "```json" in text:
                text = text.split("```json")[1].split("```")[0].strip()
            elif "```" in text:
                text = text.split("```")[1].strip()
                
            data = json.loads(text)
            return IntentClassificationResponse(
                intent=data.get("intent", "unknown"),
                confidence=float(data.get("confidence", 0.5)),
                reasoning=data.get("reasoning", "Failed to parse reasoning")
            )
        except Exception as e:
            print(f"Failed to parse Claude response: {e}")
            return self._mock_classification(req)

    def _mock_classification(self, req: IntentClassificationRequest) -> IntentClassificationResponse:
        """Heuristic-based fallback if Claude isn't available."""
        intent = "feature"
        confidence = 0.5
        reasoning = "Fallback heuristic based on branch name and files."

        if req.branch:
            branch_lower = req.branch.lower()
            if "hotfix" in branch_lower or "fix" in branch_lower:
                intent = "hotfix"
                confidence = 0.8
            elif "refactor" in branch_lower:
                intent = "refactor"
                confidence = 0.7

        for f in req.changed_files:
            f_lower = f.lower()
            if "migration" in f_lower or f_lower.endswith(".sql"):
                intent = "migration"
                confidence = 0.9
                reasoning = "SQL migration files detected."
                break
            elif "config" in f_lower or f_lower.endswith((".yml", ".yaml")):
                if intent != "migration":
                    intent = "config_update"
                    confidence = 0.7
                    reasoning = "Configuration files detected."

        return IntentClassificationResponse(
            intent=intent,
            confidence=confidence,
            reasoning=reasoning
        )
