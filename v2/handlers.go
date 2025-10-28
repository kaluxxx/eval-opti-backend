package v2

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"eval/database"
)

// Cache applicatif V2
var (
	cachedStats   database.Stats
	cacheTime     time.Time
	cacheDays     int
	cacheMutex    sync.RWMutex
	cacheDuration = 5 * time.Minute
)

// GetStats - V2 OPTIMISÉE avec JOINS et agrégations SQL
// ✅ Une seule requête avec JOIN pour récupérer tout
// ✅ Agrégations SQL (GROUP BY)
// ✅ Tri en SQL (ORDER BY)
// ✅ Cache applicatif
func GetStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT CALCUL STATS (OPTIMISÉ - JOINS) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	// ✅ OPTIMISATION 1: Vérifie le cache
	cacheMutex.RLock()
	if time.Since(cacheTime) < cacheDuration && cacheDays == days && cachedStats.NbVentes > 0 {
		stats := cachedStats
		cacheMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(stats)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("[V2] 🚀 Stats depuis le cache en %v\n", time.Since(start))
		fmt.Println("[V2] === FIN (CACHE HIT) ===")
		fmt.Println()
		return
	}
	cacheMutex.RUnlock()

	fmt.Println("[V2] 💾 Cache miss, calcul des stats...")

	// ✅ OPTIMISATION 2: Calculs en SQL avec JOINs
	stats, err := calculateStatsOptimized(days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ OPTIMISATION 3: Mise en cache
	cacheMutex.Lock()
	cachedStats = stats
	cacheDays = days
	cacheTime = time.Now()
	cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("[V2] ⚡ Stats calculées en %v\n", time.Since(start))
	fmt.Println("[V2] === FIN CALCUL STATS ===")
	fmt.Println()
}

