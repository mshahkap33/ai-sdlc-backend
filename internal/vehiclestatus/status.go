// Package vehiclestatus implements the "Manage Vehicle Status and
// Availability" backend capability: the vehicle lifecycle state machine, the
// event-driven and manual transition endpoints, and the audit trail of
// status changes.
package vehiclestatus

// Status is a vehicle lifecycle status, as defined by the
// vehicle_status_availability technical requirement document.
type Status string

// The full set of vehicle lifecycle statuses.
const (
	StatusAvailable   Status = "available"
	StatusReserved    Status = "reserved"
	StatusRented      Status = "rented"
	StatusCleaning    Status = "cleaning"
	StatusMaintenance Status = "maintenance"
	StatusDamaged     Status = "damaged"
	StatusRetired     Status = "retired"
)

// validStatuses enumerates every status accepted by the API.
var validStatuses = map[Status]bool{
	StatusAvailable:   true,
	StatusReserved:    true,
	StatusRented:      true,
	StatusCleaning:    true,
	StatusMaintenance: true,
	StatusDamaged:     true,
	StatusRetired:     true,
}

// IsValidStatus reports whether s is one of the enumerated vehicle statuses.
func IsValidStatus(s Status) bool {
	return validStatuses[s]
}

// EventType identifies a triggering event that can move a vehicle from one
// status to another (e.g. "booking_confirmed").
type EventType string

// transitions encodes the state machine from the TRD's state diagram: for a
// given current status and event type, it yields the resulting status. Any
// (status, event) pair absent from this table is not an allowed transition.
var transitions = map[Status]map[EventType]Status{
	StatusAvailable: {
		"booking_confirmed":     StatusReserved,
		"maintenance_scheduled": StatusMaintenance,
		"damage_reported":       StatusDamaged,
		"compliance_expired":    StatusMaintenance,
		"retired":               StatusRetired,
	},
	StatusReserved: {
		"handover_completed":    StatusRented,
		"booking_cancelled":     StatusAvailable,
		"maintenance_scheduled": StatusMaintenance,
	},
	StatusRented: {
		"return_completed": StatusCleaning,
		"damage_reported":  StatusDamaged,
	},
	StatusCleaning: {
		"turnaround_completed": StatusAvailable,
	},
	StatusMaintenance: {
		"maintenance_completed": StatusAvailable,
		"compliance_resolved":   StatusAvailable,
		"retired":               StatusRetired,
	},
	StatusDamaged: {
		"repair_scheduled": StatusMaintenance,
	},
}

// NextStatus returns the resulting status for the given current status and
// event type, and whether that transition is allowed by the state machine.
func NextStatus(current Status, event EventType) (Status, bool) {
	byEvent, ok := transitions[current]
	if !ok {
		return "", false
	}
	next, ok := byEvent[event]
	return next, ok
}

// IsTransitionAllowed reports whether moving from current to target is
// reachable via any single event defined in the state machine. It is used to
// validate manual status overrides, which are not tied to a specific event.
func IsTransitionAllowed(current, target Status) bool {
	for _, next := range transitions[current] {
		if next == target {
			return true
		}
	}
	return false
}
