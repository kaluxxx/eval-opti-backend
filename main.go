package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	// API handlers
	apiv1 "eval/api/v1"
	apiv2 "eval/api/v2"

	// Analytics
	analyticsapp "eval/internal/analytics/application"
	analyticsinfra "eval/internal/analytics/infrastructure"

	// Catalog
	cataloginfra "eval/internal/catalog/infrastructure"

	// Export
	exportapp "eval/internal/export/application"
	exportinfra "eval/internal/export/infrastructure"

	// Orders
	ordersinfra "eval/internal/orders/infrastructure"

	// Shared infrastructure
	sharedinfra "eval/internal/shared/infrastructure"
)

// Application contient toutes les dépendances de l'application
type Application struct {
	db *sql.DB

	// Repositories
	productQueryRepo  *cataloginfra.ProductQueryRepository
	orderQueryRepo    *ordersinfra.OrderQueryRepository
	statsQueryRepo    *analyticsinfra.StatsQueryRepository
	exportQueryRepo   *exportinfra.ExportQueryRepository

	// Services
	cache             sharedinfra.Cache
	statsServiceV1    *analyticsapp.StatsServiceV1
	statsServiceV2    *analyticsapp.StatsServiceV2
	exportServiceV1   *exportapp.ExportServiceV1
	exportServiceV2   *exportapp.ExportServiceV2

	// Handlers
	handlersV1 *apiv1.Handlers
	handlersV2 *apiv2.Handlers
}

func main() {
	// Charger les variables d'environnement
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Fichier .env non trouvé, utilisation des valeurs par défaut")
	}

	// Initialiser l'application avec DI
	app, err := initializeApplication()
	if err != nil {
		log.Fatal("❌ Erreur d'initialisation:", err)
	}
	defer app.cleanup()

	fmt.Println("✅ Application DDD initialisée avec succès")

	// Enregistrer les routes
	app.registerRoutes()

	// Démarrer le serveur
	port := getEnv("APP_PORT", "8080")
	app.printBanner(port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// initializeApplication initialise toute l'application avec dependency injection
func initializeApplication() (*Application, error) {
	app := &Application{}

	// 1. Initialiser la connexion DB
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "evaluser"),
		getEnv("DB_PASSWORD", "evalpass"),
		getEnv("DB_NAME", "evaldb"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configuration du pool de connexions
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	app.db = db

	// 2. Initialiser l'infrastructure partagée
	app.cache = sharedinfra.NewShardedCache(16) // 16 shards pour réduire contention

	// 3. Initialiser les repositories
	app.productQueryRepo = cataloginfra.NewProductQueryRepository(db)
	app.orderQueryRepo = ordersinfra.NewOrderQueryRepository(db)
	app.statsQueryRepo = analyticsinfra.NewStatsQueryRepository(db)
	app.exportQueryRepo = exportinfra.NewExportQueryRepository(db)

	// 4. Initialiser les services V1 (non-optimisés)
	app.statsServiceV1 = analyticsapp.NewStatsServiceV1(
		app.statsQueryRepo,
		app.productQueryRepo,
	)
	app.exportServiceV1 = exportapp.NewExportServiceV1(
		app.exportQueryRepo,
		app.statsServiceV1,
	)

	// 5. Initialiser les services V2 (optimisés)
	app.statsServiceV2 = analyticsapp.NewStatsServiceV2(
		app.statsQueryRepo,
		app.cache,
	)
	app.exportServiceV2 = exportapp.NewExportServiceV2(
		app.exportQueryRepo,
		app.statsServiceV2,
	)

	// 6. Initialiser les handlers
	app.handlersV1 = apiv1.NewHandlers(
		app.statsServiceV1,
		app.exportServiceV1,
	)
	app.handlersV2 = apiv2.NewHandlers(
		app.statsServiceV2,
		app.exportServiceV2,
	)

	return app, nil
}

// registerRoutes enregistre toutes les routes HTTP
func (app *Application) registerRoutes() {
	// Health check
	http.HandleFunc("/api/health", app.healthHandler)

	// API V1 - Non-optimisée (DDD)
	http.HandleFunc("/api/v1/stats", app.handlersV1.GetStats)
	http.HandleFunc("/api/v1/export/csv", app.handlersV1.ExportCSV)
	http.HandleFunc("/api/v1/export/stats-csv", app.handlersV1.ExportStatsCSV)
	http.HandleFunc("/api/v1/export/parquet", app.handlersV1.ExportParquet)

	// API V2 - Optimisée (DDD)
	http.HandleFunc("/api/v2/stats", app.handlersV2.GetStats)
	http.HandleFunc("/api/v2/export/csv", app.handlersV2.ExportCSV)
	http.HandleFunc("/api/v2/export/stats-csv", app.handlersV2.ExportStatsCSV)
	http.HandleFunc("/api/v2/export/parquet", app.handlersV2.ExportParquet)
}

// healthHandler retourne le status de l'application
func (app *Application) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "ok",
		"architecture": "DDD - Domain-Driven Design",
		"message":      "V1 (non-optimisé) et V2 (optimisé) avec architecture DDD",
	})
}

// cleanup libère les ressources
func (app *Application) cleanup() {
	if app.exportServiceV2 != nil {
		app.exportServiceV2.Cleanup()
	}
	if app.db != nil {
		app.db.Close()
	}
	fmt.Println("\n✅ Nettoyage des ressources terminé")
}

// printBanner affiche la bannière de démarrage
func (app *Application) printBanner(port string) {
	fmt.Println()
	fmt.Println("🚀 Serveur démarré sur le port", port)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🏛️  ARCHITECTURE: Domain-Driven Design (DDD)")
	fmt.Println()
	fmt.Println("📦 Bounded Contexts:")
	fmt.Println("   • Catalog (Products, Categories, Suppliers)")
	fmt.Println("   • Orders (Orders, OrderItems)")
	fmt.Println("   • Analytics (Stats, Reporting)")
	fmt.Println("   • Export (CSV, Parquet)")
	fmt.Println()
	fmt.Println("🐌 V1: Code Go NON optimisé")
	fmt.Println("   • N+1 queries problem")
	fmt.Println("   • Bubble sort O(n²)")
	fmt.Println("   • Pas de cache")
	fmt.Println("   • Charge tout en mémoire")
	fmt.Println()
	fmt.Println("⚡ V2: Code Go OPTIMISÉ")
	fmt.Println("   • Agrégations SQL optimisées")
	fmt.Println("   • Cache shardé (5 min TTL)")
	fmt.Println("   • Goroutines parallèles")
	fmt.Println("   • Worker pools pour exports")
	fmt.Println("   • Batch processing (1000 rows)")
	fmt.Println()
	fmt.Println("🎯 Patterns DDD:")
	fmt.Println("   • Value Objects (Money, DateRange, Quantity)")
	fmt.Println("   • Entities & Aggregates (Order, Product)")
	fmt.Println("   • Repositories (CQRS pattern)")
	fmt.Println("   • Domain Services")
	fmt.Println("   • Dependency Injection")
	fmt.Println()
	fmt.Println("📊 Infrastructure:")
	fmt.Println("   • PostgreSQL avec indexes")
	fmt.Println("   • Cache en mémoire (16 shards)")
	fmt.Println("   • Worker pool (4 workers)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// getEnv récupère une variable d'environnement avec fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
