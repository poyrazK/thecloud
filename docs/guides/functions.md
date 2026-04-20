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
- **Resource Limits**: 128MB RAM and 0.5 CPU by default. Adjust with `cloud fn update --memory`.
- **Process Limits**: A limit of 50 PIDs to prevent fork bombs.

## Updating a Function

You can update a function's handler, timeout, memory allocation, and environment variables without recreating it.

### Update Timeout and Memory

```bash
cloud fn update hello --timeout 60 --memory 256
```

Timeout must be between 1 and 900 seconds. Memory must be between 64 and 10240 MB.

### Update Environment Variables

Environment variables are injected into the container at runtime and accessible via `process.env` (Node.js) or `os.environ` (Python).

```bash
cloud fn update hello --env DATABASE_URL=postgres://localhost/mydb --env DEBUG=true
```

To clear all environment variables, pass an empty env list:
```bash
cloud fn update hello --env ""
```

## Viewing Logs

You can view the logs of the last 100 invocations:

```bash
cloud fn logs hello
```

## Environment Variables

All functions have access to the following built-in environment variable:

| Variable | Description |
|----------|-------------|
| `PAYLOAD` | The JSON payload passed during invocation |

Custom environment variables are set at the function level via `cloud fn update --env KEY=VALUE`. They are available at runtime just like `PAYLOAD`.

Functions run with a completely isolated network and a read-only filesystem. Environment variables are the only way to pass configuration to your function at runtime.
