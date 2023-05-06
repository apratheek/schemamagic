package schemamagic

import (
	"context"
	"fmt"

	"strings"

	pgx "github.com/jackc/pgx/v5"
	// pgx2 "gopkg.in/jackc/pgx.v2"
)

// Table holds the table details as well as all the columns inside the table
type Table struct {
	Name          string
	DefaultSchema string
	Database      string
	Tx            pgx.Tx
	Autocommit    bool
	Columns       []Column
	constraints   []Constraint
}

// NewTable creates and returns an instance of a postgres table
func NewTable(t Table) *Table {
	table := new(Table)
	table.Name = t.Name
	table.DefaultSchema = t.DefaultSchema
	table.Database = t.Database
	table.Autocommit = t.Autocommit
	table.Columns = t.Columns
	table.Tx = t.Tx
	return table
}

// Append method accepts a column and appends it to the list of columns of the table
func (t *Table) Append(col Column) {
	t.Columns = append(t.Columns, col)
}

// AddConstraint accepts a constraint and appends it to the list of constraints that need to be applied on the table
func (t *Table) AddConstraint(constraint Constraint) {
	t.constraints = append(t.constraints, constraint)
}

// Begin method initiates a DB transaction and checks if table (Name) exists in the DB. If it does, then it calls updateTable(). If it doesn't, it calls createTable(), and then updateTable()"""
func (t *Table) Begin(ctx context.Context) {
	log.Infoln("Operating on table --> ", t.Name)
	// Create the schema here
	schemaStatement := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", t.DefaultSchema)
	_, err := t.Tx.Exec(ctx, schemaStatement)
	if err != nil {
		t.Tx.Rollback(ctx)
		log.Warningln("Couldn't create schema --> ", t.DefaultSchema, " with error being --> ", err)
	}
	//  Check if table exists in the database
	presence := t.checkTableExistence(ctx)
	if !presence {
		// Table does not exist --> need to create it
		t.createTable(ctx)
	}
	// Loop over all the available columns and call updateTable() on each column
	for _, col := range t.Columns {
		log.Debugln("-----------------------------------------------")
		t.updateTable(ctx, col)
		log.Debugln("-----------------------------------------------")
	}

	// Iterate over the available constraints and apply them
	for _, constraint := range t.constraints {
		// 1. drop them first
		dropRule := constraint.createDropRule(t.Name)
		log.Debugln("Constraint drop rule is ", dropRule)
		err := t.executeSQL(ctx, dropRule)
		if err != nil {
			log.Fatalln("While trying to drop constraint rule --> ", dropRule, "\n the error is ", err.Error())
		}
		// 2. add them
		addRule := constraint.createAddRule(t.Name)
		log.Debugln("Constraint add rule is ", addRule)
		err = t.executeSQL(ctx, addRule)
		if err != nil {
			log.Fatalln("While trying to add constraint rule -->  ", addRule, "\n the error is ", err.Error())
		}
	}

	if t.Autocommit {
		t.commit(ctx)
	}
}

// checkTableExistence returns if the table already exists in the DB
func (t *Table) checkTableExistence(ctx context.Context) bool {
	var presence bool
	statement := fmt.Sprintf("SELECT exists(select 0 from pg_class where relname = '%s')", t.Name)
	err := t.Tx.QueryRow(ctx, statement).Scan(&presence)
	if err != nil {
		t.Tx.Rollback(ctx)
		log.Warningln("While querying for table existence, error is --> ", err)
	}
	log.Debugln("While checking for table existence, presence is ", presence)
	return presence
}

// createTable method creates the table in the particular DB
func (t *Table) createTable(ctx context.Context) {
	statement := fmt.Sprintf("CREATE TABLE %s()", t.Name)
	err := t.executeSQL(ctx, statement)
	if err != nil {
		t.Tx.Rollback(ctx)
		log.Warningln("While creating table --> ", t.Name, " error is --> ", err)
	}
}

// DropTable method drops the table from the DB
func (t *Table) DropTable(ctx context.Context) {
	presence := t.checkTableExistence(ctx)
	if presence {
		// Drop the table here
		log.Infoln("Trying to drop table --> ", t.Name)
		statement := fmt.Sprintf("DROP TABLE %s", t.Name)
		err := t.executeSQL(ctx, statement)
		if err != nil {
			t.Tx.Rollback(ctx)
			log.Warningln("While dropping table --> ", t.Name, " error is --> ", err)
		} else {
			log.Infoln("Successfully dropped table --> ", t.Name)
		}
	}
}

