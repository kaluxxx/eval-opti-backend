# Projet Eval - Performance Optimization API avec Architecture DDD

Projet d'évaluation et d'optimisation de performance en Go, comparant deux versions d'une API de ventes avec une architecture **Domain-Driven Design (DDD)**.

## 🏛️ Architecture DDD

Ce projet utilise une architecture **Domain-Driven Design** avec séparation en **Bounded Contexts** :

### Structure du Projet

```
eval/
├── internal/                         # Code application (DDD)
│   ├── shared/                       # Shared Kernel
│   │   ├── domain/                   # Value Objects communs
│   │   │   ├── money.go              # Money value object
│   │   │   ├── daterange.go          # DateRange value object
│   │   │   └── quantity.go           # Quantity value object
│   │   └── infrastructure/           # Infrastructure partagée
│   │       ├── cache.go              # Cache avec TTL et sharding
│   │       ├── workerpool.go         # Worker pools réutilisables
│   │       └── repository.go         # Base repository (CQRS)
│   │
│   ├── catalog/                      # Bounded Context: Catalogue
│   │   ├── domain/                   # Entités du domaine
│   │   │   ├── product.go            # Product entity
│   │   │   ├── category.go           # Category entity
│   │   │   └── supplier.go           # Supplier entity
│   │   └── infrastructure/           # Repositories
│   │       └── product_query_repository.go
│   │
│   ├── orders/                       # Bounded Context: Commandes
│   │   ├── domain/                   # Aggregate Order
│   │   │   ├── order.go              # Order aggregate root
│   │   │   └── order_item.go         # OrderItem entity
│   │   └── infrastructure/           # Repositories
│   │       └── order_query_repository.go
│   │
│   ├── analytics/                    # Bounded Context: Analytics
│   │   ├── domain/                   # Entités Stats
│   │   │   └── stats.go              # Stats, CategoryStats, ProductStats...
│   │   ├── application/              # Domain Services
│   │   │   ├── stats_service_v1.go   # Service non-optimisé
│   │   │   └── stats_service_v2.go   # Service optimisé (cache + goroutines)
│   │   └── infrastructure/           # Repositories
│   │       └── stats_query_repository.go
│   │
│   └── export/                       # Bounded Context: Exports
│       ├── domain/                   # Entités Export
│       │   └── export_job.go         # ExportJob, SaleExportRow
│       ├── application/              # Services d'export
│       │   ├── export_service_v1.go  # Export non-optimisé (N+1)
│       │   └── export_service_v2.go  # Export optimisé (worker pools)
│       └── infrastructure/           # Repositories
│           └── export_query_repository.go
│
├── api/                              # Handlers HTTP
│   ├── v1/                           # API V1 - Non optimisée
│   │   └── handlers.go               # Handlers avec services V1
│   └── v2/                           # API V2 - Optimisée
│       └── handlers.go               # Handlers avec services V2
│
├── cmd/
│   └── seed/                         # Outil de seeding DB
│       └── main.go
│
├── database/                         # Ancienne couche DB (legacy)
│   ├── db.go
│   ├── models.go
│   └── seed.go
│
├── v1/ et v2/                        # Anciens handlers (legacy - conservés)
│
├── benchmarks/                       # Benchmarking avec Hyperfine
│   ├── scripts/                      # Scripts de benchmark
│   └── results/                      # Résultats des benchmarks
│
├── profiling/                        # Profiling avec pprof
│   ├── scripts/                      # Scripts de profiling
│   ├── profiles/                     # Profils générés
│   ├── PROFILING.md
│   └── PROFILING_RESULTS.md
│
├── docs/                             # Documentation
│   ├── BENCHMARK.md
│   ├── RESULTS.md
│   └── README_API.md
│
├── postman/                          # Collections Postman
│   └── API_Ventes.postman_collection.json
│
├── main.go                           # Bootstrap avec Dependency Injection
├── main.go.old                       # Ancien main (legacy)
└── go.mod
```

## 🎯 Patterns DDD Implémentés

### 1. **Value Objects** (Shared Kernel)
- `Money` : Montant avec devise et validation
- `DateRange` : Période temporelle avec validation
- `Quantity` : Quantité avec validation (>= 0)

### 2. **Entities & Aggregates**
- **Product** : Entité avec identité (ProductID)
- **Category** : Entité catégorie
- **Supplier** : Entité fournisseur
- **Order** : **Aggregate Root** avec règles métier
- **OrderItem** : Entity dans l'aggregate Order

