package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// Vector4 represents a 4-dimensional point (X, Y, Z, W)
type Vector4 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	W float64 `json:"w"`
}

// Line2D represents a projected 2D line segment with color styling
type Line2D struct {
	X1    float64 `json:"x1"`
	Y1    float64 `json:"y1"`
	X2    float64 `json:"x2"`
	Y2    float64 `json:"y2"`
	Color string  `json:"color"` // "#00FF00", "#0088FF", "#FF0055", "#444444"
}

// Orb represents a collectible Quantum Core in 3D/4D space
type Orb struct {
	ID       int     `json:"id"`
	Position Vector4 `json:"position"`
	Active   bool    `json:"active"`
}

// PlayerState represents another player's snapshot
type PlayerState struct {
	ID       string  `json:"id"`
	Position Vector4 `json:"position"`
	Yaw      float64 `json:"yaw"`
	Pitch    float64 `json:"pitch"`
	Score    int     `json:"score"`
}

// FramePayload is sent over WebSocket to the dumb client
type FramePayload struct {
	Lines          []Line2D      `json:"lines"`
	PlayerPos      Vector4       `json:"playerPos"`
	TimeMultiplier float64       `json:"timeMultiplier"`
	TickMs         float64       `json:"tickMs"`
	PlayerID       string        `json:"playerId"`
	Score          int           `json:"score"`
	Leaderboard    []PlayerState `json:"leaderboard"`
	OtherPlayers   []PlayerState `json:"otherPlayers"`
}

// InputCommand represents raw movement or mouse look from the frontend
type InputCommand struct {
	Type  string  `json:"type"`  // "move", "look"
	Key   string  `json:"key"`   // "w", "s", "a", "d", "arrowup", "arrowdown", "shift", "space", "q", "e"
	Delta float64 `json:"delta"` // magnitude
	DX    float64 `json:"dx"`    // Mouse delta X (Yaw)
	DY    float64 `json:"dy"`    // Mouse delta Y (Pitch)
}


// Global Actor Model Hub instance
var GameHub = NewHub()

func main() {
	// Start the "Mandor" Broker Goroutine
	go GameHub.Run()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	app.Use(cors.New())
	app.Use(pprof.New())

	// Serve static frontend files
	app.Static("/", "./public")

	// WebSocket upgrade middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket route
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// 1. Master Switch (Context with Cancel)
		ctx, cancel := context.WithCancel(context.Background())

		playerID := uuid.New().String()
		player := &Player{
			ID: playerID,
			Position: Vector4{
				X: 0,
				Y: 0,
				Z: 0,
				W: 1.0, // Initial distance from singularity
			},
			Yaw:      0,
			Pitch:    0,
			HyperRot: 0,
			Conn:     c,
			Send:     make(chan []byte, 256), // Buffer antrean pengiriman keluar
		}

		// 1. Kirim surat pendaftaran ke Mandor (Zero Mutex!)
		GameHub.register <- player

		// 2. Luncurkan Goroutine "Kurir" (WritePump) di background
		go player.WritePump()

		// 3. Guaranteed Cleanup via Defer
		defer func() {
			// Pull Master Switch: instantly terminates PhysicsAndTemporalLoop
			cancel()

			// Kirim surat pengunduran diri ke Mandor (Zero Mutex contention)
			GameHub.unregister <- player

			// Close physical socket connection
			c.Close()
		}()

		// 4. Launch Physics & Temporal calculation in background goroutine with context
		go player.PhysicsLoop(ctx)

		// 5. ReadPump (Blocks main handler loop while socket is alive)
		player.ReadPump(c)
	}))

	log.Println("⚡ The Tesseract Paradox Server is starting on http://localhost:8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
