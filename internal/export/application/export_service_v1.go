package application

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"eval/internal/analytics/application"
	"eval/internal/export/domain"
	"eval/internal/export/infrastructure"
	shareddomain "eval/internal/shared/domain"
)

// ExportServiceV1 service NON-optimisé pour les exports (Version 1)
type ExportServiceV1 struct {
	exportRepo   *infrastructure.ExportQueryRepository
	statsService *application.StatsServiceV1
}

// NewExportServiceV1 crée une nouvelle instance de ExportServiceV1
func NewExportServiceV1(
	exportRepo *infrastructure.ExportQueryRepository,
	statsService *application.StatsServiceV1,
) *ExportServiceV1 {
	return &ExportServiceV1{
		exportRepo:   exportRepo,
		statsService: statsService,
	}
}

// ExportSalesToCSV exporte les ventes en CSV de manière inefficace (N+1 queries)
func (s *ExportServiceV1) ExportSalesToCSV(days int) ([]byte, error) {
	// Créer la période
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Récupérer les données avec N+1 queries (INEFFICACE!)
	salesData, err := s.exportRepo.GetSalesDataInefficient(dateRange)
	if err != nil {
		return nil, err
	}

	// Pas de pré-allocation du buffer (inefficace)
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)

	// Écrire les en-têtes
	if err := writer.Write(domain.CSVHeaders()); err != nil {
		return nil, err
	}

	// Écrire toutes les données sans flush intermédiaire (charge en mémoire)
	for _, row := range salesData {
		if err := writer.Write(row.ToCSVRow()); err != nil {
			return nil, err
		}
	}

	// Flush une seule fois à la fin
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// ExportStatsToCSV exporte les statistiques en CSV (utilise le service V1 non-optimisé)
func (s *ExportServiceV1) ExportStatsToCSV(days int) ([]byte, error) {
	// Utiliser le service de stats V1 (avec N+1 et bubble sort)
	stats, err := s.statsService.GetStats(days)
	if err != nil {
		return nil, err
	}

	// Création d’un nouveau buffer vide pour accumuler les données CSV. - stockés dans la Heap
	buffer := &bytes.Buffer{}

	//writer est une struct légère stockée sur la stack, contenant :
	// une référence (io.Writer) vers le buffer (donc un pointeur vers le heap)
	//un petit buffer interne de 4 KB environ pour écrire efficacement
	writer := csv.NewWriter(buffer)

	// En-têtes pour les stats globales
	writer.Write([]string{"Type", "Metric", "Value"})

	// Stats globales
	writer.Write([]string{"Global", "Total Revenue", fmt.Sprintf("%.2f", stats.TotalRevenue().Amount())})
	writer.Write([]string{"Global", "Total Orders", fmt.Sprintf("%d", stats.TotalOrders())})
	writer.Write([]string{"Global", "Average Order Value", fmt.Sprintf("%.2f", stats.AverageOrderValue().Amount())})

	// Saut de ligne
	writer.Write([]string{})

	// Stats par catégorie
	writer.Write([]string{"Category Stats", "", ""})
	writer.Write([]string{"Category Name", "Total Revenue", "Total Orders"})
	for _, cs := range stats.CategoryStats() {
		writer.Write([]string{
			cs.CategoryName(),
			fmt.Sprintf("%.2f", cs.TotalRevenue().Amount()),
			fmt.Sprintf("%d", cs.TotalOrders()),
		})
	}

	// Saut de ligne
	writer.Write([]string{})

	// Top produits
	writer.Write([]string{"Top Products", "", ""})
	writer.Write([]string{"Product Name", "Total Revenue", "Total Orders", "Total Quantity"})
	for _, ps := range stats.TopProducts() {
		writer.Write([]string{
			ps.ProductName(),
			fmt.Sprintf("%.2f", ps.TotalRevenue().Amount()),
			fmt.Sprintf("%d", ps.TotalOrders()),
			fmt.Sprintf("%d", ps.TotalQuantity().Value()),
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// ExportToParquet exporte en format Parquet de manière inefficace (tout en mémoire)
func (s *ExportServiceV1) ExportToParquet(days int) ([]byte, error) {
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Récupérer TOUTES les données en mémoire d'un coup (INEFFICACE!)
	salesData, err := s.exportRepo.GetSalesDataInefficient(dateRange)
	if err != nil {
		return nil, err
	}

	// TODO: Implémenter l'export Parquet inefficace (tout en mémoire)
	// Pour l'instant, on retourne juste une confirmation
	message := fmt.Sprintf("Parquet export (V1) would load all %d rows in memory at once",
		len(salesData))

	return []byte(message), nil
}
