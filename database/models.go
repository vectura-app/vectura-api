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
	RouteId          string         `gorm:"uniqueIndex"`
	AgencyId         sql.NullString // TODO: add agency association
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
	StopId             string `gorm:"uniqueIndex"`
	StopCode           sql.NullString
	StopName           sql.NullString
	StopLat            sql.NullFloat64
	StopLon            sql.NullFloat64
	StopUrl            sql.NullString
	ZoneId             sql.NullString
	ParentStation      sql.NullString
	PlatformCode       sql.NullString
	WheelchairBoarding sql.NullInt16
	LocationType       sql.NullInt16
}

type Trip struct {
	gorm.Model
	CityId               string
	TripId               string `gorm:"uniqueIndex"`
	RouteId              string
	Route                Route `gorm:"foreignKey:RouteId;references:RouteId"`
	ServiceId            string
	BlockId              sql.NullString
	TripHeadsign         sql.NullString
	TripShortName        sql.NullString
	DirectionId          sql.NullInt16
	ShapeId              sql.NullString
	WheelchairAccessible sql.NullInt16
	BikeAccessible       sql.NullInt16
}

type Departure struct {
	gorm.Model
	CityId        string
	TripId        string
	Trip          Trip `gorm:"foreignKey:TripId;references:TripId"`
	StopId        string
	Stop          Stop `gorm:"foreignKey:StopId;references:StopId"`
	ArrivalTime   string
	DepartureTime string
	StopSequence  int
	PickupType    sql.NullInt16
	DropoffType   sql.NullInt16
}

type Calendar struct {
	gorm.Model
	CityId    string
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
	gorm.Model
	CityId        string
	ServiceId     string
	Date          time.Time
	ExceptionType int
}

type Shape struct {
	gorm.Model
	CityId          string
	ShapeId         string
	ShapePtLat      float64
	ShapePtLon      float64
	ShapePtSequence int
}

func DbRouteToRoute(dbRoute Route) models.Route {
	return models.Route{
		RouteId:          dbRoute.RouteId,
		AgencyId:         dbRoute.AgencyId.String,
		RouteShortName:   dbRoute.RouteShortName.String,
		RouteLongName:    dbRoute.RouteLongName.String,
		RouteDescription: dbRoute.RouteDescription.String,
		RouteType:        models.Type(dbRoute.RouteType.Int16),
		RouteUrl:         dbRoute.RouteUrl.String,
		RouteColor:       dbRoute.RouteColor.String,
		RouteTextColor:   dbRoute.RouteTextColor.String,
	}
}

func DbStopToStop(dbStop Stop) models.Stop {
	return models.Stop{
		StopId:             dbStop.StopId,
		StopCode:           dbStop.StopCode.String,
		StopName:           dbStop.StopName.String,
		StopLat:            dbStop.StopLat.Float64,
		StopLon:            dbStop.StopLon.Float64,
		StopUrl:            dbStop.StopUrl.String,
		ZoneId:             dbStop.ZoneId.String,
		ParentStation:      dbStop.ParentStation.String,
		PlatformCode:       dbStop.PlatformCode.String,
		WheelchairBoarding: models.Accessibility(dbStop.WheelchairBoarding.Int16),
		LocationType:       models.Location(dbStop.LocationType.Int16),
	}
}

func DbTripToTrip(dbTrip Trip) models.Trip {
	return models.Trip{
		TripId:               dbTrip.TripId,
		RouteId:              dbTrip.RouteId,
		ServiceId:            dbTrip.ServiceId,
		BlockId:              dbTrip.BlockId.String,
		TripHeadsign:         dbTrip.TripHeadsign.String,
		TripShortName:        dbTrip.TripShortName.String,
		DirectionId:          models.Direction(dbTrip.DirectionId.Int16),
		ShapeId:              dbTrip.ShapeId.String,
		WheelchairAccessible: models.Accessibility(dbTrip.WheelchairAccessible.Int16),
		BikeAccessible:       models.Accessibility(dbTrip.BikeAccessible.Int16),
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
		PickupType:    models.PickupOrDropoff(dbDep.PickupType.Int16),
		DropoffType:   models.PickupOrDropoff(dbDep.DropoffType.Int16),
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

func StopToDbStop(stop models.Stop, cityId string) Stop {
	return Stop{
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

func TripToDbTrip(trip models.Trip, cityId string) Trip {
	return Trip{
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
