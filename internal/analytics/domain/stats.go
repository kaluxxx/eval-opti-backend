package domain

import (
	catalogdomain "eval/internal/catalog/domain"
	ordersdomain "eval/internal/orders/domain"
	"eval/internal/shared/domain"
)

// Stats représente les statistiques globales
type Stats struct {
	totalRevenue      domain.Money
	totalOrders       int
	averageOrderValue domain.Money
	categoryStats     []*CategoryStats
	topProducts       []*ProductStats
	topStores         []*StoreStats
	paymentDistrib    []*PaymentMethodStats
}

// NewStats crée une nouvelle instance de Stats
func NewStats() *Stats {
	revenue, _ := domain.NewMoney(0, "EUR")
	avgOrder, _ := domain.NewMoney(0, "EUR")

	return &Stats{
		totalRevenue:      revenue,
		totalOrders:       0,
		averageOrderValue: avgOrder,
		categoryStats:     make([]*CategoryStats, 0),
		topProducts:       make([]*ProductStats, 0),
		topStores:         make([]*StoreStats, 0),
		paymentDistrib:    make([]*PaymentMethodStats, 0),
	}
}

// TotalRevenue retourne le chiffre d'affaires total
func (s *Stats) TotalRevenue() domain.Money {
	return s.totalRevenue
}

// TotalOrders retourne le nombre total de commandes
func (s *Stats) TotalOrders() int {
	return s.totalOrders
}

// AverageOrderValue retourne la valeur moyenne d'une commande
func (s *Stats) AverageOrderValue() domain.Money {
	return s.averageOrderValue
}

// CategoryStats retourne les statistiques par catégorie
func (s *Stats) CategoryStats() []*CategoryStats {
	return append([]*CategoryStats{}, s.categoryStats...)
}

// TopProducts retourne les meilleurs produits
func (s *Stats) TopProducts() []*ProductStats {
	return append([]*ProductStats{}, s.topProducts...)
}

// TopStores retourne les meilleurs magasins
func (s *Stats) TopStores() []*StoreStats {
	return append([]*StoreStats{}, s.topStores...)
}

// PaymentDistribution retourne la distribution des moyens de paiement
func (s *Stats) PaymentDistribution() []*PaymentMethodStats {
	return append([]*PaymentMethodStats{}, s.paymentDistrib...)
}

// SetTotalRevenue définit le chiffre d'affaires total
func (s *Stats) SetTotalRevenue(revenue domain.Money) {
	s.totalRevenue = revenue
}

// SetTotalOrders définit le nombre total de commandes
func (s *Stats) SetTotalOrders(count int) {
	s.totalOrders = count
}

// SetAverageOrderValue définit la valeur moyenne
func (s *Stats) SetAverageOrderValue(avg domain.Money) {
	s.averageOrderValue = avg
}

// SetCategoryStats définit les statistiques par catégorie
func (s *Stats) SetCategoryStats(stats []*CategoryStats) {
	s.categoryStats = stats
}

// SetTopProducts définit les meilleurs produits
func (s *Stats) SetTopProducts(products []*ProductStats) {
	s.topProducts = products
}

// SetTopStores définit les meilleurs magasins
func (s *Stats) SetTopStores(stores []*StoreStats) {
	s.topStores = stores
}

// SetPaymentDistribution définit la distribution des paiements
func (s *Stats) SetPaymentDistribution(distrib []*PaymentMethodStats) {
	s.paymentDistrib = distrib
}

// CategoryStats représente les statistiques pour une catégorie
type CategoryStats struct {
	categoryID   catalogdomain.CategoryID
	categoryName string
	totalRevenue domain.Money
	totalOrders  int
}

// NewCategoryStats crée une nouvelle instance de CategoryStats
func NewCategoryStats(
	categoryID catalogdomain.CategoryID,
	categoryName string,
	totalRevenue domain.Money,
	totalOrders int,
) *CategoryStats {
	return &CategoryStats{
		categoryID:   categoryID,
		categoryName: categoryName,
		totalRevenue: totalRevenue,
		totalOrders:  totalOrders,
	}
}

// CategoryID retourne l'ID de la catégorie
func (cs *CategoryStats) CategoryID() catalogdomain.CategoryID {
	return cs.categoryID
}

// CategoryName retourne le nom de la catégorie
func (cs *CategoryStats) CategoryName() string {
	return cs.categoryName
}

