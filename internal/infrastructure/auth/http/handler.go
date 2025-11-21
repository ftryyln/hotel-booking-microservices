package authhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	uc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/auth"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/auth/assembler"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/utils"
)

// Handler wires HTTP routes to auth service.
type Handler struct {
	service   *uc.Service
	jwtSecret string
}

func NewHandler(service *uc.Service, jwtSecret string) *Handler {
	return &Handler{service: service, jwtSecret: jwtSecret}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Get("/me/{id}", h.me)
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWT(h.jwtSecret))
		r.Get("/users", h.listUsers)
		r.Get("/users/{id}", h.getUser)
	})
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
	resource := utils.NewResource(resp.ID, "user", "/auth/me/"+resp.ID, resp)
	utils.Respond(w, http.StatusCreated, "user registered", resource)
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
	resource := utils.NewResource(resp.ID, "user", "/auth/me/"+resp.ID, resp)
	utils.Respond(w, http.StatusOK, "login succeeded", resource)
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
	user, err := h.service.Me(r.Context(), userID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resp := assembler.ToProfile(user)
	resource := utils.NewResource(resp.ID, "user", "/auth/me/"+resp.ID, resp)
	utils.Respond(w, http.StatusOK, "profile retrieved", resource)
}

// @Summary List users (admin)
// @Tags Auth
// @Produce json
// @Success 200 {array} dto.ProfileResponse
// @Failure 403 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /auth/users [get]
func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(r) {
		writeError(w, pkgErrors.New("forbidden", "admin only"))
		return
	}
	opts := parseQueryOptions(r)
	users, err := h.service.List(r.Context(), opts)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	var resources []utils.Resource
	for _, u := range users {
		dto := assembler.ToProfile(u)
		resources = append(resources, utils.NewResource(dto.ID, "user", "/auth/users/"+dto.ID, dto))
	}
	utils.RespondWithCount(w, http.StatusOK, "users listed", resources, len(resources))
}

// @Summary Get user detail
// @Tags Auth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.ProfileResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /auth/users/{id} [get]
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	claims, ok := r.Context().Value(middleware.AuthContextKey).(*middleware.Claims)
	if !ok || (claims.Role != "admin" && claims.UserID != userID.String()) {
		writeError(w, pkgErrors.New("forbidden", "insufficient role"))
		return
	}
	user, err := h.service.Get(r.Context(), userID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dto := assembler.ToProfile(user)
	resource := utils.NewResource(dto.ID, "user", "/auth/users/"+dto.ID, dto)
	utils.Respond(w, http.StatusOK, "user retrieved", resource)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	utils.Respond(w, pkgErrors.StatusCode(err), err.Message, err)
}

func isAdmin(r *http.Request) bool {
	if claims, ok := r.Context().Value(middleware.AuthContextKey).(*middleware.Claims); ok {
		return claims.Role == "admin"
	}
	return false
}

func parseQueryOptions(r *http.Request) query.Options {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return query.Options{Limit: limit, Offset: offset}
}
