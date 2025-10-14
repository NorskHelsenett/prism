# Prism Integration Test Suite

A comprehensive, easy-to-manage integration test suite for the Prism vulnerability management platform, featuring CSV-based access control testing and performance benchmarking.

## 🎯 Features

- **CSV-Based Test Configuration**: Human-readable test cases in CSV format
- **Access Control Matrix**: Comprehensive permission testing across all user roles
- **Benchmark Suite**: Read-heavy performance tests optimized for SQLite
- **Mock OIDC Server**: Isolated authentication testing without external dependencies
- **Docker-Based**: Fully containerized for consistent testing environments
- **JSON Reports**: Machine-readable test and benchmark results
- **Easy to Manage**: Simple Makefile commands for all operations

## 📁 Directory Structure

```
tests/integration/
├── docker-compose.test.yml    # Container orchestration
├── Makefile                   # Easy command execution
├── config.test.yaml           # Test configuration for Prism API
├── go.mod                     # Go module definition
├── helpers.go                 # CSV parsers and database utilities
├── access_test.go             # Access control test runner
├── benchmark_test.go          # Performance benchmark tests
├── seed.go                    # Database seeding utility
├── oidc-mock/                 # Mock OIDC provider
│   ├── server.go
│   ├── Dockerfile
│   ├── go.mod
│   └── users.yaml
├── fixtures/                  # Test data in CSV format
│   ├── users.csv
│   ├── projects.csv
│   ├── vulnerabilities.csv
│   └── access_matrix.csv
├── benchmark/                 # Benchmark configurations
└── reports/                   # Generated test reports (JSON)
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional, but recommended)
- jq (for report formatting, optional)

### Running Tests

```bash
# Navigate to integration tests directory
cd tests/integration

# Run full test suite
make test

# Run only access control tests
make test-access

# Run benchmark tests
make bench

# View reports
make report
```

## 📊 Test Fixtures

### Users (fixtures/users.csv)

Test users with different roles:

| Email               | Role          | Description                    |
|---------------------|---------------|--------------------------------|
| alice@test.com      | admin         | Full administrative access     |
| bob@test.com        | pentester     | Project and findings access    |
| charlie@test.com    | visitor       | Read-only access              |
| diana@test.com      | pentester     | Another pentester user        |
| eve@test.com        | visitor       | Another visitor user          |
| frank@test.com      | global_viewer | Read-only access to everything |

### Projects (fixtures/projects.csv)

Test projects with various configurations.

### Vulnerabilities (fixtures/vulnerabilities.csv)

8 test vulnerabilities with different:
- Visibility levels (published, undisclosed, hidden)
- Severities (critical, high, medium, low)
- Assignment states
- Projects

### Access Matrix (fixtures/access_matrix.csv)

**48 test cases** covering:
- Admin global access
- Reporter access
- Assigned user access
- Public vs. restricted visibility
- Cross-project access restrictions
- Global viewer permissions

Each row specifies:
- `vulnerability_id`: Which vulnerability to test
- `user_email`: Which user to test as
- `should_access`: Expected outcome (true/false)
- `reason`: Human-readable explanation
- `test_category`: Category for statistics

## 🧪 Access Control Testing

The access matrix tests verify that:

1. **Admins** can access all vulnerabilities
2. **Reporters** can access their own findings
3. **Assigned users** can access undisclosed vulnerabilities they're assigned to
4. **Non-assigned users** cannot access undisclosed/hidden vulnerabilities
5. **Published vulnerabilities** are accessible to all authenticated users
6. **Global viewers** have read-only access to everything

### Example Test Cases

```csv
vulnerability_id,user_email,should_access,reason
1,alice@test.com,true,Admin has global access
2,bob@test.com,true,Reporter of the vulnerability
2,charlie@test.com,true,Explicitly assigned to this undisclosed vuln
2,diana@test.com,false,Not assigned - undisclosed vulnerability
```

## 📈 Benchmark Tests

Read-heavy performance tests optimized for SQLite:

1. **List All Vulnerabilities**: Tests query performance
2. **Get Single Vulnerability**: Tests single record retrieval
3. **Filter Published Vulnerabilities**: Tests WHERE clause performance
4. **Concurrent Reads**: Tests SQLite's handling of parallel reads
5. **Project List**: Tests join performance
6. **Access Control Check**: Measures permission check overhead
7. **Dashboard Load**: Simulates multiple concurrent API calls

### Benchmark Metrics

Each benchmark measures:
- Mean latency
- P50, P95, P99 percentiles
- Min/Max latency
- Throughput (ops/sec)
- Memory allocations
- Bytes per operation

## 📝 Reports

### Access Test Report (JSON)

```json
{
  "timestamp": "2025-10-14T14:32:10Z",
  "total_tests": 48,
  "passed": 48,
  "failed": 0,
  "pass_rate": 100.0,
  "results": [...],
  "failed_tests": [],
  "category_stats": {
    "admin_access": {"total": 8, "passed": 8, "failed": 0, "pass_rate": 100},
    "reporter_access": {"total": 8, "passed": 8, "failed": 0, "pass_rate": 100},
    ...
  }
}
```

### Benchmark Report (JSON)

```json
{
  "timestamp": "2025-10-14T14:35:22Z",
  "scenarios": [
    {
      "name": "List All Vulnerabilities",
      "operations": 100,
      "mean_latency_ns": 45000000,
      "p95_latency_ns": 78000000,
      "throughput_ops_per_sec": 220.5
    },
    ...
  ],
  "summary": {
    "total_scenarios": 7,
    "total_operations": 700,
    "avg_throughput_ops_per_sec": 185.3
  }
}
```

## 🛠️ Makefile Commands

| Command                | Description                                    |
|------------------------|------------------------------------------------|
| `make help`            | Show all available commands                    |
| `make test`            | Run full integration test suite                |
| `make test-access`     | Run only access control tests                  |
| `make bench`           | Run benchmark tests                            |
| `make report`          | Display test reports in terminal               |
| `make build`           | Build all Docker containers                    |
| `make up`              | Start services (API + OIDC mock)               |
| `make down`            | Stop all services                              |
| `make clean`           | Clean up test data and reports                 |
| `make logs`            | View logs from all services                    |
| `make seed`            | Seed database with fixtures                    |
| `make shell`           | Open interactive shell in test container       |
| `make validate-fixtures` | Validate CSV fixture format                  |

## 🔧 Advanced Usage

### Running Specific Tests

```bash
# Run only a specific test
docker-compose -f docker-compose.test.yml run --rm test-runner \
  go test -v -run TestAccessMatrix

