package domain

import (
	"errors"
	"fmt"
)

// Money représente une valeur monétaire avec garanties d'invariants
type Money struct {
	amount   float64
	currency string
}

// NewMoney crée une nouvelle instance de Money avec validation
func NewMoney(amount float64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("amount cannot be negative")
	}
	if currency == "" {
		return Money{}, errors.New("currency cannot be empty")
	}
	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

// MustNewMoney crée un Money en paniquant si invalide (pour les constantes)
func MustNewMoney(amount float64, currency string) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(fmt.Sprintf("invalid money: %v", err))
	}
	return m
}

// Amount retourne le montant
func (m Money) Amount() float64 {
	return m.amount
}

// Currency retourne la devise
func (m Money) Currency() string {
	return m.currency
}

// Add additionne deux Money (même devise requise)
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.currency, other.currency)
	}
	return Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

// Multiply multiplie le montant par un facteur
func (m Money) Multiply(factor float64) (Money, error) {
	if factor < 0 {
		return Money{}, errors.New("multiplication factor cannot be negative")
	}
	return Money{
		amount:   m.amount * factor,
		currency: m.currency,
	}, nil
}

// IsZero vérifie si le montant est zéro
func (m Money) IsZero() bool {
	return m.amount == 0
}

// String retourne une représentation textuelle
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.amount, m.currency)
}

// Equals vérifie l'égalité entre deux Money
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}
