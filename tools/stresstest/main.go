package main

import (
	"flag"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	clients := flag.Int("clients", 100, "Number of concurrent WebSocket clients")
	duration := flag.Duration("duration", 10*time.Second, "Test duration")
	serverURL := flag.String("url", "ws://localhost:8080/ws", "WebSocket server URL")
	flag.Parse()

	u, err := url.Parse(*serverURL)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	log.Printf("🚀 Starting Stress Test: %d concurrent clients connecting to %s for %v", *clients, *serverURL, *duration)

	var (
		connectedClients int64
		messagesReceived int64
		errorsCount      int64
		wg               sync.WaitGroup
	)

	stopChan := make(chan struct{})

	// Spawn concurrent clients
	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				atomic.AddInt64(&errorsCount, 1)
				return
			}
			defer conn.Close()

			atomic.AddInt64(&connectedClients, 1)
			defer atomic.AddInt64(&connectedClients, -1)

			// Reader goroutine for incoming 2D line frames
			done := make(chan struct{})
			go func() {
				defer close(done)
				for {
					var raw map[string]interface{}
					if err := conn.ReadJSON(&raw); err != nil {
						return
					}
					atomic.AddInt64(&messagesReceived, 1)
				}
			}()

			// Sender ticker for simulated input bombardment
			inputTicker := time.NewTicker(50 * time.Millisecond)
			defer inputTicker.Stop()

			keys := []string{"w", "a", "s", "d", "q", "e", "shift", " ", "arrowleft", "arrowright"}

			for {
				select {
				case <-stopChan:
					return
				case <-done:
					return
				case t := <-inputTicker.C:
					// Send pseudo-random keys
					key := keys[t.Nanosecond()%len(keys)]
					cmd := map[string]interface{}{
						"type":  "move",
						"key":   key,
						"delta": 0.05,
					}
					if err := conn.WriteJSON(cmd); err != nil {
						atomic.AddInt64(&errorsCount, 1)
						return
					}
				}
			}
		}(i)
	}

	// Live telemetry printer
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case <-time.After(1 * time.Second):
				log.Printf("📊 Active Connections: %d | Frames Recv: %d/s | Total Errors: %d",
					atomic.LoadInt64(&connectedClients),
					atomic.SwapInt64(&messagesReceived, 0),
					atomic.LoadInt64(&errorsCount),
				)
			}
		}
	}()

	time.Sleep(*duration)
	close(stopChan)
	log.Println("🛑 Stopping stress test, waiting for clients to disconnect...")
	wg.Wait()
	log.Println("✅ Stress test completed!")
}
