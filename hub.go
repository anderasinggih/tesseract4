package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
)

// Hub bertindak sebagai "Mandor" tunggal (Actor Model / Channel Broker)
type Hub struct {
	clients       map[*Player]bool
	register      chan *Player
	unregister    chan *Player
	Broadcast     chan []byte
	countRequests chan chan int
	getPlayers    chan chan []*Player
	collectOrb    chan *Player
	orbs          []Orb
}

// NewHub menginisialisasi sang Mandor dengan kotak surat dan Quantum Orbs
func NewHub() *Hub {
	h := &Hub{
		clients:       make(map[*Player]bool),
		register:      make(chan *Player, 1024),
		unregister:    make(chan *Player, 1024),
		Broadcast:     make(chan []byte, 4096),
		countRequests: make(chan chan int, 64),
		getPlayers:    make(chan chan []*Player, 64),
		collectOrb:    make(chan *Player, 256),
		orbs:          make([]Orb, 6),
	}

	// Spawn 6 Quantum Orbs di penjuru arena dunia 3D/4D
	h.respawnOrbs()
	return h
}

func (h *Hub) respawnOrbs() {
	for i := 0; i < len(h.orbs); i++ {
		h.orbs[i] = Orb{
			ID: i,
			Position: Vector4{
				X: (rand.Float64()*12.0 - 6.0),
				Y: (rand.Float64()*3.0 - 1.0),
				Z: (rand.Float64()*12.0 - 3.0),
				W: (rand.Float64()*1.6 - 0.8),
			},
			Active: true,
		}
	}
}

// Run adalah Jantung Actor Model: Event Loop tunggal
func (h *Hub) Run() {
	log.Println("👔 [Hub Mandor] Event Loop Actor Model + Quantum Orb Gameplay aktif.")
	for {
		select {
		case player := <-h.register:
			h.clients[player] = true
			fmt.Printf("[Hub] Pemain %s masuk! Total aktif: %d\n", player.ID, len(h.clients))

		case player := <-h.unregister:
			if _, ok := h.clients[player]; ok {
				delete(h.clients, player)
				close(player.Send)
				fmt.Printf("[Hub] Pemain %s keluar. Total aktif: %d\n", player.ID, len(h.clients))
			}

		case p := <-h.collectOrb:
			// Cek apakah player menyentuh Orb (Collision radius 1.2 unit)
			p.Mutex.Lock()
			for i := range h.orbs {
				if !h.orbs[i].Active {
					continue
				}
				dx := p.Position.X - h.orbs[i].Position.X
				dy := p.Position.Y - h.orbs[i].Position.Y
				dz := p.Position.Z - h.orbs[i].Position.Z
				dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
				if dist < 1.4 {
					p.Score += 100
					// Respawn orb di lokasi baru
					h.orbs[i].Position = Vector4{
						X: (rand.Float64()*14.0 - 7.0),
						Y: (rand.Float64()*3.5 - 1.5),
						Z: (rand.Float64()*14.0 - 4.0),
						W: (rand.Float64()*1.6 - 0.8),
					}
					fmt.Printf("⭐ [Score] Player %s mengambil Quantum Core! Score: %d\n", p.ID, p.Score)
				}
			}
			p.Mutex.Unlock()

		case message := <-h.Broadcast:
			for player := range h.clients {
				select {
				case player.Send <- message:
				default:
					close(player.Send)
					delete(h.clients, player)
				}
			}

		case respChan := <-h.countRequests:
			respChan <- len(h.clients)

		case respChan := <-h.getPlayers:
			list := make([]*Player, 0, len(h.clients))
			for p := range h.clients {
				list = append(list, p)
			}
			respChan <- list
		}
	}
}

// GetGameSnapshot mengambil list pemain lain, leaderboard, dan Orbs aktif
func (h *Hub) GetGameSnapshot(selfID string) ([]PlayerState, []PlayerState, []Orb) {
	resp := make(chan []*Player, 1)
	h.getPlayers <- resp
	players := <-resp

	allStates := make([]PlayerState, 0, len(players))
	otherStates := make([]PlayerState, 0, len(players))

	for _, p := range players {
		p.Mutex.RLock()
		st := PlayerState{
			ID:       p.ID,
			Position: p.Position,
			Yaw:      p.Yaw,
			Pitch:    p.Pitch,
			Score:    p.Score,
		}
		p.Mutex.RUnlock()

		allStates = append(allStates, st)
		if p.ID != selfID {
			otherStates = append(otherStates, st)
		}
	}

	// Sort leaderboard descending by score
	sort.Slice(allStates, func(i, j int) bool {
		return allStates[i].Score > allStates[j].Score
	})

	orbsCopy := make([]Orb, len(h.orbs))
	copy(orbsCopy, h.orbs)

	return otherStates, allStates, orbsCopy
}

// CheckCollectOrb mengirim sinyal collision check ke Mandor
func (h *Hub) CheckCollectOrb(p *Player) {
	select {
	case h.collectOrb <- p:
	default:
	}
}

// GetPlayerCount membaca total pemain melalui jalur pesan channel
func (h *Hub) GetPlayerCount() int {
	resp := make(chan int, 1)
	h.countRequests <- resp
	return <-resp
}
