package domain

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// ========================================
// Benchmarks: ToCSVRow Method Optimization
// ========================================

// BenchmarkSaleExportRow_ToCSVRow_Current benchmarks l'implémentation actuelle
func BenchmarkSaleExportRow_ToCSVRow_Current(b *testing.B) {
	row := NewSaleExportRow(
		1001, 501, 1, "Store Downtown", 201, "Laptop Pro",
		"Electronics", 2, 1299.99, 2599.98,
		"Credit Card", "PROMO123",
		time.Date(2024, 10, 15, 14, 30, 0, 0, time.UTC),
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = row.ToCSVRow()
	}
}

// BenchmarkSaleExportRow_ToCSVRow_Optimized benchmarks version optimisée avec strconv
func BenchmarkSaleExportRow_ToCSVRow_Optimized(b *testing.B) {
	row := NewSaleExportRow(
		1001, 501, 1, "Store Downtown", 201, "Laptop Pro",
		"Electronics", 2, 1299.99, 2599.98,
		"Credit Card", "PROMO123",
		time.Date(2024, 10, 15, 14, 30, 0, 0, time.UTC),
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = toCSVRowOptimized(row)
	}
}

// toCSVRowOptimized version optimisée utilisant strconv au lieu de fmt.Sprintf
func toCSVRowOptimized(ser *SaleExportRow) []string {
	return []string{
		strconv.FormatInt(ser.OrderID, 10),
		strconv.FormatInt(ser.CustomerID, 10),
		strconv.FormatInt(ser.StoreID, 10),
		ser.StoreName,
		strconv.FormatInt(ser.ProductID, 10),
		ser.ProductName,
		ser.CategoryName,
		strconv.Itoa(ser.Quantity),
		strconv.FormatFloat(ser.UnitPrice, 'f', 2, 64),
		strconv.FormatFloat(ser.Subtotal, 'f', 2, 64),
		ser.PaymentMethod,
		ser.PromotionCode,
		ser.OrderDate.Format("2006-01-02 15:04:05"),
	}
}

// ========================================
// Benchmarks: CSV Header Generation
// ========================================

// BenchmarkCSVHeaders teste la génération des headers CSV
func BenchmarkCSVHeaders(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = CSVHeaders()
	}
}

// ========================================
// Benchmarks: Row Creation
// ========================================

// BenchmarkNewSaleExportRow teste la création d'une ligne d'export
func BenchmarkNewSaleExportRow(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewSaleExportRow(
			1001, 501, 1, "Store Downtown", 201, "Laptop Pro",
			"Electronics", 2, 1299.99, 2599.98,
			"Credit Card", "PROMO123",
			time.Now(),
		)
	}
}

// ========================================
// Benchmarks: Batch Row Processing
// ========================================

// BenchmarkBatchRowProcessing_100 teste le traitement de 100 lignes
func BenchmarkBatchRowProcessing_100(b *testing.B) {
	rows := make([]*SaleExportRow, 100)
	for i := 0; i < 100; i++ {
		rows[i] = NewSaleExportRow(
			int64(1000+i), int64(500+i), int64(1+i%10), "Store",
			int64(200+i), "Product", "Category", 2, 99.99, 199.98,
			"Credit Card", "PROMO", time.Now(),
		)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, row := range rows {
			_ = row.ToCSVRow()
		}
	}
}

// BenchmarkBatchRowProcessing_1000 teste le traitement de 1000 lignes
func BenchmarkBatchRowProcessing_1000(b *testing.B) {
	rows := make([]*SaleExportRow, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = NewSaleExportRow(
			int64(1000+i), int64(500+i), int64(1+i%10), "Store",
			int64(200+i), "Product", "Category", 2, 99.99, 199.98,
			"Credit Card", "PROMO", time.Now(),
		)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, row := range rows {
			_ = row.ToCSVRow()
		}
	}
}

// ========================================
// Benchmarks: String Building Strategies
// ========================================

// BenchmarkStringBuilding_Concatenation teste avec concaténation simple
func BenchmarkStringBuilding_Concatenation(b *testing.B) {
	row := NewSaleExportRow(
		1001, 501, 1, "Store", 201, "Product",
		"Category", 2, 99.99, 199.98,
		"Credit", "PROMO", time.Now(),
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = strconv.FormatInt(row.OrderID, 10) + "," +
			strconv.FormatInt(row.CustomerID, 10) + "," +
			row.StoreName
	}
}

// BenchmarkStringBuilding_Builder teste avec strings.Builder
func BenchmarkStringBuilding_Builder(b *testing.B) {
	row := NewSaleExportRow(
		1001, 501, 1, "Store", 201, "Product",
		"Category", 2, 99.99, 199.98,
		"Credit", "PROMO", time.Now(),
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		builder.Grow(128)
		builder.WriteString(strconv.FormatInt(row.OrderID, 10))
		builder.WriteByte(',')
		builder.WriteString(strconv.FormatInt(row.CustomerID, 10))
		builder.WriteByte(',')
		builder.WriteString(row.StoreName)
		_ = builder.String()
	}
}

// BenchmarkStringBuilding_PreallocatedSlice teste avec slice pré-allouée
func BenchmarkStringBuilding_PreallocatedSlice(b *testing.B) {
	row := NewSaleExportRow(
		1001, 501, 1, "Store", 201, "Product",
		"Category", 2, 99.99, 199.98,
		"Credit", "PROMO", time.Now(),
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fields := make([]string, 0, 13)
		fields = append(fields,
			strconv.FormatInt(row.OrderID, 10),
			strconv.FormatInt(row.CustomerID, 10),
			row.StoreName,
		)
		_ = fields
	}
}
