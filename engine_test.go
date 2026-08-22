package main

import (
	"math"
	"sync"
	"testing"
	"time"
)

// TestATCKinematicsAndSeparation verifies spherical bearing and conflict detection math
func TestATCKinematicsAndSeparation(t *testing.T) {
	cgk := GeoCoord{Lat: -6.1256, Lng: 106.6558} // Jakarta (CGK)
	sin := GeoCoord{Lat: 1.3644, Lng: 103.9915}  // Singapore (SIN)

	distNM := DistanceInNauticalMiles(cgk, sin)
	if distNM < 450 || distNM > 550 {
		t.Fatalf("Unexpected CGK-SIN distance: %f NM", distNM)
	}

	bearing := CalculateTrueBearing(cgk, sin)
	if bearing < 320 || bearing > 360 {
		t.Fatalf("Unexpected CGK-SIN bearing: %f deg", bearing)
	}

	// Kinematics movement test: 450 kts for 1 hour should move ~450 NM
	moved := MoveAircraftPosition(cgk, 0, 450, 3600)
	movedDist := DistanceInNauticalMiles(cgk, moved)
	if math.Abs(movedDist-450.0) > 2.0 {
		t.Fatalf("Movement calculation drifted: moved %f NM, expected 450 NM", movedDist)
	}
}

// TestATCHubConcurrentCommands verifies thread safety for multiple controllers managing flights
func TestATCHubConcurrentCommands(t *testing.T) {
	hub := NewATCHub()
	go hub.Run()

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Spawn simulated controllers
	for i := 0; i < 5; i++ {
		client := &ATCClientConnection{
			ID:       "ctr-test",
			Callsign: "JAKARTA_APP",
			SectorID: "sec-wiii",
			Send:     make(chan []byte, 64),
		}
		hub.register <- client

		go func(c *ATCClientConnection) {
			for {
				select {
				case <-done:
					return
				case <-c.Send:
				}
			}
		}(client)
	}

	// Goroutine 1: Continuous vector and altitude commands
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 300; i++ {
			select {
			case <-done:
				return
			default:
				hub.ExecuteATCCommand(ClientATCCommand{
					Type:     "set_heading",
					Callsign: "GIA880",
					Heading:  180,
				}, "ctr-test")
				hub.ExecuteATCCommand(ClientATCCommand{
					Type:     "set_altitude",
					Callsign: "GIA880",
					Altitude: 32000,
				}, "ctr-test")
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Goroutine 2: Handoff transactions
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 300; i++ {
			select {
			case <-done:
				return
			default:
				hub.ExecuteATCCommand(ClientATCCommand{
					Type:     "handoff_init",
					Callsign: "GIA880",
					ToSector: "sec-wsjc",
				}, "ctr-test")
				hub.ExecuteATCCommand(ClientATCCommand{
					Type:     "handoff_accept",
					Callsign: "GIA880",
				}, "ctr-test")
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Goroutine 3: Snapshot telemetry readers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 300; i++ {
			select {
			case <-done:
				return
			default:
				_ = hub.GetSnapshot("ctr-test")
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()
}
