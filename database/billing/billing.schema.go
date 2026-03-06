package billing

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentMethod stores card details linked to a user (sourced from Stripe or similar).
// NOTE: We store only non-sensitive metadata (last4, brand, expiry).
// The actual tokenization is done client-side via Stripe.js.
type PaymentMethod struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ProviderID string             `bson:"provider_id" json:"provider_id"` // Stripe pm_xxx ID
	Brand     string             `bson:"brand" json:"brand"`             // "visa", "mastercard", etc.
	Last4     string             `bson:"last4" json:"last4"`
	ExpMonth  int                `bson:"exp_month" json:"expMonth"`
	ExpYear   int                `bson:"exp_year" json:"expYear"`
	IsDefault bool               `bson:"is_default" json:"isDefault"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// BillingInfo stores a user's billing address and tax details.
type BillingInfo struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	CompanyName  *string            `bson:"company_name,omitempty" json:"companyName,omitempty"`
	AddressLine1 string             `bson:"address_line1" json:"addressLine1"`
	AddressLine2 *string            `bson:"address_line2,omitempty" json:"addressLine2,omitempty"`
	City         string             `bson:"city" json:"city"`
	State        string             `bson:"state" json:"state"`
	PostalCode   string             `bson:"postal_code" json:"postalCode"`
	Country      string             `bson:"country" json:"country"` // ISO 3166-1 alpha-2
	TaxID        *string            `bson:"tax_id,omitempty" json:"taxId,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// Invoice represents a billing invoice for a subscription period.
type Invoice struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	InvoiceID string             `bson:"invoice_id" json:"invoiceId"` // Human-readable e.g. "INV-2026-003"
	Date     time.Time          `bson:"date" json:"date"`
	Amount   float64            `bson:"amount" json:"amount"`
	Currency string             `bson:"currency" json:"currency"` // e.g. "USD"
	// Status: "paid", "upcoming", "past_due", "void"
	Status  string  `bson:"status" json:"status"`
	PdfURL  *string `bson:"pdf_url,omitempty" json:"pdfUrl,omitempty"`
}

// Subscription stores the user's active subscription state.
type Subscription struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID `bson:"user_id" json:"user_id"`
	PlanID            string             `bson:"plan_id" json:"planId"`
	PlanName          string             `bson:"plan_name" json:"planName"`
	Status            string             `bson:"status" json:"status"`             // "active","trialing","past_due","cancelled","none"
	BillingCycle      string             `bson:"billing_cycle" json:"billingCycle"` // "monthly" | "yearly"
	CurrentPeriodEnd  time.Time          `bson:"current_period_end" json:"currentPeriodEnd"`
	CancelAtPeriodEnd bool               `bson:"cancel_at_period_end" json:"cancelAtPeriodEnd"`
	Seats             int                `bson:"seats" json:"seats"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}
