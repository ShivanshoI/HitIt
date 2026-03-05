package internal

import "context"

// MustUserID extracts the authenticated user ID from the context.
// It panics if the value is missing, which should never happen on
// an Auth-middleware-protected route — hence "Must".
func MustUserID(ctx context.Context) string {
	v := ctx.Value(UserIDKey)
	if v == nil {
		panic("MustUserID: no user ID in context — is the Auth middleware applied?")
	}
	return v.(string)
}