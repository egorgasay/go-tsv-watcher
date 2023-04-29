package queries

import (
	"database/sql"
	"errors"
	"fmt"
)

// Query text of query.
type Query string

// Name number of query.
type Name int

// Query names.
const (
	AddFilename = iota
	SaveEvent
	AddRelation
	GetEvent
)

var queriesSqlite3 = map[Name]Query{
	AddFilename: "INSERT INTO files (name, error) VALUES (?, ?)",
	SaveEvent: `INSERT INTO events (ID,
                     Number, MQTT ,InventoryID, 
                     UnitGUID, MessageID, MessageText,
                     Context  ,MessageClass, Level, 
                     Area, Address , Block, Type, Bit, 
                     InvertBit) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	AddRelation: "INSERT INTO relations (file_name, event_id) VALUES (?, ?)",
	GetEvent:    "SELECT * FROM events WHERE UnitGUID = ? LIMIT 1 OFFSET ?",
}

var queriesPostgres = map[Name]Query{
	AddFilename: "INSERT INTO files (name, error) VALUES ($1, $2)",
	SaveEvent: `INSERT INTO events (ID,
                     Number, MQTT ,InventoryID, 
                     UnitGUID, MessageID, MessageText,
                     Context  ,MessageClass, Level, 
                     Area, Address , Block, Type, Bit, 
                     InvertBit) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
	AddRelation: "INSERT INTO relations (file_name, event_id) VALUES ($1, $2::uuid)",
	GetEvent:    "SELECT * FROM events WHERE UnitGUID = $1 LIMIT 1 OFFSET $2",
}

// ErrNotFound occurs when query was not found.
var ErrNotFound = errors.New("the query was not found")

// ErrNilStatement occurs query statement is nil.
var ErrNilStatement = errors.New("query statement is nil")

var statements = make(map[Name]*sql.Stmt, 10)

// Prepare prepares all queries for db instance.
func Prepare(DB *sql.DB, vendor string) error {
	var queries map[Name]Query
	switch vendor {
	case "sqlite3":
		queries = queriesSqlite3
	case "postgres":
		queries = queriesPostgres
	}

	for n, q := range queries {
		prep, err := DB.Prepare(string(q))
		if err != nil {
			return err
		}
		statements[n] = prep
	}
	return nil
}

// GetPreparedStatement returns *sql.Stmt by name of query.
func GetPreparedStatement(name int) (*sql.Stmt, error) {
	stmt, ok := statements[Name(name)]
	if !ok {
		return nil, ErrNotFound
	}

	if stmt == nil {
		return nil, ErrNilStatement
	}

	return stmt, nil
}

// Close closes all prepared statements.
func Close() error {
	for _, stmt := range statements {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("error closing statement: %w", err)
		}
	}

	return nil
}
