package model

type Address struct {
	Country      string `json:"country"`
	City         string `json:"city"`
	Village      string `json:"village"`
	Town         string `json:"town"`
	District     string `json:"district"`
	Prefix       string `json:"prefix"`
	Street       string `json:"street"`
	HouseNumber  string `json:"housenumber"`
	Name         string `json:"name"`
	OldName      string `json:"old_name"`
	HouseName    string `json:"housename"`
	PostCode     string `json:"postcode"`
	Intersection bool   `json:"intersection"`
	Shape        []byte `json:"shape"`
}
