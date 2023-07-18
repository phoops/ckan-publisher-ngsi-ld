package entities

import "time"

// Vehicle data that will be saved to CKAN datastore
type GateCount struct {
	Parking          string    `json:"parking"`
	Gate             string    `json:"gate"`
	Coordinate1      float64   `json:"coordinate1"`
	Coordinate2      float64   `json:"coordinate2"`
	BeginObservation time.Time `json:"beginObservation"`
	EndObservation   time.Time `json:"endObservation"`
	Count            int       `json:"count"`
}
