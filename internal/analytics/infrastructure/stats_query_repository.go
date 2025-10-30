package infrastructure

import (
	"database/sql"

	"eval/internal/analytics/domain"
	catalogdomain "eval/internal/catalog/domain"
	ordersdomain "eval/internal/orders/domain"
	shareddomain "eval/internal/shared/domain"
	"eval/internal/shared/infrastructure"
)

// StatsQueryRepository repository pour les statistiques
type StatsQueryRepository struct {
	infrastructure.BaseRepository
}

// NewStatsQueryRepository crée un nouveau repository de stats
func NewStatsQueryRepository(db *sql.DB) *StatsQueryRepository {
	return &StatsQueryRepository{
		BaseRepository: infrastructure.NewBaseRepository(db),
	}
}

// GetGlobalStats récupère les statistiques globales de manière optimisée
func (r *StatsQueryRepository) GetGlobalStats(dateRange shareddomain.DateRange) (shareddomain.Money, int, shareddomain.Money, error) {
	query := `
		SELECT COALESCE(SUM(total_amount), 0) as total_revenue,
		       COALESCE(COUNT(*), 0) as total_orders,
		       COALESCE(AVG(total_amount), 0) as avg_order_value
		FROM orders
		WHERE order_date >= $1 AND order_date <= $2
	`

	var totalRevenue, avgOrderValue float64
	var totalOrders int

	err := r.QueryRow(query, dateRange.Start(), dateRange.End()).Scan(&totalRevenue, &totalOrders, &avgOrderValue)
	if err != nil {
		var emptyMoney shareddomain.Money
		return emptyMoney, 0, emptyMoney, err
	}

	revenue, _ := shareddomain.NewMoney(totalRevenue, "EUR")
	avgOrder, _ := shareddomain.NewMoney(avgOrderValue, "EUR")

	return revenue, totalOrders, avgOrder, nil
}

