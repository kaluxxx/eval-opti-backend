package testhelpers

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	analyticsinfra "eval/internal/analytics/infrastructure"
	cataloginfra "eval/internal/catalog/infrastructure"
	exportinfra "eval/internal/export/infrastructure"
	sharedinfra "eval/internal/shared/infrastructure"
)

// TestContext contient toutes les dépendances pour les tests d'intégration
// Note: Ne contient PAS les services pour éviter les import cycles
// Les tests doivent créer leurs propres services en utilisant ce contexte
type TestContext struct {
	DB *sql.DB

	// Repositories
	ProductQueryRepo *cataloginfra.ProductQueryRepository
	StatsQueryRepo   *analyticsinfra.StatsQueryRepository
	ExportQueryRepo  *exportinfra.ExportQueryRepository

	// Infrastructure
	Cache sharedinfra.Cache
}

// SetupTestDB initialise une connexion à la base de données de test
func SetupTestDB(tb testing.TB) *sql.DB {
	tb.Helper()

	// Charger les variables d'environnement
	_ = godotenv.Load("../../.env")

	// Construire la connection string
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
		tb.Fatalf("Failed to open database: %v", err)
	}

	// Configuration du pool de connexions (optimisé pour tests)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		tb.Fatalf("Failed to ping database: %v\nConnection string: %s", err, hidePassword(connStr))
	}

	return db
}

// SetupTestContext initialise un contexte de test avec DB et repositories
// Les services doivent être créés par les tests eux-mêmes pour éviter les import cycles
func SetupTestContext(tb testing.TB) *TestContext {
	tb.Helper()

	ctx := &TestContext{}

	// 1. Initialiser la connexion DB
	ctx.DB = SetupTestDB(tb)

	// 2. Initialiser l'infrastructure partagée
	ctx.Cache = sharedinfra.NewShardedCache(16)

	// 3. Initialiser les repositories
	ctx.ProductQueryRepo = cataloginfra.NewProductQueryRepository(ctx.DB)
	ctx.StatsQueryRepo = analyticsinfra.NewStatsQueryRepository(ctx.DB)
	ctx.ExportQueryRepo = exportinfra.NewExportQueryRepository(ctx.DB)

	return ctx
}

// Cleanup libère les ressources du contexte de test
func (ctx *TestContext) Cleanup() {
	if ctx.DB != nil {
		ctx.DB.Close()
	}
}

// ClearCache vide le cache (utile entre les benchmarks)
func (ctx *TestContext) ClearCache() {
	if ctx.Cache != nil {
		ctx.Cache.Clear()
	}
}

// getEnv récupère une variable d'environnement avec fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// hidePassword masque le mot de passe dans la connection string pour les logs
func hidePassword(connStr string) string {
	// Simple masquage pour ne pas exposer le mot de passe dans les logs
	return "host=... (password hidden)"
}

// SkipIfNoDatabase skip le test/benchmark si la DB n'est pas disponible
func SkipIfNoDatabase(tb testing.TB) {
	tb.Helper()

	_ = godotenv.Load("../../.env")

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
		tb.Skip("Database not available:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		tb.Skip("Database not available:", err)
	}
}
