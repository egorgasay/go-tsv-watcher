package devices

import (
	"github.com/dogenzaka/tsv"
	"github.com/google/uuid"
	"log"
	"os"
	"sync"
)

type Device struct {
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

type Devices struct {
	current *Device
	devices []Device
	parser  parser
	file    *os.File

	mu sync.Mutex
}

func New(filename string) (*Devices, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Devices{
		current: new(Device),
		file:    f,
		parser:  nil,
		devices: make([]Device, 0),
	}, nil
}

func (ds *Devices) prepare() error {
	var err error
	ds.parser, err = ds.newParser()
	if err != nil {
		return err
	}

	return nil
}

func (ds *Devices) newParser() (parser, error) {
	return tsv.NewParser(ds.file, ds.current)
}

func (ds *Devices) closeDevices() error {
	return ds.file.Close()
}

func (ds *Devices) Fill() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	err := ds.prepare()
	if err != nil {
		return err
	}

	defer ds.closeDevices()

	for {
		eof, err := ds.parser.Next()
		if eof {
			return nil
		}

		if err != nil {
			return err
		}

		ds.current.ID = uuid.New().String()

		ds.devices = append(ds.devices, *ds.current)
	}
}

func (ds *Devices) Print() {
	for _, d := range ds.devices {
		log.Println(d.Number)
	}
}

func (ds *Devices) Iter(cb func(d Device) (stop bool)) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for _, d := range ds.devices {
		if stop := cb(d); stop {
			return
		}
	}
}
