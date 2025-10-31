# Guide des Benchmarks Go

Ce document explique comment utiliser les benchmarks Go pour mesurer les performances du projet `eval`.

## 📊 Types de Benchmarks

Le projet contient **deux types** de benchmarks :

### 1. **Benchmarks Unitaires** (`*_test.go`)

**Ce qu'ils mesurent** :
- Performance du code Go pur (manipulation de strings, allocations)
- Micro-optimisations (fmt.Sprintf vs strconv)
- Algorithmes (tri, hashing)
- Structures de données (slices, maps)

**Avantages** :
- ✅ Rapides (millisecondes)
- ✅ Pas besoin de DB
- ✅ Isolent les optimisations spécifiques

**Limites** :
- ❌ Ne mesurent PAS les requêtes SQL
- ❌ Ne mesurent PAS la latence réseau
- ❌ Ne mesurent PAS les I/O réelles

**Exemples** :
```bash
# Benchmarks unitaires du cache
go test -bench=BenchmarkFNV32 ./internal/shared/infrastructure/

# Benchmarks unitaires des conversions de strings
go test -bench=BenchmarkStringFormat ./internal/export/application/
```

---

### 2. **Benchmarks d'Intégration** (`*_integration_test.go`)

**Ce qu'ils mesurent** :
- 🔥 **Performance RÉELLE** avec PostgreSQL
- Latence des requêtes SQL (JOIN, GROUP BY, ORDER BY)
- Transfert réseau (DB → App)
- Parsing des résultats SQL → structs Go
- Impact du cache (hit vs miss)
- Performance des goroutines avec I/O
- Worker pool avec données réelles

**Avantages** :
- ✅ Mesure les vraies performances end-to-end
- ✅ Inclut latence SQL + réseau
- ✅ Quantifie les gains V1 vs V2
- ✅ Détecte les régressions réelles

**Limites** :
- ⚠️ Nécessite PostgreSQL en cours d'exécution
- ⚠️ Plus lents (secondes)
- ⚠️ Nécessite données seed

**Exemples** :
```bash
# Benchmarks d'intégration des exports
go test -bench=BenchmarkExportServiceV2_RealDB ./internal/export/application/

# Benchmarks d'intégration des stats
go test -bench=BenchmarkStatsServiceV2_RealDB ./internal/analytics/application/
```

---

## 🚀 Prérequis

### Pour les Benchmarks Unitaires

Aucun prérequis, ils fonctionnent directement.

### Pour les Benchmarks d'Intégration

1. **PostgreSQL en cours d'exécution**

```bash
# Via Docker Compose
docker-compose up -d postgres

# Ou via service local
# Windows: net start postgresql
# Linux: sudo service postgresql start
```

2. **Base de données seed avec données**

```bash
# Seed la base de données (génère les données de test)
go run cmd/seed/main.go
```

3. **Variables d'environnement** (fichier `.env`)

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=evaluser
DB_PASSWORD=evalpass
DB_NAME=evaldb
DB_SSLMODE=disable
```

---

## 📝 Commandes de Base

### Exécuter TOUS les benchmarks d'un package

```bash
# Tous les benchmarks (unitaires + intégration)
go test -bench=. ./internal/export/application/

# Avec métriques mémoire détaillées
go test -bench=. -benchmem ./internal/export/application/
```

### Exécuter un benchmark spécifique

```bash
# Export CSV 30 jours (V2)
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_30Days ./internal/export/application/

# Stats avec cache hit
go test -bench=BenchmarkStatsServiceV2_RealDB_CacheHit ./internal/analytics/application/
```

### Comparer V1 vs V2

```bash
# Comparaison directe Export V1 vs V2
go test -bench=BenchmarkComparison_V1_vs_V2_CSV ./internal/export/application/

# Comparaison directe Stats V1 vs V2
go test -bench=BenchmarkComparison_V1_vs_V2_Stats ./internal/analytics/application/
```

---

## 🔬 Options Avancées

### 1. Nombre d'itérations (`-benchtime`)

Par défaut, Go exécute les benchmarks pendant 1 seconde. Vous pouvez augmenter :

```bash
# Exécuter pendant 5 secondes
go test -bench=. -benchtime=5s ./internal/export/application/