### 3. **Repositories (Pattern CQRS)**
- **Query Repositories** : Lecture seule, optimisées
  - `ProductQueryRepository`
  - `OrderQueryRepository`
  - `StatsQueryRepository`
  - `ExportQueryRepository`

### 4. **Domain Services**
- `StatsServiceV1` : Service de stats non-optimisé (N+1, bubble sort)
- `StatsServiceV2` : Service optimisé (cache, goroutines parallèles)
- `ExportServiceV1` : Export inefficace (mémoire complète)
- `ExportServiceV2` : Export optimisé (worker pools, streaming)

### 5. **Dependency Injection**
- Injection complète dans `main.go`
- Facilite les tests unitaires
- Découplage infrastructure/domaine

## 📊 Bounded Contexts

| Context | Responsabilité | Entities/Aggregates |
|---------|---------------|---------------------|
| **Catalog** | Gestion du catalogue produits | Product, Category, Supplier |
| **Orders** | Gestion des commandes | Order (aggregate), OrderItem |
| **Analytics** | Calcul de statistiques | Stats, CategoryStats, ProductStats |
| **Export** | Exports CSV/Parquet | ExportJob, SaleExportRow |

## 🚀 Endpoints API

### V1 (Non optimisée - DDD)
- `GET /api/v1/stats?days=365` - Statistiques JSON (N+1 queries, bubble sort)
- `GET /api/v1/export/csv?days=30` - Export CSV (N+1 queries, mémoire complète)
- `GET /api/v1/export/stats-csv?days=365` - Export CSV statistiques
- `GET /api/v1/export/parquet?days=30` - Export Parquet (inefficace)

### V2 (Optimisée - DDD)
- `GET /api/v2/stats?days=365` - Statistiques JSON (cache 5min, goroutines parallèles)
- `GET /api/v2/export/csv?days=30` - Export CSV (requête optimisée, batch 1000)
- `GET /api/v2/export/stats-csv?days=365` - Export CSV stats (depuis cache)
- `GET /api/v2/export/parquet?days=30` - Export Parquet (worker pool 4 workers)

### Health
- `GET /api/health` - Status de l'application

## ⚡ Démarrage Rapide

### Prérequis
- Go 1.25+
- PostgreSQL (via Docker Compose)

### 1. Lancer la base de données
```bash
docker-compose up -d
```

### 2. Seeding de la base
```bash
# 5 ans de données par défaut
go run cmd/seed/main.go

# Ou spécifier le nombre d'années
SEED_YEARS=10 go run cmd/seed/main.go
```

### 3. Lancer le serveur
```bash
go run main.go
# Serveur disponible sur http://localhost:8080
```

### 4. Tester l'API
```bash
# V1 (non-optimisée)
curl "http://localhost:8080/api/v1/stats?days=365"

# V2 (optimisée)
curl "http://localhost:8080/api/v2/stats?days=365"
```

## 🧪 Tests & Benchmarks

### Tests unitaires
```bash
# Tous les tests
go test ./...

# Tests spécifiques
go test ./internal/analytics/application/...
```

### Benchmarks Go (NOUVEAUX - avec PostgreSQL)

Le projet inclut maintenant des **benchmarks d'intégration** qui mesurent les performances réelles avec PostgreSQL :

```bash
# Benchmarks d'intégration Export Service (avec DB)
go test -bench=BenchmarkExportServiceV2_RealDB -benchmem ./internal/export/application/

# Benchmarks d'intégration Stats Service (avec cache)
go test -bench=BenchmarkStatsServiceV2_RealDB -benchmem ./internal/analytics/application/

# Comparaison directe V1 vs V2
go test -bench=BenchmarkComparison_V1_vs_V2 -benchmem ./internal/export/application/

# Benchmarks unitaires (sans DB - plus rapides)
go test -bench=. -benchmem ./internal/shared/infrastructure/
```

**Script PowerShell automatisé** :
```powershell
# Tous les benchmarks d'intégration
.\benchmarks\scripts\run-go-benchmarks.ps1 -Integration

# Export uniquement
.\benchmarks\scripts\run-go-benchmarks.ps1 -Package export -Integration

# Avec profiling CPU
.\benchmarks\scripts\run-go-benchmarks.ps1 -Package stats -Profile cpu

# Sauvegarder pour comparaison
.\benchmarks\scripts\run-go-benchmarks.ps1 -Count 10 -Save

# Afficher l'aide
.\benchmarks\scripts\run-go-benchmarks.ps1 -Help
```

