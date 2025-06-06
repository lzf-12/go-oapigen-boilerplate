package order

import "context"

type OrderImpl struct {
	Orders map[int64]Order

	// add dependency here
}

var _ StrictServerInterface = (*OrderImpl)(nil)

func NewUserHandler() *OrderImpl {
	return &OrderImpl{
		Orders: make(map[int64]Order),
	}
}

func (o *OrderImpl) GetOrder(ctx context.Context, request GetOrderRequestObject) (GetOrderResponseObject, error) {

	var result []Order

	// implement get orders from db or cache here

	// placeholder boilerplate
	amount := 100
	id := "123"
	userid := "userid"

	result = append(result, Order{
		Amount: float32(amount),
		Id:     id,
		UserId: userid,
	})

	return GetOrder200JSONResponse(result), nil
}
