package vehicle

import "fmt"

// FieldError describes a single validation failure on a request field.
type FieldError struct {
	Field   string
	Message string
}

// ValidationError aggregates one or more FieldErrors found while validating a
// request.
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s: %s", e.Errors[0].Field, e.Errors[0].Message)
}

func (e *ValidationError) add(field, message string) {
	e.Errors = append(e.Errors, FieldError{Field: field, Message: message})
}

// HasErrors reports whether any field errors have been recorded.
func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Sentinel errors returned by the vehicle service and repository. Handlers
// map these to HTTP status codes.
var (
	// ErrDuplicateVIN is returned when the requested VIN is already assigned
	// to another vehicle.
	ErrDuplicateVIN = fmt.Errorf("vin is already assigned to another vehicle")
	// ErrDuplicatePlateNumber is returned when the requested plate number is
	// already assigned to another active vehicle.
	ErrDuplicatePlateNumber = fmt.Errorf("plate_number is already assigned to another vehicle")
	// ErrCategoryNotFound is returned when category_id does not reference an
	// existing vehicle category.
	ErrCategoryNotFound = fmt.Errorf("category_id does not reference an existing vehicle category")
	// ErrCategoryInactive is returned when category_id references a vehicle
	// category that is not active.
	ErrCategoryInactive = fmt.Errorf("category_id references an inactive vehicle category")
)
