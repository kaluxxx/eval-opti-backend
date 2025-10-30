package infrastructure

import (
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

// FindByID trouve un produit par son ID
func (r *ProductQueryRepository) FindByID(id domain.ProductID) (*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.supplier_id, p.base_price, p.stock_quantity, p.created_at
		FROM products p
		WHERE p.id = $1
	`
	/*
		var ici déclare plusieurs variables locales.
		Stack : allocation ultra rapide (pointeur de pile déplacé).
		Pas de GC : ces variables sont détruites automatiquement à la fin de la fonction.
	*/
	var (
		pid        int64
		name       string
		supplierID int64
		basePrice  float64
		stockQty   int
		createdAt  time.Time
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
