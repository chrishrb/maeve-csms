// SPDX-License-Identifier: Apache-2.0

//go:build integration

package firestore_test

// Test for transaction.go

import (
	"context"
	"k8s.io/utils/clock"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/store/firestore"
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

func TestFindTransactionDoesNotExist(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	transactionStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer transactionStore.CloseConn()
	require.NoError(t, err)

	got, err := transactionStore.LookupTransaction(ctx, "unknown", "ids")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestCreateAndFindTransaction(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	transactionStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer transactionStore.CloseConn()
	require.NoError(t, err)

	meterValues := NewMeterValues(100)

	err = transactionStore.CreateTransaction(ctx, "cs001", "1234", idToken, tokenType, meterValues, 0, false)
	assert.NoError(t, err)

	got, err := transactionStore.LookupTransaction(ctx, "cs001", "1234")
	assert.NoError(t, err)

	want := &store.Transaction{
		ChargeStationId: "cs001",
		TransactionId:   "1234",
		IdToken:         idToken,
		TokenType:       tokenType,
		MeterValues:     meterValues,
		StartSeqNo:      0,
	}

	assert.Equal(t, want, got)
}

func TestCreateTransactionWithExistingTransaction(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	transactionStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer transactionStore.CloseConn()
	require.NoError(t, err)

	meterValues1 := NewMeterValues(100)

	err = transactionStore.CreateTransaction(ctx, "cs002", "1234", idToken, tokenType, meterValues1, 0, false)
	assert.NoError(t, err)

	meterValues2 := NewMeterValues(200)

	err = transactionStore.CreateTransaction(ctx, "cs002", "1234", idToken, tokenType, meterValues2, 0, false)
	assert.NoError(t, err)

	got, err := transactionStore.LookupTransaction(ctx, "cs002", "1234")
	assert.NoError(t, err)

	want := &store.Transaction{
		ChargeStationId: "cs002",
		TransactionId:   "1234",
		IdToken:         idToken,
		TokenType:       tokenType,
		MeterValues:     append(meterValues1, meterValues2...),
		StartSeqNo:      0,
	}

	assert.Equal(t, want, got)
}

func TestTransactionStoreGetAllTransactionsByChargeStation(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	transactionStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer transactionStore.CloseConn()
	require.NoError(t, err)

	transactionsBefore, err := transactionStore.ListTransactionsByChargeStation(ctx, "cs006", 0, 20)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(transactionsBefore))

	meterValues := NewMeterValues(100)
	err = transactionStore.CreateTransaction(ctx, "cs006", "1234", idToken, tokenType, meterValues, 0, false)
	assert.NoError(t, err)

	err = transactionStore.CreateTransaction(ctx, "cs006", "4567", idToken, tokenType, meterValues, 0, false)
	assert.NoError(t, err)

	err = transactionStore.CreateTransaction(ctx, "cs006", "8912", idToken, tokenType, meterValues, 0, false)
	assert.NoError(t, err)

	err = transactionStore.CreateTransaction(ctx, "cs002", "4444", idToken, tokenType, meterValues, 0, false)
	assert.NoError(t, err)

	err = transactionStore.CreateTransaction(ctx, "cs009", "5555", idToken, tokenType, meterValues, 0, false)
	assert.NoError(t, err)

	transactionsAfter, err := transactionStore.ListTransactionsByChargeStation(ctx, "cs006", 0, 20)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(transactionsAfter))
}

func TestTransactionStoreUpdateCreatedTransaction(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	transactionStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer transactionStore.CloseConn()
	require.NoError(t, err)

	meterValues1 := NewMeterValues(100)

	err = transactionStore.CreateTransaction(ctx, "cs003", "1234", idToken, tokenType, meterValues1, 0, false)
	assert.NoError(t, err)

	meterValues2 := NewMeterValues(200)

	err = transactionStore.UpdateTransaction(ctx, "cs003", "1234", meterValues2)
	assert.NoError(t, err)

	got, err := transactionStore.LookupTransaction(ctx, "cs003", "1234")
	assert.NoError(t, err)

	want := &store.Transaction{
		ChargeStationId:   "cs003",
		TransactionId:     "1234",
		IdToken:           idToken,
		TokenType:         tokenType,
		MeterValues:       append(meterValues1, meterValues2...),
		UpdatedSeqNoCount: 1,
	}

	assert.Equal(t, want, got)
}

func TestTransactionStoreEndTransaction(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	transactionStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer transactionStore.CloseConn()
	require.NoError(t, err)

	meterValues1 := NewMeterValues(100)
	err = transactionStore.CreateTransaction(ctx, "cs004", "1234", idToken, tokenType, meterValues1, 0, false)
	assert.NoError(t, err)

	meterValues2 := NewMeterValues(200)
	err = transactionStore.UpdateTransaction(ctx, "cs004", "1234", meterValues2)
	assert.NoError(t, err)

	meterValues3 := NewMeterValues(200)
	err = transactionStore.EndTransaction(ctx, "cs004", "1234", idToken, tokenType, meterValues3, 2)
	assert.NoError(t, err)

	got, err := transactionStore.LookupTransaction(ctx, "cs004", "1234")
	assert.NoError(t, err)

	want := &store.Transaction{
		ChargeStationId:   "cs004",
		TransactionId:     "1234",
		IdToken:           idToken,
		TokenType:         tokenType,
		MeterValues:       append(meterValues1, append(meterValues2, meterValues3...)...),
		StartSeqNo:        0,
		EndedSeqNo:        2,
		UpdatedSeqNoCount: 1,
		Offline:           false,
	}

	assert.Equal(t, want, got)
}

func TestTransactionStoreEndNonExistingTransaction(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	transactionStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer transactionStore.CloseConn()
	require.NoError(t, err)

	meterValues := NewMeterValues(100)
	err = transactionStore.EndTransaction(ctx, "cs005", "1234", idToken, tokenType, meterValues, 2)
	assert.NoError(t, err)

	got, err := transactionStore.LookupTransaction(ctx, "cs005", "1234")
	assert.NoError(t, err)

	want := &store.Transaction{
		ChargeStationId:   "cs005",
		TransactionId:     "1234",
		IdToken:           idToken,
		TokenType:         tokenType,
		MeterValues:       meterValues,
		StartSeqNo:        0,
		EndedSeqNo:        2,
		UpdatedSeqNoCount: 0,
		Offline:           false,
	}

	assert.Equal(t, want, got)
}
