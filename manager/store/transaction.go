// SPDX-License-Identifier: Apache-2.0

package store

import "context"

type Transaction struct {
	ChargeStationId   string       `json:"charge_station_id"`
	TransactionId     string       `json:"transaction_id"`
	IdToken           string       `json:"id_token"`
	TokenType         string       `json:"token_type"`
	MeterValues       []MeterValue `json:"meter_values"`
	StartSeqNo        int          `json:"start_seq_no"`
	EndedSeqNo        int          `json:"ended_seq_no"`
	UpdatedSeqNoCount int          `json:"updated_seq_no_count"`
	Offline           bool         `json:"offline"`
}

type MeterValue struct {
	SampledValues []SampledValue `json:"sampled_values"`
	Timestamp     string         `json:"timestamp"`
}

type SampledValue struct {
	Context       *string        `json:"context"`
	Location      *string        `json:"location"`
	Measurand     *string        `json:"measurand"`
	Phase         *string        `json:"phase"`
	UnitOfMeasure *UnitOfMeasure `json:"unit_of_measure"`
	Value         float32        `json:"value"`
}

type UnitOfMeasure struct {
	Unit      string `json:"unit"`
	Multipler int    `json:"multipler"`
}

type TransactionStore interface {
	ListTransactionsByChargeStation(ctx context.Context, chargeStationId string, offset int, limit int) ([]*Transaction, error)
	LookupTransaction(ctx context.Context, chargeStationId, transactionId string) (*Transaction, error)
	CreateTransaction(ctx context.Context, chargeStationId, transactionId, idToken, tokenType string, meterValue []MeterValue, seqNo int, offline bool) error
	UpdateTransaction(ctx context.Context, chargeStationId, transactionId string, meterValue []MeterValue) error
	EndTransaction(ctx context.Context, chargeStationId, transactionId, idToken, tokenType string, meterValue []MeterValue, seqNo int) error
}
