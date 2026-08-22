package main

import (
	"fmt"
	"log"
)

// Hub bertindak sebagai "Mandor" tunggal (Actor Model / Channel Broker)
type Hub struct {
	// Map murni, TIDAK ADA Mutex!
	clients map[*Player]bool

	// Kotak surat untuk pemain baru yang mau masuk
	register chan *Player

	// Kotak surat untuk pemain yang mau keluar/disconnect
	unregister chan *Player

	// Kotak surat utama tempat Mandor menerima pengumuman broadcast
	Broadcast chan []byte

	// Kotak surat untuk query jumlah pemain aktif
	countRequests chan chan int
}

// NewHub menginisialisasi sang Mandor dengan kotak surat (buffered channel)
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Player]bool),
		register:      make(chan *Player, 1024),
		unregister:    make(chan *Player, 1024),
		Broadcast:     make(chan []byte, 4096),
		countRequests: make(chan chan int, 64),
	}
}

// Run adalah Jantung Actor Model: Event Loop tunggal yang memproses antrean bebas lag
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
				close(player.Send) // Tutup kotak suratnya
				fmt.Printf("[Hub] Pemain %s keluar. Total aktif: %d\n", player.ID, len(h.clients))
			}

		case message := <-h.Broadcast:
			// Mandor memfotokopi pesan ke seluruh kotak surat (RAM-to-RAM nanoseconds)
			for player := range h.clients {
				select {
				case player.Send <- message:
					// Berhasil masuk ke antrean kurir pemain

				default:
					// RAHASIA ANTI-LAG:
					// Kalau buffer pemain ini penuh (karena internet lambat/stuck),
					// Mandor TIDAK AKAN MENUNGGU!
					// Mandor langsung memutuskan pemain ini agar 9.999 pemain lain tidak lag.
					close(player.Send)
					delete(h.clients, player)
					fmt.Printf("[Hub] Pemain %s di-disconnect karena buffer jenuh (Slow Consumer Protection).\n", player.ID)
				}
			}

		case respChan := <-h.countRequests:
			respChan <- len(h.clients)
		}
	}
}

// GetPlayerCount membaca total pemain melalui jalur pesan channel
func (h *Hub) GetPlayerCount() int {
	resp := make(chan int, 1)
	h.countRequests <- resp
	return <-resp
}
