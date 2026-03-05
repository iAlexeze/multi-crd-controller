# Architectural Deep Dive

This document explains the internal architecture of the Project Controller in detail. It covers the control loop, component lifecycle, leader election model, and the design decisions that make the controller highly available, predictable, and production‑ready.

---

## Controller Lifecycle

The controller follows the same lifecycle pattern used by Kubernetes’ built‑in controllers: a clear separation between **bootstrap** and **reconciliation**.

### Bootstrap Phase (`Start()`)

The bootstrap phase runs in **every pod**, regardless of leadership. It performs all infrastructure initialization:

- Starts informers and begins watching CRD resources.
- Initializes the workqueue.
- Starts the informer controller (`ctrl.Run(wait.NeverStop)`).
- Waits for the informer cache to sync.
- Ensures follower pods always maintain a warm cache.

This design ensures that when a new leader is elected, it can begin reconciling immediately without waiting for cache warm‑up.

### Reconciliation Phase (`Run()`)

The reconciliation phase runs **only in the leader**. It is responsible for:

- Starting N worker goroutines.
- Pulling items from the workqueue.
- Executing reconciliation logic.
- Respecting context cancellation.
- Draining in‑flight reconciliations when leadership is lost.

Workers stop cleanly when the leader election context is cancelled, ensuring no partial or duplicate reconciliations.

---

## Post‑Start Hooks

The controller manager supports **post‑start hooks**, which allow components that depend on fully initialized infrastructure to start only after all core components are running.

Leader election is started using a post‑start hook:

```go
manager.AddPostStartHook(func(ctx context.Context) {
    leader.Start(ctx)
})
```

This guarantees:

- The Kubernetes client is initialized.
- The event recorder is ready.
- Informers are running and caches are synced.

Post‑start hooks avoid import cycles and startup ordering issues, and they mirror the extensibility model used by controller‑runtime’s `AddRunnable`.

---

## Leader Election Model

The controller uses Kubernetes’ coordination API to ensure that only one pod performs reconciliation at any time.

### Key Properties

- **Only the leader runs workers** (`Run()`).
- **All pods run informers** (`Start()`).
- **Failover is instant** because followers maintain warm caches.
- **Lease is released on shutdown** for fast leadership transitions.
- **Events are emitted** for visibility into leadership changes.

### Leadership Loss and Draining

When leadership is lost:

1. The leader election context is cancelled.
2. Workers stop accepting new items.
3. The queue is shut down.
4. In‑flight reconciliations finish.
5. The controller exits cleanly.

This prevents double‑processing and ensures consistency.

---

## Why Raw client‑go Instead of controller‑runtime?

This controller intentionally uses **client‑go** directly rather than controller‑runtime. The goals are:

- Full visibility into the control loop.
- Explicit lifecycle management.
- No hidden abstractions.
- Lightweight, dependency‑minimal design.
- Ideal for learning and debugging.
- Flexible enough for custom lifecycle behavior (e.g., post‑start hooks, custom draining).

controller‑runtime is excellent for large operators, but it hides many details behind abstractions. This implementation is ideal for developers who want to understand and control the underlying mechanics.

---

## Component Responsibilities

### Manager

The manager orchestrates:

- Ordered startup of components.
- Post‑start hooks.
- Graceful shutdown on SIGINT/SIGTERM.
- Readiness signaling via the health server.

It does **not** start the controller workers; that is delegated to leader election.

### Informer Layer

The informer layer provides:

- List/Watch on Project CRDs.
- Local cache for fast reads.
- Workqueue for event processing.
- Automatic resync.

Informers run in all pods to ensure fast failover.

### Controller

The controller is responsible for:

- Worker lifecycle.
- Reconciliation logic.
- Finalizer handling.
- Event recording.
- Rate limiting and retries.
- Draining on leadership loss.

### Leader Election

Leader election ensures:

- Only one pod reconciles at a time.
- Followers stay warm.
- Failover is fast and predictable.
- Leadership transitions are visible via events.

---

## Performance and Scaling

### Horizontal Scaling

Multiple replicas can run simultaneously:

- Only one pod becomes leader.
- Followers maintain warm caches.
- Failover is instant.

### Vertical Scaling

You can tune:

- Worker count (`WORKERS`)
- Resync period (`DEFAULT_RESYNC`)
- Queue rate limiting
- Reconciliation concurrency

### Queue Behavior

The workqueue provides:

- Exponential backoff on errors.
- Deduplication of keys.
- Shutdown‑aware draining.

This ensures stable performance under load.

---

## Testing Strategy

The architecture supports testing at multiple layers:

- Unit tests for reconciliation logic using fake clients.
- Informer tests using fake watch streams.
- Queue tests verifying retry behavior.
- Leader election tests simulating leadership loss.
- Integration tests using envtest or a local cluster.

---

## Roadmap

Planned enhancements include:

- Prometheus metrics for reconciliation duration and queue depth.
- Multi‑CRD support.
- Admission webhooks for validation and defaulting.
- Status subresource improvements.
- Optional controller‑runtime compatibility layer.
- Pluggable reconciliation modules for external systems.
