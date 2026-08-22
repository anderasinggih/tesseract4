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

// AttackArc represents a real-time cyber projectile flying over the world map
type AttackArc struct {
	ID          string   `json:"id"`
	AttackerID  string   `json:"attackerId"`
	Attacker    string   `json:"attacker"`
	OriginCity  string   `json:"originCity"`
	OriginCoord GeoCoord `json:"originCoord"`
	TargetCity  string   `json:"targetCity"`
	TargetCoord GeoCoord `json:"targetCoord"`
	TargetIP    string   `json:"targetIp"`
	TargetASN   string   `json:"targetAsn"`
	Vector      string   `json:"vector"` // "DDoS SYN Flood", "Ransomware Encryptor", "Kernel 0-Day Exploit", "DNS Amplification"
	Color       string   `json:"color"`  // "#FF0033" (Red), "#9900FF" (Purple), "#00FF66" (Green)
	Progress    float64  `json:"progress"` // 0.0 to 1.0 (Interpolated geodesic flight)
	CurrentPos  GeoCoord `json:"currentPos"`
	PayloadSize int      `json:"payloadSize"` // in KB or MB
	Entropy     float64  `json:"entropy"`     // Shannon entropy index
	Neutralized bool     `json:"neutralized"`
	MitigatedBy string   `json:"mitigatedBy,omitempty"`
}

// SentinelAgent represents an active connected player/operator
type SentinelAgent struct {
	ID        string   `json:"id"`
	Alias     string   `json:"alias"`
	Faction   string   `json:"faction"` // "red", "blue"
	City      string   `json:"city"`
	Coord     GeoCoord `json:"coord"`
	Attacks   int      `json:"attacks"`
	Defends   int      `json:"defends"`
	Score     int      `json:"score"`
}

// TerminalLog represents a real-time cybersecurity telemetry event
type TerminalLog struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"` // "WARN", "CRIT", "DEFENSE", "INFO"
	Message   string `json:"message"`
	Tag       string `json:"tag"`
}

// TargetNode represents a global infrastructure point with live health
type TargetNode struct {
	City    string   `json:"city"`
	Country string   `json:"country"`
	Coord   GeoCoord `json:"coord"`
	IP      string   `json:"ip"`
	ASN     string   `json:"asn"`
	Health  int      `json:"health"` // 0 to 100%
	Shield  bool     `json:"shield"`
}

// WarRoomState is broadcast to clients at 30Hz
type WarRoomState struct {
	Type          string          `json:"type"` // "state_update"
	AgentID       string          `json:"agentId"`
	Agents        []SentinelAgent `json:"agents"`
	ActiveArcs    []AttackArc     `json:"activeArcs"`
	Nodes         []TargetNode    `json:"nodes"`
	Logs          []TerminalLog   `json:"logs"`
	RedScore      int             `json:"redScore"`
	BlueScore     int             `json:"blueScore"`
	GlobalEntropy float64         `json:"globalEntropy"`
	ThreatLevel   string          `json:"threatLevel"` // "DEFCON 4", "DEFCON 3", "DEFCON 2", "DEFCON 1"
}

// ClientCommand represents commands sent by users
type ClientCommand struct {
	Type       string  `json:"type"` // "join", "attack", "defend"
	Alias      string  `json:"alias,omitempty"`
	Faction    string  `json:"faction,omitempty"` // "red", "blue"
	City       string  `json:"city,omitempty"`
	TargetCity string  `json:"targetCity,omitempty"`
	Vector     string  `json:"vector,omitempty"`
	ArcID      string  `json:"arcId,omitempty"` // For intercepting
	ActionType string  `json:"actionType,omitempty"` // "bgp_null_route", "rate_limit", "firewall_patch"
	PayloadMB  int     `json:"payloadMb,omitempty"`
}

// Global Actor Model Hub instance
var GlobalWarRoom = NewWarRoomHub()

func main() {
	// Start the Cyber Warfare Hub Goroutine
	go GlobalWarRoom.Run()

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
		ctx, cancel := context.WithCancel(context.Background())
		agentID := uuid.New().String()[:8]

		// Random default city assignment for new sentinel
		defaultNode := GlobalNodes[len(agentID)%len(GlobalNodes)]
		agent := &AgentConnection{
			ID:      agentID,
			Alias:   "Agent-" + agentID,
			Faction: "red",
			City:    defaultNode.City,
			Coord:   defaultNode.Coord,
			Conn:    c,
			Send:    make(chan []byte, 256),
		}

		// Register Sentinel in War Room
		GlobalWarRoom.register <- agent

		// Launch dedicated asynchronous writer
		go agent.WritePump()

		defer func() {
			cancel()
			GlobalWarRoom.unregister <- agent
			c.Close()
		}()

		// Launch telemetry stream
		go agent.TelemetryLoop(ctx)

		// Read incoming player actions
		agent.ReadPump(c)
	}))

	log.Println("🛡️ Cyber Warfare Ops Server is live on http://localhost:8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
