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

// GlobalMajorAirports international hubs
var GlobalMajorAirports = []Airport{
	{"WIII", "CGK", "Soekarno-Hatta Intl", "Jakarta", "ID", GeoCoord{-6.1256, 106.6558}},
	{"WARR", "SUB", "Juanda Intl", "Surabaya", "ID", GeoCoord{-7.3798, 112.7875}},
	{"WADD", "DPS", "I Gusti Ngurah Rai Intl", "Bali", "ID", GeoCoord{-8.7482, 115.1672}},
	{"WIMM", "KNO", "Kualanamu Intl", "Medan", "ID", GeoCoord{3.6422, 98.8853}},
	{"WAAA", "UPG", "Sultan Hasanuddin Intl", "Makassar", "ID", GeoCoord{-5.0616, 119.5540}},
	{"WIPP", "PLM", "Sultan Mahmud Badaruddin II", "Palembang", "ID", GeoCoord{-2.8984, 104.7005}},
	{"WIOO", "PNK", "Supadio Intl", "Pontianak", "ID", GeoCoord{-0.1506, 109.4039}},
	{"WALL", "BPN", "Sultan Aji Muhammad Sulaiman", "Balikpapan", "ID", GeoCoord{-1.2683, 116.8944}},
	{"WAMM", "MDC", "Sam Ratulangi Intl", "Manado", "ID", GeoCoord{1.5494, 124.9261}},
	{"WAJJ", "DJJ", "Sentani Intl", "Jayapura", "ID", GeoCoord{-2.5769, 140.5164}},
	{"WIBB", "PKU", "Sultan Syarif Kasim II", "Pekanbaru", "ID", GeoCoord{0.4608, 101.4447}},
	{"WIEE", "PDG", "Minangkabau Intl", "Padang", "ID", GeoCoord{-0.7872, 100.2814}},
	{"WICC", "BDO", "Husein Sastranegara", "Bandung", "ID", GeoCoord{-6.9006, 107.5761}},
	{"WAHH", "YIA", "Yogyakarta Intl", "Yogyakarta", "ID", GeoCoord{-7.9072, 110.0594}},
	{"WAHQ", "SRG", "Jenderal Ahmad Yani", "Semarang", "ID", GeoCoord{-6.9722, 110.3756}},
	{"WRLS", "TRK", "Juwata Intl", "Tarakan", "ID", GeoCoord{3.3264, 117.5667}},
	{"WAKK", "MKQ", "Mopah Intl", "Merauke", "ID", GeoCoord{-8.5203, 140.4183}},
	{"WASS", "SOQ", "Domine Eduard Osok", "Sorong", "ID", GeoCoord{-0.8986, 131.2872}},
	{"WATG", "KOE", "El Tari Intl", "Kupang", "ID", GeoCoord{-10.1714, 123.6711}},
	{"WAPP", "AMQ", "Pattimura Intl", "Ambon", "ID", GeoCoord{-3.7103, 128.0894}},
	{"WSSL", "XSP", "Seletar Airport", "Singapore", "SG", GeoCoord{1.4169, 103.8678}},
	{"WMKP", "PEN", "Penang Intl", "Penang", "MY", GeoCoord{5.2971, 100.2769}},
	{"WMSA", "SZB", "Sultan Abdul Aziz Shah", "Subang", "MY", GeoCoord{3.1306, 101.5497}},
	{"WBKK", "BKI", "Kota Kinabalu Intl", "Kota Kinabalu", "MY", GeoCoord{5.9372, 116.0511}},
	{"WBGG", "KCH", "Kuching Intl", "Kuching", "MY", GeoCoord{1.4847, 110.3469}},
	{"WBSB", "BWN", "Brunei Intl", "Bandar Seri Begawan", "BN", GeoCoord{4.9442, 114.9283}},
	{"WSSL", "TGG", "Sultan Mahmud Airport", "Kuala Terengganu", "MY", GeoCoord{5.3826, 103.1030}},
	{"VTSG", "KBV", "Krabi Intl", "Krabi", "TH", GeoCoord{8.0991, 98.9862}},
	{"VTSS", "HDY", "Hat Yai Intl", "Hat Yai", "TH", GeoCoord{6.9331, 100.3928}},
	{"WIDD", "BTH", "Hang Nadim Intl", "Batam", "ID", GeoCoord{1.1211, 104.1189}},
	{"WIDN", "TNJ", "Raja Haji Fisabilillah", "Tanjung Pinang", "ID", GeoCoord{0.9239, 104.5319}},
	{"WIJJ", "DJB", "Sultan Thaha", "Jambi", "ID", GeoCoord{-1.6386, 103.6442}},
	{"WIPL", "BKS", "Fatmawati Soekarno", "Bengkulu", "ID", GeoCoord{-3.8636, 102.3400}},
	{"WILL", "TKG", "Radin Inten II", "Bandar Lampung", "ID", GeoCoord{-5.2422, 105.1786}},
	{"WIKK", "PGK", "Depati Amir", "Pangkal Pinang", "ID", GeoCoord{-2.1625, 106.1394}},
	{"WIOD", "TJQ", "H.A.S. Hanandjoeddin", "Belitung", "ID", GeoCoord{-2.7561, 107.7533}},
	{"WIOG", "PSU", "Pangsuma", "Putussibau", "ID", GeoCoord{0.8350, 112.9350}},
	{"WIOK", "KTG", "Rahadi Oesman", "Ketapang", "ID", GeoCoord{-1.8156, 109.9636}},
	{"WAOK", "SMQ", "H. Asan", "Sampit", "ID", GeoCoord{-2.5008, 112.9767}},
	{"WAON", "PKY", "Tjilik Riwut", "Palangkaraya", "ID", GeoCoord{-2.2253, 113.9458}},
	{"WAOP", "PKN", "Iskandar", "Pangkalan Bun", "ID", GeoCoord{-2.7042, 111.6733}},
	{"WALR", "AAP", "Aji Pangeran Tumenggung", "Samarinda", "ID", GeoCoord{-0.3736, 117.2581}},
	{"WALB", "BEJ", "Kalimarau", "Berau", "ID", GeoCoord{2.1550, 117.4339}},
	{"WAMP", "PLW", "Mutiara SIS Al-Jufrie", "Palu", "ID", GeoCoord{-0.9189, 119.9078}},
	{"WAMW", "LUW", "Syukuran Aminuddin Amir", "Luwuk", "ID", GeoCoord{-1.0378, 122.7744}},
	{"WAFW", "KDI", "Haluoleo", "Kendari", "ID", GeoCoord{-4.0819, 122.4172}},
	{"WAFB", "BUW", "Betoambari", "Baubau", "ID", GeoCoord{-5.4883, 122.5694}},
	{"WAPL", "LAH", "Oesman Sadik", "Labuha", "ID", GeoCoord{-0.6381, 127.4914}},
	{"WAPT", "TTE", "Sultan Babullah", "Ternate", "ID", GeoCoord{0.8319, 127.3808}},
	{"WAUA", "LUV", "Karel Sadsuitubun", "Langgur", "ID", GeoCoord{-5.6622, 132.7308}},
	{"WADB", "BMU", "Sultan Muhammad Salahuddin", "Bima", "ID", GeoCoord{-8.5392, 118.6872}},
	{"WADL", "LOP", "Lombok Intl", "Praya", "ID", GeoCoord{-8.7583, 116.2764}},
	{"WADW", "SWQ", "Sultan Muhammad Kaharuddin", "Sumbawa", "ID", GeoCoord{-8.4878, 117.4144}},
	{"WATL", "LBJ", "Komodo Intl", "Labuan Bajo", "ID", GeoCoord{-8.4875, 119.8886}},
	{"WATB", "ENE", "H. Hasan Aroeboesman", "Ende", "ID", GeoCoord{-8.8497, 121.6639}},
	{"WATM", "MOF", "Frans Seda", "Maumere", "ID", GeoCoord{-8.6414, 122.2378}},
	{"WASF", "FKQ", "Torea", "Fakfak", "ID", GeoCoord{-2.9208, 132.2667}},
	{"WASN", "NBX", "Nabire", "Nabire", "ID", GeoCoord{-3.3678, 135.4967}},
	{"WASE", "TIM", "Mozes Kilangin", "Timika", "ID", GeoCoord{-4.5283, 136.8867}},
	{"WASY", "BIK", "Frans Kaisiepo", "Biak", "ID", GeoCoord{-1.1906, 136.1089}},
	{"WASR", "MKW", "Rendani", "Manokwari", "ID", GeoCoord{-0.8917, 134.0500}},
	{"WAGI", "WMX", "Wamena", "Wamena", "ID", GeoCoord{-4.0978, 138.9519}},
	{"VTCB", "THS", "Sukhothai Airport", "Sukhothai", "TH", GeoCoord{17.2378, 99.8183}},
	{"VTUD", "UTH", "Udon Thani Intl", "Udon Thani", "TH", GeoCoord{17.3867, 102.7881}},
	{"VTUI", "SNO", "Sakon Nakhon Airport", "Sakon Nakhon", "TH", GeoCoord{17.1947, 104.1194}},
	{"VTUQ", "UBP", "Ubon Ratchathani", "Ubon Ratchathani", "TH", GeoCoord{15.2514, 104.8703}},
	{"VTUU", "ROI", "Roi Et Airport", "Roi Et", "TH", GeoCoord{16.1558, 103.7744}},
	{"VTUV", "BFV", "Buriram Airport", "Buriram", "TH", GeoCoord{15.2289, 103.2536}},
	{"VTSB", "URT", "Surat Thani Intl", "Surat Thani", "TH", GeoCoord{9.1325, 99.1417}},
	{"VTSM", "USM", "Samui Intl", "Koh Samui", "TH", GeoCoord{9.5478, 100.0628}},
	{"VTSN", "NST", "Nakhon Si Thammarat", "Nakhon Si Thammarat", "TH", GeoCoord{8.5400, 99.9442}},
	{"VTSP", "HKT", "Phuket Intl", "Phuket", "TH", GeoCoord{8.1133, 98.3169}},
	{"VTST", "TST", "Trang Airport", "Trang", "TH", GeoCoord{7.5097, 99.6161}},
	{"VTSY", "NAW", "Narathiwat Airport", "Narathiwat", "TH", GeoCoord{6.5200, 101.7431}},
	{"VTPB", "PRH", "Phrae Airport", "Phrae", "TH", GeoCoord{18.1322, 100.1639}},
	{"VTPO", "NXO", "Nan Nakhon Airport", "Nan", "TH", GeoCoord{18.8078, 100.7833}},
	{"VTPY", "PYY", "Pai Airport", "Mae Hong Son", "TH", GeoCoord{19.3697, 98.4358}},
	{"VTCL", "LPT", "Lampang Airport", "Lampang", "TH", GeoCoord{18.2711, 99.5050}},
	{"VTCH", "HGN", "Mae Hong Son", "Mae Hong Son", "TH", GeoCoord{19.3017, 97.9758}},
	{"VTBL", "KKM", "Khon Kaen Airport", "Khon Kaen", "TH", GeoCoord{16.4656, 102.7839}},
	{"WIII_ALT", "HLP", "Halim Perdanakusuma", "Jakarta", "ID", GeoCoord{-6.2656, 106.8908}},
	{"WARR_ALT", "MLG", "Abdul Rachman Saleh", "Malang", "ID", GeoCoord{-7.9264, 112.7144}},
	{"WADD_ALT", "BWX", "Banyuwangi Intl", "Banyuwangi", "ID", GeoCoord{-8.3125, 114.3394}},
	{"WIMM_ALT", "FLZ", "Ferdinand Lumban Tobing", "Sibolga", "ID", GeoCoord{1.5544, 98.8872}},
	{"WAAA_ALT", "MJU", "Tampa Padang", "Mamuju", "ID", GeoCoord{-2.5003, 119.0306}},
	{"WIPP_ALT", "LLG", "Silampari", "Lubuklinggau", "ID", GeoCoord{-3.2847, 102.9000}},
	{"WIOO_ALT", "SQG", "Susilo", "Sintang", "ID", GeoCoord{0.0639, 112.1800}},
	{"WALL_ALT", "TJG", "Warukin", "Tanjung", "ID", GeoCoord{-2.2133, 115.4417}},
	{"WAMM_ALT", "GTO", "Jalaluddin", "Gorontalo", "ID", GeoCoord{0.6372, 122.8500}},
	{"WAJJ_ALT", "DJJ_2", "Lereh Airport", "Jayapura", "ID", GeoCoord{-2.9833, 140.0167}},
	{"WIBB_ALT", "DUM", "Pinang Kampai", "Dumai", "ID", GeoCoord{1.6444, 101.4339}},
	{"WIEE_ALT", "SIW", "Silangit Intl", "Siborong-Borong", "ID", GeoCoord{2.2600, 98.9900}},
	{"WICC_ALT", "CBN", "Cakrabhuwana", "Cirebon", "ID", GeoCoord{-6.7556, 108.5397}},
	{"WAHH_ALT", "JOG", "Adisutjipto Intl", "Yogyakarta", "ID", GeoCoord{-7.7881, 110.4319}},
	{"WAHQ_ALT", "SOC", "Adisumarmo Intl", "Solo", "ID", GeoCoord{-7.5161, 110.7569}},
	{"WRLS_ALT", "NNX", "Nunukan", "Nunukan", "ID", GeoCoord{4.1383, 117.6533}},
	{"WAKK_ALT", "TMH", "Tanah Merah", "Boven Digoel", "ID", GeoCoord{-6.0967, 140.3017}},
	{"WASS_ALT", "RDE", "Raja Ampat Marinda", "Waisai", "ID", GeoCoord{-0.4333, 130.8167}},
	{"WATG_ALT", "ARD", "Alor Island", "Mali", "ID", GeoCoord{-8.2167, 124.5667}},
	{"WAPP_ALT", "NAM", "Namlea", "Buru", "ID", GeoCoord{-3.2458, 127.0944}},
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

// GlobalAirspaceSectors list of playable sectors
var GlobalAirspaceSectors = []AirspaceSector{
	{
		ID:         "sec-wiii",
		ICAO:       "WIII_FIR",
		Name:       "Jakarta FIR (Western Indonesia)",
		Center:     GeoCoord{-6.12, 106.65},
		Color:      "#00e676",
		Bounds: []GeoCoord{
			{6.0, 95.0}, {6.0, 108.0}, {-2.0, 110.0}, {-10.0, 110.0}, {-10.0, 95.0}, {6.0, 95.0},
		},
	},
	{
		ID:         "sec-wsjc",
		ICAO:       "WSJC_ACC",
		Name:       "Singapore ACC (Malacca Sector)",
		Center:     GeoCoord{1.36, 103.99},
		Color:      "#00b4d8",
		Bounds: []GeoCoord{
			{7.0, 99.0}, {7.0, 105.0}, {1.0, 105.0}, {0.0, 101.0}, {7.0, 99.0},
		},
	},
	{
		ID:         "sec-waaa",
		ICAO:       "WAAA_FIR",
		Name:       "Ujung Pandang FIR (Eastern Indonesia)",
		Center:     GeoCoord{-5.06, 119.55},
		Color:      "#ffb703",
		Bounds: []GeoCoord{
			{6.0, 110.0}, {6.0, 141.0}, {-11.0, 141.0}, {-11.0, 110.0}, {6.0, 110.0},
		},
	},
	{
		ID:         "sec-vtbb",
		ICAO:       "VTBB_ACC",
		Name:       "Bangkok ACC (Indochina Sector)",
		Center:     GeoCoord{13.69, 100.75},
		Color:      "#b026ff",
		Bounds: []GeoCoord{
			{20.0, 97.0}, {20.0, 106.0}, {7.0, 106.0}, {7.0, 97.0}, {20.0, 97.0},
		},
	},
	{
		ID:         "sec-ybbb",
		ICAO:       "YBBB_FIR",
		Name:       "Brisbane / Sydney ACC (Oceanic)",
		Center:     GeoCoord{-33.94, 151.17},
		Color:      "#ff2a5f",
		Bounds: []GeoCoord{
			{-10.0, 141.0}, {-10.0, 160.0}, {-40.0, 160.0}, {-40.0, 141.0}, {-10.0, 141.0},
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
