package maintenanceschedule

import (
	"context"
	"errors"
)

// Sentinel errors returned by Repository implementations, mapped to the
// appropriate HTTP status by the handler.
var (
	// ErrVehicleNotFound is returned when vehicleId does not reference an
	// existing, non-deleted vehicle.
	ErrVehicleNotFound = errors.New("maintenanceschedule: vehicle_id does not reference an existing vehicle")
	// ErrVehicleCategoryNotFound is returned when vehicleCategoryId does
	// not reference an existing, non-deleted vehicle category.
	ErrVehicleCategoryNotFound = errors.New("maintenanceschedule: vehicle_category_id does not reference an existing vehicle category")
)

// Repository persists and retrieves maintenance schedule rules.
type Repository interface {
	// CreateSchedule inserts a new maintenance schedule rule. It returns
	// ErrVehicleNotFound or ErrVehicleCategoryNotFound when the referenced
	// vehicle/category does not exist.
	CreateSchedule(ctx context.Context, s Schedule) (Schedule, error)
}
