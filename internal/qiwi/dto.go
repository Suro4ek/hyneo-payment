package qiwi

type QiwiBill struct {
	Amount_currency string `json:"amount.currency"`
	Amount_value    string `json:"amount.value"`
	BillId          string `json:"billId"`
	SiteID          string `json:"siteId"`
	Status_value    string `json:"status.value"`
}
