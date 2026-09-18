// Package vehicle implements the Vehicle Onboarding REST API: creating and
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
	ConditionNotes string
	CreatedAt      time.Time
	CreatedBy      string
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
	ConditionNotes string  `json:"condition_notes"`
}

// CreateResponse is the JSON representation returned for a newly created
// vehicle.
type CreateResponse struct {
	ID             string    `json:"id"`
	VIN            string    `json:"vin"`
	PlateNumber    string    `json:"plate_number"`
	Make           string    `json:"make"`
	Model          string    `json:"model"`
	Year           int       `json:"year"`
	CategoryID     string    `json:"category_id"`
	Mileage        float64   `json:"mileage"`
	FuelLevel      float64   `json:"fuel_level"`
	Location       string    `json:"location"`
	ConditionNotes string    `json:"condition_notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// toResponse converts a Vehicle into its API response representation.
func toResponse(v Vehicle) CreateResponse {
	return CreateResponse{
		ID:             v.ID,
		VIN:            v.VIN,
		PlateNumber:    v.PlateNumber,
		Make:           v.Make,
		Model:          v.Model,
		Year:           v.Year,
		CategoryID:     v.CategoryID,
		Mileage:        v.Mileage,
		FuelLevel:      v.FuelLevel,
		Location:       v.Location,
		ConditionNotes: v.ConditionNotes,
		CreatedAt:      v.CreatedAt,
	}
}
