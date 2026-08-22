package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

var AirlineFleets = []struct {
	Prefix  string
	Airline string
	Types   []string
}{
	{"GIA", "Garuda Indonesia", []string{"B77W", "A339", "B738"}},
	{"SIA", "Singapore Airlines", []string{"A359", "B78X", "B77W"}},
	{"LNI", "Lion Air", []string{"B739", "B738", "A339"}},
	{"QFA", "Qantas Airways", []string{"B789", "A332", "A388"}},
	{"THA", "Thai Airways", []string{"A359", "B77W", "B788"}},
	{"MAS", "Malaysia Airlines", []string{"A359", "A333", "B738"}},
	{"AWQ", "AirAsia", []string{"A320", "A321", "A333"}},
	{"BTK", "Batik Air", []string{"A320", "B738", "B739"}},
	{"CTV", "Citilink", []string{"A320", "A339", "ATR72"}},
	{"SJY", "Sriwijaya Air", []string{"B738", "B735", "B733"}},
	{"TRA", "TransNusa", []string{"ARJ21", "A320", "ATR42"}},
	{"BVA", "Pelita Air", []string{"A320", "AT76", "B734"}},
	{"ILM", "Super Air Jet", []string{"A320", "A321", "A320"}},
	{"WON", "Wings Air", []string{"AT76", "AT75", "AT72"}},
	{"SUS", "Susi Air", []string{"C208", "PC6", "PA18"}},
	{"ALR", "Aviastar", []string{"DHC6", "BAe146", "C208"}},
	{"TGN", "Trigana Air", []string{"AT72", "B733", "DHC6"}},
	{"NEX", "Raya Airways", []string{"B762", "B734", "B763"}},
	{"AXM", "AirAsia Malaysia", []string{"A320", "A321", "A20N"}},
	{"FFM", "Firefly", []string{"AT72", "B738", "AT76"}},
	{"MXD", "Batik Air Malaysia", []string{"B738", "B38M", "A333"}},
	{"MSR", "MYAirline", []string{"A320", "A20N", "A320"}},
	{"SCO", "Scoot", []string{"B788", "B789", "A320"}},
	{"JSA", "Jetstar Asia", []string{"A320", "A321", "A21N"}},
	{"BKP", "Bangkok Airways", []string{"A319", "A320", "AT76"}},
	{"NOK", "Nok Air", []string{"B738", "B738", "Q400"}},
	{"TLM", "Thai Lion Air", []string{"B738", "B739", "B38M"}},
	{"TVJ", "Thai VietJet", []string{"A320", "A321", "A21N"}},
	{"FDX", "FedEx Express", []string{"B77L", "B763", "MD11"}},
	{"UPS", "UPS Airlines", []string{"B748", "B763", "MD11"}},
	{"CLX", "Cargolux", []string{"B748", "B744", "B748"}},
	{"SQC", "Singapore Cargo", []string{"B744", "B77L", "B744"}},
	{"GIC", "Garuda Cargo", []string{"A332", "B738", "A333"}},
	{"LNC", "Lion Cargo", []string{"B738", "A333", "B739"}},
	{"MAS_C", "MASkargo", []string{"A332", "B744", "A332"}},
	{"PAC", "Polar Air Cargo", []string{"B748", "B77L", "B763"}},
	{"KPA", "K-Mile Air", []string{"B734", "B737", "B738"}},
	{"WIA", "Cardig Air", []string{"B733", "B734", "B738"}},
	{"BVA_C", "Pelita Cargo", []string{"AT72", "B734", "AT72"}},
	{"TRI_C", "Tri-M.G. Cargo", []string{"B733", "B734", "B722"}},
}

// ATCHub manages active aircraft kinematics, sectors, collision alerts, and handoffs
type ATCHub struct {
	clients     map[*ATCClientConnection]bool
	register    chan *ATCClientConnection
	unregister  chan *ATCClientConnection
	commands    chan func()
	
	mu          sync.RWMutex
	sectors     map[string]*AirspaceSector
	aircraft    map[string]*Aircraft
	controllers map[string]*ControllerSession
	logs        []ATCLogEntry
}

func NewATCHub() *ATCHub {
	h := &ATCHub{
		clients:     make(map[*ATCClientConnection]bool),
		register:    make(chan *ATCClientConnection, 1024),
		unregister:  make(chan *ATCClientConnection, 1024),
		commands:    make(chan func(), 1024),
		sectors:     make(map[string]*AirspaceSector),
		aircraft:    make(map[string]*Aircraft),
		controllers: make(map[string]*ControllerSession),
		logs:        make([]ATCLogEntry, 0, 100),
	}

	// Initialize default airspace sectors
	for _, sec := range GlobalAirspaceSectors {
		secCopy := sec
		secCopy.Controller = "AUTO-TOWER"
		h.sectors[sec.ID] = &secCopy
	}

	// Pre-spawn initial commercial flights across all sectors
	for i := 0; i < 40; i++ {
		h.spawnCommercialFlight()
	}

	h.addLog("SYS", "ALL", "Air Traffic Control radar network online. STCA Conflict Alert active.")
	return h
}

