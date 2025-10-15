# Benchmark Configuration

This directory contains benchmark scenario configurations and results.

## Running Benchmarks

```bash
# Run all benchmarks
make bench

# Run specific benchmark
go test -bench=BenchmarkListVulnerabilities -benchtime=30s

# With CPU and memory profiling
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Benchmark Scenarios

The suite includes read-heavy benchmarks optimized for SQLite:

1. **List All Vulnerabilities** - Tests pagination and filtering
2. **Get Single Vulnerability** - Tests single record retrieval
3. **Filter Published Vulnerabilities** - Tests WHERE clause performance
4. **Concurrent Reads** - Tests SQLite's parallel read capability
5. **Project List** - Tests join performance
6. **Access Control Check** - Measures permission overhead
7. **Dashboard Load** - Simulates real-world multi-query patterns

## Performance Baselines

Expected results on modern hardware (4-core CPU, SSD):

| Scenario | Throughput | Mean Latency | P95 Latency |
|----------|------------|--------------|-------------|
| List Vulnerabilities | 200-300 ops/s | 30-50ms | 70-100ms |
| Get Single | 500-800 ops/s | 10-20ms | 30-50ms |
| Filter Published | 250-400 ops/s | 20-40ms | 60-90ms |
| Concurrent Reads | Scales with cores | - | - |
| Project List | 300-500 ops/s | 15-30ms | 40-70ms |
| Access Check | 400-600 ops/s | 10-25ms | 35-60ms |
| Dashboard Load | 100-150 ops/s | 60-100ms | 120-180ms |

## SQLite Optimizations

These benchmarks assume:
- WAL mode enabled
- Proper indexes on foreign keys
- Read-heavy workload optimization
- Reasonable dataset size (<10,000 records)

For large datasets, consider:
- Adding composite indexes
- Query result caching
- Connection pooling
- Read replicas

## Regression Testing

Track performance over time:

```bash
# Baseline
make bench > baseline.txt

# After changes
make bench > current.txt

# Compare
benchstat baseline.txt current.txt
```

## CI/CD Integration

Add performance gates to your pipeline:

```yaml
- name: Run Benchmarks
  run: make bench
  
- name: Check Performance
  run: |
    # Fail if throughput drops below threshold
    jq -e '.summary.avg_throughput_ops_per_sec > 150' reports/benchmark_latest.json
```
