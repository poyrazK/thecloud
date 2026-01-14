---
description: Build and deploy the application locally
---
# Deploy Workflow

Build, test, and deploy the application with smoke tests.

// turbo-all

## Steps

1. **Run Linting**
```bash
golangci-lint run ./...
```

2. **Run Unit Tests**
```bash
make test
```

3. **Build Binaries**
```bash
make build
```

4. **Stop Existing Services**
```bash
make stop
```

5. **Deploy with Docker Compose**
```bash
make run
```

6. **Wait for Services to Start**
```bash
sleep 5
```

7. **Check Health Endpoint**
```bash
curl -sf http://localhost:8080/health/ready && echo "✅ Health OK" || echo "❌ Health FAILED"
```

8. **Run Smoke Tests - Auth Flow**
```bash
# Register a test user
curl -sf -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@deploy.local","password":"TestPass123!","name":"Deploy Test"}' \
  && echo "✅ Register OK" || echo "⚠️ Register (may already exist)"
```

9. **Run Smoke Tests - Login**
```bash
# Login and get API key (save for next steps)
curl -sf -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@deploy.local","password":"TestPass123!"}' \
  && echo "✅ Login OK" || echo "❌ Login FAILED"
```

10. **Run Smoke Tests - List Instances**
```bash
# Test authenticated endpoint (replace YOUR_KEY with actual key from login)
API_KEY=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@deploy.local","password":"TestPass123!"}' | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
curl -sf http://localhost:8080/instances -H "X-API-Key: $API_KEY" \
  && echo "✅ Instances API OK" || echo "❌ Instances API FAILED"
```

11. **Run Smoke Tests - List VPCs**
```bash
API_KEY=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@deploy.local","password":"TestPass123!"}' | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
curl -sf http://localhost:8080/vpcs -H "X-API-Key: $API_KEY" \
  && echo "✅ VPCs API OK" || echo "❌ VPCs API FAILED"
```

12. **Check Swagger Docs Available**
```bash
curl -sf http://localhost:8080/swagger/index.html > /dev/null \
  && echo "✅ Swagger UI OK" || echo "❌ Swagger UI FAILED"
```

13. **View API Logs (Optional)**
```bash
docker compose logs api --tail=50
```

## Summary
If all smoke tests pass (✅), deployment is successful!

## Rollback
If something goes wrong:
```bash
make stop
git checkout main
make run
```

## Manual Testing Checklist
After automated tests pass, optionally verify:
- [ ] Swagger UI loads at http://localhost:8080/swagger/index.html
- [ ] Can create an instance via API
- [ ] Can create a VPC and subnet
- [ ] WebSocket connections work
