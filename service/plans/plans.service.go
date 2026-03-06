package plans

import (
	"context"
	"fmt"
	"strings"
	"time"

	pogBillingDB "pog/database/billing"
	"pog/internal"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ── Catalogue (source of truth for plan definitions) ─────────────────
// In the future this can be seeded into a MongoDB "plans" collection.
// For now we define them in-memory and serve them as-is.

var planCatalogue = []Plan{
	{
		ID:          "free",
		Name:        "Starter",
		Description: "Perfect for individual developers and small personal projects.",
		Color:       "#10b981",
		IsPopular:   false,
		Price:       PlanPrice{Monthly: 0, Yearly: 0},
		Features: []string{
			"Up to 3 personal collections",
			"Basic request execution",
			"Local environment variables",
			"7-day execution history",
			"Community support",
		},
	},
	{
		ID:          "pro",
		Name:        "Professional",
		Description: "Advanced features and collaboration for power users and teams.",
		Color:       "#6c3fc5",
		IsPopular:   true,
		Price:       PlanPrice{Monthly: 12, Yearly: 9},
		Features: []string{
			"Unlimited collections",
			"Unlimited team members",
			"Advanced request history (30 days)",
			"Cloud environment sync",
			"Priority email support",
		},
	},
	{
		ID:          "enterprise",
		Name:        "Enterprise",
		Description: "Custom security, compliance, and administration for large orgs.",
		Color:       "#3b82f6",
		IsPopular:   false,
		Price:       PlanPrice{Monthly: 49, Yearly: 39},
		Features: []string{
			"Everything in Professional",
			"SAML Single Sign-On (SSO)",
			"Role-Based Access Control",
			"Unlimited execution history logs",
			"24/7 dedicated support",
		},
	},
}

// validPlanIDs is a quick look-up set for validation.
var validPlanIDs = func() map[string]bool {
	m := make(map[string]bool, len(planCatalogue))
	for _, p := range planCatalogue {
		m[p.ID] = true
	}
	return m
}()

func getPlanByID(id string) (*Plan, bool) {
	for i, p := range planCatalogue {
		if p.ID == id {
			return &planCatalogue[i], true
		}
	}
	return nil, false
}

// ── PlansService ──────────────────────────────────────────────────────

type PlansService struct {
	billingRepo *pogBillingDB.BillingRepository
}

func NewPlansService(billingRepo *pogBillingDB.BillingRepository) *PlansService {
	return &PlansService{billingRepo: billingRepo}
}

// ListPlans returns all plans from the in-memory catalogue.
func (s *PlansService) ListPlans(_ context.Context) PlansResponse {
	return PlansResponse{Plans: planCatalogue}
}

// GetSubscription returns a user's subscription. If no record exists, returns a
// synthetic "free" subscription.
func (s *PlansService) GetSubscription(ctx context.Context, userID string) (*SubscriptionResponse, error) {
	sub, err := s.billingRepo.GetSubscription(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to fetch subscription")
	}

	if sub == nil {
		// Default — no paid subscription
		return &SubscriptionResponse{
			PlanID:            "free",
			PlanName:          "Starter",
			Status:            "none",
			BillingCycle:      "monthly",
			CurrentPeriodEnd:  "",
			CancelAtPeriodEnd: false,
			Seats:             1,
		}, nil
	}

	return &SubscriptionResponse{
		PlanID:            sub.PlanID,
		PlanName:          sub.PlanName,
		Status:            sub.Status,
		BillingCycle:      sub.BillingCycle,
		CurrentPeriodEnd:  sub.CurrentPeriodEnd.UTC().Format(time.RFC3339),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		Seats:             sub.Seats,
	}, nil
}

// UpgradePlan changes the user's subscription plan.
// For free → paid: returns a Stripe checkout redirect URL (stubbed).
// For paid → paid: updates directly.
// NOTE: Stripe integration adds real payment gates. Currently we auto-upgrade
// if a data payment method is on file, or return a stub checkoutUrl for demo.
func (s *PlansService) UpgradePlan(ctx context.Context, userID string, req UpgradeRequest) (any, error) {
	if !validPlanIDs[req.PlanID] {
		return nil, internal.NewBadRequest("invalid planId")
	}
	if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
		return nil, internal.NewBadRequest("billingCycle must be 'monthly' or 'yearly'")
	}

	plan, _ := getPlanByID(req.PlanID)
	
	// Check whether the user has a payment method on file
	pm, err := s.billingRepo.GetPaymentMethod(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to check payment method")
	}

	// Paid plan with no payment method → redirect to (stub) checkout
	if plan.Price.Monthly > 0 && pm == nil {
		// TODO: integrate with Stripe to create a real checkout session
		return &UpgradeRedirectResponse{
			Success:         false,
			RequiresPayment: true,
			CheckoutURL:     fmt.Sprintf("https://checkout.stripe.com/pay/cs_stub_%s_%s", req.PlanID, req.BillingCycle),
		}, nil
	}

	// Direct upgrade
	objUserID, _ := primitive.ObjectIDFromHex(userID)
	periodEnd := time.Now().AddDate(0, 1, 0) // 1 month from today by default
	if req.BillingCycle == "yearly" {
		periodEnd = time.Now().AddDate(1, 0, 0)
	}

	sub := &pogBillingDB.Subscription{
		UserID:            objUserID,
		PlanID:            plan.ID,
		PlanName:          plan.Name,
		Status:            "active",
		BillingCycle:      req.BillingCycle,
		CurrentPeriodEnd:  periodEnd,
		CancelAtPeriodEnd: false,
		Seats:             1,
	}

	saved, err := s.billingRepo.UpsertSubscription(ctx, sub)
	if err != nil {
		return nil, internal.NewInternalError("failed to update subscription")
	}

	return &UpgradeResponse{
		Success: true,
		Message: fmt.Sprintf("Plan upgraded to %s successfully.", plan.Name),
		Subscription: SubscriptionResponse{
			PlanID:            saved.PlanID,
			PlanName:          saved.PlanName,
			Status:            saved.Status,
			BillingCycle:      saved.BillingCycle,
			CurrentPeriodEnd:  saved.CurrentPeriodEnd.UTC().Format(time.RFC3339),
			CancelAtPeriodEnd: saved.CancelAtPeriodEnd,
			Seats:             saved.Seats,
		},
	}, nil
}

