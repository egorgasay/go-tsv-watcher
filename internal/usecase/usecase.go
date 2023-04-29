package usecase

import (
	"fmt"
	"go-tsv-watcher/internal/devices"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/watcher"
	"log"
	"time"
)

type UseCase struct {
	storage     storage.Storage
	fileWatcher *watcher.Watcher
}

// New UseCase constructor
func New(storage storage.Storage) *UseCase {
	return &UseCase{
		storage: storage,
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
		gadgets, err := devices.New(dir + "/" + filename)
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

		err = u.storage.SaveDevices(gadgets)
		if err != nil {
			log.Println("Failed to save devices:", err)
		}

		var uuids = make([]string, 0, 20)

		gadgets.Iter(func(d devices.Device) (stop bool) {
			uuids = append(uuids, d.ID)
			return false
		})

		err = u.storage.AddRelations(filename, uuids)
		if err != nil {
			log.Println("Failed to add relations:", err)
		}
	}
}

//func (u *UseCase) Save() error {
//	return u.storage.SaveFilenames(u.fileWatcher.GetProcessed())
//}
