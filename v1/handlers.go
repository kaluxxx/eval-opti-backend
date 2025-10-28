package v1

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"eval/database"
)

type OrderItemTemp struct {
	ID        int64
	OrderID   int64
	ProductID int
	Quantity  int
	UnitPrice float64
	Subtotal  float64
}

type ProductWithCategories struct {
	ID         int
	Name       string
	Categories []string
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V1] 🐌 === DÉBUT CALCUL STATS (NON OPTIMISÉ - N+1) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

	fmt.Printf("[V1] ⏳ Chargement des order_items (≥ %d jours)...\n", days)
	loadStart := time.Now()

	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1
	`

	rows, err := database.DB.Query(query, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orderItems []OrderItemTemp
	for rows.Next() {
		var oi OrderItemTemp
		err := rows.Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orderItems = append(orderItems, oi)
	}

	fmt.Printf("[V1] 📦 %d lignes de commande chargées en %v\n", len(orderItems), time.Since(loadStart))

	fmt.Println("[V1] 🐌 Récupération des produits (N+1 problem)...")
	fetchStart := time.Now()

	productsMap := make(map[int]ProductWithCategories)

	for _, oi := range orderItems {
		if _, exists := productsMap[oi.ProductID]; !exists {
			var productName string
			err := database.DB.QueryRow("SELECT name FROM products WHERE id = $1", oi.ProductID).Scan(&productName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			catRows, err := database.DB.Query(`
				SELECT c.name
				FROM categories c
				INNER JOIN product_categories pc ON c.id = pc.category_id
				WHERE pc.product_id = $1
			`, oi.ProductID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var categories []string
			for catRows.Next() {
				var catName string
				catRows.Scan(&catName)
				categories = append(categories, catName)
			}
			catRows.Close()

			productsMap[oi.ProductID] = ProductWithCategories{
				ID:         oi.ProductID,
				Name:       productName,
				Categories: categories,
			}
		}
	}

	fmt.Printf("[V1] 📦 %d produits récupérés en %v\n", len(productsMap), time.Since(fetchStart))

	stats := calculateStatsInefficient(orderItems, productsMap)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("[V1] 🏁 Durée totale: %v\n", time.Since(start))
	fmt.Println("[V1] === FIN CALCUL STATS ===")
	fmt.Println()
}

func calculateStatsInefficient(orderItems []OrderItemTemp, productsMap map[int]ProductWithCategories) database.Stats {
	fmt.Println("[V1] 🐌 Calcul des stats avec algorithmes inefficaces...")

	stats := database.Stats{
		ParCategorie: make(map[string]database.CategoryStats),
	}

	fmt.Println("[V1]    Boucle 1: Calcul CA total...")
	totalCA := 0.0
	for _, oi := range orderItems {
		totalCA += oi.Subtotal
	}
	stats.TotalCA = totalCA
	stats.NbVentes = len(orderItems)
	if len(orderItems) > 0 {
		stats.MoyenneVente = totalCA / float64(len(orderItems))
	}

	fmt.Println("[V1]    Boucles multiples: Stats par catégorie...")
	categorySet := make(map[string]bool)
	for _, product := range productsMap {
		for _, cat := range product.Categories {
			categorySet[cat] = true
		}
	}

	for cat := range categorySet {
		fmt.Printf("[V1]       Calcul pour catégorie '%s'\n", cat)
		caCategorie := 0.0
		count := 0

		for _, oi := range orderItems {
			product := productsMap[oi.ProductID]
			hasCategory := false
			for _, c := range product.Categories {
				if c == cat {
					hasCategory = true
					break
				}
			}

			if hasCategory {
				caCategorie += oi.Subtotal
				count++
			}
		}

		stats.ParCategorie[cat] = database.CategoryStats{
			CA:       caCategorie,
			NbVentes: count,
		}
	}

	fmt.Println("[V1]    Boucle 3: Calcul CA par produit...")
	productsCA := make(map[int]struct {
		Name     string
		CA       float64
		NbVentes int
	})

	for _, oi := range orderItems {
		product := productsMap[oi.ProductID]
		existing := productsCA[oi.ProductID]
		productsCA[oi.ProductID] = struct {
			Name     string
			CA       float64
			NbVentes int
		}{
			Name:     product.Name,
			CA:       existing.CA + oi.Subtotal,
			NbVentes: existing.NbVentes + 1,
		}
	}

	fmt.Println("[V1]    🐌 Tri avec bubble sort O(n²)...")
	productsList := make([]database.ProductStat, 0, len(productsCA))
	for productID, data := range productsCA {
		productsList = append(productsList, database.ProductStat{
			ProductID:   productID,
			ProductName: data.Name,
			CA:          data.CA,
			NbVentes:    data.NbVentes,
		})
	}

	sortStart := time.Now()
	n := len(productsList)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if productsList[j].CA < productsList[j+1].CA {
				productsList[j], productsList[j+1] = productsList[j+1], productsList[j]
			}
		}
	}
	fmt.Printf("[V1]       Tri terminé en %v\n", time.Since(sortStart))

	// Top 10
	if len(productsList) > 10 {
		stats.TopProduits = productsList[:10]
	} else {
		stats.TopProduits = productsList
	}

	return stats
}

func ExportCSV(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V1] 🐌 === DÉBUT EXPORT CSV (NON OPTIMISÉ) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal, o.order_date
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1
		ORDER BY o.order_date DESC
	`

	rows, err := database.DB.Query(query, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"Date", "Commande ID", "Produit", "Quantité", "Prix Unitaire", "Sous-total"}
	writer.Write(header)

	count := 0
	for rows.Next() {
		var id int64
		var orderID int64
		var productID int
		var quantity int
		var unitPrice float64
		var subtotal float64
		var orderDate time.Time

		rows.Scan(&id, &orderID, &productID, &quantity, &unitPrice, &subtotal, &orderDate)

		var productName string
		database.DB.QueryRow("SELECT name FROM products WHERE id = $1", productID).Scan(&productName)

		row := []string{
			orderDate.Format("2006-01-02"),
			strconv.FormatInt(orderID, 10),
			productName,
			strconv.Itoa(quantity),
			fmt.Sprintf("%.2f", unitPrice),
			fmt.Sprintf("%.2f", subtotal),
		}
		writer.Write(row)

		count++
	}

	writer.Flush()

	fmt.Println("[V1] ⏳ Post-traitement...")

	fmt.Printf("[V1] 🏁 Export terminé: %d lignes en %v\n", count, time.Since(start))
	fmt.Println("[V1] === FIN EXPORT CSV ===")
	fmt.Println()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_v1.csv")
	w.Write(buf.Bytes())
}

