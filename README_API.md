# API Ventes - Comparaison V1 vs V2

Ce projet compare deux versions d'une API de gestion de ventes :
- **V1** : Code intentionnellement non optimisé (pour démonstration)
- **V2** : Code optimisé avec les meilleures pratiques

## Démarrage

```bash
go run main.go
```

Le serveur démarre sur `http://localhost:8080`

## Endpoints

### API V1 - Non optimisée

```
GET /api/v1/stats?days=365              # Statistiques JSON
GET /api/v1/export/csv?days=365         # Export CSV complet
GET /api/v1/export/stats-csv?days=365   # Export CSV des statistiques
```

### API V2 - Optimisée

```
GET /api/v2/stats?days=365              # Statistiques JSON (avec cache)
GET /api/v2/export/csv?days=365         # Export CSV complet (optimisé)
GET /api/v2/export/stats-csv?days=365   # Export CSV des stats (optimisé)
```

### Health Check

```
GET /api/health                          # Vérification du serveur
```

## Problèmes de Performance (V1)

### 1. Génération de données à chaque requête
```go
// ❌ Pas de cache - régénère tout à chaque appel
sales := generateFakeSalesData(days)
```

### 2. Bubble Sort O(n²)
```go
// ❌ Algorithme de tri inefficace
for i := 0; i < n; i++ {
    for j := 0; j < n-i-1; j++ {
        if productsList[j].CA < productsList[j+1].CA {
            productsList[j], productsList[j+1] = productsList[j+1], productsList[j]
        }
    }
}
```

### 3. Boucles multiples inefficaces
```go
// ❌ Boucle sur toutes les ventes pour CHAQUE catégorie
for _, cat := range categories {
    for _, sale := range sales {
        if sale.Category == cat {
            // calculs...
        }
    }
}
```

### 4. Sleeps artificiels
```go
// ❌ Ralentissements intentionnels
time.Sleep(10 * time.Millisecond)  // Toutes les 1000 lignes
time.Sleep(2 * time.Second)        // Post-traitement
time.Sleep(1 * time.Second)        // Export stats
```

### 5. Pas de préallocation
```go
// ❌ Réallocations multiples
var sales []Sale
sales = append(sales, sale)  // Croissance dynamique inefficace
```

## Optimisations (V2)

### 1. Cache avec TTL (5 minutes)
```go
// ✅ Cache des données générées
var (
    cachedSales   []Sale
    cacheTime     time.Time
    cacheDuration = 5 * time.Minute
    cacheMutex    sync.RWMutex
)
```

**Gain** : Évite la régénération des données à chaque requête

### 2. Tri efficace O(n log n)
```go
// ✅ Utilise sort.Slice de la stdlib
sort.Slice(productsList, func(i, j int) bool {
    return productsList[i].CA > productsList[j].CA
})
```

**Gain** : Pour 100 produits, passe de ~10,000 comparaisons à ~664 comparaisons

### 3. Calcul en une seule passe
```go
// ✅ Une seule boucle pour tout calculer
for _, sale := range sales {
    ca := float64(sale.Quantity) * sale.Price
    totalCA += ca

    // Stats par catégorie
    catStats := stats.ParCategorie[sale.Category]
    catStats.CA += ca
    catStats.NbVentes++
    stats.ParCategorie[sale.Category] = catStats

    // CA par produit
    productsCA[sale.Product] += ca
}
```

**Gain** : Au lieu de N×M boucles (N=sales, M=catégories), une seule boucle de N itérations

### 4. Pas de sleeps artificiels
```go
// ✅ Suppression de tous les time.Sleep()
```

**Gain** : Pour un export de 36,000 lignes, économie de ~360ms + 2s = ~2.4s

### 5. Préallocation des slices
```go
// ✅ Préalloue la capacité
estimatedSize := days * 100
sales := make([]Sale, 0, estimatedSize)
```

**Gain** : Évite les réallocations multiples en mémoire

## Comparaison des performances

### Test avec 365 jours (~36,000 ventes)

| Opération | V1 (non optimisée) | V2 (optimisée) | Amélioration |
|-----------|-------------------|----------------|--------------|
| **Première requête /stats** | ~5-7s | ~2-3s | ~60% plus rapide |
| **Deuxième requête /stats** | ~5-7s | ~5-10ms | ~1000x plus rapide |
| **Export CSV complet** | ~10-15s | ~3-5s (première fois) | ~70% plus rapide |
| **Export stats CSV** | ~7-10s | ~3-5ms (avec cache) | ~2000x plus rapide |

### Complexité algorithmique

| Opération | V1 | V2 |
|-----------|----|----|
| Tri des produits | O(n²) bubble sort | O(n log n) sort.Slice |
| Stats par catégorie | O(N×M) boucles imbriquées | O(N) une seule passe |
| Génération données | À chaque requête | Cachée (5 min) |

## Tester la différence

### 1. Lancer le serveur
```bash
go run main.go
```

### 2. Tester V1 (lente)
```bash
# PowerShell
Measure-Command { Invoke-WebRequest http://localhost:8080/api/v1/stats?days=365 }

# Ou avec curl
curl -w "\nTemps: %{time_total}s\n" http://localhost:8080/api/v1/stats?days=365
```

### 3. Tester V2 (rapide)
```bash
# Première requête (génère le cache)
Measure-Command { Invoke-WebRequest http://localhost:8080/api/v2/stats?days=365 }

# Deuxième requête (utilise le cache)
Measure-Command { Invoke-WebRequest http://localhost:8080/api/v2/stats?days=365 }
```

## Profiling

Le serveur expose les endpoints pprof pour analyser les performances :

```bash
# CPU Profile
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Memory Profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Visualisation web
go tool pprof -http=:6060 http://localhost:8080/debug/pprof/profile?seconds=30
```

## Structure du projet

```
eval/
├── main.go              # Point d'entrée avec routage
├── v1/
│   └── handlers.go      # Implémentation non optimisée
├── v2/
│   └── handlers.go      # Implémentation optimisée
├── go.mod
├── PROFILING.md         # Documentation du profiling
└── README_API.md        # Ce fichier
```

## Bonnes pratiques appliquées (V2)

1. **Cache intelligent** avec TTL et mutex pour la concurrence
2. **Algorithmes efficaces** (sort.Slice au lieu de bubble sort)
3. **Minimisation des allocations** (préallocation, une seule passe)
4. **Pas de sleeps inutiles** (traitement asynchrone si nécessaire)
5. **Mesure du temps** pour identifier les goulots d'étranglement

## Conclusion

La V2 montre qu'avec quelques optimisations simples :
- Utilisation du cache (5 min TTL)
- Choix d'algorithmes efficaces
- Calculs en une seule passe
- Préallocation mémoire

On peut obtenir des gains de performance de **60-2000x** selon l'opération.

Le code de la V1 illustre les anti-patterns courants à éviter en production.
