package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkReport contains benchmark results
type BenchmarkReport struct {
	Timestamp time.Time        `json:"timestamp"`
	Scenarios []ScenarioResult `json:"scenarios"`
	Summary   BenchmarkSummary `json:"summary"`
}

// ScenarioResult contains results for a single benchmark scenario
type ScenarioResult struct {
	Name        string        `json:"name"`
	Operations  int           `json:"operations"`
	Duration    time.Duration `json:"duration_ns"`
	MeanLatency time.Duration `json:"mean_latency_ns"`
	MinLatency  time.Duration `json:"min_latency_ns"`
	MaxLatency  time.Duration `json:"max_latency_ns"`
	P50Latency  time.Duration `json:"p50_latency_ns"`
	P95Latency  time.Duration `json:"p95_latency_ns"`
	P99Latency  time.Duration `json:"p99_latency_ns"`
	Throughput  float64       `json:"throughput_ops_per_sec"`
	BytesPerOp  int64         `json:"bytes_per_op"`
	AllocsPerOp int64         `json:"allocs_per_op"`
}

// BenchmarkSummary provides overall statistics
type BenchmarkSummary struct {
	TotalScenarios  int     `json:"total_scenarios"`
	TotalOperations int     `json:"total_operations"`
	TotalDuration   string  `json:"total_duration"`
	AvgThroughput   float64 `json:"avg_throughput_ops_per_sec"`
}

// BenchmarkListVulnerabilities tests listing all vulnerabilities
func BenchmarkListVulnerabilities(b *testing.B) {
	client := setupBenchmark(b)
	url := fmt.Sprintf("%s/api/vulnerability", baseURL)

	latencies := make([]time.Duration, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
	_, _, err := makeAuthenticatedRequestWithClient("GET", url, client)
		latencies[i] = time.Since(start)

		if err != nil {
			b.Errorf("Request failed: %v", err)
		}
	}
	b.StopTimer()

	reportBenchmarkStats(b, "List All Vulnerabilities", latencies)
}

// BenchmarkGetSingleVulnerability tests getting a single vulnerability
func BenchmarkGetSingleVulnerability(b *testing.B) {
	client := setupBenchmark(b)
	url := fmt.Sprintf("%s/api/vulnerability/1", baseURL)

	latencies := make([]time.Duration, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
	_, _, err := makeAuthenticatedRequestWithClient("GET", url, client)
		latencies[i] = time.Since(start)

		if err != nil {
			b.Errorf("Request failed: %v", err)
		}
	}
	b.StopTimer()

	reportBenchmarkStats(b, "Get Single Vulnerability", latencies)
}

// BenchmarkFilterPublicVulnerabilities tests filtering for published vulnerabilities
func BenchmarkFilterPublicVulnerabilities(b *testing.B) {
	client := setupBenchmark(b)
	url := fmt.Sprintf("%s/api/vulnerability?visibility=published", baseURL)

	latencies := make([]time.Duration, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
	_, _, err := makeAuthenticatedRequestWithClient("GET", url, client)
		latencies[i] = time.Since(start)

		if err != nil {
			b.Errorf("Request failed: %v", err)
		}
	}
	b.StopTimer()

	reportBenchmarkStats(b, "Filter Published Vulnerabilities", latencies)
}

// BenchmarkConcurrentReads tests concurrent read operations (SQLite optimization test)
func BenchmarkConcurrentReads(b *testing.B) {
	client := setupBenchmark(b)

	b.RunParallel(func(pb *testing.PB) {
		vulnID := 1
		for pb.Next() {
			url := fmt.Sprintf("%s/api/vulnerability/%d", baseURL, vulnID)
			_, _, err := makeAuthenticatedRequestWithClient("GET", url, client)
			if err != nil {
				b.Errorf("Request failed: %v", err)
			}
			// Rotate through different vulnerabilities
			vulnID = (vulnID % 8) + 1
		}
	})
}

// BenchmarkProjectList tests listing projects
func BenchmarkProjectList(b *testing.B) {
	client := setupBenchmark(b)
	url := fmt.Sprintf("%s/api/project", baseURL)

	latencies := make([]time.Duration, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
	_, _, err := makeAuthenticatedRequestWithClient("GET", url, client)
		latencies[i] = time.Since(start)

		if err != nil {
			b.Errorf("Request failed: %v", err)
		}
	}
	b.StopTimer()

	reportBenchmarkStats(b, "List All Projects", latencies)
}

// BenchmarkAccessControlCheck tests the overhead of access control checks
func BenchmarkAccessControlCheck(b *testing.B) {
	// Test with a restricted user (non-admin)
	client := setupBenchmark(b)

	latencies := make([]time.Duration, b.N)

	// Test access to various vulnerabilities with different permissions
	urls := []string{
		fmt.Sprintf("%s/api/vulnerability/1", baseURL), // Should have access
		fmt.Sprintf("%s/api/vulnerability/2", baseURL), // Should have access (assigned)
		fmt.Sprintf("%s/api/vulnerability/3", baseURL), // Should NOT have access
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := urls[i%len(urls)]
		start := time.Now()
		_, _, _ = makeAuthenticatedRequestWithClient("GET", url, client)
		latencies[i] = time.Since(start)
	}
	b.StopTimer()

	reportBenchmarkStats(b, "Access Control Check Overhead", latencies)
}

