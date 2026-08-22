package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// Player represents an active player session in the game
type Player struct {
	ID       string
	Position Vector4
	Yaw      float64         // Mouse horizontal angle (Look Left/Right)
	Pitch    float64         // Mouse vertical angle (Look Up/Down)
	HyperRot float64         // 4D XW dimension rotation angle
	Conn     *websocket.Conn // Physical WebSocket connection
	Send     chan []byte     // Outbound message queue (Buffered Channel)
	Mutex    sync.RWMutex    // Protects Position and Rotation mutability
}

// WritePump acts as the dedicated "Kurir" per player.
func (p *Player) WritePump() {
	defer p.Conn.Close()

	for {
		message, ok := <-p.Send
		if !ok {
			_ = p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := p.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

// ReadPump listens for FPS WASD movements and Mouse Look angles
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

		p.Mutex.Lock()
		switch cmd.Type {
		case "look":
			// FPS Mouse Look: DX -> Yaw, DY -> Pitch
			sensitivity := 0.003
			p.Yaw += cmd.DX * sensitivity
			p.Pitch += cmd.DY * sensitivity

			// Clamp pitch so camera doesn't flip upside down
			if p.Pitch > 1.5 {
				p.Pitch = 1.5
			} else if p.Pitch < -1.5 {
				p.Pitch = -1.5
			}

		case "move":
			speed := 0.08
			if cmd.Delta != 0 {
				speed = cmd.Delta
			}

			// FPS Direction Vectors based on camera Yaw
			cosY := math.Cos(p.Yaw)
			sinY := math.Sin(p.Yaw)

			switch strings.ToLower(cmd.Key) {
			// W: Maju ke Depan (arah hadap kamera di sumbu Z/X)
			case "w":
				p.Position.Z += speed * cosY
				p.Position.X += speed * sinY

			// S: Mundur ke Belakang
			case "s":
				p.Position.Z -= speed * cosY
				p.Position.X -= speed * sinY

			// A: Strafe ke Kiri
			case "a":
				p.Position.X -= speed * cosY
				p.Position.Z += speed * sinY

			// D: Strafe ke Kanan
			case "d":
				p.Position.X += speed * cosY
				p.Position.Z -= speed * sinY

			// Panah Atas / Space: Terbang Naik (Sumbu Y)
			case "arrowup", " ":
				p.Position.Y += speed

			// Panah Bawah / C: Turun ke Bawah (Sumbu Y)
			case "arrowdown", "c":
				p.Position.Y -= speed

			// Shift: Mendekati Singularity Black Hole (Sumbu W)
			case "shift":
				p.Position.W -= speed

			// E: Menjauhi Singularity Black Hole (Sumbu W)
			case "e":
				p.Position.W += speed

			// Q / R: Manual 4D Hypercube Spin
			case "q":
				p.HyperRot -= 0.05
			case "r":
				p.HyperRot += 0.05
			}
		}
		p.Mutex.Unlock()
	}
}

// PhysicsLoop computes 4D projection with full FPS camera, streaming smoothly at 60 FPS
func (p *Player) PhysicsLoop(ctx context.Context) {
	// 60 FPS ticker (16ms)
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			p.Mutex.RLock()
			currentPos := p.Position
			yaw := p.Yaw
			pitch := p.Pitch
			hyperRot := p.HyperRot
			p.Mutex.RUnlock()

			// 1. Calculate Schwarzschild Time Dilation Multiplier
			timeMultiplier := CalculateTimeMultiplier(currentPos.W)

			// 2. Compute dynamic tick latency display
			effectiveTickMs := float64(BaseTick) / timeMultiplier

			// 3. Generate 2D projected lines with FPS camera transform
			lines := GenerateProjectedLines(currentPos, yaw, pitch, hyperRot)

			// 4. Construct payload
			payload := FramePayload{
				Lines:          lines,
				PlayerPos:      currentPos,
				TimeMultiplier: timeMultiplier,
				TickMs:         effectiveTickMs,
				PlayerID:       p.ID,
			}

			// 5. Deliver to send queue
			data, err := json.Marshal(payload)
			if err == nil {
				select {
				case p.Send <- data:
				default:
				}
			}

			// Ambient 4D dimensional rotation
			p.Mutex.Lock()
			p.HyperRot += 0.008 * timeMultiplier
			p.Mutex.Unlock()
		}
	}
}
