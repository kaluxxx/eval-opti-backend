# Guide de Profiling

Ce guide explique comment profiler l'application pour identifier les bottlenecks de performance.

## Table des matières

1. [Profiling en temps réel avec pprof](#profiling-en-temps-réel-avec-pprof)
2. [Benchmarks](#benchmarks)
3. [Scripts automatisés](#scripts-automatisés)
4. [Analyse des résultats](#analyse-des-résultats)

---

## Profiling en temps réel avec pprof

L'application expose automatiquement des endpoints pprof lorsqu'elle tourne.

### Démarrer le serveur

```bash
go run main.go
```

### Endpoints pprof disponibles

Une fois le serveur démarré sur `http://localhost:8080`, les endpoints suivants sont disponibles :

| Endpoint | Description |
|----------|-------------|
| `/debug/pprof/` | Index du profiler avec tous les profils disponibles |
| `/debug/pprof/profile?seconds=30` | Profil CPU pendant 30 secondes |
| `/debug/pprof/heap` | Profil d'allocations mémoire |
| `/debug/pprof/goroutine` | État de toutes les goroutines |
| `/debug/pprof/block` | Profil des blocages |
| `/debug/pprof/mutex` | Profil de contention des mutex |

### Capture manuelle des profils

#### Profil CPU

```bash
# Capture pendant 30 secondes
curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof

# Analyse en ligne de commande
go tool pprof -top cpu.prof

# Analyse interactive
go tool pprof cpu.prof

# Visualisation web (recommandé)
go tool pprof -http=:8081 cpu.prof
```

#### Profil Mémoire

```bash
# Capture
curl http://localhost:8080/debug/pprof/heap -o mem.prof

# Analyse
go tool pprof -top mem.prof

# Visualisation web
go tool pprof -http=:8081 mem.prof
```

### Commandes pprof utiles

En mode interactif (`go tool pprof <fichier>`), vous pouvez utiliser :

```
top           # Top 10 des fonctions les plus coûteuses
top -cum      # Top 10 par temps cumulé
list <func>   # Code source annoté d'une fonction
web           # Graphe visuel (nécessite Graphviz)
pdf           # Génère un PDF du graphe
help          # Aide complète
```

---

## Benchmarks

Des benchmarks ont été créés pour tester les fonctions critiques.

### Exécuter les benchmarks

```bash
# Tous les benchmarks
go test -bench=. -benchmem

# Benchmark spécifique
go test -bench=BenchmarkCalculateStatistics -benchmem

# Avec profiling CPU et mémoire
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
```

### Benchmarks disponibles

**Avec PostgreSQL et données réelles (110k commandes, 330k lignes)** :

| Benchmark | Description | V1 vs V2 |
|-----------|-------------|----------|
| **Stats 365 jours** | Calcul statistiques 1 an | N+1 (200+ queries) vs JOINs (5 queries) |
| **Stats 100 jours** | Calcul statistiques 100j | Bubble sort vs SQL ORDER BY |
| **Export CSV 30 jours** | Export ~27k lignes | N+1 per row vs Single JOIN |
| **Export Parquet 7 jours** | Export format columnar | Full memory vs Streaming |
| **Cache Effect** | Impact du cache V2 | N/A vs 5min TTL cache |

### Interpréter les résultats

```
BenchmarkCalculateStatistics_MediumDataset-8    10    115234567 ns/op    2345678 B/op    12345 allocs/op
```

- `10` : nombre d'itérations
- `115234567 ns/op` : temps moyen par opération (115ms)
- `2345678 B/op` : mémoire allouée par opération
- `12345 allocs/op` : nombre d'allocations par opération

---

## Scripts automatisés

Deux scripts sont fournis pour automatiser le profiling.

### PowerShell (Windows)

```powershell
# Profiling complet (CPU + Mémoire)
.\profile.ps1 all

# CPU uniquement
.\profile.ps1 cpu

# Mémoire uniquement
.\profile.ps1 mem

# Benchmarks
.\profile.ps1 bench

# Avec durée personnalisée (60 secondes)
.\profile.ps1 cpu -Duration 60
```

### Bash (Linux/Mac/WSL)

```bash
# Rendre le script exécutable
chmod +x profile.sh

# Profiling complet
./profile.sh all

# CPU uniquement
./profile.sh cpu

# Mémoire uniquement
./profile.sh mem

# Benchmarks
./profile.sh bench

# Avec durée personnalisée (60 secondes)
./profile.sh cpu 60
```

Les profils sont sauvegardés dans le dossier `profiles/` avec un timestamp.

---

## Analyse des résultats

### 1. Visualisation web (recommandé)

C'est la méthode la plus intuitive :

```bash
go tool pprof -http=:8081 profiles/cpu_20250127_143022.prof
```

Ouvre un navigateur avec :
- **Graph** : Graphe des appels avec temps/mémoire
- **Flame Graph** : Graphe en flamme (très visuel)
- **Top** : Tableau des fonctions les plus coûteuses
- **Source** : Code source annoté

### 2. Ligne de commande

```bash
# Top 10 des fonctions
go tool pprof -top profiles/cpu_20250127_143022.prof

# Top 20
go tool pprof -top -nodecount=20 profiles/cpu_20250127_143022.prof

# Focus sur une fonction spécifique
go tool pprof -top -focus=calculateStatistics profiles/cpu_20250127_143022.prof
```

### 3. Comparaison de profils

Comparez les performances avant/après optimisation :

```bash
# Comparer deux profils CPU
go tool pprof -base=profiles/cpu_before.prof profiles/cpu_after.prof

# Visualiser la différence
go tool pprof -http=:8081 -base=profiles/cpu_before.prof profiles/cpu_after.prof
```

### 4. Indicateurs clés à surveiller

#### CPU Profile
- **Flat %** : Temps passé directement dans la fonction
- **Cum %** : Temps cumulé (fonction + ses appels)
- Cherchez les fonctions avec un **Flat %** élevé

#### Memory Profile
- **inuse_space** : Mémoire actuellement utilisée
- **alloc_space** : Mémoire totale allouée
- **inuse_objects** : Nombre d'objets en mémoire
- **alloc_objects** : Nombre total d'allocations

```bash
# Profil mémoire par espace alloué
go tool pprof -alloc_space profiles/mem.prof

# Profil mémoire par nombre d'allocations
go tool pprof -alloc_objects profiles/mem.prof
```

---

## Exemples de workflow

### Workflow 1 : Identifier un bottleneck CPU

```bash
# 1. Démarrer le serveur
go run main.go

# 2. Dans un autre terminal, capturer un profil CPU
.\profile.ps1 cpu -Duration 30

# 3. Pendant la capture, générer de la charge
curl "http://localhost:8080/api/stats?days=365"

# 4. Analyser le profil
go tool pprof -http=:8081 profiles/cpu_<timestamp>.prof

# 5. Identifier les fonctions gourmandes et optimiser
```

### Workflow 2 : Comparer avant/après optimisation

```bash
# 1. Capturer un profil AVANT optimisation
.\profile.ps1 all
# Sauvegarder les fichiers: cpu_before.prof, mem_before.prof

# 2. Optimiser le code

# 3. Capturer un profil APRÈS optimisation
.\profile.ps1 all

# 4. Comparer les profils
go tool pprof -http=:8081 -base=profiles/cpu_before.prof profiles/cpu_after.prof
```

### Workflow 3 : Benchmarking

```bash
# 1. Exécuter les benchmarks AVANT optimisation
go test -bench=. -benchmem | tee bench_before.txt

# 2. Optimiser le code

# 3. Exécuter les benchmarks APRÈS
go test -bench=. -benchmem | tee bench_after.txt

# 4. Comparer les résultats
diff bench_before.txt bench_after.txt
```

---

## Bottlenecks connus dans ce code

Voici les problèmes de performance identifiés dans l'application actuelle :

### 1. Bubble Sort (O(n²))
**Fichier** : `main.go:145-152`

Le tri bubble sort est utilisé pour le top des produits. C'est le pire algorithme de tri possible.

**Impact** : Très élevé avec beaucoup de produits.

**Solution** : Utiliser `sort.Slice()` de la bibliothèque standard.

### 2. Boucles multiples sur les ventes
**Fichier** : `main.go:114-131`

Le calcul par catégorie boucle sur TOUTES les ventes pour CHAQUE catégorie (5 fois).

**Impact** : Complexité O(n × c) où c = nombre de catégories.

**Solution** : Une seule boucle avec accumulation dans une map.

### 3. Génération de données à chaque requête
**Fichier** : `main.go:177, 249, 331`

Les données sont regénérées pour chaque requête, pas de cache.

**Impact** : Très élevé, génération lente.

**Solution** : Implémenter un cache avec expiration.

### 4. Allocations mémoire excessive
**Fichier** : `main.go:71`

`sales := []Sale{}` crée un slice sans capacité initiale, causant de multiples réallocations.

**Impact** : Allocations et copies inutiles.

**Solution** : Préallouer avec `make([]Sale, 0, expectedSize)`.

### 5. Sleep artificiel
**Fichier** : `main.go:210, 226, 312`

Des `time.Sleep()` sont ajoutés pour simuler des traitements lents.

**Impact** : Ralentissement artificiel.

**Solution** : Supprimer ou rendre conditionnel (dev uniquement).

---

## Ressources

- [Documentation officielle pprof](https://pkg.go.dev/net/http/pprof)
- [Go Blog : Profiling Go Programs](https://go.dev/blog/pprof)
- [Effective Go : Profiling](https://go.dev/doc/effective_go#profiling)
- [Dave Cheney's High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)