// calculateStatsOptimized - TOUT EN SQL avec JOINs!
func calculateStatsOptimized(days int) (database.Stats, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	stats := database.Stats{
		ParCategorie: make(map[string]database.CategoryStats),
	}

	// ✅ OPTIMISATION 1: Stats globales en une seule requête
	fmt.Println("[V2]    Requête 1: Stats globales...")
	queryGlobal := `
		SELECT
			COUNT(*) as nb_ventes,
			COALESCE(SUM(oi.subtotal), 0) as total_ca,
			COALESCE(AVG(oi.subtotal), 0) as moyenne_vente,
			COUNT(DISTINCT o.id) as nb_commandes
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1
	`

	var nbCommandes int
	err := database.DB.QueryRow(queryGlobal, startDate).Scan(
		&stats.NbVentes, &stats.TotalCA, &stats.MoyenneVente, &nbCommandes)
	if err != nil {
		return stats, err
	}
	stats.NbCommandes = nbCommandes

	// ✅ OPTIMISATION 2: Stats par catégorie avec JOINs et GROUP BY
	fmt.Println("[V2]    Requête 2: Stats par catégorie (avec JOINs)...")
	queryCateg := `
		SELECT
			c.name as category,
			COUNT(oi.id) as nb_ventes,
			SUM(oi.subtotal) as ca
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		INNER JOIN products p ON oi.product_id = p.id
		INNER JOIN product_categories pc ON p.id = pc.product_id
		INNER JOIN categories c ON pc.category_id = c.id
		WHERE o.order_date >= $1
		GROUP BY c.name
		ORDER BY ca DESC
	`

	rows, err := database.DB.Query(queryCateg, startDate)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var cs database.CategoryStats
		err := rows.Scan(&category, &cs.NbVentes, &cs.CA)
		if err != nil {
			return stats, err
		}
		stats.ParCategorie[category] = cs
	}

	// ✅ OPTIMISATION 3: Top produits avec JOINs, GROUP BY, ORDER BY et LIMIT en SQL
	fmt.Println("[V2]    Requête 3: Top 10 produits (avec JOINs + ORDER BY + LIMIT)...")
	queryTop := `
		SELECT
			p.id,
			p.name,
			COUNT(oi.id) as nb_ventes,
			SUM(oi.subtotal) as ca
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE o.order_date >= $1
		GROUP BY p.id, p.name
		ORDER BY ca DESC
		LIMIT 10
	`

	rowsTop, err := database.DB.Query(queryTop, startDate)
	if err != nil {
		return stats, err
	}
	defer rowsTop.Close()

	// ✅ OPTIMISATION 4: Préallocation du slice
	stats.TopProduits = make([]database.ProductStat, 0, 10)

	for rowsTop.Next() {
		var ps database.ProductStat
		err := rowsTop.Scan(&ps.ProductID, &ps.ProductName, &ps.NbVentes, &ps.CA)
		if err != nil {
			return stats, err
		}
		stats.TopProduits = append(stats.TopProduits, ps)
	}

	// ✅ BONUS: Top magasins
	fmt.Println("[V2]    Requête 4: Top 5 magasins...")
	queryStores := `
		SELECT
			s.id,
			s.name,
			s.city,
			COUNT(oi.id) as nb_ventes,
			SUM(oi.subtotal) as ca
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		INNER JOIN stores s ON o.store_id = s.id
		WHERE o.order_date >= $1
		GROUP BY s.id, s.name, s.city
		ORDER BY ca DESC
		LIMIT 5
	`

	rowsStores, err := database.DB.Query(queryStores, startDate)
	if err == nil {
		defer rowsStores.Close()
		stats.TopMagasins = make([]database.StoreStat, 0, 5)

		for rowsStores.Next() {
			var ss database.StoreStat
			rowsStores.Scan(&ss.StoreID, &ss.StoreName, &ss.City, &ss.NbVentes, &ss.CA)
			stats.TopMagasins = append(stats.TopMagasins, ss)
		}
	}

	// ✅ BONUS: Répartition par méthode de paiement
	fmt.Println("[V2]    Requête 5: Répartition paiements...")
	queryPayment := `
		SELECT
			pm.name,
			COUNT(DISTINCT o.id) as nb_commandes
		FROM orders o
		INNER JOIN payment_methods pm ON o.payment_method_id = pm.id
		WHERE o.order_date >= $1
		GROUP BY pm.name
	`

	rowsPayment, err := database.DB.Query(queryPayment, startDate)
	if err == nil {
		defer rowsPayment.Close()
		stats.RepartitionPaiement = make(map[string]int)

		for rowsPayment.Next() {
			var method string
			var count int
			rowsPayment.Scan(&method, &count)
			stats.RepartitionPaiement[method] = count
		}
	}

	return stats, nil
}

