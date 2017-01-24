package schemamagic

import (
	"errors"
	"fmt"
	"strings"
)

// Column stores all the parameters of each column inside a table
type Column struct {
	Name            string
	Datatype        string
	Action          string
	DefaultExists   bool
	DefaultValue    string
	IsUnique        bool
	IsPrimary       bool
	IsNotNull   bool
	Comment         string
	SequenceRestart int64
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
	if c.IsNotNull {
		// Set default IsNotNull to true
		col.IsNotNull = true
	} else {
		col.IsNotNull = false
	}
	col.Comment = c.Comment
	if c.SequenceRestart == 0 {
		col.SequenceRestart = 1
	} else {
		col.SequenceRestart = c.SequenceRestart
	}
	return col
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
		if c.IsNotNull {
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
