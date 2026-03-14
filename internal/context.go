package internal

import "context"

// Scope represents the unified authentication and scoping context.
type Scope struct {
	UserID string
	TeamID string
	OrgID  string
}

// MustUserID extracts the authenticated user ID from the context.
func MustUserID(ctx context.Context) string {
	v := ctx.Value(UserIDKey)
	if v == nil {
		panic("MustUserID: no user ID in context")
	}
	return v.(string)
}

// GetScope extracts the unified Scope from context.
func GetScope(ctx context.Context) Scope {
	s, ok := ctx.Value(ScopeKey).(Scope)
	if !ok {
		// Fallback to individual keys if unified scope is missing
		return Scope{
			UserID: getString(ctx, UserIDKey),
			TeamID: getString(ctx, TeamIDKey),
			OrgID:  getString(ctx, OrgIDKey),
		}
	}
	return s
}

func getString(ctx context.Context, key ContextKey) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}