# Projet Eval - Performance Optimization API

Projet d'évaluation et d'optimisation de performance en Go, comparant deux versions d'une API de ventes.

## Structure du Projet

```
eval/
├── v1/                  # API V1 - Version non optimisée
│   ├── handlers.go     # Handlers avec performances volontairement dégradées
│   └── handlers_test.go # Tests et benchmarks V1
│
├── v2/                  # API V2 - Version optimisée
│   ├── handlers.go     # Handlers optimisés (cache, algorithmes efficaces)
│   └── handlers_test.go # Tests et benchmarks V2
│
├── benchmarks/          # Benchmarking avec Hyperfine
│   ├── scripts/        # Scripts de benchmark (PowerShell et Bash)
│   │   ├── benchmark.ps1         # Script auto avec start/stop serveur
│   │   ├── benchmark-simple.ps1  # Script manuel (serveur déjà lancé)
│   │   └── benchmark.sh          # Version Bash
│   └── results/        # Résultats des benchmarks
│       ├── benchmark_*.md        # Résultats individuels
│       ├── benchmark_tests.txt   # Sortie complète des tests Go
│       └── benchmark_stats_365.json
│
├── profiling/           # Profiling avec pprof
│   ├── scripts/        # Scripts de profiling
│   │   ├── profile.ps1           # Script PowerShell
│   │   └── profile.sh            # Script Bash
│   ├── profiles/       # Profils générés (.prof)
│   │   ├── cpu_profile.prof
│   │   └── mem_profile.prof
│   ├── PROFILING.md              # Guide d'utilisation pprof
│   └── PROFILING_RESULTS.md      # Analyse détaillée des résultats
│
├── docs/                # Documentation
│   ├── BENCHMARK.md              # Documentation des benchmarks
│   ├── RESULTS.md                # Résultats de performance
│   └── README_API.md             # Documentation de l'API
│
├── postman/             # Collections Postman
│   └── API_Ventes.postman_collection.json
│
├── main.go              # Point d'entrée avec routes V1 et V2
└── go.mod               # Dépendances Go

```

## Endpoints API

### V1 (Non optimisée)
- `GET /api/v1/stats?days=365` - Statistiques JSON
- `GET /api/v1/export/csv?days=365` - Export CSV complet
- `GET /api/v1/export/stats-csv?days=365` - Export CSV statistiques

### V2 (Optimisée)
- `GET /api/v2/stats?days=365` - Statistiques JSON (avec cache)
- `GET /api/v2/export/csv?days=365` - Export CSV complet (optimisé)
- `GET /api/v2/export/stats-csv?days=365` - Export CSV statistiques (optimisé)

## Démarrage Rapide

### Lancer le serveur
```bash
go run main.go
# Serveur disponible sur http://localhost:8080
```

### Lancer les tests
```bash
go test ./...
```

### Benchmarks Go
```bash
go test -bench=. ./v1 ./v2 -benchmem
```

### Benchmarks Hyperfine
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

## Résultats de Performance

### Comparaison V1 vs V2

| Opération | V1 | V2 | Amélioration |
|-----------|----|----|--------------|
| Stats 365 jours | 49.1 ms | 22.4 ms | **2.19x plus rapide** |
| Stats 100 jours | 31.4 ms | 23.3 ms | **1.35x plus rapide** |
| Export CSV 30 jours | 2072 ms | 40.3 ms | **51.48x plus rapide** |

### Mémoire

| Métrique | V1 | V2 | Amélioration |
|----------|----|----|--------------|
| Génération données | 345.70 MB | 137.15 MB | **60% moins de mémoire** |

Voir [docs/RESULTS.md](docs/RESULTS.md) pour les résultats détaillés et [profiling/PROFILING_RESULTS.md](profiling/PROFILING_RESULTS.md) pour l'analyse approfondie.

## Optimisations Implémentées (V2)

1. **Cache avec TTL** : Données et statistiques cachées pendant 5 minutes
2. **Algorithmes efficaces** : sort.Slice (O(n log n)) au lieu de bubble sort (O(n²))
3. **Préallocation mémoire** : Slices pré-alloués avec capacité estimée
4. **Réduction I/O** : Suppression des prints dans le code de production
5. **Calculs optimisés** : Une seule boucle au lieu de boucles imbriquées

## Outils Utilisés

- **Go 1.x** : Langage de programmation
- **Hyperfine** : Benchmarking CLI (statistiques fiables avec warmup)
- **pprof** : Profiling CPU et mémoire (analyse des hot spots)
- **Postman** : Tests manuels de l'API

## Documentation

- [API Documentation](docs/README_API.md) - Guide complet de l'API
- [Benchmarking Guide](docs/BENCHMARK.md) - Guide d'utilisation des benchmarks
- [Profiling Guide](profiling/PROFILING.md) - Guide d'utilisation pprof
- [Performance Results](docs/RESULTS.md) - Résultats de performance détaillés
- [Profiling Analysis](profiling/PROFILING_RESULTS.md) - Analyse complète CPU/mémoire

## Licence

Projet éducatif d'évaluation de performance.
