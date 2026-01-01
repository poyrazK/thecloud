# ADR-006: Real-time Communication Strategy

## Status
Accepted

## Context
The web console (Phase 5) requires real-time updates for:
- Instance status changes (launch, stop, terminate)
- Live metrics (CPU, memory, network)
- Audit log streaming
- Resource creation/deletion notifications

We need to choose between WebSocket, Server-Sent Events (SSE), or polling.

## Decision
We will use a **hybrid approach**:

1. **SSE for metrics streaming** (`/api/dashboard/stream`)
   - Unidirectional (server → client)
   - Native browser support, auto-reconnect
   - Lighter than WebSocket for simple streaming

2. **WebSocket for bidirectional events** (`/ws`)
   - Instance lifecycle notifications
   - Audit log real-time feed
   - Future: interactive terminal (Xterm.js)

3. **REST for on-demand queries** (`/api/dashboard/*`)
   - Initial page load
   - Explicit refresh actions

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Next.js Frontend                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ useSSE hook │  │ useWS hook  │  │ useSWR (REST)       │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
└─────────┼────────────────┼─────────────────────┼────────────┘
          │                │                     │
          ▼                ▼                     ▼
    SSE Stream       WebSocket Hub          REST API
    /dashboard/      /ws                    /dashboard/
    stream                                  summary
```

## Consequences

### Positive
- SSE is simpler to implement and debug than full WebSocket
- WebSocket enables future interactive features (terminal, logs)
- Clear separation of concerns between streaming types
- Native browser reconnection for SSE

### Negative
- Two streaming protocols to maintain
- WebSocket requires explicit auth on handshake
- SSE not supported in older browsers (IE11)

## Implementation Notes
- SSE: Use Gin's `c.SSEvent()` with 5-second ticker
- WebSocket: Use gorilla/websocket with hub pattern
- Auth: Both require X-API-Key validation