func (h *ATCHub) addLog(logType, callsign, message string) {
	now := time.Now().Format("15:04:05")
	entry := ATCLogEntry{
		Timestamp: now,
		Type:      logType,
		Callsign:  callsign,
		Message:   message,
	}
	h.logs = append(h.logs, entry)
	if len(h.logs) > 50 {
		h.logs = h.logs[len(h.logs)-50:]
	}
}

func (h *ATCHub) spawnCommercialFlight() {
	if len(GlobalOfficialFlightSchedules) == 0 {
		return
	}

	// Pick a real-world scheduled route
	sched := GlobalOfficialFlightSchedules[rand.Intn(len(GlobalOfficialFlightSchedules))]
	callsign := sched.Callsign

	// Avoid duplicate active flights
	if _, exists := h.aircraft[callsign]; exists {
		return
	}

	// Lookup official origin and destination airports
	var orig, dest *Airport
	for _, apt := range GlobalMajorAirports {
		if apt.ICAO == sched.Origin {
			aptCopy := apt
			orig = &aptCopy
		}
		if apt.ICAO == sched.Destination {
			aptCopy := apt
			dest = &aptCopy
		}
	}

	if orig == nil || dest == nil {
		return
	}

	squawk := fmt.Sprintf("%04d", rand.Intn(7000)+1000)
	bearing := CalculateTrueBearing(orig.Coord, dest.Coord)

	// Determine starting sector
	sectorID := "sec-wiii"
	for _, sec := range GlobalAirspaceSectors {
		if IsPointInSector(orig.Coord, sec.Bounds) {
			sectorID = sec.ID
			break
		}
	}

	// Position aircraft along its authentic great-circle flight path
	startFraction := rand.Float64() * 0.85
	startCoord := GeoCoord{
		Lat: orig.Coord.Lat + (dest.Coord.Lat-orig.Coord.Lat)*startFraction,
		Lng: orig.Coord.Lng + (dest.Coord.Lng-orig.Coord.Lng)*startFraction,
	}

	bearing := CalculateTrueBearing(startCoord, dest.Coord)

	h.aircraft[callsign] = &Aircraft{
		Callsign:       callsign,
		Airline:        sched.Airline,
		AircraftType:   sched.AircraftType,
		Squawk:         squawk,
		Origin:         orig.ICAO,
		Destination:    dest.ICAO,
		Coord:          startCoord,
		Altitude:       sched.CruiseFL,
		TargetAltitude: sched.CruiseFL,
		Heading:        bearing,
		TargetHeading:  bearing,
		Speed:          float64(sched.CruiseSpeed),
		TargetSpeed:    float64(sched.CruiseSpeed),
		SectorID:       sectorID,
		HandoffState:   "NONE",
		ConflictAlert:  false,
		Trail:          make([]GeoCoord, 0, 8),
	}
}

// Run executes the continuous aircraft kinematics loop and collision checks at 10Hz
func (h *ATCHub) Run() {
	log.Println("✈️ [ATC Hub] Air traffic kinematics & collision alert engine started.")
	ticker := time.NewTicker(100 * time.Millisecond) // 10Hz physics tick
	defer ticker.Stop()

	respawnTicker := time.NewTicker(8 * time.Second)
	defer respawnTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.mu.Lock()
			h.controllers[client.ID] = &ControllerSession{
				ID:        client.ID,
				Callsign:  client.Callsign,
				SectorID:  client.SectorID,
				Score:     100,
				Handled:   0,
				Conflicts: 0,
			}
			if sec, ok := h.sectors[client.SectorID]; ok {
				sec.Controller = client.Callsign
				sec.ControllerID = client.ID
			}
			h.addLog("SYS", "ALL", fmt.Sprintf("Operator [%s] assumed control of sector [%s]", client.Callsign, client.SectorID))
			h.mu.Unlock()

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.mu.Lock()
				if sec, ok := h.sectors[client.SectorID]; ok {
					if sec.ControllerID == client.ID {
						sec.Controller = "AUTO-TOWER"
						sec.ControllerID = ""
					}
				}
				delete(h.controllers, client.ID)
				h.addLog("SYS", "ALL", fmt.Sprintf("Operator [%s] went offline. Sector assigned to AUTO-TOWER.", client.Callsign))
				h.mu.Unlock()
			}

		case cmd := <-h.commands:
			cmd()

		case <-respawnTicker.C:
			h.mu.Lock()
			if len(h.aircraft) < 25 {
				h.spawnCommercialFlight()
			}
			h.mu.Unlock()

		case <-ticker.C:
			h.updateKinematics(0.1) // dt = 0.1s
		}
	}
}

