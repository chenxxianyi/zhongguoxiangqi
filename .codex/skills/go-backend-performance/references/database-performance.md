# Database Performance Reference

## Slow SQL Evidence

Prefer, in order:

1. Measured query latency and query count from the target workload.
2. Existing slow-query logs or tracing.
3. Read-only `EXPLAIN` or `EXPLAIN ANALYZE` in a verified safe environment.
4. Code-based inference labeled as an unverified hypothesis.

Inspect GORM for queries inside loops, per-row association loading, broad `Preload`, unbounded result sets, missing pagination, `SELECT *`, unnecessary counts, unstable sorting, and transactions held across remote I/O.

Do not execute user-provided SQL or mutate data. Treat `EXPLAIN ANALYZE` as potentially executing the query; require verified read-only safety and bounded cost.

Record SQL shape with sensitive literals redacted. Never copy DSNs, credentials, tokens, or private query parameters into a report.

## Query Plan Review

Check access type, chosen indexes, examined rows, filtering, temporary tables, filesort, join order, and cardinality estimates. A missing index is a hypothesis until query shape, selectivity, write cost, and plan evidence support it.

Do not recommend an index without discussing:

- Target query and columns.
- Equality/range/order behavior.
- Selectivity and expected rows.
- Write amplification and storage cost.
- Redundant or overlapping indexes.
- Before/after query-plan and latency validation.

## Connection Pool

Inspect:

- `SetMaxOpenConns`.
- `SetMaxIdleConns`.
- `SetConnMaxLifetime`.
- `SetConnMaxIdleTime`.
- Connection, read, write, and request timeouts.
- Database connection capacity and proxy limits.

Collect `sql.DBStats` when already exposed:

- `OpenConnections`.
- `InUse`.
- `Idle`.
- `WaitCount`.
- `WaitDuration`.
- `MaxIdleClosed`.
- `MaxIdleTimeClosed`.
- `MaxLifetimeClosed`.

Relate pool waits to workload concurrency and database saturation. Never apply a universal multiplier or recommend a concrete pool size from configuration alone.

## Concurrency and Locks

Inspect bounded goroutine creation, ownership, context cancellation, channel backpressure, lock scope, shared maps, duplicate work, serialization points, and remote I/O inside critical sections.

Require workload, mutex, block, goroutine, or trace evidence before replacing synchronization. Run the race detector separately for correctness.

## Cache Decision

Recommend a cache only when repeated expensive work is measured. First consider request coalescing, duplicate suppression, batching, better query shape, or smaller payloads.

Every cache proposal must define:

- Key and value.
- Process, request, or distributed scope.
- TTL and invalidation trigger.
- Stale-data and consistency policy.
- Memory/cardinality bound.
- Stampede prevention.
- Hit/miss/eviction observability.
- Correctness and operational fallback.

Do not present a cache as free speed. Include invalidation complexity and failure behavior.
