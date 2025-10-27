package v1

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
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

// generateFakeSalesData génère des données de ventes - APPELÉ À CHAQUE REQUÊTE
func generateFakeSalesData(days int) []Sale {
	start := time.Now()
	fmt.Printf("[V1] ⏳ Génération de %d jours de données...\n", days)

	categories := []string{"Électronique", "Vêtements", "Alimentation", "Maison", "Sport"}
	var sales []Sale

	// Génère beaucoup de données - non optimisé
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

	fmt.Printf("[V1] ✅ %d ventes générées en %v\n", len(sales), time.Since(start))
	return sales
}

// calculateStatistics calcule les stats de manière TRÈS inefficace
func calculateStatistics(sales []Sale) Stats {
	start := time.Now()
	fmt.Printf("[V1] 📊 Calcul des statistiques sur %d ventes...\n", len(sales))

	stats := Stats{
		ParCategorie: make(map[string]CategoryStats),
	}

	// Calcul du CA total - boucle manuelle au lieu d'utiliser une optimisation
	totalCA := 0.0
	for _, sale := range sales {
		totalCA += float64(sale.Quantity) * sale.Price
	}
	stats.TotalCA = totalCA
	stats.NbVentes = len(sales)
	stats.MoyenneVente = totalCA / float64(len(sales))

	// Calcul par catégorie - TRÈS inefficace, on reboucle plusieurs fois
	categories := []string{"Électronique", "Vêtements", "Alimentation", "Maison", "Sport"}
	for _, cat := range categories {
		caCategorie := 0.0
		count := 0

		// Boucle sur TOUTES les ventes pour chaque catégorie
		for _, sale := range sales {
			if sale.Category == cat {
				caCategorie += float64(sale.Quantity) * sale.Price
				count++
			}
		}

		stats.ParCategorie[cat] = CategoryStats{
			CA:       caCategorie,
			NbVentes: count,
		}
	}

	// Top produits - algorithme O(n²) avec bubble sort
	productsCA := make(map[string]float64)
	for _, sale := range sales {
		productsCA[sale.Product] += float64(sale.Quantity) * sale.Price
	}

	// Conversion en slice pour trier
	productsList := make([]ProductStat, 0, len(productsCA))
	for product, ca := range productsCA {
		productsList = append(productsList, ProductStat{Product: product, CA: ca})
	}

	// BUBBLE SORT - le pire algorithme de tri possible !
	n := len(productsList)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if productsList[j].CA < productsList[j+1].CA {
				productsList[j], productsList[j+1] = productsList[j+1], productsList[j]
			}
		}
	}

	// Prend le top 10
	if len(productsList) > 10 {
		stats.TopProduits = productsList[:10]
	} else {
		stats.TopProduits = productsList
	}

	fmt.Printf("[V1] ✅ Statistiques calculées en %v\n", time.Since(start))
	return stats
}

// ExportCSV exporte TOUTES les ventes en CSV - TRÈS BLOQUANT
func ExportCSV(w http.ResponseWriter, r *http.Request) {
	startTotal := time.Now()
	fmt.Println("\n[V1] 🔥 === DÉBUT EXPORT CSV COMPLET ===")

	// Parse le paramètre days
	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	// GÉNÈRE les données À CHAQUE REQUÊTE (pas de cache)
	sales := generateFakeSalesData(days)

	fmt.Printf("[V1] 📝 Écriture de %d lignes dans le CSV...\n", len(sales))
	startWrite := time.Now()

	// Crée le CSV EN MÉMOIRE (buffer complet)
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Écrit l'en-tête
	header := []string{"Date", "Produit", "Quantité", "Prix", "Client", "Catégorie", "CA Ligne"}
	writer.Write(header)

	// Écrit TOUTES les lignes UNE PAR UNE (pas de batch)
	for i, sale := range sales {
		// Calcul du CA pour chaque ligne individuellement
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

		// Simule un traitement lent (validation, formatage, etc.)
		// Pour chaque 1000 lignes, on fait une micro-pause
		if (i+1)%1000 == 0 {
			time.Sleep(10 * time.Millisecond)
			fmt.Printf("[V1]    ... %d lignes écrites\n", i+1)
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		http.Error(w, "Erreur lors de l'écriture du CSV", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[V1] ✅ CSV écrit en %v\n", time.Since(startWrite))

	// Simule un post-traitement (compression, validation, etc.)
	fmt.Println("[V1] ⏳ Post-traitement du fichier...")
	time.Sleep(2 * time.Second)

	fmt.Printf("[V1] 🏁 DURÉE TOTALE: %v\n", time.Since(startTotal))
	fmt.Printf("[V1] 📦 Taille du fichier: %d octets\n", buf.Len())
	fmt.Println("[V1] === FIN EXPORT CSV ===\n")

	// Envoie le CSV
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_export_v1.csv")
	w.Write(buf.Bytes())
}

// ExportStatsCSV exporte les statistiques agrégées en CSV
func ExportStatsCSV(w http.ResponseWriter, r *http.Request) {
	startTotal := time.Now()
	fmt.Println("\n[V1] 📊 === DÉBUT EXPORT STATS CSV ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	// Génère et calcule tout de zéro
	sales := generateFakeSalesData(days)
	stats := calculateStatistics(sales)

	fmt.Println("[V1] 📝 Écriture du CSV des statistiques...")

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Section 1: Stats globales
	writer.Write([]string{"STATISTIQUES GLOBALES"})
	writer.Write([]string{"Métrique", "Valeur"})
	writer.Write([]string{"CA Total", fmt.Sprintf("%.2f", stats.TotalCA)})
	writer.Write([]string{"Nombre de ventes", strconv.Itoa(stats.NbVentes)})
	writer.Write([]string{"Moyenne par vente", fmt.Sprintf("%.2f", stats.MoyenneVente)})
	writer.Write([]string{}) // Ligne vide

	// Section 2: Stats par catégorie
	writer.Write([]string{"STATISTIQUES PAR CATÉGORIE"})
	writer.Write([]string{"Catégorie", "CA", "Nombre de ventes"})

	// Trie les catégories (encore un tri inutile !)
	type catSort struct {
		name string
		stat CategoryStats
	}
	catList := make([]catSort, 0, len(stats.ParCategorie))
	for name, stat := range stats.ParCategorie {
		catList = append(catList, catSort{name, stat})
	}

	// Bubble sort again !
	n := len(catList)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if catList[j].stat.CA < catList[j+1].stat.CA {
				catList[j], catList[j+1] = catList[j+1], catList[j]
			}
		}
	}

	for _, cat := range catList {
		writer.Write([]string{
			cat.name,
			fmt.Sprintf("%.2f", cat.stat.CA),
			strconv.Itoa(cat.stat.NbVentes),
		})
	}
	writer.Write([]string{}) // Ligne vide

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

	// Simule un traitement supplémentaire
	time.Sleep(1 * time.Second)

	fmt.Printf("[V1] 🏁 DURÉE TOTALE: %v\n", time.Since(startTotal))
	fmt.Println("[V1] === FIN EXPORT STATS CSV ===\n")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=statistiques_v1.csv")
	w.Write(buf.Bytes())
}

// GetStats retourne uniquement les statistiques en JSON
func GetStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	sales := generateFakeSalesData(days)
	stats := calculateStatistics(sales)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)

	fmt.Printf("[V1] ⚡ Stats générées en %v\n", time.Since(start))
}
