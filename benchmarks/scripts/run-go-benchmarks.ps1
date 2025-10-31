# Go Benchmarks Runner Script
# Exécute les benchmarks Go avec différentes options

param(
    [string]$Package = "all",           # all, export, stats, infrastructure, cache, workerpool
    [switch]$Integration,               # Exécuter uniquement les benchmarks d'intégration (avec DB)
    [switch]$Unit,                      # Exécuter uniquement les benchmarks unitaires (sans DB)
    [switch]$Short,                     # Skip les benchmarks lents (V1 avec 365 jours)
    [switch]$Compare,                   # Générer un rapport de comparaison avec benchstat
    [string]$Profile = "",              # cpu, mem, ou vide
    [int]$Count = 1,                    # Nombre de runs (-count)
    [string]$BenchTime = "1s",          # Durée ou nombre d'itérations (-benchtime)
    [switch]$Save,                      # Sauvegarder les résultats avec timestamp
    [switch]$Help                       # Afficher l'aide
)

# Couleurs
$ColorTitle = "Cyan"
$ColorSuccess = "Green"
$ColorWarning = "Yellow"
$ColorError = "Red"
$ColorInfo = "Gray"

function Show-Help {
    Write-Host ""
    Write-Host "=== Go Benchmarks Runner ===" -ForegroundColor $ColorTitle
    Write-Host ""
    Write-Host "USAGE:" -ForegroundColor $ColorSuccess
    Write-Host "  .\run-go-benchmarks.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "OPTIONS:" -ForegroundColor $ColorSuccess
    Write-Host "  -Package <name>       Package à benchmarker (default: all)"
    Write-Host "                        Valeurs: all, export, stats, infrastructure, cache, workerpool"
    Write-Host ""
    Write-Host "  -Integration          Uniquement les benchmarks d'intégration (avec PostgreSQL)"
    Write-Host "  -Unit                 Uniquement les benchmarks unitaires (sans DB)"
    Write-Host "  -Short                Skip les benchmarks lents (V1 avec 365 jours)"
    Write-Host ""
    Write-Host "  -Compare              Générer un rapport de comparaison avec benchstat"
    Write-Host "  -Profile <type>       Profiling CPU ou mémoire (cpu|mem)"
    Write-Host "  -Count <n>            Nombre de runs pour stats (default: 1)"
    Write-Host "  -BenchTime <duration> Durée ou nombre d'itérations (default: 1s)"
    Write-Host "  -Save                 Sauvegarder les résultats avec timestamp"
    Write-Host ""
    Write-Host "EXEMPLES:" -ForegroundColor $ColorSuccess
    Write-Host "  # Tous les benchmarks d'intégration"
    Write-Host "  .\run-go-benchmarks.ps1 -Integration" -ForegroundColor $ColorInfo
    Write-Host ""
    Write-Host "  # Benchmarks Export uniquement"
    Write-Host "  .\run-go-benchmarks.ps1 -Package export -Integration" -ForegroundColor $ColorInfo
    Write-Host ""
    Write-Host "  # Comparaison statistique (10 runs)"
    Write-Host "  .\run-go-benchmarks.ps1 -Package export -Count 10 -Save" -ForegroundColor $ColorInfo
    Write-Host ""
    Write-Host "  # Profiling CPU sur Stats"
    Write-Host "  .\run-go-benchmarks.ps1 -Package stats -Profile cpu" -ForegroundColor $ColorInfo
    Write-Host ""
    Write-Host "  # Skip les benchmarks lents"
    Write-Host "  .\run-go-benchmarks.ps1 -Short" -ForegroundColor $ColorInfo
    Write-Host ""
    exit 0
}

if ($Help) {
    Show-Help
}

Write-Host ""
Write-Host "=== Go Benchmarks Runner ===" -ForegroundColor $ColorTitle
Write-Host ""

# Changer le répertoire de travail à la racine du projet
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
Set-Location $projectRoot
Write-Host "Working directory: $projectRoot" -ForegroundColor $ColorInfo
Write-Host ""

