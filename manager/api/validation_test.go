// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/api"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/testutil"
)

func TestValidationMiddlewareWithInvalidRequest(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	req := httptest.NewRequest(http.MethodPost, "/cs", strings.NewReader(""))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	b, err := io.ReadAll(rr.Result().Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"status":"Bad Request","error":"request body has an error: value is required but missing"}`, string(b))
}

func TestValidationMiddlewareWithValidRequest(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	// Create location
	loc, err := engine.CreateLocation(context.Background(), &store.Location{
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

	req := httptest.NewRequest(http.MethodPost, "/cs", strings.NewReader(fmt.Sprintf(`{"security_profile":0, "location_id": "%s"}`, loc.Id)))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	require.NoError(t, err)
}

func TestValidationMiddlewareWithNonApiPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(api.ValidationMiddleware)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})
	server := httptest.NewServer(r)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", strings.NewReader(""))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
}

func TestValidationMiddlewareWithUnknownMethod(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/register", strings.NewReader(""))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Result().StatusCode)
}
