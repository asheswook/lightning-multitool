package lndrest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	// Helper function to create a temporary self-signed certificate
	createTempCert := func(t *testing.T) string {
		t.Helper()
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		template := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject:      pkix.Name{CommonName: "localhost"},
			NotBefore:    time.Now(),
			NotAfter:     time.Now().Add(time.Hour),
			KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
		require.NoError(t, err)

		tempDir := t.TempDir()
		certPath := filepath.Join(tempDir, "cert.pem")
		certOut, err := os.Create(certPath)
		require.NoError(t, err)
		defer certOut.Close()

		pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		return certPath
	}

	t.Run("with valid cert", func(t *testing.T) {
		certPath := createTempCert(t)
		client, err := NewClient("https://localhost:8080", "macaroon", certPath)
		require.NoError(t, err)
		require.NotNil(t, client)
		transport := client.httpClient.Transport.(*http.Transport)
		assert.NotNil(t, transport.TLSClientConfig.RootCAs)
		assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("with no cert (insecure)", func(t *testing.T) {
		client, err := NewClient("https://localhost:8080", "macaroon", "")
		require.NoError(t, err)
		require.NotNil(t, client)
		transport := client.httpClient.Transport.(*http.Transport)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("with non-existent cert path", func(t *testing.T) {
		_, err := NewClient("https://localhost:8080", "macaroon", "/path/to/non/existent/cert.pem")
		require.Error(t, err)
	})

	t.Run("with invalid cert content", func(t *testing.T) {
		tempDir := t.TempDir()
		certPath := filepath.Join(tempDir, "invalid.pem")
		require.NoError(t, os.WriteFile(certPath, []byte("invalid content"), 0644))
		_, err := NewClient("https://localhost:8080", "macaroon", certPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to append LND cert to pool")
	})
}
