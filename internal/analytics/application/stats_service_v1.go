package application

import (
	"eval/internal/analytics/domain"
	"eval/internal/analytics/infrastructure"
	catalogdomain "eval/internal/catalog/domain"
	cataloginfra "eval/internal/catalog/infrastructure"
	shareddomain "eval/internal/shared/domain"
)

// StatsServiceV1 service NON-optimisé pour le calcul des statistiques (Version 1)
// Reproduit volontairement les inefficacités de l'ancienne version
type StatsServiceV1 struct {
	statsRepo   *infrastructure.StatsQueryRepository
	productRepo *cataloginfra.ProductQueryRepository
}

// NewStatsServiceV1 crée une nouvelle instance de StatsServiceV1
func NewStatsServiceV1(
	statsRepo *infrastructure.StatsQueryRepository,
	productRepo *cataloginfra.ProductQueryRepository,
) *StatsServiceV1 {
	return &StatsServiceV1{
		statsRepo:   statsRepo,
		productRepo: productRepo,
	}
}

// GetStats calcule les statistiques de manière inefficace (comme V1 originale)
func (s *StatsServiceV1) GetStats(days int) (*domain.Stats, error) {
	// Créer la période
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Calculer les stats de manière inefficace
	return s.calculateStatsInefficient(dateRange)
}

