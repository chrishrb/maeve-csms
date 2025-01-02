// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/thoughtworks/maeve-csms/manager/ocpp"
	types "github.com/thoughtworks/maeve-csms/manager/ocpp/ocpp16"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"golang.org/x/exp/slog"
	"k8s.io/utils/clock"
)

type StartTransactionHandler struct {
	Clock            clock.PassiveClock
	TokenStore       store.TokenStore
	TransactionStore store.TransactionStore
}

func (t StartTransactionHandler) HandleCall(ctx context.Context, chargeStationId string, request ocpp.Request) (ocpp.Response, error) {
	req := request.(*types.StartTransactionJson)

	slog.Info("starting transaction", slog.Any("request", req))

	transactionId := -1
	status := types.StartTransactionResponseJsonIdTagInfoStatusInvalid
	tok, err := t.TokenStore.LookupToken(ctx, req.IdTag)
	if err != nil {
		return nil, err
	}
	if tok != nil {
		status = types.StartTransactionResponseJsonIdTagInfoStatusAccepted
		//#nosec G404 - transaction id does not require secure random number generator
		transactionId = int(rand.Int31())
	}

	contextTransactionBegin := types.MeterValuesJsonMeterValueElemSampledValueElemContextTransactionBegin
	meterValueMeasurand := "MeterValue"
	transactionUuid := ConvertToUUID(transactionId)
	err = t.TransactionStore.CreateTransaction(ctx, chargeStationId, transactionUuid, req.IdTag, "ISO14443",
		[]store.MeterValue{
			{
				Timestamp: t.Clock.Now().Format(time.RFC3339),
				SampledValues: []store.SampledValue{
					{
						Context:   (*string)(&contextTransactionBegin),
						Measurand: &meterValueMeasurand,
						UnitOfMeasure: &store.UnitOfMeasure{
							Unit:      string(types.MeterValuesJsonMeterValueElemSampledValueElemUnitWh),
							Multipler: 1,
						},
						Value: float32(req.MeterStart),
					},
				},
			},
		}, 0, false)
	if err != nil {
		return nil, err
	}

	return &types.StartTransactionResponseJson{
		IdTagInfo: types.StartTransactionResponseJsonIdTagInfo{
			Status: status,
		},
		TransactionId: transactionId,
	}, nil
}

func ConvertToUUID(transactionId int) string {
	uuidBytes := []byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		(byte)(transactionId >> 24),
		(byte)(transactionId >> 16),
		(byte)(transactionId >> 8),
		(byte)(transactionId),
	}
	return uuid.Must(uuid.FromBytes(uuidBytes)).String()
}
