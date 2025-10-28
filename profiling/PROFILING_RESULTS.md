# Résultats du Profiling pprof

Profiling réalisé avec `go tool pprof` sur 30 secondes avec charge mixte V1/V2.

## Configuration

- **Duration**: 30 secondes
- **Total samples**: 1.82s (6.07%)
- **Charge**: 15 requêtes parallèles V1 + V2 (stats 365j et 100j)
- **Date**: 2025-10-27

## Profil CPU

### Top Fonctions (cumulative time)

| Fonction | Temps Cumulatif | % | Analyse |
|----------|----------------|---|---------|
| `net/http.(*conn).serve` | 1.30s | 71.43% | Serveur HTTP (normal) |
| **`eval/v1.GetStats`** | 0.66s | 36.26% | **Handler V1** |
| **`eval/v2.GetStats`** | 0.62s | 34.07% | **Handler V2** |
| `eval/v2.getCachedStats` | 0.61s | 33.52% | Cache V2 |
| **`eval/v1.generateFakeSalesData`** | 0.58s | 31.87% | **Génération V1** |
| **`eval/v2.generateFakeSalesData`** | 0.51s | 28.02% | **Génération V2** ✅ |
| **`fmt.Sprintf`** | 0.45s | 24.73% | **Formatage strings** |
| `time.Time.Format` | 0.34s | 18.68% | Formatage dates |
| `runtime.systemstack` | 0.41s | 22.53% | Runtime Go (GC, etc.) |
| `runtime.gcDrain` | 0.29s | 15.93% | Garbage Collector |

### Observations CPU

#### 1. V2 est plus rapide que V1 ✅
- **V1 generateData**: 0.58s (31.87%)
- **V2 generateData**: 0.51s (28.02%)
- **Gain**: 12% plus rapide grâce au cache

#### 2. fmt.Sprintf très coûteux ⚠️
- Prend **24.73%** du temps CPU
- Utilisé pour :
  - `fmt.Sprintf("Produit_%d", ...)` - génération noms produits
  - `fmt.Sprintf("Client_%d", ...)` - génération noms clients
  - Formatage des nombres en CSV

**Optimisation possible** :
```go
// Au lieu de :
Product: fmt.Sprintf("Produit_%d", rand.Intn(100)+1)

// Utiliser :
Product: "Produit_" + strconv.Itoa(rand.Intn(100)+1)
// strconv.Itoa est 2-3x plus rapide
```

#### 3. time.Time.Format coûteux
- **18.68%** du temps CPU
- Utilisé pour formater les dates en "2006-01-02"
- Inévitable, mais pourrait être optimisé avec un cache de dates pré-formatées

#### 4. Garbage Collector actif
- `runtime.gcDrain`: 15.93%
- Beaucoup d'allocations temporaires
- Signe qu'on pourrait réduire les allocations

---

## Profil Mémoire

### Top Allocations (alloc_space)

| Fonction | Allocations | % | Analyse |
|----------|-------------|---|---------|
| **`eval/v1.generateFakeSalesData`** | 345.70 MB | 50.77% | **Énorme V1** ❌ |
| **`eval/v2.generateFakeSalesData`** | 137.15 MB | 20.14% | **V2 optimisé** ✅ |
| `bytes.growSlice` | 89.64 MB | 13.17% | Buffers CSV qui grandissent |
| **`fmt.Sprintf`** | 66.50 MB | 9.77% | **Allocations strings** |
| `time.Time.Format` | 25 MB | 3.67% | Formatage dates |
| `eval/v2.ExportCSV` | 3 MB | 0.44% | Export V2 |
| `eval/v1.ExportCSV` | 1 MB | 0.15% | Export V1 |

### Observations Mémoire

#### 1. V2 utilise 2.5x MOINS de mémoire que V1 🔥
- **V1**: 345.70 MB (50.77%)
- **V2**: 137.15 MB (20.14%)
- **Économie**: **208 MB** grâce au cache !

#### 2. fmt.Sprintf alloue beaucoup
- **66.50 MB** d'allocations (9.77%)
- Chaque `fmt.Sprintf` alloue une nouvelle string
- Optimisation : utiliser `strconv` ou `strings.Builder`

#### 3. bytes.growSlice (buffers CSV)
- **89.64 MB** (13.17%)
- Les buffers CSV grandissent dynamiquement
- Optimisation : préallouer avec une taille estimée
  ```go
  buf := bytes.NewBuffer(make([]byte, 0, estimatedSize))
  ```

#### 4. Export CSV très efficient
- V1 Export: seulement 1 MB
- V2 Export: seulement 3 MB
- Le buffer CSV est bien optimisé ✅

---

## Comparaison V1 vs V2

### CPU

| Métrique | V1 | V2 | Amélioration |
|----------|----|----|--------------|
| **GetStats** | 0.66s (36.26%) | 0.62s (34.07%) | **6% plus rapide** |
| **generateData** | 0.58s (31.87%) | 0.51s (28.02%) | **12% plus rapide** |

### Mémoire

| Métrique | V1 | V2 | Amélioration |
|----------|----|----|--------------|
| **generateData** | 345.70 MB | 137.15 MB | **60% moins de mémoire** 🔥 |
| **Total handler** | 364.74 MB | 169.35 MB | **54% moins de mémoire** |

---

## Hot Spots Identifiés

### 🔴 Hot Spot #1 : fmt.Sprintf (CPU + Mémoire)
- **CPU**: 24.73% du temps
- **Mémoire**: 66.50 MB d'allocations
- **Impact**: ÉNORME

