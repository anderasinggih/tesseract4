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
	Score    int             // Player collected Quantum Core score
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

// ReadPump listens for Minecraft/FPS standard WASD movements and Mouse Look
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
			// FPS Natural Mouse Look (Standard Minecraft / FPS)
			sensitivity := 0.003
			p.Yaw += cmd.DX * sensitivity
			p.Pitch += cmd.DY * sensitivity

			if p.Pitch > 1.45 {
				p.Pitch = 1.45
			} else if p.Pitch < -1.45 {
				p.Pitch = -1.45
			}

		case "move":
			speed := 0.12
			if cmd.Delta != 0 {
				speed = cmd.Delta
			}

			// Standar Minecraft / FPS Movement Vector
			sinY := math.Sin(p.Yaw)
			cosY := math.Cos(p.Yaw)

			switch strings.ToLower(cmd.Key) {
			case "w":
				p.Position.X += speed * sinY
				p.Position.Z += speed * cosY
			case "s":
				p.Position.X -= speed * sinY
				p.Position.Z -= speed * cosY
			case "a":
				p.Position.X -= speed * cosY
				p.Position.Z += speed * sinY
			case "d":
				p.Position.X += speed * cosY
				p.Position.Z -= speed * sinY
			case "arrowup", " ":
				p.Position.Y += speed
			case "arrowdown", "c":
				p.Position.Y -= speed
			case "shift":
				p.Position.W -= speed
			case "e":
				p.Position.W += speed
			case "q":
				p.HyperRot -= 0.05
			case "r":
				p.HyperRot += 0.05
			}

			// Check Quantum Orb collision on move
			GameHub.CheckCollectOrb(p)
		}
		p.Mutex.Unlock()
	}
}

// PhysicsLoop computes 4D projection with full FPS camera, streaming smoothly at 60 FPS
func (p *Player) PhysicsLoop(ctx context.Context) {
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
			score := p.Score
			p.Mutex.RUnlock()

			// 1. Calculate Schwarzschild Time Dilation Multiplier
			timeMultiplier := CalculateTimeMultiplier(currentPos.W)
			effectiveTickMs := float64(BaseTick) / timeMultiplier

			// 2. Fetch other players snapshot, leaderboard, and Quantum Orbs
			otherPlayers, leaderboard, orbs := GameHub.GetGameSnapshot(p.ID)

			// 3. Generate 2D projected lines with Orbs and Avatars
			lines := GenerateProjectedLines(currentPos, yaw, pitch, hyperRot, otherPlayers, orbs)

			// 4. Construct payload
			payload := FramePayload{
				Lines:          lines,
				PlayerPos:      currentPos,
				TimeMultiplier: timeMultiplier,
				TickMs:         effectiveTickMs,
				PlayerID:       p.ID,
				Score:          score,
				Leaderboard:    leaderboard,
				OtherPlayers:   otherPlayers,
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
			p.HyperRot += 0.006 * timeMultiplier
			p.Mutex.Unlock()
		}
	}
}
