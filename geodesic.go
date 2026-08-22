package main

import (
	"math"
	"math/rand"
)

// GeoCoord represents a geographic coordinate point (Latitude, Longitude)
type GeoCoord struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Major global cybersecurity hubs and infrastructure targets
var GlobalNodes = []struct {
	City    string
	Country string
	Coord   GeoCoord
	IP      string
	ASN     string
}{
	{"Jakarta", "ID", GeoCoord{-6.2088, 106.8456}, "103.119.214.1", "AS131440"},
	{"Singapore", "SG", GeoCoord{1.3521, 103.8198}, "165.21.100.2", "AS4657"},
	{"Tokyo", "JP", GeoCoord{35.6762, 139.6503}, "133.242.18.5", "AS2514"},
	{"Frankfurt", "DE", GeoCoord{50.1109, 8.6821}, "194.109.6.99", "AS8881"},
	{"Washington DC", "US", GeoCoord{38.9072, -77.0369}, "199.19.0.1", "AS396982"},
	{"London", "GB", GeoCoord{51.5074, -0.1278}, "195.66.224.1", "AS5459"},
	{"Sydney", "AU", GeoCoord{-33.8688, 151.2093}, "139.130.4.5", "AS1221"},
	{"São Paulo", "BR", GeoCoord{-23.5505, -46.6333}, "200.160.2.3", "AS22548"},
	{"Seoul", "KR", GeoCoord{37.5665, 126.9780}, "211.233.0.1", "AS9318"},
	{"Zurich", "CH", GeoCoord{47.3769, 8.5417}, "194.42.48.1", "AS3303"},
	{"Hong Kong", "HK", GeoCoord{22.3193, 114.1694}, "202.45.84.1", "AS9381"},
	{"San Francisco", "US", GeoCoord{37.7749, -122.4194}, "198.41.0.4", "AS13335"},
}

// toRad converts degrees to radians
func toRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// toDeg converts radians to degrees
func toDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// CalculateGreatCircleDistance computes the spherical distance in km between two geo-coordinates
func CalculateGreatCircleDistance(p1, p2 GeoCoord) float64 {
	const EarthRadiusKm = 6371.0
	dLat := toRad(p2.Lat - p1.Lat)
	dLng := toRad(p2.Lng - p1.Lng)

	lat1 := toRad(p1.Lat)
	lat2 := toRad(p2.Lat)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLng/2)*math.Sin(dLng/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

// CalculateGeodesicWaypoint interpolates an exact flight coordinate at fraction t (0.0 to 1.0) along Great-Circle
// Mathematical formula: Arc(t) using Spherical Linear Interpolation (Slerp) on unit sphere
func CalculateGeodesicWaypoint(p1, p2 GeoCoord, t float64) GeoCoord {
	if t <= 0.0 {
		return p1
	}
	if t >= 1.0 {
		return p2
	}

	lat1 := toRad(p1.Lat)
	lng1 := toRad(p1.Lng)
	lat2 := toRad(p2.Lat)
	lng2 := toRad(p2.Lng)

	// Angular distance d between points
	cosD := math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(lng2-lng1)
	if cosD > 1.0 {
		cosD = 1.0
	} else if cosD < -1.0 {
		cosD = -1.0
	}
	d := math.Acos(cosD)

	if d < 0.0001 {
		return p1
	}

	sinD := math.Sin(d)
	a := math.Sin((1.0-t)*d) / sinD
	b := math.Sin(t*d) / sinD

	x := a*math.Cos(lat1)*math.Cos(lng1) + b*math.Cos(lat2)*math.Cos(lng2)
	y := a*math.Cos(lat1)*math.Sin(lng1) + b*math.Cos(lat2)*math.Sin(lng2)
	z := a*math.Sin(lat1) + b*math.Sin(lat2)

	latFinal := math.Atan2(z, math.Sqrt(x*x+y*y))
	lngFinal := math.Atan2(y, x)

	return GeoCoord{
		Lat: toDeg(latFinal),
		Lng: toDeg(lngFinal),
	}
}

// CalculateShannonEntropy measures cryptographic randomness of byte traffic (0.0 = plain, 8.0 = encrypted)
func CalculateShannonEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	var freq [256]float64
	for _, b := range data {
		freq[b]++
	}

	total := float64(len(data))
	var entropy float64
	for _, count := range freq {
		if count > 0 {
			p := count / total
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// GenerateRandomPayloadBytes generates packet samples with given entropy profile
func GenerateRandomPayloadBytes(size int, isEncrypted bool) []byte {
	buf := make([]byte, size)
	if isEncrypted {
		for i := 0; i < size; i++ {
			buf[i] = byte(rand.Intn(256))
		}
	} else {
		// Patterned text payload
		charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789{}[]:\"_-"
		for i := 0; i < size; i++ {
			buf[i] = charset[rand.Intn(len(charset))]
		}
	}
	return buf
}