// updateKinematics moves all aircraft, applies heading/altitude steering, and runs STCA collision checks
func (h *ATCHub) updateKinematics(dt float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Move and steer each aircraft
	for callsign, ac := range h.aircraft {
		// Flight Management System (FMS) Auto-Nav: If heading not manually assigned by ATC, continuously track destination
		if ac.TargetHeading == ac.Heading {
			for _, apt := range GlobalMajorAirports {
				if apt.ICAO == ac.Destination {
					ac.TargetHeading = CalculateTrueBearing(ac.Coord, apt.Coord)
					break
				}
			}
		}

		// Heading turn rate (approx 3 deg/sec standard rate 1 turn)
		turnRate := 3.0 * dt
		diffHeading := ac.TargetHeading - ac.Heading
		for diffHeading > 180 {
			diffHeading -= 360
		}
		for diffHeading < -180 {
			diffHeading += 360
		}

		if math.Abs(diffHeading) <= turnRate {
			ac.Heading = ac.TargetHeading
		} else if diffHeading > 0 {
			ac.Heading += turnRate
		} else {
			ac.Heading -= turnRate
		}
		if ac.Heading < 0 {
			ac.Heading += 360
		} else if ac.Heading >= 360 {
			ac.Heading -= 360
		}

		// Altitude climb/descent rate (~1500 fpm = 25 fps)
		altRate := int(25.0 * dt * 10)
		if ac.Altitude < ac.TargetAltitude {
			ac.Altitude += altRate
			if ac.Altitude > ac.TargetAltitude {
				ac.Altitude = ac.TargetAltitude
			}
		} else if ac.Altitude > ac.TargetAltitude {
			ac.Altitude -= altRate
			if ac.Altitude < ac.TargetAltitude {
				ac.Altitude = ac.TargetAltitude
			}
		}

		// Speed adjust rate
		if ac.Speed < ac.TargetSpeed {
			ac.Speed += 10.0 * dt
		} else if ac.Speed > ac.TargetSpeed {
			ac.Speed -= 10.0 * dt
		}

		// Advance position along heading (Kinematics formula)
		ac.Coord = MoveAircraftPosition(ac.Coord, ac.Heading, ac.Speed, dt)

		// Record radar breadcrumb trail
		if len(ac.Trail) == 0 || DistanceInNauticalMiles(ac.Trail[len(ac.Trail)-1], ac.Coord) > 5.0 {
			ac.Trail = append(ac.Trail, ac.Coord)
			if len(ac.Trail) > 6 {
				ac.Trail = ac.Trail[1:]
			}
		}

		// Check airspace sector transition
		for _, sec := range GlobalAirspaceSectors {
			if IsPointInSector(ac.Coord, sec.Bounds) {
				if ac.SectorID != sec.ID && ac.HandoffState == "NONE" {
					// Aircraft crossed into new sector
					ac.SectorID = sec.ID
				}
				break
			}
		}

		// Check if arrived near destination airport (< 8 NM)
		for _, apt := range GlobalMajorAirports {
			if apt.ICAO == ac.Destination && DistanceInNauticalMiles(ac.Coord, apt.Coord) < 10.0 {
				h.addLog("RADIO", callsign, fmt.Sprintf("%s safely landed at %s (%s)", callsign, apt.Name, apt.ICAO))
				delete(h.aircraft, callsign)
				break
			}
		}
	}

	// 2. Short-Term Conflict Alert (STCA) Loss of Separation calculation
	// Lateral < 5 NM and Vertical < 1000 ft
	for _, ac1 := range h.aircraft {
		ac1.ConflictAlert = false
	}

	for _, ac1 := range h.aircraft {
		for _, ac2 := range h.aircraft {
			if ac1.Callsign == ac2.Callsign {
				continue
			}

			latDist := DistanceInNauticalMiles(ac1.Coord, ac2.Coord)
			altDiff := math.Abs(float64(ac1.Altitude - ac2.Altitude))

			if latDist < 5.0 && altDiff < 1000.0 {
				ac1.ConflictAlert = true
				ac2.ConflictAlert = true
				if rand.Float64() < 0.05 {
					h.addLog("ALERT", ac1.Callsign, fmt.Sprintf("TRAFFIC ALERT! %s & %s loss of separation (%.1f NM, %.0f ft diff)", ac1.Callsign, ac2.Callsign, latDist, altDiff))
				}
			}
		}
	}
}

