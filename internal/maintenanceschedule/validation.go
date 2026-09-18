package maintenanceschedule

import (
	"fmt"
	"strings"
)

// ValidationError describes a single invalid field in a request.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of field-level validation failures.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// validTriggerTypes are the enumerated trigger types accepted by the
// maintenance_schedules table.
var validTriggerTypes = map[TriggerType]bool{
	TriggerTypeDateInterval:         true,
	TriggerTypeMileageInterval:      true,
	TriggerTypeManufacturerSchedule: true,
	TriggerTypeTelematicsCondition:  true,
}

// validate checks req against the Common Validation Rules defined in the
// Maintenance and Service Scheduling TRD and returns any violations found.
func (req CreateRequest) validate() ValidationErrors {
	var errs ValidationErrors

	hasCategory := req.VehicleCategoryID != ""
	hasVehicle := req.VehicleID != ""
	if hasCategory == hasVehicle {
		errs = append(errs, ValidationError{"vehicleCategoryId", "exactly one of vehicleCategoryId or vehicleId must be provided"})
	}

	triggerType := TriggerType(req.TriggerType)
	if req.TriggerType == "" {
		errs = append(errs, ValidationError{"triggerType", "is required"})
	} else if !validTriggerTypes[triggerType] {
		errs = append(errs, ValidationError{"triggerType", "must be one of date_interval, mileage_interval, manufacturer_schedule, telematics_condition"})
	} else {
		switch triggerType {
		case TriggerTypeDateInterval:
			if req.IntervalDays == nil || *req.IntervalDays <= 0 {
				errs = append(errs, ValidationError{"intervalDays", "is required and must be greater than 0 for trigger_type date_interval"})
			}
		case TriggerTypeMileageInterval:
			if req.IntervalMileage == nil || *req.IntervalMileage <= 0 {
				errs = append(errs, ValidationError{"intervalMileage", "is required and must be greater than 0 for trigger_type mileage_interval"})
			}
		case TriggerTypeManufacturerSchedule:
			if req.ManufacturerReference == "" {
				errs = append(errs, ValidationError{"manufacturerReference", "is required for trigger_type manufacturer_schedule"})
			}
		case TriggerTypeTelematicsCondition:
			if req.TelematicsCondition == "" {
				errs = append(errs, ValidationError{"telematicsCondition", "is required for trigger_type telematics_condition"})
			}
		}
	}

	return errs
}
