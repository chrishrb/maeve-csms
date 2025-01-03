// SPDX-License-Identifier: Apache-2.0

package ocpi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/handlers"
	"github.com/thoughtworks/maeve-csms/manager/handlers/ocpp16"
	"github.com/thoughtworks/maeve-csms/manager/handlers/ocpp201"
	"github.com/thoughtworks/maeve-csms/manager/ocpi"
	"github.com/thoughtworks/maeve-csms/manager/services"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/store/inmemory"
	"github.com/thoughtworks/maeve-csms/manager/transport"
	"k8s.io/utils/clock"
	fakeclock "k8s.io/utils/clock/testing"
)

func setupHandler(t *testing.T) (http.Handler, store.Engine, time.Time) {
	engine := inmemory.NewStore(clock.RealClock{})
	err := engine.SetRegistrationDetails(context.Background(), "123", &store.OcpiRegistration{
		Status: store.OcpiRegistrationStatusRegistered,
	})
	require.NoError(t, err)

	ocpiApi := ocpi.NewOCPI(engine, http.DefaultClient, "GB", "TWK")
	v16CallMaker := newNoopV16CallMaker()
	v201CallMaker := newNoopV201CallMaker()
	now := time.Now().UTC()
	evseUidSvc := services.NewEvseUIDService(`^([A-Z]{2})\*([A-Z0-9]{3})\*E([0-9]+)\*?(.*)$`)
	server, err := ocpi.NewServer(ocpiApi, fakeclock.NewFakePassiveClock(now), v16CallMaker, v201CallMaker, evseUidSvc)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Mount("/", ocpi.Handler(server))
	return r, engine, now
}

