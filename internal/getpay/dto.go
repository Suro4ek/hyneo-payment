package getpay

type GetPay struct {
	MerchantID        string `form:"WALLET_ID"`
	Amount            string `form:"SUM"`
	Merchant_order_id string `form:"ORDER_ID"`
	SIGN              string `form:"SIGN"`
}

type GetPayBill struct {
	Name    string  `json:"name" form:"name"`
	Item_id string  `json:"item_id" form:"item_id"`
	Promo   *string `json:"promo" form:"promo"`
}
