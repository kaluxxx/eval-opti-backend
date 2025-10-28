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
func (s *StatsServiceV1) calculateStatsInefficient(dateRange shareddomain.DateRange) (*domain.Stats, error) {
	stats := domain.NewStats()

	// Récupérer TOUS les items (pas de GROUP BY, charge tout en mémoire)
	allItems, err := s.statsRepo.GetAllOrderItems(dateRange)
	if err != nil {
		return nil, err
	}

	// Calcul 1: Chiffre d'affaires total (boucle simple)
	totalRevenue := 0.0
	totalOrders := make(map[int64]bool)
	for _, item := range allItems {
		totalRevenue += item.Subtotal
		totalOrders[item.OrderID] = true
	}

	revenue, _ := shareddomain.NewMoney(totalRevenue, "EUR")
	stats.SetTotalRevenue(revenue)
	stats.SetTotalOrders(len(totalOrders))

	if len(totalOrders) > 0 {
		avgOrder, _ := shareddomain.NewMoney(totalRevenue/float64(len(totalOrders)), "EUR")
		stats.SetAverageOrderValue(avgOrder)
	}

	// Calcul 2: Stats par produit (N+1 queries - UN QUERY PAR PRODUIT!)
	// Ceci est volontairement inefficace
	productStats := make(map[int64]*productStatTemp)
	for _, item := range allItems {
		if _, exists := productStats[item.ProductID]; !exists {
			// N+1 QUERY PROBLEM: Une requête par produit distinct!
			product, err := s.productRepo.FindByID(catalogdomain.ProductID(item.ProductID))
			if err != nil {
				// Si erreur, on utilise un nom par défaut
				productStats[item.ProductID] = &productStatTemp{
					productID:   item.ProductID,
					productName: "Unknown Product",
					revenue:     0,
					orders:      make(map[int64]bool),
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

	// Convertir en slice pour le tri
	var productStatsList []*productStatTemp
	for _, ps := range productStats {
		productStatsList = append(productStatsList, ps)
	}

	// BUBBLE SORT O(n²) - volontairement inefficace!
	n := len(productStatsList)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
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
	for i := 0; i < limit; i++ {
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
type productStatTemp struct {
	productID   int64
	productName string
	revenue     float64
	orders      map[int64]bool
	quantity    int
}
