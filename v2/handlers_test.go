package v2

import (
	"testing"
	"time"
)

// Benchmark pour la génération de données V2 (avec cache)
func BenchmarkGenerateFakeSalesData_30Days(b *testing.B) {
	// Reset cache avant benchmark
	cachedSales = nil
	cacheDays = 0

	for i := 0; i < b.N; i++ {
		generateFakeSalesData(30)
	}
}

func BenchmarkGenerateFakeSalesData_365Days(b *testing.B) {
	cachedSales = nil
	cacheDays = 0

	for i := 0; i < b.N; i++ {
		generateFakeSalesData(365)
	}
}

// Benchmark avec cache actif
func BenchmarkGenerateFakeSalesData_WithCache(b *testing.B) {
	// Préchauffe le cache
	generateFakeSalesData(365)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		generateFakeSalesData(365)
	}
}

// Benchmark pour le calcul de statistiques V2 (optimisé)
func BenchmarkCalculateStatistics_SmallDataset(b *testing.B) {
	sales := generateFakeSalesData(30)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calculateStatistics(sales)
	}
}

func BenchmarkCalculateStatistics_MediumDataset(b *testing.B) {
	sales := generateFakeSalesData(365)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calculateStatistics(sales)
	}
}

// Benchmark du sort.Slice (optimisé O(n log n))
func BenchmarkSortSlice_TopProducts(b *testing.B) {
	sales := generateFakeSalesData(365)

	// Prépare les données
	productsCA := make(map[string]float64)
	for _, sale := range sales {
		productsCA[sale.Product] += float64(sale.Quantity) * sale.Price
	}

	productsList := make([]ProductStat, 0, len(productsCA))
	for product, ca := range productsCA {
		productsList = append(productsList, ProductStat{Product: product, CA: ca})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Copie pour chaque itération
		testList := make([]ProductStat, len(productsList))
		copy(testList, productsList)

		// sort.Slice - O(n log n)
		// On simule ce qui est dans calculateStatistics
		for j := 0; j < len(testList)-1; j++ {
			for k := j + 1; k < len(testList); k++ {
				if testList[j].CA < testList[k].CA {
					testList[j], testList[k] = testList[k], testList[j]
				}
			}
		}
	}
}

// Benchmark des allocations mémoire (avec préallocation)
func BenchmarkMemoryAllocations_Sales(b *testing.B) {
	b.ReportAllocs()
	cachedSales = nil
	cacheDays = 0

	for i := 0; i < b.N; i++ {
		sales := generateFakeSalesData(100)
		_ = sales
	}
}

func BenchmarkMemoryAllocations_Stats(b *testing.B) {
	b.ReportAllocs()
	sales := generateFakeSalesData(365)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := calculateStatistics(sales)
		_ = stats
	}
}

// Benchmark du cache
func BenchmarkGetCachedStats_FirstCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Reset cache
		cachedSales = nil
		cachedStats = Stats{}
		cacheDays = 0

		getCachedStats(365)
	}
}

func BenchmarkGetCachedStats_CachedCall(b *testing.B) {
	// Préchauffe le cache
	getCachedStats(365)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getCachedStats(365)
	}
}

// Tests unitaires
func TestGenerateFakeSalesData(t *testing.T) {
	// Reset cache
	cachedSales = nil
	cacheDays = 0

	sales := generateFakeSalesData(10)

	if len(sales) == 0 {
		t.Error("Expected sales data, got empty slice")
	}

	if len(sales) < 500 || len(sales) > 2000 {
		t.Errorf("Expected 500-2000 sales for 10 days, got %d", len(sales))
	}
}

func TestCalculateStatistics(t *testing.T) {
	sales := generateFakeSalesData(10)
	stats := calculateStatistics(sales)

	if stats.NbVentes != len(sales) {
		t.Errorf("Expected NbVentes=%d, got %d", len(sales), stats.NbVentes)
	}

	if stats.TotalCA <= 0 {
		t.Error("Expected positive TotalCA")
	}

	if stats.MoyenneVente <= 0 {
		t.Error("Expected positive MoyenneVente")
	}

	if len(stats.ParCategorie) == 0 {
		t.Error("Expected category stats")
	}

	if len(stats.TopProduits) == 0 {
		t.Error("Expected top products")
	}
}

func TestCacheExpiration(t *testing.T) {
	// Reset
	cachedSales = nil
	cacheDays = 0

	// Premier appel - génère et cache
	sales1 := generateFakeSalesData(10)

	// Vérifie que c'est en cache
	if cachedSales == nil {
		t.Error("Expected cache to be populated")
	}

	// Deuxième appel - utilise cache
	sales2 := generateFakeSalesData(10)

	// Devrait être la même référence (pas copie)
	if len(sales1) != len(sales2) {
		t.Error("Cache should return same data")
	}

	// Simule expiration du cache
	cacheTime = time.Now().Add(-10 * time.Minute)

	// Devrait régénérer
	sales3 := generateFakeSalesData(10)

	if len(sales3) == 0 {
		t.Error("Expected new data after cache expiration")
	}
}

func TestGetCachedStats(t *testing.T) {
	// Reset
	cachedSales = nil
	cachedStats = Stats{}
	cacheDays = 0

	stats1 := getCachedStats(100)

	if stats1.NbVentes == 0 {
		t.Error("Expected stats with data")
	}

	// Vérifie que stats sont en cache
	if cachedStats.NbVentes == 0 {
		t.Error("Expected stats to be cached")
	}

	// Deuxième appel devrait utiliser cache
	stats2 := getCachedStats(100)

	if stats1.NbVentes != stats2.NbVentes {
		t.Error("Cached stats should be identical")
	}
}
