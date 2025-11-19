package paymenthttp

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/payment"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Handler exposes payment endpoints.
type Handler struct {
	service *payment.Service
}

func NewHandler(service *payment.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/payments", h.createPayment)
	r.Get("/payments/{id}", h.getPayment)
	r.Post("/payments/webhook", h.handleWebhook)
	r.Post("/payments/refund", h.refund)
	return r
}

// @Summary Initiate payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body dto.PaymentRequest true "Payment payload"
// @Success 201 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /payments [post]
func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	var req dto.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	resp, err := h.service.Initiate(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// @Summary Payment webhook
// @Tags Payments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Router /payments/webhook [post]
func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req dto.WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid webhook"))
		return
	}
	if err := h.service.HandleWebhook(r.Context(), req, string(body)); err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

// @Summary Refund payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body dto.RefundRequest true "Refund payload"
// @Success 200 {object} dto.RefundResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /payments/refund [post]
func (h *Handler) refund(w http.ResponseWriter, r *http.Request) {
	var req dto.RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	resp, err := h.service.Refund(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary Get payment
// @Tags Payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /payments/{id} [get]
func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	paymentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	resp, err := h.service.GetPayment(r.Context(), paymentID)
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
