package infrastructure

import (
	"database/sql"
	"time"

	"eval/internal/export/domain"
	shareddomain "eval/internal/shared/domain"
	"eval/internal/shared/infrastructure"
)

// ExportQueryRepository repository pour les requêtes d'export
type ExportQueryRepository struct {
	infrastructure.BaseRepository
}

// NewExportQueryRepository crée un nouveau repository d'export
func NewExportQueryRepository(db *sql.DB) *ExportQueryRepository {
	return &ExportQueryRepository{
		BaseRepository: infrastructure.NewBaseRepository(db),
	}
}

// GetSalesDataOptimized récupère les données de vente de manière optimisée (une seule requête)
// PERFORMANCE: ✓ OPTIMISÉ - UNE SEULE requête avec tous les JOINs
//   - Vs V1 qui fait 1 query initiale + 6 queries par order_item (N+1 × 6!)
//   - Ex: 10k order_items → V1 = 60,001 queries vs V2 = 1 query
//   - Temps: V1 ≈ 60s (1ms/query) vs V2 ≈ 100ms
func (r *ExportQueryRepository) GetSalesDataOptimized(dateRange shareddomain.DateRange) ([]*domain.SaleExportRow, error) {
	// SYNTAXE SQL optimisée avec JOINS:
	//   - INNER JOIN = seulement les lignes avec correspondance (orders, order_items, etc.)
	//   - LEFT JOIN = garde la ligne même si pas de correspondance (promotions optionnelles)
	//   - COALESCE(value, 'default') = retourne 'default' si value est NULL
	// PERFORMANCE: PostgreSQL fait tous les JOINs en UNE PASSE
	//   - Query planner optimise l'ordre des joins
	//   - Utilise les index pour accélérer les joins
	//   - Dénormalise les données côté DB (plus efficace qu'en Go)
	// MÉMOIRE: Transfère toutes les colonnes nécessaires d'un coup
	//   - Évite les round-trips réseau (latence majeure en DB)
	query := `
		SELECT
			o.id as order_id,
			o.customer_id,
			o.store_id,
			s.name as store_name,
			oi.product_id,
			p.name as product_name,
			COALESCE(c.name, 'Uncategorized') as category_name,
			oi.quantity,
			oi.unit_price,
			oi.subtotal,
			pm.name as payment_method,
			COALESCE(pr.code, '') as promotion_code,
			o.order_date
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		INNER JOIN stores s ON o.store_id = s.id
		INNER JOIN payment_methods pm ON o.payment_method_id = pm.id
		LEFT JOIN promotions pr ON o.promotion_id = pr.id
		LEFT JOIN product_categories pc ON p.id = pc.product_id
		LEFT JOIN categories c ON pc.category_id = c.id
		WHERE o.order_date >= $1 AND o.order_date <= $2
		ORDER BY o.order_date DESC, o.id, oi.id
	`

	rows, err := r.Query(query, dateRange.Start(), dateRange.End())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var salesData []*domain.SaleExportRow
	//
	for rows.Next() {
		var (
			orderID       int64
			customerID    int64
			storeID       int64
			storeName     string
			productID     int64
			productName   string
			categoryName  string
			quantity      int
			unitPrice     float64
			subtotal      float64
			paymentMethod string
			promotionCode string
			orderDate     time.Time
		)

		if err := rows.Scan(
			&orderID, &customerID, &storeID, &storeName,
			&productID, &productName, &categoryName,
			&quantity, &unitPrice, &subtotal,
			&paymentMethod, &promotionCode, &orderDate,
		); err != nil {
			return nil, err
		}

		row := domain.NewSaleExportRow(
			orderID, customerID, storeID, productID,
			storeName, productName, categoryName,
			quantity, unitPrice, subtotal,
			paymentMethod, promotionCode, orderDate,
		)
		salesData = append(salesData, row)
	}

	return salesData, nil
}

