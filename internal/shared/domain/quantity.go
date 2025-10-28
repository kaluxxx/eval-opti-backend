package domain

import (
	"errors"
	"fmt"
)

// Quantity représente une quantité avec validation
type Quantity struct {
	value int
}

// NewQuantity crée une nouvelle instance de Quantity avec validation
func NewQuantity(value int) (Quantity, error) {
	if value < 0 {
		return Quantity{}, errors.New("quantity cannot be negative")
	}
	return Quantity{value: value}, nil
}

// MustNewQuantity crée une Quantity en paniquant si invalide
func MustNewQuantity(value int) Quantity {
	q, err := NewQuantity(value)
	if err != nil {
		panic(fmt.Sprintf("invalid quantity: %v", err))
	}
	return q
}

// Value retourne la valeur
func (q Quantity) Value() int {
	return q.value
}

// Add additionne deux quantités
func (q Quantity) Add(other Quantity) Quantity {
	return Quantity{value: q.value + other.value}
}

// Subtract soustrait une quantité (ne peut pas être négatif)
func (q Quantity) Subtract(other Quantity) (Quantity, error) {
	if q.value < other.value {
		return Quantity{}, errors.New("resulting quantity would be negative")
	}
	return Quantity{value: q.value - other.value}, nil
}

// Multiply multiplie la quantité
func (q Quantity) Multiply(factor int) (Quantity, error) {
	if factor < 0 {
		return Quantity{}, errors.New("multiplication factor cannot be negative")
	}
	return Quantity{value: q.value * factor}, nil
}

// IsZero vérifie si la quantité est nulle
func (q Quantity) IsZero() bool {
	return q.value == 0
}

// IsGreaterThan compare deux quantités
func (q Quantity) IsGreaterThan(other Quantity) bool {
	return q.value > other.value
}

// IsLessThan compare deux quantités
func (q Quantity) IsLessThan(other Quantity) bool {
	return q.value < other.value
}

// Equals vérifie l'égalité
func (q Quantity) Equals(other Quantity) bool {
	return q.value == other.value
}

// String retourne une représentation textuelle
func (q Quantity) String() string {
	return fmt.Sprintf("%d", q.value)
}
