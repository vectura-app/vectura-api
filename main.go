package main

import (
	"os"

	"git.marceeli.ovh/vectura/vectura-api/api"
	"git.marceeli.ovh/vectura/vectura-api/database"
	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func loadDSN() string {
	enverr := godotenv.Load()

	if enverr != nil {
		if enverr.Error() == "open .env: no such file or directory" {
			// there is no .env so don't give a fuck
		} else {
			panic(enverr)
		}
	}

	return os.Getenv("DSN")
}

func loadDB(dsn string) *gorm.DB {
	var (
		db  *gorm.DB
		err error
	)

	if dsn != "" {
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN: dsn,
		}), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   logger.Default.LogMode(logger.Error),
		})
	} else {
		db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
	}

	if err != nil {
		panic(err)
	}

	return db
}
func main() {
	dsn := loadDSN()
	db := loadDB(dsn)

	database.PreloadCities(db)
	api.StartServer(db)
}
