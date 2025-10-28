package infrastructure

import (
	"database/sql"

	catalogdomain "eval/internal/catalog/domain"
	"eval/internal/analytics/domain"
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
func (r *StatsQueryRepository) GetTopProducts(dateRange shareddomain.DateRange, limit int) ([]*domain.ProductStats, error) {
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

	var stats []*domain.ProductStats
	for rows.Next() {
		var (
			prodID       int64
			prodName     string
			totalRevenue float64
			totalOrders  int
			totalQty     int
		)

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
func (r *StatsQueryRepository) GetAllOrderItems(dateRange shareddomain.DateRange) ([]OrderItemData, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal,
		       o.order_date, o.customer_id, o.store_id, o.payment_method_id
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.order_date >= $1 AND o.order_date <= $2
		ORDER BY o.order_date DESC
	`

	rows, err := r.Query(query, dateRange.Start(), dateRange.End())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItemData
	for rows.Next() {
		var item OrderItemData
		if err := rows.Scan(
			&item.ItemID, &item.OrderID, &item.ProductID, &item.Quantity,
			&item.UnitPrice, &item.Subtotal, &item.OrderDate,
			&item.CustomerID, &item.StoreID, &item.PaymentMethodID,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// OrderItemData structure pour les données brutes d'items
type OrderItemData struct {
	ItemID          int64
	OrderID         int64
	ProductID       int64
	Quantity        int
	UnitPrice       float64
	Subtotal        float64
	OrderDate       string
	CustomerID      int64
	StoreID         int64
	PaymentMethodID int64
}
