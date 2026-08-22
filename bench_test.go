package main

import (

	"testing"
	"time"
)

func TestKinematicsTickCost(t *testing.T) {
	hub := NewATCHub()
	go hub.Run()
	time.Sleep(200 * time.Millisecond)

	// Force-spawn up to 150 aircraft like production
	for i := 0; i < 150; i++ {
		hub.mu.Lock()
		if len(hub.aircraft) < 150 {
			hub.spawnCommercialFlight()
		}
		hub.mu.Unlock()
	}
	hub.mu.RLock()
	n := len(hub.aircraft)
	hub.mu.RUnlock()

	// Measure raw updateKinematics cost
	start := time.Now()
	runs := 300
	for i := 0; i < runs; i++ {
		hub.updateKinematics(0.016)
	}
	elapsed := time.Since(start)

	// Also measure GetSnapshot cost (telemetry path)
	start2 := time.Now()
	for i := 0; i < 100; i++ {
		_ = hub.GetSnapshot("bench")
	}
	snapElapsed := time.Since(start2)

	t.Logf("aircraft=%d | avg updateKinematics=%v/tick | avg GetSnapshot=%v", n, elapsed/time.Duration(runs), snapElapsed/100)
	if avg := elapsed / time.Duration(runs); avg > 8*time.Millisecond {
		t.Fatalf("physics tick too slow: %v per tick (budget 16ms)", avg)
	}
}

func TestCoordinatedTurnRealtime(t *testing.T) {
	hub := NewATCHub()
	go hub.Run()
	time.Sleep(100 * time.Millisecond)

	hub.mu.Lock()
	hub.spawnCommercialFlight()
	var ac *Aircraft
	for _, a := range hub.aircraft {
		ac = a
		break
	}
	_ = ac.Callsign
	ac.Speed = 450
	ac.TargetSpeed = 450
	ac.Roll = 0
	ac.TargetRoll = -25 // simulate stick held left
	hub.mu.Unlock()

	start := time.Now()
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for time.Since(start) < 4*time.Second {
		<-ticker.C
		i++
		hub.updateKinematics(0.016)
	}
	hub.mu.RLock()
	roll := ac.Roll
	hdg := ac.Heading
	hub.mu.RUnlock()

	simSeconds := float64(i) * 0.016
	expectedRoll := -25 * (simSeconds - 3.57) // reaches full bank after 25/7 s
	if simSeconds > 25.0/7.0 && roll > -24.5 {
		t.Errorf("roll should saturate at -25 by now, got %.2f", roll)
	}
	t.Logf("sim_time=%.2fs wall=%.2fs roll=%.2f hdgΔ=%.2f° expectedROT=%.3f°/s",
		simSeconds, time.Since(start).Seconds(), roll, hdg, 1091.0*0.4663/450)
	_ = expectedRoll
}
