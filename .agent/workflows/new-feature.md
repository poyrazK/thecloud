---
description: Start developing a new feature with proper git workflow
---
# New Feature Workflow

Start a new feature with proper branch management and frequent commits.

## Steps

1. **Sync with Main**
// turbo
```bash
git checkout main && git pull origin main
```

2. **Create Feature Branch**
```bash
git checkout -b feature/<feature-name>
```

3. **Make Initial Commit**
After making first changes:
```bash
git add .
git commit -m "feat: initial setup for <feature-name>"
```

4. **Run Tests Frequently**
// turbo
```bash
make test
```

5. **Commit Often**
After each logical unit of work:
```bash
git add .
git commit -m "feat: <description of change>"
```

6. **Update Swagger (if API changed)**
// turbo
```bash
make swagger
```

7. **Push to Remote**
```bash
git push -u origin feature/<feature-name>
```

8. **Create PR**
Open GitHub/GitLab and create a Pull Request to `main`.

## Commit Message Format
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks
