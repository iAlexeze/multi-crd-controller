# Execution Model

Orkestra is built around a simple runtime boundary: **Kubernetes
resources are the canonical representation consumed by the
reconciliation system**.

Resources may originate from different sources. A target-aware gateway
may [translate higher-level intent](../self-service/05-gateway-as-delivery-layer.md) into a CR, while GitOps, CI/CD,
`kubectl`, another controller, or another Kubernetes client may produce
the CR directly.

Once the resource exists in Kubernetes, the runtime follows the same
execution model regardless of where that resource came from.

```text
                         RESOURCE SOURCES

       Target-aware        GitOps          CI/CD        kubectl
          Gateway            │               │             │
             │               │               │             │
        intent → CR          CR              CR            CR
             │               │               │             │
             └───────────────┴───────────────┴─────────────┘
                                     │
                                     ▼
                              Kubernetes API
                                     │
                                     ▼
                              Orkestra Runtime
```

The gateway is therefore not a required stage in the execution
pipeline. It is one producer of Kubernetes resources, with the
additional capability of understanding Orkestra's target model while
translating intent, enforcing admission policy, stamping provenance,
and applying field translation before the CR reaches the API server.
That provenance — target, alias, source, OIDC identity — travels on
the CR as annotations that subsequent runtime stages can observe.

---

## The effective operator box

The `OperatorBox` is one of the central composition points in the
runtime.

A resource's effective operator box is resolved from the CRD
configuration and the target associated with the resource. Runtime
features that consult the effective box automatically inherit
target-specific configuration.

```text
                         primary object
                              │
                        resolve target
                        annotation
                              │
                              ▼
                         effectiveBox
                              │
             ┌────────────────┼────────────────┐
             │                │                │
           watch[]        enqueueGate     reconcileGate
                                                │
                                           eventAware
```

This is important because runtime features do not need separate
target-resolution implementations. If a target changes the watch
configuration, the effective watch is used. If a target changes the
enqueue gate, the effective enqueue gate is used. If a target enables
`eventAware`, the resulting queue identity follows that effective
configuration.

The effective-box resolution is therefore a semantic boundary between
configuration and runtime behaviour.

---

## Sentinels and conditions

Orkestra reuses the same condition model across every part of the
system.

[`when` and `or` conditions](../../reference/schema/02-katalog/06-when-conditions.md) are not separate implementations for each
feature. They are evaluated through common condition machinery against
a resolver context.

