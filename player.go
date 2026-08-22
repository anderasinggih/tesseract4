package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// Player represents an active player session in the game
type Player struct {
	ID       string
	Position Vector4
	Rotation float64         // Angle theta for XW rotation
	Conn     *websocket.Conn // Physical WebSocket connection
	Send     chan []byte     // Outbound message queue (Buffered Channel)
	Mutex    sync.RWMutex    // Protects Position and Rotation mutability
}

// WritePump acts as the dedicated "Kurir" per player.
// It delivers messages from the in-memory Send channel to the physical network.
func (p *Player) WritePump() {
	defer p.Conn.Close()

	for {
		message, ok := <-p.Send
		if !ok {
			// Hub closed the channel (player kicked / disconnected)
			_ = p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := p.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

// ReadPump listens for raw input commands from the dumb client
func (p *Player) ReadPump(c *websocket.Conn) {
	for {
		var cmd InputCommand
		err := c.ReadJSON(&cmd)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ReadPump] Player %s disconnected with err: %v", p.ID, err)
			}
			break
		}

		speed := 0.05
		if cmd.Delta != 0 {
			speed = cmd.Delta
		}

		p.Mutex.Lock()
		switch strings.ToLower(cmd.Key) {
		// X-axis movement
		case "a":
			p.Position.X -= speed
		case "d":
			p.Position.X += speed

		// Y-axis movement
		case "w":
			p.Position.Y += speed
		case "s":
			p.Position.Y -= speed

		// Z-axis movement
		case "q":
			p.Position.Z -= speed
		case "e":
			p.Position.Z += speed

		// W-axis movement (Proximity to Black Hole)
		case "shift":
			p.Position.W -= speed
		case " ": // Space
			p.Position.W += speed

		// XW 4D Rotation
		case "arrowleft":
			p.Rotation -= 0.04
		case "arrowright":
			p.Rotation += 0.04
		}
		p.Mutex.Unlock()
	}
}

// PhysicsLoop computes 4D projection, time dilation, and enqueues 2D lines to the Send channel
func (p *Player) PhysicsLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			p.Mutex.RLock()
			currentPos := p.Position
			currentRotation := p.Rotation
			p.Mutex.RUnlock()

			// 1. Calculate Schwarzschild Time Dilation Multiplier
			timeMultiplier := CalculateTimeMultiplier(currentPos.W)

			// 2. Dynamic Tick Rate (BaseTick = 16ms / Multiplier)
			actualSleepDuration := time.Duration(float64(BaseTick)/timeMultiplier) * time.Millisecond

			if actualSleepDuration < 1*time.Millisecond {
				actualSleepDuration = 1 * time.Millisecond
			} else if actualSleepDuration > 1000*time.Millisecond {
				actualSleepDuration = 1000 * time.Millisecond
			}

			// 3. Generate 2D projected lines from 4D Tesseract
			lines := GenerateProjectedLines(currentPos, currentRotation)

			// 4. Construct frame payload
			payload := FramePayload{
				Lines:          lines,
				PlayerPos:      currentPos,
				TimeMultiplier: timeMultiplier,
				TickMs:         float64(actualSleepDuration.Milliseconds()),
				PlayerID:       p.ID,
			}

			// 5. Serialize and deliver to Kurir's Send channel
			data, err := json.Marshal(payload)
			if err == nil {
				select {
				case p.Send <- data:
				default:
					// Drop frame if buffer is congested
				}
			}

			// Auto ambient rotation
			p.Mutex.Lock()
			p.Rotation += 0.01
			p.Mutex.Unlock()

			// 6. Non-blocking sleep with context cancellation check
			select {
			case <-ctx.Done():
				return
			case <-time.After(actualSleepDuration):
			}
		}
	}
}
