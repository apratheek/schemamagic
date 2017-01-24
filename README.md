# schemamagic
Go library that automatically updates PostgreSQL schema without touching existing data.

# Motivation
This library aims to update SQL (PostgreSQL) schema changes painlessly, bridging the gap between SQL and NoSQL. I've been using this in production in a personal project of mine, and it has made my life a million times easier. No more messy table updates or setting defaults. 

# Installation
```
go get https://github.com/apratheek/schemamagic
```

# Usage
```
import "github.com/apratheek/schemamagic"
```
The above command should import the package into your code, assuming that your `$GOPATH` is set. To connect to the database, the following function needs to be called.

```
schemamagic.SetupDB(host, port, database, username, password)
```
which returns a connection to the underlying database and an error. `schemamagic` uses [pgx](https://github.com/jackc/pgx) as the PostgreSQL driver, so all the methods available on [*pgx.ConnPool](https://godoc.org/github.com/jackc/pgx#ConnPool) are valid on this connection.

## Table (struct)
```
type Table struct {
	Name          string // This denotes the name of the PostgreSQL table
	DefaultSchema string // This is the default schema (usually "public")
	Database      string // This is the name of the database
	Tx            *pgx.Tx // This is *pgx.Tx
	Autocommit    bool // Denotes if the operation on each table needs to be autocommitted (default is False)
	Columns       []Column // Stores all the columns in this table
}
```

### Create table
```
table := schemamagic.NewTable(schemamagic.Table{Name: "temp_table", DefaultSchema: "public", Database: database, Tx: tx})
```

### Available methods
1. `Append(col Column):`
This method appends a column to the table

2. `DropTable():`
This method drops the table from the database

3. `Begin():`
This method creates the table (along with all the columns) if it does not exist, or updates the schema if it has changed. 

## Column (Struct)
```
type Column struct {
	Name            string // Name of the column
	Datatype        string // Datatype of the column (bigint, bigserial, text, jsonb, bigint[], etc)
	Action          string // Default is "Add", does not support anything else as of this moment
	DefaultExists   bool // Default is false. Stores if a default value needs to be assigned to this column
        DefaultValue    string // This is the default value that will be set to the column if DefaultExists is true. Eg.: 400 (integer/bigint), 'Hello' (text), array[]::bigint[] (bigint[]), date_part('epoch'::text, now())::bigint (timestamp)
	IsUnique        bool // Default is false. If true, the unique key contraint is added
	IsPrimary       bool // Default is false. If true, the primary key constraint is added
	IsNotNull   bool // Default is false. If true, the 'NOT NULL' constraint is added
	Comment         string // This is just a comment regarding this column. Does not affect the execution of the library
	SequenceRestart int64 // In case the Datatype is either bigserial or serial, a number can be mentioned here to restart the sequence
}
```

### Create Column
```
c1 := schemamagic.NewColumn(schemamagic.Column{Name: "action", Datatype: "text", IsNotNull: true, IsUnique: true})
c2 := schemamagic.NewColumn(schemamagic.Column{Name: "created_at", Datatype: "bigint", DefaultExists: true, DefaultValue: "400"})
c3 := schemamagic.NewColumn(schemamagic.Column{Name: "version_description", Datatype: "text", DefaultExists: true, DefaultValue: "'Hello'"})
c4 := schemamagic.NewColumn(schemamagic.Column{Name: "version_new", Datatype: "bigserial"})
c5 := schemamagic.NewColumn(schemamagic.Column{Name: "arr", Datatype: "bigint[]", DefaultExists: true, DefaultValue: "array[]::bigint[]"})
c6 := schemamagic.NewColumn(schemamagic.Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint", IsPrimary: true, IsUnique: true})
```

### Add columns to a table
```
table.Append(c1)
table.Append(c2)
table.Append(c3)
table.Append(c4)
table.Append(c5)
table.Append(c6)
```

## Example
Check out a minimal [example](https://github.com/apratheek/schemamagic/blob/master/example/main.go) here.

### Things not implemented
1. Haven't yet implemented addition of foreign keys. This wasn't something I required.

### Gotchas
1. If a column has `DefaultExists` set to `true` and a corresponding `DefaultValue`, and its `DefaultValue` is changed at the next iteration, then all the columns in the table are updated with the new `DefaultValue`. This was deliberate, as in case you want to update the default value, you'd probably want to update the value in all the columns.
2. You can pass along an individual `Tx` object to update each table, or you could use the same `Tx` object to update all the tables at once. The choice is left to the developer. Of course, the changes will have to be explicitly committed by the developer (in case Autocommit is set to false). Otherwise, none of the changes would reflect (duh!).