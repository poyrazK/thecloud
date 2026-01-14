---
description: Run tests with coverage report and analysis
---
# Test Coverage Workflow

Run comprehensive tests with coverage reporting.

## Steps

1. **Run All Tests with Coverage**
// turbo
```bash
go test -coverprofile=coverage.out ./...
```

2. **View Coverage Summary**
// turbo
```bash
go tool cover -func=coverage.out | tail -20
```

3. **Generate HTML Report**
// turbo
```bash
go tool cover -html=coverage.out -o coverage.html
```

4. **Open Coverage Report**
```bash
open coverage.html
```

5. **Run Specific Package Tests**
For focused testing:
```bash
go test -v -cover ./internal/core/services/...
```

6. **Run With Race Detector**
For concurrency issues:
```bash
go test -race ./...
```

7. **Cleanup**
// turbo
```bash
rm -f coverage.out coverage.html
```

## Coverage Targets
- Services: Target 70%+
- Handlers: Target 65%+
- Repositories: Target 60%+
