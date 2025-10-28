# Résultats des Benchmarks

Résultats obtenus avec **hyperfine** sur Windows avec le script `benchmark-simple.ps1`.

## Configuration
- **OS**: Windows
- **Tool**: Hyperfine 1.19.0
- **Server**: Go 1.25
- **Dataset**: Données aléatoires générées

## Résultats Complets

### 1. Stats JSON - 365 jours (~36,000 ventes)

#### V1 - Non Optimisée
```
Time (mean ± σ):      49.1 ms ±   3.2 ms    [User: 8.7 ms, System: 18.1 ms]
Range (min … max):    43.2 ms …  52.3 ms    10 runs
```

#### V2 - Optimisée
```
Time (mean ± σ):      22.4 ms ±   0.9 ms    [User: 6.2 ms, System: 19.6 ms]
Range (min … max):    21.2 ms …  23.9 ms    10 runs
```

**Résultat : V2 est 2.19x plus rapide que V1** ⚡

---

### 2. Stats JSON - 100 jours (~10,000 ventes)

#### V1 - Non Optimisée
```
Time (mean ± σ):      31.4 ms ±   2.6 ms    [User: 5.2 ms, System: 12.8 ms]
Range (min … max):    26.9 ms …  35.2 ms    10 runs
```

#### V2 - Optimisée
```
Time (mean ± σ):      23.3 ms ±   0.7 ms    [User: 7.9 ms, System: 5.0 ms]
Range (min … max):    22.2 ms …  24.5 ms    10 runs
```

**Résultat : V2 est 1.35x plus rapide que V1** ⚡

---

### 3. Export CSV - 30 jours (~3,000 lignes)

#### V1 - Non Optimisée
```
Time (mean ± σ):      2.072 s ±  0.005 s    [User: 0.017 s, System: 0.020 s]
Range (min … max):    2.068 s …  2.082 s    5 runs
```

#### V2 - Optimisée
```
Time (mean ± σ):      40.3 ms ±   1.6 ms    [User: 8.4 ms, System: 37.8 ms]
Range (min … max):    37.7 ms …  42.0 ms    5 runs
```

**Résultat : V2 est 51.48x plus rapide que V1** 🔥🔥🔥

---

### 4. Cache V2 - Stats 365 jours (avec cache)

```
Time (mean ± σ):      24.5 ms ±   3.9 ms    [User: 13.8 ms, System: 12.1 ms]
Range (min … max):    19.0 ms …  41.7 ms    50 runs
```

**Note** : Performance stable avec cache actif (TTL 5 min)

---

## Analyse des Résultats

### Gains de Performance

| Test | V1 | V2 | Amélioration |
|------|----|----|--------------|
| Stats 365j | 49.1 ms | 22.4 ms | **2.19x** |
| Stats 100j | 31.4 ms | 23.3 ms | **1.35x** |
| CSV 30j | 2072 ms | 40.3 ms | **51.48x** |
| Cache V2 | - | 24.5 ms | Stable |

### Pourquoi V2 est Plus Rapide ?

#### 1. Suppression des prints (fmt.Printf/Println)
- **Impact** : Élimine l'overhead des I/O
- **Gain estimé** : 10-15%

#### 2. Pas de sleeps artificiels
- **V1 CSV** : `time.Sleep(10ms)` toutes les 1000 lignes + 2s post-traitement + 1s stats
- **V2 CSV** : Aucun sleep
- **Gain** : **Énorme sur CSV** (51x plus rapide !)

#### 3. Tri efficace
- **V1** : Bubble sort O(n²)
- **V2** : `sort.Slice` O(n log n)
- **Impact** : Plus visible sur gros datasets

#### 4. Calculs optimisés
- **V1** : Boucles imbriquées pour chaque catégorie (O(N×M))
- **V2** : Une seule passe (O(N))
- **Gain** : ~20-30% sur calculs

#### 5. Cache avec TTL
- **V2** : Données mises en cache 5 minutes
- **Impact** : Requêtes suivantes ~constant 24ms

#### 6. Préallocation des slices
- **V2** : `make([]Sale, 0, estimatedSize)`
- **Gain** : Évite réallocations mémoire

---

## Impact des Optimisations par Catégorie

### Stats JSON (Petit Impact des Sleeps)
- V1 : Pas de sleeps majeurs, juste calculs inefficaces
- Gain modéré : **1.35x - 2.19x**

### Export CSV (Gros Impact des Sleeps)
- V1 : 2+ secondes de sleeps artificiels
- Gain massif : **51.48x** 🚀

### Cache V2
- Performance constante ~24ms
- Pas de régénération des données

---

## Recommandations

### Pour Production

1. ✅ **Utiliser V2** : Gains substantiels sans compromis
2. ✅ **Activer le cache** : TTL configurable selon besoin
3. ✅ **Pas de logs verbeux** : Utiliser un logger avec niveaux
4. ✅ **Profiling régulier** : Utiliser pprof pour détecter bottlenecks

### Pour Aller Plus Loin

- **Goroutines** : Paralléliser la génération de données
- **Pool de workers** : Pour les exports CSV volumineux
- **Compression** : gzip pour les CSV avant envoi
- **Pagination** : Limiter les datasets renvoyés
- **Base de données** : Remplacer génération aléatoire par vraies données

---

## Conclusion

Les optimisations de la V2 démontrent qu'avec quelques bonnes pratiques :
- Suppression des I/O inutiles
- Choix d'algorithmes efficaces
- Cache intelligent
- Préallocation mémoire

On peut obtenir des **gains de 1.3x à 50x** selon le contexte !

Le code V1 illustre les **anti-patterns** à éviter en production :
- ❌ Bubble sort pour tri
- ❌ Boucles imbriquées inefficaces
- ❌ Sleeps artificiels
- ❌ Logs verbeux sans contrôle
- ❌ Génération à chaque requête

---

## Reproductibilité

Pour reproduire ces résultats :

```powershell
# Terminal 1 : Lancer le serveur
go run main.go

# Terminal 2 : Lancer le benchmark
.\benchmark-simple.ps1
```

Les résultats détaillés sont dans :
- `benchmark_stats_365.md`
- `benchmark_stats_100.md`
- `benchmark_csv_30.md`
- `benchmark_cache.md`
