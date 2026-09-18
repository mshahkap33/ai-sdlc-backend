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
		CategoryID:  "b3f1f2a0-1111-4a2b-8c3d-000000000001",
		Mileage:     1200,
		FuelLevel:   80,
		Location:    "Depot A",
	}
}

func TestValidateCreateRequestValid(t *testing.T) {
	req := validRequest()
	now := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)

	if verr := ValidateCreateRequest(req, now); verr != nil {
		t.Fatalf("ValidateCreateRequest() = %v, want nil", verr)
	}
}

func TestValidateCreateRequestFieldErrors(t *testing.T) {
	now := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		mutate    func(*CreateRequest)
		wantField string
	}{
		{"missing vin", func(r *CreateRequest) { r.VIN = "" }, "vin"},
		{"vin too short", func(r *CreateRequest) { r.VIN = "SHORT" }, "vin"},
		{"vin contains excluded letter", func(r *CreateRequest) { r.VIN = "1HGCM82633A00435I" }, "vin"},
		{"missing plate number", func(r *CreateRequest) { r.PlateNumber = "" }, "plate_number"},
		{"invalid plate number", func(r *CreateRequest) { r.PlateNumber = "!!!" }, "plate_number"},
		{"missing make", func(r *CreateRequest) { r.Make = "" }, "make"},
		{"missing model", func(r *CreateRequest) { r.Model = "" }, "model"},
		{"year too far in future", func(r *CreateRequest) { r.Year = 2026 }, "year"},
		{"year not 4 digits", func(r *CreateRequest) { r.Year = 99 }, "year"},
		{"missing category", func(r *CreateRequest) { r.CategoryID = "" }, "category_id"},
		{"negative mileage", func(r *CreateRequest) { r.Mileage = -1 }, "mileage"},
		{"fuel level too low", func(r *CreateRequest) { r.FuelLevel = -1 }, "fuel_level"},
		{"fuel level too high", func(r *CreateRequest) { r.FuelLevel = 101 }, "fuel_level"},
		{"missing location", func(r *CreateRequest) { r.Location = "" }, "location"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validRequest()
			tt.mutate(&req)

			verr := ValidateCreateRequest(req, now)
			if verr == nil {
				t.Fatalf("ValidateCreateRequest() = nil, want error for field %q", tt.wantField)
			}

			found := false
			for _, fe := range verr.Errors {
				if fe.Field == tt.wantField {
					found = true
				}
			}
			if !found {
				t.Fatalf("ValidateCreateRequest() errors = %+v, want an error for field %q", verr.Errors, tt.wantField)
			}
		})
	}
}

func TestValidateCreateRequestYearAllowsNextYear(t *testing.T) {
	now := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)
	req := validRequest()
	req.Year = 2025

	if verr := ValidateCreateRequest(req, now); verr != nil {
		t.Fatalf("ValidateCreateRequest() = %v, want nil for next year", verr)
	}
}
