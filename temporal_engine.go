package main

import (
	"math"
)

const (
	Rs       = 0.5 // Schwarzschild radius
	BaseTick = 16  // Base tick rate in milliseconds (~60 FPS)
)

// CalculateTimeMultiplier computes the Schwarzschild temporal factor based on W coordinate.
// Multiplier = sqrt(|1 - Rs / (|w| + 0.1)|)
func CalculateTimeMultiplier(w float64) float64 {
	absW := math.Abs(w)
	val := math.Abs(1.0 - (Rs / (absW + 0.1)))
	multiplier := math.Sqrt(val)

	// Clamp to bounds to prevent complete freezing or infinite acceleration
	if multiplier < 0.05 {
		multiplier = 0.05
	}
	if multiplier > 20.0 {
		multiplier = 20.0
	}
	return multiplier
}