func TestServerGetVersions(t *testing.T) {
	handler, _, now := setupHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/ocpi/versions", nil)
	req.Header.Set("Authorization", "Token 123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	want := ocpi.OcpiResponseListVersion{
		Data: &[]ocpi.Version{
			{
				Version: "2.2",
				Url:     "/ocpi/2.2",
			},
		},
		StatusCode:    ocpi.StatusSuccess,
		StatusMessage: &ocpi.StatusSuccessMessage,
		Timestamp:     now.Format(time.RFC3339),
	}

	var got ocpi.OcpiResponseListVersion
	err = json.Unmarshal(b, &got)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestServerGetVersion(t *testing.T) {
	handler, _, now := setupHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/ocpi/2.2", nil)
	req.Header.Set("Authorization", "Token 123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	want := ocpi.OcpiResponseVersionDetail{
		Data: &ocpi.VersionDetail{
			Endpoints: []ocpi.Endpoint{
				{
					Identifier: "credentials",
					Url:        "/ocpi/2.2/credentials",
					Role:       ocpi.RECEIVER,
				},
				{
					Identifier: "commands",
					Url:        "/ocpi/receiver/2.2/commands",
					Role:       ocpi.RECEIVER,
				},
				{
					Identifier: "tokens",
					Url:        "/ocpi/receiver/2.2/tokens/",
					Role:       ocpi.RECEIVER,
				},
			},
			Version: "2.2",
		},
		StatusCode:    ocpi.StatusSuccess,
		StatusMessage: &ocpi.StatusSuccessMessage,
		Timestamp:     now.Format(time.RFC3339),
	}

	var got ocpi.OcpiResponseVersionDetail
	err = json.Unmarshal(b, &got)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestServerGetClientOwnedToken(t *testing.T) {
	handler, engine, now := setupHandler(t)

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "DEADBEEF",
		ContractId:  "GBTWKTWTW000018",
		Issuer:      "Thoughtworks",
		Valid:       true,
		CacheMode:   "ALWAYS",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/ocpi/receiver/2.2/tokens/GB/TWK/DEADBEEF", nil)
	req.Header.Set("Authorization", "Token 123")
	req.Header.Set("X-Request-ID", "123")
	req.Header.Set("X-Correlation-ID", "123")
	req.Header.Set("OCPI-from-country-code", "GB")
	req.Header.Set("OCPI-from-party-id", "TWK")
	req.Header.Set("OCPI-to-country-code", "GB")
	req.Header.Set("OCPI-to-party-id", "TWK")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	want := ocpi.OcpiResponseToken{
		Data: &ocpi.Token{
			ContractId:  "GBTWKTWTW000018",
			CountryCode: "GB",
			Issuer:      "Thoughtworks",
			PartyId:     "TWK",
			Type:        "RFID",
			Uid:         "DEADBEEF",
			Valid:       true,
			Whitelist:   "ALWAYS",
		},
		StatusCode:    ocpi.StatusSuccess,
		StatusMessage: &ocpi.StatusSuccessMessage,
		Timestamp:     now.Format(time.RFC3339),
	}

	var got ocpi.OcpiResponseToken
	err = json.Unmarshal(b, &got)
	require.NoError(t, err)

	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, got.Data.LastUpdated)
	got.Data.LastUpdated = ""
	assert.Equal(t, want, got)
}

func TestServerPutClientOwnedToken(t *testing.T) {
	handler, engine, _ := setupHandler(t)

	tok := ocpi.Token{
		ContractId:  "GBTWKTWTW000018",
		CountryCode: "GB",
		Issuer:      "Thoughtworks",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "DEADBEEF",
		Valid:       true,
		Whitelist:   "ALWAYS",
	}
	b, err := json.Marshal(tok)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/ocpi/receiver/2.2/tokens/GB/TWK/DEADBEEF", bytes.NewReader(b))
	req.Header.Set("Authorization", "Token 123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "123")
	req.Header.Set("X-Correlation-ID", "123")
	req.Header.Set("OCPI-from-country-code", "GB")
	req.Header.Set("OCPI-from-party-id", "TWK")
	req.Header.Set("OCPI-to-country-code", "GB")
	req.Header.Set("OCPI-to-party-id", "TWK")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	want := &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "DEADBEEF",
		ContractId:  "GBTWKTWTW000018",
		Issuer:      "Thoughtworks",
		Valid:       true,
		CacheMode:   "ALWAYS",
	}

	got, err := engine.LookupToken(context.Background(), "DEADBEEF")
	require.NoError(t, err)
	got.LastUpdated = ""
	assert.Equal(t, want, got)
}

func TestServerPatchClientOwnedToken(t *testing.T) {
	handler, engine, _ := setupHandler(t)

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "DEADBEEF",
		ContractId:  "GBTWKTWTW000018",
		Issuer:      "Thoughtworks",
		Valid:       true,
		CacheMode:   "ALWAYS",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/ocpi/receiver/2.2/tokens/GB/TWK/DEADBEEF",
		strings.NewReader(`{
			"contract_id": "GBTWKTWTW000025",
			"issuer": "TW",
			"valid": false,
			"whitelist": "NEVER"
		}`))
	req.Header.Set("Authorization", "Token 123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "123")
	req.Header.Set("X-Correlation-ID", "123")
	req.Header.Set("OCPI-from-country-code", "GB")
	req.Header.Set("OCPI-from-party-id", "TWK")
	req.Header.Set("OCPI-to-country-code", "GB")
	req.Header.Set("OCPI-to-party-id", "TWK")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	want := &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "DEADBEEF",
		ContractId:  "GBTWKTWTW000025",
		Issuer:      "TW",
		Valid:       false,
		CacheMode:   "NEVER",
	}

	got, err := engine.LookupToken(context.Background(), "DEADBEEF")
	require.NoError(t, err)
	got.LastUpdated = ""
	assert.Equal(t, want, got)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("%s", string(b))
}

