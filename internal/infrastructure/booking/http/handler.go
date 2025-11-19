package bookinghttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Handler exposes booking endpoints.
type Handler struct {
	service *booking.Service
}

func NewHandler(service *booking.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/bookings", h.createBooking)
	r.Get("/bookings/{id}", h.getBooking)
	r.Post("/bookings/{id}/cancel", h.cancelBooking)
	r.Post("/bookings/{id}/status", h.updateStatus)
	r.Post("/bookings/{id}/checkpoint", h.checkpoint)
	return r
}

// @Summary Create booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Param request body dto.BookingRequest true "Booking payload"
// @Success 201 {object} dto.BookingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /bookings [post]
func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	var req dto.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	resp, err := h.service.CreateBooking(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// @Summary Cancel booking
// @Tags Bookings
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /bookings/{id}/cancel [post]
func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	if err := h.service.CancelBooking(r.Context(), bookingID); err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// @Summary Get booking
// @Tags Bookings
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /bookings/{id} [get]
func (h *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	resp, err := h.service.GetBooking(r.Context(), bookingID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary Booking checkpoint
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param request body dto.CheckpointRequest true "Checkpoint payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /bookings/{id}/checkpoint [post]
func (h *Handler) checkpoint(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	var req dto.CheckpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	if err := h.service.Checkpoint(r.Context(), bookingID, req.Action); err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": req.Action})
}

// @Summary Update booking status (internal)
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param request body map[string]string true "Status payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Router /bookings/{id}/status [post]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	if err := h.service.ApplyStatus(r.Context(), bookingID, payload.Status); err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": payload.Status})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	writeJSON(w, pkgErrors.StatusCode(err), err)
}
