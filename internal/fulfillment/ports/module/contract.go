package module

type FulfillmentModule interface{}

func NewFulfillmentModule() FulfillmentModule {
	return struct{}{}
}
