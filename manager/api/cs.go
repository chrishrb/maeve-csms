package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/render"
	handlers "github.com/thoughtworks/maeve-csms/manager/handlers/ocpp201"
	"github.com/thoughtworks/maeve-csms/manager/ocpi"
	"github.com/thoughtworks/maeve-csms/manager/store"
)

func (s *Server) ListChargeStations(w http.ResponseWriter, r *http.Request, params ListChargeStationsParams) {
	offset, limit := getPaginationDefaults(params.Offset, params.Limit)

	chargeStations, err := s.store.ListChargeStations(r.Context(), offset, limit)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	var resp = make([]render.Renderer, len(chargeStations))
	for i, cs := range chargeStations {
		var numEvses int
		if cs.Evses != nil {
			numEvses = len(*cs.Evses)
		}
		outputEvses := make([]Evse, numEvses)
		if cs.Evses != nil {
			for j, evse := range *cs.Evses {
				outputConnectors := make([]Connector, len(evse.Connectors))
				for k, connector := range evse.Connectors {
					lastUpdated, err := parseTime(connector.LastUpdated)
					if err != nil {
						_ = render.Render(w, r, ErrInternalError(err))
						return
					}
					outputConnectors[k] = Connector{
						Format:      ConnectorFormat(connector.Format),
						Id:          connector.Id,
						LastUpdated: lastUpdated,
						MaxAmperage: connector.MaxAmperage,
						MaxVoltage:  connector.MaxVoltage,
						PowerType:   ConnectorPowerType(connector.PowerType),
						Standard:    ConnectorStandard(connector.Standard),
					}
				}
				lastUpdated, err := parseTime(evse.LastUpdated)
				if err != nil {
					_ = render.Render(w, r, ErrInternalError(err))
					return
				}
				outputEvses[j] = Evse{
					Connectors:  outputConnectors,
					EvseId:      evse.EvseId,
					LastUpdated: lastUpdated,
					Status:      EvseStatus(evse.Status),
					Uid:         evse.Uid,
				}
			}
		}
		resp[i] = ChargeStation{
			Id:                     cs.Id,
			LocationId:             cs.LocationId,
			InvalidUsernameAllowed: &cs.InvalidUsernameAllowed,
			Base64SHA256Password:   &cs.Base64SHA256Password,
			SecurityProfile:        int(cs.SecurityProfile),
			Evses:                  &outputEvses,
		}
	}

	_ = render.RenderList(w, r, resp)
}

func (s *Server) LookupChargeStation(w http.ResponseWriter, r *http.Request, csId string) {
	cs, err := s.store.LookupChargeStation(r.Context(), csId)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
	if cs == nil {
		_ = render.Render(w, r, ErrNotFound)
		return
	}

	var numEvses int
	if cs.Evses != nil {
		numEvses = len(*cs.Evses)
	}
	outputEvses := make([]Evse, numEvses)
	if cs.Evses != nil {
		for j, evse := range *cs.Evses {
			outputConnectors := make([]Connector, len(evse.Connectors))
			for k, connector := range evse.Connectors {
				lastUpdated, err := parseTime(connector.LastUpdated)
				if err != nil {
					_ = render.Render(w, r, ErrInternalError(err))
					return
				}
				outputConnectors[k] = Connector{
					Format:      ConnectorFormat(connector.Format),
					Id:          connector.Id,
					LastUpdated: lastUpdated,
					MaxAmperage: connector.MaxAmperage,
					MaxVoltage:  connector.MaxVoltage,
					PowerType:   ConnectorPowerType(connector.PowerType),
					Standard:    ConnectorStandard(connector.Standard),
				}
			}
			lastUpdated, err := parseTime(evse.LastUpdated)
			if err != nil {
				_ = render.Render(w, r, ErrInternalError(err))
				return
			}
			outputEvses[j] = Evse{
				Connectors:  outputConnectors,
				EvseId:      evse.EvseId,
				LastUpdated: lastUpdated,
				Status:      EvseStatus(evse.Status),
				Uid:         evse.Uid,
			}
		}
	}

	resp := ChargeStation{
		Id:                     cs.Id,
		LocationId:             cs.LocationId,
		InvalidUsernameAllowed: &cs.InvalidUsernameAllowed,
		Base64SHA256Password:   &cs.Base64SHA256Password,
		SecurityProfile:        int(cs.SecurityProfile),
		Evses:                  &outputEvses,
	}

	_ = render.Render(w, r, resp)
}

func (s *Server) UpdateChargeStation(w http.ResponseWriter, r *http.Request, csId string) {
	// TODO
}

func (s *Server) DeleteChargeStation(w http.ResponseWriter, r *http.Request, csId string) {
	// TODO
}

