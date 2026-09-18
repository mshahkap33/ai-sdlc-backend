// Package maintenanceschedule implements the Create Maintenance Schedule
// REST API described in the Maintenance and Service Scheduling TRD:
// configuring the date/mileage/manufacturer/telematics rules that trigger
// maintenance work orders.
package maintenanceschedule

// TriggerType enumerates the supported maintenance-schedule trigger types.
type TriggerType string

// Supported trigger types, as defined by the maintenance_schedules table
// check constraint.
const (
	TriggerTypeDateInterval         TriggerType = "date_interval"
	TriggerTypeMileageInterval      TriggerType = "mileage_interval"
	TriggerTypeManufacturerSchedule TriggerType = "manufacturer_schedule"
	TriggerTypeTelematicsCondition  TriggerType = "telematics_condition"
)

// Schedule is a configured maintenance-schedule rule, as persisted in the
// maintenance_schedules table.
type Schedule struct {
	ID                    string
	VehicleCategoryID     string
	VehicleID             string
	TriggerType           TriggerType
	IntervalDays          *int
	IntervalMileage       *float64
	ManufacturerReference string
	TelematicsCondition   string
	Active                bool
	CreatedBy             string
}

// CreateRequest is the payload accepted by the Create Maintenance Schedule
// API.
type CreateRequest struct {
	VehicleCategoryID     string   `json:"vehicleCategoryId"`
	VehicleID             string   `json:"vehicleId"`
	TriggerType           string   `json:"triggerType"`
	IntervalDays          *int     `json:"intervalDays"`
	IntervalMileage       *float64 `json:"intervalMileage"`
	ManufacturerReference string   `json:"manufacturerReference"`
	TelematicsCondition   string   `json:"telematicsCondition"`
}

// CreateResponse is the JSON representation returned for a newly created
// maintenance schedule rule.
type CreateResponse struct {
	ID                    string   `json:"id"`
	VehicleCategoryID     string   `json:"vehicleCategoryId,omitempty"`
	VehicleID             string   `json:"vehicleId,omitempty"`
	TriggerType           string   `json:"triggerType"`
	IntervalDays          *int     `json:"intervalDays,omitempty"`
	IntervalMileage       *float64 `json:"intervalMileage,omitempty"`
	ManufacturerReference string   `json:"manufacturerReference,omitempty"`
	TelematicsCondition   string   `json:"telematicsCondition,omitempty"`
	Active                bool     `json:"active"`
}

// toResponse converts a Schedule into its API response representation.
func toResponse(s Schedule) CreateResponse {
	return CreateResponse{
		ID:                    s.ID,
		VehicleCategoryID:     s.VehicleCategoryID,
		VehicleID:             s.VehicleID,
		TriggerType:           string(s.TriggerType),
		IntervalDays:          s.IntervalDays,
		IntervalMileage:       s.IntervalMileage,
		ManufacturerReference: s.ManufacturerReference,
		TelematicsCondition:   s.TelematicsCondition,
		Active:                s.Active,
	}
}
