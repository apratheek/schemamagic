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
func (c Constraint) createDropRule(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", tableName, c.Name)
}

// createAddRule generates the SQL statement that will be used to add this constraint
func (c Constraint) createAddRule(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", tableName, c.Name, c.Value)
}
