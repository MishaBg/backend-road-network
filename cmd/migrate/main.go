package main

import (
	"lab1-design-backend/internal/app/ds"
	"lab1-design-backend/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.City{},
		&ds.User{},
		&ds.ServiceCalc{},
		&ds.CityForService{},
		&ds.CityInfo{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
