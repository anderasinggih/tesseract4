package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WarRoomHub coordinates real-time cyber battles using the zero-lock Actor Model
type WarRoomHub struct {
	clients        map[*AgentConnection]bool
	register       chan *AgentConnection
	unregister     chan *AgentConnection
	broadcast      chan []byte
	agentAction    chan func()
	getStateChan   chan chan WarRoomState

	// Global State
	mu             sync.RWMutex
	agents         map[string]*SentinelAgent
	activeArcs     map[string]*AttackArc
	nodes          map[string]*TargetNode
	logs           []TerminalLog
	redScore       int
	blueScore      int
}

func NewWarRoomHub() *WarRoomHub {
	h := &WarRoomHub{
		clients:      make(map[*AgentConnection]bool),
		register:     make(chan *AgentConnection, 1024),
		unregister:   make(chan *AgentConnection, 1024),
		broadcast:    make(chan []byte, 4096),
		agentAction:  make(chan func(), 1024),
		getStateChan: make(chan chan WarRoomState, 64),
		agents:       make(map[string]*SentinelAgent),
		activeArcs:   make(map[string]*AttackArc),
		nodes:        make(map[string]*TargetNode),
		logs:         make([]TerminalLog, 0, 100),
	}

	// Initialize global target infrastructure nodes
	for _, n := range GlobalNodes {
		h.nodes[n.City] = &TargetNode{
			City:    n.City,
			Country: n.Country,
			Coord:   n.Coord,
			IP:      n.IP,
			ASN:     n.ASN,
			Health:  100,
			Shield:  false,
		}
	}

	h.addLog("INFO", "Global Cyber Defense Grid initialized. Shannon entropy monitors online.", "SYS")
	return h
}

func (h *WarRoomHub) addLog(level, message, tag string) {
	now := time.Now().Format("15:04:05.000")
	entry := TerminalLog{
		Timestamp: now,
		Level:     level,
		Message:   message,
		Tag:       tag,
	}
	h.logs = append(h.logs, entry)
	if len(h.logs) > 60 {
		h.logs = h.logs[len(h.logs)-60:]
	}
}

