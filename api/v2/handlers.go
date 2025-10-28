package v2

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	analyticsapp "eval/internal/analytics/application"
	exportapp "eval/internal/export/application"
)

// Handlers contient tous les handlers pour l'API V2 (optimisée)
type Handlers struct {
	statsService  *analyticsapp.StatsServiceV2
	exportService *exportapp.ExportServiceV2
}

// NewHandlers crée une nouvelle instance des handlers V2
func NewHandlers(
	statsService *analyticsapp.StatsServiceV2,
	exportService *exportapp.ExportServiceV2,
) *Handlers {
	return &Handlers{
		statsService:  statsService,
		exportService: exportService,
	}
}

// GetStats handler pour GET /api/v2/stats
func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	// Récupérer le paramètre days
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 365 // Valeur par défaut
	}

	// Utiliser le service V2 (optimisé avec cache + goroutines parallèles)
	stats, err := h.statsService.GetStats(days)
	if err != nil {
		log.Printf("Error getting stats (V2): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convertir en format JSON pour la réponse
	response := h.statsToJSON(stats)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportCSV handler pour GET /api/v2/export/csv
func (h *Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30 // Valeur par défaut
	}

	// Export avec requête optimisée + batch processing
	csvData, err := h.exportService.ExportSalesToCSV(days)
	if err != nil {
		log.Printf("Error exporting CSV (V2): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_v2.csv")
	w.Write(csvData)
}

// ExportStatsCSV handler pour GET /api/v2/export/stats-csv
func (h *Handlers) ExportStatsCSV(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 365
	}

	// Utilise le service stats V2 avec cache
	csvData, err := h.exportService.ExportStatsToCSV(days)
	if err != nil {
		log.Printf("Error exporting stats CSV (V2): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=stats_v2.csv")
	w.Write(csvData)
}

// ExportParquet handler pour GET /api/v2/export/parquet
func (h *Handlers) ExportParquet(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	// Export avec worker pool + batch processing
	parquetData, err := h.exportService.ExportToParquet(days)
	if err != nil {
		log.Printf("Error exporting Parquet (V2): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_v2.parquet")
	w.Write(parquetData)
}

// statsToJSON convertit les stats du domaine en format JSON
func (h *Handlers) statsToJSON(stats interface{}) map[string]interface{} {
	// Pour simplifier, on retourne une structure générique
	// Dans un vrai projet, on créerait des DTOs spécifiques
	return map[string]interface{}{
		"version": "v2",
		"message": "Stats calculated with V2 (optimized: cached + parallel SQL queries)",
		"stats":   stats,
	}
}
