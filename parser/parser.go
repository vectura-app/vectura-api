package parser

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
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

// Optimization: Pre-allocate map to avoid resizing if possible,
// though global maps can be dangerous for memory leaks if not cleared.
var interner = make(map[string]string, 10000)

func ClearInterner() {
	// Re-make rather than range-delete for speed
	interner = make(map[string]string)
}

func intern(s string) string {
	if v, ok := interner[s]; ok {
		return v
	}
	// Make a copy of the string to avoid pinning the underlying CSV buffer
	// if we are using ReuseRecord (which we will).
	sCopy := strings.Clone(s)
	interner[sCopy] = sCopy
	return sCopy
}

func parseCSV(file fs.File, callback func(record []string, idxMap map[string]int)) {
	reader := csv.NewReader(file)

	reader.ReuseRecord = true
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return
	}

	idxMap := make(map[string]int, len(header))
	for i, h := range header {
		idxMap[h] = i // Use raw header name
		// If headers might have BOM or whitespace, trim here:
		// idxMap[strings.TrimSpace(h)] = i
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		gostfu(err)

		callback(record, idxMap)
	}
}

// Helper to safely get a value from the slice using the index map
func getVal(record []string, idxMap map[string]int, key string) string {
	if idx, ok := idxMap[key]; ok && idx < len(record) {
		return record[idx]
	}
	return ""
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseUint(s string) uint8 {
	if s == "" {
		return 0
	}
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
	defer file.Close()

	var stops []models.Stop

	parseCSV(file, func(row []string, idx map[string]int) {
		stop := models.Stop{
			StopId:             intern(getVal(row, idx, "stop_id")),
			StopCode:           intern(getVal(row, idx, "stop_code")),
			StopName:           intern(getVal(row, idx, "stop_name")),
			StopLat:            parseFloat(getVal(row, idx, "stop_lat")),
			StopLon:            parseFloat(getVal(row, idx, "stop_lon")),
			StopUrl:            intern(getVal(row, idx, "stop_url")),
			ZoneId:             intern(getVal(row, idx, "zone_id")),
			ParentStation:      intern(getVal(row, idx, "parent_station")),
			PlatformCode:       intern(getVal(row, idx, "platform_code")),
			WheelchairBoarding: models.Accessibility(parseUint(getVal(row, idx, "wheelchair_boarding"))),
			LocationType:       models.Location(parseUint(getVal(row, idx, "location_type"))),
		}

		stops = append(stops, stop)
	})

	return stops
}

func GetRoutes(data []byte) []models.Route {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("routes.txt")
	defer file.Close()

	var routes []models.Route

	parseCSV(file, func(row []string, idx map[string]int) {
		route := models.Route{
			RouteId:          intern(getVal(row, idx, "route_id")),
			AgencyId:         intern(getVal(row, idx, "agency_id")),
			RouteShortName:   intern(getVal(row, idx, "route_short_name")),
			RouteLongName:    intern(getVal(row, idx, "route_long_name")),
			RouteDescription: intern(getVal(row, idx, "route_desc")),
			RouteType:        models.Type(parseUint(getVal(row, idx, "route_type"))),
			RouteUrl:         intern(getVal(row, idx, "route_url")),
			RouteColor:       intern(getVal(row, idx, "route_color")),
			RouteTextColor:   intern(getVal(row, idx, "route_text_color")),
		}

		routes = append(routes, route)
	})

	return routes
}

func GetTrips(data []byte) []models.Trip {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("trips.txt")
	defer file.Close()

	var trips []models.Trip

	parseCSV(file, func(row []string, idx map[string]int) {
		trip := models.Trip{
			TripId:               intern(getVal(row, idx, "trip_id")),
			RouteId:              intern(getVal(row, idx, "route_id")),
			ServiceId:            intern(getVal(row, idx, "service_id")),
			BlockId:              intern(getVal(row, idx, "block_id")),
			TripHeadsign:         intern(getVal(row, idx, "trip_headsign")),
			TripShortName:        intern(getVal(row, idx, "trip_short_name")),
			DirectionId:          models.Direction(parseUint(getVal(row, idx, "direction_id"))),
			ShapeId:              intern(getVal(row, idx, "shape_id")),
			WheelchairAccessible: models.Accessibility(parseUint(getVal(row, idx, "wheelchair_accessible"))),
			BikeAccessible:       models.Accessibility(parseUint(getVal(row, idx, "bikes_allowed"))),
		}

		trips = append(trips, trip)
	})

	return trips
}

func GetDepartures(data []byte) []models.Departure {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("stop_times.txt")
	defer file.Close()

	var departures []models.Departure

	parseCSV(file, func(row []string, idx map[string]int) {
		departure := models.Departure{
			TripId:        intern(getVal(row, idx, "trip_id")),
			StopId:        intern(getVal(row, idx, "stop_id")),
			ArrivalTime:   intern(getVal(row, idx, "arrival_time")),
			DepartureTime: intern(getVal(row, idx, "departure_time")),
			StopSequence:  parseInt(getVal(row, idx, "stop_sequence")),
			PickupType:    models.PickupOrDropoff(parseUint(getVal(row, idx, "pickup_type"))),
			DropoffType:   models.PickupOrDropoff(parseUint(getVal(row, idx, "drop_off_type"))),
		}

		departures = append(departures, departure)
	})

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
				file.Close()
				return []models.Calendar{}
			}
		}
	}
	defer file.Close()

	var calendars []models.Calendar

	parseCSV(file, func(row []string, idx map[string]int) {
		calendar := models.Calendar{
			ServiceId: intern(getVal(row, idx, "service_id")),
			Monday:    parseBool(getVal(row, idx, "monday")),
			Tuesday:   parseBool(getVal(row, idx, "tuesday")),
			Wednesday: parseBool(getVal(row, idx, "wednesday")),
			Thursday:  parseBool(getVal(row, idx, "thursday")),
			Friday:    parseBool(getVal(row, idx, "friday")),
			Saturday:  parseBool(getVal(row, idx, "saturday")),
			Sunday:    parseBool(getVal(row, idx, "sunday")),
			StartDate: parseDate(getVal(row, idx, "start_date")),
			EndDate:   parseDate(getVal(row, idx, "end_date")),
		}

		calendars = append(calendars, calendar)
	})

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
				file.Close()
				return []models.CalendarDate{}
			}
		}
	}
	defer file.Close()

	var calendarDates []models.CalendarDate

	parseCSV(file, func(row []string, idx map[string]int) {
		calendarDate := models.CalendarDate{
			ServiceId:     intern(getVal(row, idx, "service_id")),
			Date:          parseDate(getVal(row, idx, "date")),
			ExceptionType: models.ExceptionType(parseUint(getVal(row, idx, "exception_type"))),
		}

		calendarDates = append(calendarDates, calendarDate)
	})

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
	defer file.Close()

	var shapes []models.Shape

	parseCSV(file, func(row []string, idx map[string]int) {
		shape := models.Shape{
			ShapeId:         intern(getVal(row, idx, "shape_id")),
			ShapePtLat:      parseFloat(getVal(row, idx, "shape_pt_lat")),
			ShapePtLon:      parseFloat(getVal(row, idx, "shape_pt_lon")),
			ShapePtSequence: parseInt(getVal(row, idx, "shape_pt_sequence")),
		}

		shapes = append(shapes, shape)
	})

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
			if date.Before(cal.StartDate) || date.After(cal.EndDate) {
				continue
			}
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
