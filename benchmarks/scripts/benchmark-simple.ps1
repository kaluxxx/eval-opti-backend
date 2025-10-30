# ============================================================================
# SCRIPT DE BENCHMARK - PowerShell (.ps1)
# Mesure les performances de l'API V1 (non-optimisée) vs V2 (optimisée)
# ============================================================================

# SYNTAXE PowerShell: Write-Host affiche du texte coloré dans la console
Write-Host "=== Benchmark API V1 vs V2 ===" -ForegroundColor Cyan
Write-Host ""

# =============================================================================
# TEST DE CONNEXION AU SERVEUR
# =============================================================================
Write-Host "Testing server connection..." -ForegroundColor Yellow

try {
    # OPTIONS:
    # -s (silent) = pas de barre de progression (économise CPU)
    # -o nul = jette la réponse (pas d'I/O disque)
    #
    # SYNTAXE PowerShell: $null = ignore le retour (pas d'affichage)
    # MÉMOIRE: Réponse HTTP (quelques KB) allouée puis libérée immédiatement
    $null = curl.exe -s http://localhost:8080/api/health
    
    # Si curl.exe réussit (exit code 0), affiche succès
    Write-Host "[OK] Server is running" -ForegroundColor Green
    
} catch {
    Write-Host "[ERROR] Server not running. Start it with: go run main.go" -ForegroundColor Red
    exit 1
}
Write-Host ""

$resultsDir = "..\results"

# SYNTAXE: Test-Path = vérifie existence fichier/dossier
if (-not (Test-Path $resultsDir)) {
    # New-Item = crée dossier (mkdir en Unix)
    # SYSTÈME: Appel syscall CreateDirectoryW ~50µs
    # | Out-Null = ignore output (évite affichage console)
    New-Item -ItemType Directory -Path $resultsDir | Out-Null
}

# =============================================================================
# BENCHMARK 1: STATS 365 JOURS - PROBLÈME N+1
# =============================================================================

Write-Host "=== Benchmark 1: Stats JSON (365 days) - PostgreSQL ===" -ForegroundColor Cyan

# CONTEXTE MÉTIER:
# V1: Fait 200+ queries SQL (1 principale + N queries par produit)
# V2: Fait 5 queries SQL parallèles avec JOINs optimisés
Write-Host "V1: N+1 problem (200+ queries) | V2: JOINs (5 queries)" -ForegroundColor Gray

# OUTIL: Hyperfine = CLI benchmark tool écrit en Rust
# POURQUOI Hyperfine?
# - Statistiques fiables (médiane, écart-type, outliers)
# - Warmup runs (évite cold cache/CPU)
# - Export Markdown/JSON
# - Plus précis que mesures manuelles
#
# INSTALLATION: scoop install hyperfine (Windows)
# MÉMOIRE: Hyperfine ~5MB RAM (Rust = binaire optimisé)
#
# SYNTAXE PowerShell: & 'path' = exécute un programme externe
# Backticks (`) = continuation de ligne (comme \ en Bash)
& 'C:\Users\Lucas\scoop\shims\hyperfine.exe' `
    # --warmup 2 : Exécute 2 fois AVANT de mesurer
    # POURQUOI? Élimine les effets de cold start:
    # - CPU cache (L1/L2/L3) vide
    # - OS page cache vide
    # - PostgreSQL buffer cache vide
    # - Go runtime pas encore optimisé (JIT inlining)
    --warmup 2 `
    
    # --runs 10 : Mesure 10 fois pour statistiques
    # POURQUOI 10? Balance précision vs temps total
    --runs 10 `
    
    # --export-markdown : Exporte résultats en tableau Markdown
    --export-markdown "$resultsDir\benchmark_stats_365.md" `
    
    # --command-name : Nom affiché dans les résultats
    --command-name "V1 (N+1 + Bubble Sort)" `
    --command-name "V2 (JOINs + SQL ORDER BY)" `

    "curl -s http://localhost:8080/api/v1/stats?days=365 -o nul" `
    "curl -s http://localhost:8080/api/v2/stats?days=365 -o nul"

Write-Host ""

# =============================================================================
# BENCHMARK 2: STATS 100 JOURS - Dataset plus petit
# =============================================================================

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

# =============================================================================
# BENCHMARK 5: EFFET DU CACHE (V2 uniquement)
# =============================================================================

Write-Host "=== Benchmark 5: V2 Cache Effect (365 days) ===" -ForegroundColor Cyan

# OBJECTIF: Mesurer speedup du cache
# V2 utilise un cache shardé (16 shards) avec TTL 5 minutes

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

# Liste des fichiers générés (Markdown tables)
Write-Host "  - $resultsDir\benchmark_stats_365.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_stats_100.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_csv_30.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_parquet_7.md" -ForegroundColor White
Write-Host "  - $resultsDir\benchmark_cache.md" -ForegroundColor White
Write-Host ""

Write-Host "=== Benchmark complete ===" -ForegroundColor Cyan
Write-Host ""

# RÉSUMÉ DES OPTIMISATIONS TESTÉES
Write-Host "Summary of optimizations tested:" -ForegroundColor Cyan

# 1. N+1 → JOINs: 200 queries → 5 queries
Write-Host "1. N+1 Problem -> JOINs SQL (Stats 365)" -ForegroundColor Green

# 2. Bubble Sort O(n²) → SQL ORDER BY O(n log n)
Write-Host "2. Bubble Sort O(n²) -> SQL ORDER BY (Stats)" -ForegroundColor Green

# 3. Boucles imbriquées → Agrégations SQL
Write-Host "3. Multiple loops -> SQL Aggregations (Stats)" -ForegroundColor Green

# 4. N+1 par ligne CSV → 1 query globale
Write-Host "4. N+1 per row -> Single JOIN (CSV Export)" -ForegroundColor Green

# 5. Tout en RAM → Streaming par batches
Write-Host "5. Full memory load -> Streaming (Parquet)" -ForegroundColor Green

# 6. Pas de cache → Cache 5min TTL avec sharding
Write-Host "6. No cache -> 5min TTL cache (V2 Cache)" -ForegroundColor Green
