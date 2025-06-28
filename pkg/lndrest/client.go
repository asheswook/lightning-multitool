package lndrest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client is a client for the LND REST API.
type Client struct {
	httpClient     *http.Client
	host           string
	macaroonBase64 string
}

// NewClient creates a new LND client.
// It configures an HTTP client that trusts the LND's TLS certificate.
func NewClient(host, macaroonBase64, certPath string) (*Client, error) {
	httpClient := &http.Client{Timeout: 20 * time.Second}

	if certPath != "" {
		caCert, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read LND cert file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append LND cert to pool")
		}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
	} else {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		httpClient:     httpClient,
		host:           host,
		macaroonBase64: macaroonBase64,
	}, nil
}
