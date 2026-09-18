package vehicle

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
)

func TestCreateVehicleHandlerSuccess(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true}
	h := NewHandler(newService(repo))

	body, err := json.Marshal(validRequest())
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(body))
	req = req.WithContext(auth.NewTestContext(req.Context(), auth.User{Subject: "staff-1", Roles: []string{"service_staff"}}))
	rr := httptest.NewRecorder()

	h.CreateVehicle(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp createResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.ID == "" {
		t.Fatalf("response ID is empty")
	}
}

func TestCreateVehicleHandlerForbiddenWithoutRole(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true}
	h := NewHandler(newService(repo))

	body, _ := json.Marshal(validRequest())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(body))
	req = req.WithContext(auth.NewTestContext(req.Context(), auth.User{Subject: "staff-1", Roles: []string{"delivery_staff"}}))
	rr := httptest.NewRecorder()

	h.CreateVehicle(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestCreateVehicleHandlerUnauthenticated(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true}
	h := NewHandler(newService(repo))

	body, _ := json.Marshal(validRequest())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.CreateVehicle(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestCreateVehicleHandlerValidationError(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true}
	h := NewHandler(newService(repo))

	req := validRequest()
	req.VIN = ""
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(body))
	r = r.WithContext(auth.NewTestContext(r.Context(), auth.User{Subject: "staff-1", Roles: []string{"service_staff"}}))
	rr := httptest.NewRecorder()

	h.CreateVehicle(rr, r)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestCreateVehicleHandlerCategoryNotFound(t *testing.T) {
	repo := &fakeRepository{categoryFound: false}
	h := NewHandler(newService(repo))

	body, _ := json.Marshal(validRequest())
	r := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(body))
	r = r.WithContext(auth.NewTestContext(r.Context(), auth.User{Subject: "staff-1", Roles: []string{"service_staff"}}))
	rr := httptest.NewRecorder()

	h.CreateVehicle(rr, r)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusUnprocessableEntity, rr.Body.String())
	}
}

func TestCreateVehicleHandlerDuplicateVIN(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true, createErr: ErrDuplicateVIN}
	h := NewHandler(newService(repo))

	body, _ := json.Marshal(validRequest())
	r := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(body))
	r = r.WithContext(auth.NewTestContext(r.Context(), auth.User{Subject: "staff-1", Roles: []string{"service_staff"}}))
	rr := httptest.NewRecorder()

	h.CreateVehicle(rr, r)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusConflict, rr.Body.String())
	}
}
