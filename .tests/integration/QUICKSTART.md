# 🚀 Integration Test Suite - Quick Reference

## One-Liners

```bash
# Run everything
cd tests/integration && make test

# Just access tests
make test-access

# Just benchmarks
make bench

# View results
make report

# Clean up
make clean
```

## File Locations

```
fixtures/access_matrix.csv    ← Add access test cases here
fixtures/vulnerabilities.csv  ← Add test vulnerabilities here
fixtures/users.csv           ← Add test users here
reports/*.json              ← Test results appear here
```

## Adding a Test Case

```csv
# Edit: fixtures/access_matrix.csv
# Add line:
8,user@test.com,false,Reason here,category_name

# Run:
make test-access
```

## Test Categories

- `admin_access` - Admin global permissions
- `reporter_access` - Reporter can see their findings
- `assigned_access` - Assigned users on undisclosed
- `public_access` - Published vulnerabilities
- `undisclosed_restricted` - Should NOT access undisclosed
- `hidden_restricted` - Should NOT access hidden
- `global_viewer_access` - Read-only global access

## Expected Performance

| Operation | Throughput | Latency |
|-----------|------------|---------|
| List      | 200-300/s  | 30-50ms |
| Get       | 500-800/s  | 10-20ms |
| Filter    | 250-400/s  | 20-40ms |

## Troubleshooting

```bash
# Services won't start
make clean && make build && make test

# Database issues
rm -rf test-data/*.db && make test

# View logs
make logs

# Interactive debug
make shell
```

## CI/CD

```yaml
# .github/workflows/test.yml
- run: cd tests/integration && make test
- uses: actions/upload-artifact@v3
  with:
    name: reports
    path: tests/integration/reports/
```

## Makefile Commands

```bash
make help              # Show all commands
make test              # Full test suite
make test-access       # Access control only
make bench             # Benchmarks only
make report            # Display reports
make build             # Build containers
make up                # Start services
make down              # Stop services
make clean             # Clean everything
make logs              # View logs
make seed              # Seed database
make shell             # Interactive shell
make validate-fixtures # Check CSV format
```

## Quick Debugging

```bash
# Check service health
curl http://localhost:9999/health  # OIDC
curl http://localhost:8080/health  # API

# Manual test
make shell
go test -v -run TestAccessMatrix

# Single benchmark
go test -bench=BenchmarkListVulnerabilities -benchtime=5s
```

## Report Structure

```json
{
  "total_tests": 48,
  "passed": 48,
  "failed": 0,
  "pass_rate": 100.0,
  "category_stats": {...},
  "failed_tests": [...]
}
```

## Test Users

| Email              | Role          | Password    |
|--------------------|---------------|-------------|
| alice@test.com     | admin         | password123 |
| bob@test.com       | pentester     | password123 |
| charlie@test.com   | visitor       | password123 |
| diana@test.com     | pentester     | password123 |
| eve@test.com       | visitor       | password123 |
| frank@test.com     | global_viewer | password123 |

## Quick Checks

```bash
# Are fixtures valid?
make validate-fixtures

# How many tests?
wc -l fixtures/access_matrix.csv

# Latest results?
cat reports/access_test_latest.json | jq '.pass_rate'

# Benchmark summary?
cat reports/benchmark_latest.json | jq '.summary'
```

---

**Need help?** Run `make help` or check README.md
