package database

import (
	"time"

	"git.marceeli.ovh/vectura/vectura-api/models"
	"git.marceeli.ovh/vectura/vectura-api/parser"
	"git.marceeli.ovh/vectura/vectura-api/utils"
	"gorm.io/gorm"
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
		data, err := utils.FetchGTFS(city.URL)
		if err != nil {
			panic(err)
		}

		limit := 2000

		db.CreateInBatches(parser.GetStops(data), limit)
		db.CreateInBatches(parser.GetRoutes(data), limit)
		db.CreateInBatches(parser.GetTrips(data), limit)
		db.CreateInBatches(parser.GetDepartures(data), limit)
		db.CreateInBatches(parser.GetCalendar(data), limit)
		db.CreateInBatches(parser.GetCalendarDates(data), limit)
		db.CreateInBatches(parser.GetShapes(data), limit)

		// Clear the downloaded data to help GC
		data = nil

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

	// Get departures for this stop with preloaded Trip
	db.Model(&Departure{}).
		Table("departures").
		Preload("Trip.Route").
		Where("city_id = ?", city).
		Where("stop_id = ?", stop).
		Order("departure_time").
		Find(&dbDeps)

	// Filter departures by active services using preloaded Trip
	for _, dep := range dbDeps {
		if services[dep.Trip.ServiceId] {
			deps = append(deps, DbDepartureToDeparture(dep))
		}
	}

	return deps
}
