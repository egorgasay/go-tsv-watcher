package storage

import (
	"github.com/dolthub/swiss"
	"sync"
)

type database interface {
	AddFilename(filename string) error
	loadFilenames() (string, error)

	Save() error
}

type Storage struct {
	db         database
	mu         sync.RWMutex
	ramStorage *swiss.Map[string, struct{}]
}

type Config struct {
	Type           string
	DataSourceCred string
}

func New(Config) *Storage {
	return &Storage{
		//TODO: implement
	}
}