// updateTable alters the table by adding a new column to it, passed as the method parameter
func (t *Table) updateTable(ctx context.Context, col Column) {
	var steps = make([]int, 0)
	columnPresence := t.checkColumnPresence(ctx, col.Name)
	if columnPresence {
		log.Debugln("Column --> ", col.Name, " already exists")
		columnDatatypeMatch := t.checkColumnDatatype(ctx, col)
		log.Debugln("Column --> ", col.Name, " datatype match value is --> ", columnDatatypeMatch)
		if columnDatatypeMatch {
			// Do nothing
		} else {
			// If the datatype does not match, then step=101, and set minNumberOfSteps to 1, since the step=1 has already been executed in the form of datatype modificaton
			statement, statementErr := col.prepareSQLStatement(101, t.Name, columnPresence)
			if statementErr != nil {
				err := t.executeSQL(ctx, statement)
				if err != nil {
					t.Tx.Rollback(ctx)
					log.Warningln("While executing SQL --> \n", statement, "\nerror is ", err)
				}
			}
		}
		// Run these steps to check for other updates
		steps = []int{2, 3, 5, 7, 8}
	} else {
		// Column does not exist
		// DefaultExists -> 2
		// DefaultValue -> 3
		// IsUnique -> 5
		// IsPrimary -> 6
		// IsNotNull -> 7
		// IndexRequired -> 8
		log.Debugln("Column --> ", col.Name, " does not exist")
		steps = []int{1, 2, 3, 4, 5, 6, 7, 8}
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
	for _, step := range steps {
		// for step := minNumberOfSteps; step < maxNumberOfSteps; step++ {
		statement, statementErr := col.prepareSQLStatement(step, t.Name, columnPresence)
		log.Debugln("In steps, statement is \n", statement, " and error is ", statementErr)
		if statementErr == nil {
			err := t.executeSQL(ctx, statement)
			if err != nil {
				t.Tx.Rollback(ctx)
				log.Warningln("Statement --> ", statement, " could not be executed because of error --> ", err)
			}
		}
	}
}

// executeSQL executes the SQL query
func (t *Table) executeSQL(ctx context.Context, sql string) error {
	// var err error
	if sql != "" {
		_, err := t.Tx.Exec(ctx, sql)
		log.Debugln("Executing Statement --> \n", sql, " and error is ", err)
		return err
	}
	return nil
}

// checkColumnPresence checks if the column name passed is present in the current table
func (t *Table) checkColumnPresence(ctx context.Context, columnName string) bool {
	var presence bool
	statement := fmt.Sprintf("SELECT EXISTS(SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = '%s' AND table_catalog = '%s' AND column_name = '%s')", t.Name, t.Database, columnName)
	log.Debugln("Statement in checkColumnPresence is: \n", statement)
	err := t.Tx.QueryRow(ctx, statement).Scan(&presence)
	if err != nil {
		t.Tx.Rollback(ctx)
		log.Warningln("In checkColumnPresence, error for table --> ", t.Name, " and Column --> ", columnName, " is ", err)
	}
	log.Debugln("Presence is ", presence)
	return presence
}

// checkColumnDatatype checks if the column datatype of the column name passed is equal to the column datatype present in the table
func (t *Table) checkColumnDatatype(ctx context.Context, col Column) bool {
	columnName := col.Name
	columnDatatype := col.Datatype
	columnDefault := col.DefaultValue
	var (
		dbDatatype      string
		columnDefaultDB *string
		presence        bool
	)
	statement := fmt.Sprintf("SELECT data_type, column_default FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = '%s' AND table_catalog = '%s' AND column_name = '%s'", t.Name, t.Database, columnName)

	err := t.Tx.QueryRow(ctx, statement).Scan(&dbDatatype, &columnDefaultDB)
	if err != nil {
		t.Tx.Rollback(ctx)
		log.Warningln("While querying for column data type in table --> ", t.Name, " error is --> ", err)
	}

	log.Debugln("Datatype DB is ", dbDatatype, " and ColumnDefaultDB is ", columnDefaultDB, " and ColumnDefault is ", columnDefault, " and Column Datatype is ", columnDatatype)
	// Check if there are sequences here
	if true {
		// This is a sequence
		if columnDatatype == "bigserial" && dbDatatype == "bigint" {
			presence = true
		} else if columnDatatype == "serial" && dbDatatype == "integer" {
			presence = true
		} else if columnDefaultDB == &columnDefault {
			presence = true
		} else if columnDatatype == dbDatatype {
			presence = true
			// presence = true
		} else if dbDatatype == "ARRAY" {
			// This is the case when the datatype is an array
			joinedDatatype := strings.Join([]string{columnDefault, columnDatatype}, "::")
			log.Debugln("Datatype is ARRAY, and join is ", joinedDatatype)
			if *columnDefaultDB == joinedDatatype {
				presence = true
			}
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

	// Finally, check if there's a pseudo name for this column (where the specified data type might have a bigger name, as in PostgreSQL standard)
	// Example for this: time with/without timezone --> datatype is time, but PostgreSQL returns it as time with/without time zone. This needs to be handled
	if !presence {
		// Run this check only if the datatypes haven't already been matched
		log.Debugln("Performing check on pseudo datatype as well and pseudo datatype is ", col.PseudoDatatype)
		if dbDatatype == col.PseudoDatatype {
			presence = true
		}
	}

	log.Debugln("In check column datatype, the value of presence is ", presence)
	return presence
}

func (t *Table) commit(ctx context.Context) {
	commitErr := t.Tx.Commit(ctx)
	if commitErr != nil {
		t.Tx.Rollback(ctx)
		log.Warningln("Couldn't commit changes to the TABLE --> ", t.Name, " with error being --> ", commitErr)
	}
}
