# Go Backend Performance Review: <scope>

**Date:** YYYY-MM-DD  
**Mode:** Phase A — analysis only  
**Repository/Service:** <name>  
**Environment:** <local/test/staging/production and relevant limits>

## Executive Summary

Summarize measured bottlenecks, highest-priority hypotheses, blocked checks, and whether the available evidence supports implementation.

## Scope and Safety

- Requested scope:
- Workload:
- Persistent data safety:
- Secrets handling:
- Exclusions:

## Evidence Ledger

| ID | Stage | Evidence type | Command/source | Status |
|---|---|---|---|---|
| E-01 | Benchmark | Measured fact / inference / hypothesis | `<command or file>` | Collected / blocked / not applicable |

## 1. Benchmark

Document the workload, command, samples, environment, time/op or latency, throughput, allocations, and variability. If unavailable, state the missing instrumentation and proposed baseline.

## 2. pprof

Document CPU, heap/allocation, goroutine, mutex, block, or trace evidence. Separate collected evidence from proposed instrumentation.

## 3. Slow SQL

Document query counts/timing, slow-query logs, read-only `EXPLAIN` evidence, suspected N+1 patterns, and index hypotheses.

## 4. Connection Pool

Document configured limits and observed `sql.DBStats` values, especially waits and saturation. Do not recommend a numeric pool change without workload and database-capacity evidence.

## 5. Concurrency and Locks

Document goroutine ownership, bounded concurrency, cancellation, race findings, mutex/block evidence, and blocking I/O.

## 6. Cache

Document measured duplicate work. For any cache proposal, define key, scope, TTL, invalidation, consistency, memory bound, and observability.

## 7. Before/After Protocol

Define the exact commands, workload, environment, sample count, metrics, and acceptable variance to use after an approved implementation.

## Prioritized Findings

| ID | Priority | Classification | Symptom | Evidence | Proposed change | Risk | Validation |
|---|---|---|---|---|---|---|---|
| P-01 | P0/P1/P2/P3 | Measured fact / inference / hypothesis | | | | | |

## Proposed Implementation Order

List only scoped recommendations. Include dependencies and rollback considerations.

## Evidence Gaps and Blocked Checks

List missing benchmarks, profiles, workloads, permissions, metrics, or safe environments. State the minimum instrumentation required.

## Approval Gate

No application source, tests, dependencies, configuration, schema, or persistent data were changed during Phase A. Review the findings and explicitly approve the finding IDs to enter Phase B implementation and validation.
