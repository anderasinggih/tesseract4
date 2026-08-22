package main

// RealWorldFlightSchedule represents an authentic international airline route with ICAO designators
type RealWorldFlightSchedule struct {
	Callsign     string // e.g. "GIA880"
	FlightNumber string // e.g. "GA880"
	Airline      string // e.g. "Garuda Indonesia"
	AircraftType string // e.g. "B77W" (ICAO Doc 8643)
	Origin       string // e.g. "WIII" (Jakarta)
	Destination  string // e.g. "RJAA" (Tokyo Narita)
	CruiseFL     int    // e.g. 35000 (FL350)
	CruiseSpeed  int    // e.g. 475 (Knots)
}

// GlobalOfficialFlightSchedules is an authentic curated database of real commercial flight routes worldwide
var GlobalOfficialFlightSchedules = []RealWorldFlightSchedule{
	// 🇮🇩 Garuda Indonesia (GA/GIA)
	{"GIA880", "GA880", "Garuda Indonesia", "B77W", "WIII", "RJAA", 37000, 480},
	{"GIA881", "GA881", "Garuda Indonesia", "B77W", "RJAA", "WIII", 36000, 480},
	{"GIA870", "GA870", "Garuda Indonesia", "A339", "WADD", "RKSI", 38000, 470},
	{"GIA871", "GA871", "Garuda Indonesia", "A339", "RKSI", "WADD", 39000, 470},
	{"GIA716", "GA716", "Garuda Indonesia", "A333", "WIII", "YSSY", 37000, 475},
	{"GIA717", "GA717", "Garuda Indonesia", "A333", "YSSY", "WIII", 38000, 475},
	{"GIA86",  "GA086", "Garuda Indonesia", "B77W", "WIII", "EHAM", 34000, 485},
	{"GIA87",  "GA087", "Garuda Indonesia", "B77W", "EHAM", "WIII", 35000, 485},
	{"GIA836", "GA836", "Garuda Indonesia", "B738", "WIII", "WSSS", 28000, 430},
	{"GIA837", "GA837", "Garuda Indonesia", "B738", "WSSS", "WIII", 29000, 430},

	// 🇸🇬 Singapore Airlines (SQ/SIA)
	{"SIA318", "SQ318", "Singapore Airlines", "A359", "WSSS", "EGLL", 38000, 480},
	{"SIA319", "SQ319", "Singapore Airlines", "A359", "EGLL", "WSSS", 37000, 480},
	{"SIA22",  "SQ022", "Singapore Airlines", "A359", "WSSS", "KJFK", 41000, 490},
	{"SIA21",  "SQ021", "Singapore Airlines", "A359", "KJFK", "WSSS", 40000, 490},
	{"SIA221", "SQ221", "Singapore Airlines", "A388", "WSSS", "YSSY", 38000, 480},
	{"SIA222", "SQ222", "Singapore Airlines", "A388", "YSSY", "WSSS", 39000, 480},
	{"SIA958", "SQ958", "Singapore Airlines", "B78X", "WSSS", "WIII", 28000, 430},
	{"SIA959", "SQ959", "Singapore Airlines", "B78X", "WIII", "WSSS", 27000, 430},
	{"SIA638", "SQ638", "Singapore Airlines", "B77W", "WSSS", "RJAA", 37000, 485},

	// 🇯🇵 All Nippon Airways & Japan Airlines (NH/ANA & JL/JAL)
	{"ANA855", "NH855", "All Nippon Airways", "B789", "RJTT", "WIII", 38000, 475},
	{"ANA856", "NH856", "All Nippon Airways", "B789", "WIII", "RJTT", 39000, 475},
	{"JAL725", "JL725", "Japan Airlines", "B788", "RJAA", "WIII", 37000, 470},
	{"JAL726", "JL726", "Japan Airlines", "B788", "WIII", "RJAA", 38000, 470},
	{"ANA008", "NH008", "All Nippon Airways", "B77W", "RJAA", "KLAX", 35000, 490},
	{"ANA007", "NH007", "All Nippon Airways", "B77W", "KLAX", "RJAA", 36000, 490},
	{"JAL043", "JL043", "Japan Airlines", "A359", "RJTT", "EGLL", 38000, 485},

	// 🇦🇺 Qantas Airways (QF/QFA)
	{"QFA1",   "QF001", "Qantas Airways", "A388", "YSSY", "EGLL", 36000, 480},
	{"QFA2",   "QF002", "Qantas Airways", "A388", "EGLL", "YSSY", 37000, 480},
	{"QFA11",  "QF011", "Qantas Airways", "B789", "YSSY", "KLAX", 38000, 490},
	{"QFA12",  "QF012", "Qantas Airways", "B789", "KLAX", "YSSY", 39000, 490},
	{"QFA41",  "QF041", "Qantas Airways", "A332", "YSSY", "WIII", 38000, 475},
	{"QFA42",  "QF042", "Qantas Airways", "A332", "WIII", "YSSY", 37000, 475},
	{"QFA29",  "QF029", "Qantas Airways", "A333", "YMML", "VHHH", 38000, 475},

	// 🇬🇧 British Airways (BA/BAW) & 🇫🇷 Air France (AF/AFR) & 🇩🇪 Lufthansa (LH/DLH)
	{"BAW11",  "BA011", "British Airways", "B77W", "EGLL", "WSSS", 35000, 485},
	{"BAW12",  "BA012", "British Airways", "B77W", "WSSS", "EGLL", 36000, 485},
	{"BAW177", "BA177", "British Airways", "B772", "EGLL", "KJFK", 37000, 480},
	{"BAW178", "BA178", "British Airways", "B772", "KJFK", "EGLL", 38000, 480},
	{"AFR258", "AF258", "Air France", "B77W", "LFPG", "WSSS", 36000, 485},
	{"AFR259", "AF259", "Air France", "B77W", "WSSS", "LFPG", 37000, 485},
	{"DLH778", "LH778", "Lufthansa", "A359", "EDDF", "WSSS", 37000, 480},
	{"DLH779", "LH779", "Lufthansa", "A359", "WSSS", "EDDF", 38000, 480},
	{"KLM809", "KL809", "KLM Royal Dutch", "B772", "EHAM", "WIII", 34000, 485},
	{"KLM810", "KL810", "KLM Royal Dutch", "B772", "WIII", "EHAM", 35000, 485},

	// 🇦🇪 Emirates (EK/UAE) & 🇶🇦 Qatar Airways (QR/QTR)
	{"UAE356", "EK356", "Emirates Airline", "A388", "OMDB", "WIII", 37000, 485},
	{"UAE357", "EK357", "Emirates Airline", "A388", "WIII", "OMDB", 38000, 485},
	{"UAE414", "EK414", "Emirates Airline", "A388", "OMDB", "YSSY", 38000, 485},
	{"UAE001", "EK001", "Emirates Airline", "A388", "OMDB", "EGLL", 36000, 480},
	{"UAE201", "EK201", "Emirates Airline", "A388", "OMDB", "KJFK", 39000, 490},
	{"QTR956", "QR956", "Qatar Airways", "B77W", "OTHH", "WIII", 37000, 485},
	{"QTR957", "QR957", "Qatar Airways", "B77W", "WIII", "OTHH", 38000, 485},
	{"QTR701", "QR701", "Qatar Airways", "A35K", "OTHH", "KJFK", 39000, 490},

	// 🇺🇸 Delta, United, American Airlines (DL/DAL, UA/UAL, AA/AAL)
	{"UAL871", "UA871", "United Airlines", "B77W", "KLAX", "RJAA", 36000, 490},
	{"UAL872", "UA872", "United Airlines", "B77W", "RJAA", "KLAX", 37000, 490},
	{"DAL283", "DL283", "Delta Air Lines", "A359", "KLAX", "RJAA", 38000, 485},
	{"AAL100", "AA100", "American Airlines", "B772", "KJFK", "EGLL", 37000, 480},
	{"AAL101", "AA101", "American Airlines", "B772", "EGLL", "KJFK", 38000, 480},
	{"UAL1",   "UA001", "United Airlines", "B789", "KORD", "EGLL", 38000, 485},

	// 🇭🇰 Cathay Pacific (CX/CPA) & 🇰🇷 Korean Air (KE/KAL)
	{"CPA719", "CX719", "Cathay Pacific", "A359", "VHHH", "WIII", 37000, 475},
	{"CPA718", "CX718", "Cathay Pacific", "A359", "WIII", "VHHH", 38000, 475},
	{"CPA888", "CX888", "Cathay Pacific", "B77W", "VHHH", "KJFK", 39000, 490},
	{"KAL627", "KE627", "Korean Air", "B77W", "RKSI", "WADD", 38000, 480},
	{"KAL628", "KE628", "Korean Air", "B77W", "WADD", "RKSI", 39000, 480},
	{"KAL011", "KE011", "Korean Air", "B748", "RKSI", "KLAX", 36000, 490},

	// 🇲🇾 Malaysia Airlines (MH/MAS) & 🇹🇭 Thai Airways (TG/THA)
	{"MAS725", "MH725", "Malaysia Airlines", "A333", "WMKK", "WIII", 28000, 430},
	{"MAS726", "MH726", "Malaysia Airlines", "A333", "WIII", "WMKK", 29000, 430},
	{"MAS001", "MH001", "Malaysia Airlines", "A359", "EGLL", "WMKK", 37000, 480},
	{"THA435", "TG435", "Thai Airways", "B77W", "VTBS", "WIII", 36000, 475},
	{"THA436", "TG436", "Thai Airways", "B77W", "WIII", "VTBS", 37000, 475},
	{"THA910", "TG910", "Thai Airways", "A359", "VTBS", "EGLL", 38000, 480},
}
