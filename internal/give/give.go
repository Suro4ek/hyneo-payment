package give

import "hyneo-payment/internal/model"

type Give interface {
	Give(order model.Order) error
}
