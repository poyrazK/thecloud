# CloudFunctions Guide

CloudFunctions allows you to run serverless code in response to API requests. It uses Docker containers to provide a secure, isolated environment for your code.

## Supported Runtimes

| Runtime | Image | Entrypoint | Extension |
|---------|-------|------------|-----------|
| `nodejs20` | `node:20-alpine` | `node` | `.js` |
| `python312` | `python:3.12-alpine` | `python` | `.py` |
| `go122` | `golang:1.22-alpine` | `go run` | `.go` |
| `ruby33` | `ruby:3.3-alpine` | `ruby` | `.rb` |
| `java21` | `eclipse-temurin:21-alpine` | `java -jar` | `.jar` |

## Creating a Function

To create a function, you need a zip file containing your code.

### Example: Node.js

1. Create `index.js`:
   ```javascript
   const payload = JSON.parse(process.env.PAYLOAD || '{}');
   console.log(`Hello, ${payload.name || 'World'}!`);
   ```

2. Zip it:
   ```bash
   zip code.zip index.js
   ```

3. Create the function:
   ```bash
   cloud fn create --name hello --runtime nodejs20 --handler index.js --code code.zip
   ```

## Updating a Function

You can update a function's configuration without redeploying code:

```bash
# Update timeout and memory
cloud fn update hello --timeout 300 --memory 256

# Update handler
cloud fn update hello --handler newhandler.js

# Set environment variables
cloud fn update hello --env FOO=bar --env DB_HOST=localhost
```

### Environment Variables

Environment variables are injected at runtime into the function container:

```bash
cloud fn update my-func --env API_KEY=secret --env DEBUG=true
```

Environment variables are available via `process.env` (Node.js), `os.environ` (Python), or `os.Getenv` (Go, Java).

### Available Update Options

| Flag | Description | Valid Range |
|------|-------------|------------|
| `--handler` | Entry point file | string |
| `--timeout` | Execution timeout (seconds) | 1–900 |
| `--memory` | Memory allocation (MB) | 64–10240 |
| `--env` | Environment variable `KEY=VALUE` | multiple |

## Invoking a Function

You can invoke a function synchronously or asynchronously.

### Synchronous Invocation (Default)

The CLI will wait for the function to complete and return the logs and exit status.

```bash
cloud fn invoke hello --payload '{"name": "Antigravity"}'
```

### Asynchronous Invocation

The API returns immediately with an invocation ID. The function runs in the background.

```bash
cloud fn invoke hello --payload '{"name": "Antigravity"}' --async
```

## Security & Isolation

Each invocation runs in a fresh Docker container with:
- **No Network Access**: The container cannot reach the internet or other services.
- **Read-Only Root Filesystem**: The container cannot modify its environment.
- **Resource Limits**: 128MB RAM and 0.5 CPU by default.
- **Process Limits**: A limit of 50 PIDs to prevent fork bombs.

## Viewing Logs

You can view the logs of the last 100 invocations:

```bash
cloud fn logs hello
```
