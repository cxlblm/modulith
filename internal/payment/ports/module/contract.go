package module

type PaymentModule interface{}

func NewPaymentModule() PaymentModule {
	return struct{}{}
}
