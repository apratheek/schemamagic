package main

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/apratheek/schemamagic"
)

var log *logger.Entry

func init() {
	logger.SetLevel(logger.InfoLevel)
	log = logger.WithFields(logger.Fields{
	// "time": time.Now().Format("Mon Jan _2 15:04:05 2006"),
	})
}

func main() {
	host := "localhost"
	port := uint16(5432)
	database := "dbName"
	username := "username"
	password := "password_in_string"
	dbConn, err := schemamagic.SetupDB(host, port, database, username, password)

	if err != nil {
		log.Fatal("Database error --> ", err)
	}
	tx, _ := dbConn.Begin()
	table := schemamagic.NewTable(schemamagic.Table{Name: "temp_table", DefaultSchema: "public", Database: database, Tx: tx})

	c1 := schemamagic.NewColumn(schemamagic.Column{Name: "action", Datatype: "text", IsNotNull: true, IsUnique: true})
	c2 := schemamagic.NewColumn(schemamagic.Column{Name: "created_at", Datatype: "bigint", DefaultExists: true, DefaultValue: "400"})
	c3 := schemamagic.NewColumn(schemamagic.Column{Name: "version_description", Datatype: "text", DefaultExists: true, DefaultValue: "'Hello'"})
	c4 := schemamagic.NewColumn(schemamagic.Column{Name: "version_new", Datatype: "bigserial"})
	c5 := schemamagic.NewColumn(schemamagic.Column{Name: "arr", Datatype: "bigint[]", DefaultExists: true, DefaultValue: "array[]::bigint[]"})
	c6 := schemamagic.NewColumn(schemamagic.Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint", IsPrimary: true, IsUnique: true})
	table.Append(c1)
	table.Append(c2)
	table.Append(c3)
	table.Append(c4)
	table.Append(c5)
	table.Append(c6)

	table.DropTable()
	table.Begin()

	commitErr := tx.Commit()
	if commitErr != nil {
		tx.Rollback()
		log.Warningln("Couldn't commit transaction --> error is ", commitErr)
	}

}
