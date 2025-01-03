// SPDX-License-Identifier: Apache-2.0

//go:build integration

package firestore_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/store/firestore"
	"github.com/thoughtworks/maeve-csms/manager/testutil"
	"golang.org/x/net/context"
	"k8s.io/utils/clock"
)

func TestSetAndLookupLocation(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()
	locationStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer locationStore.CloseConn()
	require.NoError(t, err)

	want := &store.Location{
		Id:      "loc001",
		Address: "F.Rooseveltlaan 3A",
		City:    "Gent",
		Coordinates: store.GeoLocation{
			Latitude:  "51.047599",
			Longitude: "3.729944",
		},
		Country:     "BEL",
		Name:        testutil.StringPtr("Gent Zuid"),
		ParkingType: testutil.StringPtr("ON_STREET"),
		PostalCode:  testutil.StringPtr("9000"),
	}
	err = locationStore.CreateLocation(ctx, want)
	require.NoError(t, err)
	assert.NotEmpty(t, "loc001")

	got, err := locationStore.LookupLocation(ctx, "loc001")
	require.NoError(t, err)

	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, got.LastUpdated)
	got.LastUpdated = ""

	assert.Equal(t, want, got)
}

func TestListLocations(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()
	locationStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer locationStore.CloseConn()
	require.NoError(t, err)

	locations := make([]*store.Location, 20)
	for i := 0; i < 20; i++ {
		locations[i] = &store.Location{
			Id:      fmt.Sprintf("loc%02d", i),
			Address: "Randomstreet 3A",
			City:    "Randomtown",
			Coordinates: store.GeoLocation{
				Latitude:  fmt.Sprintf("%f", rand.Float32()*90),
				Longitude: fmt.Sprintf("%f", rand.Float32()*180),
			},
			Country:     "RAND",
			Name:        testutil.StringPtr("Random Location"),
			ParkingType: testutil.StringPtr("ON_STREET"),
			PostalCode:  testutil.StringPtr("12345"),
		}
	}

	for _, loc := range locations {
		err = locationStore.CreateLocation(ctx, loc)
		require.NoError(t, err)
	}

	got, err := locationStore.ListLocations(ctx, 0, 10)
	require.NoError(t, err)

	assert.Equal(t, 10, len(got))
	if !cmp.Equal(locations[:10], got, cmpopts.IgnoreFields(store.Location{}, "LastUpdated")) {
		t.Errorf("locations list wrong")
	}
}
