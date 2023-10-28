package db

type Team struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Merchants []Merchant `json:"merchants,omitempty"`
}

type Teams struct {
	Teams []Team `json:"teams,omitempty"`
}

type CreateTeamParams struct {
	Name string `json:"name"`
}

type UpdateTeamParams struct {
	Name string `json:"name"`
}
