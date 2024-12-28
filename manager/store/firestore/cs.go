// SPDX-License-Identifier: Apache-2.0

package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/thoughtworks/maeve-csms/manager/store"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Store) CreateChargeStation(ctx context.Context, cs *store.ChargeStation) (*store.ChargeStation, error) {
	id := uuid.NewString()
	csRef := s.client.Doc(fmt.Sprintf("ChargeStation/%s", id))
	cs.Id = id
	_, err := csRef.Set(ctx, &cs)
	if err != nil {
		return nil, fmt.Errorf("setting cs %s: %w", id, err) 
	}
	return s.LookupChargeStation(ctx, id)
}

func (s *Store) UpdateChargeStation(ctx context.Context, csId string, cs *store.ChargeStation) (*store.ChargeStation, error) {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStation/%s", csId))
	_, err := csRef.Set(ctx, &cs)
	if err != nil {
		return nil, fmt.Errorf("setting cs %s: %w", csId, err) 
	}
	return s.LookupChargeStation(ctx, csId)
}

func (s *Store) DeleteChargeStation(ctx context.Context, chargeStationId string) error {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStation/%s", chargeStationId))
	_, err := csRef.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupChargeStation(ctx context.Context, chargeStationId string) (*store.ChargeStation, error) {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStation/%s", chargeStationId))
	snap, err := csRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup charge station %s: %w", chargeStationId, err)
	}
	var cs store.ChargeStation
	if err = snap.DataTo(&cs); err != nil {
		return nil, fmt.Errorf("map charge station %s: %w", chargeStationId, err)
	}
  return &cs, nil
}

func (s *Store) ListChargeStations(context context.Context, offset, limit int) ([]*store.ChargeStation, error) {
	var chargeStations []*store.ChargeStation
	iter := s.client.Collection("ChargeStation").OrderBy("Id", firestore.Asc).Offset(offset).Limit(limit).Documents(context)
	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("next chargeStation: %w", err)
		}
	  var cs store.ChargeStation
		if err = snap.DataTo(&cs); err != nil {
			return nil, fmt.Errorf("map chargeStation: %w", err)
		}
		chargeStations = append(chargeStations, &cs)
	}
	if chargeStations == nil {
		chargeStations = make([]*store.ChargeStation, 0)
	}
	return chargeStations, nil
}

type chargeStationSetting struct {
	Value     string    `firestore:"v"`
	Status    string    `firestore:"s"`
	SendAfter time.Time `firestore:"u"`
}

func (s *Store) UpdateChargeStationSettings(ctx context.Context, chargeStationId string, settings *store.ChargeStationSettings) error {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationSettings/%s", chargeStationId))
	var set = make(map[string]*chargeStationSetting)
	for k, v := range settings.Settings {
		set[k] = &chargeStationSetting{
			Value:  v.Value,
			Status: string(v.Status),
		}
	}
	_, err := csRef.Set(ctx, set, firestore.MergeAll)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupChargeStationSettings(ctx context.Context, chargeStationId string) (*store.ChargeStationSettings, error) {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationSettings/%s", chargeStationId))
	snap, err := csRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup charge station settings %s: %w", chargeStationId, err)
	}
	var csData map[string]*chargeStationSetting
	if err = snap.DataTo(&csData); err != nil {
		return nil, fmt.Errorf("map charge station settings %s: %w", chargeStationId, err)
	}
	var settings = mapChargeStationSettings(csData)
	return &store.ChargeStationSettings{
		ChargeStationId: chargeStationId,
		Settings:        settings,
	}, nil
}

