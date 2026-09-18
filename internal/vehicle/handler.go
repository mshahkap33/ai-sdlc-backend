package vehicle

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
)

// createResponse is the JSON representation of a vehicle returned by the
// Create Vehicle API, matching the Vehicle Onboarding TRD response body.
type createResponse struct {
	ID             string  `json:"id"`
	VIN            string  `json:"vin"`
	PlateNumber    string  `json:"plate_number"`
	Make           string  `json:"make"`
	Model          string  `json:"model"`
	Year           int     `json:"year"`
	CategoryID     string  `json:"category_id"`
	Mileage        float64 `json:"mileage"`
	FuelLevel      float64 `json:"fuel_level"`
	Location       string  `json:"location"`
	ConditionNotes *string `json:"condition_notes,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// errorResponse is the JSON representation of an API error.
type errorResponse struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

// Handler exposes the Vehicle Onboarding REST API endpoints.
type Handler struct {
	service *Service
}

// NewHandler builds a Handler backed by the given Service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register wires the Vehicle Onboarding routes onto mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/vehicles", h.CreateVehicle)
}

// CreateVehicle handles POST /api/v1/vehicles.
func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	actor, ok := auth.UserFromContext(r.Context())
	if !ok || !actor.HasRole("service_staff") {
		writeError(w, http.StatusForbidden, "service_staff role is required to create vehicles", nil)
		return
	}

	var req CreateRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON matching the Create Vehicle schema", nil)
		return
	}

	v, err := h.service.Create(r.Context(), req, actor.Subject)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toCreateResponse(v))
}

func writeServiceError(w http.ResponseWriter, err error) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		writeError(w, http.StatusBadRequest, "validation failed", verr.Errors)
	case errors.Is(err, ErrDuplicateVIN), errors.Is(err, ErrDuplicatePlateNumber):
		writeError(w, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, ErrCategoryNotFound), errors.Is(err, ErrCategoryInactive):
		writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
	default:
		log.Printf("vehicle: internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
	}
}

func toCreateResponse(v *Vehicle) createResponse {
	return createResponse{
		ID:             v.ID,
		VIN:            v.VIN,
		PlateNumber:    v.PlateNumber,
		Make:           v.Make,
		Model:          v.Model,
		Year:           v.Year,
		CategoryID:     v.CategoryID,
		Mileage:        v.Mileage,
		FuelLevel:      v.FuelLevel,
		Location:       v.Location,
		ConditionNotes: v.ConditionNotes,
		CreatedAt:      v.CreatedAt.Format(rfc3339Milli),
	}
}

const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string, fields []FieldError) {
	writeJSON(w, status, errorResponse{Error: message, Fields: fields})
}
