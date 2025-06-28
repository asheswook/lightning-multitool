package lndrest

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateInvoice(t *testing.T) {
	t.Run("successful invoice creation", func(t *testing.T) {
		const macaroon = "test-macaroon"
		const memo = "test memo"
		const value = 1000
		const paymentRequest = "lnbc10u1pjx5g8zpp5z..."
		var rHash = []byte{1, 2, 3}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/invoices", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, macaroon, r.Header.Get("Grpc-Metadata-macaroon"))

			var reqBody CreateInvoiceParams
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, memo, reqBody.Memo)
			assert.Equal(t, int64(value), reqBody.Value)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CreateInvoiceResponse{
				PaymentRequest: paymentRequest,
				RHash:          rHash,
			})
		}))
		defer server.Close()

		client, err := NewClient(server.URL, macaroon, "")
		require.NoError(t, err)

		params := CreateInvoiceParams{
			Memo:  memo,
			Value: value,
		}
		resp, err := client.CreateInvoice(context.Background(), params)

		require.NoError(t, err)
		assert.Equal(t, paymentRequest, resp.PaymentRequest)
		assert.Equal(t, rHash, resp.RHash)
	})

	t.Run("LND API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		}))
		defer server.Close()

		client, err := NewClient(server.URL, "macaroon", "")
		require.NoError(t, err)

		_, err = client.CreateInvoice(context.Background(), CreateInvoiceParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "LND API error: 500 Internal Server Error")
		assert.Contains(t, err.Error(), "internal error")
	})
}
