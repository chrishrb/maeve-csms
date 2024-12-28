package store

import "context"

type GeoLocation struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type Location struct {
	Address     string      `json:"address"`
	City        string      `json:"city"`
	Coordinates GeoLocation `json:"coordinates"`
	Country     string      `json:"country"`
	CountryCode string      `json:"country_code"`
	Id          string      `json:"id"`
	LastUpdated string      `json:"last_updated"`
	Name        *string     `json:"name,omitempty"`
	ParkingType *string     `json:"parking_type,omitempty"`
	PostalCode  *string     `json:"postal_code,omitempty"`
	PartyId     string      `json:"party_id"`
}

type LocationStore interface {
	CreateLocation(ctx context.Context, location *Location) (*Location, error)
	UpdateLocation(ctx context.Context, locationId string, location *Location) (*Location, error)
	DeleteLocation(ctx context.Context, locationId string) error
	LookupLocation(ctx context.Context, locationId string) (*Location, error)
	ListLocations(context context.Context, offset int, limit int) ([]*Location, error)
}
