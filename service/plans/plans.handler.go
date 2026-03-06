package plans

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pog/internal"
	"pog/middleware"
)

// PlansHandler wires the plans/subscription/billing routes.
type PlansHandler struct {
	service *PlansService
}

func NewPlansHandler(service *PlansService) *PlansHandler {
	return &PlansHandler{service: service}
}

// RegisterRoutes attaches all plans, subscription, and billing routes.
func (h *PlansHandler) RegisterRoutes(mux *http.ServeMux) {
	// ── Plan catalogue ────────────────────────────────────────────────
	mux.Handle("GET "+internal.APIPrefix+"/plans",
		middleware.Auth(http.HandlerFunc(h.ListPlans)))

	// ── Subscription ─────────────────────────────────────────────────
	mux.Handle("GET "+internal.APIPrefix+"/user/subscription",
		middleware.Auth(http.HandlerFunc(h.GetSubscription)))

	mux.Handle("POST "+internal.APIPrefix+"/subscription/upgrade",
		middleware.Auth(http.HandlerFunc(h.UpgradePlan)))

	mux.Handle("POST "+internal.APIPrefix+"/subscription/cancel",
		middleware.Auth(http.HandlerFunc(h.CancelSubscription)))

	// ── Billing — payment method ──────────────────────────────────────
	mux.Handle("GET "+internal.APIPrefix+"/billing/payment-method",
		middleware.Auth(http.HandlerFunc(h.GetPaymentMethod)))

	mux.Handle("PUT "+internal.APIPrefix+"/billing/payment-method",
		middleware.Auth(http.HandlerFunc(h.UpdatePaymentMethod)))

	// ── Billing — billing info ────────────────────────────────────────
	mux.Handle("GET "+internal.APIPrefix+"/billing/info",
		middleware.Auth(http.HandlerFunc(h.GetBillingInfo)))

	mux.Handle("PUT "+internal.APIPrefix+"/billing/info",
		middleware.Auth(http.HandlerFunc(h.UpdateBillingInfo)))

	// ── Billing — invoices ────────────────────────────────────────────
	mux.Handle("GET "+internal.APIPrefix+"/billing/invoices",
		middleware.Auth(http.HandlerFunc(h.ListInvoices)))

	mux.Handle("GET "+internal.APIPrefix+"/billing/invoices/{id}/download",
		middleware.Auth(http.HandlerFunc(h.DownloadInvoice)))
}

// ── Handlers ─────────────────────────────────────────────────────────

// ListPlans handles GET /api/plans
func (h *PlansHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	result := h.service.ListPlans(r.Context())
	internal.SuccessResponse(w, http.StatusOK, result)
}

// GetSubscription handles GET /api/user/subscription
func (h *PlansHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	sub, err := h.service.GetSubscription(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to fetch subscription"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, map[string]any{"subscription": sub})
}

// UpgradePlan handles POST /api/subscription/upgrade
func (h *PlansHandler) UpgradePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var req UpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	result, err := h.service.UpgradePlan(r.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("upgrade failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, result)
}

// CancelSubscription handles POST /api/subscription/cancel
func (h *PlansHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var req CancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	resp, err := h.service.CancelSubscription(r.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("cancellation failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, resp)
}

// GetPaymentMethod handles GET /api/billing/payment-method
func (h *PlansHandler) GetPaymentMethod(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	pm, err := h.service.GetPaymentMethod(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to fetch payment method"))
		}
		return
	}

	// pm can be nil when no payment method is on file
	internal.SuccessResponse(w, http.StatusOK, map[string]any{"paymentMethod": pm})
}

// UpdatePaymentMethod handles PUT /api/billing/payment-method
func (h *PlansHandler) UpdatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var req UpdatePaymentMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	resp, err := h.service.UpdatePaymentMethod(r.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to update payment method"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, resp)
}

// GetBillingInfo handles GET /api/billing/info
func (h *PlansHandler) GetBillingInfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	info, err := h.service.GetBillingInfo(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to fetch billing info"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, map[string]any{"billingInfo": info})
}

// UpdateBillingInfo handles PUT /api/billing/info
func (h *PlansHandler) UpdateBillingInfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var req UpdateBillingInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	resp, err := h.service.UpdateBillingInfo(r.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to update billing info"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, resp)
}

// ListInvoices handles GET /api/billing/invoices?page=1&limit=20
func (h *PlansHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 50 {
		limit = 50
	}

	result, err := h.service.ListInvoices(r.Context(), userID, page, limit)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to fetch invoices"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, result)
}

// DownloadInvoice handles GET /api/billing/invoices/{id}/download
// Redirects to the PDF URL stored on the invoice document.
// NOTE: We do NOT generate PDFs server-side. The pdfUrl in the invoice record
// must point to a pre-generated PDF (e.g. from Stripe or stored in S3).
// If you need server-side PDF generation, add a library like `go-pdf` or use
// Stripe's built-in invoice PDF and store that URL.
func (h *PlansHandler) DownloadInvoice(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	// Extract {id} from URL path — Go 1.22+ pattern variable extraction
	invoiceID := strings.TrimPrefix(r.PathValue("id"), "")
	if invoiceID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("invoice id is required"))
		return
	}

	inv, err := h.service.GetInvoiceForDownload(r.Context(), userID, invoiceID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to fetch invoice"))
		}
		return
	}

	if inv.PdfURL == nil || *inv.PdfURL == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("no PDF available for this invoice"))
		return
	}

	// Redirect the client to the PDF URL (pre-signed S3 / Stripe hosted PDF)
	http.Redirect(w, r, *inv.PdfURL, http.StatusTemporaryRedirect)
}
