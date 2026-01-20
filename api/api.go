package api

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"git.marceeli.ovh/vectura/vectura-api/models"
	"git.marceeli.ovh/vectura/vectura-api/parser"
	"git.marceeli.ovh/vectura/vectura-api/utils"
	"github.com/gin-gonic/gin"
)

var SupportedCities = utils.LoadCitiesFromYAML("cities.yaml")

var cityData = make(map[string]*models.GTFSData)
var cityDataMutex sync.RWMutex

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

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

func preloadCityData() {
	for _, city := range SupportedCities {
		fmt.Printf("Loading GTFS for %s...\n", city.ID)
		start := time.Now()

		data, err := fetchGTFS(city.URL)
		if err != nil {
			fmt.Printf("Failed to fetch GTFS for %s: %v\n", city.ID, err)
			continue
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

		parser.ClearInterner()

		cityDataMutex.Lock()
		cityData[city.ID] = gtfs
		cityDataMutex.Unlock()

		elapsed := time.Since(start)
		fmt.Printf("Loaded %s in %s\n", city.ID, elapsed)
	}
}

func StartServer() {
	preloadCityData()

	r := gin.Default()

	r.GET("/api/cities", func(c *gin.Context) {
		cities := make([]string, 0)

		for _, city := range SupportedCities {
			cities = append(cities, city.ID)
		}

		c.JSON(http.StatusOK, gin.H{
			"cities": cities,
		})
	})

	r.GET("/api/:city/stops", func(c *gin.Context) {
		cityID := c.Param("city")

		cityDataMutex.RLock()
		data, exists := cityData[cityID]
		cityDataMutex.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"city":  cityID,
			"stops": data.Stops,
		})
	})

	r.GET("/api/:city/routes", func(c *gin.Context) {
		cityID := c.Param("city")
		cityDataMutex.RLock()
		data, exists := cityData[cityID]
		cityDataMutex.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not supported"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"city":   cityID,
			"routes": data.Routes,
		})
	})

	r.GET("/api/:city/trips", func(c *gin.Context) {
		cityID := c.Param("city")
		cityDataMutex.RLock()
		data, exists := cityData[cityID]
		cityDataMutex.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not supported"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"city":  cityID,
			"trips": data.Trips,
		})
	})

	r.GET("/api/:city/departures", func(c *gin.Context) {
		cityID := c.Param("city")
		stopID := c.Query("stop")
		date := c.Query("date")
		cityDataMutex.RLock()
		data, exists := cityData[cityID]
		cityDataMutex.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not supported"})
			return
		}

		if stopID != "" {
			if date != "" {
				parsedDate, err := parseDate(date)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"city":       cityID,
					"date":       date,
					"departures": parser.GetDeparturesForStopOnDate(stopID, parsedDate, data.Calendars, data.CalendarDates, data.Departures, data.Trips),
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"city":       cityID,
					"departures": parser.GetDeparturesForStopToday(stopID, data.Calendars, data.CalendarDates, data.Departures, data.Trips),
				})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"city":       cityID,
				"departures": data.Departures,
			})
		}

	})

	r.GET("/api/:city/shapes", func(c *gin.Context) {
		cityID := c.Param("city")
		cityDataMutex.RLock()
		data, exists := cityData[cityID]
		cityDataMutex.RUnlock()

		shape := c.Query("shape")

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not supported"})
			return
		}

		if shape != "" {
			c.JSON(http.StatusOK, gin.H{
				"city":   cityID,
				"shapes": parser.GetShapeById(shape, data.Shapes),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"city":   cityID,
				"shapes": data.Shapes,
			})
		}

	})

	r.Run()
}
