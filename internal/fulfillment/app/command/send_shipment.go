package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/fulfillment/domain/shipment"
)

type SendShipment struct {
	ShipmentID string
}

type SendShipmentHandler struct {
	Shipments shipment.Repository
}

func (h SendShipmentHandler) Handle(ctx context.Context, cmd SendShipment) error {
	if err := h.Shipments.Send(ctx, shipment.ShipmentUUID(cmd.ShipmentID)); err != nil {
		return fmt.Errorf("send shipment: %w", err)
	}
	return nil
}
