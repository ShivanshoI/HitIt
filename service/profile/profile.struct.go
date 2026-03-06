package profile

// ── Request structs ──────────────────────────────────────────────────

// UpdateProfileRequest is the payload for PUT /api/user/profile
type UpdateProfileRequest struct {
	Name  string `json:"name"`  // Full display name (required)
	Email string `json:"email"` // Email address (required)
	Theme string `json:"theme"` // "system" | "dark" | "light" (optional)
}

// UpdatePasswordRequest is the payload for PUT /api/user/password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"` // Must match stored hash
	NewPassword     string `json:"newPassword"`     // Min 8 chars
}

// ── Response structs ─────────────────────────────────────────────────

// ProfileStatsResponse is the response shape for GET /api/user/me/stats
type ProfileStatsResponse struct {
	TotalCollections int `json:"totalCollections"`
	RequestsSent     int `json:"requestsSent"`
	TeamsJoined      int `json:"teamsJoined"`
}

// ActivityEventResponse is a single item in the activity feed.
type ActivityEventResponse struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Time  string `json:"time"` // ISO 8601
	Icon  string `json:"icon"`
}

// ActivityFeedResponse wraps the activity list.
type ActivityFeedResponse struct {
	Activity []ActivityEventResponse `json:"activity"`
}

// UserProfileResponse is the trimmed user object returned after a profile update.
type UserProfileResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Theme string `json:"theme"`
}

// UpdateProfileResponse wraps the updated user.
type UpdateProfileResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	User    UserProfileResponse `json:"user"`
}

// UpdatePasswordResponse is the success body for PUT /api/user/password.
type UpdatePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RevokeSessionsResponse is the success body for DELETE /api/user/sessions.
type RevokeSessionsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
