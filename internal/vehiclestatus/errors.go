package vehiclestatus

import "errors"

// Sentinel errors returned by the service layer. Handlers translate these
// into the appropriate HTTP status codes.
var (
	// ErrVehicleNotFound is returned when the referenced vehicle does not
	// exist or has been deleted.
	ErrVehicleNotFound = errors.New("vehicle not found")

	// ErrUnknownEventType is returned when the submitted event type has no
	// corresponding entry in the vehicle_status_triggers configuration.
	ErrUnknownEventType = errors.New("unknown event type")

	// ErrInvalidStatus is returned when a requested status is not one of
	// the enumerated vehicle lifecycle statuses.
	ErrInvalidStatus = errors.New("invalid status")

	// ErrTransitionNotAllowed is returned when the requested transition is
	// not permitted by the state machine from the vehicle's current
	// status.
	ErrTransitionNotAllowed = errors.New("transition not allowed from current status")

	// ErrReasonRequired is returned when a manual override (or an event
	// that requires a reason) is submitted without one.
	ErrReasonRequired = errors.New("reason is required")
)
