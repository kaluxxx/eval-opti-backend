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

# Benchmark 1: Stats 365 days
Write-Host "=== Benchmark 1: Stats JSON (365 days) ===" -ForegroundColor Cyan
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 2 `
    --runs 10 `
    --export-markdown "$resultsDir\benchmark_stats_365.md" `
    "curl -s http://localhost:8080/api/v1/stats?days=365 -o nul" `
    "curl -s http://localhost:8080/api/v2/stats?days=365 -o nul"
Write-Host ""

# Benchmark 2: Stats 100 days
Write-Host "=== Benchmark 2: Stats JSON (100 days) ===" -ForegroundColor Cyan
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 2 `
    --runs 10 `
    --export-markdown "$resultsDir\benchmark_stats_100.md" `
    "curl -s http://localhost:8080/api/v1/stats?days=100 -o nul" `
    "curl -s http://localhost:8080/api/v2/stats?days=100 -o nul"
Write-Host ""

# Benchmark 3: CSV Export 30 days
Write-Host "=== Benchmark 3: Export CSV (30 days) ===" -ForegroundColor Cyan
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 1 `
    --runs 5 `
    --export-markdown "$resultsDir\benchmark_csv_30.md" `
    "curl -s http://localhost:8080/api/v1/export/csv?days=30 -o nul" `
    "curl -s http://localhost:8080/api/v2/export/csv?days=30 -o nul"
Write-Host ""

# Benchmark 4: Cache effect
Write-Host "=== Benchmark 4: V2 Cache Effect ===" -ForegroundColor Cyan
Write-Host "Warming up cache..." -ForegroundColor Yellow
curl.exe -s http://localhost:8080/api/v2/stats?days=365 -o nul
Start-Sleep -Seconds 1

& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    --warmup 0 `
    --runs 50 `
    --export-markdown "$resultsDir\benchmark_cache.md" `
    "curl -s http://localhost:8080/api/v2/stats?days=365 -o nul"
Write-Host ""

Write-Host "=== Results ===" -ForegroundColor Cyan
Write-Host "Results exported to:" -ForegroundColor Green
Write-Host "  - $resultsDir\benchmark_stats_365.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_stats_100.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_csv_30.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_cache.md" -ForegroundColor White
Write-Host ""
Write-Host "=== Benchmark complete ===" -ForegroundColor Cyan
