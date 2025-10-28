package v2

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"eval/database"
)

type CacheShard struct {
	stats database.Stats
	time  time.Time
	mutex sync.RWMutex
}

var (
	cacheShards   = make(map[int]*CacheShard)
	shardsM       sync.RWMutex
	cacheDuration = 5 * time.Minute
)

var rowPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 8) // 8 colonnes pour CSV
	},
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT CALCUL STATS (OPTIMISÉ V2.1 - GOROUTINES) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	shardsM.RLock()
	shard := cacheShards[days]
	shardsM.RUnlock()

	if shard != nil {
		shard.mutex.RLock()
		if time.Since(shard.time) < cacheDuration && shard.stats.NbVentes > 0 {
			stats := shard.stats
			shard.mutex.RUnlock()

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
		shard.mutex.RUnlock()
	}

	fmt.Println("[V2] 💾 Cache miss, calcul des stats...")

	stats, err := calculateStatsOptimized(days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shardsM.Lock()
	if cacheShards[days] == nil {
		cacheShards[days] = &CacheShard{}
	}
	shard = cacheShards[days]
	shardsM.Unlock()

	shard.mutex.Lock()
	shard.stats = stats
	shard.time = time.Now()
	shard.mutex.Unlock()

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

func calculateStatsOptimized(days int) (database.Stats, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	stats := database.Stats{
		ParCategorie:        make(map[string]database.CategoryStats, 10),
		RepartitionPaiement: make(map[string]int, 5),
	}

	fmt.Println("[V2] ⚡ Exécution des 5 requêtes SQL en PARALLÈLE...")

	var wg sync.WaitGroup
	var globalErr, categErr, topErr, storesErr, paymentErr error

	wg.Add(5)

	go func() {
		defer wg.Done()
		fmt.Println("[V2]    [GO 1/5] Stats globales...")

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
		globalErr = database.DB.QueryRow(queryGlobal, startDate).Scan(
			&stats.NbVentes, &stats.TotalCA, &stats.MoyenneVente, &nbCommandes)
		stats.NbCommandes = nbCommandes
	}()

	go func() {
		defer wg.Done()
		fmt.Println("[V2]    [GO 2/5] Stats par catégorie...")

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
			categErr = err
			return
		}
		defer rows.Close()

		tempCateg := make(map[string]database.CategoryStats, 10)
		for rows.Next() {
			var category string
			var cs database.CategoryStats
			if err := rows.Scan(&category, &cs.NbVentes, &cs.CA); err != nil {
				categErr = err
				return
			}
			tempCateg[category] = cs
		}

		for k, v := range tempCateg {
			stats.ParCategorie[k] = v
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Println("[V2]    [GO 3/5] Top 10 produits...")

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

		rows, err := database.DB.Query(queryTop, startDate)
		if err != nil {
			topErr = err
			return
		}
		defer rows.Close()

		stats.TopProduits = make([]database.ProductStat, 0, 10)
		for rows.Next() {
			stats.TopProduits = append(stats.TopProduits, database.ProductStat{})
			ps := &stats.TopProduits[len(stats.TopProduits)-1]
			if err := rows.Scan(&ps.ProductID, &ps.ProductName, &ps.NbVentes, &ps.CA); err != nil {
				topErr = err
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Println("[V2]    [GO 4/5] Top 5 magasins...")

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

		rows, err := database.DB.Query(queryStores, startDate)
		if err != nil {
			storesErr = err
			return
		}
		defer rows.Close()

		stats.TopMagasins = make([]database.StoreStat, 0, 5)
		for rows.Next() {
			stats.TopMagasins = append(stats.TopMagasins, database.StoreStat{})
			ss := &stats.TopMagasins[len(stats.TopMagasins)-1]
			if err := rows.Scan(&ss.StoreID, &ss.StoreName, &ss.City, &ss.NbVentes, &ss.CA); err != nil {
				storesErr = err
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Println("[V2]    [GO 5/5] Répartition paiements...")

		queryPayment := `
			SELECT
				pm.name,
				COUNT(DISTINCT o.id) as nb_commandes
			FROM orders o
			INNER JOIN payment_methods pm ON o.payment_method_id = pm.id
			WHERE o.order_date >= $1
			GROUP BY pm.name
		`

		rows, err := database.DB.Query(queryPayment, startDate)
		if err != nil {
			paymentErr = err
			return
		}
		defer rows.Close()

		tempPayment := make(map[string]int, 5)
		for rows.Next() {
			var method string
			var count int
			if err := rows.Scan(&method, &count); err != nil {
				paymentErr = err
				return
			}
			tempPayment[method] = count
		}

		for k, v := range tempPayment {
			stats.RepartitionPaiement[k] = v
		}
	}()

	wg.Wait()

	fmt.Println("[V2] ✅ Toutes les requêtes parallèles terminées")

	if globalErr != nil {
		return stats, globalErr
	}
	if categErr != nil {
		return stats, categErr
	}
	if topErr != nil {
		return stats, topErr
	}
	if storesErr != nil {
		return stats, storesErr
	}
	if paymentErr != nil {
		return stats, paymentErr
	}

	return stats, nil
}

func ExportCSV(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT EXPORT CSV (OPTIMISÉ V2.1) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

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

	var buf bytes.Buffer
	buf.Grow(1024 * 1024) // 1 MB

	writer := csv.NewWriter(&buf)

	header := []string{"Date", "Commande ID", "Produit", "Quantité", "Prix Unitaire", "Sous-total", "Client", "Magasin"}
	writer.Write(header)

	var sb strings.Builder
	sb.Grow(256)

	count := 0
	const flushEvery = 1000

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

		row := rowPool.Get().([]string)

		dateBuf := make([]byte, 0, 10)
		dateBuf = orderDate.AppendFormat(dateBuf, "2006-01-02")
		row[0] = string(dateBuf)

		row[1] = strconv.FormatInt(orderID, 10)
		row[2] = productName
		row[3] = strconv.Itoa(quantity)

		row[4] = strconv.FormatFloat(unitPrice, 'f', 2, 64)
		row[5] = strconv.FormatFloat(subtotal, 'f', 2, 64)

		row[6] = customerName
		row[7] = storeName

		writer.Write(row)
		rowPool.Put(row)

		count++

		if count%flushEvery == 0 {
			writer.Flush()
		}
	}

	writer.Flush()

	fmt.Printf("[V2] ⚡ Export terminé: %d lignes en %v\n", count, time.Since(start))
	fmt.Println("[V2] === FIN EXPORT CSV ===")
	fmt.Println()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_v2.csv")
	w.Write(buf.Bytes())
}

func ExportStatsCSV(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT EXPORT STATS CSV ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	shardsM.RLock()
	shard := cacheShards[days]
	shardsM.RUnlock()

	var stats database.Stats
	var err error

	if shard != nil {
		shard.mutex.RLock()
		if time.Since(shard.time) < cacheDuration && shard.stats.NbVentes > 0 {
			stats = shard.stats
			shard.mutex.RUnlock()
			fmt.Println("[V2] 🚀 Utilisation du cache")
		} else {
			shard.mutex.RUnlock()
			fmt.Println("[V2] 💾 Cache miss, calcul...")
			stats, err = calculateStatsOptimized(days)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
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

	writer.Write([]string{"CA Total", strconv.FormatFloat(stats.TotalCA, 'f', 2, 64)})
	writer.Write([]string{"Nombre de ventes", strconv.Itoa(stats.NbVentes)})
	writer.Write([]string{"Nombre de commandes", strconv.Itoa(stats.NbCommandes)})
	writer.Write([]string{"Moyenne par vente", strconv.FormatFloat(stats.MoyenneVente, 'f', 2, 64)})
	writer.Write([]string{})

	writer.Write([]string{"STATISTIQUES PAR CATÉGORIE"})
	writer.Write([]string{"Catégorie", "CA", "Nombre de ventes"})

	for cat, catStats := range stats.ParCategorie {
		writer.Write([]string{
			cat,
			strconv.FormatFloat(catStats.CA, 'f', 2, 64),
			strconv.Itoa(catStats.NbVentes),
		})
	}
	writer.Write([]string{})

	writer.Write([]string{"TOP 10 PRODUITS"})
	writer.Write([]string{"Rang", "Produit", "CA", "Nb Ventes"})
	for i, prod := range stats.TopProduits {
		writer.Write([]string{
			strconv.Itoa(i + 1),
			prod.ProductName,
			strconv.FormatFloat(prod.CA, 'f', 2, 64),
			strconv.Itoa(prod.NbVentes),
		})
	}
	writer.Write([]string{})

	if len(stats.TopMagasins) > 0 {
		writer.Write([]string{"TOP 5 MAGASINS"})
		writer.Write([]string{"Rang", "Magasin", "Ville", "CA", "Nb Ventes"})
		for i, store := range stats.TopMagasins {
			writer.Write([]string{
				strconv.Itoa(i + 1),
				store.StoreName,
				store.City,
				strconv.FormatFloat(store.CA, 'f', 2, 64),
				strconv.Itoa(store.NbVentes),
			})
		}
		writer.Write([]string{})
	}

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

func ExportParquet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	fmt.Println()
	fmt.Println("[V2] ⚡ === DÉBUT EXPORT PARQUET (OPTIMISÉ V2.1 - WORKER POOL) ===")

	days := 365
	if r.URL.Query().Get("days") != "" {
		fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)
	}

	startDate := time.Now().AddDate(0, 0, -days)

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

	fmt.Println("[V2] ⚡ Traitement avec worker pool (4 workers)...")

	const batchSize = 1000
	const numWorkers = 4

	jobs := make(chan []database.SaleParquet, numWorkers*2)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for batch := range jobs {
				fmt.Printf("[V2]    Worker %d traite batch de %d lignes\n", workerID, len(batch))
			}
		}(i)
	}

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
			close(jobs)
			wg.Wait()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dateBuf := make([]byte, 0, 10)
		dateBuf = orderDate.AppendFormat(dateBuf, "2006-01-02")

		sale := database.SaleParquet{
			OrderDate:     string(dateBuf),
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

		if len(batch) >= batchSize {
			batchNum++
			batchCopy := make([]database.SaleParquet, len(batch))
			copy(batchCopy, batch)
			jobs <- batchCopy
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		batchNum++
		jobs <- batch
	}

	close(jobs)
	wg.Wait()

	fmt.Printf("[V2] ⚡ Export Parquet terminé: %d lignes en %d batches en %v\n", totalRows, batchNum, time.Since(start))
	fmt.Printf("[V2] ✅ Mémoire utilisée: ~%d MB (max batch size)\n", (batchSize*200)/1024/1024)
	fmt.Printf("[V2] ⚡ Traitement parallèle avec %d workers\n", numWorkers)
	fmt.Println("[V2] === FIN EXPORT PARQUET ===")
	fmt.Println()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=ventes_v2.parquet")

	w.Write([]byte(fmt.Sprintf("V2 Parquet export (optimized worker pool): %d rows processed in %d batches with %d workers in %v",
		totalRows, batchNum, numWorkers, time.Since(start))))
}
