# Fichiers Générés par les Outils de Mesure

Ce document explique tous les fichiers générés par les différents outils de mesure de performance.

---

## 1. Hyperfine (Benchmark HTTP)

**Outil** : Benchmark d'endpoints HTTP avec statistiques
**Script** : `benchmarks/scripts/benchmark-simple.ps1`
**Localisation** : `benchmarks/results/`

### Fichiers générés

| Fichier | Format | Contenu |
|---------|--------|---------|
| `benchmark_stats_365.md` | Markdown | Benchmark stats JSON avec 365 jours de données |
| `benchmark_stats_100.md` | Markdown | Benchmark stats JSON avec 100 jours de données |
| `benchmark_csv_30.md` | Markdown | Benchmark export CSV avec 30 jours |
| `benchmark_cache.md` | Markdown | Benchmark effet du cache V2 (50 runs) |

### Structure d'un fichier Hyperfine (.md)

```markdown
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `curl -s http://localhost:8080/api/v1/stats?days=365` | 49.1 ± 2.3 | 45.2 | 53.7 | 2.19 ± 0.15 |
| `curl -s http://localhost:8080/api/v2/stats?days=365` | 22.4 ± 1.1 | 20.8 | 25.1 | 1.00 |
```

**Colonnes** :
- **Mean** : Temps moyen d'exécution
- **Min/Max** : Temps minimum et maximum observés
- **Relative** : Facteur de vitesse relatif (1.00 = référence la plus rapide)

### Options Hyperfine utilisées

```powershell
hyperfine.exe `
    --warmup 2 `           # 2 exécutions de chauffe (ignore du cache OS)
    --runs 10 `            # 10 exécutions pour la moyenne
    --export-markdown file.md `  # Export en tableau markdown
    "commande1" `
    "commande2"
```

---

## 2. pprof (Profiling Go)

**Outil** : Profiler CPU et mémoire de Go
**Script** : `profiling/scripts/profile.ps1` ou `profile.sh`
**Localisation** : `profiling/profiles/`

### Fichiers générés

#### Profils CPU

| Fichier | Format | Contenu |
|---------|--------|---------|
| `cpu_YYYYMMDD_HHMMSS.prof` | Binaire pprof | Profil CPU (échantillonnage pendant N secondes) |
| `cpu_profile.prof` | Binaire pprof | Dernier profil CPU (lien/copie) |

