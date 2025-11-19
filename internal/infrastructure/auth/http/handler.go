package authhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	uc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/auth"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Handler wires HTTP routes to auth service.
type Handler struct {
	service *uc.Service
}

func NewHandler(service *uc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Get("/me/{id}", h.me)
	return r
}

// @Summary Register user
// @Description Create a new user with customer or admin role.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register payload"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// @Summary Login
// @Description Authenticate user and return JWT tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary Get profile
// @Description Retrieve profile information for a user ID.
// @Tags Auth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.ProfileResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /auth/me/{id} [get]
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	resp, err := h.service.Me(r.Context(), userID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	writeJSON(w, pkgErrors.StatusCode(err), err)
}
