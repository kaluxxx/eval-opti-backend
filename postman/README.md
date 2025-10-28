# 📮 Collection Postman - Eval API

Collection complète pour tester et comparer les performances entre **V1 (non optimisé)** et **V2 (optimisé)**.

## 📦 Fichiers

- `eval-api.postman_collection.json` - Collection complète avec tous les endpoints
- `eval-local.postman_environment.json` - Environnement local (localhost:8080)

## 🚀 Installation

### 1. Importer dans Postman

#### Option A : Via l'interface
1. Ouvrir Postman
2. Cliquer sur **Import** (en haut à gauche)
3. Glisser-déposer les 2 fichiers JSON ou cliquer **Upload Files**
4. Sélectionner les 2 fichiers :
   - `eval-api.postman_collection.json`
   - `eval-local.postman_environment.json`

#### Option B : Via URL (si dépôt Git)
```
https://raw.githubusercontent.com/YOUR_REPO/postman/eval-api.postman_collection.json
```

### 2. Sélectionner l'environnement

1. En haut à droite, dans le sélecteur d'environnement
2. Choisir **"Eval - Local"**
3. Vérifier que `baseUrl` = `http://localhost:8080`

### 3. Vérifier que le serveur tourne

```bash
# Dans le terminal
go run main.go

# Ou
./eval.exe
```

Le serveur doit afficher :
```
✅ Connexion PostgreSQL établie
🚀 Serveur démarré sur le port 8080
```

## 📂 Structure de la collection

### 1. **Health Check**
- `GET /api/health` - Vérifier que l'API fonctionne

### 2. **V1 - Non Optimisé** 🐌
- `GET /api/v1/stats?days=365` - Calcul stats (N+1 problem)
- `GET /api/v1/stats?days=30` - Stats 30 jours
- `GET /api/v1/stats?days=7` - Stats 7 jours
- `GET /api/v1/export/csv` - Export CSV (N+1)
- `GET /api/v1/export/stats-csv` - Export stats CSV
- `GET /api/v1/export/parquet` - Export Parquet (⚠️ mémoire intensive)

### 3. **V2 - Optimisé** ⚡
- `GET /api/v2/stats?days=365` - Calcul stats optimisé
- `GET /api/v2/stats?days=365` (2ème appel) - Avec cache
- `GET /api/v2/stats?days=30` - Stats 30 jours optimisé
- `GET /api/v2/stats?days=7` - Stats 7 jours optimisé
- `GET /api/v2/export/csv` - Export CSV optimisé
- `GET /api/v2/export/stats-csv` - Export stats CSV optimisé
- `GET /api/v2/export/parquet` - Export Parquet streaming

### 4. **Tests de Comparaison** 🔬
Séquences prédéfinies pour comparer V1 vs V2 :
- **Comparaison Stats** : V1 → V2 → V2 (cache)
- **Comparaison Export CSV** : V1 → V2
- **Comparaison Parquet** : V1 (days=30) → V2 (days=365)

## 🧪 Scénarios de test recommandés

### Test 1 : Différence de performance Stats

**Objectif** : Comparer le temps de calcul des statistiques

1. Exécuter `V1 - Stats 365 jours`
   - Observer le temps de réponse : ~10-15 secondes
   - Observer les logs console : N+1 queries, boucles multiples

2. Exécuter `V2 - Stats 365 jours`
   - Observer le temps : ~1-2 secondes
   - Observer les logs : 5 requêtes avec JOINs

3. Réexécuter `V2 - Stats 365 jours` immédiatement
   - Observer le temps : < 5 ms (CACHE !)
   - Pas de requête SQL

**Résultat attendu** :
- V2 est **80-90% plus rapide** que V1
- V2 avec cache est **99.9% plus rapide** que V1

### Test 2 : Impact mémoire Parquet

**Objectif** : Démontrer la différence de consommation mémoire

⚠️ **Attention** : V1 peut consommer beaucoup de RAM

1. Exécuter `V1 - Export Parquet` avec `days=30`
   - Observer les logs : chargement complet en mémoire
   - Observer l'estimation mémoire : ~500 MB - 1 GB

2. Exécuter `V2 - Export Parquet` avec `days=365`
   - Observer les logs : traitement par batches de 1000
   - Mémoire constante : ~0.2 MB

**Résultat attendu** :
- V1 utilise **99.99% plus de mémoire** que V2
- V2 est scalable (peut gérer millions de lignes)

### Test 3 : Cache applicatif

**Objectif** : Démontrer l'impact du cache

1. Attendre 5 minutes (expiration du cache)

2. Exécuter `V2 - Stats 365 jours`
   - 1ère exécution : ~1-2 secondes (MISS)

3. Immédiatement réexécuter `V2 - Stats 365 jours`
   - 2ème exécution : < 5 ms (HIT)

4. Attendre 5 minutes puis réexécuter
   - Cache expiré : retour à ~1-2 secondes

**Résultat attendu** :
- Cache valide = réponse **instantanée**
- TTL de 5 minutes fonctionne

### Test 4 : Export CSV comparé

**Objectif** : Comparer les exports

