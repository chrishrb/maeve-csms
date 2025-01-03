package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/testutil"
)

func TestRegisterChargeStation(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	// Create location
	err := engine.CreateLocation(context.Background(), &store.Location{
		Id:      "loc001",
		Address: "F.Rooseveltlaan 3A",
		City:    "Gent",
		Coordinates: store.GeoLocation{
			Latitude:  "51.047599",
			Longitude: "3.729944",
		},
		Country:     "BEL",
		CountryCode: "BEL",
		Name:        testutil.StringPtr("Gent Zuid"),
		ParkingType: testutil.StringPtr("ON_STREET"),
		PostalCode:  testutil.StringPtr("9000"),
		PartyId:     "TWK",
	})
	require.NoError(t, err)

	// Create cs
	req := httptest.NewRequest(http.MethodPost, "/cs", strings.NewReader(fmt.Sprintf(`{"security_profile": 0, "location_id": "%s"}`, "loc001")))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var res store.ChargeStation
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	want := &store.ChargeStation{
		Id:              res.Id,
		SecurityProfile: 0,
		LocationId:      "loc001",
	}

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	assert.Equal(t, *want, res)
}

func TestLookupChargeStation(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	// Given
	err := engine.CreateLocation(context.Background(), &store.Location{
		Id:      "loc001",
		Address: "F.Rooseveltlaan 3A",
		City:    "Gent",
		Coordinates: store.GeoLocation{
			Latitude:  "51.047599",
			Longitude: "3.729944",
		},
		Country:     "BEL",
		CountryCode: "BEL",
		Name:        testutil.StringPtr("Gent Zuid"),
		ParkingType: testutil.StringPtr("ON_STREET"),
		PostalCode:  testutil.StringPtr("9000"),
		PartyId:     "TWK",
	})
	require.NoError(t, err)

	err = engine.CreateChargeStation(context.Background(), &store.ChargeStation{
		Id:              "cs001",
		SecurityProfile: 1,
		LocationId:      "loc001",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/cs/%s", "cs001"), strings.NewReader("{}"))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var res store.ChargeStation
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	want := &store.ChargeStation{
		Id:              res.Id,
		SecurityProfile: 1,
		LocationId:      "loc001",
		Evses:           &[]store.Evse{},
	}

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Equal(t, *want, res)
}

func TestListChargeStations(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	// Create location
	err := engine.CreateLocation(context.Background(), &store.Location{
		Id:      "loc001",
		Address: "F.Rooseveltlaan 3A",
		City:    "Gent",
		Coordinates: store.GeoLocation{
			Latitude:  "51.047599",
			Longitude: "3.729944",
		},
		Country:     "BEL",
		CountryCode: "BEL",
		Name:        testutil.StringPtr("Gent Zuid"),
		ParkingType: testutil.StringPtr("ON_STREET"),
		PostalCode:  testutil.StringPtr("9000"),
		PartyId:     "TWK",
	})
	require.NoError(t, err)

	// Create charge stations
	cs1 := &store.ChargeStation{
		Id:              "cs1",
		SecurityProfile: 1,
		LocationId:      "loc001",
		Evses:           &[]store.Evse{},
	}
	err = engine.CreateChargeStation(context.Background(), cs1)
	require.NoError(t, err)

	cs2 := &store.ChargeStation{
		Id:              "cs2",
		SecurityProfile: 2,
		LocationId:      "loc001",
		Evses:           &[]store.Evse{},
	}
	err = engine.CreateChargeStation(context.Background(), cs2)
	require.NoError(t, err)

	// List charge stations
	req := httptest.NewRequest(http.MethodGet, "/cs", nil)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var res []store.ChargeStation
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Set Id
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Len(t, res, 2)
	assert.Contains(t, res, *cs1)
	assert.Contains(t, res, *cs2)
}

func TestLookupChargeStationThatDoesNotExist(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/cs/unknown", strings.NewReader("{}"))
	req.Header.Set("accept", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestLookupChargeStationRuntimeDetails(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	err := engine.SetChargeStationRuntimeDetails(context.Background(), "cs001", &store.ChargeStationRuntimeDetails{
		OcppVersion: "2.0.1",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/cs/cs001/runtime-details", strings.NewReader("{}"))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var res store.ChargeStationRuntimeDetails
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	want := &store.ChargeStationRuntimeDetails{
		OcppVersion: res.OcppVersion,
	}

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Equal(t, *want, res)
}

func TestLookupChargeStationRuntimeDetailsThatDoesNotExist(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/cs/cs001/runtime-details", strings.NewReader("{}"))
	req.Header.Set("accept", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}
