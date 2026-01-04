# CloudQueue (Message Queue Service)

CloudQueue is a distributed message queuing service designed for reliable asynchronous communication between microservices.

## Architecture
- **Storage**: PostgreSQL is used to ensure message persistence and ACID compliance.
- **Visibility Timeout**: Messages remain in the queue but are "hidden" from other consumers once received, until either deleted or the timeout expires.
- **Multi-tenancy**: Every queue is scoped to a `user_id`.

## Logic
Messages are stored in the `queue_messages` table. When `ReceiveMessages` is called:
1. It looks for messages where `visible_at <= NOW()`.
2. It sets `visible_at = NOW() + visibility_timeout`.
3. It returns the message to the consumer.

## CLI Usage
```bash
# Create a queue
cloud queue create my-tasks --visibility 30

# Send a message
cloud queue send <queue-id> "hello world"

# Receive messages
cloud queue receive <queue-id> --count 5
```

## SDK Usage
```go
q, _ := client.CreateQueue("my-tasks", 30)
client.SendMessage(q.ID, "data")
msgs, _ := client.ReceiveMessages(q.ID, 1)
```
