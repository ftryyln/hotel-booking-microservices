package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
)

// Handler exposes gateway features.
type Handler struct {
	bookingURL string
	paymentURL string
	proxy      *httputil.ReverseProxy
	client     *http.Client
	limit      int
}

func NewHandler(bookingURL, paymentURL, aggregateTarget string, limit int) *Handler {
	target, _ := url.Parse(aggregateTarget)
	return &Handler{
		bookingURL: bookingURL,
		paymentURL: paymentURL,
		proxy:      httputil.NewSingleHostReverseProxy(target),
		client:     &http.Client{Timeout: 5 * time.Second},
		limit:      limit,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RateLimiter(h.limit))
	r.Get("/aggregate/bookings/{id}", h.aggregateBooking)
	r.Mount("/proxy", http.StripPrefix("/proxy", h.proxy))
	return r
}

// @Summary Aggregate booking data
// @Tags Gateway
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} dto.BookingAggregateResponse
// @Failure 502 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /gateway/aggregate/bookings/{id} [get]
func (h *Handler) aggregateBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	bookingResp, apiErr := h.fetch(fmt.Sprintf("%s/bookings/%s", h.bookingURL, id))
	if apiErr.Code != "" {
		writeError(w, apiErr)
		return
	}
	paymentResp, apiErr := h.fetch(fmt.Sprintf("%s/payments/%s", h.paymentURL, id))
	if apiErr.Code != "" {
		writeError(w, apiErr)
		return
	}

	var booking dto.BookingResponse
	var payment dto.PaymentResponse
	_ = json.Unmarshal(bookingResp, &booking)
	_ = json.Unmarshal(paymentResp, &payment)

	writeJSON(w, http.StatusOK, dto.BookingAggregateResponse{Booking: booking, Payment: payment})
}

func (h *Handler) fetch(url string) ([]byte, pkgErrors.APIError) {
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, pkgErrors.New("bad_gateway", err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgErrors.New("bad_gateway", err.Error())
	}
	if resp.StatusCode >= 400 {
		return nil, pkgErrors.APIError{Code: "upstream_error", Message: string(body)}
	}
	return body, pkgErrors.APIError{}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	if err.Code == "" {
		return
	}
	writeJSON(w, pkgErrors.StatusCode(err), err)
}
