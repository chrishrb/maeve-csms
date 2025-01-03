package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/thoughtworks/maeve-csms/manager/store"
)

func (s *Server) RegisterLocation(w http.ResponseWriter, r *http.Request) {
	if s.ocpi == nil {
		_ = render.Render(w, r, ErrNotFound)
		return
	}

	req := new(Location)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	var parkingType *string
	if req.ParkingType != nil {
		pt := string(*req.ParkingType)
		parkingType = &pt
	}
	err := s.store.CreateLocation(r.Context(), &store.Location{
		Id:      req.Id,
		Address: req.Address,
		City:    req.City,
		Coordinates: store.GeoLocation{
			Latitude:  req.Coordinates.Latitude,
			Longitude: req.Coordinates.Longitude,
		},
		Country:     req.Country,
		CountryCode: req.CountryCode,
		Name:        req.Name,
		ParkingType: parkingType,
		PostalCode:  req.PostalCode,
		PartyId:     req.PartyId,
	})
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = render.Render(w, r, req)
}

func (s *Server) ListLocations(w http.ResponseWriter, r *http.Request, params ListLocationsParams) {
	offset, limit := getPaginationDefaults(params.Offset, params.Limit)

	locations, err := s.store.ListLocations(r.Context(), offset, limit)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	var resp = make([]render.Renderer, len(locations))
	for i, loc := range locations {
		coords := GeoLocation{
			Latitude:  loc.Coordinates.Latitude,
			Longitude: loc.Coordinates.Longitude,
		}
		var parkingType *LocationParkingType
		if loc.ParkingType != nil {
			pt := LocationParkingType(*loc.ParkingType)
			parkingType = &pt
		}
		resp[i] = Location{
			Id:          loc.Id,
			Address:     loc.Address,
			City:        loc.City,
			Coordinates: coords,
			Country:     loc.Country,
			CountryCode: loc.CountryCode,
			PartyId:     loc.PartyId,
			Name:        loc.Name,
			ParkingType: parkingType,
			PostalCode:  loc.PostalCode,
		}
	}

	_ = render.RenderList(w, r, resp)
}

func (s *Server) LookupLocation(w http.ResponseWriter, r *http.Request, locationId string) {
	loc, err := s.store.LookupLocation(r.Context(), locationId)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	coords := GeoLocation{
		Latitude:  loc.Coordinates.Latitude,
		Longitude: loc.Coordinates.Longitude,
	}
	var parkingType *LocationParkingType
	if loc.ParkingType != nil {
		pt := LocationParkingType(*loc.ParkingType)
		parkingType = &pt
	}
	resp := Location{
		Id:          loc.Id,
		Address:     loc.Address,
		City:        loc.City,
		Coordinates: coords,
		Country:     loc.Country,
		CountryCode: loc.CountryCode,
		PartyId:     loc.PartyId,
		Name:        loc.Name,
		ParkingType: parkingType,
		PostalCode:  loc.PostalCode,
	}

	_ = render.Render(w, r, resp)
}

func (s *Server) UpdateLocation(w http.ResponseWriter, r *http.Request, locationId string) {
	// TODO
}

func (s *Server) DeleteLocation(w http.ResponseWriter, r *http.Request, locationId string) {
	// TODO
}
