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

	// Pre-spawn initial commercial flights across all sectors (Dense global traffic: 150 flights)
	for i := 0; i < 150; i++ {
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

	// If callsign exists, generate unique flight suffix (e.g. GIA880B, SIA318X) to allow dense multi-flight traffic
	if _, exists := h.aircraft[callsign]; exists {
		suffixes := []string{"A", "B", "C", "D", "X", "Y", "Z", "1", "2"}
		callsign = fmt.Sprintf("%s%s", sched.Callsign, suffixes[rand.Intn(len(suffixes))])
		if _, exists := h.aircraft[callsign]; exists {
			return
		}
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
		EmergencyState: false,
		EmergencyType:  "",
		Trail:          make([]GeoCoord, 0, 8),
	}
}

// Run executes the continuous aircraft kinematics loop and collision checks at 60Hz
func (h *ATCHub) Run() {
	log.Println("[ATC Hub] Air traffic kinematics & collision alert engine started at 60 FPS.")
	ticker := time.NewTicker(16 * time.Millisecond) // 60Hz (16.6ms) physics tick
	defer ticker.Stop()

	respawnTicker := time.NewTicker(4 * time.Second)
	defer respawnTicker.Stop()

	emergencyTicker := time.NewTicker(90 * time.Second) // Occasional in-flight emergency challenge
	defer emergencyTicker.Stop()

	lastTick := time.Now()
	// ── TEMP diagnostics: verify live tick cadence matches wall clock ──
	diagWindow := 300 // ~5s of ticks
	tickCount := 0
	wallStart := time.Now()
	simAdvance := 0.0

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
			if len(h.aircraft) < 150 {
				h.spawnCommercialFlight()
			}
			h.mu.Unlock()

		case <-emergencyTicker.C:
			h.mu.Lock()
			// Limit to maximum 1 active emergency at any time across the entire world
			activeEmergencies := 0
			for _, ac := range h.aircraft {
				if ac.EmergencyState {
					activeEmergencies++
				}
			}

			if activeEmergencies == 0 && len(h.aircraft) > 0 {
				// Pick one random flight
				var candidateList []*Aircraft
				for _, ac := range h.aircraft {
					candidateList = append(candidateList, ac)
				}
				if len(candidateList) > 0 {
					targetAc := candidateList[rand.Intn(len(candidateList))]
					targetAc.EmergencyState = true
					targetAc.Squawk = "7700"
					emergencyTypes := []string{"ENGINE_FAIL", "MED_EMERG", "CABIN_ALT"}
					targetAc.EmergencyType = emergencyTypes[rand.Intn(len(emergencyTypes))]
					targetAc.TargetAltitude = 10000 // Priority emergency descent to FL100
					h.addLog("ALERT", targetAc.Callsign, fmt.Sprintf("MAYDAY 7700: %s reports %s! Descending to FL100.", targetAc.Callsign, targetAc.EmergencyType))
				}
			}
			h.mu.Unlock()

		case now := <-ticker.C:
			dt := now.Sub(lastTick).Seconds()
			lastTick = now
			if dt > 0.1 {
				dt = 0.1
			}
			h.updateKinematics(dt)
			tickCount++
			simAdvance += dt
			if tickCount >= diagWindow {
				wallElapsed := time.Since(wallStart).Seconds()
				log.Printf("[DIAG] %d ticks | wall=%.3fs sim=%.3fs ratio=%.2fx | aircraft=%d",
					diagWindow, wallElapsed, simAdvance,
					simAdvance/wallElapsed, h.aircraftCount())
				tickCount = 0
				simAdvance = 0
				wallStart = time.Now()
			}
		}
	}
}

// aircraftCount returns the current number of simulated flights (diagnostics).
func (h *ATCHub) aircraftCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.aircraft)
}

// headingDiff returns shortest signed angular difference a-b in (-180, +180]
func headingDiff(a, b float64) float64 {
	d := a - b
	for d > 180 {
		d -= 360
	}
	for d < -180 {
		d += 360
	}
	return d
}