[Sentinels](../../reference/schema/02-katalog/04-operatorbox.md#prereconcilesentinels) provide event-specific facts to that evaluation — computed
from the delta between `oldObj` and `newObj` at the moment of an
update event.

```text
Kubernetes update event
      │
      ▼
old object + new object
      │
      ▼
sentinel computation
(generationChanged, labelsChanged, annotationsChanged, ...)
      │
      ▼
resolver context
      │
      ├── when / or
      ├── sentinels
      ├── metrics
      ├── health
      ├── intent
      └── other context
```

This allows gates at different stages of the execution pipeline to use
the same underlying evaluation model. The condition machinery is not
reimplemented per feature. It is evaluated once, against the resolver,
wherever conditions are declared.

---

## From Kubernetes events to work

Once Kubernetes contains the resource, informers observe changes.

There are two relevant event sources:

1. **Primary informers** — watching the CRD's own resources.
2. **Secondary watch informers** — watching resources that may affect
   those primary resources.

Both eventually converge on the same enqueue path.

```text
                         Kubernetes events
                                │
                  ┌─────────────┴─────────────┐
                  │                           │
           primary event              secondary event
                  │                           │
             primary key                 resolve owner /
             already known              affected primary key
                  │                           │
                  └─────────────┬─────────────┘
                                ▼
                     common enqueue boundary
```

A secondary watch therefore does not implement its own queue semantics.
Its responsibility is to:

```text
observe
  ↓
apply watch / on filtering
  ↓
compute sentinels from old + new objects
  ↓
resolve affected primary key(s)
  ↓
hand key + sentinels to the common enqueue path
```

A `watch entry` does not need to know how the queue represents
event identity, how items are deduplicated, how retries work, or what
eventually reconciles the item. It provides the affected primary
identity and event context and hands control back to the informer
factory.

---

## Admission before the queue

The common enqueue path is where an observed event becomes eligible
work. Admission can take several pieces of context into account:

- namespace restrictions
- queue behaviour (`onLimit`, `onThreshold`)
- enqueue gates (`when`, `or`, sentinels, external calls)
- effective target configuration

```text
Informer event
      │
      ├── identity
      ├── sentinels
      │
      ▼
Queue behaviour
(onLimit / onThreshold — capacity-based decisions, no object context)
      │
      ▼
Enqueue gate
      │
      ├── when / or conditions
      ├── sentinel conditions
      └── external evaluation
      │
      ▼
Queue
```

Queue behaviour and the enqueue gate are evaluated in sequence. Queue
behaviour operates at the capacity level — it has no object context
and makes decisions based on queue depth alone. When conditions are
declared, queue behaviour delegates to the informer for Tier 2
evaluation, which has full object context.

The enqueue gate evaluates conditions against the resolver. If the
gate passes, the event enters the queue. If it fails, the event is
silently discarded — the object is not lost, only this event is.
The next resync or external event will re-enqueue it.

---

## Queue identity

When an admitted event reaches the workqueue, it becomes a `QueueItem`.

The queue owns the mechanics of work identity and delivery.

```text
QueueItem
 ├── Key       (object key)
 ├── GVK       (group/version/kind)
 └── EventID   (present only when eventAware: true)
```

In normal operation, client-go's deduplication coalesces multiple
events for the same object into one work item. Only the latest state
matters. This is the correct behaviour for level-triggered
reconciliation.

When `eventAware: true` is declared on `reconcileGate`, each event
receives a unique `EventID`. Two events for the same key become two
distinct `QueueItem` values and survive deduplication as separate work
items. The sentinel context computed at event time travels with each
item independently.

The architectural point: **producers of queue work do not need to
understand queue identity mechanics**. A watch informer supplies the
affected primary identity and event context. The informer factory
interprets the effective configuration. The queue implements the
identity model. Each layer knows only what its responsibility requires.

```text
watch informer
      │  affected primary identity + sentinels
      ▼
informer factory
      │  interprets effectiveBox → eventAware?
      ▼
queue
      │  implements QueueItem identity
      ▼
kordinator
```

---

## The Kordinator

The Kordinator sits between the workqueue and reconciliation admission.

It dequeues work items, resolves the effective operator box for the
item's GVK and target, evaluates preReconcile admission, and dispatches
to the MuxReconciler. It coordinates without knowing the implementation
details of the reconciler it dispatches to.

```text
Workqueue
      │
      ▼
Kordinator
      │
      ├── resolve effectiveBox (GVK + target annotation)
      ├── evaluate preReconcile admission
      └── dispatch to MuxReconciler
```

The Kordinator sees one queue item at a time. It does not know whether
that item originated from a primary informer or a secondary watch. It
does not know whether the reconciler it dispatches to is declarative,
hybrid, or a Go `Reconcile()` function. It receives a work item and
coordinates its path to reconciliation.

---

## Reconciliation admission

Getting an item into the queue does not mean that reconciliation
automatically occurs. There is a second admission boundary at
`preReconcile`.

The two gates serve different stages and answer different questions:

> **enqueueGate:** Should this event become work?

> **reconcileGate:** Should this queued work be reconciled now?

Both use the same underlying condition and sentinel machinery.

```text
Workqueue
      │
      ▼
Kordinator
      │
      ▼
preReconcile admission
      │
      └── reconcileGate
             │
             ├── when / or conditions
             ├── sentinel conditions
             ├── external evaluation
             └── eventAware declaration
             │
             ▼
        MuxReconciler
```

The full journey from event to reconciliation:

```text
Kubernetes event
      │
      ▼
event admission
      │
      ├── queue behaviour        (capacity — no object context)
      ├── enqueueGate            (conditions — full object context)
      └── sentinels              (event facts)
      │
      ▼
Workqueue
      │
      ▼
preReconcile admission
      │
      └── reconcileGate
             │
             ├── when / or
             ├── sentinels
             └── external
      │
      ▼
reconciliation
```

---

## Reconciliation

Once work passes reconciliation admission, the Kordinator dispatches
to the MuxReconciler.

The MuxReconciler reads the `serve-target` annotation on the CR and
routes to the reconciler registered for that target. When no target
annotation is present — the CR arrived via `kubectl apply` directly —
it falls back to the CRD-level reconciler.

```text
                         MuxReconciler
                              │
              reads serve-target annotation
                              │
             ┌────────────────┼────────────────┐
             │                │                │
         target A          target B       no target
         (declarative)     (hooks)        (CRD-level)
```

This allows one CRD to have multiple reconcile strategies — one per
serve target — without the Kordinator knowing anything about which
strategy is in use. The routing is internal to the MuxReconciler.

Across all paths, reconciliation can take three forms:

```text
             ┌────────────────┼────────────────┐
             │                │                │
        declarative     hybrid (hooks)       Constructor
      (100% YAML)    (90% YAML + 10% hooks)  (100% ─ Reconcile())
```

The Kordinator coordinates the work but does not need to know the
internal implementation of the selected reconciler beyond the contract
it exposes.

---

## Status, metrics, and the feedback loop

Reconciliation produces more than immediate resource changes.

The runtime stamps health and metrics onto each CR after reconcile —
borrowing the same pattern as gateway provenance injection, but
originating from the runtime rather than the delivery layer.

Those values are subsequently available to gate evaluation in future
reconcile cycles:

```text
                         reconciliation
                               │
                    ┌──────────┴──────────┐
                    │                     │
                 status           health + metrics
                                    stamped on CR
                                    exposed at /katalog
                                          │
                                          ▼
                                  resolver context
                                          │
                              ┌───────────┴───────────┐
                              │                       │
                           when / or               gates
                                          │
                                          ▼
                                   future decisions
```

This creates a feedback loop where the operator's own runtime state
becomes a gate condition. A `reconcileGate` that checks
`.metrics.queueDepth` is asking: was the operator healthy at the end
of the last cycle, and is there capacity to process this object now?

```text
configuration
      ↓
observation
      ↓
admission
      ↓
queue
      ↓
reconciliation
      ↓
status / metrics / health
      ↓
evaluation context
      ↓
admission  ←─────── feedback loop
      ↓
...
```

This is the property that makes Orkestra operators self-aware. The
operator can gate its own reconciliation based on its own state —
without external monitoring systems, without custom feedback
controllers, without anything beyond the runtime that was already
running.

---

## The complete model

```text
                         CR PRODUCERS

       Target-aware        GitOps          CI/CD        kubectl
          Gateway            │               │             │
             │               │               │             │
   intent → CR + provenance  CR              CR            CR
             │               │               │             │
             └───────────────┴───────────────┴─────────────┘
                                     │
                                     ▼
                              Kubernetes API
                                     │
                                     ▼
                                Kubernetes
                                  events
                                     │
                    ┌────────────────┴────────────────┐
                    │                                 │
             primary informer                  watch informer
                    │                                 │
             object → key                 sentinels + affected key
                    │                                 │
                    └────────────────┬────────────────┘
                                     │
                                     ▼
                            common enqueue path
                                     │
                          ┌──────────┴──────────┐
                          │                     │
                    queue behaviour        enqueueGate
                    (Tier 1: capacity)   (conditions + sentinels)
                          │                     │
                          └──────────┬──────────┘
                                     │
                                 Workqueue
                                     │
                                  QueueItem
                                (key + GVK + EventID?)
                                     │
                                     ▼
                                Kordinator
                                     │
                             resolve effectiveBox
                                     │
                                     ▼
                            preReconcile admission
                                     │
                                reconcileGate
                                     │
                         ┌───────────┼───────────┐
                         │           │           │
                      when/or    sentinels    external
                                     │
                                 eventAware
                                     │
                                     ▼
                              MuxReconciler
                              (route by serve-target)
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
                declarative         hybrid (hooks)  Constructor — Reconcile()
                    │                │                │
                    └────────────────┼────────────────┘
                                     │
                                     ▼
                               reconciliation
                                     │
                    ┌────────────────┴────────────────┐
                    │                                 │
                 status                    health + metrics
                                           stamped on CR
                                           exposed at /katalog
                                                 │
                                                 ▼
                                        evaluation context
                                                 │
                                                 └──► gates (feedback loop)
```

---

## Key properties

The model demonstrates five properties that each follow from the
convergence design:

**Convergence** — all CR producers feed the same runtime path once the
resource exists in Kubernetes. A CR from the gateway, from ArgoCD, or
from kubectl is the same thing to the runtime.

**Separation** — each layer knows only what its responsibility requires.
A watch informer does not understand queues. A queue does not understand
reconcilers. A reconcile gate does not need to know event IDs. A
reconciler does not know whether work came from a primary informer or
a secondary watch.

**Self-awareness** — reconciliation produces health and metrics that
feed back into gate evaluation. The operator can gate its own future
cycles based on its own current state. No external system required.

**Uniform conditions** — `when`, `or`, sentinels, and external calls
evaluate identically at the enqueue gate, the reconcile gate, resource
conditions, status fields, autoscaler conditions, and validation rules.
One mental model applies everywhere.

**Target transparency** — `effectiveBox` abstracts target selection from
runtime features. Watch configuration, enqueue gate, reconcile gate,
and event awareness all resolve through the effective box without
needing separate target-aware implementations in each feature.


## Where to go next

- [Event Aware Reconciliation](../reconciler-model/08-event-aware-reonciliation.md)
- [The Queue That Grew Up](/blog/the-queue-that-grew-up)