# Exécuter exactement 100 fois
go test -bench=. -benchtime=100x ./internal/export/application/
```

### 2. Statistiques avec `-count`

Pour obtenir des résultats statistiquement significatifs :

```bash
# Exécuter 10 fois et calculer moyenne/écart-type
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_30Days -benchmem -count=10 ./internal/export/application/
```

### 3. Profiling CPU et Mémoire

```bash
# CPU profiling
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_365Days \
    -benchmem \
    -cpuprofile=cpu.prof \
    ./internal/export/application/

# Analyser le profil
go tool pprof -http=:8081 cpu.prof

# Memory profiling
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_365Days \
    -benchmem \
    -memprofile=mem.prof \
    ./internal/export/application/

# Analyser le profil
go tool pprof -http=:8081 mem.prof
```

### 4. Skip les benchmarks lents (`-short`)

Les benchmarks V1 avec 365 jours sont très lents. Pour les skip :

```bash
# Skip les benchmarks marqués avec testing.Short()
go test -bench=. -short ./internal/analytics/application/
```

---

## 📊 Utiliser `benchstat` pour Comparer

`benchstat` permet de comparer scientifiquement deux séries de benchmarks.

### Installation

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

### Workflow de Comparaison

```bash
# 1. Baseline (avant optimisation)
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_30Days -benchmem -count=10 \
    ./internal/export/application/ | tee baseline.txt

# 2. Faire vos optimisations dans le code...

# 3. Nouveau benchmark (après optimisation)
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_30Days -benchmem -count=10 \
    ./internal/export/application/ | tee optimized.txt

# 4. Comparer avec benchstat
benchstat baseline.txt optimized.txt
```

### Exemple de Sortie `benchstat`

```
name                                    old time/op    new time/op    delta
ExportServiceV2_RealDB_CSV_30Days-8       45.2ms ± 2%    27.1ms ± 1%   -40.04%  (p=0.000 n=10+10)

name                                    old alloc/op   new alloc/op   delta
ExportServiceV2_RealDB_CSV_30Days-8       5.24MB ± 0%    2.11MB ± 0%   -59.73%  (p=0.000 n=10+10)

name                                    old allocs/op  new allocs/op  delta
ExportServiceV2_RealDB_CSV_30Days-8        12.3k ± 0%      3.2k ± 0%   -73.98%  (p=0.000 n=10+10)
```

**Interprétation** :
- `-40.04%` : 40% plus rapide
- `-59.73%` : 60% moins de mémoire allouée
- `-73.98%` : 74% moins d'allocations
- `p=0.000` : Statistiquement significatif (< 0.05)

---

## 🎯 Benchmarks par Catégorie

### Export Service

#### **CSV Export (V2 Optimisé)**
```bash
# Petite charge (7 jours)
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_7Days -benchmem ./internal/export/application/

# Charge moyenne (30 jours)
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_30Days -benchmem ./internal/export/application/

# Charge élevée (365 jours)
go test -bench=BenchmarkExportServiceV2_RealDB_CSV_365Days -benchmem ./internal/export/application/
```

#### **CSV Export (V1 Baseline - N+1)**
```bash
# V1 avec N+1 queries (LENT!)
go test -bench=BenchmarkExportServiceV1_RealDB_CSV_30Days -benchmem ./internal/export/application/

# V1 vs V2 comparaison
go test -bench=BenchmarkComparison_V1_vs_V2_CSV -benchmem ./internal/export/application/
```

#### **Parquet Export (WorkerPool)**
```bash
# Export Parquet avec 4 workers
go test -bench=BenchmarkExportServiceV2_RealDB_Parquet -benchmem ./internal/export/application/
```

#### **Requêtes SQL Isolées**
```bash
# Mesurer uniquement la requête SQL (sans génération CSV)
go test -bench=BenchmarkExportServiceV2_RealDB_QueryOnly -benchmem ./internal/export/application/

