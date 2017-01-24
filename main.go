package main

import (
	"errors"
	"fmt"
	"strings"

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

// Column stores all the parameters of each column inside a table
type Column struct {
	Name            string
	Datatype        string
	Action          string
	DefaultExists   bool
	DefaultValue    string
	IsUnique        bool
	IsPrimary       bool
	IsNotNullable   bool
	Comment         string
	SequenceRestart int64
}

// prepareSQLStatement prepares and returns the statement that needs to be executed by the table
func (c *Column) prepareSQLStatement(step int, tableName string) (string, error) {
	log.Infoln("Executing ", c.Name, " with step --> ", step)
	var statement string
	if step == 1 {
		// This is the step where the column is added without a default value
		// statement = "ALTER TABLE %s ADD %s %s"%(table_name, self.column_name, self.datatype)
		statement = fmt.Sprintf("ALTER TABLE %s ADD %s %s", tableName, c.Name, c.Datatype)
	} else if step == 2 {
		// This is the step where a default value is set for the column
		// statement = cursor.mogrify("ALTER TABLE %(table)s ALTER COLUMN %(column)s SET DEFAULT %(value)s", {"table" : AsIs(table_name), "column" : AsIs(self.column_name), "value" : AsIs(self.default_value)})
		if c.DefaultExists {
			statement = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s", tableName, c.Name, c.DefaultValue)
		}
	} else if step == 3 {
		// This is the step where the default value is updated for all the existing rows
		// statement = cursor.mogrify("UPDATE %(table)s SET %(column)s = %(value)s", {"table" : AsIs(table_name), "column" : AsIs(self.column_name), "value" : AsIs(self.default_value)})
		if c.DefaultExists {
			statement = fmt.Sprintf("UPDATE %s SET %s = %s", tableName, c.Name, c.DefaultValue)
		}
	} else if step == 4 {
		// This is the step where the sequence is altered, in case sequence_restart is > 0 and datatype is either bigserial or serial
		// statement = cursor.mogrify("ALTER SEQUENCE %(sequence_name)s RESTART WITH %(value)s", {"sequence_name" : AsIs(sequence_name), "value" : self.sequence_restart})
		if strings.Contains(c.Datatype, "serial") {
			statement = fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH %d", tableName+"_"+c.Name+"_seq", c.SequenceRestart)
		}
	} else if step == 5 {
		// This is the step where a unique constraint is added, in case the column in unique
		if c.IsUnique {
			// statement = "ALTER TABLE %s ADD UNIQUE (%s)"%(table_name, self.column_name)
			statement = fmt.Sprintf("ALTER TABLE %s ADD UNIQUE (%s)", tableName, c.Name)
		}
	} else if step == 6 {
		// This is the step where a primary key constraint is added, in case the column is a primary key
		if c.IsPrimary {
			// statement = "ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY(%s)"%(table_name, table_name + "_" +self.column_name, self.column_name)
			statement = fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY(%s)", tableName, tableName+"_"+c.Name, c.Name)
		}
	} else if step == 7 {
		// This is the step where NOT NULL is applied to a particular column
		if c.IsNotNullable {
			// statement = "ALTER TABLE %s ALTER COLUMN %s SET NOT NULL"%(table_name, self.column_name)
			statement = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, c.Name)
		}
	} else if step == 101 {
		// This is the step where the column's datatype is altered
		if strings.Contains(c.Datatype, "serial") {
			errorStatement := fmt.Sprintf("Can't modify datatype to SERIAL versions, while modifying \nTable --> %s \nColumn --> %s \nDatatype --> %s", tableName, c.Name, c.Datatype)
			err := errors.New(errorStatement)
			return "", err
		}
		// statement = "ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s"%(table_name, self.column_name, self.datatype, self.column_name, altered_datatype)
		statement = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s", tableName, c.Name, c.Datatype, c.Name, c.Datatype)
	}

	log.Infoln("In prepareSQLStatement, statement is \n", statement)
	return statement, nil
}

// NewColumn initializes the Column with the default parameters
func NewColumn(c Column) Column {
	// col := new(Column)
	var col Column
	col.Name = c.Name
	col.Datatype = c.Datatype
	if c.Action == "" {
		col.Action = "Add"
	} else {
		col.Action = c.Action
	}
	col.DefaultExists = c.DefaultExists
	if c.DefaultValue == "" {
		col.DefaultValue = "'NULL'"
	} else {
		col.DefaultValue = c.DefaultValue
	}
	col.IsUnique = c.IsUnique
	col.IsPrimary = c.IsPrimary
	if c.IsNotNullable {
		// Set default IsNotNullable to true
		col.IsNotNullable = true
	} else {
		col.IsNotNullable = false
	}
	col.Comment = c.Comment
	if c.SequenceRestart == 0 {
		col.SequenceRestart = 1
	} else {
		col.SequenceRestart = c.SequenceRestart
	}
	return col
}