// updateKinematics moves all aircraft, applies heading/altitude steering, and runs STCA collision checks
func (h *ATCHub) updateKinematics(dt float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Move and steer each aircraft
	for callsign, ac := range h.aircraft {
		// Flight Management System (FMS) Auto-Nav: when wings level and on assigned
		// track, continuously re-track the destination airport
		if math.Abs(headingDiff(ac.TargetHeading, ac.Heading)) < 0.05 && math.Abs(ac.Roll) < 0.5 {
			for _, apt := range GlobalMajorAirports {
				if apt.ICAO == ac.Destination {
					ac.TargetHeading = CalculateTrueBearing(ac.Coord, apt.Coord)
					break
				}
			}
		}

		// ── Coordinated turn dynamics (FAA Pilot's Handbook of Aeronautical Knowledge) ──
		// Bank command: pilot stick overrides directly; otherwise bank-angle pursuit
		// of the assigned heading (3x gain, saturated at airliner limit 25 deg).
		bankCmd := ac.TargetRoll
		if ac.TargetRoll == 0 {
			bankCmd = math.Max(-25.0, math.Min(25.0, 3.0*headingDiff(ac.TargetHeading, ac.Heading)))
		}

		// Airliner roll rate (Fast arcade response: 60 deg/s so bank is instantaneous)
		if ac.Roll < bankCmd {
			ac.Roll = math.Min(ac.Roll+60.0*dt, bankCmd)
		} else if ac.Roll > bankCmd {
			ac.Roll = math.Max(ac.Roll-60.0*dt, bankCmd)
		}

		// Rate of turn: Responsive turn rate (~18 deg/s at 25 deg bank)
		// Multiplied for responsive arcade-style control like GTA
		tas := math.Max(ac.Speed, 120.0)
		rot := 1091.0 * math.Tan(ac.Roll*math.Pi/180.0) / tas * 15.0
		ac.Heading += rot * dt
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

		// Check airspace sector transition & trigger clean single-shot handoff protocol
		for _, sec := range GlobalAirspaceSectors {
			if IsPointInSector(ac.Coord, sec.Bounds) {
				if ac.SectorID != sec.ID {
					if ac.HandoffState == "NONE" && ac.HandoffTarget != sec.ID {
						// Aircraft is entering a new sector: initiate clean handoff request
						ac.HandoffState = "PENDING"
						ac.HandoffTarget = sec.ID
						fromName := ac.SectorID
						if oldSec, ok := h.sectors[ac.SectorID]; ok {
							fromName = oldSec.Name
						}
						h.addLog("HANDOFF", ac.Callsign, fmt.Sprintf("INBOUND HANDOFF: %s requesting entry from %s to %s", ac.Callsign, fromName, sec.Name))
					}
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

	// 2. Short-Term Conflict Alert (STCA) & Mid-Air Crash Detection
	for _, ac1 := range h.aircraft {
		ac1.ConflictAlert = false
	}

	var crashedFlights []string
	for _, ac1 := range h.aircraft {
		for _, ac2 := range h.aircraft {
			if ac1.Callsign == ac2.Callsign {
				continue
			}

			latDist := DistanceInNauticalMiles(ac1.Coord, ac2.Coord)
			altDiff := math.Abs(float64(ac1.Altitude - ac2.Altitude))

			// CRITICAL MID-AIR COLLISION / CRASH (< 0.4 NM and < 200 ft)
			if latDist < 0.4 && altDiff < 200.0 {
				crashedFlights = append(crashedFlights, ac1.Callsign, ac2.Callsign)
				h.addLog("CRASH", "EMERGENCY", fmt.Sprintf("⚠️ MID-AIR COLLISION! %s and %s collided at FL%d (%.2f NM).", ac1.Callsign, ac2.Callsign, ac1.Altitude/100, latDist))
				break
			}

			// Loss of Separation Warning (< 5 NM and < 1000 ft)
			if latDist < 5.0 && altDiff < 1000.0 {
				ac1.ConflictAlert = true
				ac2.ConflictAlert = true
				if rand.Float64() < 0.05 {
					h.addLog("ALERT", ac1.Callsign, fmt.Sprintf("TRAFFIC ALERT! %s & %s loss of separation (%.1f NM, %.0f ft diff)", ac1.Callsign, ac2.Callsign, latDist, altDiff))
				}
			}
		}
	}

	// Remove destroyed aircraft from radar
	if len(crashedFlights) > 0 {
		for _, cs := range crashedFlights {
			delete(h.aircraft, cs)
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

		case "steer_start":
			// Pilot stick: hold a coordinated bank. cmd.Heading = -1 (left) / +1 (right)
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				dir := cmd.Heading
				if dir > 0 {
					ac.TargetRoll = 25
				} else if dir < 0 {
					ac.TargetRoll = -25
				}
			}

		case "steer_stop":
			// Release: roll back to wings level and hold the CURRENT track —
			// no snap-back to stale targets. FMS auto-nav resumes naturally.
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				ac.TargetRoll = 0
				ac.TargetHeading = ac.Heading
			}

		case "adjust_altitude":
			if ac, ok := h.aircraft[cmd.Callsign]; ok {
				na := ac.Altitude + cmd.Altitude
				if na < 1000 {
					na = 1000
				}
				if na > 45000 {
					na = 45000
				}
				ac.TargetAltitude = na
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
