// SPDX-License-Identifier: Apache-2.0

package api

import "net/http"

func (c ChargeStation) Bind(r *http.Request) error {
	return nil
}

func (c ChargeStation) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (c ChargeStationSettings) Bind(r *http.Request) error {
	return nil
}

func (c ChargeStationInstallCertificates) Bind(r *http.Request) error {
	return nil
}

func (c ChargeStationTrigger) Bind(r *http.Request) error {
	return nil
}

func (t Token) Bind(r *http.Request) error {
	return nil
}

func (t Token) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (t Certificate) Bind(r *http.Request) error {
	return nil
}

func (t Certificate) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (r Registration) Bind(req *http.Request) error {
	return nil
}

func (r Location) Bind(req *http.Request) error {
	return nil
}

func (t Location) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (t Transaction) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
