package vehicle

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
)

// Handler exposes the Vehicle Onboarding REST API.
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

// CreateVehicle handles POST /api/v1/vehicles.
func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body", nil)
		return
	}

	actor := ""
	if user, ok := auth.UserFromContext(r.Context()); ok {
		actor = user.Subject
	}

	created, err := h.service.CreateVehicle(r.Context(), req, actor)
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
	case errors.Is(err, ErrDuplicateVIN):
		writeError(w, http.StatusConflict, ErrDuplicateVIN.Error(), nil)
	case errors.Is(err, ErrDuplicatePlateNumber):
		writeError(w, http.StatusConflict, ErrDuplicatePlateNumber.Error(), nil)
	case errors.Is(err, ErrCategoryNotFound):
		writeError(w, http.StatusUnprocessableEntity, ErrCategoryNotFound.Error(), nil)
	default:
		log.Printf("vehicle: unexpected error creating vehicle: %v", err)
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