# Mesurer les N+1 queries de V1
go test -bench=BenchmarkExportServiceV1_RealDB_QueryOnly -benchmem ./internal/export/application/
```

---

### Stats Service

#### **Stats avec Cache Hit**
```bash
# Cache hit (données déjà en cache) - TRÈS RAPIDE
go test -bench=BenchmarkStatsServiceV2_RealDB_CacheHit -benchmem ./internal/analytics/application/
```

#### **Stats avec Cache Miss**
```bash
# Cache miss (premier appel, calcul complet)
go test -bench=BenchmarkStatsServiceV2_RealDB_CacheMiss -benchmem ./internal/analytics/application/
```

#### **Impact du Cache**
```bash
# Compare cache hit vs cache miss
go test -bench=BenchmarkStatsServiceV2_RealDB_CacheImpact -benchmem ./internal/analytics/application/
```

#### **Stats V1 (N+1 + Bubble Sort)**
```bash
# V1 avec N+1 queries + bubble sort O(n²) (TRÈS LENT!)
go test -bench=BenchmarkStatsServiceV1_RealDB -benchmem -short ./internal/analytics/application/

# V1 vs V2 comparaison
go test -bench=BenchmarkComparison_V1_vs_V2_Stats -benchmem ./internal/analytics/application/
```

#### **Requêtes SQL Isolées**
```bash
# GetGlobalStats (agrégation SQL)
go test -bench=BenchmarkStatsRepo_RealDB_GetGlobalStats -benchmem ./internal/analytics/application/

# GetCategoryStats (GROUP BY)
go test -bench=BenchmarkStatsRepo_RealDB_GetCategoryStats -benchmem ./internal/analytics/application/

# GetTopProducts (JOIN + ORDER BY + LIMIT)
go test -bench=BenchmarkStatsRepo_RealDB_GetTopProducts -benchmem ./internal/analytics/application/
```

---

### Infrastructure

#### **Cache Performance**
```bash
# Cache InMemoryCache vs ShardedCache
go test -bench=BenchmarkComparison_InMemory_vs_Sharded -benchmem ./internal/shared/infrastructure/

# Cache avec haute contention (concurrence)
go test -bench=BenchmarkShardedCache.*HighContention -benchmem ./internal/shared/infrastructure/

# Hash FNV-1a performance
go test -bench=BenchmarkFNV32 -benchmem ./internal/shared/infrastructure/
```

#### **WorkerPool Performance**
```bash
# Variation du nombre de workers
go test -bench=BenchmarkWorkerPool.*Workers -benchmem ./internal/shared/infrastructure/

# WorkerPool vs goroutines directes
go test -bench=BenchmarkComparison_WorkerPool_vs_Goroutines -benchmem ./internal/shared/infrastructure/

# Throughput du worker pool
go test -bench=BenchmarkWorkerPool_Throughput -benchmem ./internal/shared/infrastructure/
```

#### **Domain Objects**
```bash
# ToCSVRow optimisations (fmt.Sprintf vs strconv)
go test -bench=BenchmarkToCSVRow -benchmem ./internal/export/domain/
```

---

## 🧪 Tests de Charge Concurrente

### Export Service sous Charge

```bash
# Charge concurrente sur Export CSV
go test -bench=BenchmarkExportServiceV2_RealDB_ConcurrentLoad -benchmem ./internal/export/application/
```

### Stats Service sous Charge

```bash
# Charge concurrente sur Stats
go test -bench=BenchmarkStatsServiceV2_RealDB_ConcurrentLoad -benchmem ./internal/analytics/application/

