# CloudCron (Scheduled Tasks)

CloudCron allows you to run recurring tasks using standard Cron syntax.

## Internal Workings
- **Worker**: A background goroutine (`CronWorker`) polls every 10 seconds for due jobs via `ClaimNextJobsToRun`.
- **Transactional claiming**: Jobs are claimed atomically using `FOR UPDATE SKIP LOCKED` inside a `BEGIN...COMMIT` transaction. The `next_run_at` is set to a far-future value (1 year) and a `claimed_until` timestamp is recorded to prevent double-execution across workers.
- **Atomic completion**: After execution, `CompleteJobRun` atomically inserts the run record and advances `next_run_at` to the true next scheduled time — in a single transaction.
- **Crash recovery**: A reaper runs every 1 minute to reset stale claims where `claimed_until` has expired (worker died mid-execution). Those jobs are immediately re-queued.
- **Parsing**: Uses the `robfig/cron/v3` library for reliable cron expression parsing.
- **Execution**: All jobs are currently "HTTP Targets" — they trigger an HTTP/REST call to a user-defined URL.

## Features
- **Run History**: Every execution's status, code, and duration are recorded in `cron_job_runs`.
- **States**: Jobs can be `ACTIVE` or `PAUSED`.
- **Multi-worker safe**: Due to transactional claiming with claim timeouts, the same job cannot be executed by multiple workers simultaneously.

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
