package main

import (
	logger "github.com/Unaxiom/ulogger"
	"github.com/apratheek/schemamagic"
)

var log *logger.Logger

func init() {
	log = logger.New()
	log.SetLogLevel(logger.InfoLevel)
}

func main() {
	host := "localhost"
	port := uint16(5432)
	database := "play"
	username := "username"
	password := "password_in_string"
	dbConn, err := schemamagic.SetupDB(host, port, database, username, password)

	if err != nil {
		log.Fatalln("Database error --> ", err)
	}
	tx, _ := dbConn.Begin()
	table := schemamagic.NewTable(schemamagic.Table{Name: "temp_table", DefaultSchema: "public", Database: database, Tx: tx})

	c1 := schemamagic.NewColumn(schemamagic.Column{Name: "action", Datatype: "text", IsNotNull: true, IsUnique: true, IndexRequired: true})
	c2 := schemamagic.NewColumn(schemamagic.Column{Name: "created_at", Datatype: "bigint", DefaultExists: true, DefaultValue: "400"})
	c3 := schemamagic.NewColumn(schemamagic.Column{Name: "version_description", Datatype: "text", DefaultExists: false}) // , DefaultValue: "'Hello'"
	c4 := schemamagic.NewColumn(schemamagic.Column{Name: "version_new", Datatype: "bigserial"})
	c5 := schemamagic.NewColumn(schemamagic.Column{Name: "arr", Datatype: "bigint[]", DefaultExists: true, DefaultValue: "array[]::bigint[]"})
	c6 := schemamagic.NewColumn(schemamagic.Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint"})
	c7 := schemamagic.NewColumn(schemamagic.Column{Name: "indexed_col", Datatype: "text", DefaultExists: true, DefaultValue: "''", IndexRequired: true})
	c8 := schemamagic.NewColumn(schemamagic.Column{Name: "id", Datatype: "bigserial", IsPrimary: true, IsUnique: true})
	c9 := schemamagic.NewColumn(schemamagic.Column{Name: "id2", Datatype: "bigserial", IsUnique: true, SequenceRestart: 2000})
	table.Append(c1)
	table.Append(c2)
	table.Append(c3)
	table.Append(c4)
	table.Append(c5)
	table.Append(c6)
	table.Append(c7)
	table.Append(c8)
	table.Append(c9)
	compositeKey := schemamagic.Constraint{
		Name:  "new_id",
		Value: "UNIQUE (action, version_description)",
	}
	table.AddConstraint(compositeKey)

	// table.DropTable()
	// log.Infoln("Dropping table here...")
	table.Begin()
	log.Infoln("Starting table creation here...")

	commitErr := tx.Commit()
	if commitErr != nil {
		tx.Rollback()
		log.Warningln("Couldn't commit transaction --> error is ", commitErr)
	}

}