// Run executes the core event loop and 30Hz ballistic arc physics simulation
func (h *WarRoomHub) Run() {
	log.Println("⚡ [WarRoom Hub] Actor Model cyber battle engine active.")
	ticker := time.NewTicker(33 * time.Millisecond) // ~30Hz simulation
	defer ticker.Stop()

	// Periodic automated background ambient botnet activity
	ambientTicker := time.NewTicker(4 * time.Second)
	defer ambientTicker.Stop()

	for {
		select {
		case agentConn := <-h.register:
			h.clients[agentConn] = true
			h.mu.Lock()
			h.agents[agentConn.ID] = &SentinelAgent{
				ID:      agentConn.ID,
				Alias:   agentConn.Alias,
				Faction: agentConn.Faction,
				City:    agentConn.City,
				Coord:   agentConn.Coord,
				Score:   0,
			}
			h.addLog("INFO", fmt.Sprintf("Operator [%s] connected from %s (Faction: %s)", agentConn.Alias, agentConn.City, agentConn.Faction), "SENTINEL")
			h.mu.Unlock()
			fmt.Printf("[WarRoom] Sentinel %s joined. Total active operators: %d\n", agentConn.ID, len(h.clients))

		case agentConn := <-h.unregister:
			if _, ok := h.clients[agentConn]; ok {
				delete(h.clients, agentConn)
				close(agentConn.Send)
				h.mu.Lock()
				delete(h.agents, agentConn.ID)
				h.addLog("WARN", fmt.Sprintf("Operator [%s] went offline.", agentConn.Alias), "SENTINEL")
				h.mu.Unlock()
				fmt.Printf("[WarRoom] Sentinel %s disconnected. Active: %d\n", agentConn.ID, len(h.clients))
			}

		case action := <-h.agentAction:
			action()

		case <-ambientTicker.C:
			// Spawn random ambient cyber attack if fewer than 8 active arcs
			h.mu.Lock()
			if len(h.activeArcs) < 8 {
				srcIdx := rand.Intn(len(GlobalNodes))
				dstIdx := rand.Intn(len(GlobalNodes))
				if srcIdx != dstIdx {
					src := GlobalNodes[srcIdx]
					dst := GlobalNodes[dstIdx]
					vectors := []string{"DDoS SYN Flood", "Ransomware Encryptor", "Kernel 0-Day Exploit", "DNS Amplification"}
					vec := vectors[rand.Intn(len(vectors))]
					color := "#FF0033"
					if vec == "Ransomware Encryptor" {
						color = "#9900FF"
					} else if vec == "Kernel 0-Day Exploit" {
						color = "#00FF66"
					}

					payload := GenerateRandomPayloadBytes(1024, vec == "Ransomware Encryptor")
					entropy := CalculateShannonEntropy(payload)

					arcID := uuid.New().String()[:6]
					h.activeArcs[arcID] = &AttackArc{
						ID:          arcID,
						AttackerID:  "BOTNET-AUTO",
						Attacker:    "Autonomous Botnet",
						OriginCity:  src.City,
						OriginCoord: src.Coord,
						TargetCity:  dst.City,
						TargetCoord: dst.Coord,
						TargetIP:    dst.IP,
						TargetASN:   dst.ASN,
						Vector:      vec,
						Color:       color,
						Progress:    0.0,
						CurrentPos:  src.Coord,
						PayloadSize: rand.Intn(800) + 100,
						Entropy:     entropy,
						Neutralized: false,
					}
					h.addLog("WARN", fmt.Sprintf("Inbound %s detected from %s -> %s (Target ASN: %s)", vec, src.City, dst.City, dst.ASN), "DETECT")
				}
			}
			h.mu.Unlock()

		case <-ticker.C:
			// Advance ballistic projectile arcs along Geodesic Great-Circle paths
			h.mu.Lock()
			for id, arc := range h.activeArcs {
				// Speed factor (completes in ~4-6 seconds)
				arc.Progress += 0.008
				if arc.Progress >= 1.0 {
					// Arc reached target!
					if !arc.Neutralized {
						if node, ok := h.nodes[arc.TargetCity]; ok {
							if node.Shield {
								node.Shield = false
								h.blueScore += 50
								h.addLog("DEFENSE", fmt.Sprintf("Quantum Firewall deflected %s at %s!", arc.Vector, arc.TargetCity), "FIREWALL")
							} else {
								dmg := rand.Intn(15) + 10
								node.Health -= dmg
								if node.Health < 0 {
									node.Health = 0
								}
								h.redScore += 100
								h.addLog("CRIT", fmt.Sprintf("IMPACT! %s struck %s (%s). Node Health: %d%%", arc.Vector, arc.TargetCity, arc.TargetIP, node.Health), "EXPLOIT")
							}
						}
					}
					delete(h.activeArcs, id)
				} else {
					// Calculate precise current coordinate on spherical globe
					arc.CurrentPos = CalculateGeodesicWaypoint(arc.OriginCoord, arc.TargetCoord, arc.Progress)
				}
			}

			// Slowly regenerate node health
			for _, node := range h.nodes {
				if node.Health < 100 && rand.Float64() < 0.05 {
					node.Health += 1
				}
			}
			h.mu.Unlock()
		}
	}
}

// GetSnapshot retrieves the current war room state for client streaming
func (h *WarRoomHub) GetSnapshot(selfID string) WarRoomState {
	h.mu.RLock()
	defer h.mu.RUnlock()

	agentsList := make([]SentinelAgent, 0, len(h.agents))
	for _, a := range h.agents {
		agentsList = append(agentsList, *a)
	}

	arcsList := make([]AttackArc, 0, len(h.activeArcs))
	var totalEntropy float64
	for _, arc := range h.activeArcs {
		arcsList = append(arcsList, *arc)
		totalEntropy += arc.Entropy
	}

	nodesList := make([]TargetNode, 0, len(h.nodes))
	for _, n := range h.nodes {
		nodesList = append(nodesList, *n)
	}

	logsCopy := make([]TerminalLog, len(h.logs))
	copy(logsCopy, h.logs)

	avgEntropy := 4.2
	if len(h.activeArcs) > 0 {
		avgEntropy = totalEntropy / float64(len(h.activeArcs))
	}

	threat := "DEFCON 4"
	if len(h.activeArcs) > 6 {
		threat = "DEFCON 1"
	} else if len(h.activeArcs) > 4 {
		threat = "DEFCON 2"
	} else if len(h.activeArcs) > 2 {
		threat = "DEFCON 3"
	}

	return WarRoomState{
		Type:          "state_update",
		AgentID:       selfID,
		Agents:        agentsList,
		ActiveArcs:    arcsList,
		Nodes:         nodesList,
		Logs:          logsCopy,
		RedScore:      h.redScore,
		BlueScore:     h.blueScore,
		GlobalEntropy: math.Round(avgEntropy*100) / 100,
		ThreatLevel:   threat,
	}
}

