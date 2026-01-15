package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"git.marceeli.ovh/vectura/vectura-api/api"
	"git.marceeli.ovh/vectura/vectura-api/parser"
)

func gostfu(err error) {
	if err != nil {
		panic(err)
	}
}

func fetchGTFS(url string) ([]byte, error) {
	resp, err := http.Get(url)

	gostfu(err)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)

	gostfu(err)

	return data, nil
}

func testParser() {
	start := time.Now()
	data, err := fetchGTFS("https://mkuran.pl/gtfs/warsaw.zip")
	if err != nil {
		panic(err)
	}
	elapsed := time.Since(start)
	fmt.Printf("Fetching took %s\n", elapsed)

	start = time.Now()

	_ = parser.GetStops(data)
	parser.GetRoutes(data)
	trips := parser.GetTrips(data)
	departures := parser.GetDepartures(data)
	calendars := parser.GetCalendar(data)
	calendarDates := parser.GetCalendarDates(data)

	elapsed = time.Since(start)
	fmt.Printf("Mapping took %s\n", elapsed)

	start = time.Now()

	dep := parser.GetDeparturesForStopToday("701301", calendars, calendarDates, departures, trips)

	elapsed = time.Since(start)
	fmt.Printf("Query took %s\n", elapsed)

	fmt.Println(dep[:5])
}

func main() {
	api.StartServer()
}
