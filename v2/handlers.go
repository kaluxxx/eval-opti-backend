package v2

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Sale représente une vente
type Sale struct {
	Date     string  `json:"date"`
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Customer string  `json:"customer"`
	Category string  `json:"category"`
}

// Stats contient les statistiques calculées
type Stats struct {
	TotalCA      float64                  `json:"total_ca"`
	ParCategorie map[string]CategoryStats `json:"par_categorie"`
	TopProduits  []ProductStat            `json:"top_produits"`
	NbVentes     int                      `json:"nb_ventes"`
	MoyenneVente float64                  `json:"moyenne_vente"`
}

type CategoryStats struct {
	CA       float64 `json:"ca"`
	NbVentes int     `json:"nb_ventes"`
}

type ProductStat struct {
	Product string  `json:"product"`
	CA      float64 `json:"ca"`
}

// Cache simple pour les données
var (
	cachedSales      []Sale
	cachedStats      Stats
	cacheTime        time.Time
	cacheDays        int
	cacheMutex       sync.RWMutex
	cacheDuration    = 5 * time.Minute // Cache valide pendant 5 minutes
)

// generateFakeSalesData génère des données de ventes - OPTIMISÉ avec cache
func generateFakeSalesData(days int) []Sale {
	// Vérifie le cache
	cacheMutex.RLock()
	if time.Since(cacheTime) < cacheDuration && cacheDays == days && len(cachedSales) > 0 {
		fmt.Printf("[V2] ⚡ Utilisation du cache (%d ventes)\n", len(cachedSales))
		cacheMutex.RUnlock()
		return cachedSales
	}
	cacheMutex.RUnlock()

	// Génère les données
	start := time.Now()
	fmt.Printf("[V2] ⏳ Génération de %d jours de données...\n", days)

	categories := []string{"Électronique", "Vêtements", "Alimentation", "Maison", "Sport"}

	// OPTIMISATION: Préalloue la capacité du slice
	estimatedSize := days * 100 // estimation de 100 ventes/jour en moyenne
	sales := make([]Sale, 0, estimatedSize)

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		numSales := 50 + rand.Intn(150) // 50-200 ventes par jour

		for j := 0; j < numSales; j++ {
			sale := Sale{
				Date:     date.Format("2006-01-02"),
				Product:  fmt.Sprintf("Produit_%d", rand.Intn(100)+1),
				Quantity: rand.Intn(10) + 1,
				Price:    10 + rand.Float64()*490,
				Customer: fmt.Sprintf("Client_%d", rand.Intn(1000)+1),
				Category: categories[rand.Intn(len(categories))],
			}
			sales = append(sales, sale)
		}
	}

	// Met en cache
	cacheMutex.Lock()
	cachedSales = sales
	cacheDays = days
	cacheTime = time.Now()
	cacheMutex.Unlock()

	fmt.Printf("[V2] ✅ %d ventes générées en %v (mises en cache)\n", len(sales), time.Since(start))
	return sales
}

// calculateStatistics calcule les stats de manière OPTIMISÉE
func calculateStatistics(sales []Sale) Stats {
	start := time.Now()
	fmt.Printf("[V2] 📊 Calcul des statistiques sur %d ventes...\n", len(sales))

	stats := Stats{
		ParCategorie: make(map[string]CategoryStats),
	}

	// OPTIMISATION: Une seule boucle pour tout calculer
	totalCA := 0.0
	productsCA := make(map[string]float64)

	for _, sale := range sales {
		ca := float64(sale.Quantity) * sale.Price
		totalCA += ca

		// Stats par catégorie
		catStats := stats.ParCategorie[sale.Category]
		catStats.CA += ca
		catStats.NbVentes++
		stats.ParCategorie[sale.Category] = catStats

		// CA par produit
		productsCA[sale.Product] += ca
	}

	stats.TotalCA = totalCA
	stats.NbVentes = len(sales)
	if len(sales) > 0 {
		stats.MoyenneVente = totalCA / float64(len(sales))
	}

	// OPTIMISATION: Tri efficace avec sort.Slice au lieu de bubble sort
	productsList := make([]ProductStat, 0, len(productsCA))
	for product, ca := range productsCA {
		productsList = append(productsList, ProductStat{Product: product, CA: ca})
	}

	// Tri avec sort.Slice - O(n log n) au lieu de O(n²)
	sort.Slice(productsList, func(i, j int) bool {
		return productsList[i].CA > productsList[j].CA
	})

	// Prend le top 10
	if len(productsList) > 10 {
		stats.TopProduits = productsList[:10]
	} else {
		stats.TopProduits = productsList
	}

	fmt.Printf("[V2] ✅ Statistiques calculées en %v\n", time.Since(start))
	return stats
}

