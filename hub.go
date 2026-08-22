package main

import (
	"fmt"
	"log"
)

// Hub bertindak sebagai "Mandor" tunggal (Actor Model / Channel Broker)
type Hub struct {
	clients       map[*Player]bool
	register      chan *Player
	unregister    chan *Player
	Broadcast     chan []byte
	countRequests chan chan int
	getPlayers    chan chan []*Player
}

// NewHub menginisialisasi sang Mandor dengan kotak surat
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Player]bool),
		register:      make(chan *Player, 1024),
		unregister:    make(chan *Player, 1024),
		Broadcast:     make(chan []byte, 4096),
		countRequests: make(chan chan int, 64),
		getPlayers:    make(chan chan []*Player, 64),
	}
}

// Run adalah Jantung Actor Model: Event Loop tunggal
func (h *Hub) Run() {
	log.Println("👔 [Hub Mandor] Event Loop Actor Model + Non-Blocking Broadcast Pump aktif.")
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

// GetOtherPlayersSnapshot mengambil list pemain lain untuk digambar di dunia 3D/4D
func (h *Hub) GetOtherPlayersSnapshot(selfID string) []PlayerState {
	resp := make(chan []*Player, 1)
	h.getPlayers <- resp
	players := <-resp

	states := make([]PlayerState, 0, len(players))
	for _, p := range players {
		if p.ID == selfID {
			continue
		}
		p.Mutex.RLock()
		states = append(states, PlayerState{
			ID:       p.ID,
			Position: p.Position,
			Yaw:      p.Yaw,
			Pitch:    p.Pitch,
		})
		p.Mutex.RUnlock()
	}
	return states
}

// GetPlayerCount membaca total pemain melalui jalur pesan channel
func (h *Hub) GetPlayerCount() int {
	resp := make(chan int, 1)
	h.countRequests <- resp
	return <-resp
}
