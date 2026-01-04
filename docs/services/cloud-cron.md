# CloudCron (Scheduled Tasks)

CloudCron allows you to run recurring tasks using standard Cron syntax.

## Internal Workings
- **Worker**: A background goroutine (`CronWorker`) polls the database every 10 seconds for jobs whose `next_run_at <= NOW()`.
- **Parsing**: Uses the `robfig/cron/v3` library for reliable cron expression parsing.
- **Execution**: All jobs are currently "HTTP Targets" - they trigger an HTTP/REST call to a user-defined URL.

## Features
- **Run History**: Every execution's status, code, and duration are recorded in `cron_job_runs`.
- **States**: Jobs can be `ACTIVE` or `PAUSED`.

## CLI Usage
```bash
# Create a job (Daily 3AM)
cloud cron create cleanup "0 3 * * *" "http://my-api/cleanup" -X POST

# List/Status
cloud cron list

# Control
cloud cron pause <job-id>
cloud cron resume <job-id>
```
