package application

import (
	"fmt"
	"sync"
	"time"

	"eval/internal/analytics/domain"
	"eval/internal/analytics/infrastructure"
	shareddomain "eval/internal/shared/domain"
	sharedinfra "eval/internal/shared/infrastructure"
)

// StatsServiceV2 service optimisé pour le calcul des statistiques (Version 2)
// PERFORMANCE: Structure allouée sur le HEAP si retournée par une fonction ou passée par pointeur
type StatsServiceV2 struct {
	statsRepo *infrastructure.StatsQueryRepository // Pointeur : heap
	cache     sharedinfra.Cache                    // “Stocké inline” = “stocké directement dans la stack”
	cacheTTL  time.Duration
}

// NewStatsServiceV2 crée une nouvelle instance de StatsServiceV2
// Pattern "constructor" - retourne un pointeur pour éviter la copie de la struct
func NewStatsServiceV2(
	statsRepo *infrastructure.StatsQueryRepository,
	cache sharedinfra.Cache,
) *StatsServiceV2 {
	return &StatsServiceV2{
		statsRepo: statsRepo,
		cache:     cache,
		cacheTTL:  5 * time.Minute,
	}
}

// GetStats calcule les statistiques de manière optimisée avec cache et goroutines
// Pattern "cache-aside" pour éviter les calculs coûteux
func (s *StatsServiceV2) GetStats(days int) (*domain.Stats, error) {
	// CACHE: Vérifier le cache en premier (hot path optimization)
	// MÉMOIRE: buildCacheKey alloue une string sur le HEAP
	cacheKey := s.buildCacheKey(days)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*domain.Stats), nil
	}

	// Créer la période
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Calculer les stats avec goroutines parallèles
	stats, err := s.calculateStatsOptimized(dateRange)
	if err != nil {
		return nil, err
	}

	// Mettre en cache
	s.cache.Set(cacheKey, stats, s.cacheTTL)

	return stats, nil
}

// calculateStatsOptimized calcule les stats avec des requêtes SQL parallèles
// CONCURRENCE: Utilise des goroutines pour paralléliser 5 requêtes SQL indépendantes
// PERFORMANCE: Temps total ≈ max(temps requête) au lieu de Σ(temps requêtes)
// ATTENTION: Trade-off mémoire/vitesse - 5 goroutines actives simultanément (sur la stack)
func (s *StatsServiceV2) calculateStatsOptimized(dateRange shareddomain.DateRange) (*domain.Stats, error) {
	stats := domain.NewStats()

	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	// Goroutine 1: Stats globales
	wg.Add(1)
	go func() {
		defer wg.Done()
		revenue, orders, avgOrder, err := s.statsRepo.GetGlobalStats(dateRange)
		if err != nil {
			errChan <- fmt.Errorf("global stats error: %w", err)
			return
		}
		stats.SetTotalRevenue(revenue)
		stats.SetTotalOrders(orders)
		stats.SetAverageOrderValue(avgOrder)
	}()

	// Goroutine 2: Stats par catégorie
	wg.Add(1)
	go func() {
		defer wg.Done()
		categoryStats, err := s.statsRepo.GetCategoryStats(dateRange)
		if err != nil {
			errChan <- fmt.Errorf("category stats error: %w", err)
			return
		}
		stats.SetCategoryStats(categoryStats)
	}()

	// Goroutine 3: Top 10 produits
	wg.Add(1)
	go func() {
		defer wg.Done()
		topProducts, err := s.statsRepo.GetTopProducts(dateRange, 10)
		if err != nil {
			errChan <- fmt.Errorf("top products error: %w", err)
			return
		}
		stats.SetTopProducts(topProducts)
	}()

	// Goroutine 4: Top 5 magasins
	wg.Add(1)
	go func() {
		defer wg.Done()
		topStores, err := s.statsRepo.GetTopStores(dateRange, 5)
		if err != nil {
			errChan <- fmt.Errorf("top stores error: %w", err)
			return
		}
		stats.SetTopStores(topStores)
	}()

	// Goroutine 5: Distribution des moyens de paiement
	wg.Add(1)
	go func() {
		defer wg.Done()
		paymentDistrib, err := s.statsRepo.GetPaymentMethodDistribution(dateRange)
		if err != nil {
			errChan <- fmt.Errorf("payment distribution error: %w", err)
			return
		}
		stats.SetPaymentDistribution(paymentDistrib)
	}()

	// Attendre la fin de toutes les goroutines
	wg.Wait()
	close(errChan)

	// Vérifier s'il y a eu des erreurs
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return stats, nil
}

// buildCacheKey construit la clé de cache
// PERFORMANCE: Utilise un builder pour éviter les concatenations multiples de strings
// MÉMOIRE : Les strings en Go sont immuables, chaque concat créerait une nouvelle string
// ALLOCATION: Le builder utilise un buffer interne (probablement bytes.Buffer ou strings.Builder)
func (s *StatsServiceV2) buildCacheKey(days int) string {
	// BUILDER PATTERN: Efficace pour construire des strings avec plusieurs parties
	// MÉMOIRE: NewCacheKeyBuilder() alloue probablement une struct avec un buffer sur le HEAP
	// PERFORMANCE: Évite N-1 allocations intermédiaires (où N = nombre de Add)
	// Sans builder: "stats" + "v2" + strconv.Itoa(days) = 2 allocations intermédiaires
	// Avec builder: 1 seule allocation finale
	return sharedinfra.NewCacheKeyBuilder().
		Add("stats").
		Add("v2").
		AddInt(days).
		Build()
}

// InvalidateCache invalide le cache pour un nombre de jours donné
func (s *StatsServiceV2) InvalidateCache(days int) {
	cacheKey := s.buildCacheKey(days)
	s.cache.Delete(cacheKey)
}

// ClearCache vide tout le cache
func (s *StatsServiceV2) ClearCache() {
	s.cache.Clear()
}