// LaunchAttack creates a new targeted cyber attack arc
func (h *WarRoomHub) LaunchAttack(attackerID, originCity, targetCity, vector string, payloadMB int) {
	h.agentAction <- func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		var srcCoord GeoCoord
		var srcFound bool
		for _, n := range GlobalNodes {
			if n.City == originCity {
				srcCoord = n.Coord
				srcFound = true
				break
			}
		}
		if !srcFound {
			srcCoord = GlobalNodes[0].Coord
		}

		var dstNode *TargetNode
		for _, n := range h.nodes {
			if n.City == targetCity {
				dstNode = n
				break
			}
		}
		if dstNode == nil {
			return
		}

		color := "#FF0033"
		if vector == "Ransomware Encryptor" {
			color = "#9900FF"
		} else if vector == "Kernel 0-Day Exploit" {
			color = "#00FF66"
		}

		payload := GenerateRandomPayloadBytes(1024, vector == "Ransomware Encryptor")
		entropy := CalculateShannonEntropy(payload)

		attackerAlias := "Operator-" + attackerID
		if a, ok := h.agents[attackerID]; ok {
			attackerAlias = a.Alias
			a.Attacks++
			a.Score += 20
		}

		arcID := uuid.New().String()[:6]
		h.activeArcs[arcID] = &AttackArc{
			ID:          arcID,
			AttackerID:  attackerID,
			Attacker:    attackerAlias,
			OriginCity:  originCity,
			OriginCoord: srcCoord,
			TargetCity:  targetCity,
			TargetCoord: dstNode.Coord,
			TargetIP:    dstNode.IP,
			TargetASN:   dstNode.ASN,
			Vector:      vector,
			Color:       color,
			Progress:    0.0,
			CurrentPos:  srcCoord,
			PayloadSize: payloadMB,
			Entropy:     entropy,
			Neutralized: false,
		}

		h.addLog("WARN", fmt.Sprintf("[%s] launched %s at %s (%s, %d MB, Entropy: %.2f)", attackerAlias, vector, targetCity, dstNode.IP, payloadMB, entropy), "RED-OPS")
	}
}

// DeployDefense intercepts or shields an infrastructure node
func (h *WarRoomHub) DeployDefense(defenderID, actionType, arcID, targetCity string) {
	h.agentAction <- func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		defenderAlias := "Sentinel-" + defenderID
		if a, ok := h.agents[defenderID]; ok {
			defenderAlias = a.Alias
			a.Defends++
			a.Score += 50
		}

		switch actionType {
		case "bgp_null_route", "intercept":
			if arc, ok := h.activeArcs[arcID]; ok {
				arc.Neutralized = true
				arc.MitigatedBy = defenderAlias
				arc.Color = "#0088FF"
				h.blueScore += 100
				h.addLog("DEFENSE", fmt.Sprintf("[%s] successfully neutralized attack %s via BGP Null-Route!", defenderAlias, arc.Vector), "BLUE-OPS")
			}

		case "firewall_patch", "shield":
			if node, ok := h.nodes[targetCity]; ok {
				node.Shield = true
				h.blueScore += 30
				h.addLog("DEFENSE", fmt.Sprintf("[%s] deployed Quantum Firewall Shield to %s (%s)", defenderAlias, targetCity, node.IP), "BLUE-OPS")
			}
		}
	}
}

// UpdateAgent updates an operator's alias, faction, or home city
func (h *WarRoomHub) UpdateAgent(agentID, alias, faction, city string) {
	h.agentAction <- func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		if a, ok := h.agents[agentID]; ok {
			a.Alias = alias
			a.Faction = faction
			a.City = city
			for _, n := range GlobalNodes {
				if n.City == city {
					a.Coord = n.Coord
					break
				}
			}
			h.addLog("INFO", fmt.Sprintf("[%s] switched faction to %s stationed at %s", alias, faction, city), "SENTINEL")
		}
	}
}
