package events

import (
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestEvents_Fill(t *testing.T) {
	type fields struct {
		current *Event
		events  []Event
		parser  parser
		file    *os.File
		mu      *sync.Mutex
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		filename string
		tsvData  string
	}{
		{
			name: "test1",
			fields: fields{
				current: nil,
				events: []Event{
					{
						Number:       1,
						InventoryID:  "G-044322",
						MQTT:         "",
						UnitGUID:     "01749246-95f6-57db-b7c3-2ae0e8be6715",
						MessageID:    "cold7_Defrost_status",
						MessageText:  "Разморозка",
						Context:      "",
						MessageClass: "waiting",
						Level:        100,
						Area:         "LOCAL",
						Address:      "cold7_status.Defrost_status",
						Block:        false,
						Type:         "",
						Bit:          0,
						InvertBit:    0,
					},
				},
				mu: &sync.Mutex{},
			},
			filename: "/tmp/test1",
			tsvData:  "n\tmqtt\tinvid\tunit_guid\tmsg_id\ttext\tcontext\tclass\tlevel\tarea\taddr\tblock\ttype\tbit\tinvert_bit\n1\t\tG-044322\t01749246-95f6-57db-b7c3-2ae0e8be6715\tcold7_Defrost_status\tРазморозка\t\twaiting\t100\tLOCAL\tcold7_status.Defrost_status\t\t\t\t",
		},
		{
			name: "test2",
			fields: fields{
				current: nil,
				events: []Event{
					{
						Number:       1,
						InventoryID:  "G-044322",
						MQTT:         "",
						UnitGUID:     "01749246-95f6-57db-b7c3-2ae0e8be6715",
						MessageID:    "cold7_Defrost_status",
						MessageText:  "Разморозка",
						Context:      "",
						MessageClass: "waiting",
						Level:        100,
						Area:         "LOCAL",
						Address:      "cold7_status.Defrost_status",
						Block:        false,
						Type:         "",
						Bit:          0,
						InvertBit:    0,
					},
					{
						Number:       2,
						InventoryID:  "G-044322",
						MQTT:         "",
						UnitGUID:     "01749246-95f6-57db-b7c3-2ae0e8be6715",
						MessageID:    "cold7_VentSK_status",
						MessageText:  "Вентилятор",
						Context:      "",
						MessageClass: "working",
						Level:        100,
						Area:         "LOCAL",
						Address:      "cold7_status.VentSK_status",
						Block:        false,
						Type:         "",
						Bit:          0,
						InvertBit:    0,
					},
				},
				mu: &sync.Mutex{},
			},
			filename: "/tmp/test1",
			tsvData:  "n\tmqtt\tinvid\tunit_guid\tmsg_id\ttext\tcontext\tclass\tlevel\tarea\taddr\tblock\ttype\tbit\tinvert_bit\n1\t\tG-044322\t01749246-95f6-57db-b7c3-2ae0e8be6715\tcold7_Defrost_status\tРазморозка\t\twaiting\t100\tLOCAL\tcold7_status.Defrost_status\t\t\t\t\n2\t\tG-044322\t01749246-95f6-57db-b7c3-2ae0e8be6715\tcold7_VentSK_status\tВентилятор\t\tworking\t100\tLOCAL\tcold7_status.VentSK_status\t\t\t\t",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.OpenFile(tt.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				t.Fatalf("os.OpenFile() error = %v", err)
			}

			_, err = f.WriteString(tt.tsvData)
			if err != nil {
				t.Fatalf("f.WriteString() error = %v", err)
			}

			f.Sync()
			f.Close()

			es, err := New(tt.filename)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if err := es.Fill(); (err != nil) != tt.wantErr {
				t.Errorf("Fill() error = %v,\n wantErr %v", err, tt.wantErr)
			}

			for i := range es.events {
				es.events[i].ID = ""
				if !reflect.DeepEqual(es.events[i], tt.fields.events[i]) {
					t.Errorf("Fill() es.events = %v,\n want %v", es.events[i], tt.fields.events[i])
				}
			}

			err = os.Remove(tt.filename)
			if err != nil {
				t.Fatalf("os.Remove() error = %v", err)
			}
		})
	}
}

func TestEvents_Iter(t *testing.T) {
	ev := &Events{events: []Event{{Number: 0}, {Number: 1}, {Number: 2}}, mu: &sync.Mutex{}}
	evts := make([]Event, 0)

	ev.Iter(func(e Event) (stop bool) {
		evts = append(evts, e)
		return false
	})

	for i := range ev.events {
		if !reflect.DeepEqual(ev.events[i], evts[i]) {
			t.Errorf("Iter() ev.events = %v,\n want %v", ev.events[i], evts[i])
		}
	}
}
