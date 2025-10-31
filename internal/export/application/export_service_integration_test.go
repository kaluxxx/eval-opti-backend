package application

import (
	"testing"

	analyticsapp "eval/internal/analytics/application"
	shareddomain "eval/internal/shared/domain"
	"eval/internal/testhelpers"
)

// ========================================
// INTEGRATION BENCHMARKS - REAL DATABASE
// ========================================
// Ces benchmarks utilisent PostgreSQL et mesurent les performances RÉELLES
// incluant : latence SQL, transfert réseau, parsing, allocations, etc.

// ========================================
// Test Helpers
// ========================================

// setupExportServices crée les services nécessaires pour les tests d'export
func setupExportServices(ctx *testhelpers.TestContext) (*ExportServiceV1, *ExportServiceV2) {
	// Services V1
	statsServiceV1 := analyticsapp.NewStatsServiceV1(ctx.StatsQueryRepo, ctx.ProductQueryRepo)
	exportServiceV1 := NewExportServiceV1(ctx.ExportQueryRepo, statsServiceV1)

	// Services V2
	statsServiceV2 := analyticsapp.NewStatsServiceV2(ctx.StatsQueryRepo, ctx.Cache)
	exportServiceV2 := NewExportServiceV2(ctx.ExportQueryRepo, statsServiceV2)

	return exportServiceV1, exportServiceV2
}

// ========================================
// V1 vs V2 Direct Comparison
// ========================================

// BenchmarkComparison_V1_vs_V2_CSV compare directement V1 et V2 sur 30 jours
func BenchmarkComparison_V1_vs_V2_CSV_30Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	exportServiceV1, exportServiceV2 := setupExportServices(ctx)
	defer exportServiceV2.Cleanup()

	b.Run("V1_N+1_Queries", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			data, err := exportServiceV1.ExportSalesToCSV(30)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportMetric(float64(len(data)), "bytes")
		}
	})

	b.Run("V2_Single_JOIN", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			data, err := exportServiceV2.ExportSalesToCSV(30)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportMetric(float64(len(data)), "bytes")
		}
	})
}

// ========================================
// V2 CSV Export Benchmarks
// ========================================

// BenchmarkExportServiceV2_CSV_7Days teste l'export CSV avec 7 jours de données réelles
func BenchmarkExportServiceV2_CSV_7Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	_, exportServiceV2 := setupExportServices(ctx)
	defer exportServiceV2.Cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := exportServiceV2.ExportSalesToCSV(7)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(data)), "bytes")
	}
}

// BenchmarkExportServiceV2_CSV_30Days teste l'export CSV avec 30 jours de données réelles
func BenchmarkExportServiceV2_CSV_30Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	_, exportServiceV2 := setupExportServices(ctx)
	defer exportServiceV2.Cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := exportServiceV2.ExportSalesToCSV(30)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(data)), "bytes")
	}
}

// BenchmarkExportServiceV2_CSV_365Days teste l'export CSV avec 365 jours (charge élevée)
func BenchmarkExportServiceV2_CSV_365Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	_, exportServiceV2 := setupExportServices(ctx)
	defer exportServiceV2.Cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := exportServiceV2.ExportSalesToCSV(365)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(data)), "bytes")
	}
}

// ========================================
// Parquet Export Benchmark
// ========================================

// BenchmarkExportServiceV2_Parquet_30Days teste l'export Parquet avec WorkerPool
func BenchmarkExportServiceV2_Parquet_30Days(b *testing.B) {
	testhelpers.SkipIfNoDatabase(b)

	ctx := testhelpers.SetupTestContext(b)
	defer ctx.Cleanup()

	_, exportServiceV2 := setupExportServices(ctx)
	defer exportServiceV2.Cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := exportServiceV2.ExportToParquet(30)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(data)), "bytes")
	}
}

// ========================================
// Repository Benchmarks (SQL Queries)
// ========================================

// BenchmarkExportRepo_GetSalesDataOptimized mesure uniquement la requête SQL optimisée
func BenchmarkExportRepo_GetSalesDataOptimized_30Days(b *testing.B) {
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

		salesData, err := ctx.ExportQueryRepo.GetSalesDataOptimized(dateRange)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportMetric(float64(len(salesData)), "rows")
	}
}
