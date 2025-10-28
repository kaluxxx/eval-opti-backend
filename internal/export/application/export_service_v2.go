package application

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"eval/internal/analytics/application"
	"eval/internal/export/domain"
	"eval/internal/export/infrastructure"
	shareddomain "eval/internal/shared/domain"
	sharedinfra "eval/internal/shared/infrastructure"
)

// ExportServiceV2 service optimisé pour les exports (Version 2)
type ExportServiceV2 struct {
	exportRepo   *infrastructure.ExportQueryRepository
	statsService *application.StatsServiceV2
	workerPool   *sharedinfra.WorkerPool
	batchSize    int
}

// NewExportServiceV2 crée une nouvelle instance de ExportServiceV2
func NewExportServiceV2(
	exportRepo *infrastructure.ExportQueryRepository,
	statsService *application.StatsServiceV2,
) *ExportServiceV2 {
	return &ExportServiceV2{
		exportRepo:   exportRepo,
		statsService: statsService,
		workerPool:   sharedinfra.NewWorkerPool(4), // 4 workers
		batchSize:    1000,
	}
}

// ExportSalesToCSV exporte les ventes en CSV de manière optimisée
func (s *ExportServiceV2) ExportSalesToCSV(days int) ([]byte, error) {
	// Créer la période
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Récupérer les données de manière optimisée (une seule requête)
	salesData, err := s.exportRepo.GetSalesDataOptimized(dateRange)
	if err != nil {
		return nil, err
	}

	// Pré-allouer le buffer (optimisation V2)
	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024)) // 1 MB initial
	writer := csv.NewWriter(buffer)

	// Écrire les en-têtes
	if err := writer.Write(domain.CSVHeaders()); err != nil {
		return nil, err
	}

	// Écrire les données par batch avec flush périodique
	for i, row := range salesData {
		if err := writer.Write(row.ToCSVRow()); err != nil {
			return nil, err
		}

		// Flush tous les 1000 lignes pour optimiser la mémoire
		if (i+1)%s.batchSize == 0 {
			writer.Flush()
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// ExportStatsToCSV exporte les statistiques en CSV
func (s *ExportServiceV2) ExportStatsToCSV(days int) ([]byte, error) {
	// Utiliser le service de stats optimisé avec cache
	stats, err := s.statsService.GetStats(days)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 64*1024)) // 64 KB
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

// ExportToParquet exporte en format Parquet avec worker pool (simplifié ici - juste structure)
// Note: L'implémentation complète de Parquet nécessiterait la library parquet-go
func (s *ExportServiceV2) ExportToParquet(days int) ([]byte, error) {
	// Pour simplifier, on retourne juste une indication
	// Dans le vrai V2, ceci utiliserait un worker pool avec batches de 1000 rows
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Récupérer les données optimisées
	salesData, err := s.exportRepo.GetSalesDataOptimized(dateRange)
	if err != nil {
		return nil, err
	}

	// TODO: Implémenter l'export Parquet avec worker pool
	// Pour l'instant, on retourne juste une confirmation
	message := fmt.Sprintf("Parquet export would process %d rows with worker pool (batch size: %d)",
		len(salesData), s.batchSize)

	return []byte(message), nil
}

// Cleanup nettoie les ressources
func (s *ExportServiceV2) Cleanup() {
	if s.workerPool != nil {
		s.workerPool.Stop()
	}
}
