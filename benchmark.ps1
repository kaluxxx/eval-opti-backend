# Script de benchmark pour comparer V1 vs V2 avec hyperfine

Write-Host "=== Benchmark API V1 vs V2 avec Hyperfine ===" -ForegroundColor Cyan
Write-Host ""

# Demarre le serveur en arriere-plan
Write-Host "Demarrage du serveur..." -ForegroundColor Yellow
$server = Start-Process -FilePath "go" -ArgumentList "run", "main.go" -WorkingDirectory $PSScriptRoot -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 3

Write-Host "Serveur demarre (PID: $($server.Id))" -ForegroundColor Green
Write-Host ""

try {
    # Teste que le serveur repond
    Write-Host "Test de connexion au serveur..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/api/health" -TimeoutSec 5 -UseBasicParsing
        Write-Host "[OK] Serveur operationnel" -ForegroundColor Green
    } catch {
        Write-Host "[ERROR] Le serveur ne repond pas" -ForegroundColor Red
        exit 1
    }
    Write-Host ""

    # Benchmark 1: Stats JSON avec 365 jours
    Write-Host "=== Benchmark 1: Stats JSON (365 days) ===" -ForegroundColor Cyan
    & 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
        --warmup 2 `
        --runs 10 `
        --export-markdown benchmark_stats_365.md `
        --export-json benchmark_stats_365.json `
        "Invoke-WebRequest -Uri 'http://localhost:8080/api/v1/stats?days=365' -UseBasicParsing | Out-Null" `
        "Invoke-WebRequest -Uri 'http://localhost:8080/api/v2/stats?days=365' -UseBasicParsing | Out-Null"
    Write-Host ""

    # Benchmark 2: Stats JSON avec 100 jours
    Write-Host "=== Benchmark 2: Stats JSON (100 days) ===" -ForegroundColor Cyan
    & 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
        --warmup 2 `
        --runs 10 `
        --export-markdown benchmark_stats_100.md `
        "Invoke-WebRequest -Uri 'http://localhost:8080/api/v1/stats?days=100' -UseBasicParsing | Out-Null" `
        "Invoke-WebRequest -Uri 'http://localhost:8080/api/v2/stats?days=100' -UseBasicParsing | Out-Null"
    Write-Host ""

    # Benchmark 3: Export CSV (petit dataset)
    Write-Host "=== Benchmark 3: Export CSV (30 days) ===" -ForegroundColor Cyan
    & 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
        --warmup 1 `
        --runs 5 `
        --export-markdown benchmark_csv_30.md `
        "Invoke-WebRequest -Uri 'http://localhost:8080/api/v1/export/csv?days=30' -UseBasicParsing | Out-Null" `
        "Invoke-WebRequest -Uri 'http://localhost:8080/api/v2/export/csv?days=30' -UseBasicParsing | Out-Null"
    Write-Host ""

    # Benchmark 4: Test du cache V2
    Write-Host "=== Benchmark 4: Cache effect V2 ===" -ForegroundColor Cyan
    Write-Host "Warming up cache..." -ForegroundColor Yellow
    Invoke-WebRequest -Uri "http://localhost:8080/api/v2/stats?days=365" -UseBasicParsing | Out-Null
    Start-Sleep -Seconds 1

    & 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
        --warmup 0 `
        --runs 50 `
        --export-markdown benchmark_cache.md `
        "Invoke-WebRequest -Uri 'http://localhost:8080/api/v2/stats?days=365' -UseBasicParsing | Out-Null"
    Write-Host ""

    Write-Host "=== Results ===" -ForegroundColor Cyan
    Write-Host "Results exported to:" -ForegroundColor Green
    Write-Host "  - benchmark_stats_365.md (Stats 365 days)" -ForegroundColor White
    Write-Host "  - benchmark_stats_100.md (Stats 100 days)" -ForegroundColor White
    Write-Host "  - benchmark_csv_30.md (Export CSV 30 days)" -ForegroundColor White
    Write-Host "  - benchmark_cache.md (V2 cache performance)" -ForegroundColor White
    Write-Host ""

} finally {
    # Stop server
    Write-Host "Stopping server..." -ForegroundColor Yellow
    Stop-Process -Id $server.Id -Force
    Write-Host "[OK] Server stopped" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== Benchmark complete ===" -ForegroundColor Cyan
