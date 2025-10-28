package database

import "time"

// ============================================================================
// MODÈLES DE DONNÉES - Base normalisée
// ============================================================================

// Category - Catégorie de produit
type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Supplier - Fournisseur
type Supplier struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	ContactName string    `json:"contact_name,omitempty"`
	Email       string    `json:"email,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	Address     string    `json:"address,omitempty"`
	City        string    `json:"city,omitempty"`
	Country     string    `json:"country,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Product - Produit
type Product struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	SupplierID    *int      `json:"supplier_id,omitempty"`
	BasePrice     float64   `json:"base_price"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
}

// ProductCategory - Relation N-N produit-catégorie
type ProductCategory struct {
	ProductID  int `json:"product_id"`
	CategoryID int `json:"category_id"`
}

// Customer - Client
type Customer struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Address    string    `json:"address,omitempty"`
	City       string    `json:"city,omitempty"`
	PostalCode string    `json:"postal_code,omitempty"`
	Country    string    `json:"country"`
	CreatedAt  time.Time `json:"created_at"`
}

// Store - Magasin / Point de vente
type Store struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	Region    string    `json:"region,omitempty"`
	Country   string    `json:"country"`
	Address   string    `json:"address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// PaymentMethod - Méthode de paiement
type PaymentMethod struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active"`
}

// Promotion - Promotion / Remise
type Promotion struct {
	ID              int       `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	DiscountPercent float64   `json:"discount_percent"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
}

// Order - Commande (header)
type Order struct {
	ID              int64     `json:"id"`
	CustomerID      int       `json:"customer_id"`
	StoreID         int       `json:"store_id"`
	PaymentMethodID int       `json:"payment_method_id"`
	PromotionID     *int      `json:"promotion_id,omitempty"`
	OrderDate       time.Time `json:"order_date"`
	TotalAmount     float64   `json:"total_amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// OrderItem - Ligne de commande (détail)
type OrderItem struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// MODÈLES POUR LES STATISTIQUES (API Response)
// ============================================================================

// Stats - Statistiques globales (réponse API)
type Stats struct {
	TotalCA         float64                  `json:"total_ca"`
	ParCategorie    map[string]CategoryStats `json:"par_categorie"`
	TopProduits     []ProductStat            `json:"top_produits"`
	NbVentes        int                      `json:"nb_ventes"`
	MoyenneVente    float64                  `json:"moyenne_vente"`
	NbCommandes     int                      `json:"nb_commandes,omitempty"`
	TopMagasins     []StoreStat              `json:"top_magasins,omitempty"`
	RepartitionPaiement map[string]int       `json:"repartition_paiement,omitempty"`
}

// CategoryStats - Statistiques par catégorie
type CategoryStats struct {
	CA       float64 `json:"ca"`
	NbVentes int     `json:"nb_ventes"`
}

// ProductStat - Statistiques par produit
type ProductStat struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product"`
	CA          float64 `json:"ca"`
	NbVentes    int     `json:"nb_ventes"`
}

// StoreStat - Statistiques par magasin
type StoreStat struct {
	StoreID   int     `json:"store_id"`
	StoreName string  `json:"store_name"`
	City      string  `json:"city"`
	CA        float64 `json:"ca"`
	NbVentes  int     `json:"nb_ventes"`
}

// ============================================================================
// VUES / QUERIES COMPLEXES
// ============================================================================

// SaleComplete - Vente complète avec toutes les informations (jointure complète)
type SaleComplete struct {
	SaleID          int64      `json:"sale_id"`
	OrderDate       time.Time  `json:"order_date"`
	OrderID         int64      `json:"order_id"`
	CustomerID      int        `json:"customer_id"`
	CustomerName    string     `json:"customer_name"`
	CustomerEmail   string     `json:"customer_email,omitempty"`
	ProductID       int        `json:"product_id"`
	ProductName     string     `json:"product_name"`
	StoreName       string     `json:"store_name"`
	StoreCity       string     `json:"store_city"`
	PaymentMethod   string     `json:"payment_method"`
	PromotionCode   *string    `json:"promotion_code,omitempty"`
	DiscountPercent *float64   `json:"discount_percent,omitempty"`
	Quantity        int        `json:"quantity"`
	UnitPrice       float64    `json:"unit_price"`
	Subtotal        float64    `json:"subtotal"`
	OrderTotal      float64    `json:"order_total"`
}

// ============================================================================
// MODÈLES POUR EXPORT PARQUET
// ============================================================================

// SaleParquet - Structure optimisée pour export Parquet
type SaleParquet struct {
	OrderDate     string  `parquet:"name=order_date, type=BYTE_ARRAY, convertedtype=UTF8"`
	OrderID       int64   `parquet:"name=order_id, type=INT64"`
	ProductName   string  `parquet:"name=product_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	CustomerName  string  `parquet:"name=customer_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	StoreName     string  `parquet:"name=store_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	StoreCity     string  `parquet:"name=store_city, type=BYTE_ARRAY, convertedtype=UTF8"`
	PaymentMethod string  `parquet:"name=payment_method, type=BYTE_ARRAY, convertedtype=UTF8"`
	Quantity      int32   `parquet:"name=quantity, type=INT32"`
	UnitPrice     float64 `parquet:"name=unit_price, type=DOUBLE"`
	Subtotal      float64 `parquet:"name=subtotal, type=DOUBLE"`
}
