package watcher

import (
	"fmt"
	"github.com/dolthub/swiss"
	"os"
	"time"
)

// Watcher watches a directory for new files
type Watcher struct {
	refreshInterval time.Duration
	dir             string
	processed       *swiss.Map[string, struct{}]
	files           chan string
}

// New creates a new watcher
func New(refreshInterval time.Duration, dir string, files chan string) *Watcher {
	return &Watcher{
		refreshInterval: refreshInterval,
		dir:             dir,
		processed:       swiss.NewMap[string, struct{}](100),
		files:           files,
	}
}

// AddFile adds a file to the list of processed files
func (w *Watcher) AddFile(filename string) {
	w.processed.Put(filename, struct{}{})
}

// Run starts the watcher
func (w *Watcher) Run() error {
	ticker := time.NewTicker(w.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		dir, err := os.Open(w.dir)
		if err != nil {
			return fmt.Errorf("failed to open directory: %s", err)
		}

		fis, err := dir.Readdir(-1)
		if err != nil {
			return fmt.Errorf("failed to read directory: %s", err)
		}

		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}
			if w.processed.Has(fi.Name()) {
				continue
			}

			if len(fi.Name()) < 4 || fi.Name()[len(fi.Name())-4:] != ".tsv" {
				continue
			}

			w.files <- fi.Name()
			w.processed.Put(fi.Name(), struct{}{})
		}
		dir.Close()
	}

	return nil
}
