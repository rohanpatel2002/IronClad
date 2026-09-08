# Risk Scoring Engine Architecture & Weights Specification

The `scoring-python` microservice computes deployment risk using a multi-axis scoring model:

## 1. Scoring Axes
- **Blast Radius (40%)**: Pre-computed topology score scaled by target service criticality.
- **Reversibility Risk (35%)**: Evaluates change complexity (e.g., SQL migrations vs pure code changes vs test updates).
- **Timing Risk (25%)**: Assesses deployment timing window (Friday PM, weekend, night vs business hours).

## 2. Environment Amplification
- `production`: 1.3x multiplier
- `staging`: 1.0x multiplier
- `dev`: 0.6x multiplier

## 3. Failure Grammar Analysis
Analyzes patch diffs against learned failure signatures (e.g., raw SQL string concatenation, unhandled nil pointers). Matches amplify reversibility and blast radius scores.
