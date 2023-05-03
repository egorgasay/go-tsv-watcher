package service

import (
	"errors"
	"go-tsv-watcher/internal/events"
)

type Adder interface {
	AddFile(filename string)
}

type IEvents interface {
	Fill() error
	Print()
	Iter(cb func(d events.Event) (stop bool))
}

var ErrEventNotFound = errors.New("event not found")
