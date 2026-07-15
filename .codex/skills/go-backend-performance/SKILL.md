---
name: go-backend-performance
description: Use when analyzing or validating performance of Go backends with slow APIs, high latency, low throughput, high CPU or memory use, excessive allocations, slow SQL, GORM N+1 queries, database connection waits, goroutine or lock contention, or suspected cache opportunities.
---

# Go Backend Performance

## Overview

Find bottlenecks from reproducible evidence, not intuition. Follow every stage in order and distinguish measured facts, code-based inferences, and unverified hypotheses.

**Default to Phase A: analyze and report only. Do not change application code until the user reviews the report and explicitly approves named findings.**

**REQUIRED BACKGROUND:** Use `systematic-debugging` for root-cause discipline.

**REQUIRED FOR PHASE B:** Use `test-driven-development` and `verification-before-completion`.

## Authorization Gate

### Phase A — Analysis Only

Allow:

- Read source, examples, tests, migrations, logs, and documentation.
- Run existing non-mutating tests, benchmarks, and safe local profiles.
- Inspect existing slow-query evidence, read-only plans, pool configuration, and metrics.
- Keep profile artifacts in the system temporary directory.
- Write the final report under `docs/performance/`.

Do not:

- Modify source, tests, benchmarks, configuration, `.env`, dependencies, migrations, schemas, or persistent data.
- Add or expose profiling endpoints.
- Install tools without approval.
- Run load tests or database commands until target safety is verified.
- Claim improvement without comparable before/after measurements.

Treat missing instrumentation as a finding. Describe the minimum proposed instrumentation, then continue to the next stage without adding it.

Finish Phase A by writing the report and stopping for explicit approval. Approval applies only to finding IDs the user names or clearly selects.

### Phase B — Approved Changes

Enter Phase B only after the user reviews the Phase A report and explicitly approves implementation.

For each approved finding:

1. Preserve or create a repeatable baseline.
2. Write regression or behavior tests first.
3. Make the smallest targeted change.
4. Run correctness, race, and relevant performance checks.
5. Repeat the same measurement under comparable conditions.
6. Report raw before/after values, variance, regressions, and trade-offs.

Do not implement unapproved findings from the same report.

## Fixed Workflow

Execute every stage in this order. Mark a stage `blocked` or `not applicable` with a reason; never silently skip or reorder it.

| Stage | Required output |
|---|---|
| 1. Benchmark | Workload, command, environment, samples, baseline metrics, or instrumentation gap |
| 2. pprof | CPU and memory evidence; conditional goroutine/mutex/block/trace evidence, or gap |
| 3. Slow SQL | Query count/timing, logs/plans, GORM patterns, and evidence classification |
| 4. Connection pool | Limits, `sql.DBStats`, database-capacity relationship, or metric gap |
| 5. Concurrency/locks | Goroutine, cancellation, blocking, race, mutex/trace evidence, or gap |
| 6. Cache | Measured repeated work and full cache contract, or `not justified` |
| 7. Before/after | Phase A comparison protocol; Phase B comparable results |

### 1. Benchmark

Identify existing benchmark and endpoint workload commands. Establish reproducibility before interpreting numbers. Record latency, throughput, and allocations when applicable. Separate cold start, external API, disk, and database effects.

### 2. pprof

Start with CPU and heap/allocation profiles from the representative workload. Use goroutine, mutex, block, or trace profiles only when symptoms require them. Never optimize a static hot-looking function without workload correlation.

Read [references/go-profiling.md](references/go-profiling.md) before running profiling commands or comparing measurements.

### 3. Slow SQL

Prefer measured query timing/counts, existing slow logs, and safe read-only plans. Label N+1, missing-index, and query-shape observations as hypotheses until measured.

### 4. Connection Pool

Inspect pool configuration and `sql.DBStats`. Relate waits to workload concurrency and database capacity. Never infer a numeric pool size from configuration alone.

### 5. Concurrency and Locks

Inspect bounded concurrency, goroutine ownership, channels, context cancellation, shared state, critical-section I/O, and blocking. Race results are correctness evidence, not performance measurements.

### 6. Cache

Recommend caching only for measured repeated expensive work. Define key, scope, TTL, invalidation, consistency, memory bound, stampede behavior, observability, and fallback.

Read [references/database-performance.md](references/database-performance.md) before database, pool, concurrency, or cache conclusions.

### 7. Before/After Comparison

In Phase A, record the baseline and exact future comparison protocol. In Phase B, keep workload, machine, Go version, build flags, dataset, concurrency, warm-up, and external behavior comparable. Report samples, variability, absolute differences, percentages, regressions, and uncertainty.

## Evidence Rules

Classify every finding:

- **Measured fact:** Directly supported by a command, profile, metric, log, or plan.
- **Code-based inference:** Plausible from source but not measured.
- **Unverified hypothesis:** Requires instrumentation, access, or a safe workload.

Never invent values or expected percentages. Redact credentials, DSNs, tokens, and sensitive parameters. Do not print or copy `.env` values.

## Report

Copy [assets/performance-report-template.md](assets/performance-report-template.md) to:

`docs/performance/YYYY-MM-DD-<scope>-performance-review.md`

Include commands, environment, evidence ledger, all seven stages, prioritized finding IDs, risks, validation methods, blocked checks, and the approval gate.

The final statement must say no application source was changed during Phase A and ask the user to approve specific finding IDs before Phase B.

## Red Flags — Stop

- “The fix is obvious, so measurement can come later.”
- “Adding a benchmark or pprof endpoint is only instrumentation.”
- “A larger connection pool is always faster.”
- “The code pattern proves this is the bottleneck.”
- “The user asked for analysis, but a one-line change is harmless.”
- “One favorable run proves improvement.”
- “Cache invalidation can be decided during implementation.”

These are evidence or authorization failures. Return to the current stage, document the gap, or stop at the approval gate.

## Common Mistakes

| Mistake | Required correction |
|---|---|
| Skip unavailable stages | Mark them blocked and propose minimum instrumentation |
| Treat `-race` timings as performance | Run race checks separately |
| Tune the pool from config alone | Require waits, concurrency, and database capacity |
| Recommend cache from intuition | Measure duplicate expensive work first |
| Compare different workloads | Re-run under comparable conditions |
| Change code during Phase A | Revert the unauthorized Skill-produced change and report only |
