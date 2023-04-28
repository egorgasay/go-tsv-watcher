package base

import "database/sql"

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

func (db *DB) Begin() (StorageTx, error) {

}