// TotalRevenue retourne le CA de la catégorie
func (cs *CategoryStats) TotalRevenue() domain.Money {
	return cs.totalRevenue
}

// TotalOrders retourne le nombre de commandes
func (cs *CategoryStats) TotalOrders() int {
	return cs.totalOrders
}

// ProductStats représente les statistiques pour un produit
type ProductStats struct {
	productID    catalogdomain.ProductID
	productName  string
	totalRevenue domain.Money
	totalOrders  int
	totalQty     domain.Quantity
}

// NewProductStats crée une nouvelle instance de ProductStats
func NewProductStats(
	productID catalogdomain.ProductID,
	productName string,
	totalRevenue domain.Money,
	totalOrders int,
	totalQty domain.Quantity,
) *ProductStats {
	return &ProductStats{
		productID:    productID,
		productName:  productName,
		totalRevenue: totalRevenue,
		totalOrders:  totalOrders,
		totalQty:     totalQty,
	}
}

// ProductID retourne l'ID du produit
func (ps *ProductStats) ProductID() catalogdomain.ProductID {
	return ps.productID
}

// ProductName retourne le nom du produit
func (ps *ProductStats) ProductName() string {
	return ps.productName
}

// TotalRevenue retourne le CA du produit
func (ps *ProductStats) TotalRevenue() domain.Money {
	return ps.totalRevenue
}

// TotalOrders retourne le nombre de commandes
func (ps *ProductStats) TotalOrders() int {
	return ps.totalOrders
}

// TotalQuantity retourne la quantité totale vendue
func (ps *ProductStats) TotalQuantity() domain.Quantity {
	return ps.totalQty
}

// StoreStats représente les statistiques pour un magasin
type StoreStats struct {
	storeID      ordersdomain.StoreID
	storeName    string
	totalRevenue domain.Money
	totalOrders  int
}

// NewStoreStats crée une nouvelle instance de StoreStats
func NewStoreStats(
	storeID ordersdomain.StoreID,
	storeName string,
	totalRevenue domain.Money,
	totalOrders int,
) *StoreStats {
	return &StoreStats{
		storeID:      storeID,
		storeName:    storeName,
		totalRevenue: totalRevenue,
		totalOrders:  totalOrders,
	}
}

// StoreID retourne l'ID du magasin
func (ss *StoreStats) StoreID() ordersdomain.StoreID {
	return ss.storeID
}

// StoreName retourne le nom du magasin
func (ss *StoreStats) StoreName() string {
	return ss.storeName
}

// TotalRevenue retourne le CA du magasin
func (ss *StoreStats) TotalRevenue() domain.Money {
	return ss.totalRevenue
}

// TotalOrders retourne le nombre de commandes
func (ss *StoreStats) TotalOrders() int {
	return ss.totalOrders
}

// PaymentMethodStats représente les statistiques pour un moyen de paiement
type PaymentMethodStats struct {
	paymentMethodID   ordersdomain.PaymentMethodID
	paymentMethodName string
	totalRevenue      domain.Money
	totalOrders       int
	percentage        float64
}

// NewPaymentMethodStats crée une nouvelle instance de PaymentMethodStats
func NewPaymentMethodStats(
	paymentMethodID ordersdomain.PaymentMethodID,
	paymentMethodName string,
	totalRevenue domain.Money,
	totalOrders int,
	percentage float64,
) *PaymentMethodStats {
	return &PaymentMethodStats{
		paymentMethodID:   paymentMethodID,
		paymentMethodName: paymentMethodName,
		totalRevenue:      totalRevenue,
		totalOrders:       totalOrders,
		percentage:        percentage,
	}
}

// PaymentMethodID retourne l'ID du moyen de paiement
func (pms *PaymentMethodStats) PaymentMethodID() ordersdomain.PaymentMethodID {
	return pms.paymentMethodID
}

// PaymentMethodName retourne le nom du moyen de paiement
func (pms *PaymentMethodStats) PaymentMethodName() string {
	return pms.paymentMethodName
}

// TotalRevenue retourne le CA pour ce moyen de paiement
func (pms *PaymentMethodStats) TotalRevenue() domain.Money {
	return pms.totalRevenue
}

// TotalOrders retourne le nombre de commandes
func (pms *PaymentMethodStats) TotalOrders() int {
	return pms.totalOrders
}

// Percentage retourne le pourcentage d'utilisation
func (pms *PaymentMethodStats) Percentage() float64 {
	return pms.percentage
}
