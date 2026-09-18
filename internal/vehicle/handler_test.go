package vehicle

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
	return NewHandler(newTestService(repo))
}

func doCreateRequest(h *Handler, body any) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	h.CreateVehicle(rec, req)
	return rec
}

func TestHandlerCreateVehicleSuccess(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, v Vehicle) (Vehicle, error) {
			v.ID = "vehicle-1"
			return v, nil
		},
	}
	h := newTestHandler(repo)

	rec := doCreateRequest(h, validRequest())

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp CreateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.ID != "vehicle-1" {
		t.Errorf("resp.ID = %q, want %q", resp.ID, "vehicle-1")
	}
}

func TestHandlerCreateVehicleValidationError(t *testing.T) {
	repo := &fakeRepository{}
	h := newTestHandler(repo)

	req := validRequest()
	req.VIN = ""

	rec := doCreateRequest(h, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.called {
		t.Fatal("expected repository not to be called for invalid request")
	}
}

func TestHandlerCreateVehicleDuplicateVIN(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, v Vehicle) (Vehicle, error) {
			return Vehicle{}, ErrDuplicateVIN
		},
	}
	h := newTestHandler(repo)

	rec := doCreateRequest(h, validRequest())

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestHandlerCreateVehicleCategoryNotFound(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, v Vehicle) (Vehicle, error) {
			return Vehicle{}, ErrCategoryNotFound
		},
	}
	h := newTestHandler(repo)

	rec := doCreateRequest(h, validRequest())

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
}

func TestHandlerCreateVehicleInvalidJSON(t *testing.T) {
	repo := &fakeRepository{}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	h.CreateVehicle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerCreateVehicleUsesAuthenticatedUserAsActor(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, v Vehicle) (Vehicle, error) {
			v.ID = "vehicle-1"
			return v, nil
		},
	}
	h := newTestHandler(repo)

	data, _ := json.Marshal(validRequest())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(data))
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{Subject: "staff-42", Roles: []string{"service_staff"}}))
	rec := httptest.NewRecorder()

	h.CreateVehicle(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if repo.received.CreatedBy != "staff-42" {
		t.Errorf("received.CreatedBy = %q, want %q", repo.received.CreatedBy, "staff-42")
	}
}