func TestPostStartSession16(t *testing.T) {
	handler, engine, _ := setupHandler(t)

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "DEADBEEF",
		ContractId:  "GBTWKTWTW000018",
		Issuer:      "Thoughtworks",
		Valid:       true,
		CacheMode:   "ALWAYS",
	})
	require.NoError(t, err)

	err = engine.SetChargeStationRuntimeDetails(context.Background(), "00188", &store.ChargeStationRuntimeDetails{
		OcppVersion: store.OcppVersion16,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ocpi/receiver/2.2/commands/START_SESSION",
		strings.NewReader(`{
			"response_url": "https://example.com/ocpi/receiver/2.2/commands/START_SESSION/12345",
			"evse_uid": "DE*GCE*E00188*001",
			"connector_id": "2",
			"token": {	
				"type": "APP_USER",
				"uid": "DEADBEEF",
				"whitelist": "NEVER",
				"country_code": "GB",
				"party_id": "TWK",
				"contract_id": "GBTWKTWTW000018",
				"issuer": "Thoughtworks",
				"valid": true
			},
			"location_id": "loc001",
			"authorization_reference": "56789"
		}`))
	req.Header.Set("Authorization", "Token 123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "123")
	req.Header.Set("X-Correlation-ID", "123")
	req.Header.Set("OCPI-from-country-code", "GB")
	req.Header.Set("OCPI-from-party-id", "TWK")
	req.Header.Set("OCPI-to-country-code", "GB")
	req.Header.Set("OCPI-to-party-id", "TWK")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var ocpiResponseCommandResponse ocpi.OcpiResponseCommandResponse
	err = json.Unmarshal(b, &ocpiResponseCommandResponse)
	require.NoError(t, err)
	assert.Equal(t, ocpi.StatusSuccess, ocpiResponseCommandResponse.StatusCode)
	t.Logf("%s", string(b))
	require.NotNilf(t, ocpiResponseCommandResponse.Data, "ocpiResponseCommandResponse.Data should not be nil")
	assert.Equal(t, ocpi.CommandResponseResultACCEPTED, ocpiResponseCommandResponse.Data.Result)
}

func TestPostStartSession201(t *testing.T) {
	handler, engine, _ := setupHandler(t)

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "DEADBEEF",
		ContractId:  "GBTWKTWTW000018",
		Issuer:      "Thoughtworks",
		Valid:       true,
		CacheMode:   "ALWAYS",
	})
	require.NoError(t, err)

	err = engine.SetChargeStationRuntimeDetails(context.Background(), "041503001", &store.ChargeStationRuntimeDetails{
		OcppVersion: store.OcppVersion201,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ocpi/receiver/2.2/commands/START_SESSION",
		strings.NewReader(`{
			"response_url": "https://example.com/ocpi/receiver/2.2/commands/START_SESSION/12345",
			"evse_uid": "BE*BEC*E041503001",
			"connector_id": "2",
			"token": {	
				"type": "APP_USER",
				"uid": "DEADBEEF",
				"whitelist": "NEVER",
				"country_code": "GB",
				"party_id": "TWK",
				"contract_id": "GBTWKTWTW000018",
				"issuer": "Thoughtworks",
				"valid": true
			},
			"location_id": "loc001",
			"authorization_reference": "56789"
		}`))
	req.Header.Set("Authorization", "Token 123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "123")
	req.Header.Set("X-Correlation-ID", "123")
	req.Header.Set("OCPI-from-country-code", "GB")
	req.Header.Set("OCPI-from-party-id", "TWK")
	req.Header.Set("OCPI-to-country-code", "GB")
	req.Header.Set("OCPI-to-party-id", "TWK")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var ocpiResponseCommandResponse ocpi.OcpiResponseCommandResponse
	err = json.Unmarshal(b, &ocpiResponseCommandResponse)
	require.NoError(t, err)
	assert.Equal(t, ocpi.StatusSuccess, ocpiResponseCommandResponse.StatusCode)
	t.Logf("%s", string(b))
	require.NotNilf(t, ocpiResponseCommandResponse.Data, "ocpiResponseCommandResponse.Data should not be nil")
	assert.Equal(t, ocpi.CommandResponseResultACCEPTED, ocpiResponseCommandResponse.Data.Result)
}

func newNoopV16CallMaker() *handlers.OcppCallMaker {
	emitter := transport.EmitterFunc(func(ctx context.Context, ocppVersion transport.OcppVersion, chargeStationId string, message *transport.Message) error {
		return nil
	})
	return ocpp16.NewCallMaker(emitter)
}

func newNoopV201CallMaker() *handlers.OcppCallMaker {
	emitter := transport.EmitterFunc(func(ctx context.Context, ocppVersion transport.OcppVersion, chargeStationId string, message *transport.Message) error {
		return nil
	})
	return ocpp201.NewCallMaker(emitter)
}
