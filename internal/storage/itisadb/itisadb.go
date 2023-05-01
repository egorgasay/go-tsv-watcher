package itisadb

import (
	"context"
	"errors"
	"fmt"
	"github.com/egorgasay/itisadb-go-sdk"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/service"
	"log"
	"reflect"
	"strconv"
)

type Itisadb struct {
	client *itisadb.Client
}

var ErrEventNotFound = errors.New("event not found")

func New(creds string) (*Itisadb, error) {
	client, err := itisadb.New(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Itisadb{
		client: client,
	}, nil
}

func (i *Itisadb) LoadFilenames(adder service.Adder) error {
	files, err := i.client.Index(context.Background(), "files")
	if err != nil {
		return fmt.Errorf("failed to get files index: %w", err)
	}

	filesMap, err := files.GetIndex(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	for name := range filesMap {
		adder.AddFile(name)
	}

	return nil
}

func (i *Itisadb) AddFilename(filename string, err error) error {
	files, err := i.client.Index(context.Background(), "files")
	if err != nil {
		return err
	}

	var errMsg = ""
	if err != nil {
		errMsg = err.Error()
	}

	err = files.Set(context.Background(), filename, errMsg, false)
	if err != nil {
		return err
	}

	return nil
}

func (i *Itisadb) SaveEvents(evs *events.Events) error {
	save := func(e events.Event) (stop bool) {
		guidIndex, err := i.client.Index(context.Background(), e.UnitGUID)
		if err != nil {
			log.Printf("failed to create or get guid index: %v", err)
			return true
		}

		num, err := guidIndex.Size(context.Background()) // TODO: GET CONTEXT
		if err != nil {
			log.Printf("failed to get size: %v", err)
			return true
		}

		numIndex, err := guidIndex.Index(context.Background(), fmt.Sprintf("%d", num+1))
		if err != nil {
			log.Printf("failed to create or get index: %v", err)
			return true
		}

		ev := reflect.ValueOf(e)
		for j := 0; j < ev.NumField(); j++ {
			// get field name
			field := ev.Type().Field(j)
			// get field value
			value := ev.Field(j)
			switch field.Type.Kind() {
			case reflect.String:
				err = numIndex.Set(context.Background(), field.Name, value.String(), false)
				if err != nil {
					log.Printf("failed to save %s: %s", field.Name, err)
				}
			case reflect.Int:
				err = numIndex.Set(context.Background(), field.Name, fmt.Sprintf("%d", value.Int()), false)
				if err != nil {
					log.Printf("failed to save %s: %s", field.Name, err)
				}
			}
		}
		return false
	}

	evs.Iter(save)

	return nil
}

func (i *Itisadb) GetEventByNumber(guid string, number int) (events.Event, error) {
	guidIndex, err := i.client.Index(context.Background(), guid)
	if err != nil {
		return events.Event{}, err
	}

	numIndex, err := guidIndex.Index(context.Background(), fmt.Sprintf("%d", number))
	if err != nil {
		return events.Event{}, err
	}

	numMap, err := numIndex.GetIndex(context.Background())
	if err != nil {
		return events.Event{}, err
	}

	if len(numMap) == 0 {
		return events.Event{}, ErrEventNotFound
	}

	var event events.Event
	ev := reflect.ValueOf(&event).Elem()

	// reflect over event
	for j := 0; j < ev.NumField(); j++ {
		// get field name
		field := ev.Field(j)
		tField := ev.Type().Field(j)

		switch field.Type().Kind() {
		case reflect.String:
			field.SetString(numMap[tField.Name])
		case reflect.Int:
			num, err := strconv.Atoi(numMap[tField.Name])
			if err != nil {
				return events.Event{}, err
			}
			field.SetInt(int64(num))
		}
	}

	return event, nil
}
