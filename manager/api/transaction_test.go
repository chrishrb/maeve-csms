package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/store"
)

func makePtr[T any](t T) *T {
	v := t
	return &v
}

const idToken = "SOMERFID"
const tokenType = "ISO14443"

func NewMeterValues(energyReactiveExportValue float32) []store.MeterValue {
	return []store.MeterValue{
		{
			Timestamp: time.Now().Format(time.RFC3339),
			SampledValues: []store.SampledValue{
				{
					Measurand: makePtr("Energy.Active.Import.Register"),
					Value:     energyReactiveExportValue,
				},
			},
		},
	}
}

func TestLookupTransaction(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	ctx := context.Background()

	meterValues1 := NewMeterValues(100)
	err := engine.CreateTransaction(ctx, "cs003", "1234", idToken, tokenType, meterValues1, 0, false)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/transactions/%s/%s", "cs003", "1234"), strings.NewReader("{}"))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var res store.Transaction
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	want := &store.Transaction{
		ChargeStationId: "cs003",
		TransactionId:   "1234",
		IdToken:         idToken,
		TokenType:       tokenType,
		MeterValues:     meterValues1,
		StartSeqNo:      0,
	}

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Equal(t, *want, res)
}

func TestListTransactionsByChargePoint(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	ctx := context.Background()

	meterValues1 := NewMeterValues(100)
	err := engine.CreateTransaction(ctx, "cs003", "1234", idToken, tokenType, meterValues1, 0, false)
	require.NoError(t, err)

	meterValues2 := NewMeterValues(200)
	err = engine.CreateTransaction(ctx, "cs003", "4567", idToken, tokenType, meterValues2, 0, false)
	require.NoError(t, err)

	err = engine.CreateTransaction(ctx, "cs002", "12121", idToken, tokenType, meterValues2, 0, false)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/transactions/%s", "cs003"), strings.NewReader("{}"))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var res []store.Transaction
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	want := []store.Transaction{
		{
			ChargeStationId: "cs003",
			TransactionId:   "1234",
			IdToken:         idToken,
			TokenType:       tokenType,
			MeterValues:     meterValues1,
			StartSeqNo:      0,
		},
		{
			ChargeStationId: "cs003",
			TransactionId:   "4567",
			IdToken:         idToken,
			TokenType:       tokenType,
			MeterValues:     meterValues2,
			StartSeqNo:      0,
		},
	}

	assert.Len(t, res, 2)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.ElementsMatch(t, want, res)
}
