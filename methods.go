package schemamagic

import (
	logger "github.com/Sirupsen/logrus"
	"gopkg.in/jackc/pgx.v2"
)

var log *logger.Entry

func init() {
	logger.SetLevel(logger.InfoLevel)
	log = logger.WithFields(logger.Fields{
	// "time": time.Now().Format("Mon Jan _2 15:04:05 2006"),
	})
}

// SetupDB establishes the connection between the postgres server
func SetupDB(dbHost string, port uint16, database string, username string, password string) (*pgx.ConnPool, error) {
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
	return dbConn, dbErr
}
