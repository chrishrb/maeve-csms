package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/api"
	"github.com/thoughtworks/maeve-csms/manager/store"
)

func TestSetToken(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	token := api.Token{
		CacheMode:   "ALWAYS",
		ContractId:  "GB-TWK-012345678-V",
		CountryCode: "GB",
		Issuer:      "Thoughtworks",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "012345678",
		Valid:       true,
	}
	tokenPayload, err := json.Marshal(token)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(tokenPayload))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	b, err := io.ReadAll(rr.Result().Body)
	require.NoError(t, err)
	assert.Equal(t, "", string(b))

	want := &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "012345678",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Thoughtworks",
		Valid:       true,
		CacheMode:   "ALWAYS",
	}

	got, err := engine.LookupToken(context.Background(), "012345678")

	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, got.LastUpdated)
	got.LastUpdated = ""

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestLookupToken(t *testing.T) {
	server, r, engine, c := setupServer(t)
	now := c.Now()
	defer server.Close()

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "012345678",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Thoughtworks",
		Valid:       true,
		CacheMode:   "ALWAYS",
		LastUpdated: now.Format(time.RFC3339),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/token/012345678", nil)
	req.Header.Set("accept", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	decoder := json.NewDecoder(rr.Result().Body)
	var got api.Token
	err = decoder.Decode(&got)
	require.NoError(t, err)

	lastUpdatedStr := now.Format(time.RFC3339)
	lastUpdated, err := time.Parse(time.RFC3339, lastUpdatedStr)
	require.NoError(t, err)
	want := api.Token{
		CacheMode:   "ALWAYS",
		ContractId:  "GBTWK012345678V",
		CountryCode: "GB",
		Issuer:      "Thoughtworks",
		LastUpdated: &lastUpdated,
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "012345678",
		Valid:       true,
	}

	assert.Equal(t, want, got)
}

func TestListTokens(t *testing.T) {
	ctx := context.Background()
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	tokens := make([]*store.Token, 20)
	for i := 0; i < 20; i++ {
		tokens[i] = &store.Token{
			CountryCode: "GB",
			PartyId:     "TWK",
			Type:        "RFID",
			Uid:         fmt.Sprintf("123456%02d", i),
			ContractId:  "GBTWK012345678V",
			Issuer:      "TWK",
			Valid:       true,
			CacheMode:   store.CacheModeAllowed,
		}
	}

	for _, token := range tokens {
		err := engine.SetToken(ctx, token)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/token", nil)
	req.Header.Set("accept", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	decoder := json.NewDecoder(rr.Result().Body)
	var got []api.Token
	err := decoder.Decode(&got)
	require.NoError(t, err)

	t.Logf("got: %+v", got)
}
