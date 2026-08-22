package main

import (
	"math"
)

// GeoCoord represents a geographic coordinate point (Latitude, Longitude)
type GeoCoord struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Airport represents an international aviation hub
type Airport struct {
	ICAO    string   `json:"icao"`
	IATA    string   `json:"iata"`
	Name    string   `json:"name"`
	City    string   `json:"city"`
	Country string   `json:"country"`
	Coord   GeoCoord `json:"coord"`
}

// GlobalMajorAirports international hubs (Worldwide International Grid)
var GlobalMajorAirports = []Airport{
	// Southeast Asia & Australia
	{"WIII", "CGK", "Soekarno-Hatta", "Jakarta", "ID", GeoCoord{-6.1256, 106.6558}},
	{"WSSS", "SIN", "Singapore Changi", "Singapore", "SG", GeoCoord{1.3644, 103.9915}},
	{"WMKK", "KUL", "Kuala Lumpur Intl", "Kuala Lumpur", "MY", GeoCoord{2.7456, 101.7072}},
	{"VTBS", "BKK", "Suvarnabhumi", "Bangkok", "TH", GeoCoord{13.6900, 100.7501}},
	{"WADD", "DPS", "Ngurah Rai Bali", "Denpasar", "ID", GeoCoord{-8.7482, 115.1672}},
	{"YSSY", "SYD", "Sydney Kingsford", "Sydney", "AU", GeoCoord{-33.9461, 151.1772}},
	{"YMML", "MEL", "Melbourne Intl", "Melbourne", "AU", GeoCoord{-37.6690, 144.8410}},
	
	// East Asia & Pacific
	{"RJTT", "HND", "Tokyo Haneda", "Tokyo", "JP", GeoCoord{35.5494, 139.7798}},
	{"RJAA", "NRT", "Tokyo Narita", "Tokyo", "JP", GeoCoord{35.7720, 140.3929}},
	{"RKSI", "ICN", "Incheon Intl", "Seoul", "KR", GeoCoord{37.4602, 126.4407}},
	{"VHHH", "HKG", "Hong Kong Intl", "Hong Kong", "HK", GeoCoord{22.3080, 113.9185}},
	{"ZBAA", "PEK", "Beijing Capital", "Beijing", "CN", GeoCoord{40.0799, 116.6031}},
	{"ZSPD", "PVG", "Shanghai Pudong", "Shanghai", "CN", GeoCoord{31.1443, 121.8083}},
	
	// Middle East & South Asia
	{"OMDB", "DXB", "Dubai Intl", "Dubai", "AE", GeoCoord{25.2532, 55.3657}},
	{"OTHH", "DOH", "Hamad Intl", "Doha", "QA", GeoCoord{25.2731, 51.6081}},
	{"VIDP", "DEL", "Indira Gandhi", "Delhi", "IN", GeoCoord{28.5562, 77.1000}},
	{"VABB", "BOM", "Chhatrapati Shivaji", "Mumbai", "IN", GeoCoord{19.0896, 72.8656}},
	
	// Europe
	{"EGLL", "LHR", "London Heathrow", "London", "GB", GeoCoord{51.4700, -0.4543}},
	{"LFPG", "CDG", "Paris Charles de Gaulle", "Paris", "FR", GeoCoord{49.0097, 2.5479}},
	{"EDDF", "FRA", "Frankfurt Main", "Frankfurt", "DE", GeoCoord{50.0379, 8.5622}},
	{"EHAM", "AMS", "Amsterdam Schiphol", "Amsterdam", "NL", GeoCoord{52.3105, 4.7683}},
	{"LEMD", "MAD", "Madrid Barajas", "Madrid", "ES", GeoCoord{40.4839, -3.5680}},
	{"LTFM", "IST", "Istanbul Airport", "Istanbul", "TR", GeoCoord{41.2753, 28.7519}},

	// Americas
	{"KJFK", "JFK", "John F. Kennedy", "New York", "US", GeoCoord{40.6413, -73.7781}},
	{"KLAX", "LAX", "Los Angeles Intl", "Los Angeles", "US", GeoCoord{33.9416, -118.4085}},
	{"KORD", "ORD", "Chicago O'Hare", "Chicago", "US", GeoCoord{41.9742, -87.9073}},
	{"KATL", "ATL", "Hartsfield-Jackson", "Atlanta", "US", GeoCoord{33.6407, -84.4277}},
	{"SBGR", "GRU", "São Paulo Guarulhos", "São Paulo", "BR", GeoCoord{-23.4356, -46.4731}},
}

// AirspaceSector defines a Flight Information Region (FIR) controlled by a player
type AirspaceSector struct {
	ID          string     `json:"id"`
	ICAO        string     `json:"icao"`
	Name        string     `json:"name"`
	Controller  string     `json:"controller"` // Player Alias or "AUTO-TOWER"
	ControllerID string    `json:"controllerId,omitempty"`
	Center      GeoCoord   `json:"center"`
	Bounds      []GeoCoord `json:"bounds"` // Polygon boundaries
	Color       string     `json:"color"`
}

