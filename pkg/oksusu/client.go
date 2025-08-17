package oksusu

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Handler interface {
	OnLNURLPRequest(ctx context.Context, payload *LNURLRequestPayload) (*LNURLResponsePayload, error)
	OnInvoiceRequest(ctx context.Context, payload *InvoiceRequestPayload) (*InvoiceResponsePayload, error)
}

type Client struct {
	conn    *Conn
	handler Handler
	token   string
	host    string
}

// NewClient creates a new Oksusu Connect client.
func NewClient(host, token string, handler Handler) *Client {
	return &Client{
		host:    host,
		token:   token,
		handler: handler,
	}
}

// ConnectAndServe is a blocking function that connects to the Oksusu server and handles incoming messages.
// It takes a context as input and returns an error if the connection fails.
func (c *Client) ConnectAndServe(ctx context.Context) error {
	u := url.URL{Scheme: "wss", Host: c.host, Path: "/connect"}
	slog.Info("Connecting to Oksu server", "url", u.String())

	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	ws, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.conn = NewConn(ws)
	defer c.conn.Close()

	// 1. authentication
	authCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err = c.authenticate(authCtx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	slog.Info("Successfully authenticated with Oksu server")

	// 2. listening loop
	return c.listen(ctx)
}

func (c *Client) authenticate(ctx context.Context) error {
	authPayload := AuthRequestPayload{Token: c.token}
	payloadBytes, _ := json.Marshal(authPayload)

	req := Message{
		ID:      fmt.Sprintf("auth-%d", time.Now().UnixNano()),
		Type:    C2SAuth,
		Payload: payloadBytes,
	}

	if err := c.conn.WriteMessage(ctx, &req); err != nil {
		return err
	}

	// wait for auth response
	resp, err := c.conn.ReadMessage(ctx)
	if err != nil {
		return err
	}

	var p AuthResponsePayload
	_ = json.Unmarshal(resp.Payload, &p)

	if resp.Type == S2CAuthOK {
		return nil
	}

	if resp.Type == S2CAuthFail {
		return fmt.Errorf("auth failed: %s", p.Message)
	}

	return fmt.Errorf("unexpected response type during auth: %s", resp.Type)
}

func (c *Client) listen(ctx context.Context) error {
	for {
		msg, err := c.conn.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Info("Oksu client disconnected gracefully")
				return nil
			}
			slog.Error("Failed to read message, disconnecting", "error", err)
			return err
		}

		go c.handleMessage(context.Background(), msg)
	}
}

func (c *Client) handleMessage(ctx context.Context, msg *Message) {
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var err error
	var responsePayload interface{}

	switch msg.Type {
	case S2CLNURLPRequest:
		var p LNURLRequestPayload
		if err = json.Unmarshal(msg.Payload, &p); err == nil {
			responsePayload, err = c.handler.OnLNURLPRequest(reqCtx, &p)
		}
	case S2CInvoiceRequest:
		var p InvoiceRequestPayload
		if err = json.Unmarshal(msg.Payload, &p); err == nil {
			responsePayload, err = c.handler.OnInvoiceRequest(reqCtx, &p)
		}
	default:
		slog.Warn("Received unknown message type, ignoring", "type", msg.Type, "request_id", msg.ID)
		return
	}

	respMsg := Message{ID: msg.ID} // Response ID is same as the request ID

	if err != nil {
		slog.Error("Error handling request", "type", msg.Type, "request_id", msg.ID, "error", err)
		respMsg.Type = C2SError
		payload, _ := json.Marshal(ErrorPayload{Message: err.Error()})
		respMsg.Payload = payload
	} else {
		switch msg.Type {
		case S2CLNURLPRequest:
			respMsg.Type = C2SLNURLPResponse
		case S2CInvoiceRequest:
			respMsg.Type = C2SInvoiceResponse
		}
		payload, _ := json.Marshal(responsePayload)
		respMsg.Payload = payload
	}

	writeCtx, writeCancel := context.WithTimeout(ctx, 10*time.Second)
	defer writeCancel()
	if err := c.conn.WriteMessage(writeCtx, &respMsg); err != nil {
		slog.Error("Failed to send response", "request_id", msg.ID, "error", err)
	}
}
