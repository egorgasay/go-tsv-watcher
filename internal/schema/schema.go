package schema

// EventRequest is the schema for the event reques
type EventRequest struct {
	UnitGUID string `json:"unit_guid"`
	Page     int    `json:"page"`
}
