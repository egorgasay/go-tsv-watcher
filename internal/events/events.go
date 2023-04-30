package events

import (
	"github.com/dogenzaka/tsv"
	"github.com/google/uuid"
	"log"
	"os"
	"sync"
)

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

type parser interface {
	Next() (bool, error)
}

type Events struct {
	current *Event
	devices []Event
	parser  parser
	file    *os.File

	mu sync.Mutex
}

func New(filename string) (*Events, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Events{
		current: new(Event),
		file:    f,
		parser:  nil,
		devices: make([]Event, 0),
	}, nil
}

func (es *Events) prepare() error {
	var err error
	es.parser, err = es.newParser()
	if err != nil {
		return err
	}

	return nil
}

func (es *Events) newParser() (parser, error) {
	return tsv.NewParser(es.file, es.current)
}

func (es *Events) closeDevices() error {
	return es.file.Close()
}

func (es *Events) Fill() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	err := es.prepare()
	if err != nil {
		return err
	}

	defer es.closeDevices()

	for {
		eof, err := es.parser.Next()
		if eof {
			return nil
		}

		if err != nil {
			return err
		}

		es.current.ID = uuid.New().String()

		es.devices = append(es.devices, *es.current)
	}
}

func (es *Events) Print() {
	for _, d := range es.devices {
		log.Println(d.Number)
	}
}

func (es *Events) Iter(cb func(d Event) (stop bool)) {
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, d := range es.devices {
		if stop := cb(d); stop {
			return
		}
	}
}
