package usecase

import (
	"go-tsv-watcher/internal/devices"
	"go-tsv-watcher/internal/storage"
)

type UseCase struct {
	storage *storage.Storage
	gadgets *devices.Devices
}

func New(storage *storage.Storage) *UseCase {
	return &UseCase{
		storage: storage,
		gadgets: nil,
	}
}

func (u *UseCase) Process(filename string) error {
	gadgets, err := devices.New(filename)
	if err != nil {
		return err
	}
	u.gadgets = gadgets

	err = u.gadgets.Fill()
	if err != nil {
		return err
	}
	return nil
}
