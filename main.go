package main

import (
	"encoding/json"
	"fmt"
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

	fmt.Println("🚀 Serveur démarré sur http://localhost:8080")
	fmt.Println("\n📊 API V1 - Non optimisée (code original):")
	fmt.Println("   GET /api/v1/health - Vérification du serveur")
	fmt.Println("   GET /api/v1/export/csv?days=365 - Export CSV complet (TRÈS BLOQUANT)")
	fmt.Println("   GET /api/v1/export/stats-csv?days=365 - Export CSV des stats")
	fmt.Println("   GET /api/v1/stats?days=365 - Statistiques JSON")
	fmt.Println("\n⚡ API V2 - Optimisée:")
	fmt.Println("   GET /api/v2/export/csv?days=365 - Export CSV complet (avec cache)")
	fmt.Println("   GET /api/v2/export/stats-csv?days=365 - Export CSV des stats (optimisé)")
	fmt.Println("   GET /api/v2/stats?days=365 - Statistiques JSON (avec cache)")
	fmt.Println("\n🔍 Endpoints de profiling (pprof):")
	fmt.Println("   GET /debug/pprof/ - Index du profiler")
	fmt.Println("   GET /debug/pprof/profile?seconds=30 - CPU profile (30s)")
	fmt.Println("   GET /debug/pprof/heap - Memory profile")
	fmt.Println("   GET /debug/pprof/goroutine - Goroutines")
	fmt.Println("   GET /debug/pprof/block - Blocking profile")
	fmt.Println("   GET /debug/pprof/mutex - Mutex contention")
	fmt.Println("\n💡 Optimisations de la V2:")
	fmt.Println("   ✅ Cache des données générées (5 min)")
	fmt.Println("   ✅ Tri efficace O(n log n) au lieu de bubble sort O(n²)")
	fmt.Println("   ✅ Calcul des stats en une seule passe")
	fmt.Println("   ✅ Pas de sleeps artificiels")
	fmt.Println("   ✅ Préallocation des slices")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "API v1 (non-opti) et v2 (opti) disponibles",
	})
}
