// Package vehicle implements the Vehicle Onboarding capability described in
// the "Vehicle Onboarding" technical requirement document: creating and
// storing canonical vehicle master data records.
package vehicle

import "time"

// Vehicle is the canonical vehicle master data record.
type Vehicle struct {
	ID             string
	VIN            string
	PlateNumber    string
	Make           string
	Model          string
	Year           int
	CategoryID     string
	Mileage        float64
	FuelLevel      float64
	Location       string
	ConditionNotes *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedBy      string
	UpdatedBy      string
}

// CreateRequest is the payload accepted by the Create Vehicle API.
type CreateRequest struct {
	VIN            string  `json:"vin"`
	PlateNumber    string  `json:"plate_number"`
	Make           string  `json:"make"`
	Model          string  `json:"model"`
	Year           int     `json:"year"`
	CategoryID     string  `json:"category_id"`
	Mileage        float64 `json:"mileage"`
	FuelLevel      float64 `json:"fuel_level"`
	Location       string  `json:"location"`
	ConditionNotes *string `json:"condition_notes,omitempty"`
}
