package database

import (
	"archive/zip"
	"fmt"
	"time"

	"git.marceeli.ovh/vectura/vectura-api/models"
	"git.marceeli.ovh/vectura/vectura-api/parser"
	"git.marceeli.ovh/vectura/vectura-api/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func PreloadCities(db *gorm.DB) {
	cities := utils.LoadCitiesFromYAML()

	// Fuck the entire db
	db.Migrator().DropTable(
		&Stop{},
		&Route{},
		&Trip{},
		&Departure{},
		&Calendar{},
		&CalendarDate{},
		&Shape{},
	)

	// just kidding lmao
	db.AutoMigrate(
		&Stop{},
		&Route{},
		&Trip{},
		&Departure{},
		&Calendar{},
		&CalendarDate{},
		&Shape{},
	)

	for _, city := range cities {
		filePath := fmt.Sprintf("/tmp/%s.zip", city.ID)
		err := utils.SaveGTFS(city.URL, filePath)
		if err != nil {
			panic(err)
		}

		zipReader, err := zip.OpenReader(filePath)
		if err != nil {
			panic(err)
		}

		limit := 2000

		var dbStops []Stop
		for _, stop := range parser.GetStops(zipReader) {
			dbStops = append(dbStops, StopToDbStop(stop, city.ID))
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(dbStops, limit)

		dbStops = nil

		var dbRoutes []Route
		for _, route := range parser.GetRoutes(zipReader) {
			dbRoutes = append(dbRoutes, RouteToDbRoute(route, city.ID))
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(dbRoutes, limit)

		dbRoutes = nil

		var dbTrips []Trip
		for _, trip := range parser.GetTrips(zipReader) {
			dbTrips = append(dbTrips, TripToDbTrip(trip, city.ID))
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(dbTrips, limit)

		dbTrips = nil

		parser.ProcessDeparturesChunked(zipReader, 15000, func(departures []models.Departure) {
			if len(departures) > 0 {
				var dbDepartures []Departure
				for _, dep := range departures {
					dbDepartures = append(dbDepartures, DepartureToDbDeparture(dep, city.ID))
				}
				db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(dbDepartures, limit)
			}
		})

		var dbCalendars []Calendar
		for _, cal := range parser.GetCalendar(zipReader) {
			dbCalendars = append(dbCalendars, CalendarToDbCalendar(cal, city.ID))
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(dbCalendars, limit)

		dbCalendars = nil

		var dbCalendarDates []CalendarDate
		for _, cd := range parser.GetCalendarDates(zipReader) {
			dbCalendarDates = append(dbCalendarDates, CalendarDateToDbCalendarDate(cd, city.ID))
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(dbCalendarDates, limit)

		dbCalendarDates = nil

		var dbShapes []Shape
		for _, shape := range parser.GetShapes(zipReader) {
			dbShapes = append(dbShapes, ShapeToDbShape(shape, city.ID))
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(dbShapes, limit)

		dbShapes = nil

		zipReader.Close()

		// Clear interner cache between cities to prevent memory bloat
		parser.ClearInterner()

		println("Successfully loaded GTFS data for city:", city.ID)
	}
}

func GetStops(db *gorm.DB, city string) []models.Stop {
	var dbdata []Stop
	var data []models.Stop

	db.Table("stops").Where("city_id = ?", city).Find(&dbdata)

	for _, dat := range dbdata {
		data = append(data, DbStopToStop(dat))
	}

	return data
}

func GetRoutes(db *gorm.DB, city string) []models.Route {
	var dbdata []Route
	var data []models.Route

	db.Table("routes").Where("city_id = ?", city).Find(&dbdata)

	for _, dat := range dbdata {
		data = append(data, DbRouteToRoute(dat))
	}

	return data
}

func GetTrips(db *gorm.DB, city string) []models.Trip {
	var dbdata []Trip
	var data []models.Trip

	db.Table("trips").Where("city_id = ?", city).Find(&dbdata)

	for _, dat := range dbdata {
		data = append(data, DbTripToTrip(dat))
	}

	return data
}

func GetDepartures(db *gorm.DB, city string) []models.Departure {
	var dbdata []Departure
	var data []models.Departure

	db.Table("departures").Where("city_id = ?", city).Find(&dbdata)

	for _, dat := range dbdata {
		data = append(data, DbDepartureToDeparture(dat))
	}

	return data
}

func GetShapes(db *gorm.DB, city string) []models.Shape {
	var dbdata []Shape
	var data []models.Shape

	db.Table("shapes").Where("city_id = ?", city).Find(&dbdata)

	for _, dat := range dbdata {
		data = append(data, DbShapeToShape(dat))
	}

	return data
}

func GetShapeById(db *gorm.DB, city string, id string) []models.Shape {
	var dbshapes []Shape
	var shapes []models.Shape

	db.Table("shapes").Where("city_id = ?", city).Where("shape_id = ?", id).Order("shape_pt_sequence").Limit(-1).Find(&dbshapes)

	for _, shape := range dbshapes {
		shapes = append(shapes, DbShapeToShape(shape))
	}

	return shapes
}

func GetDeparturesForStop(db *gorm.DB, city string, id string) []models.Departure {
	var dbdeps []Departure
	var deps []models.Departure

	db.Table("departures").Where("city_id = ?", city).Where("stop_id = ?", id).Order("arrival_time").Limit(-1).Find(&dbdeps)

	for _, dep := range dbdeps {
		deps = append(deps, DbDepartureToDeparture(dep))
	}

	return deps
}

func GetActiveServicesForDate(db *gorm.DB, city string, date time.Time) map[string]bool {
	activeServices := make(map[string]bool)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	var calendars []models.Calendar
	db.Table("calendars").
		Where("city_id = ?", city).
		Where("start_date <= ?", date).
		Where("end_date >= ?", date).
		Find(&calendars)

	for _, cal := range calendars {
		var runs bool
		switch date.Weekday() {
		case time.Monday:
			runs = cal.Monday
		case time.Tuesday:
			runs = cal.Tuesday
		case time.Wednesday:
			runs = cal.Wednesday
		case time.Thursday:
			runs = cal.Thursday
		case time.Friday:
			runs = cal.Friday
		case time.Saturday:
			runs = cal.Saturday
		case time.Sunday:
			runs = cal.Sunday
		}
		if runs {
			activeServices[cal.ServiceId] = true
		}
	}

	var calendarDates []models.CalendarDate
	db.Table("calendar_dates").
		Where("city_id = ?", city).
		Where("date = ?", date).
		Find(&calendarDates)

	for _, cd := range calendarDates {
		switch cd.ExceptionType {
		case models.SERVICE_ADDED:
			activeServices[cd.ServiceId] = true
		case models.SERVICE_REMOVED:
			delete(activeServices, cd.ServiceId)
		}
	}

	return activeServices
}

func GetDeparturesForStopToday(db *gorm.DB, city string, stop string) []models.Departure {
	date := time.Now()

	return GetDeparturesForStopOnDate(db, city, stop, date)
}

func GetDeparturesForStopOnDate(db *gorm.DB, city string, stop string, date time.Time) []models.Departure {
	var (
		dbDeps []Departure
		deps   []models.Departure
	)

	services := GetActiveServicesForDate(db, city, date)
	if len(services) == 0 {
		return deps
	}

	// Extract service IDs from the map for use in SQL IN clause
	serviceIDs := make([]string, 0, len(services))
	for id := range services {
		serviceIDs = append(serviceIDs, id)
	}

	// Filter by active services in the DB query, avoiding in-memory filtering
	db.Model(&Departure{}).
		Preload("Trip.Route").
		Joins("JOIN trips ON trips.trip_id = departures.trip_id AND trips.city_id = departures.city_id").
		Where("departures.city_id = ?", city).
		Where("departures.stop_id = ?", stop).
		Where("trips.service_id IN ?", serviceIDs).
		Order("departures.departure_time").
		Find(&dbDeps)

	for _, dep := range dbDeps {
		deps = append(deps, DbDepartureToDeparture(dep))
	}

	return deps
}

func GetNextDeparturesForStopOnDate(db *gorm.DB, city string, stop string, date time.Time, limit int) []models.Departure {
	var (
		dbDeps []Departure
		deps   []models.Departure
	)

	services := GetActiveServicesForDate(db, city, date)
	if len(services) == 0 {
		return deps
	}

	// Get current time to filter departures
	currentTime := time.Now()
	currentTimeStr := currentTime.Format("15:04:05")

	// Extract service IDs from the map for use in SQL IN clause
	serviceIDs := make([]string, 0, len(services))
	for id := range services {
		serviceIDs = append(serviceIDs, id)
	}

	// Filter by active services and current time in the DB query, avoiding in-memory filtering
	db.Model(&Departure{}).
		Preload("Trip.Route").
		Joins("JOIN trips ON trips.trip_id = departures.trip_id AND trips.city_id = departures.city_id").
		Where("departures.city_id = ?", city).
		Where("departures.stop_id = ?", stop).
		Where("trips.service_id IN ?", serviceIDs).
		Where("departures.departure_time >= ?", currentTimeStr).
		Order("departures.departure_time").
		Limit(limit).
		Find(&dbDeps)

	for _, dep := range dbDeps {
		deps = append(deps, DbDepartureToDeparture(dep))
	}

	return deps
}
