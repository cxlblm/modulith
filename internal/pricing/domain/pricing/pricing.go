package pricing

type ProductLine struct {
	ProductID      string
	ProductName    string
	UnitPriceCents int64
	Qty            int
}

type Promotion struct {
	UUID           string
	Name           string
	ThresholdCents int64
	DiscountCents  int64
	Active         bool
}

type AppliedPromotion struct {
	UUID          string
	Name          string
	DiscountCents int64
}

type QuoteItem struct {
	ProductID              string
	ProductName            string
	Qty                    int
	OriginalUnitPriceCents int64
	OriginalSubtotalCents  int64
	DiscountCents          int64
	PayableCents           int64
	AppliedPromotions      []AppliedPromotion
}

type Quote struct {
	OriginalTotalCents int64
	DiscountTotalCents int64
	PayableTotalCents  int64
	Items              []QuoteItem
	AppliedPromotions  []AppliedPromotion
}

func CalculateQuote(userID string, lines []ProductLine, promotions []Promotion) (Quote, error) {
	if userID == "" || len(lines) == 0 {
		return Quote{}, ErrInvalidPricing
	}

	items := make([]QuoteItem, 0, len(lines))
	var originalTotal int64
	for _, line := range lines {
		if line.ProductID == "" || line.ProductName == "" || line.UnitPriceCents <= 0 || line.Qty <= 0 {
			return Quote{}, ErrInvalidPricing
		}
		originalSubtotal := line.UnitPriceCents * int64(line.Qty)
		originalTotal += originalSubtotal
		items = append(items, QuoteItem{
			ProductID:              line.ProductID,
			ProductName:            line.ProductName,
			Qty:                    line.Qty,
			OriginalUnitPriceCents: line.UnitPriceCents,
			OriginalSubtotalCents:  originalSubtotal,
			PayableCents:           originalSubtotal,
			AppliedPromotions:      []AppliedPromotion{},
		})
	}

	promo, ok := bestPromotion(originalTotal, promotions)
	if !ok {
		return Quote{
			OriginalTotalCents: originalTotal,
			DiscountTotalCents: 0,
			PayableTotalCents:  originalTotal,
			Items:              items,
			AppliedPromotions:  []AppliedPromotion{},
		}, nil
	}

	discountTotal := promo.DiscountCents
	if discountTotal > originalTotal {
		discountTotal = originalTotal
	}
	allocations := allocateDiscount(originalTotal, discountTotal, items)
	for i := range items {
		items[i].DiscountCents = allocations[i]
		items[i].PayableCents = items[i].OriginalSubtotalCents - allocations[i]
		if allocations[i] > 0 {
			items[i].AppliedPromotions = []AppliedPromotion{{
				UUID:          promo.UUID,
				Name:          promo.Name,
				DiscountCents: allocations[i],
			}}
		}
	}

	applied := []AppliedPromotion{{
		UUID:          promo.UUID,
		Name:          promo.Name,
		DiscountCents: discountTotal,
	}}
	return Quote{
		OriginalTotalCents: originalTotal,
		DiscountTotalCents: discountTotal,
		PayableTotalCents:  originalTotal - discountTotal,
		Items:              items,
		AppliedPromotions:  applied,
	}, nil
}

func bestPromotion(originalTotal int64, promotions []Promotion) (Promotion, bool) {
	var best Promotion
	var found bool
	var bestDiscount int64
	for _, promo := range promotions {
		if !promo.Active || promo.UUID == "" || promo.Name == "" || promo.ThresholdCents <= 0 || promo.DiscountCents <= 0 {
			continue
		}
		if originalTotal < promo.ThresholdCents {
			continue
		}
		effectiveDiscount := promo.DiscountCents
		if effectiveDiscount > originalTotal {
			effectiveDiscount = originalTotal
		}
		if !found || effectiveDiscount > bestDiscount {
			best = promo
			bestDiscount = effectiveDiscount
			found = true
		}
	}
	return best, found
}

func allocateDiscount(originalTotal int64, discountTotal int64, items []QuoteItem) []int64 {
	allocations := make([]int64, len(items))
	var allocated int64
	for i, item := range items {
		allocations[i] = item.OriginalSubtotalCents * discountTotal / originalTotal
		allocated += allocations[i]
	}
	remainder := discountTotal - allocated
	for i := 0; remainder > 0 && i < len(allocations); i++ {
		allocations[i]++
		remainder--
	}
	return allocations
}
