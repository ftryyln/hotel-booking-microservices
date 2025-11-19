package notificationhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	notificationuc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Handler exposes notification endpoint.
type Handler struct {
	service *notificationuc.Service
}

func NewHandler(service *notificationuc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/notifications", h.send)
	return r
}

// @Summary Send notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Param request body dto.NotificationRequest true "Notification payload"
// @Success 202 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications [post]
func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	var req dto.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	if err := h.service.Send(r.Context(), req); err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "sent"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	writeJSON(w, pkgErrors.StatusCode(err), err)
}
