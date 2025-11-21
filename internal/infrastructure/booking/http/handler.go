package bookinghttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/utils"
)

// Handler exposes booking endpoints.
type Handler struct {
	service *booking.Service
}

type bookingCreateInput struct {
	UserID     string `json:"user_id"`
	RoomTypeID string `json:"room_type_id"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	Guests     int    `json:"guests,omitempty"`
}

func NewHandler(service *booking.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/bookings", h.listBookings)
	r.Post("/bookings", h.createBooking)
	r.Get("/bookings/{id}", h.getBooking)
	r.Get("/bookings/{id}/status", h.getStatus)
	r.Post("/bookings/{id}/cancel", h.cancelBooking)
	r.Post("/bookings/{id}/status", h.updateStatus)
	r.Post("/bookings/{id}/checkpoint", h.checkpoint)
	return r
}

// @Summary Create booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Param request body bookingCreateInput true "Booking payload"
// @Success 201 {object} dto.BookingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /bookings [post]
func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	var input bookingCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	req, err := toBookingRequest(input)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}

	resp, err := h.service.CreateBooking(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(resp.ID, "booking", "/api/v1/bookings/"+resp.ID, resp)
	utils.Respond(w, http.StatusCreated, "booking created", resource)
}

// @Summary Cancel booking
// @Tags Bookings
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} dto.BookingResponse
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
	resp, err := h.service.GetBooking(r.Context(), bookingID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(resp.ID, "booking", "/api/v1/bookings/"+resp.ID, resp)
	utils.Respond(w, http.StatusOK, "booking cancelled", resource)
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
	resource := utils.NewResource(resp.ID, "booking", "/api/v1/bookings/"+resp.ID, resp)
	utils.Respond(w, http.StatusOK, "booking retrieved", resource)
}

// @Summary Get booking status
// @Tags Bookings
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /bookings/{id}/status [get]
func (h *Handler) getStatus(w http.ResponseWriter, r *http.Request) {
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
	utils.Respond(w, http.StatusOK, "booking status retrieved", map[string]string{"status": resp.Status})
}

// @Summary List bookings
// @Tags Bookings
// @Produce json
// @Success 200 {array} dto.BookingResponse
// @Security BearerAuth
// @Router /bookings [get]
func (h *Handler) listBookings(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	resp, err := h.service.ListBookings(r.Context(), opts)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	var resources []utils.Resource
	for _, b := range resp {
		resources = append(resources, utils.NewResource(b.ID, "booking", "/api/v1/bookings/"+b.ID, b))
	}
	utils.RespondWithCount(w, http.StatusOK, "bookings listed", resources, len(resources))
}

// @Summary Booking checkpoint
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param request body dto.CheckpointRequest true "Checkpoint payload"
// @Success 200 {object} dto.BookingResponse
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
	resp, err := h.service.GetBooking(r.Context(), bookingID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(resp.ID, "booking", "/api/v1/bookings/"+resp.ID, resp)
	utils.Respond(w, http.StatusOK, "status updated", resource)
}

// @Summary Update booking status (internal)
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param request body map[string]string true "Status payload"
// @Success 200 {object} dto.BookingResponse
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
	resp, err := h.service.GetBooking(r.Context(), bookingID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(resp.ID, "booking", "/api/v1/bookings/"+resp.ID, resp)
	utils.Respond(w, http.StatusOK, "booking status updated", resource)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	utils.Respond(w, pkgErrors.StatusCode(err), err.Message, err)
}

func toBookingRequest(in bookingCreateInput) (dto.BookingRequest, error) {
	checkIn, err := parseDate(in.CheckIn)
	if err != nil {
		return dto.BookingRequest{}, pkgErrors.New("bad_request", "invalid check_in")
	}
	checkOut, err := parseDate(in.CheckOut)
	if err != nil {
		return dto.BookingRequest{}, pkgErrors.New("bad_request", "invalid check_out")
	}
	return dto.BookingRequest{
		UserID:     in.UserID,
		RoomTypeID: in.RoomTypeID,
		CheckIn:    checkIn,
		CheckOut:   checkOut,
		Guests:     in.Guests,
	}, nil
}

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, pkgErrors.New("bad_request", "date required")
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", value)
}

func parseQueryOptions(r *http.Request) query.Options {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return query.Options{Limit: limit, Offset: offset}
}
