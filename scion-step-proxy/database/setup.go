package database

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Each driver should be put into a dedicated function to make it easier to maintain
// Connect to postgres DB via environment variables
func connectPostgres() (*gorm.DB, error) {
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbName := os.Getenv("PDATABASE_DATABASE")
	dbUser := os.Getenv("DATABASE_USERNAME")
	dbPassword := os.Getenv("DATABASE_PASSWORD")

	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		dbHost,
		dbPort,
		dbUser,
		dbName,
		dbPassword)

	db, err := gorm.Open(postgres.Open(connectionString))
	if err != nil {
		logrus.Error("Failed to connect to postres:", err)
		return nil, err
	}
	return db, nil
}

// Connect to sqlite DB via environment variables, create sqlite file if required
func connectSqlite() (*gorm.DB, error) {
	dbLocation := os.Getenv("DATABASE_PATH")
	if dbLocation == "" {
		dbLocation = "/opt/auth-service/gorm.db"
	}

	// Create the sqlite file if it's not available
	if _, err := os.Stat(dbLocation); err != nil {
		if _, err = os.Create(dbLocation); err != nil {
			logrus.Error("Failed to create sqlite db at: ", dbLocation)
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(dbLocation), &gorm.Config{})
	return db, err
}

// Ensure the database has all tables and keys set up properly
func migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

// Connects to the configred database and creates tables
func InitializeDatabaseLayer() (*gorm.DB, error) {

	dbs := os.Getenv("DATABASE")
	var db *gorm.DB
	var err error

	switch dbs {
	case "sqlite":
		db, err = connectSqlite()
		break
	case "postgres":
		db, err = connectPostgres()
		break
	default:
		return nil, fmt.Errorf("No database configured, set the DATABASE env to proceed")
	}

	if err != nil {
		return nil, err
	}

	err = migrate(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
