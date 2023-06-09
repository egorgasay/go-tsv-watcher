# go-tsv-watcher [![Tests](https://github.com/egorgasay/go-tsv-watcher/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/egorgasay/go-tsv-watcher/actions/workflows/ci.yml)

![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![SQLite](https://img.shields.io/badge/sqlite-%2307405e.svg?style=for-the-badge&logo=sqlite&logoColor=white)
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)

## TSV Parser and Watcher

### Point

The service scans a given directory at a given interval and parses new TSV files, and also saves them to pdf files by device name in a separate directory and database selected by the user.

### Database

The service supports the use of three databases. Supported: PostgreSQL, Sqlite, itisadb. I recommend using the itisadb database, as it is more suitable for high loads (developed by me).

### CLI arguments

```bash
-c=path/to/config.json
```

### Config
```go
// directory to watch
Directory string `json:"directory"`
// directory to write to
DirectoryOut string `json:"directory_out"`

// connection string for storage
DSN string `json:"dsn"`
// storage type (e.g. postgres, sqlite3, itisasb)
Storage string `json:"storage_type"`

// http(s) server mode
HTTP  string `json:"http"`
HTTPS string `json:"https"`

// refresh interval
Refresh string `json:"refresh_interval"`
```

### Config example
```json
{
  "http": ":80",
  "directory": "test_dir",
  "directory_out": "test_out",
  "storage_type": "itisadb",
  "dsn": "127.0.0.1:800",
  "refresh_interval": "1s"
}
```

### Request

You can receive an event via an HTTP(-S) request to your service.


### Request example
```http
POST http://IP:PORT/api/v1/event HTTP/1.1
Content-Type: application/json
{
    "unit_guid": "01749246-95f6-57db-b7c3-2ae0e8be6715",
    "page": 1
}
```

### Response example
```json
{
    "ID": "14d013b1-3de3-4dda-8ee6-42474a53e56f",
    "Number": 1,
    "MQTT": "",
    "InventoryID": "G-044322",
    "UnitGUID": "01749246-95f6-57db-b7c3-2ae0e8be6715",
    "MessageID": "cold7_Defrost_status",
    "MessageText": "Разморозка",
    "Context": "",
    "MessageClass": "waiting",
    "Level": 100,
    "Area": "LOCAL",
    "Address": "cold7_status.Defrost_status",
    "Block": false,
    "Type": "",
    "Bit": 0,
    "InvertBit": 0
}
```

### Quick Run
The default 'config.json' file will be used. Make sure you have it.
```bash
docker-compose up
```