# Multi-périodes en concurrence (7d, 30d, 90d, 365d)
go test -bench=BenchmarkStatsServiceV2_RealDB_MultiPeriod_Concurrent -benchmem ./internal/analytics/application/
```

---

## 📈 Résultats Attendus

### Export CSV (V1 vs V2)

| Benchmark | V1 (N+1) | V2 (JOIN) | Gain |
|-----------|----------|-----------|------|
| CSV 7 days | ~800ms | ~15ms | **53x plus rapide** |
| CSV 30 days | ~2100ms | ~40ms | **52x plus rapide** |
| CSV 365 days | ~25s | ~450ms | **55x plus rapide** |

**Pourquoi ?**
- V1 : N+1 queries (1 + 6×N requêtes SQL pour N lignes)
- V2 : 1 seule requête avec JOINs

### Stats (V1 vs V2)

| Benchmark | V1 (N+1+Sort) | V2 (Cache Miss) | V2 (Cache Hit) | Gain |
|-----------|---------------|-----------------|----------------|------|
| Stats 7 days | ~200ms | ~20ms | ~0.1ms | **2000x avec cache** |
| Stats 30 days | ~500ms | ~30ms | ~0.1ms | **5000x avec cache** |
| Stats 365 days | ~5s | ~150ms | ~0.1ms | **50000x avec cache** |

**Pourquoi ?**
- V1 : N+1 queries + bubble sort O(n²) + pas de cache
- V2 Cache Miss : 5 goroutines parallèles + agrégations SQL
- V2 Cache Hit : Lecture mémoire pure (< 1ms)

---

## 🛠️ Scripts PowerShell

### Script de Benchmarking Complet

Le projet inclut un script PowerShell pour automatiser les benchmarks :

```powershell
# Exécuter tous les benchmarks d'intégration
.\benchmarks\scripts\run-go-benchmarks.ps1

# Exécuter uniquement les benchmarks d'export
.\benchmarks\scripts\run-go-benchmarks.ps1 -Package export

# Exécuter avec profiling CPU
.\benchmarks\scripts\run-go-benchmarks.ps1 -Profile cpu

# Générer un rapport avec benchstat
.\benchmarks\scripts\run-go-benchmarks.ps1 -Compare
```

---

## 🐛 Troubleshooting

### Erreur : "Database not available"

**Cause** : PostgreSQL n'est pas démarré ou `.env` mal configuré.

**Solution** :
```bash
# 1. Vérifier que PostgreSQL tourne
docker-compose ps
# ou
psql -h localhost -U evaluser -d evaldb

# 2. Vérifier le fichier .env
cat .env

# 3. Tester la connexion
go test -v -run=TestDatabaseConnection ./internal/testhelpers/
```

### Benchmarks Très Lents

**Cause** : Benchmarks V1 avec beaucoup de données (N+1 problem).

**Solution** :
```bash
# Skip les benchmarks lents
go test -bench=. -short ./...

# Ou utiliser uniquement V2
go test -bench=.*V2.* ./...
```

### Résultats Instables

**Cause** : Variabilité due à d'autres processus, garbage collector, etc.

**Solution** :
```bash
# 1. Augmenter le nombre de runs
go test -bench=. -benchtime=10s -count=10 ./...

# 2. Utiliser benchstat pour statistiques
go test -bench=. -count=10 | tee results.txt
benchstat results.txt

# 3. Réduire la charge système
# - Fermer les applications inutiles
# - Désactiver antivirus temporairement
```

---

## 📚 Ressources

- [Official Go Benchmarking Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [benchstat Documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Dave Cheney - Benchmarking Go Programs](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [pprof Profiling Guide](https://go.dev/blog/pprof)

---

## ✅ Checklist: Avant de Commit une Optimisation

- [ ] Créer un baseline avec `go test -bench=. -count=10 | tee baseline.txt`
- [ ] Faire l'optimisation
- [ ] Re-benchmarker avec `go test -bench=. -count=10 | tee optimized.txt`
- [ ] Comparer avec `benchstat baseline.txt optimized.txt`
- [ ] Vérifier que le gain est statistiquement significatif (p < 0.05)
- [ ] Vérifier que les tests passent toujours (`go test ./...`)
- [ ] Documenter le gain dans le commit message

---

**Dernière mise à jour** : {{ date }}
**Auteur** : Claude Code
