package entities

import "time"

// Vehicle data that will be saved to CKAN datastore
type GateCount struct {
	Parking          string    `json:"parking"`
	Gate             string    `json:"gate"`
	Coordinates      []float64 `json:"coordinates"`
	BeginObservation time.Time `json:"beginObservation"`
	EndObservation   time.Time `json:"endObservation"`
	Count            int       `json:"count"`
}
