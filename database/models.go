package database

import (
	"database/sql"
	"time"

	"git.marceeli.ovh/vectura/vectura-api/models"
	"gorm.io/gorm"
)

type Route struct {
	gorm.Model
	CityId           string
	RouteId          string         `gorm:"uniqueIndex:idx_city_route"` 
	AgencyId         sql.NullString `gorm:"index"`                      
	RouteShortName   sql.NullString
	RouteLongName    sql.NullString
	RouteDescription sql.NullString
	RouteType        sql.NullInt16
	RouteUrl         sql.NullString
	RouteColor       sql.NullString
	RouteTextColor   sql.NullString
}

type Stop struct {
	gorm.Model
	CityId             string
	StopId             string `gorm:"uniqueIndex:idx_city_stop"` 
	StopCode           sql.NullString
	StopName           sql.NullString `gorm:"index"` 
	StopLat            sql.NullFloat64
	StopLon            sql.NullFloat64
	StopUrl            sql.NullString
	ZoneId             sql.NullString
	ParentStation      sql.NullString `gorm:"index"` 
	PlatformCode       sql.NullString
	WheelchairBoarding sql.NullInt16
	LocationType       sql.NullInt16
}

type Trip struct {
	gorm.Model
	CityId               string
	TripId               string `gorm:"uniqueIndex:idx_city_trip"` 
	RouteId              string `gorm:"index:idx_route"`           
	Route                Route  `gorm:"foreignKey:RouteId;references:RouteId"`
	ServiceId            string `gorm:"index:idx_service"` 
	BlockId              sql.NullString
	TripHeadsign         sql.NullString
	TripShortName        sql.NullString
	DirectionId          sql.NullInt16
	ShapeId              sql.NullString `gorm:"index"` 
	WheelchairAccessible sql.NullInt16
	BikeAccessible       sql.NullInt16
}

type Departure struct {
	gorm.Model
	CityId        string
	TripId        string `gorm:"index:idx_trip"`                                    
	Trip          Trip   `gorm:"foreignKey:TripId;references:TripId"`               
	StopId        string `gorm:"index:idx_stop;index:idx_stop_departure"`           
	Stop          Stop   `gorm:"foreignKey:StopId;references:StopId"`               
	ArrivalTime   string `gorm:"index:idx_arrival"`                                 
	DepartureTime string `gorm:"index:idx_stop_departure;index:idx_departure_time"` 
	StopSequence  int    `gorm:"index"`                                             
	PickupType    sql.NullInt16
	DropoffType   sql.NullInt16
}

type Calendar struct {
	gorm.Model
	CityId    string
	ServiceId string `gorm:"uniqueIndex:idx_city_service"` 
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool
	StartDate time.Time `gorm:"type:date"` 
	EndDate   time.Time `gorm:"type:date"` 
}

type CalendarDate struct {
	gorm.Model
	CityId        string
	ServiceId     string    `gorm:"index:idx_service_date;uniqueIndex:idx_city_service_date"` 
	Date          time.Time `gorm:"type:date;index:idx_service_date;uniqueIndex:idx_city_service_date"`
	ExceptionType int
}

type Shape struct {
	gorm.Model
	CityId          string
	ShapeId         string `gorm:"index:idx_shape;uniqueIndex:idx_shape_sequence"` 
	ShapePtLat      float64
	ShapePtLon      float64
	ShapePtSequence int `gorm:"uniqueIndex:idx_shape_sequence"` 
}

func DbRouteToRoute(dbRoute Route) models.Route {
	return models.Route{
		RouteId:          dbRoute.RouteId,
		AgencyId:         nullStringToString(dbRoute.AgencyId),
		RouteShortName:   nullStringToString(dbRoute.RouteShortName),
		RouteLongName:    nullStringToString(dbRoute.RouteLongName),
		RouteDescription: nullStringToString(dbRoute.RouteDescription),
		RouteType:        models.Type(nullInt16ToInt16(dbRoute.RouteType)),
		RouteUrl:         nullStringToString(dbRoute.RouteUrl),
		RouteColor:       nullStringToString(dbRoute.RouteColor),
		RouteTextColor:   nullStringToString(dbRoute.RouteTextColor),
	}
}

func DbStopToStop(dbStop Stop) models.Stop {
	return models.Stop{
		StopId:             dbStop.StopId,
		StopCode:           nullStringToString(dbStop.StopCode),
		StopName:           nullStringToString(dbStop.StopName),
		StopLat:            nullFloat64ToFloat64(dbStop.StopLat),
		StopLon:            nullFloat64ToFloat64(dbStop.StopLon),
		StopUrl:            nullStringToString(dbStop.StopUrl),
		ZoneId:             nullStringToString(dbStop.ZoneId),
		ParentStation:      nullStringToString(dbStop.ParentStation),
		PlatformCode:       nullStringToString(dbStop.PlatformCode),
		WheelchairBoarding: models.Accessibility(nullInt16ToInt16(dbStop.WheelchairBoarding)),
		LocationType:       models.Location(nullInt16ToInt16(dbStop.LocationType)),
	}
}

func DbTripToTrip(dbTrip Trip) models.Trip {
	return models.Trip{
		TripId:               dbTrip.TripId,
		RouteId:              dbTrip.RouteId,
		ServiceId:            dbTrip.ServiceId,
		BlockId:              nullStringToString(dbTrip.BlockId),
		TripHeadsign:         nullStringToString(dbTrip.TripHeadsign),
		TripShortName:        nullStringToString(dbTrip.TripShortName),
		DirectionId:          models.Direction(nullInt16ToInt16(dbTrip.DirectionId)),
		ShapeId:              nullStringToString(dbTrip.ShapeId),
		WheelchairAccessible: models.Accessibility(nullInt16ToInt16(dbTrip.WheelchairAccessible)),
		BikeAccessible:       models.Accessibility(nullInt16ToInt16(dbTrip.BikeAccessible)),
	}
}

