package domain

import (
	"errors"
	"fmt"
	"time"

	"eval/internal/shared/domain"
)

// ExportFormat représente le format d'export
type ExportFormat string

const (
	ExportFormatCSV     ExportFormat = "CSV"
	ExportFormatParquet ExportFormat = "Parquet"
)

// ExportType représente le type d'export
type ExportType string

const (
	ExportTypeSales ExportType = "sales"
	ExportTypeStats ExportType = "stats"
)

// ExportJob représente un job d'export
type ExportJob struct {
	format     ExportFormat
	exportType ExportType
	dateRange  domain.DateRange
	createdAt  time.Time
}

// NewExportJob crée un nouveau job d'export avec validation
func NewExportJob(
	format ExportFormat,
	exportType ExportType,
	dateRange domain.DateRange,
) (*ExportJob, error) {
	if format != ExportFormatCSV && format != ExportFormatParquet {
		return nil, errors.New("invalid export format")
	}
	if exportType != ExportTypeSales && exportType != ExportTypeStats {
		return nil, errors.New("invalid export type")
	}

	return &ExportJob{
		format:     format,
		exportType: exportType,
		dateRange:  dateRange,
		createdAt:  time.Now(),
	}, nil
}

// Format retourne le format d'export
func (ej *ExportJob) Format() ExportFormat {
	return ej.format
}

// ExportType retourne le type d'export
func (ej *ExportJob) ExportType() ExportType {
	return ej.exportType
}

// DateRange retourne la période d'export
func (ej *ExportJob) DateRange() domain.DateRange {
	return ej.dateRange
}

// CreatedAt retourne la date de création
func (ej *ExportJob) CreatedAt() time.Time {
	return ej.createdAt
}

// SaleExportRow représente une ligne d'export de vente
type SaleExportRow struct {
	OrderID       int64
	CustomerID    int64
	StoreID       int64
	StoreName     string
	ProductID     int64
	ProductName   string
	CategoryName  string
	Quantity      int
	UnitPrice     float64
	Subtotal      float64
	PaymentMethod string
	PromotionCode string
	OrderDate     time.Time
}

// NewSaleExportRow crée une nouvelle ligne d'export
// C’est ce qu’on appelle un constructeur idiomatique en Go.
// Même si Go n’a pas de mot-clé constructor, on suit ce pattern pour plusieurs raisons :
// tu as un point unique pour créer les objets (pratique si tu veux ajouter des validations ou conversions plus tard)
func NewSaleExportRow(
	orderID, customerID, storeID, productID int64,
	storeName, productName, categoryName string,
	quantity int,
	unitPrice, subtotal float64,
	paymentMethod, promotionCode string,
	orderDate time.Time,
) *SaleExportRow {
	return &SaleExportRow{
		OrderID:       orderID,
		CustomerID:    customerID,
		StoreID:       storeID,
		StoreName:     storeName,
		ProductID:     productID,
		ProductName:   productName,
		CategoryName:  categoryName,
		Quantity:      quantity,
		UnitPrice:     unitPrice,
		Subtotal:      subtotal,
		PaymentMethod: paymentMethod,
		PromotionCode: promotionCode,
		OrderDate:     orderDate,
	}
}

// ToCSVRow convertit en tableau pour CSV
func (ser *SaleExportRow) ToCSVRow() []string {
	return []string{
		fmt.Sprintf("%d", ser.OrderID),
		fmt.Sprintf("%d", ser.CustomerID),
		fmt.Sprintf("%d", ser.StoreID),
		ser.StoreName,
		fmt.Sprintf("%d", ser.ProductID),
		ser.ProductName,
		ser.CategoryName,
		fmt.Sprintf("%d", ser.Quantity),
		fmt.Sprintf("%.2f", ser.UnitPrice),
		fmt.Sprintf("%.2f", ser.Subtotal),
		ser.PaymentMethod,
		ser.PromotionCode,
		ser.OrderDate.Format("2006-01-02 15:04:05"),
	}
}

// CSVHeaders retourne les en-têtes CSV
func CSVHeaders() []string {
	return []string{
		"order_id",
		"customer_id",
		"store_id",
		"store_name",
		"product_id",
		"product_name",
		"category_name",
		"quantity",
		"unit_price",
		"subtotal",
		"payment_method",
		"promotion_code",
		"order_date",
	}
}