// CancelSubscription marks the subscription to cancel at the end of the period.
func (s *PlansService) CancelSubscription(ctx context.Context, userID string, req CancelRequest) (*CancelResponse, error) {
	sub, err := s.billingRepo.GetSubscription(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to fetch subscription")
	}
	if sub == nil || sub.Status != "active" {
		return nil, internal.NewBadRequest("no active subscription to cancel")
	}

	sub.CancelAtPeriodEnd = true
	saved, err := s.billingRepo.UpsertSubscription(ctx, sub)
	if err != nil {
		return nil, internal.NewInternalError("failed to cancel subscription")
	}

	periodStr := saved.CurrentPeriodEnd.UTC().Format("Jan 02, 2006")

	return &CancelResponse{
		Success: true,
		Message: fmt.Sprintf(
			"Your subscription has been cancelled. You will retain access to %s features until %s.",
			saved.PlanName, periodStr,
		),
		Subscription: SubscriptionResponse{
			PlanID:            saved.PlanID,
			PlanName:          saved.PlanName,
			Status:            saved.Status,
			BillingCycle:      saved.BillingCycle,
			CurrentPeriodEnd:  saved.CurrentPeriodEnd.UTC().Format(time.RFC3339),
			CancelAtPeriodEnd: saved.CancelAtPeriodEnd,
			Seats:             saved.Seats,
		},
	}, nil
}

// ── Billing methods ───────────────────────────────────────────────────

// GetPaymentMethod returns the user's payment method, or nil if none.
func (s *PlansService) GetPaymentMethod(ctx context.Context, userID string) (*PaymentMethodResponse, error) {
	pm, err := s.billingRepo.GetPaymentMethod(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to fetch payment method")
	}
	if pm == nil {
		return nil, nil
	}
	return &PaymentMethodResponse{
		ID:        pm.ProviderID,
		Brand:     pm.Brand,
		Last4:     pm.Last4,
		ExpMonth:  pm.ExpMonth,
		ExpYear:   pm.ExpYear,
		IsDefault: pm.IsDefault,
	}, nil
}

