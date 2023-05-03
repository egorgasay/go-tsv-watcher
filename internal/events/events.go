package events

import (
	"fmt"
	"github.com/dogenzaka/tsv"
	"github.com/google/uuid"
	"log"
	"os"
	"sync"
)

// Event is event struct for parsing
type Event struct {
	ID           string
	Number       int    `tsv:"n"`
	MQTT         string `tsv:"mqtt"`
	InventoryID  string `tsv:"invid"`
	UnitGUID     string `tsv:"unit_guid"`
	MessageID    string `tsv:"msg_id"`
	MessageText  string `tsv:"text"`
	Context      string `tsv:"context"`
	MessageClass string `tsv:"class"`
	Level        int    `tsv:"level"`
	Area         string `tsv:"area"`
	Address      string `tsv:"addr"`
	Block        bool   `tsv:"block"`
	Type         string `tsv:"type"`
	Bit          int    `tsv:"bit"`
	InvertBit    int    `tsv:"invert_bit"`
}

// parser is interface for parsing
type parser interface {
	Next() (bool, error)
}

// Events is events struct
// Contains []Event
type Events struct {
	current *Event
	events  []Event
	parser  parser
	file    *os.File

	mu *sync.Mutex
}

// New creates new events
func New(filename string) (*Events, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Events{
		current: new(Event),
		file:    f,
		parser:  nil,
		events:  make([]Event, 0),
		mu:      &sync.Mutex{},
	}, nil
}

// prepare prepares events
func (es *Events) prepare() error {
	var err error
	es.parser, err = es.newParser()
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	return nil
}

// newParser creates new parser
func (es *Events) newParser() (parser, error) {
	return tsv.NewParser(es.file, es.current)
}

// closeDevices closes events
func (es *Events) closeEvents() error {
	return es.file.Close()
}

// Fill fills events.
func (es *Events) Fill() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	err := es.prepare()
	if err != nil {
		return err
	}

	defer es.closeEvents()

	for {
		eof, err := es.parser.Next()
		if eof {
			return nil
		}

		if err != nil {
			return err
		}

		es.current.ID = uuid.New().String()

		es.events = append(es.events, *es.current)
	}
}

// Print prints events.
func (es *Events) Print() {
	for _, d := range es.events {
		log.Println(d.Number)
	}
}

// Iter iterates over events by giving function.
func (es *Events) Iter(cb func(d Event) (stop bool)) {
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, d := range es.events {
		if stop := cb(d); stop {
			return
		}
	}
}
