package v1

import (
	"testing"
)

// Benchmark pour la génération de données V1 (non optimisée)
func BenchmarkGenerateFakeSalesData_30Days(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateFakeSalesData(30)
	}
}

func BenchmarkGenerateFakeSalesData_365Days(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateFakeSalesData(365)
	}
}

// Benchmark pour le calcul de statistiques V1 (avec bubble sort)
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

// Benchmark du bubble sort (le bottleneck de V1)
func BenchmarkBubbleSort_TopProducts(b *testing.B) {
	sales := generateFakeSalesData(365)

	// Prépare les données pour le tri
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

		// Bubble sort O(n²)
		n := len(testList)
		for j := 0; j < n; j++ {
			for k := 0; k < n-j-1; k++ {
				if testList[k].CA < testList[k+1].CA {
					testList[k], testList[k+1] = testList[k+1], testList[k]
				}
			}
		}
	}
}

// Benchmark des allocations mémoire
func BenchmarkMemoryAllocations_Sales(b *testing.B) {
	b.ReportAllocs()

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

// Tests unitaires
func TestGenerateFakeSalesData(t *testing.T) {
	sales := generateFakeSalesData(10)

	if len(sales) == 0 {
		t.Error("Expected sales data, got empty slice")
	}

	// Vérifie qu'on a environ 50-200 ventes par jour
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