// Table holds the table details as well as all the columns inside the table
type Table struct {
	Name          string
	DefaultSchema string
	Database      string
	tx            *pgx.Tx
	Autocommit    bool
	Columns       []Column
}

// NewTable creates and returns an instance of a postgres table
func NewTable(t Table) *Table {
	table := new(Table)
	table.Name = t.Name
	table.DefaultSchema = t.DefaultSchema
	table.Database = t.Database
	table.Autocommit = t.Autocommit
	table.Columns = t.Columns
	table.tx = t.tx
	return table
}

// Append method accepts a column and appends it to the list of columns of the table
func (t *Table) Append(col Column) {
	t.Columns = append(t.Columns, col)
}

// Begin method initiates a DB transaction and checks if table (Name) exists in the DB. If it does, then it calls updateTable(). If it doesn't, it calls createTable(), and then updateTable()"""
func (t *Table) Begin() {
	// Create the schema here
	schemaStatement := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", t.DefaultSchema)
	_, err := t.tx.Exec(schemaStatement)
	if err != nil {
		t.tx.Rollback()
		log.Warningln("Couldn't create schema --> ", t.DefaultSchema, " with error being --> ", err)
	}
	//  Check if table exists in the database
	presence := t.checkTableExistence()
	if !presence {
		// Table does not exist --> need to create it
		t.createTable()
	}
	// Loop over all the available columns and call updateTable() on each column
	for _, col := range t.Columns {
		log.Infoln("-----------------------------------------------")
		t.updateTable(col)
		log.Infoln("-----------------------------------------------")
	}
	if t.Autocommit {
		commitErr := t.tx.Commit()
		if commitErr != nil {
			t.tx.Rollback()
			log.Warningln("Couldn't commit changes to the TABLE --> ", t.Name, " with error being --> ", commitErr)
		}
	}
}

// checkTableExistence returns if the table already exists in the DB
func (t *Table) checkTableExistence() bool {
	var presence bool
	statement := fmt.Sprintf("SELECT exists(select 0 from pg_class where relname = '%s')", t.Name)
	err := t.tx.QueryRow(statement).Scan(&presence)
	if err != nil {
		t.tx.Rollback()
		log.Warningln("While querying for table existence, error is --> ", err)
	}
	log.Infoln("While checking for table existence, presence is ", presence)
	return presence
}

// createTable method creates the table in the particular DB
func (t *Table) createTable() {
	statement := fmt.Sprintf("CREATE TABLE %s()", t.Name)
	err := t.executeSQL(statement)
	if err != nil {
		t.tx.Rollback()
		log.Warningln("While creating table --> ", t.Name, " error is --> ", err)
	}
}

// DropTable method drops the table from the DB
func (t *Table) DropTable() {
	presence := t.checkTableExistence()
	if presence {
		// Drop the table here
		statement := fmt.Sprintf("DROP TABLE %s", t.Name)
		err := t.executeSQL(statement)
		if err != nil {
			t.tx.Rollback()
			log.Warningln("While dropping table --> ", t.Name, " error is --> ", err)
		}
	}
}

// updateTable alters the table by adding a new column to it, passed as the method parameter
func (t *Table) updateTable(col Column) {
	minNumberOfSteps := 0
	maxNumberOfSteps := 7
	columnPresence := t.checkColumnPresence(col.Name)
	if columnPresence {
		log.Infoln("Column --> ", col.Name, " already exists")
		columnDatatypeMatch := t.checkColumnDatatype(col)
		log.Infoln("Column --> ", col.Name, " datatype match value is --> ", columnDatatypeMatch)
		if columnDatatypeMatch {
			// If the datatype matches, pass, and SET minNumberOfSteps = 0 and maxNumberOfSteps = 0, since there should be no further execution, as the column is in its actual state
			minNumberOfSteps = 0
			maxNumberOfSteps = 0
		} else {
			// If the datatype does not match, then step=101, and set minNumberOfSteps to 1, since the step=1 has already been executed in the form of datatype modificaton
			minNumberOfSteps = 1
			statement, statementErr := col.prepareSQLStatement(101, t.Name)
			if statementErr != nil {
				err := t.executeSQL(statement)
				if err != nil {
					t.tx.Rollback()
					log.Warningln("While executing SQL --> \n", statement, "\nerror is ", err)
				}
			}
		}
	} else {
		// Column does not exist
		log.Infoln("Column --> ", col.Name, " does not exist")
		minNumberOfSteps = 0
	}

	//  There are 4 steps involved here.

	// The first step is to call col.prepareSQLStatement with a step=1, which would return an SQL statement that would be used
	// to alter the table structure with a default set to NULL -> this is a very cheap operation,
	//  since the access exclusive lock would be acquired for a short time.

	//  The second step is to call col.prepareSQLStatement with a step=2, which would return an SQL statement that would be used
	//  to set the defaults to Column.default_value. In case Column.default_exists is False, then it returns an empty statement.
	//  This can be checked before executing the statement

	//  The third step is to call col.prepareSQLStatement with a step=3, which would update all the rows in the table with the default
	//  value.

	//  The fourth step is to call col.prepareSQLStatement with a step=3, which would return an SQL statement that would be used to
	//  alter the sequence start, in case the datatype is serial/bigserial. Similar to step=2, if it returns an empty statement,
	//  it either means that the datatype doesn't support a sequence, or the sequence needs to begin at 0.
	for step := minNumberOfSteps; step < maxNumberOfSteps; step++ {
		statement, statementErr := col.prepareSQLStatement(step+1, t.Name)
		log.Infoln("In steps, statement is \n", statement, " and error is ", statementErr)
		if statementErr == nil {
			err := t.executeSQL(statement)
			if err != nil {
				t.tx.Rollback()
				log.Warningln("Statement --> ", statement, " could not be executed because of error --> ", err)
			}
		}
	}
}