// calculateStatsInefficient calcule les stats de manière volontairement inefficace
// pour reproduire les problèmes de performance de V1
// SYNTAXE: (s *StatsServiceV1) = méthode receiver avec pointeur (comme "self" en Python)
//   - Le * permet de modifier la struct sans la copier
//
// PERFORMANCE: ⚠️ EXTRÊMEMENT LENT - O(n²) + N+1 queries
func (s *StatsServiceV1) calculateStatsInefficient(dateRange shareddomain.DateRange) (*domain.Stats, error) {
	stats := domain.NewStats()

	// ⚠️ PROBLÈME MAJEUR 1: Charge TOUTES les lignes de commande en mémoire!
	// MÉMOIRE: Si 100k order_items, cela peut faire ~50MB de données en RAM
	//   - Chaque OrderItemData ≈ 80 bytes (int64*7 + float64*2 + string)
	//   - Slice overhead: 24 bytes (pointeur + len + cap)
	//   - Pas de GROUP BY SQL = base de données fait tout le travail puis envoie TOUT
	// PERFORMANCE: I/O réseau important, latence élevée, GC pressure
	allItems, err := s.statsRepo.GetAllOrderItems(dateRange)
	if err != nil {
		return nil, err
	}

	// Calcul 1: Chiffre d'affaires total (boucle simple)
	totalRevenue := 0.0
	// SYNTAXE: make(map[K]V) = alloue un hashmap (dictionnaire) en Go
	//   - map[int64]bool = clés de type int64, valeurs de type bool
	// MÉMOIRE: Map est toujours alloué sur le HEAP
	//   - Structure interne: tableau de buckets + linked lists pour collisions
	//   - Overhead ~48 bytes + ~16 bytes par entrée minimum
	// UTILISATION: bool comme "set" pour dédupliquer les order IDs (true = présent)
	totalOrders := make(map[int64]bool)

	// SYNTAXE: for _, item := range allItems
	//   - range = itère sur la slice
	//   - _ = ignore l'index (underscore = variable jetable en Go)
	//   - item = copie de la valeur à chaque itération
	// PERFORMANCE: O(n) - Complexité linéaire acceptable
	// MÉMOIRE: item est copié sur la STACK à chaque itération (~80 bytes copiés)
	//   - Pourrait être optimisé avec range &allItems[i] (pointeur, pas de copie)
	for _, item := range allItems {
		totalRevenue += item.Subtotal    // addition des montants des produits
		totalOrders[item.OrderID] = true // Marque l'order comme vu (déduplication)
	}

	// Definition de notre statistique
	// SYNTAXE: _ = ignore la valeur d'erreur (dangereux en prod, ok pour démo)
	revenue, _ := shareddomain.NewMoney(totalRevenue, "EUR")
	stats.SetTotalRevenue(revenue)
	stats.SetTotalOrders(len(totalOrders))

	if len(totalOrders) > 0 {
		avgOrder, _ := shareddomain.NewMoney(totalRevenue/float64(len(totalOrders)), "EUR")
		stats.SetAverageOrderValue(avgOrder)
	}

	// ⚠️ PROBLÈME MAJEUR 2: N+1 QUERIES PROBLEM!
	// PERFORMANCE: Si 1000 produits distincts = 1000 requêtes SQL individuelles!
	//   - Chaque requête: latence réseau (~1ms) + parsing SQL + query plan
	//   - Au lieu de 1 requête JOIN, on fait 1 + N requêtes séquentielles
	//   - Total: 1000ms minimum juste pour la latence réseau
	// SYNTAXE: map[int64]*productStatTemp
	//   - Clé: int64 (product ID)
	//   - Valeur: *productStatTemp = POINTEUR vers la struct
	// MÉMOIRE: Pourquoi pointeur? Pour modifier la struct sans la recopier
	//   - Pointeur = 8 bytes, Struct = ~60 bytes
	//   - Si on stockait la struct directement, chaque map[key] créerait une copie
	productStats := make(map[int64]*productStatTemp)
	for _, item := range allItems {
		// SYNTAXE: _, exists := map[key]
		//   - Idiome Go pour tester l'existence d'une clé
		//   - _ = ignore la valeur, exists = bool (true si clé présente)
		if _, exists := productStats[item.ProductID]; !exists {

			// ⚠️ N+1 QUERY: Une requête SQL PAR PRODUIT DISTINCT!
			// PERFORMANCE: Requête synchrone bloquante, latence ~1-5ms par produit
			product, err := s.productRepo.FindByID(catalogdomain.ProductID(item.ProductID))
			if err != nil {
				// Si erreur, on utilise un nom par défaut
				// SYNTAXE: &productStatTemp{} = alloue struct sur HEAP et retourne pointeur
				//   - HEAP car besoin de survivre au-delà de ce scope
				//   - Si on retournait la struct directement, elle serait copiée
				productStats[item.ProductID] = &productStatTemp{
					productID:   item.ProductID,
					productName: "Unknown Product",
					revenue:     0,
					orders:      make(map[int64]bool), // Nouveau map pour ce produit
					quantity:    0,
				}
			} else {
				productStats[item.ProductID] = &productStatTemp{
					productID:   item.ProductID,
					productName: product.Name(),
					revenue:     0,
					orders:      make(map[int64]bool),
					quantity:    0,
				}
			}
		}

		ps := productStats[item.ProductID]
		ps.revenue += item.Subtotal
		ps.orders[item.OrderID] = true
		ps.quantity += item.Quantity
	}

	// Convertir map en slice pour pouvoir trier
	// MÉMOIRE: var déclare sans initialiser (nil slice), capacité = 0
	//   - Chaque append peut déclencher réallocation (doublement de capacité)
	//   - Mieux: productStatsList := make([]*productStatTemp, 0, len(productStats))
	// SYNTAXE: []*productStatTemp = slice de pointeurs vers productStatTemp
	//   - [] = slice (tableau dynamique), * = pointeurs
	var productStatsList []*productStatTemp
	for _, ps := range productStats {
		// PERFORMANCE: append peut réalloquer si capacité insuffisante
		//   - Réallocation = nouvelle zone mémoire + copie de tous les pointeurs
		productStatsList = append(productStatsList, ps)
	}

	// ⚠️ PROBLÈME MAJEUR 3: BUBBLE SORT - Complexité O(n²)!
	// ALGO: Tri à bulles = compare chaque paire d'éléments adjacents
	//   - Itération externe: n-1 passes
	//   - Itération interne: n-i-1 comparaisons par passe
	//   - Total: ~n²/2 comparaisons et swaps
	// PERFORMANCE: Si 1000 produits = ~500,000 comparaisons!
	//   - sort.Slice() utilise quicksort/introsort: O(n log n) = ~10,000 ops
	//   - C'est 50x plus lent qu'un tri optimisé!
	// MÉMOIRE: Tri en place, pas d'allocation supplémentaire (bien)
	n := len(productStatsList)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// SYNTAXE: a, b = b, a = swap simultané (feature Go)
			//   - Pas besoin de variable temporaire
			//   - Compilateur optimise en 3 MOV instructions
			if productStatsList[j].revenue < productStatsList[j+1].revenue {
				productStatsList[j], productStatsList[j+1] = productStatsList[j+1], productStatsList[j]
			}
		}
	}

	// Prendre le top 10
	limit := 10
	if len(productStatsList) < limit {
		limit = len(productStatsList)
	}

	var topProducts []*domain.ProductStats
	// on crée les meilleurs produits
	for i := 0; i < limit; i++ {
		// on récupère leur donnes
		ps := productStatsList[i]
		rev, _ := shareddomain.NewMoney(ps.revenue, "EUR")
		qty, _ := shareddomain.NewQuantity(ps.quantity)
		topProducts = append(topProducts, domain.NewProductStats(
			catalogdomain.ProductID(ps.productID),
			ps.productName,
			rev,
			len(ps.orders),
			qty,
		))
	}
	stats.SetTopProducts(topProducts)

	// Pour les autres stats, on utilise les méthodes optimisées du repository
	// (sinon ce serait trop long à implémenter toutes les inefficacités)
	// Dans le vrai V1, elles utilisaient aussi des boucles imbriquées

	categoryStats, err := s.statsRepo.GetCategoryStats(dateRange)
	if err != nil {
		return nil, err
	}
	stats.SetCategoryStats(categoryStats)

	topStores, err := s.statsRepo.GetTopStores(dateRange, 5)
	if err != nil {
		return nil, err
	}
	stats.SetTopStores(topStores)

	paymentDistrib, err := s.statsRepo.GetPaymentMethodDistribution(dateRange)
	if err != nil {
		return nil, err
	}
	stats.SetPaymentDistribution(paymentDistrib)

	return stats, nil
}

// productStatTemp structure temporaire pour les calculs inefficaces
// MÉMOIRE: Taille approximative de cette struct:
//   - productID: 8 bytes (int64)
//   - productName: 16 bytes (string = pointeur 8b + length 8b)
//   - revenue: 8 bytes (float64)
//   - orders: 24 bytes (map header = pointeur + 2 int)
//   - quantity: 8 bytes (int, aligné sur 8)
//   - Padding: possiblement 0-7 bytes pour alignement mémoire
//   - TOTAL: ~64-72 bytes par struct
//
// STACK vs HEAP: Allouée sur HEAP via &productStatTemp{} dans le code
type productStatTemp struct {
	productID   int64
	productName string
	revenue     float64
	orders      map[int64]bool // Set d'order IDs (utilise bool comme marker)
	quantity    int
}
