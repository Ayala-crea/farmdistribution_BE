package config

import (
	"app_farm_be/helper/atdb"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var MongoString = os.Getenv("MONGODB_URI")
var PostgresString = os.Getenv("POSTGRESSTRING")

var (
	Mongoconn, ErrorMongoconn = atdb.MongoConnect(atdb.DBInfo{
		DBString: MongoString,
		DBName:   "gobizdev",
	})

	PostgresDB *gorm.DB
)

func init() {
	if ErrorMongoconn != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", ErrorMongoconn)
	} else {
		fmt.Println("Successfully connected to MongoDB!")
	}

	var err error
	PostgresDB, err = gorm.Open(postgres.Open(PostgresString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL with GORM: %v", err)
	} else {
		fmt.Println("Successfully connected to PostgreSQL with GORM!")
	}
}
