package models

import "time"

type Type uint8
type Location uint8
type Accessibility uint8
type PickupOrDropoff uint8
type Direction uint8
type ExceptionType uint8

const (
	TRAM       Type = 0
	METRO      Type = 1
	RAIL       Type = 2
	BUS        Type = 3
	FERRY      Type = 4
	CABLE_TRAM Type = 5
	LIFT       Type = 6
	FUNICULAR  Type = 7
	TROLLEYBUS Type = 11
	MONORAIL   Type = 12
)

const (
	STOP          Location = 0
	STATION       Location = 1
	ENTRANCE_EXIT Location = 2
	NODE          Location = 3
	BOARDING      Location = 4
)

const (
	NOT_AVAILABLE  Accessibility = 0
	ACCESSIBLE     Accessibility = 1
	NOT_ACCESSIBLE Accessibility = 2
)

const (
	REGULAR              PickupOrDropoff = 0
	PICKUP_NOT_AVAILABLE PickupOrDropoff = 1
	ON_CALL              PickupOrDropoff = 2
	ON_DEMAND            PickupOrDropoff = 3
)

const (
	INBOUND  Direction = 0
	OUTBOUND Direction = 1
)

const (
	SERVICE_ADDED   ExceptionType = 1
	SERVICE_REMOVED ExceptionType = 2
)

type GTFSData struct {
	Stops         []Stop
	Routes        []Route
	Trips         []Trip
	Departures    []Departure
	Calendars     []Calendar
	CalendarDates []CalendarDate
}

type Route struct {
	RouteId          string
	AgencyId         string
	RouteShortName   string
	RouteLongName    string
	RouteDescription string
	RouteType        Type
	RouteUrl         string
	RouteColor       string
	RouteTextColor   string
}

type Stop struct {
	StopId             string
	StopCode           string
	StopName           string
	StopLat            float64
	StopLon            float64
	StopUrl            string
	ZoneId             string
	ParentStation      string
	PlatformCode       string
	WheelchairBoarding Accessibility
	LocationType       Location
}

type Trip struct {
	TripId               string
	RouteId              string
	ServiceId            string
	BlockId              string
	TripHeadsign         string
	TripShortName        string
	DirectionId          Direction
	ShapeId              string
	WheelchairAccessible Accessibility
	BikeAccessible       Accessibility
}

type Departure struct {
	TripId        string
	StopId        string
	ArrivalTime   int64
	DepartureTime int64
	StopSequence  int
	PickupType    PickupOrDropoff
	DropoffType   PickupOrDropoff
}

type Calendar struct {
	ServiceId string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool
	StartDate time.Time
	EndDate   time.Time
}

type CalendarDate struct {
	ServiceId     string
	Date          time.Time
	ExceptionType ExceptionType
}