**Solution** :
```go
// Avant (lent)
Product: fmt.Sprintf("Produit_%d", id)
Customer: fmt.Sprintf("Client_%d", id)

// Après (rapide)
Product: "Produit_" + strconv.Itoa(id)
Customer: "Client_" + strconv.Itoa(id)

// Ou encore mieux avec string builder pour concaténations multiples
var sb strings.Builder
sb.WriteString("Produit_")
sb.WriteString(strconv.Itoa(id))
Product: sb.String()
```

### 🟠 Hot Spot #2 : time.Time.Format
- **CPU**: 18.68% du temps
- **Mémoire**: 25 MB d'allocations
- **Impact**: Significatif

**Solution** : Cache de dates pré-formatées
```go
var dateCache = make(map[time.Time]string)

func formatDate(t time.Time) string {
    if cached, ok := dateCache[t]; ok {
        return cached
    }
    formatted := t.Format("2006-01-02")
    dateCache[t] = formatted
    return formatted
}
```

### 🟡 Hot Spot #3 : Allocations V1
- **Mémoire**: 345 MB pour générer les données
- **Impact**: Fort impact sur GC

**Solution** : Déjà implémenté en V2 ✅
- Cache avec TTL
- Préallocation des slices

---

## Optimisations Recommandées

### Court Terme (Quick Wins)

1. ✅ **Cache V2** - Déjà implémenté
   - Économie: 60% mémoire
   - Gain: 12% CPU

2. 🔴 **Remplacer fmt.Sprintf par strconv**
   - Gain attendu: **20-30% CPU**
   - Gain mémoire: **60+ MB**
   - Effort: Faible (chercher/remplacer)

3. 🟠 **Cache des dates formatées**
   - Gain attendu: **15-20% CPU**
   - Gain mémoire: **20+ MB**
   - Effort: Moyen

4. 🟡 **Préallocation buffers CSV**
   - Gain mémoire: **10-20 MB**
   - Effort: Faible
   ```go
   buf := bytes.NewBuffer(make([]byte, 0, len(sales)*100))
   ```

### Moyen Terme

5. **Pool d'objets pour Sales**
   ```go
   var salePool = sync.Pool{
       New: func() interface{} {
           return &Sale{}
       },
   }
   ```

6. **Génération parallèle avec goroutines**
   ```go
   // Générer les données par chunks en parallèle
   numWorkers := runtime.NumCPU()
   chunkSize := days / numWorkers
   ```

7. **Remplacement de encoding/csv**
   - `encoding/csv` est assez lent
   - Alternative: écriture manuelle optimisée

### Long Terme

8. **Base de données réelle**
   - Remplacer génération aléatoire
   - Index sur catégories
   - Requêtes SQL optimisées

9. **Compression gzip**
   - Compresser CSV avant envoi
   - Réduction 70-80% taille

10. **API streaming**
    - Ne pas tout charger en mémoire
    - Streamer les résultats ligne par ligne

---

## Impact Estimé des Optimisations

| Optimisation | CPU | Mémoire | Effort | Priority |
|--------------|-----|---------|--------|----------|
| Remplacer fmt.Sprintf | -25% | -60 MB | Faible | 🔴 Haute |
| Cache dates | -18% | -25 MB | Moyen | 🟠 Haute |
| Préalloc buffers | -2% | -15 MB | Faible | 🟡 Moyenne |
| Pool objets | -5% | -30 MB | Moyen | 🟡 Moyenne |
| Génération // | -30% | 0 | Élevé | 🟢 Basse |
| DB réelle | -50% | -200 MB | Très élevé | 🟢 Basse |

---

## Visualisation Web

Le profil interactif est disponible à l'adresse :
**http://localhost:6060**

### Vues disponibles :

1. **Top** : Tableau des fonctions les plus coûteuses
2. **Graph** : Graphe de flamme (flame graph)
3. **Peek** : Code source annoté
4. **Source** : Code source avec temps CPU par ligne
5. **Disasm** : Assembly annoté

### Comment utiliser :

```bash
# Vue graphique (recommandé)
http://localhost:6060/ui/

# Flame graph
http://localhost:6060/ui/flamegraph

# Top liste
http://localhost:6060/ui/top
```

---

## Conclusion

### Points Positifs ✅

1. **V2 déjà bien optimisé**
   - 60% moins de mémoire que V1
   - 12% plus rapide en CPU
   - Cache très efficace

2. **Export CSV efficient**
   - Utilisation mémoire minimale
   - Pas de hot spot majeur

### Points d'Amélioration 🔴

1. **fmt.Sprintf est le #1 bottleneck**
   - 25% du temps CPU
   - 67 MB de mémoire
   - **Fix simple : utiliser strconv**

2. **time.Time.Format coûteux**
   - 19% du temps CPU
   - Cache de dates améliorerait

3. **GC actif (16%)**
   - Signe de beaucoup d'allocations
   - Les optimisations ci-dessus réduiront la pression sur le GC

### Prochaines Étapes

1. Implémenter remplacement fmt.Sprintf → strconv
2. Ajouter cache des dates formatées
3. Préallouer les buffers CSV
4. Re-profiler pour mesurer l'impact

---

## Fichiers Générés

- `profiles/cpu_profile.prof` - Profil CPU (30s)
- `profiles/mem_profile.prof` - Profil mémoire
- Visualisation web : http://localhost:6060
