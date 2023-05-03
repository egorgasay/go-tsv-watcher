package itisadb

import (
	"context"
	"fmt"
	"github.com/egorgasay/itisadb-go-sdk"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/service"
	"log"
	"reflect"
	"strconv"
)

type Itisadb struct {
	files  *itisadb.Index
	client *itisadb.Client
}

func New(ctx context.Context, creds string) (*Itisadb, error) {
	client, err := itisadb.New(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	files, err := client.Index(ctx, "files")
	if err != nil {
		return nil, fmt.Errorf("failed to get files index: %w", err)
	}

	return &Itisadb{
		client: client,
		files:  files,
	}, nil
}

func (i *Itisadb) LoadFilenames(ctx context.Context, adder service.Adder) error {
	filesMap, err := i.files.GetIndex(ctx)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	for name := range filesMap {
		adder.AddFile(name)
	}

	return nil
}

func (i *Itisadb) AddFilename(ctx context.Context, filename string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var errMsg = ""
	if err != nil {
		errMsg = err.Error()
	}

	err = i.files.Set(context.Background(), filename, errMsg, true)
	if err != nil {
		return err
	}

	return nil
}

func (i *Itisadb) SaveEvents(ctx context.Context, evs service.IEvents) error {
	save := func(e events.Event) (stop bool) {
		if ctx.Err() != nil {
			log.Println(ctx.Err())
			return true
		}

		guidIndex, err := i.client.Index(ctx, e.UnitGUID)
		if err != nil {
			log.Printf("failed to create or get guid index: %v", err)
			return true
		}

		num, err := guidIndex.Size(ctx)
		if err != nil {
			log.Printf("failed to get size: %v", err)
			return true
		}

		numIndex, err := guidIndex.Index(ctx, fmt.Sprintf("%d", num+1))
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
				err = numIndex.Set(ctx, field.Name, value.String(), false)
				if err != nil {
					log.Printf("failed to save %s: %s", field.Name, err)
				}
			case reflect.Int:
				err = numIndex.Set(ctx, field.Name, fmt.Sprintf("%d", value.Int()), false)
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

func (i *Itisadb) GetEventByNumber(ctx context.Context, guid string, number int) (events.Event, error) {
	if ctx.Err() != nil {
		return events.Event{}, ctx.Err()
	}

	guidIndex, err := i.client.Index(ctx, guid)
	if err != nil {
		return events.Event{}, err
	}

	numIndex, err := guidIndex.Index(ctx, fmt.Sprintf("%d", number))
	if err != nil {
		return events.Event{}, err
	}

	numMap, err := numIndex.GetIndex(ctx)
	if err != nil {
		return events.Event{}, err
	}

	if len(numMap) == 0 {
		return events.Event{}, service.ErrEventNotFound
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
				continue
			}
			field.SetInt(int64(num))
		}
	}

	return event, nil
}
