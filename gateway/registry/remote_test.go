package registry_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoughtworks/maeve-csms/gateway/registry"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLookupChargeStation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"security_profile":1,"base64_SHA256_password":"DEADBEEF"}`))
	}))
	defer server.Close()

	reg := registry.RemoteRegistry{
		ManagerApiAddr: server.URL,
	}

	want := &registry.ChargeStation{
		ClientId:             "cs001",
		SecurityProfile:      1,
		Base64SHA256Password: "DEADBEEF",
	}

	got, _ := reg.LookupChargeStation("cs001")
	require.NotNil(t, got)

	assert.Equal(t, want, got)
}

func TestLookupCertificate(t *testing.T) {
	want := generateCertificate(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		block := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: want.Raw})
		blockWithNewlinesReplaced := strings.Replace(string(block), "\n", "\\n", -1)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"certificate":"%s"}`, blockWithNewlinesReplaced)))
	}))
	defer server.Close()

	reg := registry.RemoteRegistry{
		ManagerApiAddr: server.URL,
	}

	certHash := sha256.Sum256(want.Raw)
	b64CertHash := base64.StdEncoding.EncodeToString(certHash[:])

	got, err := reg.LookupCertificate(b64CertHash)
	require.NoError(t, err)

	assert.Equal(t, want.Raw, got.Raw)
}

func TestLookupCertificateWithSlashesAndPlusesInHash(t *testing.T) {
	var want *x509.Certificate
	var certHash [32]byte
	var b64CertHash string

	count := 0
	for {
		count++
		want = generateCertificate(t)
		certHash = sha256.Sum256(want.Raw)
		b64CertHash = base64.StdEncoding.EncodeToString(certHash[:])

		if strings.Contains(b64CertHash, "/") && strings.Contains(b64CertHash, "+") {
			break
		}
	}
	t.Logf("Generated %d certificates before finding one with slashes and pluses in the hash", count)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashToMatch := strings.Replace(b64CertHash, "/", "_", -1)
		hashToMatch = strings.Replace(hashToMatch, "+", "-", -1)
		if r.URL.Path != fmt.Sprintf("/api/v0/certificate/%s", hashToMatch) {
			http.NotFound(w, r)
			return
		}
		block := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: want.Raw})
		blockWithNewlinesReplaced := strings.Replace(string(block), "\n", "\\n", -1)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"certificate":"%s"}`, blockWithNewlinesReplaced)))
	}))
	defer server.Close()

	reg := registry.RemoteRegistry{
		ManagerApiAddr: server.URL,
	}

	got, err := reg.LookupCertificate(b64CertHash)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, want.Raw, got.Raw)
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
