# Go Profiling Reference

## Benchmark Baseline

Prefer an existing workload. Do not add benchmarks during Phase A.

For package benchmarks:

```bash
go test ./path/to/package -run '^$' -bench 'BenchmarkName$' -benchmem -count=5
```

Record Go version, OS/architecture, CPU count, power mode, relevant environment, command, sample count, and background-load caveats. Keep cold-start and steady-state results separate.

For comparisons, use identical commands and conditions. Prefer `benchstat` only when it is already available; installing it requires approval. Report raw values and variance even when using a summary tool.

## CPU and Memory Profiles

Use benchmark-integrated profiles when existing benchmarks are representative:

```bash
go test ./path/to/package -run '^$' -bench 'BenchmarkName$' -cpuprofile cpu.pprof -memprofile mem.pprof
go tool pprof -top cpu.pprof
go tool pprof -top -alloc_space mem.pprof
```

Use `inuse_space` for retained heap and `alloc_space` for allocation pressure. Correlate hot functions with the measured workload. A static code pattern is not profile evidence.

For an already-authorized HTTP profile endpoint:

```bash
go tool pprof -top 'http://127.0.0.1:6060/debug/pprof/profile?seconds=30'
go tool pprof -top 'http://127.0.0.1:6060/debug/pprof/heap'
```

Do not add or expose `net/http/pprof` during Phase A. Never expose it publicly.

## Goroutines, Mutexes, Blocking, and Trace

Collect only when symptoms justify the cost:

```bash
go tool pprof -top 'http://127.0.0.1:6060/debug/pprof/goroutine'
go tool pprof -top 'http://127.0.0.1:6060/debug/pprof/mutex'
go tool pprof -top 'http://127.0.0.1:6060/debug/pprof/block'
go test ./path/to/package -run TestName -trace trace.out
go tool trace trace.out
```

Mutex and block profiles require appropriate profiling rates. If they are absent, propose instrumentation; do not edit during Phase A.

Use trace for scheduler delays, goroutine lifecycles, GC pauses, network blocking, and synchronization behavior that CPU profiles cannot explain.

## Race Detection

Use the race detector for correctness, not benchmark comparisons:

```bash
go test -race ./...
```

Race instrumentation changes timing and memory use. Never compare race-enabled performance with normal builds.

## Comparability Checklist

- Use the same revision except for the approved change.
- Use the same Go version, build flags, machine, CPU/power mode, dataset, concurrency, and external-service behavior.
- Warm caches consistently or report cold and warm results separately.
- Use multiple samples.
- Report regressions and noisy results.
- Never select only the fastest favorable run.