// BenchmarkDashboardLoad simulates loading dashboard with multiple queries
func BenchmarkDashboardLoad(b *testing.B) {
	client := setupBenchmark(b)

	// Simulate multiple API calls that a dashboard might make
	endpoints := []string{
		"/api/vulnerability",
		"/api/project",
		"/api/notification",
	}

	latencies := make([]time.Duration, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		for _, endpoint := range endpoints {
			url := fmt.Sprintf("%s%s", baseURL, endpoint)
			_, _, err := makeAuthenticatedRequestWithClient("GET", url, client)
			if err != nil {
				b.Logf("Request to %s failed: %v", endpoint, err)
			}
		}
		latencies[i] = time.Since(start)
	}
	b.StopTimer()

	reportBenchmarkStats(b, "Dashboard Load (Multiple Queries)", latencies)
}

// Helper functions

func setupBenchmark(b *testing.B) *http.Client {
	b.Helper()
	// Try cookie-based flow first
	client, err := authenticateUserCookieClient("alice@test.com", "password123")
	if err == nil && client != nil {
		return client
	}

	// Fallback: obtain a bearer token and return a client that injects Authorization header
	token, terr := authenticateUser("alice@test.com", "password123")
	if terr != nil {
		b.Fatalf("Failed to authenticate (cookie error: %v, token error: %v)", err, terr)
	}
	return newTokenClient(token)
}

// makeAuthenticatedRequestWithClient performs a request using a cookie-aware client
func makeAuthenticatedRequestWithClient(method, url string, client *http.Client) (int, []byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use provided client which should carry session cookie
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	return resp.StatusCode, body, nil
}

// authTransport injects an Authorization header for each request
type authTransport struct {
	token string
	rt    http.RoundTripper
}

func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
	return a.rt.RoundTrip(req)
}

// newTokenClient returns an http.Client that injects the given bearer token into requests
func newTokenClient(token string) *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &authTransport{
			token: token,
			rt:    http.DefaultTransport,
		},
	}
}

func reportBenchmarkStats(b *testing.B, name string, latencies []time.Duration) {
	if len(latencies) == 0 {
		return
	}

	// Calculate statistics
	var total time.Duration
	min := latencies[0]
	max := latencies[0]

	for _, lat := range latencies {
		total += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	mean := total / time.Duration(len(latencies))

	// Calculate percentiles (simple approximation)
	p50 := latencies[len(latencies)/2]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]

	throughput := float64(len(latencies)) / total.Seconds()

	result := ScenarioResult{
		Name:        name,
		Operations:  len(latencies),
		Duration:    total,
		MeanLatency: mean,
		MinLatency:  min,
		MaxLatency:  max,
		P50Latency:  p50,
		P95Latency:  p95,
		P99Latency:  p99,
		Throughput:  throughput,
	}

	// Save to benchmark report
	saveBenchmarkResult(result)

	// Log summary
	b.Logf("\n=== %s ===", name)
	b.Logf("Operations: %d", result.Operations)
	b.Logf("Mean Latency: %v", result.MeanLatency)
	b.Logf("P50: %v, P95: %v, P99: %v", result.P50Latency, result.P95Latency, result.P99Latency)
	b.Logf("Min: %v, Max: %v", result.MinLatency, result.MaxLatency)
	b.Logf("Throughput: %.2f ops/sec", result.Throughput)
}

func saveBenchmarkResult(result ScenarioResult) {
	reportsDir := "./reports"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return
	}

	reportFile := filepath.Join(reportsDir, "benchmark_latest.json")

	// Load existing report or create new
	var report BenchmarkReport
	if data, err := os.ReadFile(reportFile); err == nil {
		_ = json.Unmarshal(data, &report)
	}

	// Update report
	if report.Timestamp.IsZero() {
		report.Timestamp = time.Now()
	}
	report.Scenarios = append(report.Scenarios, result)

	// Calculate summary
	report.Summary.TotalScenarios = len(report.Scenarios)
	var totalOps int
	var totalThroughput float64
	for _, s := range report.Scenarios {
		totalOps += s.Operations
		totalThroughput += s.Throughput
	}
	report.Summary.TotalOperations = totalOps
	if len(report.Scenarios) > 0 {
		report.Summary.AvgThroughput = totalThroughput / float64(len(report.Scenarios))
	}

	// Save report
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(reportFile, data, 0644)

	// Also save timestamped version
	timestampedFile := filepath.Join(reportsDir, fmt.Sprintf("benchmark_%s.json",
		time.Now().Format("20060102_150405")))
	_ = os.WriteFile(timestampedFile, data, 0644)
}
