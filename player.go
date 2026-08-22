package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
)

// ATCClientConnection represents an active air traffic controller session
type ATCClientConnection struct {
	ID       string
	Callsign string
	SectorID string
	Conn     *websocket.Conn
	Send     chan []byte
}

func (c *ATCClientConnection) WritePump() {
	defer c.Conn.Close()

	for {
		message, ok := <-c.Send
		if !ok {
			_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

func (c *ATCClientConnection) ReadPump(conn *websocket.Conn) {
	for {
		var cmd ClientATCCommand
		err := conn.ReadJSON(&cmd)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ReadPump] Controller %s disconnected: %v", c.ID, err)
			}
			break
		}

		GlobalATCHub.ExecuteATCCommand(cmd, c.ID)
	}
}

func (c *ATCClientConnection) TelemetryLoop(ctx context.Context) {
	ticker := time.NewTicker(33 * time.Millisecond) // 30 FPS high-efficiency real-time telemetry stream
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			payload := GlobalATCHub.GetSnapshot(c.ID)
			data, err := json.Marshal(payload)
			if err == nil {
				select {
				case c.Send <- data:
				default:
				}
			}
		}
	}
}
