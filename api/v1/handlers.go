package v1

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	analyticsapp "eval/internal/analytics/application"
	exportapp "eval/internal/export/application"
)

// Handlers contient tous les handlers pour l'API V1 (non-optimisée)
// MÉMOIRE: Cette struct est allouée sur le HEAP (pointeur retourné par NewHandlers)
// Les pointeurs vers services évitent de copier les structs complètes (économie mémoire)
type Handlers struct {
	statsService  *analyticsapp.StatsServiceV1 // Pointeur: 8 bytes sur 64-bit
	exportService *exportapp.ExportServiceV1   // Pointeur: 8 bytes sur 64-bit
}

// NewHandlers crée une nouvelle instance des handlers V1
// SYNTAXE: * devant un type = pointeur vers ce type
// SYNTAXE: & devant une valeur = adresse mémoire de cette valeur
// MÉMOIRE: &Handlers{} alloue la struct sur le HEAP et retourne son adresse
//   - HEAP car la struct survit après le return de la fonction
//   - Le garbage collector gérera la libération plus tard
//
// PERFORMANCE: Évite de copier toute la struct (passage par référence)
func NewHandlers(
	statsService *analyticsapp.StatsServiceV1,
	exportService *exportapp.ExportServiceV1,
) *Handlers {
	return &Handlers{
		statsService:  statsService,
		exportService: exportService,
	}
}

// GetStats handler pour GET /api/v1/stats
// SYNTAXE: (h *Handlers) = méthode receiver (similaire à "self" en Python)
// PERFORMANCE: Très lent car déclenche N+1 queries + bubble sort O(n²)
func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	// Récupérer le paramètre days de l'URL query string
	daysStr := r.URL.Query().Get("days")

	// SYNTAXE: strconv.Atoi = ASCII to Integer, converti string "365" -> int 365
	// MÉMOIRE: daysStr est un string (HEAP: pointeur + longueur), days est un int (STACK: 8 bytes)
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 365 // Valeur par défaut
	}

	// ⚠️ PERFORMANCE CRITIQUE: Cette ligne déclenche des MILLIERS de requêtes SQL!
	// Voir stats_service_v1.go pour le détail des inefficacités:
	//   - Récupère TOUS les order_items en mémoire (plusieurs MB de données)
	//   - N+1 queries: une requête SQL par produit distinct
	//   - Bubble sort O(n²) sur potentiellement des milliers de produits
	stats, err := h.statsService.GetStats(days)
	if err != nil {
		log.Printf("Error getting stats (V1): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convertir en format JSON pour la réponse
	response := h.statsToJSON(stats)

	// MÉMOIRE: json.NewEncoder encode directement dans le writer (streaming)
	// Évite d'allouer toute la string JSON en mémoire avant d'écrire
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportCSV handler pour GET /api/v1/export/csv
func (h *Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30 // Valeur par défaut
	}

	// Export avec N+1 queries (inefficace)
	csvData, err := h.exportService.ExportSalesToCSV(days)
	if err != nil {
		log.Printf("Error exporting CSV (V1): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_v1.csv")
	w.Write(csvData)
}

// ExportStatsCSV handler pour GET /api/v1/export/stats-csv
func (h *Handlers) ExportStatsCSV(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 365
	}

	csvData, err := h.exportService.ExportStatsToCSV(days)
	if err != nil {
		log.Printf("Error exporting stats CSV (V1): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=stats_v1.csv")
	w.Write(csvData)
}

// ExportParquet handler pour GET /api/v1/export/parquet
func (h *Handlers) ExportParquet(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	parquetData, err := h.exportService.ExportToParquet(days)
	if err != nil {
		log.Printf("Error exporting Parquet (V1): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_v1.parquet")
	w.Write(parquetData)
}

// statsToJSON convertit les stats du domaine en format JSON
// SYNTAXE: interface{} = type "any" en Go, accepte n'importe quel type
//   - Similaire à Object en Java ou any en TypeScript
//   - Perte de type safety mais gain de flexibilité
//
// SYNTAXE: map[string]interface{} = dictionnaire/hashmap clé=string, valeur=n'importe quoi
//   - map[KeyType]ValueType est la syntaxe des hashmaps en Go
//
// MÉMOIRE: map est alloué sur le HEAP (structure dynamique)
//   - Contient des pointeurs vers les buckets internes
//   - Les strings sont aussi sur le HEAP (immutables en Go)
//
// PERFORMANCE: ⚠️ Interface{} nécessite boxing/unboxing (overhead léger)
//   - Dans un vrai projet, utiliser des structs typées (DTOs) est plus performant
func (h *Handlers) statsToJSON(stats interface{}) map[string]interface{} {
	// Pour simplifier, on retourne une structure générique
	// Dans un vrai projet, on créerait des DTOs spécifiques
	return map[string]interface{}{
		"version": "v1",
		"message": "Stats calculated with V1 (inefficient: N+1 queries + bubble sort)",
		"stats":   stats,
	}
}