// GlobalAirspaceSectors list of playable international sectors
var GlobalAirspaceSectors = []AirspaceSector{
	{
		ID:         "sec-wiii",
		ICAO:       "WIII_FIR",
		Name:       "Jakarta FIR (Indonesia)",
		Center:     GeoCoord{-6.12, 106.65},
		Color:      "#00e676",
		Bounds: []GeoCoord{
			{6.0, 95.0}, {6.0, 110.0}, {-2.0, 112.0}, {-11.0, 112.0}, {-11.0, 95.0}, {6.0, 95.0},
		},
	},
	{
		ID:         "sec-wsjc",
		ICAO:       "WSJC_ACC",
		Name:       "Singapore ACC (Southeast Asia)",
		Center:     GeoCoord{1.36, 103.99},
		Color:      "#00b4d8",
		Bounds: []GeoCoord{
			{8.0, 99.0}, {8.0, 108.0}, {0.0, 108.0}, {0.0, 99.0}, {8.0, 99.0},
		},
	},
	{
		ID:         "sec-rjjj",
		ICAO:       "RJJJ_ACC",
		Name:       "Tokyo ACC (East Asia/Pacific)",
		Center:     GeoCoord{35.55, 139.78},
		Color:      "#ffb703",
		Bounds: []GeoCoord{
			{45.0, 125.0}, {45.0, 150.0}, {25.0, 150.0}, {25.0, 125.0}, {45.0, 125.0},
		},
	},
	{
		ID:         "sec-egtt",
		ICAO:       "EGTT_ACC",
		Name:       "London / Eurocontrol ACC",
		Center:     GeoCoord{51.47, -0.45},
		Color:      "#b026ff",
		Bounds: []GeoCoord{
			{60.0, -10.0}, {60.0, 15.0}, {40.0, 15.0}, {40.0, -10.0}, {60.0, -10.0},
		},
	},
	{
		ID:         "sec-kzny",
		ICAO:       "KZNY_ACC",
		Name:       "New York ARTCC (North America)",
		Center:     GeoCoord{40.64, -73.78},
		Color:      "#ff2a5f",
		Bounds: []GeoCoord{
			{50.0, -90.0}, {50.0, -60.0}, {30.0, -60.0}, {30.0, -90.0}, {50.0, -90.0},
		},
	},
	{
		ID:         "sec-ybbb",
		ICAO:       "YBBB_FIR",
		Name:       "Sydney / Brisbane FIR (Oceania)",
		Center:     GeoCoord{-33.94, 151.17},
		Color:      "#00f5d4",
		Bounds: []GeoCoord{
			{-10.0, 140.0}, {-10.0, 160.0}, {-45.0, 160.0}, {-45.0, 140.0}, {-10.0, 140.0},
		},
	},
}

// DistanceInNauticalMiles computes spherical distance in Nautical Miles (1 NM = 1.852 km)
func DistanceInNauticalMiles(p1, p2 GeoCoord) float64 {
	const EarthRadiusNM = 3440.065
	lat1 := p1.Lat * math.Pi / 180.0
	lng1 := p1.Lng * math.Pi / 180.0
	lat2 := p2.Lat * math.Pi / 180.0
	lng2 := p2.Lng * math.Pi / 180.0

	dLat := lat2 - lat1
	dLng := lng2 - lng1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusNM * c
}

// CalculateTrueBearing computes the compass heading (0 to 360 deg) from p1 to p2
func CalculateTrueBearing(p1, p2 GeoCoord) float64 {
	lat1 := p1.Lat * math.Pi / 180.0
	lat2 := p2.Lat * math.Pi / 180.0
	dLng := (p2.Lng - p1.Lng) * math.Pi / 180.0

	y := math.Sin(dLng) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLng)

	bearing := math.Atan2(y, x) * 180.0 / math.Pi
	if bearing < 0 {
		bearing += 360.0
	}
	return bearing
}

// MoveAircraftPosition calculates next position along current heading at given ground speed (knots)
func MoveAircraftPosition(curr GeoCoord, headingDeg float64, speedKnots float64, dtSeconds float64) GeoCoord {
	const EarthRadiusNM = 3440.065
	distNM := speedKnots * (dtSeconds / 3600.0)
	angularDist := distNM / EarthRadiusNM

	lat1 := curr.Lat * math.Pi / 180.0
	lng1 := curr.Lng * math.Pi / 180.0
	brng := headingDeg * math.Pi / 180.0

	lat2 := math.Asin(math.Sin(lat1)*math.Cos(angularDist) +
		math.Cos(lat1)*math.Sin(angularDist)*math.Cos(brng))
	lng2 := lng1 + math.Atan2(math.Sin(brng)*math.Sin(angularDist)*math.Cos(lat1),
		math.Cos(angularDist)-math.Sin(lat1)*math.Sin(lat2))

	return GeoCoord{
		Lat: lat2 * 180.0 / math.Pi,
		Lng: lng2 * 180.0 / math.Pi,
	}
}

// IsPointInSector tests if an aircraft coordinate is inside a polygon sector (Ray-Casting Algorithm)
func IsPointInSector(pt GeoCoord, polygon []GeoCoord) bool {
	inside := false
	n := len(polygon)
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := polygon[i].Lng, polygon[i].Lat
		xj, yj := polygon[j].Lng, polygon[j].Lat

		intersect := ((yi > pt.Lat) != (yj > pt.Lat)) &&
			(pt.Lng < (xj-xi)*(pt.Lat-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}