# Vérifier les prérequis si benchmarks d'intégration
if (-not $Unit) {
    Write-Host "Checking prerequisites..." -ForegroundColor $ColorWarning

    # Vérifier que Docker est accessible
    try {
        $null = docker ps 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "Docker not running"
        }
    } catch {
        Write-Host "[ERROR] Docker not running" -ForegroundColor $ColorError
        Write-Host "Start Docker Desktop first" -ForegroundColor $ColorInfo
        exit 1
    }

    # Vérifier que le container PostgreSQL existe et tourne
    try {
        $containerStatus = docker ps --filter "name=eval-postgres" --format "{{.Status}}" 2>&1
        if ([string]::IsNullOrEmpty($containerStatus)) {
            throw "Container not running"
        }
        Write-Host "[OK] PostgreSQL container is running" -ForegroundColor $ColorSuccess
    } catch {
        Write-Host "[ERROR] PostgreSQL container not running" -ForegroundColor $ColorError
        Write-Host "Start PostgreSQL with: docker-compose up -d" -ForegroundColor $ColorInfo
        Write-Host "Wait 10 seconds for PostgreSQL to be ready, then retry" -ForegroundColor $ColorInfo
        exit 1
    }

    # Vérifier que PostgreSQL accepte les connexions
    Write-Host "Checking PostgreSQL connection..." -ForegroundColor $ColorInfo
    $maxRetries = 10
    $retryCount = 0
    $isReady = $false

    while (-not $isReady -and $retryCount -lt $maxRetries) {
        try {
            $result = docker exec eval-postgres pg_isready -U evaluser 2>&1
            if ($result -match "accepting connections") {
                $isReady = $true
                Write-Host "[OK] PostgreSQL is accepting connections" -ForegroundColor $ColorSuccess
            } else {
                $retryCount++
                if ($retryCount -lt $maxRetries) {
                    Write-Host "  Waiting for PostgreSQL... ($retryCount/$maxRetries)" -ForegroundColor $ColorInfo
                    Start-Sleep -Seconds 2
                }
            }
        } catch {
            $retryCount++
            if ($retryCount -lt $maxRetries) {
                Write-Host "  Waiting for PostgreSQL... ($retryCount/$maxRetries)" -ForegroundColor $ColorInfo
                Start-Sleep -Seconds 2
            }
        }
    }

    if (-not $isReady) {
        Write-Host "[ERROR] PostgreSQL not ready after $maxRetries attempts" -ForegroundColor $ColorError
        Write-Host "Try: docker-compose restart" -ForegroundColor $ColorInfo
        exit 1
    }

    # Vérifier données seed
    try {
        $count = docker exec eval-postgres psql -U evaluser -d evaldb -t -c "SELECT COUNT(*) FROM orders;" 2>&1
        $count = $count.Trim()
        if ($count -match '^\d+$' -and [int]$count -gt 0) {
            Write-Host "[OK] Database has data ($count orders)" -ForegroundColor $ColorSuccess
        } else {
            Write-Host "[WARNING] Database is empty or error checking data" -ForegroundColor $ColorWarning
            Write-Host "Seed the database with: go run cmd/seed/main.go" -ForegroundColor $ColorInfo
            $response = Read-Host "Continue anyway? (y/n)"
            if ($response -ne "y") {
                exit 1
            }
        }
    } catch {
        Write-Host "[WARNING] Could not check database data" -ForegroundColor $ColorWarning
    }

    Write-Host ""
}

# Créer le dossier de résultats
$resultsDir = ".\benchmarks\results\go"
if (-not (Test-Path $resultsDir)) {
    New-Item -ItemType Directory -Path $resultsDir -Force | Out-Null
}

# Timestamp pour les fichiers
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"

# Construire les arguments de base
$baseArgs = @("-bench=.")
$baseArgs += "-benchmem"

if ($Short) {
    $baseArgs += "-short"
}

if ($Count -gt 1) {
    $baseArgs += "-count=$Count"
}

if ($BenchTime -ne "1s") {
    $baseArgs += "-benchtime=$BenchTime"
}

# Profiling
$profileFile = ""
if ($Profile -eq "cpu") {
    $profileFile = "$resultsDir\cpu_$timestamp.prof"
    $baseArgs += "-cpuprofile=$profileFile"
    Write-Host "CPU profiling enabled -> $profileFile" -ForegroundColor $ColorInfo
} elseif ($Profile -eq "mem") {
    $profileFile = "$resultsDir\mem_$timestamp.prof"
    $baseArgs += "-memprofile=$profileFile"
    Write-Host "Memory profiling enabled -> $profileFile" -ForegroundColor $ColorInfo
}

# Déterminer les packages à benchmarker
$packages = @()

switch ($Package.ToLower()) {
    "all" {
        if ($Integration) {
            $packages = @(
                "./internal/export/application",
                "./internal/analytics/application"
            )
        } elseif ($Unit) {
            $packages = @(
                "./internal/shared/infrastructure",
                "./internal/export/domain"
            )
        } else {
            $packages = @(
                "./internal/export/application",
                "./internal/analytics/application",
                "./internal/shared/infrastructure",
                "./internal/export/domain"
            )
        }
    }
    "export" {
        $packages = @("./internal/export/application")
    }
    "stats" {
        $packages = @("./internal/analytics/application")
    }
    "infrastructure" {
        $packages = @("./internal/shared/infrastructure")
    }
    "cache" {
        $packages = @("./internal/shared/infrastructure")
        $baseArgs[0] = "-bench=.*Cache.*"
    }
    "workerpool" {
        $packages = @("./internal/shared/infrastructure")
        $baseArgs[0] = "-bench=.*WorkerPool.*"
    }
    default {
        Write-Host "[ERROR] Invalid package: $Package" -ForegroundColor $ColorError
        Write-Host "Valid values: all, export, stats, infrastructure, cache, workerpool" -ForegroundColor $ColorInfo
        exit 1
    }
}

