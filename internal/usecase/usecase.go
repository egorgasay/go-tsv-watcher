package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/signintech/gopdf"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/storage/service"
	"go-tsv-watcher/internal/watcher"
	"go-tsv-watcher/pkg/logger"
	"log"
	"reflect"
	"time"
)

// UseCase struct for the logic layer.
type UseCase struct {
	storage     storage.Storage
	fileWatcher *watcher.Watcher
	dirOut      string
	logger      logger.ILogger
}

// ErrStorageIsUnavailable error occurs when is unavailable
var ErrStorageIsUnavailable = errors.New("storage is unavailable")

// IUseCase interface for mock testing.
//
//go:generate mockgen -source=usecase.go -destination=mocks/mock.go
type IUseCase interface {
	Process(ctx context.Context, refresh time.Duration, dir string) error
	GetEventByNumber(ctx context.Context, unitGUID string, number int) (events.Event, error)
}

// New UseCase constructor
func New(storage storage.Storage, dirOut string, loggerInstance logger.ILogger) *UseCase {
	return &UseCase{
		storage: storage,
		dirOut:  dirOut + "/",
		logger:  loggerInstance,
	}
}

// Process the files in the directory
func (u *UseCase) Process(ctx context.Context, refresh time.Duration, dir string) error {
	files := make(chan string, 100)

	u.fileWatcher = watcher.New(refresh, dir, files)
	err := u.storage.LoadFilenames(ctx, u.fileWatcher)
	if err != nil {
		return fmt.Errorf("failed to load filenames: %w", err)
	}

	go func() {
		err = u.fileWatcher.Run()
		if err != nil {
			u.logger.Warn(err.Error())
		}
	}()

	for ctx.Err() == nil {
		fmt.Println("Waiting for new file...")
		filename := <-files
		fmt.Println("New file:", filename)
		gadgets, err := events.New(dir + "/" + filename)
		if err != nil {
			return fmt.Errorf("failed to create events: %w", err)
		}

		errFill := gadgets.Fill()
		errAdd := u.storage.AddFilename(ctx, filename, errFill)
		if errAdd != nil {
			u.logger.Warn(fmt.Sprintf("Failed to add filename: %v", errAdd))
		}

		if errFill != nil {
			u.logger.Warn(fmt.Sprintf("Failed to fill gadgets: %v", errFill))
			continue
		} else {
			gadgets.Print()
		}

		err = u.storage.SaveEvents(ctx, gadgets)
		if err != nil {
			log.Println("Failed to save devices:", err)
		}

		err = u.savePDF(gadgets)
		if err != nil {
			u.logger.Warn(err.Error())
		}
	}
	close(files)
	return ctx.Err()
}

func (u *UseCase) savePDF(devs service.IEvents) error {
	var devicesGroups = make(map[string][]events.Event, 20)
	devs.Iter(func(d events.Event) (stop bool) {
		devicesGroups[d.UnitGUID] = append(devicesGroups[d.UnitGUID], d)
		return false
	})

	for unitGUID, group := range devicesGroups {
		err := u.process(group, unitGUID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UseCase) process(group []events.Event, unitGUID string) error {
	pdf := gopdf.GoPdf{}
	defer pdf.Close()

	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	err := pdf.AddTTFFont("LiberationSerif-Regular", "resources/LiberationSerif-Regular.ttf")
	if err != nil {
		u.logger.Warn(fmt.Sprintf("failed to add font: %v", err.Error()))
		return fmt.Errorf("failed to add font: %w", err)
	}

	err = pdf.SetFont("LiberationSerif-Regular", "", 14)
	if err != nil {
		u.logger.Warn(fmt.Sprintf("failed to set font: %v", err.Error()))
		return fmt.Errorf("failed to set font: %w", err)
	}

	for _, d := range group {
		pdf.AddPage()

		// for reflection over devices.Device
		dv := reflect.ValueOf(d)
		for i := 0; i < dv.NumField(); i++ {
			f := dv.Field(i)
			switch f.Kind() {
			case reflect.String:
				err = pdf.Cell(nil, fmt.Sprintf("%s:  %v", dv.Type().Field(i).Name, f.String()))
			case reflect.Int:
				err = pdf.Cell(nil, fmt.Sprintf("%s:  %v", dv.Type().Field(i).Name, f.Int()))
			case reflect.Bool:
				err = pdf.Cell(nil, fmt.Sprintf("%s:  %v", dv.Type().Field(i).Name, f.Bool()))
			default:
				err = pdf.Cell(nil, fmt.Sprintf("unknown type:  %v", f.Kind()))
			}
			if err != nil {
				u.logger.Warn(fmt.Sprintf("Failed to add text: %v", err))
				return fmt.Errorf("failed to add text: %w", err)
			}
			pdf.Br(20)
		}
	}

	finalName := u.dirOut + unitGUID + ".pdf"
	err = pdf.WritePdf(finalName)
	if err != nil {
		u.logger.Warn(fmt.Sprintf("Failed to save PDF: %v", err))
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}

// GetEventByNumber gets an event by number
func (u *UseCase) GetEventByNumber(ctx context.Context, unitGUID string, number int) (events.Event, error) {
	if number <= 0 {
		return events.Event{}, service.ErrEventNotFound
	}

	ev, err := u.storage.GetEventByNumber(ctx, unitGUID, number)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return ev, err
		}
		u.logger.Warn(err.Error())
		return ev, ErrStorageIsUnavailable
	}
	return ev, nil
}
