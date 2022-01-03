package main

import (
	"errors"
	"log"
	"os"
	"time"

	_ "github.com/docker/go-healthcheck"
	health "github.com/docker/go-healthcheck"
	"github.com/sirupsen/logrus"

	dbclient "patelmaulik.com/maulik/v1/dbclient"
	service "patelmaulik.com/maulik/v1/services"
)

var appName = "GoServiceSvc"
var checkFileName = "/tmp/disable"

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v\n", appName)

	/*
		if _, err := os.Create(checkFileName); err != nil {
			log.Fatalf("Startup error : Couldn't create check file: %v", checkFileName)
		}
	*/

	d := Initialise()
	h := service.AccountHandler{}
	h.DBClient = d

	health.Register("healthchecker", health.PeriodicChecker(DatabaseCheck(d), time.Second*5))
	health.Register("fileChecker", health.PeriodicThresholdChecker(MyFileChecker("/tmp/disable"), time.Second*5, 2))
	//health.Register("fileChecker", health.PeriodicChecker(checks.FileChecker("/tmp/disable"), time.Second*5))

	// an := negroni.New(negroni.HandlerFunc(mw.HandlerWithNext), negroni.Wrap(ar))

	logger := log.New(os.Stderr, "logger: ", log.Lshortfile)

	server := service.NewServer()

	server.StartWebServer("8080", &h, logger)
}

// DatabaseCheck - check connection
func DatabaseCheck(db *dbclient.DatabaseRepository) health.Checker {
	return health.CheckFunc(func() error {
		if cnn := db.CheckConection(); cnn == false {
			return errors.New("Database connection lost!")
		}
		return nil
	})
}

// MyFileChecker - if the file exists.
func MyFileChecker(f string) health.Checker {
	return health.CheckFunc(func() error {
		if _, err := os.Stat(f); err == nil {
			return errors.New("file exists")
		}
		return nil
	})
}

// Initialise - Seed data
func Initialise() *dbclient.DatabaseRepository {
	DatabaseClient := dbclient.DatabaseRepository{}
	DatabaseClient.OpenDatabase()
	DatabaseClient.SeedDatabase()
	return &DatabaseClient // dbclient.IDatabaseRepository(DatabaseClient)
	// return &DatabaseClient // dbclient.IDatabaseRepository(DatabaseClient)
}
