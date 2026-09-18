package maintenanceschedule

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
)

// Handler exposes the Maintenance Schedule REST API.
type Handler struct {
	service *Service
}

// NewHandler creates a Handler backed by service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// errorResponse is the JSON body returned for failed requests.
type errorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// CreateSchedule handles POST /api/v1/maintenance-schedules.
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body", nil)
		return
	}

	actor := ""
	if user, ok := auth.UserFromContext(r.Context()); ok {
		actor = user.Subject
	}

	created, err := h.service.CreateSchedule(r.Context(), req, actor)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toResponse(created))
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validationErrs ValidationErrors
	if errors.As(err, &validationErrs) {
		fields := make(map[string]string, len(validationErrs))
		for _, fieldErr := range validationErrs {
			fields[fieldErr.Field] = fieldErr.Message
		}
		writeError(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	switch {
	case errors.Is(err, ErrVehicleNotFound):
		writeError(w, http.StatusNotFound, ErrVehicleNotFound.Error(), nil)
	case errors.Is(err, ErrVehicleCategoryNotFound):
		writeError(w, http.StatusNotFound, ErrVehicleCategoryNotFound.Error(), nil)
	default:
		log.Printf("maintenanceschedule: unexpected error creating schedule: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
	}
}

func writeError(w http.ResponseWriter, status int, message string, fields map[string]string) {
	writeJSON(w, status, errorResponse{Error: message, Fields: fields})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
