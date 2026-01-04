# CloudNotify (Pub/Sub Service)

CloudNotify is a highly scalable Pub/Sub service for many-to-many communication.

## Architecture
- **Topics**: Logical channels identified by ARNs (`arn:thecloud:notify:local:{user}:topic/{name}`).
- **Fan-out Logic**: When a message is published, `NotifyService` fetches all subscribers and launches a goroutine for each delivery.
- **Protocols**: 
    - `webhook`: Delivers via HTTP POST.
    - `queue`: Delivers directly into a CloudQueue.

## Technology
- **Concurrency**: Go goroutines for non-blocking parallel delivery.
- **Persistence**: PostgreSQL for topics and subscription registry.

## CLI Usage
```bash
# Create topic
cloud notify create-topic engine-updates

# Subscribe a webhook
cloud notify subscribe <topic-id> --protocol webhook --endpoint http://my-api/hook

# Subscribe a queue
cloud notify subscribe <topic-id> --protocol queue --endpoint <queue-id>

# Publish
cloud notify publish <topic-id> "Update available"
```
