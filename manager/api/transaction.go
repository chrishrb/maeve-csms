package api

import (
	"net/http"

	"github.com/go-chi/render"
)

func (s *Server) ListTransactionsByChargeStation(w http.ResponseWriter, r *http.Request, csId string, params ListTransactionsByChargeStationParams) {
	offset, limit := getPaginationDefaults(params.Offset, params.Limit)

	transactions, err := s.store.ListTransactionsByChargeStation(r.Context(), csId, offset, limit)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
	var resp = make([]render.Renderer, len(transactions))
	for t, transaction := range transactions {
		outputMeterValues := make([]MeterValue, len(transaction.MeterValues))
		for i, meterValue := range transaction.MeterValues {
			outputSampledValues := make([]SampledValue, len(meterValue.SampledValues))
			for j, sampledValue := range meterValue.SampledValues {
				var unitOfMeasure *UnitOfMeasure
				if sampledValue.UnitOfMeasure != nil {
					unitOfMeasure = &UnitOfMeasure{
						Multipler: sampledValue.UnitOfMeasure.Multipler,
						Unit:      sampledValue.UnitOfMeasure.Unit,
					}
				}

				outputSampledValues[j] = SampledValue{
					Context:       sampledValue.Context,
					Location:      sampledValue.Location,
					Measurand:     sampledValue.Measurand,
					Phase:         sampledValue.Phase,
					UnitOfMeasure: unitOfMeasure,
					Value:         sampledValue.Value,
				}
			}
			outputMeterValues[i] = MeterValue{
				SampledValues: outputSampledValues,
				Timestamp:     meterValue.Timestamp,
			}
		}

		resp[t] = Transaction{
			ChargeStationId:   transaction.ChargeStationId,
			EndedSeqNo:        transaction.EndedSeqNo,
			IdToken:           transaction.IdToken,
			MeterValues:       outputMeterValues,
			Offline:           transaction.Offline,
			StartSeqNo:        transaction.StartSeqNo,
			TokenType:         transaction.TokenType,
			TransactionId:     transaction.TransactionId,
			UpdatedSeqNoCount: transaction.UpdatedSeqNoCount,
		}
	}

	_ = render.RenderList(w, r, resp)
}

func (s *Server) LookupTransaction(w http.ResponseWriter, r *http.Request, csId, transactionId string) {
	transaction, err := s.store.LookupTransaction(r.Context(), csId, transactionId)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if transaction == nil {
		_ = render.Render(w, r, ErrNotFound)
		return
	}

	outputMeterValues := make([]MeterValue, len(transaction.MeterValues))
	for i, meterValue := range transaction.MeterValues {
		outputSampledValues := make([]SampledValue, len(meterValue.SampledValues))
		for j, sampledValue := range meterValue.SampledValues {
			var unitOfMeasure *UnitOfMeasure
			if sampledValue.UnitOfMeasure != nil {
				unitOfMeasure = &UnitOfMeasure{
					Multipler: sampledValue.UnitOfMeasure.Multipler,
					Unit:      sampledValue.UnitOfMeasure.Unit,
				}
			}

			outputSampledValues[j] = SampledValue{
				Context:       sampledValue.Context,
				Location:      sampledValue.Location,
				Measurand:     sampledValue.Measurand,
				Phase:         sampledValue.Phase,
				UnitOfMeasure: unitOfMeasure,
				Value:         sampledValue.Value,
			}
		}
		outputMeterValues[i] = MeterValue{
			SampledValues: outputSampledValues,
			Timestamp:     meterValue.Timestamp,
		}
	}

	resp := Transaction{
		ChargeStationId:   transaction.ChargeStationId,
		EndedSeqNo:        transaction.EndedSeqNo,
		IdToken:           transaction.IdToken,
		MeterValues:       outputMeterValues,
		Offline:           transaction.Offline,
		StartSeqNo:        transaction.StartSeqNo,
		TokenType:         transaction.TokenType,
		TransactionId:     transaction.TransactionId,
		UpdatedSeqNoCount: transaction.UpdatedSeqNoCount,
	}
	_ = render.Render(w, r, resp)
}