// GetCategoryStats récupère les statistiques par catégorie (optimisé)
func (r *StatsQueryRepository) GetCategoryStats(dateRange shareddomain.DateRange) ([]*domain.CategoryStats, error) {
	query := `
		SELECT c.id, c.name,
		       COALESCE(SUM(oi.subtotal), 0) as total_revenue,
		       COALESCE(COUNT(DISTINCT o.id), 0) as total_orders
		FROM categories c
		LEFT JOIN product_categories pc ON c.id = pc.category_id
		LEFT JOIN order_items oi ON pc.product_id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id AND o.order_date >= $1 AND o.order_date <= $2
		GROUP BY c.id, c.name
		ORDER BY total_revenue DESC
	`

	rows, err := r.Query(query, dateRange.Start(), dateRange.End())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.CategoryStats
	for rows.Next() {
		// Ces variables locales servent de tampons temporaires pour Scan.
		var (
			catID        int64
			catName      string
			totalRevenue float64
			totalOrders  int
		)

		if err := rows.Scan(&catID, &catName, &totalRevenue, &totalOrders); err != nil {
			return nil, err
		}

		revenue, _ := shareddomain.NewMoney(totalRevenue, "EUR")
		stat := domain.NewCategoryStats(
			catalogdomain.CategoryID(catID),
			catName,
			revenue,
			totalOrders,
		)
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetTopProducts récupère les N meilleurs produits (optimisé)
// PERFORMANCE: ✓ Version optimisée avec GROUP BY en SQL
//   - Agrégation faite par PostgreSQL (moteur C optimisé)
//   - Seulement les résultats agrégés sont transférés sur le réseau
//   - Si 100k order_items → 1000 products: on transfère 1000 rows au lieu de 100k!
func (r *StatsQueryRepository) GetTopProducts(dateRange shareddomain.DateRange, limit int) ([]*domain.ProductStats, error) {
	// SYNTAXE SQL optimisée:
	//   - COALESCE(value, 0) = retourne 0 si value est NULL (évite NULL en Go)
	//   - SUM() et COUNT() = agrégations faites par le moteur DB (très rapide)
	//   - COUNT(DISTINCT) = déduplique les order IDs (pas besoin de map[int64]bool!)
	//   - GROUP BY = une ligne de résultat par produit (agrégation)
	//   - ORDER BY + LIMIT = tri et pagination côté DB (utilise index si disponible)
	// PERFORMANCE: Query plan optimal si index sur (product_id, order_date)
	query := `
		SELECT p.id, p.name,
		       COALESCE(SUM(oi.subtotal), 0) as total_revenue,
		       COALESCE(COUNT(DISTINCT oi.order_id), 0) as total_orders,
		       COALESCE(SUM(oi.quantity), 0) as total_quantity
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id AND o.order_date >= $1 AND o.order_date <= $2
		GROUP BY p.id, p.name
		ORDER BY total_revenue DESC
		LIMIT $3
	`

	rows, err := r.Query(query, dateRange.Start(), dateRange.End(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// MÉMOIRE: []*domain.ProductStats = slice de POINTEURS
	//   - Chaque pointeur: 8 bytes
	//   - Struct ProductStats allouée sur HEAP séparément
	var stats []*domain.ProductStats
	for rows.Next() {
		// SYNTAXE: var ( ... ) = déclaration de plusieurs variables
		//   - Variables locales sur la STACK (scope de la boucle)
		//   - Réutilisées à chaque itération (pas d'allocation répétée)
		// MÉMOIRE: Total ~56 bytes sur STACK par itération
		var (
			prodID       int64   // 8 bytes
			prodName     string  // 16 bytes (header)
			totalRevenue float64 // 8 bytes
			totalOrders  int     // 8 bytes
			totalQty     int     // 8 bytes
		)

		// PERFORMANCE: Scan très rapide ici, seulement 'limit' rows (ex: 10)
		//   - Vs V1 qui scannait potentiellement 100k rows!
		if err := rows.Scan(&prodID, &prodName, &totalRevenue, &totalOrders, &totalQty); err != nil {
			return nil, err
		}

		revenue, _ := shareddomain.NewMoney(totalRevenue, "EUR")
		qty, _ := shareddomain.NewQuantity(totalQty)

		stat := domain.NewProductStats(
			catalogdomain.ProductID(prodID),
			prodName,
			revenue,
			totalOrders,
			qty,
		)
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetTopStores récupère les N meilleurs magasins (optimisé)
func (r *StatsQueryRepository) GetTopStores(dateRange shareddomain.DateRange, limit int) ([]*domain.StoreStats, error) {
	query := `
		SELECT s.id, s.name,
		       COALESCE(SUM(o.total_amount), 0) as total_revenue,
		       COALESCE(COUNT(o.id), 0) as total_orders
		FROM stores s
		LEFT JOIN orders o ON s.id = o.store_id AND o.order_date >= $1 AND o.order_date <= $2
		GROUP BY s.id, s.name
		ORDER BY total_revenue DESC
		LIMIT $3
	`

	rows, err := r.Query(query, dateRange.Start(), dateRange.End(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.StoreStats
	for rows.Next() {
		var (
			storeID      int64
			storeName    string
			totalRevenue float64
			totalOrders  int
		)

		if err := rows.Scan(&storeID, &storeName, &totalRevenue, &totalOrders); err != nil {
			return nil, err
		}

		revenue, _ := shareddomain.NewMoney(totalRevenue, "EUR")
		stat := domain.NewStoreStats(
			ordersdomain.StoreID(storeID),
			storeName,
			revenue,
			totalOrders,
		)
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetPaymentMethodDistribution récupère la distribution des moyens de paiement (optimisé)
func (r *StatsQueryRepository) GetPaymentMethodDistribution(dateRange shareddomain.DateRange) ([]*domain.PaymentMethodStats, error) {
	query := `
		SELECT pm.id, pm.name,
		       COALESCE(SUM(o.total_amount), 0) as total_revenue,
		       COALESCE(COUNT(o.id), 0) as total_orders
		FROM payment_methods pm
		LEFT JOIN orders o ON pm.id = o.payment_method_id AND o.order_date >= $1 AND o.order_date <= $2
		GROUP BY pm.id, pm.name
		ORDER BY total_revenue DESC
	`

	rows, err := r.Query(query, dateRange.Start(), dateRange.End())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Premier passage: collecter les données et calculer le total
	type pmData struct {
		id           ordersdomain.PaymentMethodID
		name         string
		totalRevenue shareddomain.Money
		totalOrders  int
	}

	var data []pmData
	var grandTotal float64

	for rows.Next() {
		var (
			pmID         int64
			pmName       string
			totalRevenue float64
			totalOrders  int
		)

		if err := rows.Scan(&pmID, &pmName, &totalRevenue, &totalOrders); err != nil {
			return nil, err
		}

		revenue, _ := shareddomain.NewMoney(totalRevenue, "EUR")
		data = append(data, pmData{
			id:           ordersdomain.PaymentMethodID(pmID),
			name:         pmName,
			totalRevenue: revenue,
			totalOrders:  totalOrders,
		})
		grandTotal += totalRevenue
	}

	// Deuxième passage: calculer les pourcentages
	var stats []*domain.PaymentMethodStats
	for _, d := range data {
		percentage := 0.0
		if grandTotal > 0 {
			percentage = (d.totalRevenue.Amount() / grandTotal) * 100
		}

		stat := domain.NewPaymentMethodStats(
			d.id,
			d.name,
			d.totalRevenue,
			d.totalOrders,
			percentage,
		)
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetAllOrderItems récupère tous les items de commande dans une période (pour V1 inefficace)
// PERFORMANCE: ⚠️ Problème majeur - récupère TOUTES les lignes sans agrégation
//   - Transfert réseau: Si 100k rows × 80 bytes = 8 MB de données transférées
//   - Base de données fait un FULL SCAN puis envoie tout au client
//   - Mieux: faire des GROUP BY en SQL pour agréger côté DB
func (r *StatsQueryRepository) GetAllOrderItems(dateRange shareddomain.DateRange) ([]OrderItemData, error) {
	// SYNTAXE SQL: $1, $2 = paramètres positionnels (protection contre SQL injection)
	// PERFORMANCE: INNER JOIN = ok, mais manque de GROUP BY
	//   - ORDER BY est coûteux sur gros volumes (nécessite tri en mémoire ou index)
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal,
		       o.order_date, o.customer_id, o.store_id, o.payment_method_id
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1 AND o.order_date <= $2
		ORDER BY o.order_date DESC
	`
	// SYNTAXE: r.Query() exécute la requête et retourne un itérateur de lignes
	// MÉMOIRE: rows est un curseur (léger), pas toutes les données en RAM immédiatement
	//   - Mais on va tout charger dans []OrderItemData après (là c'est lourd!)
	rows, err := r.Query(query, dateRange.Start(), dateRange.End())
	if err != nil {
		return nil, err
	}

	// SYNTAXE: defer = exécute à la fin de la fonction (comme finally en Java)
	// IMPORTANT: Toujours Close() pour libérer la connexion DB au pool
	//   - Sans defer, si erreur avant, connexion DB leak!
	defer rows.Close()

	// MÉMOIRE: []OrderItemData = slice sans capacité initiale
	//   - Chaque append peut déclencher réallocation (x2 capacité)
	//   - Mieux: items := make([]OrderItemData, 0, 10000) si on connaît la taille
	var items []OrderItemData

	// SYNTAXE: rows.Next() = avance au prochain résultat
	//   - Retourne false quand plus de lignes ou erreur
	// PERFORMANCE: Itération O(n), rien à optimiser ici
	for rows.Next() {
		var item OrderItemData

		// SYNTAXE: rows.Scan(&variable) = lit les colonnes du SELECT dans les variables
		//   - & nécessaire car Scan modifie les variables (passage par référence)
		//   - L'ordre doit correspondre exactement à l'ordre du SELECT
		// MÉMOIRE: Scan copie les données de la réponse SQL vers la struct
		//   - Conversions: PostgreSQL types → Go types (int64, float64, string)
		//   - Strings: allocation HEAP pour chaque string (OrderDate)
		// PERFORMANCE: Scan est optimisé, mais copie obligatoire des données
		if err := rows.Scan(
			&item.ItemID, &item.OrderID, &item.ProductID, &item.Quantity,
			&item.UnitPrice, &item.Subtotal, &item.OrderDate,
			&item.CustomerID, &item.StoreID, &item.PaymentMethodID,
		); err != nil {
			return nil, err
		}

		// SYNTAXE: append(slice, element) = ajoute élément à la fin
		// MÉMOIRE: Si capacité insuffisante:
		//   - Alloue nouveau tableau 2x plus grand
		//   - Copie toutes les anciennes valeurs
		//   - Ancien tableau sera GC plus tard
		// PERFORMANCE: Coût amorti O(1), mais pics de latence possibles lors réallocation
		items = append(items, item)
	}

	return items, nil
}

// OrderItemData structure pour les données brutes d'items
// MÉMOIRE: Calcul de la taille en mémoire:
//   - ItemID: 8 bytes (int64)
//   - OrderID: 8 bytes (int64)
//   - ProductID: 8 bytes (int64)
//   - Quantity: 8 bytes (int sur 64-bit, aligné)
//   - UnitPrice: 8 bytes (float64)
//   - Subtotal: 8 bytes (float64)
//   - OrderDate: 16 bytes (string = pointer 8b + length 8b, data sur HEAP séparément)
//   - CustomerID: 8 bytes (int64)
//   - StoreID: 8 bytes (int64)
//   - PaymentMethodID: 8 bytes (int64)
//   - TOTAL: ~88 bytes par struct (sans compter les données string sur HEAP)
//
// PERFORMANCE: Struct assez grande, préférer passer par pointeur si beaucoup de copies
type OrderItemData struct {
	ItemID          int64
	OrderID         int64
	ProductID       int64
	Quantity        int
	UnitPrice       float64
	Subtotal        float64
	OrderDate       string // String alloue sur HEAP séparément
	CustomerID      int64
	StoreID         int64
	PaymentMethodID int64
}
