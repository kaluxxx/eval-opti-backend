package domain

import (
	"errors"
	"time"

	catalogdomain "eval/internal/catalog/domain"
	"eval/internal/shared/domain"
)

// OrderID représente l'identifiant unique d'une commande
type OrderID int64

// CustomerID représente l'identifiant d'un client
type CustomerID int64

// StoreID représente l'identifiant d'un magasin
type StoreID int64

// PaymentMethodID représente l'identifiant d'un moyen de paiement
type PaymentMethodID int64

// PromotionID représente l'identifiant d'une promotion
type PromotionID int64

// OrderStatus représente le statut d'une commande
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order représente une commande (aggregate root)
type Order struct {
	id              OrderID
	customerID      CustomerID
	storeID         StoreID
	paymentMethodID PaymentMethodID
	promotionID     *PromotionID
	orderDate       time.Time
	totalAmount     domain.Money
	status          OrderStatus
	items           []*OrderItem
	createdAt       time.Time
}

// NewOrder crée une nouvelle commande avec validation
func NewOrder(
	id OrderID,
	customerID CustomerID,
	storeID StoreID,
	paymentMethodID PaymentMethodID,
	promotionID *PromotionID,
	orderDate time.Time,
	status OrderStatus,
	createdAt time.Time,
) (*Order, error) {
	if customerID <= 0 {
		return nil, errors.New("invalid customer ID")
	}
	if storeID <= 0 {
		return nil, errors.New("invalid store ID")
	}
	if paymentMethodID <= 0 {
		return nil, errors.New("invalid payment method ID")
	}

	totalAmount, _ := domain.NewMoney(0, "EUR")

	return &Order{
		id:              id,
		customerID:      customerID,
		storeID:         storeID,
		paymentMethodID: paymentMethodID,
		promotionID:     promotionID,
		orderDate:       orderDate,
		totalAmount:     totalAmount,
		status:          status,
		items:           make([]*OrderItem, 0),
		createdAt:       createdAt,
	}, nil
}

// ID retourne l'identifiant de la commande
func (o *Order) ID() OrderID {
	return o.id
}

// CustomerID retourne l'identifiant du client
func (o *Order) CustomerID() CustomerID {
	return o.customerID
}

// StoreID retourne l'identifiant du magasin
func (o *Order) StoreID() StoreID {
	return o.storeID
}

// PaymentMethodID retourne l'identifiant du moyen de paiement
func (o *Order) PaymentMethodID() PaymentMethodID {
	return o.paymentMethodID
}

// PromotionID retourne l'identifiant de la promotion (peut être nil)
func (o *Order) PromotionID() *PromotionID {
	return o.promotionID
}

// OrderDate retourne la date de commande
func (o *Order) OrderDate() time.Time {
	return o.orderDate
}

// TotalAmount retourne le montant total
func (o *Order) TotalAmount() domain.Money {
	return o.totalAmount
}

// Status retourne le statut de la commande
func (o *Order) Status() OrderStatus {
	return o.status
}

// Items retourne les items de la commande
func (o *Order) Items() []*OrderItem {
	return append([]*OrderItem{}, o.items...)
}

// CreatedAt retourne la date de création
func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

// AddItem ajoute un item à la commande (invariant: recalcule le total)
func (o *Order) AddItem(item *OrderItem) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	// Vérifier que l'item n'existe pas déjà
	for _, existingItem := range o.items {
		if existingItem.ProductID() == item.ProductID() {
			return errors.New("item already exists in order")
		}
	}

	o.items = append(o.items, item)

	// Recalculer le total
	return o.recalculateTotal()
}

// RemoveItem supprime un item de la commande
func (o *Order) RemoveItem(productID catalogdomain.ProductID) error {
	found := false
	newItems := make([]*OrderItem, 0, len(o.items))

	for _, item := range o.items {
		if item.ProductID() != productID {
			newItems = append(newItems, item)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New("item not found in order")
	}

	o.items = newItems
	return o.recalculateTotal()
}

// recalculateTotal recalcule le montant total de la commande
func (o *Order) recalculateTotal() error {
	total, _ := domain.NewMoney(0, "EUR")

	for _, item := range o.items {
		newTotal, err := total.Add(item.Subtotal())
		if err != nil {
			return err
		}
		total = newTotal
	}

	o.totalAmount = total
	return nil
}

// Complete marque la commande comme complétée
func (o *Order) Complete() error {
	if o.status == OrderStatusCompleted {
		return errors.New("order is already completed")
	}
	if o.status == OrderStatusCancelled {
		return errors.New("cannot complete a cancelled order")
	}
	if len(o.items) == 0 {
		return errors.New("cannot complete an order without items")
	}

	o.status = OrderStatusCompleted
	return nil
}

// Cancel annule la commande
func (o *Order) Cancel() error {
	if o.status == OrderStatusCancelled {
		return errors.New("order is already cancelled")
	}
	if o.status == OrderStatusCompleted {
		return errors.New("cannot cancel a completed order")
	}

	o.status = OrderStatusCancelled
	return nil
}

// HasPromotion vérifie si la commande a une promotion
func (o *Order) HasPromotion() bool {
	return o.promotionID != nil
}

// ItemCount retourne le nombre d'items dans la commande
func (o *Order) ItemCount() int {
	return len(o.items)
}

// SetItems définit les items de la commande (pour hydratation depuis DB)
func (o *Order) SetItems(items []*OrderItem) error {
	o.items = items
	return o.recalculateTotal()
}
