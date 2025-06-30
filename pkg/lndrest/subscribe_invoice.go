package lndrest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// SubscribeInvoicesParams defines the request for the SubscribeInvoices call.
type SubscribeInvoicesParams struct {
	AddIndex    uint64 `json:"add_index,omitempty"`
	SettleIndex uint64 `json:"settle_index,omitempty"`
}

type SubscribeInvoicesResponse struct {
	Result *Invoice `json:"result,omitempty"`
	Error  *struct {
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

// SubscribeInvoices subscribes to invoices from the LND node.
// It returns a channel for invoices, a channel for errors, and an error for setup issues.
func (c *Client) SubscribeInvoices(ctx context.Context, req SubscribeInvoicesParams) (<-chan Invoice, error) {
	q := url.Values{}
	if req.AddIndex > 0 {
		q.Set("add_index", fmt.Sprintf("%d", req.AddIndex))
	}
	if req.SettleIndex > 0 {
		q.Set("settle_index", fmt.Sprintf("%d", req.SettleIndex))
	}

	path := "/v1/invoices/subscribe"
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}

	wsurl, err := url.JoinPath("wss://", c.host, path)
	if err != nil {
		return nil, fmt.Errorf("failed to join path: %w", err)
	}

	dialer := websocket.DefaultDialer
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok && transport.TLSClientConfig != nil {
		dialer = &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 30 * time.Second,
			TLSClientConfig:  transport.TLSClientConfig.Clone(),
		}
	}

	header := http.Header{}
	header.Set("Grpc-Metadata-macaroon", c.macaroon)

	conn, resp, err := dialer.DialContext(ctx, wsurl, header)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil, fmt.Errorf("failed to dial websocket: %w", err)
	}

	invoiceChan := make(chan Invoice)
	go func() {
		defer close(invoiceChan)
		defer conn.Close()

		go func() {
			<-ctx.Done()
			conn.Close()
		}()

		for {
			// Wait for the next message (blocked until a new message is received)
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					// 정상 종료 또는 컨텍스트 취소로 인한 종료
					return
				}
				slog.Error("error reading websocket message", "error", err)
				return
			}

			var streamResp SubscribeInvoicesResponse
			if err := json.Unmarshal(message, &streamResp); err != nil {
				slog.Error("error unmarshalling invoice stream response", "error", err)
				continue
			}

			if streamResp.Error != nil {
				slog.Error("lnd stream error", "error", streamResp.Error.Message)
				continue
			}

			select {
			case invoiceChan <- *streamResp.Result:
			case <-ctx.Done():
				return
			}
		}
	}()

	return invoiceChan, nil
}
