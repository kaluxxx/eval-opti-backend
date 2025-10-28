package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"eval/database"
	"eval/v1"
	"eval/v2"

	"github.com/joho/godotenv"
)

func main() {
	// Charge .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Attention: fichier .env non trouvé, utilisation des valeurs par défaut")
	}

	// Connexion PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "evaluser"),
		getEnv("DB_PASSWORD", "evalpass"),
		getEnv("DB_NAME", "evaldb"),
		getEnv("DB_SSLMODE", "disable"),
	)

	err = database.Init(connStr)
	if err != nil {
		log.Fatal("❌ Erreur connexion DB:", err)
	}
	defer database.Close()

	fmt.Println("✅ Connexion PostgreSQL établie")

	// Routes
	http.HandleFunc("/api/health", healthHandler)

	// V1 - Code Go non optimisé
	http.HandleFunc("/api/v1/stats", v1.GetStats)
	http.HandleFunc("/api/v1/export/csv", v1.ExportCSV)
	http.HandleFunc("/api/v1/export/stats-csv", v1.ExportStatsCSV)
	http.HandleFunc("/api/v1/export/parquet", v1.ExportParquet)

	// V2 - Code Go optimisé
	http.HandleFunc("/api/v2/stats", v2.GetStats)
	http.HandleFunc("/api/v2/export/csv", v2.ExportCSV)
	http.HandleFunc("/api/v2/export/stats-csv", v2.ExportStatsCSV)
	http.HandleFunc("/api/v2/export/parquet", v2.ExportParquet)

	port := getEnv("APP_PORT", "8080")

	fmt.Println()
	fmt.Println("🚀 Serveur démarré sur le port", port)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🐌 V1: Code Go NON optimisé")
	fmt.Println("   • Charge tout en mémoire")
	fmt.Println("   • Boucles multiples + N+1 problem")
	fmt.Println("   • Bubble sort O(n²)")
	fmt.Println("   • Pas de cache")
	fmt.Println("   • Export Parquet: mémoire complète")
	fmt.Println()
	fmt.Println("⚡ V2: Code Go OPTIMISÉ")
	fmt.Println("   • Agrégations SQL avec JOINs")
	fmt.Println("   • Tri en SQL (ORDER BY)")
	fmt.Println("   • Cache applicatif (5 min)")
	fmt.Println("   • Préallocation mémoire")
	fmt.Println("   • Export Parquet: streaming par batches")
	fmt.Println()
	fmt.Println("📊 Base de données PostgreSQL optimisée avec index")
	fmt.Println("📦 Export Parquet analytique disponible")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "V1 (non-optimisé) et V2 (optimisé) avec PostgreSQL",
	})
	if err != nil {
		return
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
