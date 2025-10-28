package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"eval/database"

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

	years, _ := strconv.Atoi(getEnv("SEED_YEARS", "5"))

	fmt.Println("🌱 Démarrage du seed de la base de données...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	err = database.SeedDatabase(years)
	if err != nil {
		log.Fatal("❌ Erreur lors du seed:", err)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Seed terminé avec succès!")
	fmt.Println()
	fmt.Println("Vous pouvez maintenant démarrer l'application avec:")
	fmt.Println("  go run main.go")
	fmt.Println()
	fmt.Println("Et tester les endpoints:")
	fmt.Println("  V1: http://localhost:8080/api/v1/stats?days=365")
	fmt.Println("  V2: http://localhost:8080/api/v2/stats?days=365")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
