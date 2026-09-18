package vehicle

import (
	"testing"
	"time"
)

func validRequest() CreateRequest {
	return CreateRequest{
		VIN:         "1HGCM82633A004352",
		PlateNumber: "ABC-1234",
		Make:        "Honda",
		Model:       "Accord",
		Year:        2023,
		CategoryID:  "11111111-1111-1111-1111-111111111111",
		Mileage:     1200,
		FuelLevel:   80,
		Location:    "Depot A",
	}
}

func TestCreateRequestValidateValid(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	if errs := validRequest().validate(now); len(errs) != 0 {
		t.Fatalf("validate() = %v, want no errors", errs)
	}
}

func TestCreateRequestValidateInvalidVIN(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := validRequest()
	req.VIN = "too-short"

	errs := req.validate(now)
	if !hasFieldError(errs, "vin") {
		t.Fatalf("validate() = %v, want vin error", errs)
	}
}

func TestCreateRequestValidateVINExcludesIOQ(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := validRequest()
	req.VIN = "1HGCM82633A00435I" // contains I, which is disallowed

	errs := req.validate(now)
	if !hasFieldError(errs, "vin") {
		t.Fatalf("validate() = %v, want vin error for VIN containing I", errs)
	}
}

func TestCreateRequestValidateMissingRequiredFields(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := CreateRequest{}

	errs := req.validate(now)
	for _, field := range []string{"vin", "plate_number", "make", "model", "category_id", "location"} {
		if !hasFieldError(errs, field) {
			t.Errorf("validate() missing expected error for field %q, got %v", field, errs)
		}
	}
}

func TestCreateRequestValidateYearTooFarInFuture(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := validRequest()
	req.Year = 2026

	errs := req.validate(now)
	if !hasFieldError(errs, "year") {
		t.Fatalf("validate() = %v, want year error", errs)
	}
}

func TestCreateRequestValidateNegativeMileage(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := validRequest()
	req.Mileage = -1

	errs := req.validate(now)
	if !hasFieldError(errs, "mileage") {
		t.Fatalf("validate() = %v, want mileage error", errs)
	}
}

func TestCreateRequestValidateFuelLevelOutOfRange(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := validRequest()
	req.FuelLevel = 150

	errs := req.validate(now)
	if !hasFieldError(errs, "fuel_level") {
		t.Fatalf("validate() = %v, want fuel_level error", errs)
	}
}

func hasFieldError(errs ValidationErrors, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}
