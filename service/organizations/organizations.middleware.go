package organizations

import (
	"context"
	"log"
	"net/http"

	"pog/database/users"
	"pog/internal"
	"pog/middleware"

	"go.mongodb.org/mongo-driver/mongo"
)

// AuthWithOrg returns a middleware that validates the X-Org-Id header.
// It ensures the user is authenticated (applying middleware.Auth if needed).
func AuthWithOrg(db *mongo.Database) func(http.Handler) http.Handler {
	userRepo := users.NewUserRepository(db)

	return func(next http.Handler) http.Handler {
		// The actual logic for organization validation
		logic := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgIDHeader := r.Header.Get("X-Org-Id")
			if orgIDHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := r.Context().Value(internal.UserIDKey).(string)
			if !ok {
				internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
				return
			}

			user, err := userRepo.GetByID(r.Context(), userID)
			if err != nil || user == nil {
				internal.ErrorResponse(w, internal.NewUnauthorized("user not found"))
				return
			}

			if user.OrganizationID == nil || user.OrganizationID.Hex() != orgIDHeader {
				log.Printf("[ORG-MW] user %s is not affiliated with organization %s", userID, orgIDHeader)
				internal.ErrorResponse(w, internal.NewForbidden("not affiliated with this organization"))
				return
			}

			// Set org_id in context
			ctx := context.WithValue(r.Context(), internal.OrgIDKey, orgIDHeader)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If already authenticated, just run the logic. Otherwise, wrap with Auth.
			if r.Context().Value(internal.UserIDKey) != nil {
				logic.ServeHTTP(w, r)
			} else {
				middleware.Auth(logic).ServeHTTP(w, r)
			}
		})
	}
}
