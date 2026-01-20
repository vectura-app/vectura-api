package models

import (
	"database/sql"
	"time"
)

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
	Shapes        []Shape
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
	ArrivalTime   string
	DepartureTime string
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

type Shape struct {
	ShapeId         string
	ShapePtLat      float64
	ShapePtLon      float64
	ShapePtSequence int
}

func RouteToDbRoute(route Route, cityId string) DbRoute {
	return DbRoute{
		CityId:           cityId,
		RouteId:          route.RouteId,
		AgencyId:         sql.NullString{String: route.AgencyId, Valid: route.AgencyId != ""},
		RouteShortName:   sql.NullString{String: route.RouteShortName, Valid: route.RouteShortName != ""},
		RouteLongName:    sql.NullString{String: route.RouteLongName, Valid: route.RouteLongName != ""},
		RouteDescription: sql.NullString{String: route.RouteDescription, Valid: route.RouteDescription != ""},
		RouteType:        sql.NullInt16{Int16: int16(route.RouteType), Valid: true},
		RouteUrl:         sql.NullString{String: route.RouteUrl, Valid: route.RouteUrl != ""},
		RouteColor:       sql.NullString{String: route.RouteColor, Valid: route.RouteColor != ""},
		RouteTextColor:   sql.NullString{String: route.RouteTextColor, Valid: route.RouteTextColor != ""},
	}
}

func StopToDbStop(stop Stop, cityId string) DbStop {
	return DbStop{
		CityId:             cityId,
		StopId:             stop.StopId,
		StopCode:           sql.NullString{String: stop.StopCode, Valid: stop.StopCode != ""},
		StopName:           sql.NullString{String: stop.StopName, Valid: stop.StopName != ""},
		StopLat:            sql.NullFloat64{Float64: stop.StopLat, Valid: true},
		StopLon:            sql.NullFloat64{Float64: stop.StopLon, Valid: true},
		StopUrl:            sql.NullString{String: stop.StopUrl, Valid: stop.StopUrl != ""},
		ZoneId:             sql.NullString{String: stop.ZoneId, Valid: stop.ZoneId != ""},
		ParentStation:      sql.NullString{String: stop.ParentStation, Valid: stop.ParentStation != ""},
		PlatformCode:       sql.NullString{String: stop.PlatformCode, Valid: stop.PlatformCode != ""},
		WheelchairBoarding: sql.NullInt16{Int16: int16(stop.WheelchairBoarding), Valid: true},
		LocationType:       sql.NullInt16{Int16: int16(stop.LocationType), Valid: true},
	}
}

func TripToDbTrip(trip Trip, cityId string) DbTrip {
	return DbTrip{
		CityId:               cityId,
		TripId:               trip.TripId,
		RouteId:              trip.RouteId,
		ServiceId:            trip.ServiceId,
		BlockId:              sql.NullString{String: trip.BlockId, Valid: trip.BlockId != ""},
		TripHeadsign:         sql.NullString{String: trip.TripHeadsign, Valid: trip.TripHeadsign != ""},
		TripShortName:        sql.NullString{String: trip.TripShortName, Valid: trip.TripShortName != ""},
		DirectionId:          sql.NullInt16{Int16: int16(trip.DirectionId), Valid: true},
		ShapeId:              sql.NullString{String: trip.ShapeId, Valid: trip.ShapeId != ""},
		WheelchairAccessible: sql.NullInt16{Int16: int16(trip.WheelchairAccessible), Valid: true},
		BikeAccessible:       sql.NullInt16{Int16: int16(trip.BikeAccessible), Valid: true},
	}
}

func DepartureToDbDeparture(dep Departure, cityId string) DbDeparture {
	return DbDeparture{
		CityId:        cityId,
		TripId:        dep.TripId,
		StopId:        dep.StopId,
		ArrivalTime:   dep.ArrivalTime,
		DepartureTime: dep.DepartureTime,
		StopSequence:  dep.StopSequence,
		PickupType:    sql.NullInt16{Int16: int16(dep.PickupType), Valid: true},
		DropoffType:   sql.NullInt16{Int16: int16(dep.DropoffType), Valid: true},
	}
}

func CalendarToDbCalendar(cal Calendar, cityId string) DbCalendar {
	return DbCalendar{
		CityId:    cityId,
		ServiceId: cal.ServiceId,
		Monday:    cal.Monday,
		Tuesday:   cal.Tuesday,
		Wednesday: cal.Wednesday,
		Thursday:  cal.Thursday,
		Friday:    cal.Friday,
		Saturday:  cal.Saturday,
		Sunday:    cal.Sunday,
		StartDate: cal.StartDate,
		EndDate:   cal.EndDate,
	}
}

func CalendarDateToDbCalendarDate(calDate CalendarDate, cityId string) DbCalendarDate {
	return DbCalendarDate{
		CityId:        cityId,
		ServiceId:     calDate.ServiceId,
		Date:          calDate.Date,
		ExceptionType: int(calDate.ExceptionType),
	}
}

func ShapeToDbShape(shape Shape, cityId string) DbShape {
	return DbShape{
		CityId:          cityId,
		ShapeId:         shape.ShapeId,
		ShapePtLat:      shape.ShapePtLat,
		ShapePtLon:      shape.ShapePtLon,
		ShapePtSequence: shape.ShapePtSequence,
	}
}
