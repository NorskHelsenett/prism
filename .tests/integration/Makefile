.PHONY: help test test-access bench report clean seed build up down logs

# Container runtime (use podman-compose or docker-compose)
COMPOSE ?= podman-compose

# Default target
help:
	@echo "Prism Integration Test Suite"
	@echo "============================="
	@echo ""
	@echo "Available targets:"
	@echo "  make test          - Run full integration test suite"
	@echo "  make test-access   - Run only access control tests"
	@echo "  make bench         - Run benchmark tests"
	@echo "  make report        - Generate and display test reports"
	@echo "  make seed          - Seed test database with fixtures"
	@echo "  make build         - Build all Docker containers"
	@echo "  make up            - Start all services"
	@echo "  make down          - Stop all services"
	@echo "  make clean         - Clean up test data and reports"
	@echo "  make logs          - Show logs from all services"
	@echo ""

# Build all containers
build:
	@echo "Building containers with $(COMPOSE)..."
	$(COMPOSE) -f docker-compose.test.yml build

# Start services in background
up:
	@echo "Starting services..."
	$(COMPOSE) -f docker-compose.test.yml up -d oidc-mock prism-api
	@echo "Waiting for services to be healthy..."
	@sleep 10
	@echo "Services are ready!"

# Stop all services
down:
	@echo "Stopping services..."
	$(COMPOSE) -f docker-compose.test.yml down

# View logs
logs:
	$(COMPOSE) -f docker-compose.test.yml logs -f

# Run full test suite
test: clean build
	@echo "Starting services..."
	$(COMPOSE) -f docker-compose.test.yml up -d oidc-mock prism-api
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Seeding database with test fixtures..."
	$(COMPOSE) -f docker-compose.test.yml run --rm test-runner go run cmd/seed/main.go
	@echo "Running full integration test suite..."
	$(COMPOSE) -f docker-compose.test.yml up --abort-on-container-exit test-runner
	@echo ""
	@echo "Test completed! Check reports/ for detailed results."
	@make report

# Run only access control tests
test-access: clean build
	@echo "Running access control tests..."
	$(COMPOSE) -f docker-compose.test.yml run --rm test-runner \
		go test -v -run TestAccessMatrix -timeout 15m
	@echo ""
	@echo "Access tests completed! Check reports/ for detailed results."
	@make report

# Run benchmark tests
bench: clean build
	@echo "Running benchmark tests..."
	$(COMPOSE) -f docker-compose.test.yml --profile benchmark up --abort-on-container-exit benchmark-runner
	@echo ""
	@echo "Benchmarks completed! Check reports/ for detailed results."
	@make report

# Generate and display reports
report:
	@echo ""
	@echo "=========================================="
	@echo "Test Reports"
	@echo "=========================================="
	@if [ -f reports/access_test_latest.json ]; then \
		echo ""; \
		echo "Access Control Test Summary:"; \
		echo "----------------------------"; \
		cat reports/access_test_latest.json | jq -r '"Total Tests: \(.total_tests)\nPassed: \(.passed)\nFailed: \(.failed)\nPass Rate: \(.pass_rate)%"'; \
		echo ""; \
		if [ "$$(cat reports/access_test_latest.json | jq -r '.failed')" != "0" ]; then \
			echo "Failed Tests:"; \
			cat reports/access_test_latest.json | jq -r '.failed_tests[] | "  ✗ vuln=\(.vulnerability_id) user=\(.user_email) expected=\(.expected) got=\(.actual) (\(.reason))"'; \
		fi; \
		echo ""; \
		echo "Category Breakdown:"; \
		cat reports/access_test_latest.json | jq -r '.category_stats | to_entries[] | "  \(.key): \(.value.passed)/\(.value.total) (\(.value.pass_rate | floor)%)"'; \
		echo ""; \
	fi
	@if [ -f reports/benchmark_latest.json ]; then \
		echo ""; \
		echo "Benchmark Results:"; \
		echo "------------------"; \
		cat reports/benchmark_latest.json | jq -r '.scenarios[] | "  \(.name):\n    Ops: \(.operations)\n    Mean: \(.mean_latency_ns / 1000000)ms\n    P95: \(.p95_latency_ns / 1000000)ms\n    Throughput: \(.throughput_ops_per_sec | floor) ops/sec\n"'; \
		echo ""; \
		echo "Summary:"; \
		cat reports/benchmark_latest.json | jq -r '.summary | "  Total Scenarios: \(.total_scenarios)\n  Total Operations: \(.total_operations)\n  Avg Throughput: \(.avg_throughput_ops_per_sec | floor) ops/sec"'; \
		echo ""; \
	fi
	@echo "Full JSON reports available in reports/"
	@echo "=========================================="

# Seed database with test fixtures
seed: up
	@echo "Cleaning database before seeding..."
	rm -rf test-data/prism.db

	@echo "Seeding test database..."
	$(COMPOSE) -f docker-compose.test.yml run --rm test-runner \
		go run ./cmd/seed
	@echo "Database seeded successfully!"

# Clean up test data and reports
clean:
	@echo "Cleaning up test data..."
# 	rm -rf test-data/*.db
	sqlite3 test-data/prism.db "VACUUM;drop table if exists project_data; drop table if exists json_data; drop table if exists api_keys;"
	rm -rf reports/*.json
	$(COMPOSE) -f docker-compose.test.yml down -v
	@echo "Cleanup completed!"

# Quick test (doesn't rebuild containers)
quick-test:
	@echo "Running quick test (no rebuild)..."
	$(COMPOSE) -f docker-compose.test.yml up --abort-on-container-exit test-runner

# Interactive shell in test container
shell:
	$(COMPOSE) -f docker-compose.test.yml run --rm test-runner /bin/sh

# Validate fixtures (check CSV format)
validate-fixtures:
	@echo "Validating CSV fixtures..."
	@python3 -c "import csv; [list(csv.DictReader(open(f))) for f in ['fixtures/users.csv', 'fixtures/projects.csv', 'fixtures/vulnerabilities.csv', 'fixtures/access_matrix.csv']]" && echo "✓ All fixtures are valid CSV files" || echo "✗ Fixture validation failed"
