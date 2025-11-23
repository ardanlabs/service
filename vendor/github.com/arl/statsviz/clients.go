package statsviz

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/arl/statsviz/internal/plot"
)

type clients struct {
	cfg *plot.Config
	ctx context.Context

	mu sync.RWMutex
	m  map[*websocket.Conn]chan []byte
}

func newClients(ctx context.Context, cfg *plot.Config) *clients {
	return &clients{
		m:   make(map[*websocket.Conn]chan []byte),
		cfg: cfg,
		ctx: ctx,
	}
}

type wsmsg struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func (c *clients) add(conn *websocket.Conn) {
	dbglog("adding client")

	// Send config first.
	err := conn.WriteJSON(wsmsg{Event: "config", Data: c.cfg})
	if err != nil {
		dbglog("failed to send config: %v", err)
		return
	}

	ch := make(chan []byte)

	go func() {
		defer func() {
			c.mu.Lock()
			delete(c.m, conn)
			c.mu.Unlock()

			dbglog("removed client")
		}()

		for {
			select {
			case <-c.ctx.Done():
				return
			case msg := <-ch:
				if err := sendbuf(conn, msg); err != nil {
					dbglog("failed to send data: %v", err)
					return
				}
			}
		}
	}()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[conn] = ch
}

func sendbuf(conn *websocket.Conn, buf []byte) error {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	_, err1 := w.Write(buf)
	err2 := w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func (c *clients) broadcast(buf []byte) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, ch := range c.m {
		select {
		case ch <- buf:
		default:
			// if a client is not keeping up, we
			// drop the message for that client.
			dbglog("dropping message to client")
		}
	}
}
