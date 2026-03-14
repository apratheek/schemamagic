package schemamagic

import (
	"fmt"
)

// Constraint stores the applicable constraint on a table
type Constraint struct {
	Name  string // This stores the name of the constrant
	Value string // This stores the constrant that needs to be applied
}

// createDropRule generates the SQL statement that will be used to drop this constraint
func (c Constraint) createDropRule(tableName string, schema string) string {
	return fmt.Sprintf("ALTER TABLE %s.%s DROP CONSTRAINT IF EXISTS %s", schema, tableName, c.Name)
}

// createAddRule generates the SQL statement that will be used to add this constraint
func (c Constraint) createAddRule(tableName string, schema string) string {
	return fmt.Sprintf("ALTER TABLE %s.%s ADD CONSTRAINT %s %s", schema, tableName, c.Name, c.Value)
}
