package plans

// ── Plan structs ─────────────────────────────────────────────────────

// PlanPrice holds the monthly and yearly per-seat prices.
type PlanPrice struct {
	Monthly float64 `json:"monthly"`
	Yearly  float64 `json:"yearly"`
}

// Plan represents a subscription tier returned by GET /api/plans.
type Plan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	IsPopular   bool      `json:"isPopular"`
	Price       PlanPrice `json:"price"`
	Features    []string  `json:"features"`
}

// PlansResponse wraps the plans list.
type PlansResponse struct {
	Plans []Plan `json:"plans"`
}

// ── Subscription structs ─────────────────────────────────────────────

// SubscriptionResponse is the shape returned by GET /api/user/subscription.
type SubscriptionResponse struct {
	PlanID            string `json:"planId"`
	PlanName          string `json:"planName"`
	Status            string `json:"status"`
	BillingCycle      string `json:"billingCycle"`
	CurrentPeriodEnd  string `json:"currentPeriodEnd"`  // ISO 8601
	CancelAtPeriodEnd bool   `json:"cancelAtPeriodEnd"`
	Seats             int    `json:"seats"`
}

// UpgradeRequest is the payload for POST /api/subscription/upgrade.
type UpgradeRequest struct {
	PlanID       string `json:"planId"`
	BillingCycle string `json:"billingCycle"`
}

// UpgradeResponse is returned after a successful upgrade where no payment is needed.
type UpgradeResponse struct {
	Success      bool                 `json:"success"`
	Message      string               `json:"message"`
	Subscription SubscriptionResponse `json:"subscription"`
}

// UpgradeRedirectResponse is returned when payment is required (Stripe checkout).
type UpgradeRedirectResponse struct {
	Success         bool   `json:"success"`
	RequiresPayment bool   `json:"requiresPayment"`
	CheckoutURL     string `json:"checkoutUrl"`
}

// CancelRequest is the payload for POST /api/subscription/cancel.
type CancelRequest struct {
	Reason   string `json:"reason"`   // Optional enum
	Feedback string `json:"feedback"` // Optional free text
}

// CancelResponse is returned after a successful cancellation.
type CancelResponse struct {
	Success      bool                 `json:"success"`
	Message      string               `json:"message"`
	Subscription SubscriptionResponse `json:"subscription"`
}

// ── Billing structs ──────────────────────────────────────────────────

// PaymentMethodResponse is the client-facing payment method view.
type PaymentMethodResponse struct {
	ID        string `json:"id"`
	Brand     string `json:"brand"`
	Last4     string `json:"last4"`
	ExpMonth  int    `json:"expMonth"`
	ExpYear   int    `json:"expYear"`
	IsDefault bool   `json:"isDefault"`
}

// UpdatePaymentMethodRequest carries the Stripe tokenised payment method ID.
type UpdatePaymentMethodRequest struct {
	PaymentMethodID string `json:"paymentMethodId"`
}

// UpdatePaymentMethodResponse wraps the result of a payment method update.
type UpdatePaymentMethodResponse struct {
	Success       bool                  `json:"success"`
	Message       string                `json:"message"`
	PaymentMethod PaymentMethodResponse  `json:"paymentMethod"`
}

// BillingInfoResponse is the client-facing billing address view.
type BillingInfoResponse struct {
	CompanyName  *string `json:"companyName"`
	AddressLine1 string  `json:"addressLine1"`
	AddressLine2 *string `json:"addressLine2"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	PostalCode   string  `json:"postalCode"`
	Country      string  `json:"country"`
	TaxID        *string `json:"taxId"`
}

// UpdateBillingInfoRequest is the payload for PUT /api/billing/info.
type UpdateBillingInfoRequest struct {
	CompanyName  *string `json:"companyName"`
	AddressLine1 string  `json:"addressLine1"`
	AddressLine2 *string `json:"addressLine2"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	PostalCode   string  `json:"postalCode"`
	Country      string  `json:"country"`
	TaxID        *string `json:"taxId"`
}

// UpdateBillingInfoResponse wraps the updated billing info.
type UpdateBillingInfoResponse struct {
	Success     bool                `json:"success"`
	Message     string              `json:"message"`
	BillingInfo BillingInfoResponse `json:"billingInfo"`
}

// InvoiceResponse is a single invoice item.
type InvoiceResponse struct {
	ID       string  `json:"id"`
	Date     string  `json:"date"` // ISO 8601
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
	PdfURL   *string `json:"pdfUrl"`
}

// InvoiceListResponse wraps the paginated invoice list.
type InvoiceListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}
