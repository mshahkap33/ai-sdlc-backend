package maintenanceschedule

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
)

func newTestHandler(repo Repository) *Handler {
	return NewHandler(NewService(repo))
}

func doCreateRequest(h *Handler, body any) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/maintenance-schedules", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, req)
	return rec
}

func TestHandlerCreateScheduleSuccess(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, s Schedule) (Schedule, error) {
			s.ID = "schedule-1"
			s.Active = true
			return s, nil
		},
	}
	h := newTestHandler(repo)

	rec := doCreateRequest(h, validDateIntervalRequest())

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp CreateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.ID != "schedule-1" {
		t.Errorf("resp.ID = %q, want %q", resp.ID, "schedule-1")
	}
	if !resp.Active {
		t.Errorf("resp.Active = %v, want true", resp.Active)
	}
}

func TestHandlerCreateScheduleValidationError(t *testing.T) {
	repo := &fakeRepository{}
	h := newTestHandler(repo)

	req := validDateIntervalRequest()
	req.IntervalDays = nil

	rec := doCreateRequest(h, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.called {
		t.Fatal("expected repository not to be called for invalid request")
	}
}

func TestHandlerCreateScheduleVehicleCategoryNotFound(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, s Schedule) (Schedule, error) {
			return Schedule{}, ErrVehicleCategoryNotFound
		},
	}
	h := newTestHandler(repo)

	rec := doCreateRequest(h, validDateIntervalRequest())

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestHandlerCreateScheduleVehicleNotFound(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, s Schedule) (Schedule, error) {
			return Schedule{}, ErrVehicleNotFound
		},
	}
	h := newTestHandler(repo)

	req := CreateRequest{
		VehicleID:    "22222222-2222-2222-2222-222222222222",
		TriggerType:  string(TriggerTypeDateInterval),
		IntervalDays: intPtr(180),
	}
	rec := doCreateRequest(h, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestHandlerCreateScheduleInvalidJSON(t *testing.T) {
	repo := &fakeRepository{}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/maintenance-schedules", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerCreateScheduleUsesAuthenticatedUserAsActor(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, s Schedule) (Schedule, error) {
			s.ID = "schedule-1"
			return s, nil
		},
	}
	h := newTestHandler(repo)

	data, _ := json.Marshal(validDateIntervalRequest())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/maintenance-schedules", bytes.NewReader(data))
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{Subject: "staff-42", Roles: []string{"service_staff"}}))
	rec := httptest.NewRecorder()

	h.CreateSchedule(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if repo.received.CreatedBy != "staff-42" {
		t.Errorf("received.CreatedBy = %q, want %q", repo.received.CreatedBy, "staff-42")
	}
}
