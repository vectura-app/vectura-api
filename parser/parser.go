package parser

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.marceeli.ovh/vectura/vectura-api/models"
)

func parseDate(s string) time.Time {
	t, _ := time.ParseInLocation("20060102", s, time.Local)
	return t
}

func parseBool(s string) bool {
	return s == "1"
}

func gostfu(err error) {
	if err != nil {
		panic(err)
	}
}

func parseCSV(file fs.File) []map[string]string {
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	gostfu(err)

	if len(records) == 0 {
		return nil // or return []map[string]string{} if you prefer an empty slice
	}

	header := records[0]
	var csvfile []map[string]string

	for _, record := range records[1:] {
		row := make(map[string]string)
		for i, value := range record {
			row[header[i]] = value
		}
		csvfile = append(csvfile, row)
	}

	return csvfile
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseUint(s string) uint8 {
	v, _ := strconv.ParseUint(s, 10, 8)
	return uint8(v)
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseGTFSTime(s string) int64 {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.Atoi(parts[2])

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return midnight.Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second).Unix()
}

func GetStops(data []byte) []models.Stop {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("stops.txt")

	parsedfile := parseCSV(file)

	var stops []models.Stop

	for _, row := range parsedfile {
		stop := models.Stop{
			StopId:             row["stop_id"],
			StopCode:           row["stop_code"],
			StopName:           row["stop_name"],
			StopLat:            parseFloat(row["stop_lat"]),
			StopLon:            parseFloat(row["stop_lon"]),
			StopUrl:            row["stop_url"],
			ZoneId:             row["zone_id"],
			ParentStation:      row["parent_station"],
			PlatformCode:       row["platform_code"],
			WheelchairBoarding: models.Accessibility(parseUint(row["wheelchair_boarding"])),
			LocationType:       models.Location(parseUint(row["location_type"])),
		}

		stops = append(stops, stop)
	}

	return stops
}

func GetRoutes(data []byte) []models.Route {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("routes.txt")

	parsedfile := parseCSV(file)

	var routes []models.Route

	for _, row := range parsedfile {
		route := models.Route{
			RouteId:          row["route_id"],
			AgencyId:         row["agency_id"],
			RouteShortName:   row["route_short_name"],
			RouteLongName:    row["route_long_name"],
			RouteDescription: row["route_desc"],
			RouteType:        models.Type(parseUint(row["route_type"])),
			RouteUrl:         row["route_url"],
			RouteColor:       row["route_color"],
			RouteTextColor:   row["route_text_color"],
		}

		routes = append(routes, route)
	}

	return routes
}

func GetTrips(data []byte) []models.Trip {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("trips.txt")

	parsedfile := parseCSV(file)

	var trips []models.Trip

	for _, row := range parsedfile {
		trip := models.Trip{
			TripId:               row["trip_id"],
			RouteId:              row["route_id"],
			ServiceId:            row["service_id"],
			BlockId:              row["block_id"],
			TripHeadsign:         row["trip_headsign"],
			TripShortName:        row["trip_short_name"],
			DirectionId:          models.Direction(parseUint(row["direction_id"])),
			ShapeId:              row["shape_id"],
			WheelchairAccessible: models.Accessibility(parseUint(row["wheelchair_accessible"])),
			BikeAccessible:       models.Accessibility(parseUint(row["bikes_allowed"])),
		}

		trips = append(trips, trip)
	}

	return trips
}

func GetDepartures(data []byte) []models.Departure {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("stop_times.txt")

	parsedfile := parseCSV(file)

	var departures []models.Departure

	for _, row := range parsedfile {
		departure := models.Departure{
			TripId:        row["trip_id"],
			StopId:        row["stop_id"],
			ArrivalTime:   row["arrival_time"],
			DepartureTime: row["departure_time"],
			StopSequence:  parseInt(row["stop_sequence"]),
			PickupType:    models.PickupOrDropoff(parseUint(row["pickup_type"])),
			DropoffType:   models.PickupOrDropoff(parseUint(row["drop_off_type"])),
		}

		departures = append(departures, departure)
	}

	return departures
}

func GetCalendar(data []byte) []models.Calendar {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, err := zipReader.Open("calendar.txt")
	if err != nil {
		if err.Error() == "open calendar.txt: file does not exist" {
			file, err = zipReader.Open("calendar_dates.txt")

			if err != nil {
				panic("no calendar or calendar dates")
			} else {
				return []models.Calendar{}
			}
		}
	}

	parsedfile := parseCSV(file)

	var calendars []models.Calendar

	for _, row := range parsedfile {
		calendar := models.Calendar{
			ServiceId: row["service_id"],
			Monday:    parseBool(row["monday"]),
			Tuesday:   parseBool(row["tuesday"]),
			Wednesday: parseBool(row["wednesday"]),
			Thursday:  parseBool(row["thursday"]),
			Friday:    parseBool(row["friday"]),
			Saturday:  parseBool(row["saturday"]),
			Sunday:    parseBool(row["sunday"]),
			StartDate: parseDate(row["start_date"]),
			EndDate:   parseDate(row["end_date"]),
		}

		calendars = append(calendars, calendar)
	}

	return calendars
}

func GetCalendarDates(data []byte) []models.CalendarDate {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, err := zipReader.Open("calendar_dates.txt")
	if err != nil {
		if err.Error() == "open calendar_dates.txt: file does not exist" {
			file, err = zipReader.Open("calendar.txt")

			if err != nil {
				panic("no calendar or calendar dates")
			} else {
				return []models.CalendarDate{}
			}
		}
	}

	parsedfile := parseCSV(file)

	var calendarDates []models.CalendarDate

	for _, row := range parsedfile {
		calendarDate := models.CalendarDate{
			ServiceId:     row["service_id"],
			Date:          parseDate(row["date"]),
			ExceptionType: models.ExceptionType(parseUint(row["exception_type"])),
		}

		calendarDates = append(calendarDates, calendarDate)
	}

	return calendarDates
}

func GetShapes(data []byte) []models.Shape {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, err := zipReader.Open("shapes.txt")
	if err != nil {
		if err.Error() == "open shapes.txt: file does not exist" {
			return []models.Shape{}
		} else {
			panic(err)
		}
	}

	parsedfile := parseCSV(file)

	var shapes []models.Shape

	for _, row := range parsedfile {
		shape := models.Shape{
			ShapeId:         row["shape_id"],
			ShapePtLat:      parseFloat(row["shape_pt_lat"]),
			ShapePtLon:      parseFloat(row["shape_pt_lon"]),
			ShapePtSequence: parseInt(row["shape_pt_sequence"]),
		}

		shapes = append(shapes, shape)
	}

	return shapes
}

func GetShapeById(id string, shapes []models.Shape) []models.Shape {
	shps := make([]models.Shape, 0)

	for _, shape := range shapes {
		if shape.ShapeId == id {
			shps = append(shps, shape)
		}
	}

	sort.Slice(shps, func(i, j int) bool {
		return shps[i].ShapePtSequence < shps[j].ShapePtSequence
	})

	return shps
}

func GetDeparturesForStop(allDepartures []models.Departure, stop string) []models.Departure {
	deps := make([]models.Departure, 0)

	for _, departure := range allDepartures {
		if departure.StopId == stop {
			deps = append(deps, departure)
		}
	}

	return deps
}

func GetActiveServicesForDate(date time.Time, calendars []models.Calendar, calendarDates []models.CalendarDate) map[string]bool {
	activeServices := make(map[string]bool)

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	if calendars != nil {
		for _, cal := range calendars {
			// Skip if date is outside service range
			if date.Before(cal.StartDate) || date.After(cal.EndDate) {
				continue
			}

			// Check weekday match
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
	}

	if calendarDates != nil {
		for _, cd := range calendarDates {
			cdDate := time.Date(cd.Date.Year(), cd.Date.Month(), cd.Date.Day(), 0, 0, 0, 0, cd.Date.Location())
			if !cdDate.Equal(date) {
				continue
			}

			switch cd.ExceptionType {
			case models.SERVICE_ADDED:
				activeServices[cd.ServiceId] = true
			case models.SERVICE_REMOVED:
				delete(activeServices, cd.ServiceId)
			}
		}
	}

	return activeServices
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
	activeServices := GetActiveServicesForDate(date, calendars, calendarDates)

	tripToService := make(map[string]string)
	for _, trip := range trips {
		tripToService[trip.TripId] = trip.ServiceId
	}

	everyDeparture := GetDeparturesForStop(allDepartures, stop)

	var departures []models.Departure
	for _, dep := range everyDeparture {
		serviceId, exists := tripToService[dep.TripId]
		if !exists {
			continue
		}

		if activeServices[serviceId] {
			departures = append(departures, dep)
		}
	}

	return departures
}
