package main

import (
	"sync"
	"testing"
	"time"
)

// TestWarRoomHubConcurrentAttackDefense tests high concurrency between red attackers and blue defenders
func TestWarRoomHubConcurrentAttackDefense(t *testing.T) {
	hub := NewWarRoomHub()
	go hub.Run()

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Spawn 20 simulated operators
	for i := 0; i < 20; i++ {
		agent := &AgentConnection{
			ID:      "agent-test",
			Alias:   "Operator",
			Faction: "red",
			City:    "Jakarta",
			Send:    make(chan []byte, 64),
		}
		hub.register <- agent

		// Consume outbound stream
		go func(a *AgentConnection) {
			for {
				select {
				case <-done:
					return
				case <-a.Send:
				}
			}
		}(agent)
	}

	// Goroutine 1: Rapid red team attack launches
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			select {
			case <-done:
				return
			default:
				hub.LaunchAttack("agent-red", "Jakarta", "Washington DC", "DDoS SYN Flood", 500)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Goroutine 2: Rapid blue team defense mitigations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			select {
			case <-done:
				return
			default:
				hub.DeployDefense("agent-blue", "bgp_null_route", "test-arc", "Washington DC")
				hub.DeployDefense("agent-blue", "firewall_patch", "", "Tokyo")
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Goroutine 3: Continuous snapshot readers (Simulating 30Hz WebSocket streamers)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			select {
			case <-done:
				return
			default:
				_ = hub.GetSnapshot("agent-test")
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()
}

// TestGeodesicGreatCircleMath verifies mathematical correctness of spherical slerp waypoints
func TestGeodesicGreatCircleMath(t *testing.T) {
	p1 := GeoCoord{Lat: -6.2088, Lng: 106.8456} // Jakarta
	p2 := GeoCoord{Lat: 35.6762, Lng: 139.6503} // Tokyo

	dist := CalculateGreatCircleDistance(p1, p2)
	if dist < 5000 || dist > 6500 {
		t.Fatalf("Unexpected Great-Circle distance Jakarta-Tokyo: %f km", dist)
	}

	mid := CalculateGeodesicWaypoint(p1, p2, 0.5)
	if mid.Lat == 0 && mid.Lng == 0 {
		t.Fatalf("Failed to calculate mid-waypoint")
	}

	// Shannon entropy test
	plain := []byte("AAAAABBBBBCCCCCDDDDD")
	plainEntropy := CalculateShannonEntropy(plain)

	randomBytes := GenerateRandomPayloadBytes(256, true)
	randomEntropy := CalculateShannonEntropy(randomBytes)

	if randomEntropy <= plainEntropy {
		t.Fatalf("Random encrypted payload should have higher entropy than patterned plain text")
	}
}
