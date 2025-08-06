# Flight Search

This project implements a **simulated flight search backend** using:

- Go Fiber for HTTP and SSE handling
- Redis Streams for event-driven communication
- OpenTelemetry (OTel) + Jaeger for distributed tracing

The system consists of two services:

- **Main Service**: Handles HTTP requests, publishes search queries, and streams results via SSE
- **Provider Service**: Subscribes to Redis stream, processes requests, and publishes matching flight results

## 📣 Event-Driven Architecture

This system adopts an **event-driven architecture**, where services communicate by publishing and consuming events via Redis Streams. This pattern allows asynchronous, decoupled, and real-time interaction between services.

## 🧱 Architecture Flow

```text
Client (Browser)
│
│ 1. POST /api/flights/search
▼
Main Service
│
│ 2. XADD flight.search.requested (with trace_context)
▼
Redis Stream
│
│ 3. XREADGROUP by Provider Service
▼
Provider Service
│
│ 4. Simulate processing and search
│ 5. XADD flight.search.results:{search_id}
▼
Main Service (SSE Handler)
│
│ 6. XREAD from flight.search.results:{search_id}
▼
Client receives results via SSE
```   

## Assumptions

- The provider service works in a **finite processing flow** per request.
- Once a search is requested, the provider processes it, publishes intermediate and final results via Redis stream, and **ends with a `status: completed`**.
- The SSE stream in the main service acts as a **replayer**, forwarding all messages related to a specific `search_id` to the client via Server-Sent Events.
- Once a `status: completed` message is received, the main service will **terminate the SSE connection** immediately.
- Keeping the connection open after that point is considered wasteful and may **consume unnecessary resources**.
- This pattern is **not suitable** for continuous data (e.g. stock prices or real-time sensors). For that, we would need to implement **infinite streams** that stay open.
- To optimize Redis stream efficiency and lookup time:
  - Using a **shared stream** (e.g. `flight.search.results`) would require **filtering based on `search_id`**, increasing overhead and complexity.
  - Therefore, each stream is **named using the `search_id`**, e.g., `flight.search.results:{search_id}`.
  - This avoids filtering logic entirely and ensures **faster reads, clean message delivery, and simpler cleanup**.

## Flow Overview

1. Client sends a flight search request to the provider.
2. The provider:
   - Creates a Redis stream with a unique name: `flight.search.results:{search_id}`
   - Publishes messages to the stream (e.g., `status: processing`, `status: result`, `status: completed`)
   - Optionally sets a TTL (e.g. 30 minutes) using `EXPIRE`
3. The main service opens an SSE connection:
   - Reads the stream using `XREADGROUP`
   - Pushes each message to the client
   - Closes the connection when `status: completed` is received
   - Deletes the stream if needed


## 🔧 Optional Cleanup

After sending `status: completed`, the main service should:

- Call `DEL` on the stream key: `flight.search.results:{search_id}`
- Or set `EXPIRE` with a reasonable TTL (e.g., 30 minutes)

[Back to README](README.md)