Voir [docs/BENCHMARKS.md](docs/BENCHMARKS.md) pour le guide complet.

### Benchmarks Hyperfine (HTTP end-to-end)
```powershell
# Windows PowerShell
.\benchmarks\scripts\benchmark.ps1
```

```bash
# Linux/Mac
./benchmarks/scripts/benchmark.sh
```

### Profiling pprof
```powershell
# Windows PowerShell
.\profiling\scripts\profile.ps1
```

```bash
# Linux/Mac
./profiling/scripts/profile.sh
```

## 📈 Résultats de Performance

### Comparaison V1 vs V2 (Architecture DDD)

| Opération | V1 | V2 | Amélioration |
|-----------|----|----|--------------|
| Stats 365 jours | 49.1 ms | 22.4 ms | **2.19x plus rapide** |
| Stats 100 jours | 31.4 ms | 23.3 ms | **1.35x plus rapide** |
| Export CSV 30 jours | 2072 ms | 40.3 ms | **51.48x plus rapide** |

### Mémoire

| Métrique | V1 | V2 | Amélioration |
|----------|----|----|--------------|
| Génération données | 345.70 MB | 137.15 MB | **60% moins de mémoire** |
| Export Parquet | ~345 MB | ~200 KB | **Streaming par batch** |

Voir [docs/RESULTS.md](docs/RESULTS.md) pour les résultats détaillés.

## 🔧 Optimisations Implémentées (V2)

### Infrastructure
1. **Cache shardé avec TTL** : 16 shards, 5 minutes TTL, réduit contention
2. **Worker pools** : 4 workers pour traitement parallèle des exports
3. **Object pooling** : Réutilisation d'objets pour réduire allocations

### Algorithmes
4. **SQL optimisé** : JOINs, GROUP BY, agrégations côté DB
5. **Goroutines parallèles** : 5 queries SQL en parallèle pour stats
6. **Sort optimisé** : `sort.Slice` O(n log n) au lieu de bubble sort O(n²)

### Mémoire
7. **Batch processing** : Traitement par lots de 1000 rows
8. **Streaming** : Export par chunks au lieu de tout charger
9. **Préallocation** : Buffers pré-alloués (1 MB pour CSV)

## 🐌 Non-Optimisations Conservées (V1)

Pour démontrer l'impact des optimisations, V1 conserve volontairement :

1. **N+1 queries problem** : Une query par produit distinct
2. **Bubble sort O(n²)** : Tri inefficace des produits
3. **Pas de cache** : Recalcul à chaque requête
4. **Chargement en mémoire** : Toutes les données chargées d'un coup
5. **Boucles imbriquées** : Calculs inefficaces

## 🛠️ Outils Utilisés

- **Go 1.25** : Langage de programmation
- **PostgreSQL** : Base de données avec indexes optimisés
- **Docker Compose** : Orchestration PostgreSQL + pgAdmin
- **Hyperfine** : Benchmarking CLI (statistiques fiables avec warmup)
- **pprof** : Profiling CPU et mémoire (analyse des hot spots)
- **Postman** : Tests manuels de l'API

## 📚 Documentation

- **[Go Benchmarks Guide](docs/BENCHMARKS.md)** - Guide complet des benchmarks Go (NOUVEAU)
- [Benchmarking Guide](docs/BENCHMARK.md) - Guide d'utilisation des benchmarks Hyperfine
- [Performance Results](docs/RESULTS.md) - Résultats de performance détaillés
- [Optimisations](docs/OPTIMISATIONS.md) - Détails des optimisations implémentées
- [Profiling Guide](profiling/PROFILING.md) - Guide d'utilisation pprof
- [Profiling Analysis](profiling/PROFILING_RESULTS.md) - Analyse complète CPU/mémoire

## 🎓 Concepts Démontrés

### Architecture
- Domain-Driven Design (DDD)
- Bounded Contexts
- CQRS (Command Query Responsibility Segregation)
- Dependency Injection
- Layered Architecture (Domain, Application, Infrastructure)

### Performance
- Caching strategies (TTL, sharding)
- Algorithmic optimization (O(n²) → O(n log n))
- Database optimization (indexes, JOINs, GROUP BY)
- Concurrent programming (goroutines, channels)
- Memory optimization (pooling, streaming, batching)

### Best Practices Go
- Value Objects avec validation
- Repository pattern
- Interface-based design
- Error handling
- Testing & benchmarking

## 📝 Licence

Projet éducatif d'évaluation de performance et d'architecture DDD.