func (s *Store) DeleteChargeStationSettings(ctx context.Context, chargeStationId string) error {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationSettings/%s", chargeStationId))
	_, err := csRef.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func mapChargeStationSettings(csData map[string]*chargeStationSetting) map[string]*store.ChargeStationSetting {
	var settings = make(map[string]*store.ChargeStationSetting)
	for k, v := range csData {
		settings[k] = &store.ChargeStationSetting{
			Value:     v.Value,
			Status:    store.ChargeStationSettingStatus(v.Status),
			SendAfter: v.SendAfter,
		}
	}
	return settings
}

func (s *Store) ListChargeStationSettings(ctx context.Context, pageSize int, previousCsId string) ([]*store.ChargeStationSettings, error) {
	var chargeStationSettings []*store.ChargeStationSettings
	var docIt *firestore.DocumentIterator
	if previousCsId == "" {
		docIt = s.client.Collection("ChargeStationSettings").OrderBy(firestore.DocumentID, firestore.Asc).
			Limit(pageSize).Documents(ctx)
	} else {
		docIt = s.client.Collection("ChargeStationSettings").OrderBy(firestore.DocumentID, firestore.Asc).
			StartAfter(previousCsId).Limit(pageSize).Documents(ctx)
	}
	snaps, err := docIt.GetAll()
	if err != nil {
		return nil, fmt.Errorf("list charge station settings: %w", err)
	}
	for _, snap := range snaps {
		var settings map[string]*chargeStationSetting
		if err = snap.DataTo(&settings); err != nil {
			return nil, fmt.Errorf("map charge station settings: %w", err)
		}
		chargeStationSetting := mapChargeStationSettings(settings)
		chargeStationSettings = append(chargeStationSettings, &store.ChargeStationSettings{
			ChargeStationId: snap.Ref.ID,
			Settings:        chargeStationSetting,
		})
	}
	return chargeStationSettings, nil
}

type chargeStationInstallCertificate struct {
	Type      string    `firestore:"t"`
	Data      string    `firestore:"d"`
	Status    string    `firestore:"s"`
	SendAfter time.Time `firestore:"u"`
}

func mapChargeStationInstallCertificates(certificates map[string]*chargeStationInstallCertificate) []*store.ChargeStationInstallCertificate {
	var certs []*store.ChargeStationInstallCertificate
	for id, c := range certificates {
		certs = append(certs, &store.ChargeStationInstallCertificate{
			CertificateType:               store.CertificateType(c.Type),
			CertificateId:                 id,
			CertificateData:               c.Data,
			CertificateInstallationStatus: store.CertificateInstallationStatus(c.Status),
			SendAfter:                     c.SendAfter,
		})
	}
	return certs
}

func (s *Store) UpdateChargeStationInstallCertificates(ctx context.Context, chargeStationId string, certificates *store.ChargeStationInstallCertificates) error {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationInstallCertificates/%s", chargeStationId))
	var set = make(map[string]*chargeStationInstallCertificate)
	for _, c := range certificates.Certificates {
		set[c.CertificateId] = &chargeStationInstallCertificate{
			Type:      string(c.CertificateType),
			Data:      c.CertificateData,
			Status:    string(c.CertificateInstallationStatus),
			SendAfter: c.SendAfter,
		}
	}
	_, err := csRef.Set(ctx, set, firestore.MergeAll)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupChargeStationInstallCertificates(ctx context.Context, chargeStationId string) (*store.ChargeStationInstallCertificates, error) {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationInstallCertificates/%s", chargeStationId))
	snap, err := csRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup charge station install certificates %s: %w", chargeStationId, err)
	}
	var csData map[string]*chargeStationInstallCertificate
	if err = snap.DataTo(&csData); err != nil {
		return nil, fmt.Errorf("map charge station install certificates %s: %w", chargeStationId, err)
	}
	var certs = mapChargeStationInstallCertificates(csData)
	return &store.ChargeStationInstallCertificates{
		ChargeStationId: chargeStationId,
		Certificates:    certs,
	}, nil
}

func (s *Store) ListChargeStationInstallCertificates(ctx context.Context, pageSize int, previousCsId string) ([]*store.ChargeStationInstallCertificates, error) {
	var installCerts []*store.ChargeStationInstallCertificates
	var docIt *firestore.DocumentIterator
	if previousCsId == "" {
		docIt = s.client.Collection("ChargeStationInstallCertificates").OrderBy(firestore.DocumentID, firestore.Asc).
			Limit(pageSize).Documents(ctx)
	} else {
		docIt = s.client.Collection("ChargeStationInstallCertificates").OrderBy(firestore.DocumentID, firestore.Asc).
			StartAfter(previousCsId).Limit(pageSize).Documents(ctx)
	}
	snaps, err := docIt.GetAll()
	if err != nil {
		return nil, fmt.Errorf("list charge station install certificates: %w", err)
	}
	for _, snap := range snaps {
		var certs map[string]*chargeStationInstallCertificate
		if err = snap.DataTo(&certs); err != nil {
			return nil, fmt.Errorf("map charge station install certificates: %w", err)
		}
		installCert := mapChargeStationInstallCertificates(certs)
		installCerts = append(installCerts, &store.ChargeStationInstallCertificates{
			ChargeStationId: snap.Ref.ID,
			Certificates:    installCert,
		})
	}
	return installCerts, nil
}

type chargeStationRuntimeDetails struct {
	OcppVersion string `firestore:"v"`
}

func (s *Store) SetChargeStationRuntimeDetails(ctx context.Context, chargeStationId string, details *store.ChargeStationRuntimeDetails) error {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationRuntimeDetails/%s", chargeStationId))
	_, err := csRef.Set(ctx, &chargeStationRuntimeDetails{
		OcppVersion: details.OcppVersion,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupChargeStationRuntimeDetails(ctx context.Context, chargeStationId string) (*store.ChargeStationRuntimeDetails, error) {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationRuntimeDetails/%s", chargeStationId))
	snap, err := csRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup charge station runtime details %s: %w", chargeStationId, err)
	}
	var csData chargeStationRuntimeDetails
	if err = snap.DataTo(&csData); err != nil {
		return nil, fmt.Errorf("map charge station runtime details %s: %w", chargeStationId, err)
	}
	return &store.ChargeStationRuntimeDetails{
		OcppVersion: csData.OcppVersion,
	}, nil
}

type chargeStationTriggerMessage struct {
	Type      string    `firestore:"t"`
	Status    string    `firestore:"s"`
	SendAfter time.Time `firestore:"u"`
}

func (s *Store) SetChargeStationTriggerMessage(ctx context.Context, chargeStationId string, triggerMessage *store.ChargeStationTriggerMessage) error {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationTriggerMessage/%s", chargeStationId))
	_, err := csRef.Set(ctx, &chargeStationTriggerMessage{
		Type:      string(triggerMessage.TriggerMessage),
		Status:    string(triggerMessage.TriggerStatus),
		SendAfter: triggerMessage.SendAfter,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteChargeStationTriggerMessage(ctx context.Context, chargeStationId string) error {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationTriggerMessage/%s", chargeStationId))
	_, err := csRef.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupChargeStationTriggerMessage(ctx context.Context, chargeStationId string) (*store.ChargeStationTriggerMessage, error) {
	csRef := s.client.Doc(fmt.Sprintf("ChargeStationTriggerMessage/%s", chargeStationId))
	snap, err := csRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup charge station trigger message %s: %w", chargeStationId, err)
	}
	var csData chargeStationTriggerMessage
	if err = snap.DataTo(&csData); err != nil {
		return nil, fmt.Errorf("map charge station trigger message %s: %w", chargeStationId, err)
	}
	return &store.ChargeStationTriggerMessage{
		ChargeStationId: chargeStationId,
		TriggerMessage:  store.TriggerMessage(csData.Type),
		TriggerStatus:   store.TriggerStatus(csData.Status),
		SendAfter:       csData.SendAfter,
	}, nil
}

func (s *Store) ListChargeStationTriggerMessages(ctx context.Context, pageSize int, previousCsId string) ([]*store.ChargeStationTriggerMessage, error) {
	var triggerMessages []*store.ChargeStationTriggerMessage
	var docIt *firestore.DocumentIterator
	if previousCsId == "" {
		docIt = s.client.Collection("ChargeStationTriggerMessage").OrderBy(firestore.DocumentID, firestore.Asc).
			Limit(pageSize).Documents(ctx)
	} else {
		docIt = s.client.Collection("ChargeStationTriggerMessage").OrderBy(firestore.DocumentID, firestore.Asc).
			StartAfter(previousCsId).Limit(pageSize).Documents(ctx)
	}
	snaps, err := docIt.GetAll()
	if err != nil {
		return nil, fmt.Errorf("list charge station trigger messages: %w", err)
	}
	for _, snap := range snaps {
		var triggerMessage chargeStationTriggerMessage
		if err = snap.DataTo(&triggerMessage); err != nil {
			return nil, fmt.Errorf("map charge station trigger message: %w", err)
		}
		triggerMessages = append(triggerMessages, &store.ChargeStationTriggerMessage{
			ChargeStationId: snap.Ref.ID,
			TriggerMessage:  store.TriggerMessage(triggerMessage.Type),
			TriggerStatus:   store.TriggerStatus(triggerMessage.Status),
			SendAfter:       triggerMessage.SendAfter,
		})
	}
	return triggerMessages, nil
}
