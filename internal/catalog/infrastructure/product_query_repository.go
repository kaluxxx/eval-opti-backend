package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"eval/internal/catalog/domain"
	shareddomain "eval/internal/shared/domain"
	"eval/internal/shared/infrastructure"
)

// ProductQueryRepository repository pour les requêtes de lecture sur les produits
type ProductQueryRepository struct {
	infrastructure.BaseRepository
}

// NewProductQueryRepository crée un nouveau repository de lecture pour les produits
func NewProductQueryRepository(db *sql.DB) *ProductQueryRepository {
	return &ProductQueryRepository{
		BaseRepository: infrastructure.NewBaseRepository(db),
	}
}

// WithContext ajoute un contexte
func (r *ProductQueryRepository) WithContext(ctx context.Context) *ProductQueryRepository {
	newRepo := *r
	// Note: BaseRepository devra être étendu pour supporter WithContext proprement
	return &newRepo
}

// FindByID trouve un produit par son ID
func (r *ProductQueryRepository) FindByID(id domain.ProductID) (*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.supplier_id, p.base_price, p.stock_quantity, p.created_at
		FROM products p
		WHERE p.id = $1
	`

	var (
		pid          int64
		name         string
		supplierID   int64
		basePrice    float64
		stockQty     int
		createdAt    time.Time
	)

	err := r.QueryRow(query, int64(id)).Scan(&pid, &name, &supplierID, &basePrice, &stockQty, &createdAt)
	if err != nil {
		return nil, err
	}

	// Récupérer les catégories
	categories, err := r.findCategoriesForProduct(domain.ProductID(pid))
	if err != nil {
		return nil, err
	}

	money, _ := shareddomain.NewMoney(basePrice, "EUR")
	quantity, _ := shareddomain.NewQuantity(stockQty)

	return domain.NewProduct(
		domain.ProductID(pid),
		name,
		domain.SupplierID(supplierID),
		money,
		quantity,
		categories,
		createdAt,
	)
}

// FindAll récupère tous les produits
func (r *ProductQueryRepository) FindAll() ([]*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.supplier_id, p.base_price, p.stock_quantity, p.created_at
		FROM products p
		ORDER BY p.id
	`

	rows, err := r.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var (
			pid          int64
			name         string
			supplierID   int64
			basePrice    float64
			stockQty     int
			createdAt    time.Time
		)

		if err := rows.Scan(&pid, &name, &supplierID, &basePrice, &stockQty, &createdAt); err != nil {
			return nil, err
		}

		categories, err := r.findCategoriesForProduct(domain.ProductID(pid))
		if err != nil {
			return nil, err
		}

		money, _ := shareddomain.NewMoney(basePrice, "EUR")
		quantity, _ := shareddomain.NewQuantity(stockQty)

		product, err := domain.NewProduct(
			domain.ProductID(pid),
			name,
			domain.SupplierID(supplierID),
			money,
			quantity,
			categories,
			createdAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return products, nil
}

// findCategoriesForProduct récupère les catégories d'un produit
func (r *ProductQueryRepository) findCategoriesForProduct(productID domain.ProductID) ([]domain.CategoryID, error) {
	query := `
		SELECT category_id
		FROM product_categories
		WHERE product_id = $1
	`

	rows, err := r.Query(query, int64(productID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.CategoryID
	for rows.Next() {
		var catID int64
		if err := rows.Scan(&catID); err != nil {
			return nil, err
		}
		categories = append(categories, domain.CategoryID(catID))
	}

	return categories, nil
}

// FindByIDs récupère plusieurs produits par leurs IDs
func (r *ProductQueryRepository) FindByIDs(ids []domain.ProductID) (map[domain.ProductID]*domain.Product, error) {
	if len(ids) == 0 {
		return make(map[domain.ProductID]*domain.Product), nil
	}

	// Convertir les IDs pour la requête
	intIDs := make([]interface{}, len(ids))
	for i, id := range ids {
		intIDs[i] = int64(id)
	}

	query := `
		SELECT p.id, p.name, p.supplier_id, p.base_price, p.stock_quantity, p.created_at
		FROM products p
		WHERE p.id = ANY($1)
	`

	rows, err := r.Query(query, intIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make(map[domain.ProductID]*domain.Product)
	for rows.Next() {
		var (
			pid          int64
			name         string
			supplierID   int64
			basePrice    float64
			stockQty     int
			createdAt    time.Time
		)

		if err := rows.Scan(&pid, &name, &supplierID, &basePrice, &stockQty, &createdAt); err != nil {
			return nil, err
		}

		categories, err := r.findCategoriesForProduct(domain.ProductID(pid))
		if err != nil {
			return nil, err
		}

		money, _ := shareddomain.NewMoney(basePrice, "EUR")
		quantity, _ := shareddomain.NewQuantity(stockQty)

		product, err := domain.NewProduct(
			domain.ProductID(pid),
			name,
			domain.SupplierID(supplierID),
			money,
			quantity,
			categories,
			createdAt,
		)
		if err != nil {
			return nil, err
		}

		products[domain.ProductID(pid)] = product
	}

	return products, nil
}
