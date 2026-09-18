package maintenanceschedule

import "testing"

func intPtr(v int) *int           { return &v }
func floatPtr(v float64) *float64 { return &v }

func validDateIntervalRequest() CreateRequest {
	return CreateRequest{
		VehicleCategoryID: "11111111-1111-1111-1111-111111111111",
		TriggerType:       string(TriggerTypeDateInterval),
		IntervalDays:      intPtr(180),
	}
}

func TestValidateValidDateIntervalRequest(t *testing.T) {
	req := validDateIntervalRequest()
	if errs := req.validate(); len(errs) != 0 {
		t.Fatalf("validate() errors = %v, want none", errs)
	}
}

func TestValidateBothVehicleAndCategoryProvided(t *testing.T) {
	req := validDateIntervalRequest()
	req.VehicleID = "22222222-2222-2222-2222-222222222222"

	errs := req.validate()
	if len(errs) == 0 {
		t.Fatal("expected validation error when both vehicleCategoryId and vehicleId are set")
	}
}

func TestValidateNeitherVehicleNorCategoryProvided(t *testing.T) {
	req := validDateIntervalRequest()
	req.VehicleCategoryID = ""

	errs := req.validate()
	if len(errs) == 0 {
		t.Fatal("expected validation error when neither vehicleCategoryId nor vehicleId are set")
	}
}

func TestValidateUnknownTriggerType(t *testing.T) {
	req := validDateIntervalRequest()
	req.TriggerType = "not_a_trigger"

	errs := req.validate()
	if len(errs) == 0 {
		t.Fatal("expected validation error for unknown triggerType")
	}
}

func TestValidateDateIntervalRequiresPositiveIntervalDays(t *testing.T) {
	tests := []struct {
		name string
		days *int
	}{
		{"nil", nil},
		{"zero", intPtr(0)},
		{"negative", intPtr(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validDateIntervalRequest()
			req.IntervalDays = tt.days

			if errs := req.validate(); len(errs) == 0 {
				t.Fatal("expected validation error for invalid intervalDays")
			}
		})
	}
}

func TestValidateMileageIntervalRequiresPositiveIntervalMileage(t *testing.T) {
	req := CreateRequest{
		VehicleCategoryID: "11111111-1111-1111-1111-111111111111",
		TriggerType:       string(TriggerTypeMileageInterval),
		IntervalMileage:   floatPtr(0),
	}

	if errs := req.validate(); len(errs) == 0 {
		t.Fatal("expected validation error for non-positive intervalMileage")
	}

	req.IntervalMileage = floatPtr(5000)
	if errs := req.validate(); len(errs) != 0 {
		t.Fatalf("validate() errors = %v, want none", errs)
	}
}

func TestValidateManufacturerScheduleRequiresReference(t *testing.T) {
	req := CreateRequest{
		VehicleCategoryID: "11111111-1111-1111-1111-111111111111",
		TriggerType:       string(TriggerTypeManufacturerSchedule),
	}

	if errs := req.validate(); len(errs) == 0 {
		t.Fatal("expected validation error for missing manufacturerReference")
	}

	req.ManufacturerReference = "10k-service"
	if errs := req.validate(); len(errs) != 0 {
		t.Fatalf("validate() errors = %v, want none", errs)
	}
}

func TestValidateTelematicsConditionRequiresCondition(t *testing.T) {
	req := CreateRequest{
		VehicleCategoryID: "11111111-1111-1111-1111-111111111111",
		TriggerType:       string(TriggerTypeTelematicsCondition),
	}

	if errs := req.validate(); len(errs) == 0 {
		t.Fatal("expected validation error for missing telematicsCondition")
	}

	req.TelematicsCondition = "check_engine"
	if errs := req.validate(); len(errs) != 0 {
		t.Fatalf("validate() errors = %v, want none", errs)
	}
}
