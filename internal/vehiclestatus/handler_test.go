package vehiclestatus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testVehicleID = "11111111-1111-1111-1111-111111111111"

func newTestHandler() (*Handler, *fakeRepository) {
	repo := newFakeRepository()
	repo.vehicles[testVehicleID] = VehicleStatus{VehicleID: testVehicleID, Status: StatusAvailable, StatusSince: time.Now()}
	repo.triggers["booking_confirmed"] = Trigger{EventType: "booking_confirmed", ResultingStatus: StatusReserved}
	repo.triggers["damage_reported"] = Trigger{EventType: "damage_reported", ResultingStatus: StatusDamaged, RequiresReason: true}
	return NewHandler(NewService(repo)), repo
}

func newMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestHandleStatusEventSuccess(t *testing.T) {
	h, _ := newTestHandler()
	mux := newMux(h)

	body := bytes.NewBufferString(`{"eventType":"booking_confirmed","sourceSystem":"booking"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/"+testVehicleID+"/status-events", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp transitionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling response: %v", err)
	}
	if resp.Status != string(StatusReserved) {
		t.Errorf("resp.Status = %q, want %q", resp.Status, StatusReserved)
	}
}

func TestHandleStatusEventInvalidVehicleID(t *testing.T) {
	h, _ := newTestHandler()
	mux := newMux(h)

	body := bytes.NewBufferString(`{"eventType":"booking_confirmed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/not-a-uuid/status-events", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleStatusEventUnknownVehicle(t *testing.T) {
	h, _ := newTestHandler()
	mux := newMux(h)

	unknownID := "22222222-2222-2222-2222-222222222222"
	body := bytes.NewBufferString(`{"eventType":"booking_confirmed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/"+unknownID+"/status-events", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandleStatusEventConflict(t *testing.T) {
	h, repo := newTestHandler()
	repo.vehicles[testVehicleID] = VehicleStatus{VehicleID: testVehicleID, Status: StatusRented}
	mux := newMux(h)

	body := bytes.NewBufferString(`{"eventType":"booking_confirmed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/"+testVehicleID+"/status-events", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestHandleStatusEventMissingReason(t *testing.T) {
	h, _ := newTestHandler()
	mux := newMux(h)

	body := bytes.NewBufferString(`{"eventType":"damage_reported"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/"+testVehicleID+"/status-events", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestHandleStatusOverrideSuccess(t *testing.T) {
	h, _ := newTestHandler()
	mux := newMux(h)

	body := bytes.NewBufferString(`{"newStatus":"maintenance","reason":"scheduled service"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/vehicles/"+testVehicleID+"/status", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp transitionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling response: %v", err)
	}
	if resp.Status != string(StatusMaintenance) {
		t.Errorf("resp.Status = %q, want %q", resp.Status, StatusMaintenance)
	}
}

func TestHandleStatusOverrideMissingReason(t *testing.T) {
	h, _ := newTestHandler()
	mux := newMux(h)

	body := bytes.NewBufferString(`{"newStatus":"maintenance"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/vehicles/"+testVehicleID+"/status", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleStatusOverrideInvalidStatus(t *testing.T) {
	h, _ := newTestHandler()
	mux := newMux(h)

	body := bytes.NewBufferString(`{"newStatus":"bogus","reason":"x"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/vehicles/"+testVehicleID+"/status", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
