package model

type Address struct {
	Country      string   `json:"country"`
	City         string   `json:"city"`
	Village      string   `json:"village"`
	Town         string   `json:"town"`
	District     string   `json:"district"`
	Prefix       string   `json:"prefix"`
	Street       string   `json:"street"`
	HouseNumber  string   `json:"housenumber"`
	Name         string   `json:"name"`
	Intersection bool     `json:"intersection"`
	Location     Location `json:"location"`
}
type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
