// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
	"os"

	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/cors"
	"github.com/thoughtworks/maeve-csms/manager/api"
	"github.com/thoughtworks/maeve-csms/manager/config"
	"github.com/thoughtworks/maeve-csms/manager/ocpi"
	"github.com/thoughtworks/maeve-csms/manager/services"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/unrolled/secure"
	"k8s.io/utils/clock"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewApiHandler(settings config.ApiSettings, engine store.Engine, ocpi ocpi.Api, csCertProvider services.ChargeStationCertificateProvider) http.Handler {
	apiServer, err := api.NewServer(engine, clock.RealClock{}, ocpi)
	if err != nil {
		panic(err)
	}

	var isDevelopment bool
	if os.Getenv("ENVIRONMENT") == "dev" {
		isDevelopment = true
	}
	secureMiddleware := secure.New(secure.Options{
		IsDevelopment:         isDevelopment,
		BrowserXssFilter:      true,
		ContentTypeNosniff:    true,
		FrameDeny:             true,
		ContentSecurityPolicy: "frame-ancestors: 'none'",
	})

	r := chi.NewRouter()

	logger := middleware.RequestLogger(logFormatter{endpoint: "api"})
	swagger, _ := api.GetSwagger()

	r.Use(middleware.Recoverer, secureMiddleware.Handler, cors.Default().Handler)
	r.Get("/health", health)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/api/openapi.json", getApiSwaggerJson)
	r.With(logger, oapimiddleware.OapiRequestValidator(swagger)).Mount("/api/v0", api.Handler(apiServer))
	return r
}

func getApiSwaggerJson(w http.ResponseWriter, r *http.Request) {
	swagger, err := api.GetSwagger()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json, err := swagger.MarshalJSON()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(json)
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"OK"}`))
}