func ExportStatsCSV(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	fmt.Println("[V1] 🐌 === DÉBUT EXPORT STATS CSV ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1
	`

	rows, err := database.DB.Query(query, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orderItems []OrderItemTemp
	for rows.Next() {
		var oi OrderItemTemp
		rows.Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal)
		orderItems = append(orderItems, oi)
	}

	productsMap := make(map[int]ProductWithCategories)
	for _, oi := range orderItems {
		if _, exists := productsMap[oi.ProductID]; !exists {
			var productName string
			database.DB.QueryRow("SELECT name FROM products WHERE id = $1", oi.ProductID).Scan(&productName)

			catRows, _ := database.DB.Query(`
				SELECT c.name
				FROM categories c
				INNER JOIN product_categories pc ON c.id = pc.category_id
				WHERE pc.product_id = $1
			`, oi.ProductID)

			var categories []string
			for catRows.Next() {
				var catName string
				catRows.Scan(&catName)
				categories = append(categories, catName)
			}
			catRows.Close()

			productsMap[oi.ProductID] = ProductWithCategories{
				ID:         oi.ProductID,
				Name:       productName,
				Categories: categories,
			}
		}
	}

	stats := calculateStatsInefficient(orderItems, productsMap)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	writer.Write([]string{"STATISTIQUES GLOBALES"})
	writer.Write([]string{"Métrique", "Valeur"})
	writer.Write([]string{"CA Total", fmt.Sprintf("%.2f", stats.TotalCA)})
	writer.Write([]string{"Nombre de ventes", strconv.Itoa(stats.NbVentes)})
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

	writer.Flush()

	fmt.Println("[V1] === FIN EXPORT STATS CSV ===")
	fmt.Println()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=stats_v1.csv")
	w.Write(buf.Bytes())
}

func ExportParquet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V1] 🐌 === DÉBUT EXPORT PARQUET (NON OPTIMISÉ) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

	fmt.Println("[V1] 🐌 Chargement de TOUTES les données en mémoire...")
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal,
		       o.order_date, o.customer_id, o.store_id, o.payment_method_id
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1
		ORDER BY o.order_date DESC
	`

	rows, err := database.DB.Query(query, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TempRow struct {
		OrderItemID     int64
		OrderID         int64
		ProductID       int
		Quantity        int
		UnitPrice       float64
		Subtotal        float64
		OrderDate       time.Time
		CustomerID      int
		StoreID         int
		PaymentMethodID int
	}

	var allRows []TempRow
	for rows.Next() {
		var row TempRow
		rows.Scan(&row.OrderItemID, &row.OrderID, &row.ProductID, &row.Quantity,
			&row.UnitPrice, &row.Subtotal, &row.OrderDate, &row.CustomerID,
			&row.StoreID, &row.PaymentMethodID)
		allRows = append(allRows, row)
	}

	fmt.Printf("[V1] 📦 %d lignes chargées en mémoire (%v)\n", len(allRows), time.Since(start))

	fmt.Println("[V1] 🐌 Récupération des informations (N+1)...")
	productsMap := make(map[int]string)
	customersMap := make(map[int]string)
	storesMap := make(map[int]struct{ Name, City string })
	paymentMethodsMap := make(map[int]string)

	count := 0
	for _, row := range allRows {
		// Produit
		if _, exists := productsMap[row.ProductID]; !exists {
			var name string
			database.DB.QueryRow("SELECT name FROM products WHERE id = $1", row.ProductID).Scan(&name)
			productsMap[row.ProductID] = name
		}

		// Client
		if _, exists := customersMap[row.CustomerID]; !exists {
			var firstName, lastName string
			database.DB.QueryRow("SELECT first_name, last_name FROM customers WHERE id = $1", row.CustomerID).
				Scan(&firstName, &lastName)
			customersMap[row.CustomerID] = firstName + " " + lastName
		}

		// Magasin
		if _, exists := storesMap[row.StoreID]; !exists {
			var name, city string
			database.DB.QueryRow("SELECT name, city FROM stores WHERE id = $1", row.StoreID).Scan(&name, &city)
			storesMap[row.StoreID] = struct{ Name, City string }{name, city}
		}

		// Méthode de paiement
		if _, exists := paymentMethodsMap[row.PaymentMethodID]; !exists {
			var name string
			database.DB.QueryRow("SELECT name FROM payment_methods WHERE id = $1", row.PaymentMethodID).Scan(&name)
			paymentMethodsMap[row.PaymentMethodID] = name
		}

		count++
	}

	fmt.Printf("[V1] 📦 Informations récupérées: %d produits, %d clients, %d magasins\n",
		len(productsMap), len(customersMap), len(storesMap))

	fmt.Println("[V1] 🐌 Conversion en structures Parquet (tout en mémoire)...")
	parquetRows := make([]database.SaleParquet, len(allRows))
	for i, row := range allRows {
		store := storesMap[row.StoreID]
		parquetRows[i] = database.SaleParquet{
			OrderDate:     row.OrderDate.Format("2006-01-02"),
			OrderID:       row.OrderID,
			ProductName:   productsMap[row.ProductID],
			CustomerName:  customersMap[row.CustomerID],
			StoreName:     store.Name,
			StoreCity:     store.City,
			PaymentMethod: paymentMethodsMap[row.PaymentMethodID],
			Quantity:      int32(row.Quantity),
			UnitPrice:     row.UnitPrice,
			Subtotal:      row.Subtotal,
		}
	}

	fmt.Println("[V1] ⏳ Post-traitement...")

	fmt.Printf("[V1] 🏁 Export Parquet terminé: %d lignes en %v\n", len(parquetRows), time.Since(start))
	fmt.Printf("[V1] ⚠️  Mémoire utilisée: ~%d MB (estimation)\n", (len(parquetRows)*200)/1024/1024)
	fmt.Println("[V1] === FIN EXPORT PARQUET ===")
	fmt.Println()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_v1.parquet")

	w.Write([]byte(fmt.Sprintf("V1 Parquet export: %d rows processed in %v", len(parquetRows), time.Since(start))))
}
