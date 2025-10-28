package domain

import (
	"errors"
	"time"

	catalogdomain "eval/internal/catalog/domain"
	"eval/internal/shared/domain"
)

// OrderItemID représente l'identifiant unique d'un item de commande
type OrderItemID int64

// OrderItem représente un item dans une commande (entity dans l'aggregate Order)
type OrderItem struct {
	id        OrderItemID
	orderID   OrderID
	productID catalogdomain.ProductID
	quantity  domain.Quantity
	unitPrice domain.Money
	subtotal  domain.Money
	createdAt time.Time
}

// NewOrderItem crée un nouvel item de commande avec validation
func NewOrderItem(
	id OrderItemID,
	orderID OrderID,
	productID catalogdomain.ProductID,
	quantity domain.Quantity,
	unitPrice domain.Money,
	createdAt time.Time,
) (*OrderItem, error) {
	if orderID <= 0 {
		return nil, errors.New("invalid order ID")
	}
	if productID <= 0 {
		return nil, errors.New("invalid product ID")
	}
	if quantity.IsZero() {
		return nil, errors.New("quantity cannot be zero")
	}
	if unitPrice.IsZero() {
		return nil, errors.New("unit price cannot be zero")
	}

	// Calculer le subtotal
	subtotal, err := unitPrice.Multiply(float64(quantity.Value()))
	if err != nil {
		return nil, err
	}

	return &OrderItem{
		id:        id,
		orderID:   orderID,
		productID: productID,
		quantity:  quantity,
		unitPrice: unitPrice,
		subtotal:  subtotal,
		createdAt: createdAt,
	}, nil
}

// ID retourne l'identifiant de l'item
func (oi *OrderItem) ID() OrderItemID {
	return oi.id
}

// OrderID retourne l'identifiant de la commande
func (oi *OrderItem) OrderID() OrderID {
	return oi.orderID
}

// ProductID retourne l'identifiant du produit
func (oi *OrderItem) ProductID() catalogdomain.ProductID {
	return oi.productID
}

// Quantity retourne la quantité
func (oi *OrderItem) Quantity() domain.Quantity {
	return oi.quantity
}

// UnitPrice retourne le prix unitaire
func (oi *OrderItem) UnitPrice() domain.Money {
	return oi.unitPrice
}

// Subtotal retourne le sous-total (quantité × prix unitaire)
func (oi *OrderItem) Subtotal() domain.Money {
	return oi.subtotal
}

// CreatedAt retourne la date de création
func (oi *OrderItem) CreatedAt() time.Time {
	return oi.createdAt
}

// UpdateQuantity met à jour la quantité et recalcule le subtotal
func (oi *OrderItem) UpdateQuantity(newQuantity domain.Quantity) error {
	if newQuantity.IsZero() {
		return errors.New("quantity cannot be zero")
	}

	oi.quantity = newQuantity

	// Recalculer le subtotal
	subtotal, err := oi.unitPrice.Multiply(float64(newQuantity.Value()))
	if err != nil {
		return err
	}

	oi.subtotal = subtotal
	return nil
}
