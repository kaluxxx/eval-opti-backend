#!/bin/bash

# Script de benchmark pour comparer V1 vs V2 avec hyperfine

echo "=== Benchmark API V1 vs V2 avec Hyperfine ==="
echo ""

# Démarre le serveur en arrière-plan
echo "Démarrage du serveur..."
go run main.go &
SERVER_PID=$!
sleep 3

echo "Serveur démarré (PID: $SERVER_PID)"
echo ""

# Fonction pour nettoyer à la fin
cleanup() {
    echo ""
    echo "Arrêt du serveur..."
    kill $SERVER_PID 2>/dev/null
    echo "✓ Serveur arrêté"
}
trap cleanup EXIT

# Teste que le serveur répond
echo "Test de connexion au serveur..."
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo "✓ Serveur opérationnel"
else
    echo "✗ Erreur: Le serveur ne répond pas"
    exit 1
fi
echo ""

# Benchmark 1: Stats JSON avec 365 jours
echo "=== Benchmark 1: Stats JSON (365 jours) ==="
hyperfine \
    --warmup 2 \
    --runs 10 \
    --export-markdown benchmark_stats_365.md \
    --export-json benchmark_stats_365.json \
    'curl -s http://localhost:8080/api/v1/stats?days=365 > /dev/null' \
    'curl -s http://localhost:8080/api/v2/stats?days=365 > /dev/null'
echo ""

# Benchmark 2: Stats JSON avec 100 jours
echo "=== Benchmark 2: Stats JSON (100 jours) ==="
hyperfine \
    --warmup 2 \
    --runs 10 \
    --export-markdown benchmark_stats_100.md \
    'curl -s http://localhost:8080/api/v1/stats?days=100 > /dev/null' \
    'curl -s http://localhost:8080/api/v2/stats?days=100 > /dev/null'
echo ""

# Benchmark 3: Export CSV (petit dataset)
echo "=== Benchmark 3: Export CSV (30 jours) ==="
hyperfine \
    --warmup 1 \
    --runs 5 \
    --export-markdown benchmark_csv_30.md \
    'curl -s http://localhost:8080/api/v1/export/csv?days=30 > /dev/null' \
    'curl -s http://localhost:8080/api/v2/export/csv?days=30 > /dev/null'
echo ""

# Benchmark 4: Test du cache V2
echo "=== Benchmark 4: Effet du cache V2 ==="
echo "Préchauffage du cache..."
curl -s http://localhost:8080/api/v2/stats?days=365 > /dev/null
sleep 1

hyperfine \
    --warmup 0 \
    --runs 50 \
    --export-markdown benchmark_cache.md \
    'curl -s http://localhost:8080/api/v2/stats?days=365 > /dev/null'
echo ""

echo "=== Résultats ==="
echo "Les résultats ont été exportés dans:"
echo "  - benchmark_stats_365.md (Stats 365 jours)"
echo "  - benchmark_stats_100.md (Stats 100 jours)"
echo "  - benchmark_csv_30.md (Export CSV 30 jours)"
echo "  - benchmark_cache.md (Performance du cache V2)"
echo ""

echo "=== Benchmark terminé ==="
