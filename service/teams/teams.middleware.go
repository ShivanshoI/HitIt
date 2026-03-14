package teams

import (
	"context"
	"log"
	"net/http"

	teamsDB "pog/database/teams"
	"pog/database/teams_mapping"
	"pog/internal"
	"pog/middleware"

	"go.mongodb.org/mongo-driver/mongo"
)

// AuthWithTeam returns a middleware that validates the X-Team-Id header.
// It ensures the user is authenticated (applying middleware.Auth if needed).
func AuthWithTeam(db *mongo.Database) func(http.Handler) http.Handler {
	mappingRepo := teams_mapping.NewTeamsMappingRepository(db)
	teamRepo := teamsDB.NewTeamsRepository(db)

	return func(next http.Handler) http.Handler {
		logic := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// Organization Mode Check: If OrgID is in context, the team must belong to it
			orgID, _ := r.Context().Value(internal.OrgIDKey).(string)

			isMember, err := mappingRepo.IsMember(r.Context(), teamIDHeader, userID)
			if err != nil || !isMember {
				log.Printf("[TEAM-MW] user %s is not a member of team %s", userID, teamIDHeader)
				internal.ErrorResponse(w, internal.NewForbidden("not a member of this team"))
				return
			}

			if orgID != "" {
				team, err := teamRepo.FindID(r.Context(), teamIDHeader)
				if err != nil || team == nil {
					internal.ErrorResponse(w, internal.NewNotFound("team not found"))
					return
				}
				if team.OrganizationID == nil || team.OrganizationID.Hex() != orgID {
					log.Printf("[TEAM-MW] team %s does not belong to organization %s", teamIDHeader, orgID)
					internal.ErrorResponse(w, internal.NewForbidden("team does not belong to this organization"))
					return
				}
			}

			// Set team_id in context
			ctx := context.WithValue(r.Context(), internal.TeamIDKey, teamIDHeader)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(internal.UserIDKey) != nil {
				logic.ServeHTTP(w, r)
			} else {
				middleware.Auth(logic).ServeHTTP(w, r)
			}
		})
	}
}
