package usecase

import (
	"fmt"
	"github.com/signintech/gopdf"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/watcher"
	"log"
	"reflect"
	"time"
)

type UseCase struct {
	storage     storage.Storage
	fileWatcher *watcher.Watcher
	dirOut      string
}

// New UseCase constructor
func New(storage storage.Storage, dirOut string) *UseCase {
	return &UseCase{
		storage: storage,
		dirOut:  dirOut + "/",
	}
}

// Process the files in the directory
func (u *UseCase) Process(refresh time.Duration, dir string) error {
	files := make(chan string, 100)

	u.fileWatcher = watcher.New(refresh, dir, files)
	err := u.storage.LoadFilenames(u.fileWatcher)
	if err != nil {
		return err
	}

	go func() {
		err = u.fileWatcher.Run()
		if err != nil {
			panic(err)
		}
	}()

	for {
		fmt.Println("Waiting for new file...")
		filename := <-files
		fmt.Println("New file:", filename)
		gadgets, err := events.New(dir + "/" + filename)
		if err != nil {
			return err
		}

		errFill := gadgets.Fill()
		errAdd := u.storage.AddFilename(filename, errFill)
		if errAdd != nil {
			log.Println("Failed to add filename:", errAdd)
		}

		if errFill != nil {
			log.Println("Failed to fill gadgets:", errFill)
			continue
		} else {
			gadgets.Print()
		}

		err = u.storage.SaveEvents(gadgets)
		if err != nil {
			log.Println("Failed to save devices:", err)
		}

		err = u.savePDF(gadgets)
		if err != nil {
			log.Println("Failed to save RTFs:", err)
		}
	}
}

func (u *UseCase) savePDF(devs *events.Events) error {
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
	err := pdf.AddTTFFont("LiberationSerif-Regular", "./resources/LiberationSerif-Regular.ttf")
	if err != nil {
		log.Print(err.Error())
		return err
	}

	err = pdf.SetFont("LiberationSerif-Regular", "", 14)
	if err != nil {
		log.Print(err.Error())
		return err
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
			default:
				err = pdf.Cell(nil, fmt.Sprintf("unknown type:  %v", f.Kind()))
			}
			if err != nil {
				log.Println("Failed to add text:", err)
				return err
			}
			pdf.Br(20)
		}
	}

	finalName := u.dirOut + unitGUID + ".pdf"
	err = pdf.WritePdf(finalName)
	if err != nil {
		log.Println("Failed to save PDF:", err)
		return err
	}

	return nil
}

func (u *UseCase) GetEventByNumber(unitGUID string, number int) (events.Event, error) {
	return u.storage.GetEventByNumber(unitGUID, number)
}
