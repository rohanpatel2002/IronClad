# Semantic Analysis Service Technical Specification

The `semantic-python` service analyzes pull request diffs and commit messages to classify deployment intent.

## Architecture
1. **Classifier Engine**: Uses Anthropic Claude 3 Haiku API (or heuristic fallback) to categorize code changes into `feature`, `hotfix`, `migration`, `rollout`, `refactor`, `config_update`, or `unknown`.
2. **Vector DB (ChromaDB)**: Stores dense embedding representations of code diffs and security incident logs to facilitate vector search and historical motif matching.
3. **Resilience**: Integrated with circuit breaking and exponential backoff retry in `gate-go`.
