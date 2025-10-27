package main

import (
	"testing"
)

// Benchmark pour la génération de données
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

func BenchmarkGenerateFakeSalesData_730Days(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateFakeSalesData(730)
	}
}

// Benchmark pour le calcul de statistiques
func BenchmarkCalculateStatistics_SmallDataset(b *testing.B) {
	sales := generateFakeSalesData(30) // ~1500-6000 ventes
	b.ResetTimer() // Ne pas compter la génération dans le benchmark

	for i := 0; i < b.N; i++ {
		calculateStatistics(sales)
	}
}

func BenchmarkCalculateStatistics_MediumDataset(b *testing.B) {
	sales := generateFakeSalesData(365) // ~18k-73k ventes
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calculateStatistics(sales)
	}
}

func BenchmarkCalculateStatistics_LargeDataset(b *testing.B) {
	sales := generateFakeSalesData(730) // ~36k-146k ventes
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calculateStatistics(sales)
	}
}

// Benchmark spécifique pour le bubble sort (le bottleneck probable)
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
		// Copie la liste pour chaque itération
		testList := make([]ProductStat, len(productsList))
		copy(testList, productsList)

		// Bubble sort
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

// Benchmark pour tester les allocations mémoire
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