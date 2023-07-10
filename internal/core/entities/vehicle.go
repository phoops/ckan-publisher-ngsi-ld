package entities

import "time"

type Vehicles []struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"description"`
	Speed struct {
		Type       string    `json:"type"`
		Value      int       `json:"value"`
		ObservedAt time.Time `json:"observedAt"`
	} `json:"speed"`
	VehicleType struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"vehicleType"`
	Location struct {
		Type  string `json:"type"`
		Value struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"value"`
		ObservedAt time.Time `json:"observedAt"`
	} `json:"location"`
	Heading struct {
		Type       string    `json:"type"`
		Value      int       `json:"value"`
		ObservedAt time.Time `json:"observedAt"`
	} `json:"heading"`
}
