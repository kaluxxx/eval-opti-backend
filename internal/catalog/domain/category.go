package domain

import (
	"errors"
	"time"
)

// CategoryID représente l'identifiant unique d'une catégorie
type CategoryID int64

// Category représente une catégorie de produits
type Category struct {
	id          CategoryID
	name        string
	description string
	createdAt   time.Time
}

// NewCategory crée une nouvelle instance de Category avec validation
func NewCategory(
	id CategoryID,
	name string,
	description string,
	createdAt time.Time,
) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}

	return &Category{
		id:          id,
		name:        name,
		description: description,
		createdAt:   createdAt,
	}, nil
}
