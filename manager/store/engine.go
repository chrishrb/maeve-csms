// SPDX-License-Identifier: Apache-2.0

package store

type Engine interface {
	Store
	ChargeStationStore
	ChargeStationSettingsStore
	ChargeStationRuntimeDetailsStore
	ChargeStationInstallCertificatesStore
	ChargeStationTriggerMessageStore
	TokenStore
	TransactionStore
	CertificateStore
	OcpiStore
	LocationStore
}