// GetSalesDataInefficient récupère les données avec N+1 queries (version inefficace)
// PERFORMANCE: ⚠️ CATASTROPHIQUE - N+1 QUERIES PROBLEM × 6!
//   - 1 query pour order_items
//   - Puis POUR CHAQUE order_item (loop):
//   - 1 query pour order
//   - 1 query pour store
//   - 1 query pour product
//   - 1 query pour category
//   - 1 query pour payment_method
//   - 1 query pour promotion
//   - Total: 1 + (N × 6) queries où N = nombre d'order_items
//   - Ex: 10,000 items = 60,001 queries! Temps: ~60 secondes minimum
func (r *ExportQueryRepository) GetSalesDataInefficient(dateRange shareddomain.DateRange) ([]*domain.SaleExportRow, error) {
	// Première query: récupère tous les order items
	// PERFORMANCE: Cette query est ok, mais c'est ce qui suit qui est terrible
	query1 := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1 AND o.order_date <= $2
		ORDER BY o.order_date DESC
	`

	rows, err := r.Query(query1, dateRange.Start(), dateRange.End())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// MÉMOIRE: Struct temporaire pour stocker les données partielles
	type itemData struct {
		itemID    int64
		orderID   int64
		productID int64
		quantity  int
		unitPrice float64
		subtotal  float64
	}

	var items []itemData
	for rows.Next() {
		var item itemData
		if err := rows.Scan(&item.itemID, &item.orderID, &item.productID, &item.quantity, &item.unitPrice, &item.subtotal); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	// ⚠️ LE DÉSASTRE COMMENCE ICI: Boucle avec 6 queries par itération!
	// PERFORMANCE: Chaque QueryRow = round-trip réseau complet
	//   - Latence réseau: ~1ms par query (optimiste)
	//   - Si 10k items: 10k × 6 × 1ms = 60 secondes MINIMUM
	//   - Plus le temps d'exécution SQL + parsing + query planning
	var salesData []*domain.SaleExportRow
	for _, item := range items {
		// ⚠️ QUERY 1: Order details
		// SYNTAXE: sql.NullInt64 = type Go pour gérer les NULL SQL
		//   - .Valid = true si pas NULL
		//   - .Int64 = valeur si Valid
		var customerID, storeID, paymentMethodID int64
		var promotionID sql.NullInt64
		var orderDate time.Time
		orderQuery := `SELECT customer_id, store_id, payment_method_id, promotion_id, order_date FROM orders WHERE id = $1`
		err := r.QueryRow(orderQuery, item.orderID).Scan(&customerID, &storeID, &paymentMethodID, &promotionID, &orderDate)
		if err != nil {
			continue
		}

		// ⚠️ QUERY 2: Store name
		var storeName string
		storeQuery := `SELECT name FROM stores WHERE id = $1`
		_ = r.QueryRow(storeQuery, storeID).Scan(&storeName)

		// ⚠️ QUERY 3: Product name
		var productName string
		productQuery := `SELECT name FROM products WHERE id = $1`
		_ = r.QueryRow(productQuery, item.productID).Scan(&productName)

		// ⚠️ QUERY 4: Category (with JOIN!)
		var categoryName string
		categoryQuery := `
			SELECT c.name FROM categories c
			INNER JOIN product_categories pc ON c.id = pc.category_id
			WHERE pc.product_id = $1 LIMIT 1
		`
		_ = r.QueryRow(categoryQuery, item.productID).Scan(&categoryName)

		// ⚠️ QUERY 5: Payment method
		var paymentMethod string
		pmQuery := `SELECT name FROM payment_methods WHERE id = $1`
		_ = r.QueryRow(pmQuery, paymentMethodID).Scan(&paymentMethod)

		// ⚠️ QUERY 6: Promotion (conditional)
		promotionCode := ""
		if promotionID.Valid {
			prQuery := `SELECT code FROM promotions WHERE id = $1`
			_ = r.QueryRow(prQuery, promotionID.Int64).Scan(&promotionCode)
		}

		row := domain.NewSaleExportRow(
			item.orderID, customerID, storeID, item.productID,
			storeName, productName, categoryName,
			item.quantity, item.unitPrice, item.subtotal,
			paymentMethod, promotionCode, orderDate,
		)
		salesData = append(salesData, row)
	}

	return salesData, nil
}
