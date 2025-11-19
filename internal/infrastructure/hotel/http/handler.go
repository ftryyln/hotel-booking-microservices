package hotelhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	hoteluc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Handler exposes hotel endpoints.
type Handler struct {
	service *hoteluc.Service
}

func NewHandler(service *hoteluc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/hotels", h.createHotel)
	r.Get("/hotels", h.listHotels)
	r.Post("/room-types", h.createRoomType)
	r.Post("/rooms", h.createRoom)
	return r
}

// @Summary Create hotel
// @Tags Hotels
// @Accept json
// @Produce json
// @Param request body dto.HotelRequest true "Hotel payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /hotels [post]
func (h *Handler) createHotel(w http.ResponseWriter, r *http.Request) {
	var req dto.HotelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	id, err := h.service.CreateHotel(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id.String()})
}

// @Summary List hotels
// @Tags Hotels
// @Produce json
// @Success 200 {array} dto.HotelResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /hotels [get]
func (h *Handler) listHotels(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.ListHotels(r.Context())
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary Create room type
// @Tags Hotels
// @Accept json
// @Produce json
// @Param request body dto.RoomTypeRequest true "Room type payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /room-types [post]
func (h *Handler) createRoomType(w http.ResponseWriter, r *http.Request) {
	var req dto.RoomTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	id, err := h.service.CreateRoomType(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id.String()})
}

// @Summary Create room
// @Tags Hotels
// @Accept json
// @Produce json
// @Param request body dto.RoomRequest true "Room payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /rooms [post]
func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	var req dto.RoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	id, err := h.service.CreateRoom(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id.String()})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	writeJSON(w, pkgErrors.StatusCode(err), err)
}
