package vehiclestatus

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
)

var vehicleIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Handler exposes the vehicle status transition endpoints over HTTP.
type Handler struct {
	service *Service
}

// NewHandler creates a Handler backed by the given Service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register wires the handler's routes onto mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/vehicles/{vehicleId}/status-events", h.handleStatusEvent)
	mux.HandleFunc("PATCH /api/v1/vehicles/{vehicleId}/status", h.handleStatusOverride)
}

type statusEventRequest struct {
	EventType    string     `json:"eventType"`
	EffectiveAt  *time.Time `json:"effectiveAt"`
	Reason       string     `json:"reason"`
	SourceSystem string     `json:"sourceSystem"`
}

type statusOverrideRequest struct {
	NewStatus string `json:"newStatus"`
	Reason    string `json:"reason"`
}

type transitionResponse struct {
	VehicleID   string    `json:"vehicleId"`
	Status      string    `json:"status"`
	StatusSince time.Time `json:"statusSince"`
}

func (h *Handler) handleStatusEvent(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.PathValue("vehicleId")
	if !vehicleIDPattern.MatchString(vehicleID) {
		writeError(w, http.StatusBadRequest, "vehicleId must be a valid UUID")
		return
	}

	var req statusEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.EventType == "" {
		writeError(w, http.StatusBadRequest, "eventType is required")
		return
	}

	var effectiveAt time.Time
	if req.EffectiveAt != nil {
		effectiveAt = *req.EffectiveAt
	}

	actor := ""
	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		actor = claims.Subject
	}

	result, err := h.service.ApplyEvent(r.Context(), ApplyEventInput{
		VehicleID:    vehicleID,
		EventType:    EventType(req.EventType),
		EffectiveAt:  effectiveAt,
		Reason:       req.Reason,
		SourceSystem: req.SourceSystem,
		Actor:        actor,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, transitionResponse{
		VehicleID:   result.VehicleID,
		Status:      string(result.Status),
		StatusSince: result.StatusSince,
	})
}

func (h *Handler) handleStatusOverride(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.PathValue("vehicleId")
	if !vehicleIDPattern.MatchString(vehicleID) {
		writeError(w, http.StatusBadRequest, "vehicleId must be a valid UUID")
		return
	}

	var req statusOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !IsValidStatus(Status(req.NewStatus)) {
		writeError(w, http.StatusBadRequest, "newStatus must be one of the enumerated vehicle statuses")
		return
	}
	if len(req.Reason) > 500 {
		writeError(w, http.StatusBadRequest, "reason must not exceed 500 characters")
		return
	}

	actor := ""
	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		actor = claims.Subject
	}

	result, err := h.service.Override(r.Context(), OverrideInput{
		VehicleID: vehicleID,
		NewStatus: Status(req.NewStatus),
		Reason:    req.Reason,
		Actor:     actor,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, transitionResponse{
		VehicleID:   result.VehicleID,
		Status:      string(result.Status),
		StatusSince: result.StatusSince,
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrVehicleNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrUnknownEventType), errors.Is(err, ErrInvalidStatus), errors.Is(err, ErrReasonRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrTransitionNotAllowed):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
