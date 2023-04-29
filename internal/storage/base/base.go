package base

import (
	"database/sql"
	"github.com/pkg/errors"
	"go-tsv-watcher/internal/devices"
	"go-tsv-watcher/internal/storage/queries"
	"log"
)

// DB is a basic implementation of the storage.Repository interface.
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

type Adder interface {
	AddFile(filename string)
}

// LoadFilenames loads filenames from the database into the RAM.
func (db *DB) LoadFilenames(storage Adder) error {
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

type Iterator[K any, V any] interface {
	Iter(func(k K, v V) (stop bool))
}

// Prepare prepares the database for usage.
func (db *DB) Prepare(vendor string) error {
	return queries.Prepare(db.DB, vendor)
}

// SaveDevices saves the devices to the database.
func (db *DB) SaveDevices(devs *devices.Devices) error {
	statement, err := queries.GetPreparedStatement(queries.SaveEvent)
	if err != nil {
		return err
	}

	save := func(d devices.Device) (stop bool) {
		_, err := statement.Exec(d.ID, d.Number, d.MQTT, d.InventoryID, d.UnitGUID,
			d.MessageID, d.MessageText, d.Context, d.MessageClass,
			d.Level, d.Area, d.Address, d.Block, d.Type, d.Bit, d.InvertBit)
		if err != nil {
			log.Println(err) // TODO: replace with logger
			return true
		}
		return false
	}

	devs.Iter(save)

	return nil
}

// AddRelations adds relations to the database.
func (db *DB) AddRelations(filename string, uuids []string) error {
	statement, err := queries.GetPreparedStatement(queries.AddRelation)
	if err != nil {
		return err
	}

	for _, uniqueID := range uuids {
		_, err = statement.Exec(filename, uniqueID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetEventByNumber(guid string, number int) (devices.Device, error) {
	statement, err := queries.GetPreparedStatement(queries.GetEvent)
	if err != nil {
		return devices.Device{}, err
	}

	var d devices.Device
	err = statement.QueryRow(guid, number).Scan(&d.ID, &d.Number, &d.MQTT, &d.InventoryID, &d.UnitGUID,
		&d.MessageID, &d.MessageText, &d.Context, &d.MessageClass, &d.Level, &d.Area, &d.Address, &d.Block, &d.Type,
		&d.Bit, &d.InvertBit)
	if err != nil {
		return devices.Device{}, err
	}

	return d, nil
}
