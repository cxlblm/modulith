package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"modular_monolith/internal/platform/dbtx"
	platformmysql "modular_monolith/internal/platform/mysql"
	"modular_monolith/internal/shop/domain/product"
)

type ProductModel struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID       string `gorm:"type:char(36);not null;uniqueIndex"`
	Name       string `gorm:"size:255;not null"`
	PriceCents int64  `gorm:"not null"`
	Stock      int    `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type StockReservationModel struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	ProductUUID string `gorm:"type:char(36);not null;uniqueIndex:idx_stock_reservation_product_order"`
	OrderUUID   string `gorm:"type:char(36);not null;uniqueIndex:idx_stock_reservation_product_order"`
	Qty         int    `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductRepository struct {
	db      *gorm.DB
	txBound bool
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func NewProductRepositoryWithTx(tx *gorm.DB) *ProductRepository {
	return &ProductRepository{db: tx, txBound: true}
}

func Models() []any {
	return []any{&ProductModel{}, &StockReservationModel{}}
}

func (r *ProductRepository) Save(ctx context.Context, p *product.Product) error {
	return r.executeWrite(ctx, "shop.Repository.Save", func(tx *gorm.DB) error {
		model := ProductModel{UUID: p.UUID().String(), Name: p.Name(), PriceCents: p.PriceCents(), Stock: p.Stock()}
		if err := tx.Create(&model).Error; err != nil {
			return fmt.Errorf("create product: %w", err)
		}
		return nil
	})
}

func (r *ProductRepository) FindByUUID(ctx context.Context, uuid product.ProductUUID) (*product.Product, error) {
	var model ProductModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", uuid.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.NewProductNotFound(err)
		}
		return nil, fmt.Errorf("find product: %w", err)
	}
	return product.Rehydrate(product.ProductUUID(model.UUID), model.Name, model.PriceCents, model.Stock, nil), nil
}

func (r *ProductRepository) ReserveStock(ctx context.Context, productUUID product.ProductUUID, orderUUID string, qty int) error {
	if productUUID.String() == "" || orderUUID == "" || qty <= 0 {
		return product.ErrInvalidProduct
	}
	return r.executeWrite(ctx, "shop.Repository.ReserveStock", func(tx *gorm.DB) error {
		reservation := StockReservationModel{ProductUUID: productUUID.String(), OrderUUID: orderUUID, Qty: qty}
		if err := tx.Create(&reservation).Error; err != nil {
			if platformmysql.IsDuplicateKey(err) {
				return nil
			}
			return fmt.Errorf("create stock reservation: %w", err)
		}
		result := tx.Model(&ProductModel{}).
			Where("uuid = ? AND stock >= ?", productUUID.String(), qty).
			Update("stock", gorm.Expr("stock - ?", qty))
		if result.Error != nil {
			return fmt.Errorf("update product stock: %w", result.Error)
		}
		if result.RowsAffected == 1 {
			return nil
		}

		var existing ProductModel
		if err := tx.Select("id").First(&existing, "uuid = ?", productUUID.String()).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return product.NewProductNotFound(err)
			}
			return fmt.Errorf("find product after stock reservation miss: %w", err)
		}
		return product.ErrInsufficientStock
	})
}

func (r *ProductRepository) executeWrite(ctx context.Context, op string, fn func(tx *gorm.DB) error) error {
	return dbtx.ExecuteWrite(ctx, r.db, r.txBound, op, fn)
}
