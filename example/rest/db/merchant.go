package db

type Merchant struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	TeamID int    `json:"teamId"`
}

type Merchants struct {
	Merchants []Merchant `json:"merchants,omitempty"`
}

type CreateMerchantParams struct {
	Name   string `json:"name"`
	TeamID int    `json:"teamId"`
}

type UpdateMerchantParams struct {
	Name   string `json:"name"`
	TeamID int    `json:"teamId"`
}
