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
