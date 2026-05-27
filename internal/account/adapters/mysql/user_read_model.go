package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"modular_monolith/internal/account/app/query"
	"modular_monolith/internal/account/domain/user"
)

type ReadModel struct {
	db *gorm.DB
}

func NewReadModel(db *gorm.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (r *ReadModel) GetUser(ctx context.Context, userID string) (query.UserDTO, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.UserDTO{}, user.NewUserNotFound(err)
		}
		return query.UserDTO{}, fmt.Errorf("get user: %w", err)
	}
	return query.UserDTO{
		ID:        model.UUID,
		Name:      model.Name,
		Email:     model.Email,
		Status:    model.Status,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *ReadModel) ListAddresses(ctx context.Context, userID string) ([]query.AddressDTO, error) {
	var models []AddressModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	items := make([]query.AddressDTO, 0, len(models))
	for _, model := range models {
		items = append(items, addressDTO(model))
	}
	return items, nil
}

func (r *ReadModel) GetAddress(ctx context.Context, userID string, addressID string) (query.AddressDTO, error) {
	var model AddressModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ? AND user_id = ?", addressID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.AddressDTO{}, user.NewAddressNotFound(err)
		}
		return query.AddressDTO{}, fmt.Errorf("get address: %w", err)
	}
	return addressDTO(model), nil
}

func addressDTO(model AddressModel) query.AddressDTO {
	return query.AddressDTO{
		ID:        model.UUID,
		UserID:    model.UserID,
		Receiver:  model.Receiver,
		Phone:     model.Phone,
		City:      model.City,
		Detail:    model.Detail,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
