package lndrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateInvoiceParams holds all possible parameters for creating an LND invoice.
type CreateInvoiceParams struct {
	Memo            string `json:"memo,omitempty"`
	RPreimage       []byte `json:"r_preimage,omitempty"`
	RHash           []byte `json:"r_hash,omitempty"`
	Value           int64  `json:"value,omitempty"`
	ValueMsat       int64  `json:"value_msat,omitempty"`
	DescriptionHash []byte `json:"description_hash,omitempty"`
	Expiry          int64  `json:"expiry,omitempty"`
	FallbackAddr    string `json:"fallback_addr,omitempty"`
	CltvExpiry      int64  `json:"cltv_expiry,omitempty"`
	Private         bool   `json:"private,omitempty"`
}

// CreateInvoiceResponse is the response from LND after creating an invoice.
type CreateInvoiceResponse struct {
	RHash          []byte `json:"r_hash"`
	PaymentRequest string `json:"payment_request"`
	AddIndex       uint64 `json:"add_index"`
	PaymentAddr    []byte `json:"payment_addr"`
}

// CreateInvoice creates a new invoice on the LND node.
func (c *Client) CreateInvoice(ctx context.Context, params CreateInvoiceParams) (CreateInvoiceResponse, error) {
	bodyBytes, err := json.Marshal(params)
	if err != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("failed to marshal invoice request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/invoices", c.host)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Grpc-Metadata-macaroon", c.macaroonBase64)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("failed to send request to LND: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return CreateInvoiceResponse{}, fmt.Errorf("LND API error: %s, body: %s", resp.Status, string(body))
	}

	var invResp CreateInvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&invResp); err != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("failed to decode LND response: %w", err)
	}

	return invResp, nil
}
