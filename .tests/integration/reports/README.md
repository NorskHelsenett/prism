# Integration Test Reports

This directory contains generated test reports from the integration test suite.

## Report Types

### Access Control Tests
- `access_test_latest.json` - Most recent access control test results
- `access_test_YYYYMMDD_HHMMSS.json` - Timestamped historical results

### Benchmark Tests
- `benchmark_latest.json` - Most recent benchmark results
- `benchmark_YYYYMMDD_HHMMSS.json` - Timestamped historical results

### HTML Report
- `report.html` - Visual dashboard of test results (generated with `make report`)

## JSON Schema

### Access Test Report
```json
{
  "timestamp": "2025-10-14T14:32:10Z",
  "total_tests": 48,
  "passed": 48,
  "failed": 0,
  "pass_rate": 100.0,
  "results": [
    {
      "vulnerability_id": 1,
      "user_email": "alice@test.com",
      "expected": true,
      "actual": true,
      "passed": true,
      "reason": "Admin has global access",
      "test_category": "admin_access",
      "response_status": 200
    }
  ],
  "failed_tests": [],
  "category_stats": {
    "admin_access": {
      "total": 8,
      "passed": 8,
      "failed": 0,
      "pass_rate": 100
    }
  }
}
```

### Benchmark Report
```json
{
  "timestamp": "2025-10-14T14:35:22Z",
  "scenarios": [
    {
      "name": "List All Vulnerabilities",
      "operations": 100,
      "duration_ns": 12345678900,
      "mean_latency_ns": 45000000,
      "min_latency_ns": 20000000,
      "max_latency_ns": 120000000,
      "p50_latency_ns": 42000000,
      "p95_latency_ns": 78000000,
      "p99_latency_ns": 95000000,
      "throughput_ops_per_sec": 220.5
    }
  ],
  "summary": {
    "total_scenarios": 7,
    "total_operations": 700,
    "total_duration": "2m15s",
    "avg_throughput_ops_per_sec": 185.3
  }
}
```

## Viewing Reports

```bash
# Terminal-based report
make report

# Generate HTML report
./generate-html-report.sh
```

## Analyzing Results

### Tracking Performance Over Time

```bash
# Compare latest with previous
diff -u reports/benchmark_20251014_143522.json reports/benchmark_latest.json

# Extract throughput from all reports
grep -h "throughput_ops_per_sec" reports/benchmark_*.json
```

### CI/CD Integration

Upload these JSON files as artifacts in your CI pipeline for historical tracking.

Example GitHub Actions:
```yaml
- name: Upload Reports
  uses: actions/upload-artifact@v3
  with:
    name: test-reports
    path: tests/integration/reports/
```
