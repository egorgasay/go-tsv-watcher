package sqllike

import (
	"database/sql"
	"github.com/pkg/errors"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/storage/service"
	"log"
)

// DB is a abstract implementation of the storage.Database interface for sql like databases.
type DB struct {
	*sql.DB
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping checks the database connection.
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// AddFilename adds a filename and error to the database.
func (db *DB) AddFilename(filename string, errFill error) error {
	statement, err := queries.GetPreparedStatement(queries.AddFilename)
	if err != nil {
		return err
	}

	var errMsg = ""
	if errFill != nil {
		errMsg = errFill.Error()
	}

	_, err = statement.Exec(filename, errMsg)
	return err
}

// LoadFilenames loads filenames from the database into the RAM.
func (db *DB) LoadFilenames(storage service.Adder) error {
	rows, err := db.Query("SELECT name FROM files")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string

		if err := rows.Scan(&filename); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		storage.AddFile(filename)
	}
	return nil
}

// SaveEvents saves the devices to the database.
func (db *DB) SaveEvents(evs service.IEvents) error {
	statement, err := queries.GetPreparedStatement(queries.SaveEvent)
	if err != nil {
		return err
	}

	save := func(d events.Event) (stop bool) {
		_, err := statement.Exec(d.ID, d.Number, d.MQTT, d.InventoryID, d.UnitGUID,
			d.MessageID, d.MessageText, d.Context, d.MessageClass,
			d.Level, d.Area, d.Address, d.Block, d.Type, d.Bit, d.InvertBit)
		if err != nil {
			log.Println(err) // TODO: replace with logger
			return true
		}
		return false
	}

	evs.Iter(save)

	return nil
}

// GetEventByNumber returns the event by number.
func (db *DB) GetEventByNumber(guid string, number int) (events.Event, error) {
	number--
	statement, err := queries.GetPreparedStatement(queries.GetEvent)
	if err != nil {
		return events.Event{}, err
	}

	var d events.Event
	err = statement.QueryRow(guid, number).Scan(&d.ID, &d.Number, &d.MQTT, &d.InventoryID, &d.UnitGUID,
		&d.MessageID, &d.MessageText, &d.Context, &d.MessageClass, &d.Level, &d.Area, &d.Address, &d.Block, &d.Type,
		&d.Bit, &d.InvertBit)
	if err != nil {
		return events.Event{}, err
	}

	return d, nil
}
