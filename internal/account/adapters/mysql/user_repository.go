package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"modular_monolith/internal/account/domain/user"
	"modular_monolith/internal/platform/dbtx"
)

type UserModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID      string `gorm:"type:char(36);not null;uniqueIndex"`
	Name      string `gorm:"size:255;not null"`
	Email     string `gorm:"size:255;not null;uniqueIndex"`
	Status    string `gorm:"size:32;not null;default:active;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AddressModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID      string `gorm:"type:char(36);not null;uniqueIndex"`
	UserID    string `gorm:"type:char(36);not null;index"`
	Receiver  string `gorm:"size:255;not null"`
	Phone     string `gorm:"size:64;not null"`
	City      string `gorm:"size:255;not null"`
	Detail    string `gorm:"size:512;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository struct {
	db      *gorm.DB
	txBound bool
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func NewUserRepositoryWithTx(tx *gorm.DB) *UserRepository {
	return &UserRepository{db: tx, txBound: true}
}

func Models() []any {
	return []any{&UserModel{}, &AddressModel{}}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	return r.executeWrite(ctx, "account.Repository.Save", func(tx *gorm.DB) error {
		model := UserModel{UUID: u.UUID().String(), Name: u.Name(), Email: u.Email(), Status: string(u.Status())}
		if err := tx.Create(&model).Error; err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		return nil
	})
}

func (r *UserRepository) FindByUUID(ctx context.Context, uuid user.UserUUID) (*user.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", uuid.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.NewUserNotFound(err)
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return user.Rehydrate(user.UserUUID(model.UUID), model.Name, model.Email, user.Status(model.Status), nil), nil
}

func (r *UserRepository) SaveAddress(ctx context.Context, address user.Address) error {
	return r.executeWrite(ctx, "account.Repository.SaveAddress", func(tx *gorm.DB) error {
		model := toAddressModel(address)
		if err := tx.Create(&model).Error; err != nil {
			return fmt.Errorf("create address: %w", err)
		}
		return nil
	})
}

func (r *UserRepository) UpdateAddress(ctx context.Context, address user.Address) error {
	return r.executeWrite(ctx, "account.Repository.UpdateAddress", func(tx *gorm.DB) error {
		model := toAddressModel(address)
		result := tx.Model(&AddressModel{}).
			Where("uuid = ? AND user_id = ?", address.UUID().String(), address.UserUUID().String()).
			Updates(map[string]any{
				"receiver": model.Receiver,
				"phone":    model.Phone,
				"city":     model.City,
				"detail":   model.Detail,
			})
		if result.Error != nil {
			return fmt.Errorf("update address: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return user.ErrAddressNotFound
		}
		return nil
	})
}

func (r *UserRepository) DeleteAddress(ctx context.Context, userUUID user.UserUUID, addressUUID user.AddressUUID) error {
	return r.executeWrite(ctx, "account.Repository.DeleteAddress", func(tx *gorm.DB) error {
		result := tx.Where("uuid = ? AND user_id = ?", addressUUID.String(), userUUID.String()).Delete(&AddressModel{})
		if result.Error != nil {
			return fmt.Errorf("delete address: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return user.ErrAddressNotFound
		}
		return nil
	})
}

func (r *UserRepository) FindAddress(ctx context.Context, userUUID user.UserUUID, addressUUID user.AddressUUID) (user.Address, error) {
	var model AddressModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ? AND user_id = ?", addressUUID.String(), userUUID.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user.Address{}, user.NewAddressNotFound(err)
		}
		return user.Address{}, fmt.Errorf("find address: %w", err)
	}
	return toAddress(model), nil
}

func toAddressModel(address user.Address) AddressModel {
	return AddressModel{
		UUID:     address.UUID().String(),
		UserID:   address.UserUUID().String(),
		Receiver: address.Receiver(),
		Phone:    address.Phone(),
		City:     address.City(),
		Detail:   address.Detail(),
	}
}

func toAddress(model AddressModel) user.Address {
	return user.RehydrateAddress(
		user.AddressUUID(model.UUID),
		user.UserUUID(model.UserID),
		model.Receiver,
		model.Phone,
		model.City,
		model.Detail,
	)
}

func (r *UserRepository) executeWrite(ctx context.Context, op string, fn func(tx *gorm.DB) error) error {
	return dbtx.ExecuteWrite(ctx, r.db, r.txBound, op, fn)
}
