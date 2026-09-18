package vehicle

import (
	"context"
	"errors"
)

// Sentinel errors returned by Repository implementations, mapped to the
// appropriate HTTP status by the handler.
var (
	// ErrDuplicateVIN is returned when the requested VIN is already
	// assigned to another vehicle.
	ErrDuplicateVIN = errors.New("vehicle: vin already assigned to another vehicle")
	// ErrDuplicatePlateNumber is returned when the requested plate number
	// is already assigned to another active vehicle.
	ErrDuplicatePlateNumber = errors.New("vehicle: plate_number already assigned to another vehicle")
	// ErrCategoryNotFound is returned when category_id does not reference
	// an existing, active vehicle category.
	ErrCategoryNotFound = errors.New("vehicle: category_id does not reference an existing, active category")
)

// Repository persists and retrieves vehicle master data records.
type Repository interface {
	// CreateVehicle inserts a new vehicle record. It returns
	// ErrDuplicateVIN, ErrDuplicatePlateNumber, or ErrCategoryNotFound when
	// the corresponding validation fails at the data layer.
	CreateVehicle(ctx context.Context, v Vehicle) (Vehicle, error)
}
