package database

import (
	"time"

	"git.marceeli.ovh/vectura/vectura-api/models"
	"git.marceeli.ovh/vectura/vectura-api/parser"
	"git.marceeli.ovh/vectura/vectura-api/utils"
	"gorm.io/gorm"
)

func PreloadCities(db *gorm.DB) {
	cities := utils.LoadCitiesFromYAML("cities.yaml")

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

		// Process stops in chunks
		parser.ProcessStopsChunked(data, 1250, func(stops []models.Stop) {
			if len(stops) > 0 {
				var dbStops []Stop
				for _, stop := range stops {
					dbStops = append(dbStops, StopToDbStop(stop, city.ID))
				}
				db.CreateInBatches(dbStops, limit)
			}
		})

		// Process routes in chunks
		parser.ProcessRoutesChunked(data, 1500, func(routes []models.Route) {
			if len(routes) > 0 {
				var dbRoutes []Route
				for _, route := range routes {
					dbRoutes = append(dbRoutes, RouteToDbRoute(route, city.ID))
				}
				db.CreateInBatches(dbRoutes, limit)
			}
		})

		// Process trips in chunks
		parser.ProcessTripsChunked(data, 750, func(trips []models.Trip) {
			if len(trips) > 0 {
				var dbTrips []Trip
				for _, trip := range trips {
					dbTrips = append(dbTrips, TripToDbTrip(trip, city.ID))
				}
				db.CreateInBatches(dbTrips, limit)
			}
		})

		// Process departures in smaller chunks (this is usually the largest dataset)
		parser.ProcessDeparturesChunked(data, 750, func(departures []models.Departure) {
			if len(departures) > 0 {
				var dbDepartures []Departure
				for _, dep := range departures {
					dbDepartures = append(dbDepartures, DepartureToDbDeparture(dep, city.ID))
				}
				db.CreateInBatches(dbDepartures, limit)
			}
		})

		// Process calendars in chunks
		parser.ProcessCalendarsChunked(data, 1000, func(calendars []models.Calendar) {
			if len(calendars) > 0 {
				var dbCalendars []Calendar
				for _, cal := range calendars {
					dbCalendars = append(dbCalendars, CalendarToDbCalendar(cal, city.ID))
				}
				db.CreateInBatches(dbCalendars, limit)
			}
		})

		// Process calendar dates in chunks
		parser.ProcessCalendarDatesChunked(data, 1750, func(calendarDates []models.CalendarDate) {
			if len(calendarDates) > 0 {
				var dbCalendarDates []CalendarDate
				for _, cd := range calendarDates {
					dbCalendarDates = append(dbCalendarDates, CalendarDateToDbCalendarDate(cd, city.ID))
				}
				db.CreateInBatches(dbCalendarDates, limit)
			}
		})

		// Process shapes in chunks
		parser.ProcessShapesChunked(data, 1500, func(shapes []models.Shape) {
			if len(shapes) > 0 {
				var dbShapes []Shape
				for _, shape := range shapes {
					dbShapes = append(dbShapes, ShapeToDbShape(shape, city.ID))
				}
				db.CreateInBatches(dbShapes, limit)
			}
		})

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
