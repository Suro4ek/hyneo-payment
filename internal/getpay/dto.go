package getpay

type GetPay struct {
	MerchantID        string `form:"WALLET_ID"`
	Amount            string `form:"SUM"`
	Merchant_order_id string `form:"ORDER_ID"`
	SIGN              string `form:"SIGN"`
}
