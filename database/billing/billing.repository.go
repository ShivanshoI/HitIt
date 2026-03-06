package billing

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ── Collection name constants ────────────────────────────────────────

const (
	PaymentMethodCollection = "payment_methods"
	BillingInfoCollection   = "billing_info"
	InvoiceCollection       = "invoices"
	SubscriptionCollection  = "subscriptions"
)

// ── Repository structs ───────────────────────────────────────────────

type BillingRepository struct {
	paymentMethods *mongo.Collection
	billingInfo    *mongo.Collection
	invoices       *mongo.Collection
	subscriptions  *mongo.Collection
}

func NewBillingRepository(db *mongo.Database) *BillingRepository {
	return &BillingRepository{
		paymentMethods: db.Collection(PaymentMethodCollection),
		billingInfo:    db.Collection(BillingInfoCollection),
		invoices:       db.Collection(InvoiceCollection),
		subscriptions:  db.Collection(SubscriptionCollection),
	}
}

// ── Payment Method ───────────────────────────────────────────────────

// GetPaymentMethod returns the default payment method for a user.
// Returns (nil, nil) when no payment method is on file.
func (r *BillingRepository) GetPaymentMethod(ctx context.Context, userID string) (*PaymentMethod, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var pm PaymentMethod
	err = r.paymentMethods.FindOne(ctx, bson.M{
		"user_id":    objUserID,
		"is_default": true,
	}).Decode(&pm)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &pm, nil
}

// UpsertPaymentMethod stores or replaces a user's default payment method.
// We store only metadata — the actual Stripe provider_id and card details already tokenized.
func (r *BillingRepository) UpsertPaymentMethod(ctx context.Context, pm *PaymentMethod) (*PaymentMethod, error) {
	// Mark all existing methods for this user as non-default
	objUserID := pm.UserID
	_, err := r.paymentMethods.UpdateMany(ctx,
		bson.M{"user_id": objUserID},
		bson.M{"$set": bson.M{"is_default": false}},
	)
	if err != nil {
		return nil, err
	}

	pm.IsDefault = true
	pm.UpdatedAt = time.Now()

	// Try to upsert based on provider_id
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": objUserID, "provider_id": pm.ProviderID}
	update := bson.M{"$set": pm}
	if pm.ID.IsZero() {
		pm.ID = primitive.NewObjectID()
		pm.CreatedAt = time.Now()
	}

	_, err = r.paymentMethods.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	return pm, nil
}

// ── Billing Info ─────────────────────────────────────────────────────

// GetBillingInfo returns the billing address for a user.
// Returns (nil, nil) when none exists.
func (r *BillingRepository) GetBillingInfo(ctx context.Context, userID string) (*BillingInfo, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var info BillingInfo
	err = r.billingInfo.FindOne(ctx, bson.M{"user_id": objUserID}).Decode(&info)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

// UpsertBillingInfo creates or replaces a user's billing info.
func (r *BillingRepository) UpsertBillingInfo(ctx context.Context, info *BillingInfo) (*BillingInfo, error) {
	info.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": info.UserID}
	update := bson.M{"$set": info}

	_, err := r.billingInfo.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// ── Invoices ─────────────────────────────────────────────────────────

// ListInvoices returns paginated invoices for a user, sorted newest first.
func (r *BillingRepository) ListInvoices(ctx context.Context, userID string, page, limit int) ([]Invoice, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"user_id": objUserID}

	total, err := r.invoices.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.M{"date": -1}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.invoices.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var invoices []Invoice
	if err = cursor.All(ctx, &invoices); err != nil {
		return nil, 0, err
	}
	if invoices == nil {
		invoices = []Invoice{}
	}
	return invoices, total, nil
}

// GetInvoiceByInvoiceID retrieves a single invoice by its human-readable ID (e.g. "INV-2026-003")
// for a specific user (prevents cross-user access).
func (r *BillingRepository) GetInvoiceByInvoiceID(ctx context.Context, userID, invoiceID string) (*Invoice, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	err = r.invoices.FindOne(ctx, bson.M{
		"user_id":    objUserID,
		"invoice_id": invoiceID,
	}).Decode(&invoice)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

// ── Subscriptions ─────────────────────────────────────────────────────

// GetSubscription returns the active subscription for a user.
// Returns (nil, nil) when no subscription record exists.
func (r *BillingRepository) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var sub Subscription
	err = r.subscriptions.FindOne(ctx, bson.M{"user_id": objUserID}).Decode(&sub)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// UpsertSubscription creates or updates a user's subscription record.
func (r *BillingRepository) UpsertSubscription(ctx context.Context, sub *Subscription) (*Subscription, error) {
	sub.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": sub.UserID}
	update := bson.M{"$set": sub}

	_, err := r.subscriptions.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
