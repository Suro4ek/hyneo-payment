package qiwi

type QiwiBill struct {
	SiteId             string            `json:"siteId"`
	BillId             string            `json:"billId"`
	Amount             Amount            `json:"amount"`
	Status             Status            `json:"status"`
	CustomFields       map[string]string `json:"customFields"`
	Comment            string            `json:"comment"`
	CreationDateTime   string            `json:"creationDateTime"`
	ExpirationDateTime string            `json:"expirationDateTime"`
	PayUrl             string            `json:"payUrl"`
}

type Status struct {
	Value           string `json:"value"`
	ChangedDateTime string `json:"changedDateTime"`
}

type QIWIPay struct {
	Name    string  `json:"name" form:"name"`
	Item_id string  `json:"item_id" form:"item_id"`
	Promo   *string `json:"promo" form:"promo"`
}
