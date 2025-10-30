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

// Méthode ExportSalesToCSV : génère un CSV en mémoire contenant les ventes récentes
// Retourne un tableau d’octets ([]byte) sans écrire sur disque — rapide, en RAM (heap)
func (s *ExportServiceV2) ExportSalesToCSV(days int) ([]byte, error) {

	// Crée une plage de dates à partir du nombre de jours demandé
	// Alloue un petit objet DateRange sur le heap (via retour de fonction)
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Récupère toutes les ventes sur la période via une requête SQL optimisée
	// Retourne une slice allouée sur le heap contenant les structs de ventes
	salesData, err := s.exportRepo.GetSalesDataOptimized(dateRange)
	if err != nil {
		return nil, err
	}

	// Pré-alloue un buffer de 1 Mo sur le heap pour éviter les reallocations successives
	// bytes.NewBuffer référence directement ce slice interne pour y écrire le CSV
	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024)) // 1 MB initial

	// Crée un writer CSV qui écrit dans le buffer en mémoire (aucun I/O disque)
	writer := csv.NewWriter(buffer)

	// Écrit la première ligne du CSV (en-têtes)
	// Petits objets temporaires sur la stack (slice de string courte)
	if err := writer.Write(domain.CSVHeaders()); err != nil {
		return nil, err
	}

	// Parcourt chaque vente (chaque ligne du CSV)
	// Chaque itération crée une petite slice temporaire (ToCSVRow) sur le stack
	for i, row := range salesData {

		// Écrit la ligne courante dans le buffer CSV (copie des données dans le heap)
		if err := writer.Write(row.ToCSVRow()); err != nil {
			return nil, err
		}

		// Tous les batchSize (ex : 1000 lignes), on force le flush pour vider le buffer interne
		// Réduit la pression mémoire (heap) et améliore le débit global
		if (i+1)%s.batchSize == 0 {
			writer.Flush()
		}
	}

	// Flush final pour s’assurer que tout est écrit dans le buffer
	writer.Flush()

	// Vérifie si des erreurs d’écriture sont survenues (buffer plein, etc.)
	if err := writer.Error(); err != nil {
		return nil, err
	}

	// Retourne le contenu du buffer sous forme d’octets (slice sur le heap)
	// Aucun fichier créé, parfait pour une réponse HTTP rapide
	return buffer.Bytes(), nil
}

// ExportStatsToCSV exporte les statistiques en CSV
func (s *ExportServiceV2) ExportStatsToCSV(days int) ([]byte, error) {
	// Utiliser le service de stats optimisé avec cache
	stats, err := s.statsService.GetStats(days)
	if err != nil {
		return nil, err
	}

	// Un buffer est une zone temporaire en mémoire pour accumuler des données
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
	dateRange, err := shareddomain.NewDateRangeFromDays(days)
	if err != nil {
		return nil, err
	}

	// Récupérer les données optimisées
	salesData, err := s.exportRepo.GetSalesDataOptimized(dateRange)
	if err != nil {
		return nil, err
	}

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
