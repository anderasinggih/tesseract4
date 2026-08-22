package main

import (
	"math"
)

const (
	DistanceD = 2.0 // Projection distance factor
)

// Base unit hypercube vertices (16 vertices for a 4D tesseract centered at origin)
var BaseTesseractVertices = [16]Vector4{
	{-1, -1, -1, -1},
	{1, -1, -1, -1},
	{1, 1, -1, -1},
	{-1, 1, -1, -1},
	{-1, -1, 1, -1},
	{1, -1, 1, -1},
	{1, 1, 1, -1},
	{-1, 1, 1, -1},
	{-1, -1, -1, 1},
	{1, -1, -1, 1},
	{1, 1, -1, 1},
	{-1, 1, -1, 1},
	{-1, -1, 1, 1},
	{1, -1, 1, 1},
	{1, 1, 1, 1},
	{-1, 1, 1, 1},
}

// 32 edges of the tesseract connecting indices of BaseTesseractVertices
var TesseractEdges = [32][2]int{
	// Inner cube 1 (w = -1, z = -1)
	{0, 1}, {1, 2}, {2, 3}, {3, 0},
	// Inner cube 1 (w = -1, z = 1)
	{4, 5}, {5, 6}, {6, 7}, {7, 4},
	// Connecting z on w = -1
	{0, 4}, {1, 5}, {2, 6}, {3, 7},

	// Outer cube 2 (w = 1, z = -1)
	{8, 9}, {9, 10}, {10, 11}, {11, 8},
	// Outer cube 2 (w = 1, z = 1)
	{12, 13}, {13, 14}, {14, 15}, {15, 12},
	// Connecting z on w = 1
	{8, 12}, {9, 13}, {10, 14}, {11, 15},

	// Connecting the two 3D cubes across w-dimension
	{0, 8}, {1, 9}, {2, 10}, {3, 11},
	{4, 12}, {5, 13}, {6, 14}, {7, 15},
}

// RotateXZ applies standard Camera Yaw (Kiri/Kanan)
func RotateXZ(v Vector4, yaw float64) Vector4 {
	cosT := math.Cos(yaw)
	sinT := math.Sin(yaw)
	return Vector4{
		X: v.X*cosT - v.Z*sinT,
		Y: v.Y,
		Z: v.X*sinT + v.Z*cosT,
		W: v.W,
	}
}

// RotateYZ applies standard Camera Pitch (Atas/Bawah)
func RotateYZ(v Vector4, pitch float64) Vector4 {
	cosT := math.Cos(pitch)
	sinT := math.Sin(pitch)
	return Vector4{
		X: v.X,
		Y: v.Y*cosT - v.Z*sinT,
		Z: v.Y*sinT + v.Z*cosT,
		W: v.W,
	}
}

// RotateXW applies 4D rotation along the X-W hyper-plane
func RotateXW(v Vector4, theta float64) Vector4 {
	cosT := math.Cos(theta)
	sinT := math.Sin(theta)
	return Vector4{
		X: v.X*cosT - v.W*sinT,
		Y: v.Y,
		Z: v.Z,
		W: v.X*sinT + v.W*cosT,
	}
}

// Project4Dto2D projects relative 4D camera coordinate to 2D screen
// World Point -> Camera Space (Offset Pos) -> Camera Rotation -> Perspective Projection
func Project4Dto2D(v Vector4, d float64) (float64, float64, bool) {
	denomW := d - v.W
	if math.Abs(denomW) < 0.0001 {
		if denomW < 0 {
			denomW = -0.0001
		} else {
			denomW = 0.0001
		}
	}

	x3d := v.X / denomW
	y3d := v.Y / denomW
	z3d := v.Z / denomW

	// Standard perspective division with camera forward distance
	denomZ := d + z3d
	if math.Abs(denomZ) < 0.0001 {
		if denomZ < 0 {
			denomZ = -0.0001
		} else {
			denomZ = 0.0001
		}
	}

	x2d := x3d / denomZ
	y2d := y3d / denomZ

	return x2d, y2d, true
}

// GenerateProjectedLines transforms vertices relative to player (FPS Camera View Matrix)
func GenerateProjectedLines(pos Vector4, yaw, pitch, hyperRot float64) []Line2D {
	var projectedPoints [16][2]float64

	for i, v := range BaseTesseractVertices {
		// 1. 4D Hyper-dimension rotation on base shape
		r4d := RotateXW(v, hyperRot)

		// 2. Translate world relative to camera position (Camera Eye: pos)
		rel := Vector4{
			X: r4d.X - pos.X,
			Y: r4d.Y - pos.Y,
			Z: r4d.Z - pos.Z,
			W: r4d.W - pos.W,
		}

		// 3. View Matrix: Camera Rotation (Yaw & Pitch)
		rYaw := RotateXZ(rel, -yaw)
		rPitch := RotateYZ(rYaw, -pitch)

		// 4. Projection to 2D Screen
		x2d, y2d, _ := Project4Dto2D(rPitch, DistanceD)
		projectedPoints[i] = [2]float64{x2d, y2d}
	}

	lines := make([]Line2D, len(TesseractEdges))
	for i, edge := range TesseractEdges {
		p1 := projectedPoints[edge[0]]
		p2 := projectedPoints[edge[1]]
		lines[i] = Line2D{
			X1: p1[0],
			Y1: p1[1],
			X2: p2[0],
			Y2: p2[1],
		}
	}

	return lines
}
