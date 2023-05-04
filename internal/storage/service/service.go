package service

import (
	"errors"
	"go-tsv-watcher/internal/events"
)

// Adder common interface for adding files
type Adder interface {
	AddFile(filename string)
}

// IEvents common interface for events
type IEvents interface {
	Fill() error
	Print()
	Iter(cb func(d events.Event) (stop bool))
}

// ErrEventNotFound error for not found event
var ErrEventNotFound = errors.New("event not found")
