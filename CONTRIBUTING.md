# CONTRIBUTING to IRONCLAD

Thank you for your interest in contributing to IRONCLAD! This document outlines our development workflow, code standards, and collaboration practices.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Code Standards](#code-standards)
- [Contribution Workflow](#contribution-workflow)
- [Testing Requirements](#testing-requirements)
- [Commit Message Standards](#commit-message-standards)

## Getting Started

1. Fork the repository to your GitHub account
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/IronClad.git
   cd IronClad
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/rohanpatel2002/IronClad.git
   ```

## Development Environment

### Prerequisites

- Go 1.22+
- Python 3.11+
- Node.js 20+
- PostgreSQL 16+
- Docker (for local postgres)

### Local Setup

```bash
# Start postgres (requires docker)
cd infra/docker
docker-compose up -d postgres

# Setup Go service
cd services/gate-go
go mod download
go test ./...

# Setup Python services
cd services/semantic-python
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
pytest

# Setup Dashboard
cd apps/dashboard
npm install
npm run build
npm run dev
```

## Code Standards

### Go Services

- **Format**: Run `go fmt ./...` before committing
- **Lint**: Use `go vet ./...` to check for issues
- **Tests**: Minimum 80% coverage for new code
- **Naming**: Follow Go conventions (CamelCase, exported = Capitalized)
- **Error Handling**: Always handle errors explicitly; no silent failures

### Python Services

- **Format**: Use `black` for code formatting (line length: 100)
- **Type Hints**: All functions must have type annotations
- **Lint**: Pass `pylint` and `mypy` checks
- **Tests**: Use pytest; minimum 80% coverage
- **Docstrings**: Google-style docstrings for all public functions

### TypeScript / Dashboard

- **Format**: Use `prettier` for code formatting
- **Lint**: Pass `eslint` and `tsc` type checking
- **React**: Use functional components and hooks
- **Tests**: Jest with React Testing Library
- **Accessibility**: Follow WCAG 2.1 AA standards

## Contribution Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feat/your-feature-name
   ```

2. **Make your changes** and write tests

3. **Run local quality checks**:
   ```bash
   make fmt lint test  # In each service directory
   ```

4. **Commit with clear messages** (see below)

5. **Push to your fork**:
   ```bash
   git push origin feat/your-feature-name
   ```

6. **Open a Pull Request** against `main` with:
   - Clear description of changes
   - Link to any related issues
   - Screenshot/demo if UI changes
   - Checklist of testing performed

## Testing Requirements

All pull requests must include:

- **Unit tests** for new functions/methods
- **Integration tests** for cross-service interactions
- **No test regressions** (all existing tests still pass)
- **Coverage reports** (must not decrease overall coverage)

Run tests locally:

```bash
# Go
go test -v -cover ./...

# Python
pytest --cov=services/semantic-python

# TypeScript
npm test -- --coverage
```

## Commit Message Standards

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation updates
- `test`: Test additions or updates
- `refactor`: Code refactoring (no functional change)
- `perf`: Performance improvements
- `chore`: Build, tooling, or dependency updates

### Example

```
feat(gate): add blast radius caching

Implemented in-memory cache for blast radius calculations to reduce
latency on frequently-deployed services. Cache invalidates when
dependency graph changes.

Fixes #123
```

## Code Review Process

Pull requests require:

- ✅ CI checks passing (lint, test, build)
- ✅ At least one approval from maintainers
- ✅ All conversations resolved
- ✅ Up to date with `main`

## Questions?

Open an issue or start a discussion if you have questions about the development process or architecture.

---

**Happy contributing!**
