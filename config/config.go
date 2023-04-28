package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-tsv-watcher/internal/storage"
	"io"
	"log"
	"os"
)

// Flag struct for parsing from env and cmd args.
type Flag struct {
	// config filename
	ConfigFile *string `json:"-"`
	// directory to watch
	Directory string `json:"directory"`

	// connection string for storage
	DSN string `json:"dsn"`
	// storage type (e.g. postgres, sqlite3)
	Storage string `json:"storage_type"`

	// http(s) server config
	HTTP  string `json:"http,omitempty"`
	HTTPS string `json:"https,omitempty"`
}

// Config struct for storing config values.
type Config struct {
	// mode of the server
	HTTP  string
	HTTPS string

	// directory to watch
	Directory string

	// storage config
	DBConfig *storage.Config
}

var f Flag

func init() {
	f.ConfigFile = flag.String("c", "config.json", "-c=config.json")
}

// New returns a new Config struct.
func New() *Config {
	flag.Parse()

	if f.ConfigFile == nil {
		log.Fatal("config file is required")
	}

	err := Modify(*f.ConfigFile)
	if err != nil {
		log.Fatalf("can't modify config: %v", err)
	}

	return &Config{
		HTTP:  f.HTTP,
		HTTPS: f.HTTPS,

		DBConfig: &storage.Config{
			Type:           f.Storage,
			DataSourceCred: f.DSN,
		},
		Directory: f.Directory,
	}
}

// Modify modifies the config by the file provided.
func Modify(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("can't open %s: %v", filename, err)
	}
	defer file.Close()

	all, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can't read %s: %v", filename, err)
	}

	err = json.Unmarshal(all, &f)
	if err != nil {
		return fmt.Errorf("can't unmarshal %s: %v", filename, err)
	}
	return nil
}