// getCachedStats retourne les stats en utilisant le cache si disponible
func getCachedStats(days int) Stats {
	cacheMutex.RLock()
	if time.Since(cacheTime) < cacheDuration && cacheDays == days && cachedStats.NbVentes > 0 {
		fmt.Printf("[V2] ⚡ Utilisation des stats en cache\n")
		cacheMutex.RUnlock()
		return cachedStats
	}
	cacheMutex.RUnlock()

	sales := generateFakeSalesData(days)
	stats := calculateStatistics(sales)

	cacheMutex.Lock()
	cachedStats = stats
	cacheMutex.Unlock()

	return stats
}

// ExportCSV exporte TOUTES les ventes en CSV - VERSION OPTIMISÉE
func ExportCSV(w http.ResponseWriter, r *http.Request) {
	startTotal := time.Now()
	fmt.Println("\n[V2] 🚀 === DÉBUT EXPORT CSV OPTIMISÉ ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	// OPTIMISATION: Utilise le cache
	sales := generateFakeSalesData(days)

	fmt.Printf("[V2] 📝 Écriture de %d lignes dans le CSV...\n", len(sales))
	startWrite := time.Now()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Écrit l'en-tête
	header := []string{"Date", "Produit", "Quantité", "Prix", "Client", "Catégorie", "CA Ligne"}
	writer.Write(header)

	// OPTIMISATION: Batch write, pas de sleep artificiel
	for _, sale := range sales {
		caLigne := float64(sale.Quantity) * sale.Price

		row := []string{
			sale.Date,
			sale.Product,
			strconv.Itoa(sale.Quantity),
			fmt.Sprintf("%.2f", sale.Price),
			sale.Customer,
			sale.Category,
			fmt.Sprintf("%.2f", caLigne),
		}

		writer.Write(row)
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		http.Error(w, "Erreur lors de l'écriture du CSV", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[V2] ✅ CSV écrit en %v\n", time.Since(startWrite))

	// OPTIMISATION: Pas de post-traitement inutile

	fmt.Printf("[V2] 🏁 DURÉE TOTALE: %v\n", time.Since(startTotal))
	fmt.Printf("[V2] 📦 Taille du fichier: %d octets\n", buf.Len())
	fmt.Println("[V2] === FIN EXPORT CSV ===\n")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_export_v2.csv")
	w.Write(buf.Bytes())
}

// ExportStatsCSV exporte les statistiques agrégées en CSV - VERSION OPTIMISÉE
func ExportStatsCSV(w http.ResponseWriter, r *http.Request) {
	startTotal := time.Now()
	fmt.Println("\n[V2] 📊 === DÉBUT EXPORT STATS CSV OPTIMISÉ ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	// OPTIMISATION: Utilise le cache
	stats := getCachedStats(days)

	fmt.Println("[V2] 📝 Écriture du CSV des statistiques...")

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Section 1: Stats globales
	writer.Write([]string{"STATISTIQUES GLOBALES"})
	writer.Write([]string{"Métrique", "Valeur"})
	writer.Write([]string{"CA Total", fmt.Sprintf("%.2f", stats.TotalCA)})
	writer.Write([]string{"Nombre de ventes", strconv.Itoa(stats.NbVentes)})
	writer.Write([]string{"Moyenne par vente", fmt.Sprintf("%.2f", stats.MoyenneVente)})
	writer.Write([]string{})

	// Section 2: Stats par catégorie
	writer.Write([]string{"STATISTIQUES PAR CATÉGORIE"})
	writer.Write([]string{"Catégorie", "CA", "Nombre de ventes"})

	// OPTIMISATION: Tri efficace avec sort.Slice
	type catSort struct {
		name string
		stat CategoryStats
	}
	catList := make([]catSort, 0, len(stats.ParCategorie))
	for name, stat := range stats.ParCategorie {
		catList = append(catList, catSort{name, stat})
	}

	// Tri O(n log n)
	sort.Slice(catList, func(i, j int) bool {
		return catList[i].stat.CA > catList[j].stat.CA
	})

	for _, cat := range catList {
		writer.Write([]string{
			cat.name,
			fmt.Sprintf("%.2f", cat.stat.CA),
			strconv.Itoa(cat.stat.NbVentes),
		})
	}
	writer.Write([]string{})

	// Section 3: Top produits
	writer.Write([]string{"TOP 10 PRODUITS"})
	writer.Write([]string{"Rang", "Produit", "CA"})
	for i, prod := range stats.TopProduits {
		writer.Write([]string{
			strconv.Itoa(i + 1),
			prod.Product,
			fmt.Sprintf("%.2f", prod.CA),
		})
	}

	writer.Flush()

	// OPTIMISATION: Pas de sleep inutile

	fmt.Printf("[V2] 🏁 DURÉE TOTALE: %v\n", time.Since(startTotal))
	fmt.Println("[V2] === FIN EXPORT STATS CSV ===\n")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=statistiques_v2.csv")
	w.Write(buf.Bytes())
}

// GetStats retourne uniquement les statistiques en JSON - VERSION OPTIMISÉE
func GetStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	stats := getCachedStats(days)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)

	fmt.Printf("[V2] ⚡ Stats générées en %v\n", time.Since(start))
}
