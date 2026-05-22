package product

import "modular_monolith/internal/shared/bizid"

type ProductUUID string

func (uuid ProductUUID) String() string { return string(uuid) }

type Product struct {
	uuid         ProductUUID
	name         string
	priceCents   int64
	stock        int
	reservations map[string]int
}

func NewProduct(name string, priceCents int64, stock int) (*Product, error) {
	if name == "" || priceCents <= 0 || stock < 0 {
		return nil, ErrInvalidProduct
	}
	return &Product{
		uuid:         ProductUUID(bizid.New()),
		name:         name,
		priceCents:   priceCents,
		stock:        stock,
		reservations: make(map[string]int),
	}, nil
}

func Rehydrate(uuid ProductUUID, name string, priceCents int64, stock int, reservations map[string]int) *Product {
	if reservations == nil {
		reservations = make(map[string]int)
	}
	return &Product{uuid: uuid, name: name, priceCents: priceCents, stock: stock, reservations: reservations}
}

func (p *Product) ReserveStock(orderUUID string, qty int) error {
	if qty <= 0 || orderUUID == "" {
		return ErrInvalidProduct
	}
	if _, exists := p.reservations[orderUUID]; exists {
		return nil
	}
	if p.stock < qty {
		return ErrInsufficientStock
	}
	p.stock -= qty
	p.reservations[orderUUID] = qty
	return nil
}

func (p *Product) UUID() ProductUUID { return p.uuid }
func (p *Product) Name() string      { return p.name }
func (p *Product) PriceCents() int64 { return p.priceCents }
func (p *Product) Stock() int        { return p.stock }
func (p *Product) Reservations() map[string]int {
	out := make(map[string]int, len(p.reservations))
	for k, v := range p.reservations {
		out[k] = v
	}
	return out
}
