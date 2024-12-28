package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/testutil"
)

func TestRegisterLocation(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	req := httptest.NewRequest(http.MethodPost, "/location", strings.NewReader(`{
  "name": "Gent Zuid",
  "address": "F.Rooseveltlaan 3A",
  "city": "Gent",
  "party_id": "TWK",
  "postal_code": "9000",
  "country": "BEL",
  "country_code": "BEL",
  "coordinates": {
    "latitude": "51.047599",
    "longitude": "3.729944"
  },
  "parking_type": "ON_STREET"
}`))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var res store.Location
	err := json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	want := &store.Location{
		Address: "F.Rooseveltlaan 3A",
		City:    "Gent",
		Coordinates: store.GeoLocation{
			Latitude:  "51.047599",
			Longitude: "3.729944",
		},
		Country:     "BEL",
		CountryCode: "BEL",
		Id:          res.Id,
		Name:        testutil.StringPtr("Gent Zuid"),
		ParkingType: testutil.StringPtr("ON_STREET"),
		PostalCode:  testutil.StringPtr("9000"),
		PartyId:     "TWK",
	}

	// Assert response
	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	assert.Equal(t, *want, res)

	// Assert location in db
	got, err := engine.LookupLocation(context.Background(), res.Id)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
