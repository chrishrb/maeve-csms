// SPDX-License-Identifier: Apache-2.0

package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Store) CreateTransaction(ctx context.Context, chargeStationId, transactionId, idToken, tokenType string, meterValue []store.MeterValue, seqNo int, offline bool) error {
	transaction, err := s.LookupTransaction(ctx, chargeStationId, transactionId)
	if err != nil {
		return fmt.Errorf("getting transaction: %w", err)
	}

	if transaction != nil {
		transaction.IdToken = idToken
		transaction.TokenType = tokenType
		transaction.MeterValues = append(transaction.MeterValues, meterValue...)
		transaction.StartSeqNo = seqNo
		transaction.Offline = offline
	} else {
		transaction = &store.Transaction{
			ChargeStationId:   chargeStationId,
			TransactionId:     transactionId,
			IdToken:           idToken,
			TokenType:         tokenType,
			MeterValues:       meterValue,
			StartSeqNo:        seqNo,
			EndedSeqNo:        0,
			UpdatedSeqNoCount: 0,
			Offline:           offline,
		}
	}

	return s.updateTransaction(ctx, chargeStationId, transactionId, transaction)
}

func (s *Store) LookupTransaction(ctx context.Context, chargeStationId, transactionId string) (*store.Transaction, error) {
	transactionRef := s.client.Doc(getPath(chargeStationId, transactionId))
	snap, err := transactionRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup transaction %s/%s with code %v: %w", chargeStationId, transactionId, status.Code(err), err)
	}

	var transaction store.Transaction
	if err = snap.DataTo(&transaction); err != nil {
		return nil, fmt.Errorf("map transaction %s/%s: %w", chargeStationId, transactionId, err)
	}

	return &transaction, nil
}

func (s *Store) ListTransactionsByChargeStation(ctx context.Context, csId string, offset, limit int) ([]*store.Transaction, error) {
	transactionRefs, err := s.client.Collection(fmt.Sprintf("ChargeStation/%s/Transaction", csId)).OrderBy("TransactionId", firestore.Asc).Offset(offset).Limit(limit).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("getting transactions: %w", err)
	}

	transactions := make([]*store.Transaction, 0, len(transactionRefs))
	for _, transactionRef := range transactionRefs {
		var transaction store.Transaction
		if err = transactionRef.DataTo(&transaction); err != nil {
			return nil, fmt.Errorf("map transaction %s: %w", transactionRef.Ref.ID, err)
		}
		transactions = append(transactions, &transaction)
	}

	return transactions, nil
}

func (s *Store) UpdateTransaction(ctx context.Context, chargeStationId, transactionId string, meterValue []store.MeterValue) error {
	transaction, err := s.LookupTransaction(ctx, chargeStationId, transactionId)
	if err != nil {
		return fmt.Errorf("getting transaction: %w", err)
	}

	if transaction == nil {
		transaction = &store.Transaction{
			ChargeStationId:   chargeStationId,
			TransactionId:     transactionId,
			MeterValues:       meterValue,
			UpdatedSeqNoCount: 1,
		}
	} else {
		transaction.MeterValues = append(transaction.MeterValues, meterValue...)
		transaction.UpdatedSeqNoCount++
	}

	return s.updateTransaction(ctx, chargeStationId, transactionId, transaction)
}

func (s *Store) EndTransaction(ctx context.Context, chargeStationId, transactionId, idToken, tokenType string, meterValue []store.MeterValue, seqNo int) error {
	transaction, err := s.LookupTransaction(ctx, chargeStationId, transactionId)
	if err != nil {
		return fmt.Errorf("getting transaction: %w", err)
	}

	if transaction == nil {
		transaction = &store.Transaction{
			ChargeStationId: chargeStationId,
			TransactionId:   transactionId,
			IdToken:         idToken,
			TokenType:       tokenType,
			MeterValues:     meterValue,
			EndedSeqNo:      seqNo,
		}
	} else {
		transaction.MeterValues = append(transaction.MeterValues, meterValue...)
		transaction.EndedSeqNo = seqNo
	}

	return s.updateTransaction(ctx, chargeStationId, transactionId, transaction)
}

func (s *Store) updateTransaction(ctx context.Context, chargeStationId, transactionId string, transaction *store.Transaction) error {
	transactionRef := s.client.Doc(getPath(chargeStationId, transactionId))
	_, err := transactionRef.Set(ctx, transaction)
	if err != nil {
		return fmt.Errorf("setting transaction %s/%s: %w", chargeStationId, transactionId, err)
	}
	return nil
}

func getPath(chargeStationId, transactionId string) string {
	return fmt.Sprintf("ChargeStation/%s/Transaction/%s", chargeStationId, transactionId)
}
