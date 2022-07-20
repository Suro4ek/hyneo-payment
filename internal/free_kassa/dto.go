package freekassa

type FreeKassa struct {
	MerchantID        string `form:"MERCHANT_ID"`
	Amount            string `form:"AMOUNT"`
	Merchant_order_id string `form:"MERCHANT_ORDER_ID"`
	SIGN              string `form:"SIGN"`
}
