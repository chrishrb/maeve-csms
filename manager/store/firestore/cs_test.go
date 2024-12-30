// SPDX-License-Identifier: Apache-2.0

//go:build integration

package firestore_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"k8s.io/utils/clock"
	clockTest "k8s.io/utils/clock/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"github.com/thoughtworks/maeve-csms/manager/store/firestore"
	"github.com/thoughtworks/maeve-csms/manager/testutil"
)

func TestSetAndLookupChargeStation(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	csStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer csStore.CloseConn()
	require.NoError(t, err)

	want := &store.ChargeStation{
		LocationId: "location001",
		Evses: &[]store.Evse{
			{
				Connectors: []store.Connector{
					{
						Format:      "Type2",
						Id:          "1",
						MaxAmperage: 32,
						MaxVoltage:  400,
						PowerType:   "AC",
						Standard:    "IEC62196",
						LastUpdated: time.Now().Format(time.RFC3339),
					},
				},
				EvseId:      testutil.StringPtr("EVSE1"),
				Status:      "Available",
				Uid:         "UID1",
				LastUpdated: time.Now().Format(time.RFC3339),
			},
			{
				Connectors: []store.Connector{
					{
						Format:      "Type2",
						Id:          "2",
						MaxAmperage: 32,
						MaxVoltage:  400,
						PowerType:   "AC",
						Standard:    "IEC62196",
						LastUpdated: time.Now().Format(time.RFC3339),
					},
				},
				EvseId:      testutil.StringPtr("EVSE2"),
				Status:      "Available",
				Uid:         "UID2",
				LastUpdated: time.Now().Format(time.RFC3339),
			},
		},
		SecurityProfile:      store.TLSWithClientSideCertificates,
		Base64SHA256Password: "DEADBEEF",
	}

	cs, err := csStore.CreateChargeStation(ctx, want)
	require.NoError(t, err)
	assert.NotEmpty(t, cs.Id)

	got, err := csStore.LookupChargeStation(ctx, cs.Id)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestListChargeStations(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	csStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer csStore.CloseConn()
	require.NoError(t, err)

	chargeStations := []*store.ChargeStation{
		{
			LocationId: "location001",
			Evses: &[]store.Evse{
				{
					Connectors: []store.Connector{
						{
							Format:      "Type2",
							Id:          "1",
							MaxAmperage: 32,
							MaxVoltage:  400,
							PowerType:   "AC",
							Standard:    "IEC62196",
							LastUpdated: time.Now().Format(time.RFC3339),
						},
					},
					EvseId:      testutil.StringPtr("EVSE1"),
					Status:      "Available",
					Uid:         "UID1",
					LastUpdated: time.Now().Format(time.RFC3339),
				},
			},
			SecurityProfile:      store.TLSWithClientSideCertificates,
			Base64SHA256Password: "DEADBEEF",
		},
		{
			LocationId: "location002",
			Evses: &[]store.Evse{
				{
					Connectors: []store.Connector{
						{
							Format:      "Type2",
							Id:          "2",
							MaxAmperage: 32,
							MaxVoltage:  400,
							PowerType:   "AC",
							Standard:    "IEC62196",
							LastUpdated: time.Now().Format(time.RFC3339),
						},
					},
					EvseId:      testutil.StringPtr("EVSE2"),
					Status:      "Available",
					Uid:         "UID2",
					LastUpdated: time.Now().Format(time.RFC3339),
				},
			},
			SecurityProfile:      store.TLSWithClientSideCertificates,
			Base64SHA256Password: "DEADBEEF",
		},
	}

	for _, cs := range chargeStations {
		_, err := csStore.CreateChargeStation(ctx, cs)
		require.NoError(t, err)
	}

	got, err := csStore.ListChargeStations(ctx, 0, 20)
	require.NoError(t, err)

	assert.Len(t, got, len(chargeStations))
}

func TestLookupChargeStationWithUnregisteredChargeStation(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	csStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer csStore.CloseConn()
	require.NoError(t, err)

	got, err := csStore.LookupChargeStation(ctx, "not-created")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestUpdateAndLookupChargeStationSettingsWithNewSettings(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	now := time.Now()
	settingsStore, err := firestore.NewStore(ctx, "myproject", clockTest.NewFakePassiveClock(now))
	defer settingsStore.CloseConn()
	require.NoError(t, err)

	want := &store.ChargeStationSettings{
		ChargeStationId: "cs001",
		Settings: map[string]*store.ChargeStationSetting{
			"foo": {Value: "bar", Status: store.ChargeStationSettingStatusPending},
			"baz": {Value: "qux", Status: store.ChargeStationSettingStatusPending},
		},
	}

	err = settingsStore.UpdateChargeStationSettings(context.Background(), "cs001", want)
	require.NoError(t, err)

	got, err := settingsStore.LookupChargeStationSettings(context.Background(), "cs001")
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestUpdateAndLookupChargeStationSettingsWithUpdatedSettings(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	settingsStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer settingsStore.CloseConn()
	require.NoError(t, err)

	want := &store.ChargeStationSettings{
		ChargeStationId: "cs001",
		Settings: map[string]*store.ChargeStationSetting{
			"foo": {Value: "bar", Status: store.ChargeStationSettingStatusPending},
			"baz": {Value: "qux", Status: store.ChargeStationSettingStatusAccepted},
		},
	}

	err = settingsStore.UpdateChargeStationSettings(context.Background(), "cs001", &store.ChargeStationSettings{
		Settings: map[string]*store.ChargeStationSetting{
			"foo": {Value: "bar", Status: store.ChargeStationSettingStatusPending},
			"baz": {Value: "qux", Status: store.ChargeStationSettingStatusPending},
		},
	})
	require.NoError(t, err)

	err = settingsStore.UpdateChargeStationSettings(context.Background(), "cs001", &store.ChargeStationSettings{
		Settings: map[string]*store.ChargeStationSetting{
			"baz": {Value: "qux", Status: store.ChargeStationSettingStatusAccepted},
		},
	})
	require.NoError(t, err)

	got, err := settingsStore.LookupChargeStationSettings(context.Background(), "cs001")
	require.NoError(t, err)

	assert.Equal(t, want.ChargeStationId, got.ChargeStationId)
	assert.Len(t, got.Settings, len(want.Settings))
	assert.Equal(t, store.ChargeStationSettingStatusPending, got.Settings["foo"].Status)
	assert.Equal(t, store.ChargeStationSettingStatusAccepted, got.Settings["baz"].Status)
}

func TestListChargeStationSettings(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	now := time.Now()
	settingsStore, err := firestore.NewStore(ctx, "myproject", clockTest.NewFakePassiveClock(now))
	defer settingsStore.CloseConn()
	require.NoError(t, err)

	want := &store.ChargeStationSettings{
		Settings: map[string]*store.ChargeStationSetting{
			"foo": {Value: "bar", Status: store.ChargeStationSettingStatusPending},
			"baz": {Value: "qux", Status: store.ChargeStationSettingStatusPending},
		},
	}
	for i := 0; i < 25; i++ {
		csId := fmt.Sprintf("cs%03d", i)
		err := settingsStore.UpdateChargeStationSettings(ctx, csId, want)
		require.NoError(t, err)
	}

	csIds := make(map[string]struct{})

	page1, err := settingsStore.ListChargeStationSettings(ctx, 10, "")
	require.NoError(t, err)
	require.Len(t, page1, 10)
	for _, got := range page1 {
		csIds[got.ChargeStationId] = struct{}{}
		assert.Equal(t, want.Settings, got.Settings)
	}

	page2, err := settingsStore.ListChargeStationSettings(ctx, 10, page1[len(page1)-1].ChargeStationId)
	require.NoError(t, err)
	require.Len(t, page2, 10)
	for _, got := range page2 {
		csIds[got.ChargeStationId] = struct{}{}
		assert.Equal(t, want.Settings, got.Settings)
	}

	page3, err := settingsStore.ListChargeStationSettings(ctx, 10, page2[len(page2)-1].ChargeStationId)
	require.NoError(t, err)
	require.Len(t, page3, 5)
	for _, got := range page3 {
		csIds[got.ChargeStationId] = struct{}{}
		assert.Equal(t, want.Settings, got.Settings)
	}

	assert.Len(t, csIds, 25)
}

func TestUpdateAndLookupChargeStationInstallCertificates(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	installCertsStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer installCertsStore.CloseConn()
	require.NoError(t, err)

	err = installCertsStore.UpdateChargeStationInstallCertificates(ctx, "cs001", &store.ChargeStationInstallCertificates{
		Certificates: []*store.ChargeStationInstallCertificate{
			{
				CertificateType:               store.CertificateTypeChargeStation,
				CertificateId:                 "csms001",
				CertificateData:               "csms-pem-data",
				CertificateInstallationStatus: store.CertificateInstallationPending,
			},
			{
				CertificateType:               store.CertificateTypeV2G,
				CertificateId:                 "v2g001",
				CertificateData:               "v2g-pem-data",
				CertificateInstallationStatus: store.CertificateInstallationAccepted,
			},
		},
	})
	require.NoError(t, err)

	err = installCertsStore.UpdateChargeStationInstallCertificates(ctx, "cs001", &store.ChargeStationInstallCertificates{
		Certificates: []*store.ChargeStationInstallCertificate{
			{
				CertificateType:               store.CertificateTypeChargeStation,
				CertificateId:                 "csms001",
				CertificateData:               "csms-pem-data",
				CertificateInstallationStatus: store.CertificateInstallationAccepted,
			},
			{
				CertificateType:               store.CertificateTypeEVCC,
				CertificateId:                 "evcc001",
				CertificateData:               "evcc-pem-data",
				CertificateInstallationStatus: store.CertificateInstallationPending,
			},
		},
	})
	require.NoError(t, err)

	got, err := installCertsStore.LookupChargeStationInstallCertificates(ctx, "cs001")
	require.NoError(t, err)

	assert.Len(t, got.Certificates, 3)
	for _, cert := range got.Certificates {
		switch cert.CertificateId {
		case "csms001":
			assert.Equal(t, "csms-pem-data", cert.CertificateData)
			assert.Equal(t, store.CertificateInstallationAccepted, cert.CertificateInstallationStatus)
		case "v2g001":
			assert.Equal(t, "v2g-pem-data", cert.CertificateData)
			assert.Equal(t, store.CertificateInstallationAccepted, cert.CertificateInstallationStatus)
		case "evcc001":
			assert.Equal(t, "evcc-pem-data", cert.CertificateData)
			assert.Equal(t, store.CertificateInstallationPending, cert.CertificateInstallationStatus)
		default:
			t.Errorf("unexpected certificate id: %s", cert.CertificateId)
		}
	}
}

func TestListChargeStationInstallCertificates(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	now := time.Now()
	certInstallStore, err := firestore.NewStore(ctx, "myproject", clockTest.NewFakePassiveClock(now))
	defer certInstallStore.CloseConn()
	require.NoError(t, err)

	want := &store.ChargeStationInstallCertificates{
		Certificates: []*store.ChargeStationInstallCertificate{
			{
				CertificateType:               store.CertificateTypeV2G,
				CertificateId:                 "v2g001",
				CertificateData:               "v2g-pem-data",
				CertificateInstallationStatus: store.CertificateInstallationPending,
				SendAfter:                     time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	for i := 0; i < 25; i++ {
		csId := fmt.Sprintf("cs%03d", i)
		err := certInstallStore.UpdateChargeStationInstallCertificates(ctx, csId, want)
		require.NoError(t, err)
	}

	csIds := make(map[string]struct{})

	page1, err := certInstallStore.ListChargeStationInstallCertificates(ctx, 10, "")
	require.NoError(t, err)
	require.Len(t, page1, 10)
	for _, got := range page1 {
		csIds[got.ChargeStationId] = struct{}{}
		assert.Equal(t, want.Certificates, got.Certificates)
	}

	page2, err := certInstallStore.ListChargeStationInstallCertificates(ctx, 10, page1[len(page1)-1].ChargeStationId)
	require.NoError(t, err)
	require.Len(t, page2, 10)
	for _, got := range page2 {
		csIds[got.ChargeStationId] = struct{}{}
		assert.Equal(t, want.Certificates, got.Certificates)
	}

	page3, err := certInstallStore.ListChargeStationInstallCertificates(ctx, 10, page2[len(page2)-1].ChargeStationId)
	require.NoError(t, err)
	require.Len(t, page3, 5)
	for _, got := range page3 {
		csIds[got.ChargeStationId] = struct{}{}
		assert.Equal(t, want.Certificates, got.Certificates)
	}

	assert.Len(t, csIds, 25)
}

func TestSetAndLookupChargeStationRuntimeDetails(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	detailsStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer detailsStore.CloseConn()
	require.NoError(t, err)

	want := &store.ChargeStationRuntimeDetails{
		OcppVersion: "1.6",
	}

	err = detailsStore.SetChargeStationRuntimeDetails(ctx, "cs001", want)
	require.NoError(t, err)

	got, err := detailsStore.LookupChargeStationRuntimeDetails(ctx, "cs001")
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestLookupChargeStationRuntimeDetailsWithUnregisteredChargeStation(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	detailsStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer detailsStore.CloseConn()
	require.NoError(t, err)

	got, err := detailsStore.LookupChargeStationRuntimeDetails(ctx, "not-created")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestListChargeStationTriggerMessages(t *testing.T) {
	defer cleanupAllCollections(t, "myproject")

	ctx := context.Background()

	triggerStore, err := firestore.NewStore(ctx, "myproject", clock.RealClock{})
	defer triggerStore.CloseConn()
	require.NoError(t, err)

	err = triggerStore.SetChargeStationTriggerMessage(ctx, "cs001", &store.ChargeStationTriggerMessage{
		TriggerMessage: store.TriggerMessageBootNotification,
		TriggerStatus:  store.TriggerStatusPending,
	})
	require.NoError(t, err)

	got, err := triggerStore.ListChargeStationTriggerMessages(ctx, 10, "")
	require.NoError(t, err)

	t.Logf("%+v", got)
}