1. Exécuter `V1 - Export CSV` avec `days=30`
   - Temps : ~5-10 secondes
   - Observer les logs : N+1 queries

2. Exécuter `V2 - Export CSV` avec `days=30`
   - Temps : ~1-2 secondes
   - Observer les logs : 1 requête avec JOINs

**Résultat attendu** :
- V2 est **75-85% plus rapide**
- V2 inclut plus d'informations (client complet, ville magasin)

## 📊 Métriques à observer

### Dans Postman
- **Time** (en bas à droite de la réponse)
  - V1 Stats : 10-15 secondes
  - V2 Stats : 1-2 secondes
  - V2 Stats (cache) : < 5 ms

- **Size** (taille de la réponse)
  - Similaire pour V1 et V2
  - V2 contient données bonus (magasins, paiements)

### Dans la console du serveur

**V1 (logs détaillés)** :
```
[V1] 🐌 === DÉBUT CALCUL STATS (NON OPTIMISÉ - N+1) ===
[V1] ⏳ Chargement des order_items...
[V1] 📦 330191 lignes de commande chargées en 2.5s
[V1] 🐌 Récupération des produits (N+1 problem)...
[V1] 📦 100 produits récupérés en 1.8s
[V1]    Boucle 1: Calcul CA total...
[V1]    Boucles multiples: Stats par catégorie...
[V1]       Calcul pour catégorie 'Électronique'
...
[V1]    🐌 Tri avec bubble sort O(n²)...
[V1] 🏁 Durée totale: 12.4s
```

**V2 (logs optimisés)** :
```
[V2] ⚡ === DÉBUT CALCUL STATS (OPTIMISÉ - JOINS) ===
[V2] 💾 Cache miss, calcul des stats...
[V2]    Requête 1: Stats globales...
[V2]    Requête 2: Stats par catégorie (avec JOINs)...
[V2]    Requête 3: Top 10 produits (avec JOINs + ORDER BY + LIMIT)...
[V2]    Requête 4: Top 5 magasins...
[V2]    Requête 5: Répartition paiements...
[V2] ⚡ Stats calculées en 1.2s
```

**V2 avec cache** :
```
[V2] ⚡ === DÉBUT CALCUL STATS (OPTIMISÉ - JOINS) ===
[V2] 🚀 Stats depuis le cache en 2ms
[V2] === FIN (CACHE HIT) ===
```

## 🎯 Résultats attendus

| Métrique | V1 | V2 | V2 (cache) | Amélioration |
|----------|----|----|------------|--------------|
| **Temps Stats** | 10-15s | 1-2s | < 5ms | 80-99.9% ↓ |
| **Requêtes SQL** | 200+ | 5 | 0 | 97-100% ↓ |
| **Mémoire Export** | 2-5 GB | 0.2 MB | - | 99.99% ↓ |
| **Temps Export CSV** | 20-40s | 3-8s | - | 75-85% ↓ |

## 🔧 Variables d'environnement

Dans l'environnement **"Eval - Local"** :

| Variable | Valeur | Description |
|----------|--------|-------------|
| `baseUrl` | `http://localhost:8080` | URL du serveur |
| `days_short` | `7` | Période courte |
| `days_medium` | `30` | Période moyenne |
| `days_long` | `365` | Période longue |

Tu peux utiliser ces variables dans les requêtes :
```
{{baseUrl}}/api/v1/stats?days={{days_medium}}
```

## 🐛 Troubleshooting

### Erreur de connexion
```
Error: connect ECONNREFUSED 127.0.0.1:8080
```
**Solution** : Vérifier que le serveur tourne (`go run main.go`)

### Timeout sur V1
```
Error: Request timeout
```
**Solution** : C'est normal pour V1 avec beaucoup de données. Réduire `days` ou utiliser V2.

### Cache ne fonctionne pas
**Solution** : Le cache a un TTL de 5 minutes. Attendre moins de 5 min entre les appels.

### V1 Parquet crash
```
Error: out of memory
```
**Solution** : V1 Parquet peut consommer beaucoup de RAM. Utiliser `days=7` ou `days=30` maximum pour V1.

## 📚 Documentation complète

Voir `docs/OPTIMISATIONS.md` pour :
- Détails des anti-patterns V1
- Détails des optimisations V2
- Comparaison ligne par ligne du code
- Architecture de la base de données

## 🎓 Pour l'évaluation

### Démonstration recommandée

1. **Ouvrir 2 onglets Postman côte à côte**
   - Gauche : V1
   - Droite : V2

2. **Exécuter simultanément** `Stats 365 jours`
   - Observer la différence de temps
   - Observer les logs console en parallèle

3. **Montrer le cache**
   - Réexécuter V2 Stats → instantané

4. **Montrer Parquet**
   - V1 avec days=30 → mémoire intensive
   - V2 avec days=365 → streaming efficace

5. **Expliquer les optimisations**
   - N+1 → JOINs SQL
   - Boucles multiples → Agrégations SQL
   - Bubble sort → ORDER BY SQL
   - Pas de cache → Cache 5 min
   - Chargement complet → Streaming

---

**Questions ?** Voir `docs/OPTIMISATIONS.md` ou les descriptions détaillées dans chaque requête Postman.
