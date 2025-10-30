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

// IsZero vérifie si la quantité est nulle
func (q Quantity) IsZero() bool {
	return q.value == 0
}
