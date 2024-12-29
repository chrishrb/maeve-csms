package api_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/api"
	"github.com/thoughtworks/maeve-csms/manager/ocpi"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/store/inmemory"
	"k8s.io/utils/clock"
	clockTest "k8s.io/utils/clock/testing"
)

func setupServer(t *testing.T) (*httptest.Server, *chi.Mux, store.Engine, clock.PassiveClock) {
	engine := inmemory.NewStore(clock.RealClock{})
	ocpiApi := ocpi.NewOCPI(engine, nil, "GB", "TWK")

	now := time.Now().UTC()
	c := clockTest.NewFakePassiveClock(now)
	srv, err := api.NewServer(engine, c, ocpiApi)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Use(api.ValidationMiddleware("/").Handler)
	r.Mount("/", api.Handler(srv))
	server := httptest.NewServer(r)

	return server, r, engine, c
}