// executeSQL executes the SQL query
func (t *Table) executeSQL(sql string) error {
	// var err error
	if sql != "" {
		_, err := t.tx.Exec(sql)
		log.Infoln("Executing Statement --> \n", sql, " and error is ", err)
		return err
	}
	return nil
}

// checkColumnPresence checks if the column name passed is present in the current table
func (t *Table) checkColumnPresence(columnName string) bool {
	var presence bool
	statement := fmt.Sprintf("SELECT EXISTS(SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = '%s' AND table_catalog = '%s' AND column_name = '%s')", t.Name, t.Database, columnName)
	err := t.tx.QueryRow(statement).Scan(&presence)
	if err != nil {
		t.tx.Rollback()
		log.Warningln("In checkColumnPresence, error for table --> ", t.Name, " and Column --> ", columnName, " is ", err)
	}
	return presence
}

// checkColumnDatatype checks if the column datatype of the column name passed is equal to the column datatype present in the table
func (t *Table) checkColumnDatatype(col Column) bool {
	columnName := col.Name
	columnDatatype := col.Datatype
	columnDefault := col.DefaultValue
	var (
		dbDatatype      string
		columnDefaultDB pgx.NullString
		presence        bool
	)
	statement := fmt.Sprintf("SELECT data_type, column_default FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = '%s' AND table_catalog = '%s' AND column_name = '%s'", t.Name, t.Database, columnName)

	err := t.tx.QueryRow(statement).Scan(&dbDatatype, &columnDefaultDB)
	if err != nil {
		t.tx.Rollback()
		log.Warningln("While querying for column data type in table --> ", t.Name, " error is --> ", err)
	}

	log.Infoln("Datatype DB is ", dbDatatype, " and Column Default is ", columnDefaultDB, " and Column Datatype is ", columnDatatype)
	// Check if there are sequences here
	if columnDefaultDB.Valid {
		// This is a sequence
		if columnDatatype == "bigserial" && dbDatatype == "bigint" {
			presence = true
		} else if columnDatatype == "serial" && dbDatatype == "integer" {
			presence = true
		} else if columnDatatype == dbDatatype {
			if columnDefaultDB.String == columnDefault {
				presence = true
			} else {
				presence = false
			}
			// presence = true
		} else {
			presence = false
		}

	} else {
		// This is a non-sequence
		if columnDatatype == dbDatatype {
			presence = true
		} else {
			presence = false
		}
	}

	log.Infoln("In check column datatype, the value of presence is ", presence)
	return presence
}

func (t *Table) commit() {
	commitErr := t.tx.Commit()
	if commitErr != nil {
		t.tx.Rollback()
		log.Warningln("Couldn't commit table modifications --> error is ", commitErr)
	}
}

func main() {
	host := "localhost"
	port := uint16(5432)
	database := "erp"
	username := "erp"
	password := "erppassword123"
	dbConn, err := SetupDB(host, port, database, username, password)
	if err != nil {
		log.Fatal("Database error --> ", err)
	}
	tx, _ := dbConn.Begin()
	table := NewTable(Table{Name: "temp_table", DefaultSchema: "public", Database: database, tx: tx})

	c1 := NewColumn(Column{Name: "action", Datatype: "text", IsNotNullable: true, IsUnique: true})
	c2 := NewColumn(Column{Name: "created_at", Datatype: "bigint", DefaultExists: true, DefaultValue: "400"})
	c3 := NewColumn(Column{Name: "version_description", Datatype: "text", DefaultExists: true, DefaultValue: "'Hello'"})
	c4 := NewColumn(Column{Name: "version_new", Datatype: "bigserial"})
	c5 := NewColumn(Column{Name: "arr", Datatype: "bigint[]", DefaultExists: true, DefaultValue: "array[]::bigint[]"})
	c6 := NewColumn(Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint", IsPrimary: true, IsUnique: true})
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
