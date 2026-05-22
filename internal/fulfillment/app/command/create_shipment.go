package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/fulfillment/domain/shipment"
)

type CreateShipment struct {
	OrderID string
}

type CreateShipmentHandler struct {
	Shipments shipment.Repository
}

func (h CreateShipmentHandler) Handle(ctx context.Context, cmd CreateShipment) error {
	s, err := shipment.NewShipment(cmd.OrderID)
	if err != nil {
		return err
	}
	if err := h.Shipments.CreateForOrder(ctx, s); err != nil {
		return fmt.Errorf("create shipment for order: %w", err)
	}
	return nil
}
