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

		gtfs := &models.GTFSData{
			Stops:         parser.GetStops(data),
			Routes:        parser.GetRoutes(data),
			Trips:         parser.GetTrips(data),
			Departures:    parser.GetDepartures(data),
			Calendars:     parser.GetCalendar(data),
			CalendarDates: parser.GetCalendarDates(data),
			Shapes:        parser.GetShapes(data),
		}

		var dbStops []models.DbStop
		for _, stop := range gtfs.Stops {
			dbStops = append(dbStops, models.StopToDbStop(stop, city.ID))
		}

		var dbRoutes []models.DbRoute
		for _, route := range gtfs.Routes {
			dbRoutes = append(dbRoutes, models.RouteToDbRoute(route, city.ID))
		}

		var dbTrips []models.DbTrip
		for _, trip := range gtfs.Trips {
			dbTrips = append(dbTrips, models.TripToDbTrip(trip, city.ID))
		}

		var dbDepartures []models.DbDeparture
		for _, dep := range gtfs.Departures {
			dbDepartures = append(dbDepartures, models.DepartureToDbDeparture(dep, city.ID))
		}

		var dbCalendars []models.DbCalendar
		for _, cal := range gtfs.Calendars {
			dbCalendars = append(dbCalendars, models.CalendarToDbCalendar(cal, city.ID))
		}

		var dbCalendarDates []models.DbCalendarDate
		for _, cd := range gtfs.CalendarDates {
			dbCalendarDates = append(dbCalendarDates, models.CalendarDateToDbCalendarDate(cd, city.ID))
		}

		var dbShapes []models.DbShape
		for _, shape := range gtfs.Shapes {
			dbShapes = append(dbShapes, models.ShapeToDbShape(shape, city.ID))
		}

		if len(dbStops) > 0 {
			db.CreateInBatches(dbStops, 150)
		}
		if len(dbRoutes) > 0 {
			db.CreateInBatches(dbRoutes, 150)
		}
		if len(dbTrips) > 0 {
			db.CreateInBatches(dbTrips, 150)
		}
		if len(dbDepartures) > 0 {
			db.CreateInBatches(dbDepartures, 150)
		}
		if len(dbCalendars) > 0 {
			db.CreateInBatches(dbCalendars, 150)
		}
		if len(dbCalendarDates) > 0 {
			db.CreateInBatches(dbCalendarDates, 150)
		}
		if len(dbShapes) > 0 {
			db.CreateInBatches(dbShapes, 150)
		}

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
