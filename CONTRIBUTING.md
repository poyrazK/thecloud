# Contributing to TheCloud

Thank you for your interest in contributing to TheCloud! This guide will help you get started with development.

## Getting Started

### Prerequisites

- **Go** 1.24+
- **Node.js** 20+ and npm
- **Docker Desktop** (for local development)
- **Git**

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/PoyrazK/thecloud.git
   cd thecloud
   ```

2. **Create environment file**
   ```bash
   echo "DATABASE_URL=postgres://cloud:password@localhost:5432/thecloud" > .env
   ```

3. **Start infrastructure**
   ```bash
   make run
   ```

## Development Workflow

### Branch Naming

```
feature/<feature-name>    # New features
fix/<issue-description>   # Bug fixes
docs/<description>        # Documentation changes
```

### Making Changes

1. **Sync with main**
   ```bash
   git checkout main && git pull origin main
   ```

2. **Create your branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes** and commit frequently:
   ```bash
   git add .
   git commit -m "feat: description of change"
   ```

4. **Run tests** before pushing
   ```bash
   make test
   ```

5. **Update Swagger** if you changed the API
   ```bash
   make swagger
   ```

6. **Push and create PR**
   ```bash
   git push -u origin feature/your-feature-name
   ```

## Commit Message Format

| Prefix | Use Case |
|--------|----------|
| `feat:` | New feature |
| `fix:` | Bug fix |
| `refactor:` | Code refactoring |
| `test:` | Adding or updating tests |
| `docs:` | Documentation changes |
| `chore:` | Maintenance tasks |

## Code Standards

- Run `gofmt` before committing: `gofmt -w .`
- Run linting: `golangci-lint run ./...`
- Follow existing naming conventions in the codebase
- Write tests for new functionality

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run e2e tests
make test-e2e TEST_TIMEOUT=5m
```

For detailed testing information, see [docs/TESTING.md](docs/TESTING.md).

## Pull Request Process

1. Fill out the PR description with:
   - What the change does
   - Why it's needed
   - How to test it
2. Ensure all tests pass
3. Update Swagger docs if API changed
4. Request review from maintainers

## Reporting Bugs

When reporting bugs, please include:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Go version, OS, and any relevant configuration

## Getting Help

For questions or discussions, open an issue on GitHub.

## Code of Conduct

Please review our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.