# Run benchmarks with custom duration
docker-compose -f docker-compose.test.yml run --rm benchmark-runner \
  go test -bench=. -benchtime=30s
```

### Modifying Test Data

1. Edit CSV files in `fixtures/` directory
2. Re-run tests - no code changes needed!

```bash
# Example: Add new access test case
echo "8,newuser@test.com,false,Not a project member,access_denied" >> fixtures/access_matrix.csv

# Run tests with new case
make test-access
```

### Custom Configuration

Edit `config.test.yaml` to change:
- OIDC provider settings
- Database path
- Admin users
- Secrets

## 🐛 Debugging

### View Service Logs

```bash
# All services
make logs

# Specific service
docker-compose -f docker-compose.test.yml logs -f prism-api
```

### Interactive Shell

```bash
# Open shell in test container
make shell

# Manually run tests
go test -v -run TestAccessMatrix
```

### Common Issues

**Services not starting:**
```bash
make clean
make build
make test
```

**Database issues:**
```bash
rm -rf test-data/*.db
make seed
```

## 📦 CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Integration Tests
        run: |
          cd tests/integration
          make test
      - name: Upload Reports
        uses: actions/upload-artifact@v3
        with:
          name: test-reports
          path: tests/integration/reports/
```

## 🎓 Adding New Tests

### 1. Add Access Control Test

Edit `fixtures/access_matrix.csv`:
```csv
9,testuser@test.com,true,Reason for access,test_category
```

### 2. Add New Vulnerability

Edit `fixtures/vulnerabilities.csv`:
```csv
9,New Vuln,Description,high,7.5,1,published,bob@test.com,,Reported,Impact,High,Remediation
```

### 3. Add New Benchmark

Edit `benchmark_test.go`:
```go
func BenchmarkNewScenario(b *testing.B) {
    token := setupBenchmark(b)
    // ... your benchmark code
}
```

## 📊 Performance Baselines

Expected performance on modern hardware:

- **List Vulnerabilities**: ~200-300 ops/sec
- **Get Single Vulnerability**: ~500-800 ops/sec
- **Access Control Check**: ~400-600 ops/sec
- **Concurrent Reads**: Scales with CPU cores

SQLite optimizations:
- WAL mode enabled
- Read-heavy workload optimized
- Proper indexes on foreign keys

## 🔒 Security Notes

- Mock OIDC server is for **testing only** - never use in production
- Test secrets are hardcoded - use environment variables in production
- All test users have weak passwords intentionally

## 📄 License

Same as main Prism project.

## 🤝 Contributing

To contribute test cases:

1. Add fixtures to CSV files
2. Run `make validate-fixtures`
3. Run `make test` to verify
4. Submit PR with updated CSV files

## 📞 Support

For issues or questions:
- Check `make help` for commands
- View logs with `make logs`
- Open an issue in the main repository

---

**Happy Testing! 🎉**
