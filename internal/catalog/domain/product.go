package domain

import (
	"errors"
	"time"

	"eval/internal/shared/domain"
)

// ProductID représente l'identifiant unique d'un produit
type ProductID int64

// Product représente un produit du catalogue
type Product struct {
	id            ProductID
	name          string
	supplierID    SupplierID
	basePrice     domain.Money
	stockQuantity domain.Quantity
	categories    []CategoryID
	createdAt     time.Time
}

// NewProduct crée une nouvelle instance de Product avec validation
func NewProduct(
	id ProductID,
	name string,
	supplierID SupplierID,
	basePrice domain.Money,
	stockQuantity domain.Quantity,
	categories []CategoryID,
	createdAt time.Time,
) (*Product, error) {
	if name == "" {
		return nil, errors.New("product name cannot be empty")
	}
	if supplierID <= 0 {
		return nil, errors.New("invalid supplier ID")
	}
	if basePrice.IsZero() {
		return nil, errors.New("base price cannot be zero")
	}

	return &Product{
		id:            id,
		name:          name,
		supplierID:    supplierID,
		basePrice:     basePrice,
		stockQuantity: stockQuantity,
		categories:    categories,
		createdAt:     createdAt,
	}, nil
}

// ID retourne l'identifiant du produit
func (p *Product) ID() ProductID {
	return p.id
}

// Name retourne le nom du produit
func (p *Product) Name() string {
	return p.name
}

// SupplierID retourne l'identifiant du fournisseur
func (p *Product) SupplierID() SupplierID {
	return p.supplierID
}

// BasePrice retourne le prix de base
func (p *Product) BasePrice() domain.Money {
	return p.basePrice
}

// StockQuantity retourne la quantité en stock
func (p *Product) StockQuantity() domain.Quantity {
	return p.stockQuantity
}

// Categories retourne les catégories du produit
func (p *Product) Categories() []CategoryID {
	return append([]CategoryID{}, p.categories...)
}

// CreatedAt retourne la date de création
func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

// HasCategory vérifie si le produit appartient à une catégorie
func (p *Product) HasCategory(categoryID CategoryID) bool {
	for _, cid := range p.categories {
		if cid == categoryID {
			return true
		}
	}
	return false
}

// IsInStock vérifie si le produit est en stock
func (p *Product) IsInStock() bool {
	return !p.stockQuantity.IsZero()
}

// CalculatePriceWithVariation calcule le prix avec une variation (pour les ventes)
func (p *Product) CalculatePriceWithVariation(variationPercent float64) (domain.Money, error) {
	factor := 1 + (variationPercent / 100)
	return p.basePrice.Multiply(factor)
}

// UpdateStock met à jour le stock (si nécessaire pour command repo)
func (p *Product) UpdateStock(newQuantity domain.Quantity) {
	p.stockQuantity = newQuantity
}