**Contenu** :
- Échantillons CPU toutes les 10ms
- Call stacks (pile d'appels) pour identifier les hot spots
- Temps cumulatif par fonction
- Arbre d'appels (qui appelle qui)

#### Profils Mémoire

| Fichier | Format | Contenu |
|---------|--------|---------|
| `mem_YYYYMMDD_HHMMSS.prof` | Binaire pprof | Profil mémoire (heap allocations) |
| `mem_profile.prof` | Binaire pprof | Dernier profil mémoire (lien/copie) |

**Contenu** :
- Allocations mémoire (alloc_space)
- Mémoire en cours d'utilisation (inuse_space)
- Nombre d'allocations (alloc_objects)
- Call stacks d'allocations

### Comment pprof collecte les données

#### CPU Profile
```
/debug/pprof/profile?seconds=30
```
- Échantillonne le CPU toutes les 10ms pendant 30 secondes
- Enregistre la stack trace à chaque échantillon
- Génère un fichier `.prof` binaire

#### Memory Profile
```
/debug/pprof/heap
```
- Snapshot instantané des allocations mémoire
- Échantillonne les allocations (pas toutes, mais représentatif)
- Montre qui alloue quoi et où

### Visualisation des profils

#### 1. Mode texte (Top)
```bash
go tool pprof -top cpu_profile.prof
```
Affiche :
```
flat  flat%   sum%   cum   cum%
0.45s 24.73% 24.73% 0.50s 27.47%  fmt.Sprintf
0.34s 18.68% 43.41% 0.40s 21.98%  time.Time.Format
0.29s 15.93% 59.34% 0.35s 19.23%  runtime.gcDrain
```
- **flat** : Temps passé dans cette fonction uniquement
- **cum** : Temps cumulatif (fonction + appelées)

#### 2. Mode interactif
```bash
go tool pprof cpu_profile.prof
```
Commandes disponibles :
- `top` - Top fonctions
- `list nomFonction` - Code source annoté
- `web` - Graphe (nécessite Graphviz)

#### 3. Mode Web UI (recommandé)
```bash
go tool pprof -http=:8081 cpu_profile.prof
```
Interface web avec :
- **Top** : Tableau des fonctions
- **Graph** : Graphe de flamme interactif
- **Flame Graph** : Visualisation en flamme
- **Peek** : Code source annoté
- **Source** : Temps par ligne de code

---

## 3. go test -bench (Benchmarks Go)

**Outil** : Framework de benchmark intégré à Go
**Commande** : `go test -bench=. -benchmem ./v1 ./v2`
**Localisation** : Sortie console (peut être redirigée)

### Fichiers générés (optionnels)

| Fichier | Format | Contenu |
|---------|--------|---------|
| `benchmark_tests.txt` | Texte | Sortie complète des benchmarks Go |
| `bench_cpu_YYYYMMDD.prof` | Binaire pprof | Profil CPU des benchmarks |
| `bench_mem_YYYYMMDD.prof` | Binaire pprof | Profil mémoire des benchmarks |

### Structure de la sortie benchmark

```
BenchmarkGenerateFakeSalesData_365Days-8    100   12345678 ns/op   1234567 B/op   12345 allocs/op
```

**Colonnes** :
- `BenchmarkGenerateFakeSalesData_365Days` : Nom du benchmark
- `-8` : Nombre de CPUs utilisés (GOMAXPROCS)
- `100` : Nombre d'itérations (ajusté automatiquement pour ~1 seconde)
- `12345678 ns/op` : Temps moyen par opération (nanosecondes)
- `1234567 B/op` : Bytes alloués par opération
- `12345 allocs/op` : Nombre d'allocations par opération

### Options go test -bench

```bash
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
```

**Flags** :
- `-bench=.` : Lance tous les benchmarks (regex)
- `-benchmem` : Affiche les stats mémoire
- `-benchtime=10s` : Durée minimale de chaque benchmark
- `-count=5` : Répète chaque benchmark 5 fois
- `-cpuprofile=file.prof` : Génère un profil CPU
- `-memprofile=file.prof` : Génère un profil mémoire

---

## 4. Fichiers de Documentation (Manuels)

**Générés par** : Analyse manuelle des profils
**Localisation** : `profiling/` et `docs/`

| Fichier | Contenu |
|---------|---------|
| `profiling/PROFILING_RESULTS.md` | Analyse détaillée des profils pprof |
| `docs/BENCHMARK.md` | Documentation des benchmarks Hyperfine |
| `docs/RESULTS.md` | Résultats consolidés de performance |

Ces fichiers sont écrits manuellement après avoir analysé les profils.

---

## Comparaison des Outils

| Outil | Mesure | Niveau | Format Output | Usage |
|-------|--------|--------|---------------|-------|
| **Hyperfine** | Temps de réponse HTTP | End-to-end (API complète) | Markdown (stats) | Comparaison V1/V2 |
| **pprof CPU** | Temps CPU par fonction | Code Go (fonctions) | Binaire .prof | Identifier hot spots CPU |
| **pprof Memory** | Allocations mémoire | Code Go (fonctions) | Binaire .prof | Identifier fuites/gaspillage |
| **go test -bench** | Performance code Go | Code Go (benchmarks) | Texte (ns/op, B/op) | Micro-benchmarks |

---

## Workflow Typique

### 1. Benchmark End-to-End (Hyperfine)
```powershell
cd benchmarks/scripts
.\benchmark-simple.ps1
```
**Génère** : `benchmarks/results/*.md`
**But** : Comparer V1 vs V2 (temps de réponse total)

### 2. Profiling sous Charge (pprof)
```powershell
cd profiling/scripts
.\profile.ps1 all
```
**Génère** : `profiling/profiles/*.prof`
**But** : Identifier les fonctions lentes et les allocations

### 3. Micro-Benchmarks (go test)
```bash
go test -bench=. -benchmem ./v1 ./v2
```
**Génère** : Sortie console
**But** : Mesurer précisément une fonction spécifique

### 4. Analyse et Documentation
- Analyser les profils avec `go tool pprof -http=:8081`
- Écrire `PROFILING_RESULTS.md` avec les conclusions
- Mettre à jour `RESULTS.md` avec les gains

---

## Exemple Complet : Mesurer une Optimisation

### Avant l'optimisation
```bash
# 1. Benchmark Hyperfine
cd benchmarks/scripts && .\benchmark-simple.ps1
# Résultat : V1 = 49ms, V2 = 22ms

# 2. Profiling pprof
cd profiling/scripts && .\profile.ps1 all
# Résultat : fmt.Sprintf = 24.73% CPU (hot spot)

# 3. Benchmark Go
go test -bench=BenchmarkCalculateStatistics ./v1
# Résultat : 50000000 ns/op, 5000000 B/op
```

### Après optimisation (remplacer fmt.Sprintf par strconv.Itoa)
```bash
# 1. Re-benchmark
cd benchmarks/scripts && .\benchmark-simple.ps1
# Nouveau résultat : V2 = 15ms (gain de 32%)

# 2. Re-profiling
cd profiling/scripts && .\profile.ps1 all
# strconv.Itoa = 8% CPU (au lieu de 24.73%)

# 3. Re-benchmark Go
go test -bench=BenchmarkCalculateStatistics ./v2
# Nouveau : 30000000 ns/op, 3000000 B/op (40% plus rapide)
```

---

## Résumé : Où Chercher Quoi

| Question | Outil | Fichier |
|----------|-------|---------|
| V2 est combien de fois plus rapide que V1 ? | Hyperfine | `benchmarks/results/benchmark_*.md` |
| Quelle fonction prend le plus de temps CPU ? | pprof CPU | `profiling/profiles/cpu_*.prof` |
| Quelle fonction alloue le plus de mémoire ? | pprof Memory | `profiling/profiles/mem_*.prof` |
| Combien d'allocations fait cette fonction ? | go test -bench | Console ou `benchmarks/results/benchmark_tests.txt` |
| Le cache V2 fonctionne-t-il ? | Hyperfine | `benchmarks/results/benchmark_cache.md` |
| Quel est le gain total d'optimisation ? | Documentation | `docs/RESULTS.md` |

---

## Notes Importantes

### .gitignore
```gitignore
# Ignore les fichiers à la racine uniquement
/benchmark_*.md
/benchmark_*.json
/*.prof

# Autorise les résultats dans les dossiers organisés
!benchmarks/results/
!profiling/profiles/
```

### Taille des Fichiers

| Type | Taille Typique | Exemple |
|------|---------------|---------|
| Benchmark .md | 1-5 KB | Petit tableau markdown |
| CPU profile .prof | 100-500 KB | 30s d'échantillons |
| Memory profile .prof | 50-200 KB | Snapshot heap |
| benchmark_tests.txt | 500 KB - 10 MB | Sortie complète des tests |

### Conservation des Fichiers

**À garder** :
- Profils de référence (avant optimisation)
- Profils après optimisation majeure
- Benchmarks finaux (validation)

**À supprimer** :
- Profils intermédiaires pendant le développement
- Benchmarks de tests rapides

---

## Commandes Utiles

### Comparer deux profils CPU
```bash
go tool pprof -base=old_cpu.prof new_cpu.prof
```

### Voir uniquement les allocations (pas le GC)
```bash
go tool pprof -alloc_space mem_profile.prof
```

### Exporter en format texte
```bash
go tool pprof -text cpu_profile.prof > analysis.txt
```

### Générer un flame graph PNG
```bash
go tool pprof -png cpu_profile.prof > flamegraph.png
```
