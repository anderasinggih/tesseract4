package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
)

// AgentConnection represents a live physical client WebSocket session
type AgentConnection struct {
	ID      string
	Alias   string
	Faction string
	City    string
	Coord   GeoCoord
	Conn    *websocket.Conn
	Send    chan []byte
}

// WritePump pushes state updates to the browser
func (a *AgentConnection) WritePump() {
	defer a.Conn.Close()

	for {
		message, ok := <-a.Send
		if !ok {
			_ = a.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := a.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

// ReadPump handles commands from the cyber operator UI
func (a *AgentConnection) ReadPump(c *websocket.Conn) {
	for {
		var cmd ClientCommand
		err := c.ReadJSON(&cmd)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ReadPump] Operator %s disconnected: %v", a.ID, err)
			}
			break
		}

		switch cmd.Type {
		case "join":
			if cmd.Alias != "" {
				a.Alias = cmd.Alias
			}
			if cmd.Faction != "" {
				a.Faction = cmd.Faction
			}
			if cmd.City != "" {
				a.City = cmd.City
			}
			GlobalWarRoom.UpdateAgent(a.ID, a.Alias, a.Faction, a.City)

		case "attack":
			if cmd.TargetCity != "" && cmd.Vector != "" {
				payloadMB := cmd.PayloadMB
				if payloadMB <= 0 {
					payloadMB = 250
				}
				GlobalWarRoom.LaunchAttack(a.ID, a.City, cmd.TargetCity, cmd.Vector, payloadMB)
			}

		case "defend":
			GlobalWarRoom.DeployDefense(a.ID, cmd.ActionType, cmd.ArcID, cmd.TargetCity)
		}
	}
}

// TelemetryLoop streams the global war room state to the client at 30Hz
func (a *AgentConnection) TelemetryLoop(ctx context.Context) {
	ticker := time.NewTicker(33 * time.Millisecond) // 30Hz
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			state := GlobalWarRoom.GetSnapshot(a.ID)
			data, err := json.Marshal(state)
			if err == nil {
				select {
				case a.Send <- data:
				default:
				}
			}
		}
	}
}
