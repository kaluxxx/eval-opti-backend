# Simple benchmark script - Start server manually first with: go run main.go

Write-Host "=== Benchmark API V1 vs V2 ===" -ForegroundColor Cyan
Write-Host ""

# Test connection
Write-Host "Testing server connection..." -ForegroundColor Yellow
try {
    $null = curl.exe -s http://localhost:8080/api/health
    Write-Host "[OK] Server is running" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Server not running. Start it with: go run main.go" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Create results directory if needed
$resultsDir = "..\results"
if (-not (Test-Path $resultsDir)) {
    New-Item -ItemType Directory -Path $resultsDir | Out-Null
}

# Benchmark 1: Stats 365 days (PostgreSQL with N+1 vs JOINs)
Write-Host "=== Benchmark 1: Stats JSON (365 days) - PostgreSQL ===" -ForegroundColor Cyan
Write-Host "V1: N+1 problem (200+ queries) | V2: JOINs (5 queries)" -ForegroundColor Gray
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 2 `
    --runs 10 `
    --export-markdown "$resultsDir\benchmark_stats_365.md" `
    --command-name "V1 (N+1 + Bubble Sort)" `
    --command-name "V2 (JOINs + SQL ORDER BY)" `
    "curl -s http://localhost:8080/api/v1/stats?days=365 -o nul" `
    "curl -s http://localhost:8080/api/v2/stats?days=365 -o nul"
Write-Host ""

# Benchmark 2: Stats 100 days
Write-Host "=== Benchmark 2: Stats JSON (100 days) ===" -ForegroundColor Cyan
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 2 `
    --runs 10 `
    --export-markdown "$resultsDir\benchmark_stats_100.md" `
    --command-name "V1 (Non optimisé)" `
    --command-name "V2 (Optimisé)" `
    "curl -s http://localhost:8080/api/v1/stats?days=100 -o nul" `
    "curl -s http://localhost:8080/api/v2/stats?days=100 -o nul"
Write-Host ""

# Benchmark 3: CSV Export 30 days
Write-Host "=== Benchmark 3: Export CSV (30 days) ===" -ForegroundColor Cyan
Write-Host "V1: N+1 for each row | V2: Single JOIN query" -ForegroundColor Gray
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 1 `
    --runs 5 `
    --export-markdown "$resultsDir\benchmark_csv_30.md" `
    --command-name "V1 CSV (N+1)" `
    --command-name "V2 CSV (JOINs)" `
    "curl -s http://localhost:8080/api/v1/export/csv?days=30 -o nul" `
    "curl -s http://localhost:8080/api/v2/export/csv?days=30 -o nul"
Write-Host ""

# Benchmark 4: Parquet Export (memory vs streaming)
Write-Host "=== Benchmark 4: Export Parquet (7 days) ===" -ForegroundColor Cyan
Write-Host "WARNING: V1 loads ALL data in memory!" -ForegroundColor Yellow
Write-Host "V1: Full memory load | V2: Streaming batches" -ForegroundColor Gray
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 1 `
    --runs 3 `
    --export-markdown "$resultsDir\benchmark_parquet_7.md" `
    --command-name "V1 Parquet (Memory)" `
    --command-name "V2 Parquet (Streaming)" `
    "curl -s http://localhost:8080/api/v1/export/parquet?days=7 -o nul" `
    "curl -s http://localhost:8080/api/v2/export/parquet?days=7 -o nul"
Write-Host ""

# Benchmark 5: Cache effect (V2 only)
Write-Host "=== Benchmark 5: V2 Cache Effect (365 days) ===" -ForegroundColor Cyan
Write-Host "Warming up cache..." -ForegroundColor Yellow
curl.exe -s http://localhost:8080/api/v2/stats?days=365 -o nul
Start-Sleep -Seconds 1

& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 0 `
    --runs 50 `
    --export-markdown "$resultsDir\benchmark_cache.md" `
    --command-name "V2 with Cache (5min TTL)" `
    "curl -s http://localhost:8080/api/v2/stats?days=365 -o nul"
Write-Host ""

Write-Host "=== Results ===" -ForegroundColor Cyan
Write-Host "Results exported to:" -ForegroundColor Green
Write-Host "  - $resultsDir\benchmark_stats_365.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_stats_100.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_csv_30.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_parquet_7.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_cache.md" -ForegroundColor White
Write-Host ""
Write-Host "=== Benchmark complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary of optimizations tested:" -ForegroundColor Cyan
Write-Host "1. N+1 Problem -> JOINs SQL (Stats 365)" -ForegroundColor Green
Write-Host "2. Bubble Sort O(n²) -> SQL ORDER BY (Stats)" -ForegroundColor Green
Write-Host "3. Multiple loops -> SQL Aggregations (Stats)" -ForegroundColor Green
Write-Host "4. N+1 per row -> Single JOIN (CSV Export)" -ForegroundColor Green
Write-Host "5. Full memory load -> Streaming (Parquet)" -ForegroundColor Green
Write-Host "6. No cache -> 5min TTL cache (V2 Cache)" -ForegroundColor Green
