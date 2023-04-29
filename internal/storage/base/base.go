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

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Ping() error {
	return db.DB.Ping()
}

func (db *DB) AddFilename(filename string, err error) error {
	statement, err := queries.GetPreparedStatement(queries.AddFilename)
	if err != nil {
		return err
	}

	var errMsg = ""
	if err == nil {
		errMsg = err.Error()
	}

	_, err = statement.Exec(filename, errMsg)
	return err
}

type Adder interface {
	AddFile(filename string)
}

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

func (db *DB) Prepare(vendor string) error {
	return queries.Prepare(db.DB, vendor)
}

func (db *DB) SaveDevices(devs *devices.Devices) error {
	statement, err := queries.GetPreparedStatement(queries.SaveDevices)
	if err != nil {
		return err
	}

	save := func(d devices.Device) (stop bool) {
		_, err := statement.Exec(d.Number, d.MQTT, d.InventoryID, d.UnitGUID,
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

func (db *DB) AddRelations(filename string, number []int) error {
	statement, err := queries.GetPreparedStatement(queries.AddRelation)
	if err != nil {
		return err
	}

	for _, num := range number {
		_, err = statement.Exec(filename, num)
		if err != nil {
			return err
		}
	}

	return nil
}
