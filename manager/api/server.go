// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"time"

	"github.com/thoughtworks/maeve-csms/manager/ocpi"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/render"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"k8s.io/utils/clock"
)

type Server struct {
	store   store.Engine
	clock   clock.PassiveClock
	swagger *openapi3.T
	ocpi    ocpi.Api
}

func NewServer(engine store.Engine, clock clock.PassiveClock, ocpi ocpi.Api) (*Server, error) {
	swagger, err := GetSwagger()
	if err != nil {
		return nil, err
	}
	return &Server{
		store:   engine,
		clock:   clock,
		ocpi:    ocpi,
		swagger: swagger,
	}, nil
}

func (s *Server) RegisterParty(w http.ResponseWriter, r *http.Request) {
	if s.ocpi == nil {
		_ = render.Render(w, r, ErrNotFound)
		return
	}

	req := new(Registration)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if req.Url != nil {
		err := s.ocpi.RegisterNewParty(r.Context(), *req.Url, req.Token)
		if err != nil {
			_ = render.Render(w, r, ErrInternalError(err))
			return
		}
	} else {
		// store credentials in database
		status := store.OcpiRegistrationStatusPending
		if req.Status != nil && *req.Status == "REGISTERED" {
			status = store.OcpiRegistrationStatusRegistered
		}

		err := s.store.SetRegistrationDetails(r.Context(), req.Token, &store.OcpiRegistration{
			Status: status,
		})
		if err != nil {
			_ = render.Render(w, r, ErrInternalError(err))
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func getPaginationDefaults(paramOffset, paramLimit *int) (int, int) {
	offset := 0
	limit := 20

	if paramOffset != nil {
		offset = *paramOffset
	}
	if paramLimit != nil {
		limit = *paramLimit
	}
	if limit > 100 {
		limit = 100
	}

	return offset, limit
}

func parseTime(timeString string) (*time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return nil, err
	}
	return &t, err
}
