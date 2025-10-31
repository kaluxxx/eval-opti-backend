package application

import (
	"testing"

	shareddomain "eval/internal/shared/domain"
	"eval/internal/testhelpers"
)

// ========================================
// INTEGRATION BENCHMARKS - REAL DATABASE
// ========================================
// Ces benchmarks utilisent PostgreSQL et mesurent les performances RÉELLES
// incluant : agrégations SQL, goroutines parallèles, cache, etc.

// ========================================
// Test Helpers
// ========================================

// setupStatsServices crée les services nécessaires pour les tests de stats
func setupStatsServices(ctx *testhelpers.TestContext) (*StatsServiceV1, *StatsServiceV2) {
	// Services V1
	statsServiceV1 := NewStatsServiceV1(ctx.StatsQueryRepo, ctx.ProductQueryRepo)

	// Services V2
	statsServiceV2 := NewStatsServiceV2(ctx.StatsQueryRepo, ctx.Cache)

	return statsServiceV1, statsServiceV2
}

// ========================================
// V1 vs V2 Direct Comparison
// ========================================

// BenchmarkComparison_V1_vs_V2_Stats compare directement V1 et V2 sur 30 jours
func BenchmarkComparison_V1_vs_V2_Stats_30Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	statsServiceV1, statsServiceV2 := setupStatsServices(ctx)

	b.Run("V1_N+1_BubbleSort", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			stats, err := statsServiceV1.GetStats(30)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportMetric(float64(stats.TotalOrders()), "orders")
		}
	})

	b.Run("V2_Optimized_CacheMiss", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			ctx.ClearCache()
			b.StartTimer()

			stats, err := statsServiceV2.GetStats(30)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportMetric(float64(stats.TotalOrders()), "orders")
		}
	})

	b.Run("V2_Optimized_CacheHit", func(b *testing.B) {
		b.ReportAllocs()

		// Chauffer le cache
		_, _ = statsServiceV2.GetStats(30)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			stats, err := statsServiceV2.GetStats(30)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportMetric(float64(stats.TotalOrders()), "orders")
		}
	})
}

// ========================================
// V2 Performance Tests
// ========================================

// BenchmarkStatsServiceV2_7Days teste V2 avec 7 jours (cache miss)
func BenchmarkStatsServiceV2_7Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	_, statsServiceV2 := setupStatsServices(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx.ClearCache()
		b.StartTimer()

		stats, err := statsServiceV2.GetStats(7)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportMetric(float64(stats.TotalOrders()), "orders")
		b.ReportMetric(stats.TotalRevenue().Amount(), "revenue")
	}
}

// BenchmarkStatsServiceV2_365Days teste V2 avec 365 jours (cache miss)
func BenchmarkStatsServiceV2_365Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	_, statsServiceV2 := setupStatsServices(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx.ClearCache()
		b.StartTimer()

		stats, err := statsServiceV2.GetStats(365)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportMetric(float64(stats.TotalOrders()), "orders")
	}
}

// ========================================
// Repository Benchmarks
// ========================================

// BenchmarkStatsRepo_GetGlobalStats mesure uniquement GetGlobalStats (agrégation SQL)
func BenchmarkStatsRepo_GetGlobalStats_30Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dateRange, err := shareddomain.NewDateRangeFromDays(30)
		if err != nil {
			b.Fatal(err)
		}

		revenue, orders, avg, err := ctx.StatsQueryRepo.GetGlobalStats(dateRange)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportMetric(float64(orders), "orders")
		b.ReportMetric(revenue.Amount(), "revenue")
		_ = avg
	}
}
