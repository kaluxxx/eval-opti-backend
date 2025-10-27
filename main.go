package main

import (
	"encoding/json"
	"log"
	"net/http"
	_ "net/http/pprof"

	"eval/v1"
	"eval/v2"
)

func main() {
	// Health check
	http.HandleFunc("/api/health", healthHandler)

	// API V1 - Non optimisée (code original avec problèmes de performance)
	http.HandleFunc("/api/v1/export/csv", v1.ExportCSV)
	http.HandleFunc("/api/v1/export/stats-csv", v1.ExportStatsCSV)
	http.HandleFunc("/api/v1/stats", v1.GetStats)

	// API V2 - Optimisée (avec cache, tri efficace, calculs optimisés)
	http.HandleFunc("/api/v2/export/csv", v2.ExportCSV)
	http.HandleFunc("/api/v2/export/stats-csv", v2.ExportStatsCSV)
	http.HandleFunc("/api/v2/stats", v2.GetStats)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "API v1 (non-opti) et v2 (opti) disponibles",
	})
}
