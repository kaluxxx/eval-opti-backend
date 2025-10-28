package domain

import (
	"errors"
	"regexp"
	"time"
)

// SupplierID représente l'identifiant unique d'un fournisseur
type SupplierID int64

// Email représente une adresse email validée
type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewEmail crée une nouvelle instance d'Email avec validation
func NewEmail(value string) (Email, error) {
	if !emailRegex.MatchString(value) {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: value}, nil
}

// Value retourne la valeur de l'email
func (e Email) Value() string {
	return e.value
}

// String retourne la représentation textuelle
func (e Email) String() string {
	return e.value
}

// Supplier représente un fournisseur de produits
type Supplier struct {
	id          SupplierID
	name        string
	contactName string
	email       Email
	phone       string
	address     string
	city        string
	country     string
	createdAt   time.Time
}

// NewSupplier crée une nouvelle instance de Supplier avec validation
func NewSupplier(
	id SupplierID,
	name string,
	contactName string,
	email Email,
	phone string,
	address string,
	city string,
	country string,
	createdAt time.Time,
) (*Supplier, error) {
	if name == "" {
		return nil, errors.New("supplier name cannot be empty")
	}
	if country == "" {
		return nil, errors.New("country cannot be empty")
	}

	return &Supplier{
		id:          id,
		name:        name,
		contactName: contactName,
		email:       email,
		phone:       phone,
		address:     address,
		city:        city,
		country:     country,
		createdAt:   createdAt,
	}, nil
}

// ID retourne l'identifiant du fournisseur
func (s *Supplier) ID() SupplierID {
	return s.id
}

// Name retourne le nom du fournisseur
func (s *Supplier) Name() string {
	return s.name
}

// ContactName retourne le nom du contact
func (s *Supplier) ContactName() string {
	return s.contactName
}

// Email retourne l'email
func (s *Supplier) Email() Email {
	return s.email
}

// Phone retourne le téléphone
func (s *Supplier) Phone() string {
	return s.phone
}

// Address retourne l'adresse
func (s *Supplier) Address() string {
	return s.address
}

// City retourne la ville
func (s *Supplier) City() string {
	return s.city
}

// Country retourne le pays
func (s *Supplier) Country() string {
	return s.country
}

// CreatedAt retourne la date de création
func (s *Supplier) CreatedAt() time.Time {
	return s.createdAt
}
