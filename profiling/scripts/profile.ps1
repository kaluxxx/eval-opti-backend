# Script de profiling pour l'application Go
# Usage: .\profile.ps1 [cpu|mem|all]

param(
    [Parameter(Position=0)]
    [ValidateSet("cpu", "mem", "all", "bench")]
    [string]$Type = "all",

    [Parameter()]
    [int]$Duration = 30,

    [Parameter()]
    [string]$ServerUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "   Profiling de l'application Go" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Crée un dossier pour les profils (relatif au script)
$profileDir = "..\profiles"
if (-not (Test-Path $profileDir)) {
    New-Item -ItemType Directory -Path $profileDir | Out-Null
    Write-Host "[+] Dossier 'profiles' créé" -ForegroundColor Green
}

$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"

function Test-ServerRunning {
    try {
        $response = Invoke-WebRequest -Uri "$ServerUrl/api/health" -Method GET -TimeoutSec 2
        return $true
    } catch {
        return $false
    }
}

function Capture-CPUProfile {
    Write-Host "`n[*] Capture du profil CPU (durée: $Duration secondes)..." -ForegroundColor Yellow
    Write-Host "[*] Faites des requêtes pendant ce temps pour capturer le profil sous charge" -ForegroundColor Yellow

    $cpuFile = "$profileDir\cpu_$timestamp.prof"

    try {
        Invoke-WebRequest -Uri "$ServerUrl/debug/pprof/profile?seconds=$Duration" -OutFile $cpuFile
        Write-Host "[+] Profil CPU sauvegardé: $cpuFile" -ForegroundColor Green

        # Analyse du profil
        Write-Host "[*] Analyse du profil CPU..." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Top 10 des fonctions les plus coûteuses:" -ForegroundColor Cyan
        go tool pprof -top -nodecount=10 $cpuFile

        Write-Host "`n[i] Pour une analyse interactive, exécutez:" -ForegroundColor Magenta
        Write-Host "    go tool pprof $cpuFile" -ForegroundColor White
        Write-Host "[i] Pour visualiser en web UI:" -ForegroundColor Magenta
        Write-Host "    go tool pprof -http=:8081 $cpuFile" -ForegroundColor White
    } catch {
        Write-Host "[-] Erreur lors de la capture du profil CPU: $_" -ForegroundColor Red
    }
}

function Capture-MemProfile {
    Write-Host "`n[*] Capture du profil mémoire..." -ForegroundColor Yellow

    $memFile = "$profileDir\mem_$timestamp.prof"

    try {
        Invoke-WebRequest -Uri "$ServerUrl/debug/pprof/heap" -OutFile $memFile
        Write-Host "[+] Profil mémoire sauvegardé: $memFile" -ForegroundColor Green

        # Analyse du profil
        Write-Host "[*] Analyse du profil mémoire..." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Top 10 des allocations mémoire:" -ForegroundColor Cyan
        go tool pprof -top -nodecount=10 $memFile

        Write-Host "`n[i] Pour une analyse interactive, exécutez:" -ForegroundColor Magenta
        Write-Host "    go tool pprof $memFile" -ForegroundColor White
        Write-Host "[i] Pour visualiser en web UI:" -ForegroundColor Magenta
        Write-Host "    go tool pprof -http=:8081 $memFile" -ForegroundColor White
    } catch {
        Write-Host "[-] Erreur lors de la capture du profil mémoire: $_" -ForegroundColor Red
    }
}

function Run-Benchmarks {
    Write-Host "`n[*] Exécution des benchmarks..." -ForegroundColor Yellow
    Write-Host ""

    $benchFile = "$profileDir\bench_$timestamp.txt"

    # Exécute les benchmarks avec profiling mémoire
    go test -bench=. -benchmem -cpuprofile="$profileDir\bench_cpu_$timestamp.prof" -memprofile="$profileDir\bench_mem_$timestamp.prof" | Tee-Object -FilePath $benchFile

    Write-Host ""
    Write-Host "[+] Résultats des benchmarks sauvegardés: $benchFile" -ForegroundColor Green
    Write-Host "[+] Profils CPU et mémoire des benchmarks sauvegardés" -ForegroundColor Green

    Write-Host "`n[i] Pour analyser les profils des benchmarks:" -ForegroundColor Magenta
    Write-Host "    go tool pprof -http=:8081 $profileDir\bench_cpu_$timestamp.prof" -ForegroundColor White
    Write-Host "    go tool pprof -http=:8081 $profileDir\bench_mem_$timestamp.prof" -ForegroundColor White
}

function Generate-Load {
    Write-Host "`n[*] Génération de charge pour le profiling..." -ForegroundColor Yellow
    Write-Host "[*] Envoi de 5 requêtes /api/stats?days=365..." -ForegroundColor Yellow

    for ($i = 1; $i -le 5; $i++) {
        Write-Host "  Requête $i/5..." -ForegroundColor Gray
        try {
            Invoke-WebRequest -Uri "$ServerUrl/api/stats?days=365" -Method GET -TimeoutSec 60 | Out-Null
        } catch {
            Write-Host "  Erreur sur la requête $i" -ForegroundColor Red
        }
    }

    Write-Host "[+] Charge générée" -ForegroundColor Green
}

# Vérification du serveur
if ($Type -ne "bench") {
    Write-Host "[*] Vérification que le serveur est en cours d'exécution..." -ForegroundColor Yellow
    if (-not (Test-ServerRunning)) {
        Write-Host "[-] ERREUR: Le serveur n'est pas accessible sur $ServerUrl" -ForegroundColor Red
        Write-Host "[!] Démarrez le serveur avec: go run main.go" -ForegroundColor Yellow
        exit 1
    }
    Write-Host "[+] Serveur accessible" -ForegroundColor Green
}

# Exécution selon le type demandé
switch ($Type) {
    "cpu" {
        Start-Job -ScriptBlock { param($url) Start-Sleep -Seconds 2; Invoke-WebRequest -Uri "$url/api/stats?days=365" -Method GET -TimeoutSec 60 | Out-Null } -ArgumentList $ServerUrl | Out-Null
        Capture-CPUProfile
    }
    "mem" {
        Generate-Load
        Capture-MemProfile
    }
    "bench" {
        Run-Benchmarks
    }
    "all" {
        Write-Host "[*] Profiling complet: CPU + Mémoire" -ForegroundColor Yellow

        # Lance une charge en arrière-plan
        $job = Start-Job -ScriptBlock {
            param($url)
            Start-Sleep -Seconds 2
            for ($i = 1; $i -le 10; $i++) {
                try {
                    Invoke-WebRequest -Uri "$url/api/stats?days=365" -Method GET -TimeoutSec 60 | Out-Null
                } catch {}
                Start-Sleep -Milliseconds 500
            }
        } -ArgumentList $ServerUrl

        Capture-CPUProfile
        Wait-Job $job | Out-Null
        Remove-Job $job

        Generate-Load
        Capture-MemProfile
    }
}

Write-Host "`n=====================================" -ForegroundColor Cyan
Write-Host "   Profiling terminé!" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "[i] Les profils sont disponibles dans le dossier: $profileDir" -ForegroundColor Magenta
Write-Host "[i] Pour une analyse visuelle interactive:" -ForegroundColor Magenta
Write-Host "    go tool pprof -http=:8081 <fichier.prof>" -ForegroundColor White