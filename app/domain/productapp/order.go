package productapp

import (
	"errors"

	"github.com/Housiadas/backend-system/business/domain/productbus"
	"github.com/Housiadas/backend-system/business/sys/order"
	"github.com/Housiadas/backend-system/foundation/validate"
)

func parseOrder(qp QueryParams) (order.By, error) {
	const (
		orderByProductID = "product_id"
		orderByUserID    = "user_id"
		orderByName      = "name"
		orderByCost      = "cost"
		orderByQuantity  = "quantity"
	)

	var orderByFields = map[string]string{
		orderByProductID: productbus.OrderByProductID,
		orderByName:      productbus.OrderByName,
		orderByCost:      productbus.OrderByCost,
		orderByQuantity:  productbus.OrderByQuantity,
		orderByUserID:    productbus.OrderByUserID,
	}

	orderBy, err := order.Parse(qp.OrderBy, order.NewBy(orderByProductID, order.ASC))
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	orderBy.Field = orderByFields[orderBy.Field]

	return orderBy, nil
}