func (s *Server) RegisterChargeStation(w http.ResponseWriter, r *http.Request) {
	req := new(ChargeStation)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	loc, err := s.store.LookupLocation(r.Context(), req.LocationId)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
	if loc == nil {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("location with id [%s] not found", req.LocationId)))
		return
	}

	var pwd string
	if req.Base64SHA256Password != nil {
		pwd = *req.Base64SHA256Password
	}
	invalidUsernameAllowed := false
	if req.InvalidUsernameAllowed != nil {
		invalidUsernameAllowed = *req.InvalidUsernameAllowed
	}

	// Store charge station locally
	now := s.clock.Now()
	var numEvses int
	if req.Evses != nil {
		numEvses = len(*req.Evses)
	}
	storeEvses := make([]store.Evse, numEvses)
	if numEvses != 0 {
		for i, reqEvse := range *req.Evses {
			storeConnectors := make([]store.Connector, len(reqEvse.Connectors))
			for j, reqConnector := range reqEvse.Connectors {
				storeConnectors[j] = store.Connector{
					Id:          reqConnector.Id,
					Format:      string(reqConnector.Format),
					PowerType:   string(reqConnector.PowerType),
					Standard:    string(reqConnector.Standard),
					MaxVoltage:  reqConnector.MaxVoltage,
					MaxAmperage: reqConnector.MaxAmperage,
					LastUpdated: now.Format(time.RFC3339),
				}
				storeEvses[i] = store.Evse{
					Connectors:  storeConnectors,
					EvseId:      reqEvse.EvseId,
					Status:      string(ocpi.EvseStatusUNKNOWN),
					Uid:         reqEvse.Uid,
					LastUpdated: now.Format(time.RFC3339),
				}
			}
		}
	}
	err = s.store.CreateChargeStation(r.Context(), &store.ChargeStation{
		Id:                     req.Id,
		LocationId:             req.LocationId,
		SecurityProfile:        store.SecurityProfile(req.SecurityProfile),
		Base64SHA256Password:   pwd,
		InvalidUsernameAllowed: invalidUsernameAllowed,
		Evses:                  &storeEvses,
	})
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	// Store charge station for roaming
	ocpiEvses := make([]ocpi.Evse, numEvses)
	if numEvses != 0 {
		for i, reqEvse := range *req.Evses {
			ocpiConnectors := make([]ocpi.Connector, len(reqEvse.Connectors))
			for j, reqConnector := range reqEvse.Connectors {
				ocpiConnectors[j] = ocpi.Connector{
					Id:          reqConnector.Id,
					Format:      ocpi.ConnectorFormat(reqConnector.Format),
					PowerType:   ocpi.ConnectorPowerType(reqConnector.PowerType),
					Standard:    ocpi.ConnectorStandard(reqConnector.Standard),
					MaxVoltage:  reqConnector.MaxVoltage,
					MaxAmperage: reqConnector.MaxAmperage,
					LastUpdated: now.Format(time.RFC3339),
				}
				ocpiEvses[i] = ocpi.Evse{
					Connectors:  ocpiConnectors,
					EvseId:      reqEvse.EvseId,
					Status:      ocpi.EvseStatusUNKNOWN,
					Uid:         reqEvse.Uid,
					LastUpdated: now.Format(time.RFC3339),
				}
			}
		}
	}
	var parkingType *ocpi.LocationParkingType
	if loc.ParkingType != nil {
		pt := ocpi.LocationParkingType(*loc.ParkingType)
		parkingType = &pt
	}
	err = s.ocpi.PushLocation(r.Context(), ocpi.Location{
		Id:      loc.Id,
		Address: loc.Address,
		City:    loc.City,
		Coordinates: ocpi.GeoLocation{
			Latitude:  loc.Coordinates.Latitude,
			Longitude: loc.Coordinates.Longitude,
		},
		Country:     loc.Country,
		CountryCode: loc.CountryCode,
		Evses:       &ocpiEvses,
		LastUpdated: now.Format(time.RFC3339),
		Name:        loc.Name,
		ParkingType: parkingType,
		PartyId:     loc.PartyId,
		PostalCode:  loc.PostalCode,
		Publish:     true,
	})
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = render.Render(w, r, req)
}

func (s *Server) ReconfigureChargeStation(w http.ResponseWriter, r *http.Request, csId string) {
	req := new(ChargeStationSettings)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	chargeStationSettings := make(map[string]*store.ChargeStationSetting)
	for k, v := range *req {
		chargeStationSettings[k] = &store.ChargeStationSetting{
			Value:  v,
			Status: store.ChargeStationSettingStatusPending,
		}
	}

	err := s.store.UpdateChargeStationSettings(r.Context(), csId, &store.ChargeStationSettings{
		Settings: chargeStationSettings,
	})
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
}

func (s *Server) InstallChargeStationCertificates(w http.ResponseWriter, r *http.Request, csId string) {
	req := new(ChargeStationInstallCertificates)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	var certs []*store.ChargeStationInstallCertificate
	for _, cert := range req.Certificates {
		certId, err := handlers.GetCertificateId(cert.Certificate)
		if err != nil {
			_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("invalid certificate: %w", err)))
			return
		}

		status := store.CertificateInstallationPending
		if cert.Status != nil {
			status = store.CertificateInstallationStatus(*cert.Status)
		}

		certs = append(certs, &store.ChargeStationInstallCertificate{
			CertificateType:               store.CertificateType(cert.Type),
			CertificateId:                 certId,
			CertificateData:               cert.Certificate,
			CertificateInstallationStatus: status,
		})
	}

	err := s.store.UpdateChargeStationInstallCertificates(r.Context(), csId, &store.ChargeStationInstallCertificates{
		ChargeStationId: csId,
		Certificates:    certs,
	})
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
}

func (s *Server) TriggerChargeStation(w http.ResponseWriter, r *http.Request, csId string) {
	req := new(ChargeStationTrigger)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	err := s.store.SetChargeStationTriggerMessage(r.Context(), csId, &store.ChargeStationTriggerMessage{
		TriggerMessage: store.TriggerMessage(req.Trigger),
		TriggerStatus:  store.TriggerStatusPending,
	})
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) LookupChargeStationRuntimeDetails(w http.ResponseWriter, r *http.Request, csId string) {
	csDetails, err := s.store.LookupChargeStationRuntimeDetails(r.Context(), csId)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
	if csDetails == nil {
		_ = render.Render(w, r, ErrNotFound)
		return
	}

	resp := ChargeStationRuntimeDetails{
		OcppVersion: ChargeStationRuntimeDetailsOcppVersion(csDetails.OcppVersion),
	}

	_ = render.Render(w, r, resp)
}
