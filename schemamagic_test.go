package schemamagic

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"gopkg.in/jackc/pgx.v2"
)

func returnAllTables(tx *pgx.Tx) []*Table {
	tables := make([]*Table, 0)
	tables = append(tables, tableTaxParams(tx))
	tables = append(tables, tableFormsSections(tx))
	return tables
}

// taxParam stores the info of every tax param in the tax_params DB table
type taxParam struct {
	ID                    int64   `json:"id"`
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	TaxRatePercentage     float64 `json:"tax_rate_percentage"`
	InputCreditPercentage float64 `json:"input_credit_percentage"`
	AddedBy               string  `json:"added_by"`
	Active                bool    `json:"active"`
	Timestamp             int64   `json:"timestamp"`
}

func tableTaxParams(tx *pgx.Tx) *Table {
	/*
		CREATE TABLE tax_params (
			id bigserial UNIQUE,
			name text UNIQUE,
			description text DEFAULT '',
			tax_rate_percentage double precision DEFAULT 0.00,
			input_credit_percentage double precision DEFAULT 0.00, -- percentage of input credit applicable -> if all the tax has an input available, then this value is 100. If only 50% of this type of tax can be availed as input credit, its value should be 50
			added_by text DEFAULT '', // Stores the username of the person who added this
			active boolean DEFAULT true,
		    timestamp bigint DEFAULT EXTRACT(EPOCH FROM NOW())::bigint -- timestamp of when this notification was generated
		)
	*/
	table := NewTable(Table{Name: "tax_params", DefaultSchema: "public", Database: "schemamagic", Tx: tx})
	table.Append(NewColumn(Column{Name: "id", Datatype: "bigserial", IsPrimary: true, SequenceRestart: 101}))
	table.Append(NewColumn(Column{Name: "name", Datatype: "text", IsUnique: true}))
	table.Append(NewColumn(Column{Name: "description", Datatype: "text", DefaultExists: true, DefaultValue: "''"}))
	table.Append(NewColumn(Column{Name: "tax_rate_percentage", Datatype: "double precision", DefaultExists: true, DefaultValue: "0.00"}))
	/*
		input_credit_percentage is the percentage of tax of this type that can be availed as input credit. If the tax_rate_percentage is 12%, and if all of the amount can be availed,
		then this value is stored as 100. If only 25% of this tax can be availed as input credit, the value is stored as 25
	*/
	table.Append(NewColumn(Column{Name: "input_credit_percentage", Datatype: "double precision", DefaultExists: true, DefaultValue: "0.00"}))
	table.Append(NewColumn(Column{Name: "added_by", Datatype: "text", DefaultExists: true, DefaultValue: "''"}))
	table.Append(NewColumn(Column{Name: "active", Datatype: "boolean", DefaultExists: true, DefaultValue: "true"}))
	table.Append(NewColumn(Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint"}))

	return table
}

// formSection stores the info of every form section in the forms_sections DB table
type formSection struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Width       string `json:"width"`
	Active      bool   `json:"active"`
	Timestamp   int64  `json:"timestamp"`
}

func tableFormsSections(tx *pgx.Tx) *Table {
	/*
		CREATE TABLE forms_sections (
			id bigserial UNIQUE,
			type text NOT NULL DEFAULT '', -- invoice-form, client-form, purchase-form, etc
			name text NOT NULL DEFAULT '',
			description text DEFAULT '',
			width text NOT NULL DEFAULT '',
			active boolean DEFAULT true,
		    timestamp bigint DEFAULT EXTRACT(EPOCH FROM NOW())::bigint -- timestamp of when this notification was generated
		)
	*/
	table := NewTable(Table{Name: "forms_sections", DefaultSchema: "public", Database: "schemamagic", Tx: tx})
	table.Append(NewColumn(Column{Name: "id", Datatype: "bigserial", IsPrimary: true}))
	table.Append(NewColumn(Column{Name: "type", Datatype: "text", IsNotNull: true, DefaultExists: true, DefaultValue: "''", IndexRequired: true}))
	table.Append(NewColumn(Column{Name: "name", Datatype: "text", IsNotNull: true, DefaultExists: true, DefaultValue: "''"}))
	table.Append(NewColumn(Column{Name: "description", Datatype: "text", DefaultExists: true, DefaultValue: "''"}))
	table.Append(NewColumn(Column{Name: "width", Datatype: "text", IsNotNull: true, DefaultExists: true, DefaultValue: "''"}))
	table.Append(NewColumn(Column{Name: "active", Datatype: "boolean", DefaultExists: true, DefaultValue: "true"}))
	table.Append(NewColumn(Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint"}))
	return table

}

func createTables(dbConn *pgx.ConnPool, assert *require.Assertions) {
	tx := fetchTx(dbConn, assert)
	// Fetch the list of tables
	tablesList := returnAllTables(tx)
	for _, table := range tablesList {
		table.Begin()
	}
	assert.Nil(tx.Commit())
}

func deleteTables(dbConn *pgx.ConnPool, assert *require.Assertions) {
	tx := fetchTx(dbConn, assert)
	// Fetch the list of tables
	tablesList := returnAllTables(tx)
	for _, table := range tablesList {
		table.DropTable()
	}
	assert.Nil(tx.Commit())
}

func fetchTx(dbConn *pgx.ConnPool, assert *require.Assertions) *pgx.Tx {
	tx, err := dbConn.Begin()
	assert.Nil(err)
	return tx
}

func insertDataIntoTables(dbConn *pgx.ConnPool, assert *require.Assertions) (taxParam, formSection) {
	var err error
	param := taxParam{
		Name:                  "Some tax " + uuid.NewV4().String(),
		Description:           "Description about this tax",
		TaxRatePercentage:     18.94,
		InputCreditPercentage: 100.0,
		AddedBy:               "self",
		Active:                true,
	}
	section := formSection{
		Type:        "section-type",
		Name:        "Section Name",
		Description: "Description about this Section",
		Width:       "6",
		Active:      true,
	}
	tx := fetchTx(dbConn, assert)

	err = tx.QueryRow(`
		INSERT INTO tax_params (name, description, tax_rate_percentage, input_credit_percentage, added_by, active)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, timestamp
	`, param.Name, param.Description, param.TaxRatePercentage, param.InputCreditPercentage, param.AddedBy, param.Active).Scan(&param.ID, &param.Timestamp)
	assert.Nil(err)

	err = tx.QueryRow(`
		INSERT INTO forms_sections (type, name, description, width, active) 
			VALUES ($1, $2, $3, $4, $5) RETURNING id, timestamp
	`, section.Type, section.Name, section.Description, section.Width, section.Active).Scan(&section.ID, &section.Timestamp)
	assert.Nil(err)

	assert.Nil(tx.Commit())
	return param, section
}

func fetchPreviouslyEnteredData(paramID int64, sectionID int64, dbConn *pgx.ConnPool, assert *require.Assertions) (taxParam, formSection) {
	var param taxParam
	var section formSection
	var err error
	err = dbConn.QueryRow(`
		SELECT id, name, description, tax_rate_percentage, input_credit_percentage, added_by, active, timestamp
			FROM tax_params WHERE id = $1
	`, paramID).Scan(&param.ID, &param.Name, &param.Description, &param.TaxRatePercentage, &param.InputCreditPercentage, &param.AddedBy, &param.Active, &param.Timestamp)
	assert.Nil(err)

	err = dbConn.QueryRow(`
		SELECT id, type, name, description, width, active, timestamp FROM forms_sections WHERE id = $1
	`, sectionID).Scan(&section.ID, &section.Type, &section.Name, &section.Description, &section.Width, &section.Active, &section.Timestamp)
	assert.Nil(err)

	return param, section
}

func assertTaxParams(param1 taxParam, param2 taxParam, assert *require.Assertions) {
	assert.Equal(param1.ID, param2.ID)
	assert.Equal(param1.Name, param2.Name)
	assert.Equal(param1.Description, param2.Description)
	assert.Equal(param1.TaxRatePercentage, param2.TaxRatePercentage)
	assert.Equal(param1.InputCreditPercentage, param2.InputCreditPercentage)
	assert.Equal(param1.AddedBy, param2.AddedBy)
	assert.Equal(param1.Active, param2.Active)
	assert.Equal(param1.Timestamp, param2.Timestamp)
}

func assertFormSections(section1 formSection, section2 formSection, assert *require.Assertions) {
	assert.Equal(section1.ID, section2.ID)
	assert.Equal(section1.Type, section2.Type)
	assert.Equal(section1.Name, section2.Name)
	assert.Equal(section1.Description, section2.Description)
	assert.Equal(section1.Width, section2.Width)
	assert.Equal(section1.Active, section2.Active)
	assert.Equal(section1.Timestamp, section2.Timestamp)
}

func TestSchemamagic(t *testing.T) {
	assert := require.New(t)
	SetLogLevel("info")
	SetLogLevel("warn")
	SetLogLevel("debug")
	// Connect to the database
	dbConn, err := SetupDB("localhost", 5432, "schemamagic", "schemamagic", "schemamagic")
	assert.Nil(err)

	// Create these tables
	createTables(dbConn, assert)
	// Insert data into these tables
	param1, section1 := insertDataIntoTables(dbConn, assert)
	// Create these tables again
	createTables(dbConn, assert)
	// Fetch the previously entered data
	param2, section2 := fetchPreviouslyEnteredData(param1.ID, section1.ID, dbConn, assert)
	// Match that the inserted values still remain in the database
	assertTaxParams(param1, param2, assert)
	assertFormSections(section1, section2, assert)
	// Drop the tables
	deleteTables(dbConn, assert)
}
