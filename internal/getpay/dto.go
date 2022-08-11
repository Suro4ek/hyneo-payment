package getpay

type GetPay struct {
	MerchantID        string `form:"WALLET_ID" query:"WALLET_ID"`
	Amount            string `form:"SUM" query:"SUM"`
	Merchant_order_id string `form:"ORDER_ID" query:"ORDER_ID"`
	SIGN              string `form:"SIGN" query:"SIGN"`
}

type GetPayBill struct {
	Name    string  `json:"name" form:"name"`
	Item_id string  `json:"item_id" form:"item_id"`
	Promo   *string `json:"promo" form:"promo"`
}
