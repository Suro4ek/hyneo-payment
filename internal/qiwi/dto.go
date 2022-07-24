package qiwi

type QiwiBill struct {
	Amount_currency string `json:"amount.currency"`
	Amount_value    string `json:"amount.value"`
	BillId          string `json:"billId"`
	SiteID          string `json:"siteId"`
	Status_value    string `json:"status.value"`
}

type QIWIPay struct {
	Name    string  `json:"name" form:"name"`
	Item_id string  `json:"item_id" form:"item_id"`
	Promo   *string `json:"promo" form:"promo"`
}
