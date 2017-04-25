package schemamagic

import (
	logger "github.com/Unaxiom/ulogger"
	"gopkg.in/jackc/pgx.v2"
)

var log *logger.Logger

func init() {
	// logger.SetLevel(logger.InfoLevel)
	log = logger.New()
	log.SetLogLevel("info")
}

// SetupDB establishes the connection between the postgres server
func SetupDB(dbHost string, port uint16, database string, username string, password string) (*pgx.ConnPool, error) {
	log.Infoln("Establishing connection with the database...")
	dbConn, dbErr := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     dbHost,
			Port:     port,
			Database: database,
			User:     username,
			Password: password,
		},
		MaxConnections: 10,
	})
	if dbErr == nil {
		log.Infoln("Successfully connected to the database...")
	}
	return dbConn, dbErr
}

// SetLogLevel sets the log level of the output log
func SetLogLevel(level string) {
	log.Infoln("Firing up...")
	log.Infoln("Setting log to ", level)
	if level == "info" {
		log.SetLogLevel("info")
	} else if level == "warn" {
		log.SetLogLevel("warning")
	} else if level == "debug" {
		log.SetLogLevel("debug")
	}
}
