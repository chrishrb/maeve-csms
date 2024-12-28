package api_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCertificate(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	cert := generateCertificate(t)
	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	encodedPemCert := strings.Replace(string(pemCert), "\n", "\\n", -1)

	req := httptest.NewRequest(http.MethodPost, "/certificate", strings.NewReader(fmt.Sprintf(`{"certificate":"%s"}`, encodedPemCert)))
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	b, err := io.ReadAll(rr.Result().Body)
	require.NoError(t, err)
	assert.Equal(t, "", string(b))

	b64Hash := getCertificateHash(cert)

	got, err := engine.LookupCertificate(context.Background(), b64Hash)
	require.NoError(t, err)

	assert.Equal(t, string(pemCert), got)
}

func TestDeleteCertificate(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	cert := generateCertificate(t)
	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	err := engine.SetCertificate(context.Background(), string(pemCert))
	require.NoError(t, err)

	b64Hash := getCertificateHash(cert)
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/certificate/%s", b64Hash), nil)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Result().StatusCode)
	b, err := io.ReadAll(rr.Result().Body)
	require.NoError(t, err)
	assert.Equal(t, "", string(b))

	got, err := engine.LookupCertificate(context.Background(), b64Hash)
	require.NoError(t, err)

	assert.Equal(t, "", got)
}

func TestLookupCertificate(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	cert := generateCertificate(t)
	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	err := engine.SetCertificate(context.Background(), string(pemCert))
	require.NoError(t, err)

	b64Hash := getCertificateHash(cert)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/certificate/%s", b64Hash), nil)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	b, err := io.ReadAll(rr.Result().Body)
	require.NoError(t, err)
	assert.JSONEq(t, fmt.Sprintf(`{"certificate":"%s"}`, strings.Replace(string(pemCert), "\n", "\\n", -1)), string(b))
}

func generateCertificate(t *testing.T) *x509.Certificate {
	keyPair, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Thoughtworks"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &keyPair.PublicKey, keyPair)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	return cert
}

func getCertificateHash(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	b64Hash := base64.RawURLEncoding.EncodeToString(hash[:])
	return b64Hash
}
