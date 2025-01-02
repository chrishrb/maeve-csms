// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"time"
)

type SecurityProfile int8

const (
	UnsecuredTransportWithBasicAuth SecurityProfile = iota
	TLSWithBasicAuth
	TLSWithClientSideCertificates
)

type Connector struct {
	Id          string `json:"id"`
	Format      string `json:"format"`
	MaxAmperage int32  `json:"max_amperage"`
	MaxVoltage  int32  `json:"max_voltage"`
	PowerType   string `json:"power_type"`
	Standard    string `json:"standard"`
	LastUpdated string `json:"last_updated"`
}

type Evse struct {
	Connectors  []Connector `json:"connectors"`
	EvseId      *string     `json:"evse_id"`
	Status      string      `json:"status"`
	Uid         string      `json:"uid"`
	LastUpdated string      `json:"last_updated"`
}

type ChargeStation struct {
	Id                     string          `json:"id"`
	LocationId             string          `json:"location_id"`
	Evses                  *[]Evse         `json:"evses"`
	SecurityProfile        SecurityProfile `json:"security_profile"`
	Base64SHA256Password   string          `json:"base64_sha256_password"`
	InvalidUsernameAllowed bool            `json:"invalid_username_allowed"`
}

type ChargeStationStore interface {
	CreateChargeStation(ctx context.Context, cs *ChargeStation) (*ChargeStation, error)
	UpdateChargeStation(ctx context.Context, csId string, cs *ChargeStation) (*ChargeStation, error)
	DeleteChargeStation(ctx context.Context, csId string) error
	LookupChargeStation(ctx context.Context, csId string) (*ChargeStation, error)
	ListChargeStations(context context.Context, offset int, limit int) ([]*ChargeStation, error)
}

type ChargeStationSettingStatus string

var (
	ChargeStationSettingStatusPending        ChargeStationSettingStatus = "Pending"
	ChargeStationSettingStatusAccepted       ChargeStationSettingStatus = "Accepted"
	ChargeStationSettingStatusRejected       ChargeStationSettingStatus = "Rejected"
	ChargeStationSettingStatusRebootRequired ChargeStationSettingStatus = "RebootRequired"
	ChargeStationSettingStatusNotSupported   ChargeStationSettingStatus = "NotSupported"
)

type ChargeStationSetting struct {
	Value     string
	Status    ChargeStationSettingStatus
	SendAfter time.Time
}

type ChargeStationSettings struct {
	ChargeStationId string
	Settings        map[string]*ChargeStationSetting
}

type ChargeStationSettingsStore interface {
	UpdateChargeStationSettings(ctx context.Context, csId string, settings *ChargeStationSettings) error
	LookupChargeStationSettings(ctx context.Context, csId string) (*ChargeStationSettings, error)
	ListChargeStationSettings(ctx context.Context, pageSize int, previousChargeStationId string) ([]*ChargeStationSettings, error)
	DeleteChargeStationSettings(ctx context.Context, csId string) error
}

type ChargeStationRuntimeDetails struct {
	OcppVersion string `json:"ocpp_version"`
}

type ChargeStationRuntimeDetailsStore interface {
	SetChargeStationRuntimeDetails(ctx context.Context, csId string, details *ChargeStationRuntimeDetails) error
	LookupChargeStationRuntimeDetails(ctx context.Context, csId string) (*ChargeStationRuntimeDetails, error)
}

type CertificateType string

var (
	CertificateTypeChargeStation CertificateType = "ChargeStation"
	CertificateTypeEVCC          CertificateType = "EVCC"
	CertificateTypeV2G           CertificateType = "V2G"
	CertificateTypeMO            CertificateType = "MO"
	CertificateTypeMF            CertificateType = "MF"
	CertificateTypeCSMS          CertificateType = "CSMS"
)

type CertificateInstallationStatus string

var (
	CertificateInstallationPending  CertificateInstallationStatus = "Pending"
	CertificateInstallationAccepted CertificateInstallationStatus = "Accepted"
	CertificateInstallationRejected CertificateInstallationStatus = "Rejected"
)

type ChargeStationInstallCertificate struct {
	CertificateType               CertificateType
	CertificateId                 string
	CertificateData               string
	CertificateInstallationStatus CertificateInstallationStatus
	SendAfter                     time.Time
}

type ChargeStationInstallCertificates struct {
	ChargeStationId string
	Certificates    []*ChargeStationInstallCertificate
}

type ChargeStationInstallCertificatesStore interface {
	UpdateChargeStationInstallCertificates(ctx context.Context, csId string, certificates *ChargeStationInstallCertificates) error
	LookupChargeStationInstallCertificates(ctx context.Context, csId string) (*ChargeStationInstallCertificates, error)
	ListChargeStationInstallCertificates(ctx context.Context, pageSize int, previousChargeStationId string) ([]*ChargeStationInstallCertificates, error)
}

type TriggerStatus string

var (
	TriggerStatusPending        TriggerStatus = "Pending"
	TriggerStatusAccepted       TriggerStatus = "Accepted"
	TriggerStatusRejected       TriggerStatus = "Rejected"
	TriggerStatusNotImplemented TriggerStatus = "NotImplemented"
)

type TriggerMessage string

var (
	TriggerMessageBootNotification                  TriggerMessage = "BootNotification"
	TriggerMessageHeartbeat                         TriggerMessage = "Heartbeat"
	TriggerMessageStatusNotification                TriggerMessage = "StatusNotification"
	TriggerMessageFirmwareStatusNotification        TriggerMessage = "FirmwareStatusNotification"
	TriggerMessageDiagnosticStatusNotification      TriggerMessage = "DiagnosticStatusNotification"
	TriggerMessageMeterValues                       TriggerMessage = "MeterValues"
	TriggerMessageSignChargingStationCertificate    TriggerMessage = "SignChargingStationCertificate"
	TriggerMessageSignV2GCertificate                TriggerMessage = "SignV2GCertificate"
	TriggerMessageSignCombinedCertificate           TriggerMessage = "SignCombinedCertificate"
	TriggerMessagePublishFirmwareStatusNotification TriggerMessage = "PublishFirmwareStatusNotification"
)

type ChargeStationTriggerMessage struct {
	ChargeStationId string
	TriggerMessage  TriggerMessage
	TriggerStatus   TriggerStatus
	SendAfter       time.Time
}

type ChargeStationTriggerMessageStore interface {
	SetChargeStationTriggerMessage(ctx context.Context, csId string, triggerMessage *ChargeStationTriggerMessage) error
	DeleteChargeStationTriggerMessage(ctx context.Context, csId string) error
	LookupChargeStationTriggerMessage(ctx context.Context, csId string) (*ChargeStationTriggerMessage, error)
	ListChargeStationTriggerMessages(ctx context.Context, pageSize int, previousCsId string) ([]*ChargeStationTriggerMessage, error)
}
