package schema

type EventRequest struct {
	UnitGUID string `json:"unit_guid"`
	Page     int    `json:"page"`
}
