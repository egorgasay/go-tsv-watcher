package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-tsv-watcher/internal/storage"
	"io"
	"os"
	"time"
)

// Flag struct for parsing from env and cmd args.
type Flag struct {
	// config filename
	ConfigFile *string `json:"-"`
	// directory to watch
	Directory string `json:"directory"`
	// directory to write to
	DirectoryOut string `json:"directory_out"`

	// connection string for storage
	DSN string `json:"dsn"`
	// storage type (e.g. postgres, sqlite3)
	Storage string `json:"storage_type"`

	// http(s) server config
	HTTP  string `json:"http,omitempty"`
	HTTPS string `json:"https,omitempty"`

	// refresh interval
	Refresh string `json:"refresh_interval"`
}

// Config struct for storing config values.
type Config struct {
	// mode of the server
	HTTP  string
	HTTPS string

	// directories
	Directory    string
	DirectoryOut string

	// storage config
	DBConfig *storage.Config
	// refresh interval
	Refresh time.Duration
}

var f Flag

func init() {
	f.ConfigFile = flag.String("c", "config.json", "-c=config.json")
}

// New returns a new Config struct.
func New() (*Config, error) {
	flag.Parse()

	if f.ConfigFile == nil {
		return nil, fmt.Errorf("config file is required")
	}

	err := Modify(*f.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("can't modify config: %v", err)
	}

	dur, err := time.ParseDuration(f.Refresh)
	if err != nil {
		return nil, fmt.Errorf("can't parse refresh duration: %v", err)
	}

	if f.DirectoryOut == "" {
		return nil, fmt.Errorf("directory_out is required")
	}

	err = os.Mkdir(f.DirectoryOut, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("can't create directory_out: %v", err)
	}

	return &Config{
		HTTP:  f.HTTP,
		HTTPS: f.HTTPS,

		DBConfig: &storage.Config{
			Type:           f.Storage,
			DataSourceCred: f.DSN,
		},
		DirectoryOut: f.DirectoryOut,
		Directory:    f.Directory,
		Refresh:      dur,
	}, nil
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