// GetSnapshot constructs the payload for streaming to clients
func (h *ATCHub) GetSnapshot(selfID string) ATCPayload {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sectorsList := make([]AirspaceSector, 0, len(h.sectors))
	for _, s := range h.sectors {
		sectorsList = append(sectorsList, *s)
	}

	aircraftList := make([]Aircraft, 0, len(h.aircraft))
	activeAlerts := 0
	for _, a := range h.aircraft {
		aircraftList = append(aircraftList, *a)
		if a.ConflictAlert {
			activeAlerts++
		}
	}

	controllersList := make([]ControllerSession, 0, len(h.controllers))
	mySectorID := "sec-wiii"
	for _, c := range h.controllers {
		controllersList = append(controllersList, *c)
		if c.ID == selfID {
			mySectorID = c.SectorID
		}
	}

	// Inbound handoffs for current player's sector
	handoffs := make([]HandoffNotification, 0)
	for _, a := range h.aircraft {
		if a.HandoffState == "PENDING" && a.HandoffTarget == mySectorID {
			fromSecName := a.SectorID
			if s, ok := h.sectors[a.SectorID]; ok {
				fromSecName = s.Name
			}
			handoffs = append(handoffs, HandoffNotification{
				Callsign:     a.Callsign,
				FromSector:   fromSecName,
				ToSector:     mySectorID,
				FromOperator: a.SectorID,
			})
		}
	}

	logsCopy := make([]ATCLogEntry, len(h.logs))
	copy(logsCopy, h.logs)

	return ATCPayload{
		Type:         "radar_update",
		ControllerID: selfID,
		SectorID:     mySectorID,
		Sectors:      sectorsList,
		Airports:     GlobalMajorAirports,
		AircraftList: aircraftList,
		Controllers:  controllersList,
		Handoffs:     handoffs,
		Logs:         logsCopy,
		TotalFlights: len(h.aircraft),
		ActiveAlerts: activeAlerts,
	}
}

// ExecuteATCCommand modifies flight parameters or completes handoffs
func (h *ATCHub) ExecuteATCCommand(cmd ClientATCCommand, operatorID string) {
	h.commands <- func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		switch cmd.Type {
		case "claim_sector":
			if sec, ok := h.sectors[cmd.SectorID]; ok {
				// Vacate old sector
				for _, s := range h.sectors {
					if s.ControllerID == operatorID {
						s.Controller = "AUTO-TOWER"
						s.ControllerID = ""
					}
				}
				sec.Controller = cmd.Controller
				sec.ControllerID = operatorID
				if c, ok := h.controllers[operatorID]; ok {
					c.SectorID = cmd.SectorID
					c.Callsign = cmd.Controller
				}
				h.addLog("SYS", "ALL", fmt.Sprintf("[%s] took over sector %s", cmd.Controller, sec.Name))
			}

		case "set_heading":
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				ac.TargetHeading = cmd.Heading
				h.addLog("RADIO", ac.Callsign, fmt.Sprintf("%s, fly heading %03.0f", ac.Callsign, cmd.Heading))
			}

		case "set_altitude":
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				ac.TargetAltitude = cmd.Altitude
				h.addLog("RADIO", ac.Callsign, fmt.Sprintf("%s, climb/descend and maintain FL%d", ac.Callsign, cmd.Altitude/100))
			}

		case "set_speed":
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				ac.TargetSpeed = cmd.Speed
				h.addLog("RADIO", ac.Callsign, fmt.Sprintf("%s, adjust speed to %.0f knots", ac.Callsign, cmd.Speed))
			}

		case "handoff_init":
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				ac.HandoffState = "PENDING"
				ac.HandoffTarget = cmd.ToSector
				h.addLog("HANDOFF", ac.Callsign, fmt.Sprintf("Handoff initiated for %s to %s", ac.Callsign, cmd.ToSector))
			}

		case "handoff_accept":
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				ac.HandoffState = "NONE"
				ac.SectorID = ac.HandoffTarget
				ac.HandoffTarget = ""
				if c, ok := h.controllers[operatorID]; ok {
					c.Score += 25
					c.Handled++
				}
				h.addLog("HANDOFF", ac.Callsign, fmt.Sprintf("Handoff ACCEPTED: %s is now under control of %s", ac.Callsign, ac.SectorID))
			}
		}
	}
}
