package vehicle

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// vinPattern matches the standard 17-character VIN format, excluding the
// letters I, O, and Q as specified in the Vehicle Onboarding TRD.
var vinPattern = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)

// plateNumberPattern allows alphanumeric plate numbers with optional
// hyphen/space separators.
var plateNumberPattern = regexp.MustCompile(`^[A-Za-z0-9]+([- ][A-Za-z0-9]+)*$`)

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

// validate checks req against the Common Validation rules defined in the
// Vehicle Onboarding TRD and returns any violations found.
func (req CreateRequest) validate(now time.Time) ValidationErrors {
	var errs ValidationErrors

	if req.VIN == "" {
		errs = append(errs, ValidationError{"vin", "is required"})
	} else if !vinPattern.MatchString(req.VIN) {
		errs = append(errs, ValidationError{"vin", "must be a 17-character VIN using only letters (excluding I, O, Q) and digits"})
	}

	if req.PlateNumber == "" {
		errs = append(errs, ValidationError{"plate_number", "is required"})
	} else if !plateNumberPattern.MatchString(req.PlateNumber) {
		errs = append(errs, ValidationError{"plate_number", "must be alphanumeric with optional hyphen/space separators"})
	}

	if req.Make == "" {
		errs = append(errs, ValidationError{"make", "is required"})
	}

	if req.Model == "" {
		errs = append(errs, ValidationError{"model", "is required"})
	}

	maxYear := now.Year() + 1
	if req.Year < 1000 || req.Year > 9999 {
		errs = append(errs, ValidationError{"year", "must be a valid 4-digit year"})
	} else if req.Year > maxYear {
		errs = append(errs, ValidationError{"year", fmt.Sprintf("must not be greater than %d", maxYear)})
	}

	if req.CategoryID == "" {
		errs = append(errs, ValidationError{"category_id", "is required"})
	}

	if req.Mileage < 0 {
		errs = append(errs, ValidationError{"mileage", "must be greater than or equal to 0"})
	}

	if req.FuelLevel < 0 || req.FuelLevel > 100 {
		errs = append(errs, ValidationError{"fuel_level", "must be between 0 and 100"})
	}

	if req.Location == "" {
		errs = append(errs, ValidationError{"location", "is required"})
	}

	return errs
}
