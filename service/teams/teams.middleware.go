package teams

import (
	"context"
	"log"
	"net/http"

	"pog/database/teams_mapping"
	"pog/internal"
	"pog/middleware"

	"go.mongodb.org/mongo-driver/mongo"
)

// AuthWithTeam returns a middleware that wraps middleware.Auth and
// additionally validates the X-Team-Id header.
func AuthWithTeam(db *mongo.Database) func(http.Handler) http.Handler {
	mappingRepo := teams_mapping.NewTeamsMappingRepository(db)

	return func(next http.Handler) http.Handler {
		return middleware.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			teamIDHeader := r.Header.Get("X-Team-Id")
			if teamIDHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := r.Context().Value(internal.UserIDKey).(string)
			if !ok {
				internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
				return
			}

			isMember, err := mappingRepo.IsMember(r.Context(), teamIDHeader, userID)
			if err != nil || !isMember {
				log.Printf("[TEAM-MW] user %s is not a member of team %s", userID, teamIDHeader)
				internal.ErrorResponse(w, internal.NewForbidden("not a member of this team"))
				return
			}

			// Set team_id in context
			ctx := context.WithValue(r.Context(), internal.TeamIDKey, teamIDHeader)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}
