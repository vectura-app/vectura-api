package parser

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"io/fs"
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

func parseCSVChunked(file fs.File, batchSize int, callback func(records [][]string, idxMap map[string]int)) {
	reader := csv.NewReader(file)

	reader.ReuseRecord = true
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return
	}

	idxMap := make(map[string]int, len(header))
	for i, h := range header {
		idxMap[h] = i
	}

	var batch [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		gostfu(err)

		// Make a copy of the record since ReuseRecord is true
		recordCopy := make([]string, len(record))
		copy(recordCopy, record)

		batch = append(batch, recordCopy)
		if len(batch) >= batchSize {
			callback(batch, idxMap)
			batch = make([][]string, 0, batchSize)
		}
	}

	if len(batch) > 0 {
		callback(batch, idxMap)
	}
}

func parseCSV(file fs.File, callback func(record []string, idxMap map[string]int)) {
	reader := csv.NewReader(file)

	reader.ReuseRecord = true
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return
	}

	idxMap := make(map[string]int, len(header))
	for i, h := range header {
		// Remove BOM from first column if present
		if i == 0 {
			h = strings.TrimPrefix(h, "\ufeff")
		}
		idxMap[strings.TrimSpace(h)] = i
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

func ProcessDeparturesChunked(data []byte, batchSize int, callback func(departures []models.Departure)) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	gostfu(err)

	file, _ := zipReader.Open("stop_times.txt")
	defer file.Close()

	parseCSVChunked(file, batchSize, func(records [][]string, idx map[string]int) {
		var departures []models.Departure
		for _, row := range records {
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
		}
		callback(departures)
	})
}
