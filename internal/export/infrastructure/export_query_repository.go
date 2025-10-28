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
func (r *ExportQueryRepository) GetSalesDataOptimized(dateRange shareddomain.DateRange) ([]*domain.SaleExportRow, error) {
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
func (r *ExportQueryRepository) GetSalesDataInefficient(dateRange shareddomain.DateRange) ([]*domain.SaleExportRow, error) {
	// D'abord récupérer tous les order items
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

	type itemData struct {
		itemID     int64
		orderID    int64
		productID  int64
		quantity   int
		unitPrice  float64
		subtotal   float64
	}

	var items []itemData
	for rows.Next() {
		var item itemData
		if err := rows.Scan(&item.itemID, &item.orderID, &item.productID, &item.quantity, &item.unitPrice, &item.subtotal); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	// Maintenant, pour chaque item, faire des requêtes séparées (N+1!)
	var salesData []*domain.SaleExportRow
	for _, item := range items {
		// Query pour l'order
		var customerID, storeID, paymentMethodID int64
		var promotionID sql.NullInt64
		var orderDate time.Time
		orderQuery := `SELECT customer_id, store_id, payment_method_id, promotion_id, order_date FROM orders WHERE id = $1`
		err := r.QueryRow(orderQuery, item.orderID).Scan(&customerID, &storeID, &paymentMethodID, &promotionID, &orderDate)
		if err != nil {
			continue
		}

		// Query pour le store
		var storeName string
		storeQuery := `SELECT name FROM stores WHERE id = $1`
		_ = r.QueryRow(storeQuery, storeID).Scan(&storeName)

		// Query pour le product
		var productName string
		productQuery := `SELECT name FROM products WHERE id = $1`
		_ = r.QueryRow(productQuery, item.productID).Scan(&productName)

		// Query pour la catégorie (première trouvée)
		var categoryName string
		categoryQuery := `
			SELECT c.name FROM categories c
			INNER JOIN product_categories pc ON c.id = pc.category_id
			WHERE pc.product_id = $1 LIMIT 1
		`
		_ = r.QueryRow(categoryQuery, item.productID).Scan(&categoryName)

		// Query pour payment method
		var paymentMethod string
		pmQuery := `SELECT name FROM payment_methods WHERE id = $1`
		_ = r.QueryRow(pmQuery, paymentMethodID).Scan(&paymentMethod)

		// Query pour promotion
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
