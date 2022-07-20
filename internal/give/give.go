package give

type Give interface {
	Give(orderId int) error
}
