package schemamagic

import (
	"context"
	"fmt"

	logger "github.com/Unaxiom/ulogger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var log *logger.Logger

func init() {
	// logger.SetLevel(logger.InfoLevel)
	log = logger.New()
	log.SetLogLevel("info")
}

// SetupDB establishes the connection between the postgres server
func SetupDB(ctx context.Context, dbHost string, port uint16, database string, username string, password string) (*pgxpool.Pool, error) {
	log.Infoln("Establishing connection with the database...")
	var maxConn = 40
	var connectionString = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?pool_max_conns=%d", username, password, dbHost, port, database, maxConn)
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Fatalln("Couldn't parse config: ", err)
	}

	config.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, `;`)
		return err
	}

	db, dbErr := pgxpool.New(ctx, config.ConnString())
	if dbErr == nil {
		log.Infoln("Successfully connected to the database...")
	}
	return db, dbErr

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
