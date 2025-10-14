#!/bin/bash
# Generate HTML report from JSON test results

REPORTS_DIR="${1:-./reports}"

if [ ! -f "$REPORTS_DIR/access_test_latest.json" ] && [ ! -f "$REPORTS_DIR/benchmark_latest.json" ]; then
    echo "No reports found in $REPORTS_DIR"
    exit 1
fi

cat > "$REPORTS_DIR/report.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Prism Integration Test Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
        }
        .card {
            background: white;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric {
            display: inline-block;
            margin: 10px 20px 10px 0;
        }
        .metric-value {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        .metric-label {
            font-size: 0.9em;
            color: #666;
        }
        .pass { color: #10b981; }
        .fail { color: #ef4444; }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #e5e7eb;
        }
        th {
            background: #f9fafb;
            font-weight: 600;
        }
        .progress-bar {
            height: 20px;
            background: #e5e7eb;
            border-radius: 10px;
            overflow: hidden;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #10b981 0%, #059669 100%);
            transition: width 0.3s;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Prism Integration Test Report</h1>
        <p id="timestamp"></p>
    </div>
    
    <div id="access-report"></div>
    <div id="benchmark-report"></div>

    <script>
        // Load and display reports
        async function loadReports() {
            try {
                const accessResponse = await fetch('./access_test_latest.json');
                if (accessResponse.ok) {
                    const accessData = await accessResponse.json();
                    displayAccessReport(accessData);
                }
            } catch (e) {
                console.log('No access report found');
            }

            try {
                const benchResponse = await fetch('./benchmark_latest.json');
                if (benchResponse.ok) {
                    const benchData = await benchResponse.json();
                    displayBenchmarkReport(benchData);
                }
            } catch (e) {
                console.log('No benchmark report found');
            }
        }

        function displayAccessReport(data) {
            const passRate = data.pass_rate.toFixed(1);
            const html = `
                <div class="card">
                    <h2>Access Control Tests</h2>
                    <div class="metric">
                        <div class="metric-value ${data.failed === 0 ? 'pass' : 'fail'}">${passRate}%</div>
                        <div class="metric-label">Pass Rate</div>
                    </div>
                    <div class="metric">
                        <div class="metric-value">${data.passed}</div>
                        <div class="metric-label">Passed</div>
                    </div>
                    <div class="metric">
                        <div class="metric-value">${data.failed}</div>
                        <div class="metric-label">Failed</div>
                    </div>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${passRate}%"></div>
                    </div>
                </div>

                <div class="card">
                    <h3>Category Breakdown</h3>
                    <table>
                        <thead>
                            <tr>
                                <th>Category</th>
                                <th>Total</th>
                                <th>Passed</th>
                                <th>Failed</th>
                                <th>Pass Rate</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${Object.entries(data.category_stats).map(([cat, stats]) => `
                                <tr>
                                    <td>${cat}</td>
                                    <td>${stats.total}</td>
                                    <td class="pass">${stats.passed}</td>
                                    <td class="fail">${stats.failed}</td>
                                    <td>${stats.pass_rate.toFixed(1)}%</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            `;
            document.getElementById('access-report').innerHTML = html;
            document.getElementById('timestamp').textContent = `Generated: ${data.timestamp}`;
        }

        function displayBenchmarkReport(data) {
            const html = `
                <div class="card">
                    <h2>Benchmark Results</h2>
                    <table>
                        <thead>
                            <tr>
                                <th>Scenario</th>
                                <th>Operations</th>
                                <th>Mean (ms)</th>
                                <th>P95 (ms)</th>
                                <th>Throughput (ops/s)</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${data.scenarios.map(s => `
                                <tr>
                                    <td>${s.name}</td>
                                    <td>${s.operations}</td>
                                    <td>${(s.mean_latency_ns / 1000000).toFixed(2)}</td>
                                    <td>${(s.p95_latency_ns / 1000000).toFixed(2)}</td>
                                    <td>${s.throughput_ops_per_sec.toFixed(1)}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            `;
            document.getElementById('benchmark-report').innerHTML = html;
        }

        loadReports();
    </script>
</body>
</html>
EOF

echo "✓ HTML report generated: $REPORTS_DIR/report.html"
