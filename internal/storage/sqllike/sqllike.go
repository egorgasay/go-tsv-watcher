package sqllike

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/storage/service"
	"log"
)

// DB is an abstract implementation of the storage.Database interface for sql like databases.
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
func (db *DB) AddFilename(ctx context.Context, filename string, errFill error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	statement, err := queries.GetPreparedStatement(queries.AddFilename)
	if err != nil {
		return err
	}

	var errMsg = ""
	if errFill != nil {
		errMsg = errFill.Error()
	}

	_, err = statement.ExecContext(ctx, filename, errMsg)
	return err
}

// LoadFilenames loads filenames from the database into the RAM.
func (db *DB) LoadFilenames(ctx context.Context, storage service.Adder) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	rows, err := db.QueryContext(ctx, "SELECT name FROM files")
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
func (db *DB) SaveEvents(ctx context.Context, evs service.IEvents) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	statement, err := queries.GetPreparedStatement(queries.SaveEvent)
	if err != nil {
		return err
	}

	save := func(d events.Event) (stop bool) {
		if ctx.Err() != nil {
			log.Println(ctx.Err()) // TODO: replace with logger
			return true
		}
		_, err = statement.Exec(d.ID, d.Number, d.MQTT, d.InventoryID, d.UnitGUID,
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
func (db *DB) GetEventByNumber(ctx context.Context, guid string, number int) (events.Event, error) {
	if ctx.Err() != nil {
		return events.Event{}, ctx.Err()
	}

	number--
	statement, err := queries.GetPreparedStatement(queries.GetEvent)
	if err != nil {
		return events.Event{}, err
	}

	var d events.Event
	err = statement.QueryRowContext(ctx, guid, number).Scan(&d.ID, &d.Number, &d.MQTT, &d.InventoryID, &d.UnitGUID,
		&d.MessageID, &d.MessageText, &d.Context, &d.MessageClass, &d.Level, &d.Area, &d.Address, &d.Block, &d.Type,
		&d.Bit, &d.InvertBit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return events.Event{}, service.ErrEventNotFound
		}
		return events.Event{}, err
	}

	return d, nil
}
