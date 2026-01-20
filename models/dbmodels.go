package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type DbRoute struct {
	gorm.Model
	CityId           string
	RouteId          string
	AgencyId         sql.NullString
	RouteShortName   sql.NullString
	RouteLongName    sql.NullString
	RouteDescription sql.NullString
	RouteType        sql.NullInt16
	RouteUrl         sql.NullString
	RouteColor       sql.NullString
	RouteTextColor   sql.NullString
}

type DbStop struct {
	gorm.Model
	CityId             string
	StopId             string
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

type DbTrip struct {
	gorm.Model
	CityId               string
	TripId               string
	RouteId              string
	ServiceId            string
	BlockId              sql.NullString
	TripHeadsign         sql.NullString
	TripShortName        sql.NullString
	DirectionId          sql.NullInt16
	ShapeId              sql.NullString
	WheelchairAccessible sql.NullInt16
	BikeAccessible       sql.NullInt16
}

type DbDeparture struct {
	gorm.Model
	CityId        string
	TripId        string
	StopId        string
	ArrivalTime   string
	DepartureTime string
	StopSequence  int
	PickupType    sql.NullInt16
	DropoffType   sql.NullInt16
}

type DbCalendar struct {
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

type DbCalendarDate struct {
	gorm.Model
	CityId        string
	ServiceId     string
	Date          time.Time
	ExceptionType int
}

type DbShape struct {
	gorm.Model
	CityId          string
	ShapeId         string
	ShapePtLat      float64
	ShapePtLon      float64
	ShapePtSequence int
}

func DbRouteToRoute(dbRoute DbRoute) Route {
	return Route{
		RouteId:          dbRoute.RouteId,
		AgencyId:         dbRoute.AgencyId.String,
		RouteShortName:   dbRoute.RouteShortName.String,
		RouteLongName:    dbRoute.RouteLongName.String,
		RouteDescription: dbRoute.RouteDescription.String,
		RouteType:        Type(dbRoute.RouteType.Int16),
		RouteUrl:         dbRoute.RouteUrl.String,
		RouteColor:       dbRoute.RouteColor.String,
		RouteTextColor:   dbRoute.RouteTextColor.String,
	}
}

func DbStopToStop(dbStop DbStop) Stop {
	return Stop{
		StopId:             dbStop.StopId,
		StopCode:           dbStop.StopCode.String,
		StopName:           dbStop.StopName.String,
		StopLat:            dbStop.StopLat.Float64,
		StopLon:            dbStop.StopLon.Float64,
		StopUrl:            dbStop.StopUrl.String,
		ZoneId:             dbStop.ZoneId.String,
		ParentStation:      dbStop.ParentStation.String,
		PlatformCode:       dbStop.PlatformCode.String,
		WheelchairBoarding: Accessibility(dbStop.WheelchairBoarding.Int16),
		LocationType:       Location(dbStop.LocationType.Int16),
	}
}

func DbTripToTrip(dbTrip DbTrip) Trip {
	return Trip{
		TripId:               dbTrip.TripId,
		RouteId:              dbTrip.RouteId,
		ServiceId:            dbTrip.ServiceId,
		BlockId:              dbTrip.BlockId.String,
		TripHeadsign:         dbTrip.TripHeadsign.String,
		TripShortName:        dbTrip.TripShortName.String,
		DirectionId:          Direction(dbTrip.DirectionId.Int16),
		ShapeId:              dbTrip.ShapeId.String,
		WheelchairAccessible: Accessibility(dbTrip.WheelchairAccessible.Int16),
		BikeAccessible:       Accessibility(dbTrip.BikeAccessible.Int16),
	}
}

func DbDepartureToDeparture(dbDep DbDeparture) Departure {
	return Departure{
		TripId:        dbDep.TripId,
		StopId:        dbDep.StopId,
		ArrivalTime:   dbDep.ArrivalTime,
		DepartureTime: dbDep.DepartureTime,
		StopSequence:  dbDep.StopSequence,
		PickupType:    PickupOrDropoff(dbDep.PickupType.Int16),
		DropoffType:   PickupOrDropoff(dbDep.DropoffType.Int16),
	}
}

func DbCalendarToCalendar(dbCal DbCalendar) Calendar {
	return Calendar{
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

func DbCalendarDateToCalendarDate(dbCalDate DbCalendarDate) CalendarDate {
	return CalendarDate{
		ServiceId:     dbCalDate.ServiceId,
		Date:          dbCalDate.Date,
		ExceptionType: ExceptionType(dbCalDate.ExceptionType),
	}
}

func DbShapeToShape(dbShape DbShape) Shape {
	return Shape{
		ShapeId:         dbShape.ShapeId,
		ShapePtLat:      dbShape.ShapePtLat,
		ShapePtLon:      dbShape.ShapePtLon,
		ShapePtSequence: dbShape.ShapePtSequence,
	}
}