func DbDepartureToDeparture(dbDep Departure) models.Departure {
	return models.Departure{
		Trip:          DbTripToTrip(dbDep.Trip),
		Route:         DbRouteToRoute(dbDep.Trip.Route),
		TripId:        dbDep.TripId,
		StopId:        dbDep.StopId,
		ArrivalTime:   dbDep.ArrivalTime,
		DepartureTime: dbDep.DepartureTime,
		StopSequence:  dbDep.StopSequence,
		PickupType:    models.PickupOrDropoff(nullInt16ToInt16(dbDep.PickupType)),
		DropoffType:   models.PickupOrDropoff(nullInt16ToInt16(dbDep.DropoffType)),
	}
}

func DbCalendarToCalendar(dbCal Calendar) models.Calendar {
	return models.Calendar{
		ServiceId: dbCal.ServiceId,
		Monday:    dbCal.Monday,
		Tuesday:   dbCal.Tuesday,
		Wednesday: dbCal.Wednesday,
		Thursday:  dbCal.Thursday,
		Friday:    dbCal.Friday,
		Saturday:  dbCal.Saturday,
		Sunday:    dbCal.Sunday,
		StartDate: dbCal.StartDate,
		EndDate:   dbCal.EndDate,
	}
}

func DbCalendarDateToCalendarDate(dbCalDate CalendarDate) models.CalendarDate {
	return models.CalendarDate{
		ServiceId:     dbCalDate.ServiceId,
		Date:          dbCalDate.Date,
		ExceptionType: models.ExceptionType(dbCalDate.ExceptionType),
	}
}

func DbShapeToShape(dbShape Shape) models.Shape {
	return models.Shape{
		ShapeId:         dbShape.ShapeId,
		ShapePtLat:      dbShape.ShapePtLat,
		ShapePtLon:      dbShape.ShapePtLon,
		ShapePtSequence: dbShape.ShapePtSequence,
	}
}

func RouteToDbRoute(route models.Route, cityId string) Route {
	return Route{
		CityId:           cityId,
		RouteId:          route.RouteId,
		AgencyId:         stringToNullString(route.AgencyId),
		RouteShortName:   stringToNullString(route.RouteShortName),
		RouteLongName:    stringToNullString(route.RouteLongName),
		RouteDescription: stringToNullString(route.RouteDescription),
		RouteType:        sql.NullInt16{Int16: int16(route.RouteType), Valid: true},
		RouteUrl:         stringToNullString(route.RouteUrl),
		RouteColor:       stringToNullString(route.RouteColor),
		RouteTextColor:   stringToNullString(route.RouteTextColor),
	}
}

func StopToDbStop(stop models.Stop, cityId string) Stop {
	return Stop{
		CityId:             cityId,
		StopId:             stop.StopId,
		StopCode:           stringToNullString(stop.StopCode),
		StopName:           stringToNullString(stop.StopName),
		StopLat:            sql.NullFloat64{Float64: stop.StopLat, Valid: true},
		StopLon:            sql.NullFloat64{Float64: stop.StopLon, Valid: true},
		StopUrl:            stringToNullString(stop.StopUrl),
		ZoneId:             stringToNullString(stop.ZoneId),
		ParentStation:      stringToNullString(stop.ParentStation),
		PlatformCode:       stringToNullString(stop.PlatformCode),
		WheelchairBoarding: sql.NullInt16{Int16: int16(stop.WheelchairBoarding), Valid: true},
		LocationType:       sql.NullInt16{Int16: int16(stop.LocationType), Valid: true},
	}
}

func TripToDbTrip(trip models.Trip, cityId string) Trip {
	return Trip{
		CityId:               cityId,
		TripId:               trip.TripId,
		RouteId:              trip.RouteId,
		ServiceId:            trip.ServiceId,
		BlockId:              stringToNullString(trip.BlockId),
		TripHeadsign:         stringToNullString(trip.TripHeadsign),
		TripShortName:        stringToNullString(trip.TripShortName),
		DirectionId:          sql.NullInt16{Int16: int16(trip.DirectionId), Valid: true},
		ShapeId:              stringToNullString(trip.ShapeId),
		WheelchairAccessible: sql.NullInt16{Int16: int16(trip.WheelchairAccessible), Valid: true},
		BikeAccessible:       sql.NullInt16{Int16: int16(trip.BikeAccessible), Valid: true},
	}
}

func DepartureToDbDeparture(dep models.Departure, cityId string) Departure {
	return Departure{
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

func CalendarToDbCalendar(cal models.Calendar, cityId string) Calendar {
	return Calendar{
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

func CalendarDateToDbCalendarDate(calDate models.CalendarDate, cityId string) CalendarDate {
	return CalendarDate{
		CityId:        cityId,
		ServiceId:     calDate.ServiceId,
		Date:          calDate.Date,
		ExceptionType: int(calDate.ExceptionType),
	}
}

func ShapeToDbShape(shape models.Shape, cityId string) Shape {
	return Shape{
		CityId:          cityId,
		ShapeId:         shape.ShapeId,
		ShapePtLat:      shape.ShapePtLat,
		ShapePtLon:      shape.ShapePtLon,
		ShapePtSequence: shape.ShapePtSequence,
	}
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func stringToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt16ToInt16(ni sql.NullInt16) int16 {
	if ni.Valid {
		return ni.Int16
	}
	return 0
}

func nullFloat64ToFloat64(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0
}
