package main

import (
	"math"
)

const (
	DistanceD = 2.5 // Projection distance factor
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

// RotateXZ applies standard Camera Yaw
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

// RotateYZ applies standard Camera Pitch
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

// Project4Dto2D projects world coordinate to camera screen
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

	// Near clipping plane
	denomZ := d + z3d
	if denomZ <= 0.1 {
		return 0, 0, false // Behind camera
	}

	x2d := x3d / denomZ
	y2d := y3d / denomZ

	return x2d, y2d, true
}

// TransformWorldPoint applies camera view matrix (Position + Yaw + Pitch)
func TransformWorldPoint(worldPt, cameraPos Vector4, yaw, pitch float64) (float64, float64, bool) {
	rel := Vector4{
		X: worldPt.X - cameraPos.X,
		Y: worldPt.Y - cameraPos.Y,
		Z: worldPt.Z - cameraPos.Z,
		W: worldPt.W - cameraPos.W,
	}

	rYaw := RotateXZ(rel, -yaw)
	rPitch := RotateYZ(rYaw, -pitch)
	return Project4Dto2D(rPitch, DistanceD)
}

// GenerateProjectedLines builds the 4D Tesseract, 3D Cyber-Grid, Orbs, and Other Player Avatars
func GenerateProjectedLines(pos Vector4, yaw, pitch, hyperRot float64, otherPlayers []PlayerState, orbs []Orb) []Line2D {
	lines := make([]Line2D, 0, 256)

	// ==========================================
	// 1. ENVIRONMENT: CYBER-GRID FLOOR (Lantai Kotak-Kotak 3D)
	// ==========================================
	floorY := -2.0
	gridSize := 16.0
	gridStep := 2.0

	for z := -gridSize; z <= gridSize; z += gridStep {
		p1x, p1y, ok1 := TransformWorldPoint(Vector4{X: -gridSize, Y: floorY, Z: z, W: 0}, pos, yaw, pitch)
		p2x, p2y, ok2 := TransformWorldPoint(Vector4{X: gridSize, Y: floorY, Z: z, W: 0}, pos, yaw, pitch)
		if ok1 && ok2 {
			color := "#003322"
			if z == 0 {
				color = "#00FF88"
			}
			lines = append(lines, Line2D{X1: p1x, Y1: p1y, X2: p2x, Y2: p2y, Color: color})
		}
	}

	for x := -gridSize; x <= gridSize; x += gridStep {
		p1x, p1y, ok1 := TransformWorldPoint(Vector4{X: x, Y: floorY, Z: -gridSize, W: 0}, pos, yaw, pitch)
		p2x, p2y, ok2 := TransformWorldPoint(Vector4{X: x, Y: floorY, Z: gridSize, W: 0}, pos, yaw, pitch)
		if ok1 && ok2 {
			color := "#003322"
			if x == 0 {
				color = "#0088FF"
			}
			lines = append(lines, Line2D{X1: p1x, Y1: p1y, X2: p2x, Y2: p2y, Color: color})
		}
	}

	// ==========================================
	// 2. ENVIRONMENT: 4 CORNER PILLARS
	// ==========================================
	corners := [][2]float64{
		{-8, -8}, {8, -8}, {-8, 8}, {8, 8},
	}
	for _, c := range corners {
		p1x, p1y, ok1 := TransformWorldPoint(Vector4{X: c[0], Y: floorY, Z: c[1], W: 0}, pos, yaw, pitch)
		p2x, p2y, ok2 := TransformWorldPoint(Vector4{X: c[0], Y: floorY + 6.0, Z: c[1], W: 0}, pos, yaw, pitch)
		if ok1 && ok2 {
			lines = append(lines, Line2D{X1: p1x, Y1: p1y, X2: p2x, Y2: p2y, Color: "#FF0055"})
		}
	}

	// ==========================================
	// 3. CORE OBJECT: 4D TESSERACT HYPERCUBE
	// ==========================================
	var tesseractPts [16][2]float64
	var tesseractVisible [16]bool
	tesseractCenter := Vector4{X: 0, Y: 0, Z: 6.0, W: 0}

	for i, v := range BaseTesseractVertices {
		r4d := RotateXW(v, hyperRot)
		worldVertex := Vector4{
			X: r4d.X*1.2 + tesseractCenter.X,
			Y: r4d.Y*1.2 + tesseractCenter.Y,
			Z: r4d.Z*1.2 + tesseractCenter.Z,
			W: r4d.W*1.2 + tesseractCenter.W,
		}

		px, py, ok := TransformWorldPoint(worldVertex, pos, yaw, pitch)
		tesseractPts[i] = [2]float64{px, py}
		tesseractVisible[i] = ok
	}

	for _, edge := range TesseractEdges {
		if tesseractVisible[edge[0]] && tesseractVisible[edge[1]] {
			p1 := tesseractPts[edge[0]]
			p2 := tesseractPts[edge[1]]
			lines = append(lines, Line2D{
				X1:    p1[0],
				Y1:    p1[1],
				X2:    p2[0],
				Y2:    p2[1],
				Color: "#00FF00",
			})
		}
	}

	// ==========================================
	// 4. GAMEPLAY OBJECTS: QUANTUM ORBS (Bintang Kuning Emas Berlian)
	// ==========================================
	diamondOffsets := [6][3]float64{
		{0, 0.35, 0}, {0, -0.35, 0},
		{0.25, 0, 0}, {-0.25, 0, 0},
		{0, 0, 0.25}, {0, 0, -0.25},
	}
	diamondEdges := [12][2]int{
		{0, 2}, {0, 3}, {0, 4}, {0, 5},
		{1, 2}, {1, 3}, {1, 4}, {1, 5},
		{2, 4}, {4, 3}, {3, 5}, {5, 2},
	}

	for _, orb := range orbs {
		var orbPts [6][2]float64
		var orbVis [6]bool

		for i, pt := range diamondOffsets {
			worldPt := Vector4{
				X: orb.Position.X + pt[0],
				Y: orb.Position.Y + pt[1],
				Z: orb.Position.Z + pt[2],
				W: orb.Position.W,
			}
			px, py, ok := TransformWorldPoint(worldPt, pos, yaw, pitch)
			orbPts[i] = [2]float64{px, py}
			orbVis[i] = ok
		}

		for _, edge := range diamondEdges {
			if orbVis[edge[0]] && orbVis[edge[1]] {
				lines = append(lines, Line2D{
					X1:    orbPts[edge[0]][0],
					Y1:    orbPts[edge[0]][1],
					X2:    orbPts[edge[1]][0],
					Y2:    orbPts[edge[1]][1],
					Color: "#FFD700", // Gold Diamond Orb
				})
			}
		}
	}

	// ==========================================
	// 5. MULTIPLAYER: AVATAR PEMAIN LAIN (Cyan Hologram Cube)
	// ==========================================
	avatarBox := [8][3]float64{
		{-0.4, -0.4, -0.4}, {0.4, -0.4, -0.4}, {0.4, 0.4, -0.4}, {-0.4, 0.4, -0.4},
		{-0.4, -0.4, 0.4}, {0.4, -0.4, 0.4}, {0.4, 0.4, 0.4}, {-0.4, 0.4, 0.4},
	}
	avatarEdges := [12][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {7, 4}, {6, 7},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}

	for _, other := range otherPlayers {
		var box2D [8][2]float64
		var boxVis [8]bool

		for i, pt := range avatarBox {
			worldPt := Vector4{
				X: other.Position.X + pt[0],
				Y: other.Position.Y + pt[1],
				Z: other.Position.Z + pt[2],
				W: other.Position.W,
			}
			px, py, ok := TransformWorldPoint(worldPt, pos, yaw, pitch)
			box2D[i] = [2]float64{px, py}
			boxVis[i] = ok
		}

		for _, edge := range avatarEdges {
			if boxVis[edge[0]] && boxVis[edge[1]] {
				lines = append(lines, Line2D{
					X1:    box2D[edge[0]][0],
					Y1:    box2D[edge[0]][1],
					X2:    box2D[edge[1]][0],
					Y2:    box2D[edge[1]][1],
					Color: "#00FFFF", // Cyan Hologram
				})
			}
		}
	}

	return lines
}
