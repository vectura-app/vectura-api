package api

import (
	"net/http"
	"slices"
	"sync"
	"time"

	"git.marceeli.ovh/vectura/vectura-api/database"
	"git.marceeli.ovh/vectura/vectura-api/models"
	"git.marceeli.ovh/vectura/vectura-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var SupportedCities = utils.LoadCitiesFromYAML("cities.yaml")
var SCIdx = utils.GetCityIDIndex("cities.yaml")

var cityData = make(map[string]*models.GTFSData)
var cityDataMutex sync.RWMutex

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

func StartServer(db *gorm.DB) {
	r := gin.Default()

	r.GET("/api/cities", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"cities": SCIdx,
		})
	})

	r.GET("/api/:city/stops", func(c *gin.Context) {
		cityID := c.Param("city")

		exists := slices.Contains(SCIdx, cityID)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"city":  cityID,
			"stops": database.GetStops(db, cityID),
		})
	})

	r.GET("/api/:city/routes", func(c *gin.Context) {
		cityID := c.Param("city")

		exists := slices.Contains(SCIdx, cityID)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not supported"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"city":   cityID,
			"routes": database.GetRoutes(db, cityID),
		})
	})

	r.GET("/api/:city/trips", func(c *gin.Context) {
		cityID := c.Param("city")

		exists := slices.Contains(SCIdx, cityID)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not supported"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"city":  cityID,
			"trips": database.GetTrips(db, cityID),
		})
	})

	r.GET("/api/:city/departures", func(c *gin.Context) {
		cityID := c.Param("city")
		stopID := c.Query("stop")
		date := c.Query("date")

		exists := slices.Contains(SCIdx, cityID)
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
					"departures": database.GetDeparturesForStopOnDate(db, cityID, stopID, parsedDate),
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"city":       cityID,
					"departures": database.GetDeparturesForStopToday(db, cityID, stopID),
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"city":  cityID,
				"error": "You need to specify a stop!",
			})
		}

	})

	r.GET("/api/:city/shapes", func(c *gin.Context) {
		cityID := c.Param("city")
		shape := c.Query("shape")

		exists := slices.Contains(SCIdx, cityID)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not supported"})
			return
		}

		if shape != "" {
			c.JSON(http.StatusOK, gin.H{
				"city":   cityID,
				"shapes": database.GetShapeById(db, cityID, shape),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"city":  cityID,
				"error": "You need to specify a shape!",
			})
		}

	})

	r.Run()
}
