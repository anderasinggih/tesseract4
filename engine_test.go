package main

import (
	"sync"
	"testing"
	"time"
)

// TestConcurrentMovementAndPhysics tests race conditions between InputReader, Physics, and WritePump
func TestConcurrentMovementAndPhysics(t *testing.T) {
	player := &Player{
		ID: "test-player-1",
		Position: Vector4{
			X: 0, Y: 0, Z: 0, W: 1.0,
		},
		Rotation: 0,
		Send:     make(chan []byte, 256),
	}

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Drain send channel to simulate active consumer
	go func() {
		for {
			select {
			case <-done:
				return
			case <-player.Send:
			}
		}
	}()

	// Goroutine 1: Bombard state updates (Simulating InputReader writes)
	wg.Add(1)
	go func() {
		defer wg.Done()
		keys := []string{"w", "a", "s", "d", "q", "e", "shift", " ", "arrowleft", "arrowright"}
		for i := 0; i < 5000; i++ {
			select {
			case <-done:
				return
			default:
				player.Mutex.Lock()
				key := keys[i%len(keys)]
				switch key {
				case "a":
					player.Position.X -= 0.05
				case "d":
					player.Position.X += 0.05
				case "w":
					player.Position.Y += 0.05
				case "s":
					player.Position.Y -= 0.05
				case "shift":
					player.Position.W -= 0.05
				case " ":
					player.Position.W += 0.05
				case "arrowleft":
					player.Rotation -= 0.04
				case "arrowright":
					player.Rotation += 0.04
				}
				player.Mutex.Unlock()
			}
		}
	}()

	// Goroutine 2: Continuous reads & mathematical projection (Simulating Physics loop)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5000; i++ {
			select {
			case <-done:
				return
			default:
				player.Mutex.RLock()
				pos := player.Position
				rot := player.Rotation
				player.Mutex.RUnlock()

				_ = CalculateTimeMultiplier(pos.W)
				_ = GenerateProjectedLines(pos, rot)

				player.Mutex.Lock()
				player.Rotation += 0.01
				player.Mutex.Unlock()
			}
		}
	}()

	// Goroutine 3: Actor Model Hub message passing (Zero Mutex contention)
	hub := NewHub()
	go hub.Run()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			p := &Player{
				ID:   "player-temp",
				Send: make(chan []byte, 16),
			}
			hub.register <- p
			hub.Broadcast <- []byte(`{"ping": true}`)
			_ = hub.GetPlayerCount()
			hub.unregister <- p
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()
}
