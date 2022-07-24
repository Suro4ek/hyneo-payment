package freekassa

type FreeKassa struct {
	MerchantID        string `form:"MERCHANT_ID"`
	Amount            string `form:"AMOUNT"`
	Merchant_order_id string `form:"MERCHANT_ORDER_ID"`
	SIGN              string `form:"SIGN"`
}

type FreeKassaBill struct {
	Name    string  `json:"name" form:"name"`
	Item_id string  `json:"item_id" form:"item_id"`
	Promo   *string `json:"promo" form:"promo"`
}
