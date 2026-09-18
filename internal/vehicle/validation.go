package vehicle

import (
	"regexp"
	"strings"
	"time"
)

// vinPattern matches the standard 17-character VIN format, excluding the
// letters I, O, and Q as specified by the Vehicle Onboarding TRD.
var vinPattern = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)

// plateNumberPattern allows alphanumeric characters with optional hyphen or
// space separators.
var plateNumberPattern = regexp.MustCompile(`^[A-Za-z0-9]+([- ][A-Za-z0-9]+)*$`)

// ValidateCreateRequest validates a CreateRequest against the Common
// Validation rules defined in the Vehicle Onboarding TRD. now is injected to
// make the "year" upper-bound check deterministic in tests.
func ValidateCreateRequest(req CreateRequest, now time.Time) *ValidationError {
	verr := &ValidationError{}

	vin := strings.TrimSpace(req.VIN)
	if vin == "" {
		verr.add("vin", "is required")
	} else if !vinPattern.MatchString(vin) {
		verr.add("vin", "must be a 17-character VIN using only digits and letters other than I, O, and Q")
	}

	plate := strings.TrimSpace(req.PlateNumber)
	if plate == "" {
		verr.add("plate_number", "is required")
	} else if !plateNumberPattern.MatchString(plate) {
		verr.add("plate_number", "must be alphanumeric with optional hyphen or space separators")
	}

	if strings.TrimSpace(req.Make) == "" {
		verr.add("make", "is required")
	}

	if strings.TrimSpace(req.Model) == "" {
		verr.add("model", "is required")
	}

	maxYear := now.Year() + 1
	if req.Year < 1000 || req.Year > 9999 {
		verr.add("year", "must be a valid 4-digit year")
	} else if req.Year > maxYear {
		verr.add("year", "must not be greater than the current year plus one")
	}

	if strings.TrimSpace(req.CategoryID) == "" {
		verr.add("category_id", "is required")
	}

	if req.Mileage < 0 {
		verr.add("mileage", "must be greater than or equal to 0")
	}

	if req.FuelLevel < 0 || req.FuelLevel > 100 {
		verr.add("fuel_level", "must be between 0 and 100")
	}

	if strings.TrimSpace(req.Location) == "" {
		verr.add("location", "is required")
	}

	if !verr.HasErrors() {
		return nil
	}
	return verr
}
