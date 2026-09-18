package vehiclestatus

import "testing"

func TestIsValidStatus(t *testing.T) {
	for _, s := range []Status{StatusAvailable, StatusReserved, StatusRented, StatusCleaning, StatusMaintenance, StatusDamaged, StatusRetired} {
		if !IsValidStatus(s) {
			t.Errorf("IsValidStatus(%q) = false, want true", s)
		}
	}

	if IsValidStatus("bogus") {
		t.Errorf("IsValidStatus(%q) = true, want false", "bogus")
	}
}

func TestNextStatus(t *testing.T) {
	tests := []struct {
		current Status
		event   EventType
		want    Status
		wantOk  bool
	}{
		{StatusAvailable, "booking_confirmed", StatusReserved, true},
		{StatusReserved, "handover_completed", StatusRented, true},
		{StatusReserved, "booking_cancelled", StatusAvailable, true},
		{StatusRented, "return_completed", StatusCleaning, true},
		{StatusCleaning, "turnaround_completed", StatusAvailable, true},
		{StatusAvailable, "maintenance_scheduled", StatusMaintenance, true},
		{StatusReserved, "maintenance_scheduled", StatusMaintenance, true},
		{StatusMaintenance, "maintenance_completed", StatusAvailable, true},
		{StatusAvailable, "damage_reported", StatusDamaged, true},
		{StatusRented, "damage_reported", StatusDamaged, true},
		{StatusDamaged, "repair_scheduled", StatusMaintenance, true},
		{StatusAvailable, "compliance_expired", StatusMaintenance, true},
		{StatusMaintenance, "compliance_resolved", StatusAvailable, true},
		{StatusAvailable, "retired", StatusRetired, true},
		{StatusMaintenance, "retired", StatusRetired, true},
		// Not allowed: event not defined for current status.
		{StatusRented, "booking_confirmed", "", false},
		{StatusRetired, "maintenance_completed", "", false},
		{StatusAvailable, "unknown_event", "", false},
	}

	for _, tt := range tests {
		got, ok := NextStatus(tt.current, tt.event)
		if ok != tt.wantOk || (ok && got != tt.want) {
			t.Errorf("NextStatus(%q, %q) = (%q, %v), want (%q, %v)", tt.current, tt.event, got, ok, tt.want, tt.wantOk)
		}
	}
}

func TestIsTransitionAllowed(t *testing.T) {
	if !IsTransitionAllowed(StatusAvailable, StatusMaintenance) {
		t.Error("expected available -> maintenance to be allowed")
	}
	if IsTransitionAllowed(StatusAvailable, StatusRented) {
		t.Error("expected available -> rented to not be allowed (no direct edge)")
	}
	if IsTransitionAllowed(StatusRetired, StatusAvailable) {
		t.Error("expected retired -> available to not be allowed (terminal state)")
	}
}