# Filtrer par type de benchmark
if ($Integration) {
    $baseArgs[0] = "-bench=.*RealDB.*"
    Write-Host "Running INTEGRATION benchmarks (with PostgreSQL)" -ForegroundColor $ColorWarning
} elseif ($Unit) {
    $baseArgs[0] = "-bench=."
    Write-Host "Running UNIT benchmarks (no database)" -ForegroundColor $ColorWarning
}

Write-Host ""

# Exécuter les benchmarks
$allResults = @()

foreach ($pkg in $packages) {
    $pkgName = Split-Path $pkg -Leaf

    Write-Host "=== Running benchmarks: $pkgName ===" -ForegroundColor $ColorTitle
    Write-Host "Package: $pkg" -ForegroundColor $ColorInfo
    Write-Host "Args: $baseArgs" -ForegroundColor $ColorInfo
    Write-Host ""

    $outputFile = ""
    if ($Save) {
        $outputFile = "$resultsDir\${pkgName}_${timestamp}.txt"
        Write-Host "Saving results to: $outputFile" -ForegroundColor $ColorInfo
        Write-Host ""
    }

    # Exécuter le benchmark
    if ($Save) {
        $result = & go test $baseArgs $pkg 2>&1 | Tee-Object -FilePath $outputFile
        $allResults += $outputFile
    } else {
        & go test $baseArgs $pkg
    }

    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "[WARNING] Some benchmarks failed or were skipped" -ForegroundColor $ColorWarning
    }

    Write-Host ""
}

# Résumé
Write-Host "=== Summary ===" -ForegroundColor $ColorTitle
Write-Host ""

if ($Save) {
    Write-Host "Results saved to:" -ForegroundColor $ColorSuccess
    foreach ($file in $allResults) {
        Write-Host "  $file" -ForegroundColor $ColorInfo
    }
    Write-Host ""
}

if ($profileFile) {
    Write-Host "Profile saved to:" -ForegroundColor $ColorSuccess
    Write-Host "  $profileFile" -ForegroundColor $ColorInfo
    Write-Host ""
    Write-Host "Analyze with:" -ForegroundColor $ColorWarning
    Write-Host "  go tool pprof -http=:8081 $profileFile" -ForegroundColor $ColorInfo
    Write-Host ""
}

if ($Compare -and $Save -and $Count -gt 1) {
    Write-Host "Generating benchstat report..." -ForegroundColor $ColorWarning

    # Vérifier si benchstat est installé
    $benchstatExists = $null -ne (Get-Command benchstat -ErrorAction SilentlyContinue)

    if (-not $benchstatExists) {
        Write-Host "[WARNING] benchstat not installed" -ForegroundColor $ColorWarning
        Write-Host "Install with: go install golang.org/x/perf/cmd/benchstat@latest" -ForegroundColor $ColorInfo
    } else {
        foreach ($file in $allResults) {
            Write-Host ""
            Write-Host "=== Statistics for $(Split-Path $file -Leaf) ===" -ForegroundColor $ColorTitle
            & benchstat $file
        }
    }
    Write-Host ""
}

# Commandes suggérées
Write-Host "=== Next Steps ===" -ForegroundColor $ColorTitle
Write-Host ""

if (-not $Save) {
    Write-Host "Save results for comparison:" -ForegroundColor $ColorWarning
    Write-Host "  .\run-go-benchmarks.ps1 -Package $Package -Count 10 -Save" -ForegroundColor $ColorInfo
    Write-Host ""
}

if ($Save -and $Count -gt 1) {
    Write-Host "Compare with future runs:" -ForegroundColor $ColorWarning
    Write-Host "  # 1. Make your optimizations..." -ForegroundColor $ColorInfo
    Write-Host "  # 2. Run benchmarks again with -Save" -ForegroundColor $ColorInfo
    Write-Host "  # 3. Compare with benchstat:" -ForegroundColor $ColorInfo
    Write-Host "  benchstat baseline.txt optimized.txt" -ForegroundColor $ColorInfo
    Write-Host ""
}

if (-not $Profile) {
    Write-Host "Profile performance:" -ForegroundColor $ColorWarning
    Write-Host "  .\run-go-benchmarks.ps1 -Package $Package -Profile cpu" -ForegroundColor $ColorInfo
    Write-Host ""
}

if ($Integration -and -not $Unit) {
    Write-Host "Run unit benchmarks (faster):" -ForegroundColor $ColorWarning
    Write-Host "  .\run-go-benchmarks.ps1 -Unit" -ForegroundColor $ColorInfo
    Write-Host ""
}

Write-Host "=== Benchmark complete ===" -ForegroundColor $ColorTitle
Write-Host ""

# Ouvrir pprof si profiling activé
if ($profileFile) {
    $response = Read-Host "Open pprof in browser? (y/n)"
    if ($response -eq "y") {
        Write-Host "Opening pprof at http://localhost:8081..." -ForegroundColor $ColorInfo
        Start-Process "http://localhost:8081"
        & go tool pprof -http=:8081 $profileFile
    }
}