// UpdatePaymentMethod stores a new payment method from a Stripe token.
// In production: call Stripe to retrieve card details from the pm_ token before storing.
// Here we parse the minimal info we can and store a placeholder.
// NOTE: Real Stripe integration should be done server-to-server before this write.
func (s *PlansService) UpdatePaymentMethod(ctx context.Context, userID string, req UpdatePaymentMethodRequest) (*UpdatePaymentMethodResponse, error) {
	if strings.TrimSpace(req.PaymentMethodID) == "" {
		return nil, internal.NewBadRequest("paymentMethodId is required")
	}

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, internal.NewUnauthorized("invalid user")
	}

	// TODO: Call Stripe API to retrieve card details (brand, last4, exp) from pm_xxx
	// For now we store a placeholder — replace with real Stripe call.
	pm := &pogBillingDB.PaymentMethod{
		UserID:     objUserID,
		ProviderID: req.PaymentMethodID,
		Brand:      "card", // Placeholder — use Stripe API response
		Last4:      "****", // Placeholder
		ExpMonth:   0,
		ExpYear:    0,
		IsDefault:  true,
	}

	saved, err := s.billingRepo.UpsertPaymentMethod(ctx, pm)
	if err != nil {
		return nil, internal.NewInternalError("failed to save payment method")
	}

	return &UpdatePaymentMethodResponse{
		Success: true,
		Message: "Payment method updated successfully.",
		PaymentMethod: PaymentMethodResponse{
			ID:        saved.ProviderID,
			Brand:     saved.Brand,
			Last4:     saved.Last4,
			ExpMonth:  saved.ExpMonth,
			ExpYear:   saved.ExpYear,
			IsDefault: saved.IsDefault,
		},
	}, nil
}

// GetBillingInfo returns the billing address for a user.
func (s *PlansService) GetBillingInfo(ctx context.Context, userID string) (*BillingInfoResponse, error) {
	info, err := s.billingRepo.GetBillingInfo(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to fetch billing info")
	}
	if info == nil {
		return nil, nil
	}
	return billingInfoToResponse(info), nil
}

// UpdateBillingInfo saves or replaces the user's billing address.
func (s *PlansService) UpdateBillingInfo(ctx context.Context, userID string, req UpdateBillingInfoRequest) (*UpdateBillingInfoResponse, error) {
	if strings.TrimSpace(req.AddressLine1) == "" || strings.TrimSpace(req.City) == "" {
		return nil, internal.NewBadRequest("addressLine1 and city are required")
	}

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, internal.NewUnauthorized("invalid user")
	}

	info := &pogBillingDB.BillingInfo{
		UserID:       objUserID,
		CompanyName:  req.CompanyName,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		TaxID:        req.TaxID,
	}

	saved, err := s.billingRepo.UpsertBillingInfo(ctx, info)
	if err != nil {
		return nil, internal.NewInternalError("failed to save billing info")
	}

	return &UpdateBillingInfoResponse{
		Success:     true,
		Message:     "Billing information updated.",
		BillingInfo: *billingInfoToResponse(saved),
	}, nil
}

// ListInvoices returns a paginated list of invoices for a user.
func (s *PlansService) ListInvoices(ctx context.Context, userID string, page, limit int) (*InvoiceListResponse, error) {
	invoices, total, err := s.billingRepo.ListInvoices(ctx, userID, page, limit)
	if err != nil {
		return nil, internal.NewInternalError("failed to fetch invoices")
	}

	items := make([]InvoiceResponse, 0, len(invoices))
	for _, inv := range invoices {
		item := InvoiceResponse{
			ID:       inv.InvoiceID,
			Date:     inv.Date.UTC().Format(time.RFC3339),
			Amount:   inv.Amount,
			Currency: inv.Currency,
			Status:   inv.Status,
			PdfURL:   inv.PdfURL,
		}
		items = append(items, item)
	}

	return &InvoiceListResponse{
		Invoices: items,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

// GetInvoiceForDownload looks up an invoice and validates it can be downloaded.
func (s *PlansService) GetInvoiceForDownload(ctx context.Context, userID, invoiceID string) (*pogBillingDB.Invoice, error) {
	inv, err := s.billingRepo.GetInvoiceByInvoiceID(ctx, userID, invoiceID)
	if err != nil {
		return nil, internal.NewInternalError("failed to fetch invoice")
	}
	if inv == nil {
		return nil, internal.NewNotFound("invoice not found")
	}
	if inv.Status == "upcoming" {
		return nil, internal.NewBadRequest("invoice PDF not yet generated for upcoming invoices")
	}
	return inv, nil
}

// ── helpers ───────────────────────────────────────────────────────────

func billingInfoToResponse(info *pogBillingDB.BillingInfo) *BillingInfoResponse {
	return &BillingInfoResponse{
		CompanyName:  info.CompanyName,
		AddressLine1: info.AddressLine1,
		AddressLine2: info.AddressLine2,
		City:         info.City,
		State:        info.State,
		PostalCode:   info.PostalCode,
		Country:      info.Country,
		TaxID:        info.TaxID,
	}
}
