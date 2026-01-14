---
description: Generate and update Swagger API documentation
---
# Swagger Documentation Workflow

Generate and verify Swagger/OpenAPI documentation.

## Steps

1. **Install Swag (if needed)**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. **Generate Swagger Docs**
// turbo
```bash
make swagger
```

3. **Verify Docs Generated**
// turbo
```bash
ls -la docs/swagger/
```

4. **Start Server to View Docs**
```bash
make run
```

5. **Open Swagger UI**
Open browser to: http://localhost:8080/swagger/index.html

6. **Commit Updated Docs**
```bash
git add docs/swagger/
git commit -m "docs: update swagger documentation"
```

## Swagger Annotation Format
```go
// @Summary Short description
// @Description Longer description
// @Tags resourceName
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param request body RequestType true "Request body"
// @Success 200 {object} ResponseType
// @Failure 400 {object} httputil.Response
// @Router /path [method]
```
