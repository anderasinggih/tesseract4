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

// Aircraft represents a real-time commercial flight tracked on the ATC radar scope
type Aircraft struct {
	Callsign       string   `json:"callsign"`       // e.g. "GIA880", "SIA318"
	Airline        string   `json:"airline"`        // e.g. "Garuda Indonesia"
	AircraftType   string   `json:"aircraftType"`   // e.g. "B77W", "A359", "A320"
	Squawk         string   `json:"squawk"`         // e.g. "4521"
	Origin         string   `json:"origin"`         // e.g. "WIII" (CGK)
	Destination    string   `json:"destination"`    // e.g. "WSSS" (SIN)
	Coord          GeoCoord `json:"coord"`          // Current position
	Altitude       int      `json:"altitude"`       // Current altitude in feet
	TargetAltitude int      `json:"targetAltitude"` // Assigned flight level in feet
	Heading        float64  `json:"heading"`        // Current heading (0-360 deg)
	TargetHeading  float64  `json:"targetHeading"`  // Assigned vector heading
	Roll           float64  `json:"roll"`           // Current bank angle in degrees (-left / +right)
	TargetRoll     float64  `json:"-"`              // Commanded bank angle while pilot stick is held (internal)
	Speed          float64  `json:"speed"`          // Current ground speed in knots
	TargetSpeed    float64  `json:"targetSpeed"`    // Assigned speed in knots
	SectorID       string   `json:"sectorId"`       // Current controlling sector ID
	HandoffState   string   `json:"handoffState"`   // "NONE", "PENDING", "ACCEPTED"
	HandoffTarget  string   `json:"handoffTarget"`  // Sector receiving handoff
	ConflictAlert  bool     `json:"conflictAlert"`  // True if Loss of Separation detected (< 5NM & < 1000ft)
	EmergencyState bool     `json:"emergencyState"` // True if Mayday 7700 emergency active
	EmergencyType  string   `json:"emergencyType"`  // "ENGINE_FAILURE", "MEDICAL", "WEATHER_DEVIATION"
	Trail          []GeoCoord `json:"trail"`        // Past radar history trail
}

// ControllerSession represents an active ATC operator controlling a sector
type ControllerSession struct {
	ID        string `json:"id"`
	Callsign  string `json:"callsign"` // e.g. "JAKARTA_CTR"
	SectorID  string `json:"sectorId"` // e.g. "sec-wiii"
	Score     int    `json:"score"`
	Handled   int    `json:"handled"`
	Conflicts int    `json:"conflicts"`
}

// HandoffNotification alerts a controller about an inbound handoff
type HandoffNotification struct {
	Callsign     string `json:"callsign"`
	FromSector   string `json:"fromSector"`
	ToSector     string `json:"toSector"`
	FromOperator string `json:"fromOperator"`
}

// ATCLogEntry represents live controller communications and radio logs
type ATCLogEntry struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"` // "RADIO", "ALERT", "HANDOFF", "SYS"
	Callsign  string `json:"callsign"`
	Message   string `json:"message"`
}

// ATCPayload is pushed to clients at 15-30Hz
type ATCPayload struct {
	Type          string                `json:"type"` // "radar_update"
	ControllerID  string                `json:"controllerId"`
	SectorID      string                `json:"sectorId"`
	Sectors       []AirspaceSector      `json:"sectors"`
	Airports      []Airport             `json:"airports"`
	AircraftList  []Aircraft            `json:"aircraft"`
	Controllers   []ControllerSession   `json:"controllers"`
	Handoffs      []HandoffNotification `json:"handoffs"`
	Logs          []ATCLogEntry         `json:"logs"`
	TotalFlights  int                   `json:"totalFlights"`
	ActiveAlerts  int                   `json:"activeAlerts"`
}

// ClientATCCommand represents instructions sent from the radar interface
type ClientATCCommand struct {
	Type       string  `json:"type"` // "claim_sector", "set_heading", "set_altitude", "set_speed", "handoff_init", "handoff_accept"
	Callsign   string  `json:"callsign,omitempty"`
	SectorID   string  `json:"sectorId,omitempty"`
	Heading    float64 `json:"heading,omitempty"`
	Altitude   int     `json:"altitude,omitempty"`
	Speed      float64 `json:"speed,omitempty"`
	ToSector   string  `json:"toSector,omitempty"`
	Controller string  `json:"controller,omitempty"`
}

// Global Actor Model ATC Radar Hub
var GlobalATCHub = NewATCHub()

func main() {
	go GlobalATCHub.Run()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	app.Use(cors.New())
	app.Use(pprof.New())
	// Never let browsers cache the UI — stale JS has caused phantom bugs before
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, must-revalidate")
		return c.Next()
	})
	app.Static("/", "./public")

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		ctx, cancel := context.WithCancel(context.Background())
		controllerID := uuid.New().String()[:8]

		// Default claim Jakarta FIR
		session := &ATCClientConnection{
			ID:         controllerID,
			Callsign:   "CTR-" + controllerID,
			SectorID:   "sec-wiii",
			Conn:       c,
			Send:       make(chan []byte, 256),
		}

		GlobalATCHub.register <- session
		go session.WritePump()

		defer func() {
			cancel()
			GlobalATCHub.unregister <- session
			c.Close()
		}()

		go session.TelemetryLoop(ctx)
		session.ReadPump(c)
	}))

	log.Println("✈️ Global ATC Radar Server is live on http://localhost:8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("ATC Server error: %v", err)
	}
}