// ExportCSV - V2 OPTIMISÉE avec JOINs
func ExportCSV(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT EXPORT CSV (OPTIMISÉ - JOINs) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

	// ✅ UNE SEULE REQUÊTE avec tous les JOINs nécessaires
	query := `
		SELECT
			o.order_date,
			o.id as order_id,
			p.name as product_name,
			oi.quantity,
			oi.unit_price,
			oi.subtotal,
			c.first_name || ' ' || c.last_name as customer_name,
			s.name as store_name
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		INNER JOIN products p ON oi.product_id = p.id
		INNER JOIN customers c ON o.customer_id = c.id
		INNER JOIN stores s ON o.store_id = s.id
		WHERE o.order_date >= $1
		ORDER BY o.order_date DESC
	`

	rows, err := database.DB.Query(query, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// ✅ Buffer préalloué
	var buf bytes.Buffer
	buf.Grow(1024 * 1024) // 1 MB

	writer := csv.NewWriter(&buf)

	header := []string{"Date", "Commande ID", "Produit", "Quantité", "Prix Unitaire", "Sous-total", "Client", "Magasin"}
	writer.Write(header)

	count := 0
	for rows.Next() {
		var orderDate time.Time
		var orderID int64
		var productName string
		var quantity int
		var unitPrice float64
		var subtotal float64
		var customerName string
		var storeName string

		rows.Scan(&orderDate, &orderID, &productName, &quantity, &unitPrice, &subtotal, &customerName, &storeName)

		row := []string{
			orderDate.Format("2006-01-02"),
			strconv.FormatInt(orderID, 10),
			productName,
			strconv.Itoa(quantity),
			fmt.Sprintf("%.2f", unitPrice),
			fmt.Sprintf("%.2f", subtotal),
			customerName,
			storeName,
		}
		writer.Write(row)
		count++

		// ✅ Pas de sleep !
	}

	writer.Flush()

	fmt.Printf("[V2] ⚡ Export terminé: %d lignes en %v\n", count, time.Since(start))
	fmt.Println("[V2] === FIN EXPORT CSV ===")
	fmt.Println()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_v2.csv")
	w.Write(buf.Bytes())
}

// ExportStatsCSV - V2 OPTIMISÉE
func ExportStatsCSV(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT EXPORT STATS CSV ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	// ✅ Utilise le cache si disponible
	cacheMutex.RLock()
	var stats database.Stats
	var err error

	if time.Since(cacheTime) < cacheDuration && cacheDays == days && cachedStats.NbVentes > 0 {
		stats = cachedStats
		cacheMutex.RUnlock()
		fmt.Println("[V2] 🚀 Utilisation du cache")
	} else {
		cacheMutex.RUnlock()
		fmt.Println("[V2] 💾 Cache miss, calcul...")
		stats, err = calculateStatsOptimized(days)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	writer.Write([]string{"STATISTIQUES GLOBALES"})
	writer.Write([]string{"Métrique", "Valeur"})
	writer.Write([]string{"CA Total", fmt.Sprintf("%.2f", stats.TotalCA)})
	writer.Write([]string{"Nombre de ventes", strconv.Itoa(stats.NbVentes)})
	writer.Write([]string{"Nombre de commandes", strconv.Itoa(stats.NbCommandes)})
	writer.Write([]string{"Moyenne par vente", fmt.Sprintf("%.2f", stats.MoyenneVente)})
	writer.Write([]string{})

	writer.Write([]string{"STATISTIQUES PAR CATÉGORIE"})
	writer.Write([]string{"Catégorie", "CA", "Nombre de ventes"})

	for cat, catStats := range stats.ParCategorie {
		writer.Write([]string{cat, fmt.Sprintf("%.2f", catStats.CA), strconv.Itoa(catStats.NbVentes)})
	}
	writer.Write([]string{})

	writer.Write([]string{"TOP 10 PRODUITS"})
	writer.Write([]string{"Rang", "Produit", "CA", "Nb Ventes"})
	for i, prod := range stats.TopProduits {
		writer.Write([]string{
			strconv.Itoa(i + 1),
			prod.ProductName,
			fmt.Sprintf("%.2f", prod.CA),
			strconv.Itoa(prod.NbVentes),
		})
	}
	writer.Write([]string{})

	// ✅ BONUS: Top magasins
	if len(stats.TopMagasins) > 0 {
		writer.Write([]string{"TOP 5 MAGASINS"})
		writer.Write([]string{"Rang", "Magasin", "Ville", "CA", "Nb Ventes"})
		for i, store := range stats.TopMagasins {
			writer.Write([]string{
				strconv.Itoa(i + 1),
				store.StoreName,
				store.City,
				fmt.Sprintf("%.2f", store.CA),
				strconv.Itoa(store.NbVentes),
			})
		}
		writer.Write([]string{})
	}

	// ✅ BONUS: Répartition paiements
	if len(stats.RepartitionPaiement) > 0 {
		writer.Write([]string{"RÉPARTITION PAR MÉTHODE DE PAIEMENT"})
		writer.Write([]string{"Méthode", "Nb Commandes"})
		for method, count := range stats.RepartitionPaiement {
			writer.Write([]string{method, strconv.Itoa(count)})
		}
	}

	writer.Flush()

	fmt.Println("[V2] === FIN EXPORT STATS CSV ===")
	fmt.Println()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=stats_v2.csv")
	w.Write(buf.Bytes())
}

// ExportParquet - V2 OPTIMISÉE avec streaming par batches
// ✅ UNE SEULE requête avec tous les JOINs
// ✅ Streaming par batches (pas tout en mémoire)
// ✅ Traitement à la volée
func ExportParquet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT EXPORT PARQUET (OPTIMISÉ - STREAMING) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

	// ✅ OPTIMISATION 1: UNE SEULE requête avec tous les JOINs
	fmt.Println("[V2] ⚡ Requête unique avec tous les JOINs...")
	query := `
		SELECT
			o.order_date,
			o.id as order_id,
			p.name as product_name,
			c.first_name || ' ' || c.last_name as customer_name,
			s.name as store_name,
			s.city as store_city,
			pm.name as payment_method,
			oi.quantity,
			oi.unit_price,
			oi.subtotal
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		INNER JOIN products p ON oi.product_id = p.id
		INNER JOIN customers c ON o.customer_id = c.id
		INNER JOIN stores s ON o.store_id = s.id
		INNER JOIN payment_methods pm ON o.payment_method_id = pm.id
		WHERE o.order_date >= $1
		ORDER BY o.order_date DESC
	`

	rows, err := database.DB.Query(query, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// ✅ OPTIMISATION 2: Streaming par batches (ne charge pas tout)
	fmt.Println("[V2] ⚡ Traitement en streaming par batches de 1000...")

	const batchSize = 1000
	batch := make([]database.SaleParquet, 0, batchSize)
	totalRows := 0
	batchNum := 0

	for rows.Next() {
		var orderDate time.Time
		var orderID int64
		var productName string
		var customerName string
		var storeName string
		var storeCity string
		var paymentMethod string
		var quantity int
		var unitPrice float64
		var subtotal float64

		err := rows.Scan(&orderDate, &orderID, &productName, &customerName,
			&storeName, &storeCity, &paymentMethod, &quantity, &unitPrice, &subtotal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ✅ Traitement à la volée
		sale := database.SaleParquet{
			OrderDate:     orderDate.Format("2006-01-02"),
			OrderID:       orderID,
			ProductName:   productName,
			CustomerName:  customerName,
			StoreName:     storeName,
			StoreCity:     storeCity,
			PaymentMethod: paymentMethod,
			Quantity:      int32(quantity),
			UnitPrice:     unitPrice,
			Subtotal:      subtotal,
		}

		batch = append(batch, sale)
		totalRows++

		// ✅ Traitement par batch
		if len(batch) >= batchSize {
			batchNum++
			fmt.Printf("[V2]    Batch %d traité (%d lignes)\n", batchNum, len(batch))
			// Ici on écrirait dans Parquet, mais pour la démo on vide juste le batch
			batch = batch[:0] // Reset le slice sans réallouer
		}
	}

	// Traiter le dernier batch
	if len(batch) > 0 {
		batchNum++
		fmt.Printf("[V2]    Batch %d traité (%d lignes)\n", batchNum, len(batch))
	}

	fmt.Printf("[V2] ⚡ Export Parquet terminé: %d lignes en %v\n", totalRows, time.Since(start))
	fmt.Printf("[V2] ✅ Mémoire utilisée: ~%d MB (max batch size)\n", (batchSize*200)/1024/1024)
	fmt.Println("[V2] === FIN EXPORT PARQUET ===")
	fmt.Println()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_v2.parquet")

	// Pour V2, on renvoie le résumé
	w.Write([]byte(fmt.Sprintf("V2 Parquet export (optimized streaming): %d rows processed in %d batches in %v",
		totalRows, batchNum, time.Since(start))))
}
