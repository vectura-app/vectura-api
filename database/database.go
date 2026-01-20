package database

import (
	"time"

	"git.marceeli.ovh/vectura/vectura-api/models"
	"git.marceeli.ovh/vectura/vectura-api/parser"
	"git.marceeli.ovh/vectura/vectura-api/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func PreloadCities() {
	cities := utils.LoadCitiesFromYAML("cities.yaml")

	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic("failed to connect to database")
	}

	// Fuck the entire db
	db.Migrator().DropTable(
		&models.DbStop{},
		&models.DbRoute{},
		&models.DbTrip{},
		&models.DbDeparture{},
		&models.DbCalendar{},
		&models.DbCalendarDate{},
		&models.DbShape{},
	)

	// just kidding lmao
	db.AutoMigrate(
		&models.DbStop{},
		&models.DbRoute{},
		&models.DbTrip{},
		&models.DbDeparture{},
		&models.DbCalendar{},
		&models.DbCalendarDate{},
		&models.DbShape{},
	)

	for _, city := range cities {
		data, err := utils.FetchGTFS(city.URL)
		if err != nil {
			panic(err)
		}

		// Process stops in chunks
		parser.ProcessStopsChunked(data, 1000, func(stops []models.Stop) {
			if len(stops) > 0 {
				var dbStops []models.DbStop
				for _, stop := range stops {
					dbStops = append(dbStops, models.StopToDbStop(stop, city.ID))
				}
				db.CreateInBatches(dbStops, 150)
			}
		})

		// Process routes in chunks
		parser.ProcessRoutesChunked(data, 1000, func(routes []models.Route) {
			if len(routes) > 0 {
				var dbRoutes []models.DbRoute
				for _, route := range routes {
					dbRoutes = append(dbRoutes, models.RouteToDbRoute(route, city.ID))
				}
				db.CreateInBatches(dbRoutes, 150)
			}
		})

		// Process trips in chunks
		parser.ProcessTripsChunked(data, 1000, func(trips []models.Trip) {
			if len(trips) > 0 {
				var dbTrips []models.DbTrip
				for _, trip := range trips {
					dbTrips = append(dbTrips, models.TripToDbTrip(trip, city.ID))
				}
				db.CreateInBatches(dbTrips, 150)
			}
		})

		// Process departures in smaller chunks (this is usually the largest dataset)
		parser.ProcessDeparturesChunked(data, 500, func(departures []models.Departure) {
			if len(departures) > 0 {
				var dbDepartures []models.DbDeparture
				for _, dep := range departures {
					dbDepartures = append(dbDepartures, models.DepartureToDbDeparture(dep, city.ID))
				}
				db.CreateInBatches(dbDepartures, 150)
			}
		})

		// Process calendars in chunks
		parser.ProcessCalendarsChunked(data, 1000, func(calendars []models.Calendar) {
			if len(calendars) > 0 {
				var dbCalendars []models.DbCalendar
				for _, cal := range calendars {
					dbCalendars = append(dbCalendars, models.CalendarToDbCalendar(cal, city.ID))
				}
				db.CreateInBatches(dbCalendars, 150)
			}
		})

		// Process calendar dates in chunks
		parser.ProcessCalendarDatesChunked(data, 1000, func(calendarDates []models.CalendarDate) {
			if len(calendarDates) > 0 {
				var dbCalendarDates []models.DbCalendarDate
				for _, cd := range calendarDates {
					dbCalendarDates = append(dbCalendarDates, models.CalendarDateToDbCalendarDate(cd, city.ID))
				}
				db.CreateInBatches(dbCalendarDates, 150)
			}
		})

		// Process shapes in chunks
		parser.ProcessShapesChunked(data, 1000, func(shapes []models.Shape) {
			if len(shapes) > 0 {
				var dbShapes []models.DbShape
				for _, shape := range shapes {
					dbShapes = append(dbShapes, models.ShapeToDbShape(shape, city.ID))
				}
				db.CreateInBatches(dbShapes, 150)
			}
		})

		// Clear the downloaded data to help GC
		data = nil

		// Clear interner cache between cities to prevent memory bloat
		parser.ClearInterner()

		println("Successfully loaded GTFS data for city:", city.ID)
	}
}

// TODO: do

func GetShapeById(id string, shapes []models.Shape) []models.Shape {
	return []models.Shape{}
}

func GetDeparturesForStop(allDepartures []models.Departure, stop string) []models.Departure {
	return []models.Departure{}
}

func GetActiveServicesForDate(date time.Time, calendars []models.Calendar, calendarDates []models.CalendarDate) map[string]bool {
	return map[string]bool{}
}

func GetDeparturesForStopToday(
	stop string,
	calendars []models.Calendar,
	calendarDates []models.CalendarDate,
	allDepartures []models.Departure,
	trips []models.Trip,
) []models.Departure {
	today := time.Now()
	return GetDeparturesForStopOnDate(stop, today, calendars, calendarDates, allDepartures, trips)
}

func GetDeparturesForStopOnDate(
	stop string,
	date time.Time,
	calendars []models.Calendar,
	calendarDates []models.CalendarDate,
	allDepartures []models.Departure,
	trips []models.Trip,
) []models.Departure {
	return []models.Departure{}
}
