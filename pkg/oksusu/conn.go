package oksusu

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	ws   *websocket.Conn
	wMtx sync.Mutex
}

func NewConn(ws *websocket.Conn) *Conn {
	return &Conn{ws: ws}
}

func (c *Conn) ReadMessage(ctx context.Context) (*Message, error) {
	msgChan := make(chan *Message)
	errChan := make(chan error)

	go func() {
		defer close(msgChan)
		defer close(errChan)

		msgType, p, err := c.ws.ReadMessage()
		if err != nil {
			errChan <- err
			return
		}

		if msgType != websocket.TextMessage {
			return
		}

		var msg Message
		if err := json.Unmarshal(p, &msg); err != nil {
			errChan <- fmt.Errorf("failed to unmarshal message: %w", err)
			return
		}
		msgChan <- &msg
	}()

	select {
	case <-ctx.Done():
		// 컨텍스트가 완료되면 WebSocket 연결을 닫아 ReadMessage 고루틴을 종료시킵니다.
		c.ws.Close()
		return nil, ctx.Err()
	case msg := <-msgChan:
		return msg, nil
	case err := <-errChan:
		return nil, err
	}
}

func (c *Conn) WriteMessage(ctx context.Context, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	c.wMtx.Lock()
	defer c.wMtx.Unlock()

	if deadline, ok := ctx.Deadline(); ok {
		c.ws.SetWriteDeadline(deadline)
	} else {
		c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	}

	defer c.ws.SetWriteDeadline(time.Time{})

	return c.ws.WriteMessage(websocket.TextMessage, data)
}

func (c *Conn) Close() error {
	return c.ws.Close()
}
