# CloudFunctions Guide

CloudFunctions allows you to run serverless code in response to API requests or on a schedule. It uses Docker containers to provide a secure, isolated environment for your code.

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
- **Resource Limits**: 128MB RAM and 0.5 CPU by default.
- **Process Limits**: A limit of 50 PIDs to prevent fork bombs.

## Viewing Logs

You can view the logs of the last 100 invocations:

```bash
cloud fn logs hello
```

## Scheduled Invocation

You can schedule a function to run on a cron expression. This is useful for periodic tasks like data processing, batch jobs, or maintenance tasks.

### Creating a Schedule

```bash
cloud fn-schedule create \
  --name nightly-processing \
  --function my-function \
  --schedule "0 2 * * *"
```

The schedule expression follows standard cron format:

| Field | Values | Special Characters |
|-------|--------|-------------------|
| Minute | 0-59 | `*`, `,`, `-` |
| Hour | 0-23 | `*`, `,`, `-` |
| Day of Month | 1-31 | `*`, `,`, `-` |
| Month | 1-12 | `*`, `,`, `-` |
| Day of Week | 0-6 | `*`, `,`, `-` |

**Examples:**

- `*/5 * * * *` — Every 5 minutes
- `0 2 * * *` — Daily at 2:00 AM
- `0 9 * * 1-5` — Weekdays at 9:00 AM
- `0 */6 * * *` — Every 6 hours

### Listing Schedules

```bash
cloud fn-schedule list
```

### Pausing and Resuming

Pause a schedule to temporarily prevent invocations without deleting it:

```bash
cloud fn-schedule pause [schedule-id]
cloud fn-schedule resume [schedule-id]
```

### Viewing Run History

```bash
cloud fn-schedule logs [schedule-id]
```

### Deleting a Schedule

```bash
cloud fn-schedule rm [schedule-id]
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/function-schedules` | Create a schedule |
| `GET` | `/function-schedules` | List all schedules |
| `GET` | `/function-schedules/:id` | Get a schedule |
| `DELETE` | `/function-schedules/:id` | Delete a schedule |
| `POST` | `/function-schedules/:id/pause` | Pause a schedule |
| `POST` | `/function-schedules/:id/resume` | Resume a schedule |
| `GET` | `/function-schedules/:id/runs` | Get run history |
