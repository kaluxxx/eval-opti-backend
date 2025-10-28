package infrastructure

import (
	"database/sql"
	"time"

	catalogdomain "eval/internal/catalog/domain"
	"eval/internal/orders/domain"
	shareddomain "eval/internal/shared/domain"
	"eval/internal/shared/infrastructure"
)

// OrderQueryRepository repository pour les requêtes de lecture sur les commandes
type OrderQueryRepository struct {
	infrastructure.BaseRepository
}

// NewOrderQueryRepository crée un nouveau repository de lecture pour les commandes
func NewOrderQueryRepository(db *sql.DB) *OrderQueryRepository {
	return &OrderQueryRepository{
		BaseRepository: infrastructure.NewBaseRepository(db),
	}
}

// FindByDateRange trouve les commandes dans une période donnée
func (r *OrderQueryRepository) FindByDateRange(dateRange shareddomain.DateRange) ([]*domain.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.store_id, o.payment_method_id, o.promotion_id,
		       o.order_date, o.total_amount, o.status, o.created_at
		FROM orders o
		WHERE o.order_date >= $1 AND o.order_date <= $2
		ORDER BY o.order_date DESC
	`

	rows, err := r.Query(query, dateRange.Start(), dateRange.End())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order, err := r.scanOrder(rows)
		if err != nil {
			return nil, err
		}

		// Charger les items
		items, err := r.findItemsByOrderID(order.ID())
		if err != nil {
			return nil, err
		}
		if err := order.SetItems(items); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// FindByID trouve une commande par son ID
func (r *OrderQueryRepository) FindByID(id domain.OrderID) (*domain.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.store_id, o.payment_method_id, o.promotion_id,
		       o.order_date, o.total_amount, o.status, o.created_at
		FROM orders o
		WHERE o.id = $1
	`

	row := r.QueryRow(query, int64(id))
	order, err := r.scanOrderRow(row)
	if err != nil {
		return nil, err
	}

	// Charger les items
	items, err := r.findItemsByOrderID(order.ID())
	if err != nil {
		return nil, err
	}
	if err := order.SetItems(items); err != nil {
		return nil, err
	}

	return order, nil
}

// findItemsByOrderID récupère les items d'une commande
func (r *OrderQueryRepository) findItemsByOrderID(orderID domain.OrderID) ([]*domain.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.created_at
		FROM order_items oi
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`

	rows, err := r.Query(query, int64(orderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.OrderItem
	for rows.Next() {
		var (
			itemID    int64
			ordID     int64
			productID int64
			quantity  int
			unitPrice float64
			createdAt time.Time
		)

		if err := rows.Scan(&itemID, &ordID, &productID, &quantity, &unitPrice, &createdAt); err != nil {
			return nil, err
		}

		qty, _ := shareddomain.NewQuantity(quantity)
		price, _ := shareddomain.NewMoney(unitPrice, "EUR")

		item, err := domain.NewOrderItem(
			domain.OrderItemID(itemID),
			domain.OrderID(ordID),
			catalogdomain.ProductID(productID),
			qty,
			price,
			createdAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

// scanOrder scanne une ligne de résultat en Order
func (r *OrderQueryRepository) scanOrder(rows *sql.Rows) (*domain.Order, error) {
	var (
		id              int64
		customerID      int64
		storeID         int64
		paymentMethodID int64
		promotionID     sql.NullInt64
		orderDate       time.Time
		totalAmount     float64
		status          string
		createdAt       time.Time
	)

	if err := rows.Scan(&id, &customerID, &storeID, &paymentMethodID, &promotionID,
		&orderDate, &totalAmount, &status, &createdAt); err != nil {
		return nil, err
	}

	var promID *domain.PromotionID
	if promotionID.Valid {
		pid := domain.PromotionID(promotionID.Int64)
		promID = &pid
	}

	return domain.NewOrder(
		domain.OrderID(id),
		domain.CustomerID(customerID),
		domain.StoreID(storeID),
		domain.PaymentMethodID(paymentMethodID),
		promID,
		orderDate,
		domain.OrderStatus(status),
		createdAt,
	)
}

// scanOrderRow scanne une seule ligne en Order
func (r *OrderQueryRepository) scanOrderRow(row *sql.Row) (*domain.Order, error) {
	var (
		id              int64
		customerID      int64
		storeID         int64
		paymentMethodID int64
		promotionID     sql.NullInt64
		orderDate       time.Time
		totalAmount     float64
		status          string
		createdAt       time.Time
	)

	if err := row.Scan(&id, &customerID, &storeID, &paymentMethodID, &promotionID,
		&orderDate, &totalAmount, &status, &createdAt); err != nil {
		return nil, err
	}

	var promID *domain.PromotionID
	if promotionID.Valid {
		pid := domain.PromotionID(promotionID.Int64)
		promID = &pid
	}

	return domain.NewOrder(
		domain.OrderID(id),
		domain.CustomerID(customerID),
		domain.StoreID(storeID),
		domain.PaymentMethodID(paymentMethodID),
		promID,
		orderDate,
		domain.OrderStatus(status),
		createdAt,
	)
}
