package teams

// ── Request DTOs ────────────────────────────────────────────────────

type CreateTeamDTO struct {
	Name        string `json:"name"`
	Theme       string `json:"theme"`
	Description string `json:"description"`
}

func (d *CreateTeamDTO) IsValid() (string, bool) {
	if len(d.Name) < 3 || len(d.Name) > 50 {
		return "name must be 3-50 characters", false
	}
	validThemes := map[string]bool{
		"purple": true, "emerald": true, "blue": true,
		"orange": true, "rose": true, "cyan": true,
	}
	if !validThemes[d.Theme] {
		return "theme must be one of: purple, emerald, blue, orange, rose, cyan", false
	}
	if len(d.Description) > 300 {
		return "description cannot exceed 300 characters", false
	}
	return "", true
}

type UpdateTeamDTO struct {
	Name        *string `json:"name,omitempty"`
	Theme       *string `json:"theme,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ChangeRoleDTO struct {
	Role string `json:"role"`
}

type InviteDTO struct {
	Emails []string `json:"emails"`
}

type SendFeedDTO struct {
	Type     string   `json:"type"`
	Message  string   `json:"message"`
	Title    *string  `json:"title"`
	Mentions []string `json:"mentions"`
}

func (d *SendFeedDTO) IsValid() (string, bool) {
	if d.Type != "message" && d.Type != "issue" {
		return "type must be 'message' or 'issue'", false
	}
	if len(d.Message) == 0 || len(d.Message) > 2000 {
		return "message must be 1-2000 characters", false
	}
	if d.Type == "issue" && (d.Title == nil || len(*d.Title) == 0 || len(*d.Title) > 200) {
		return "issue title is required and must be 1-200 characters", false
	}
	return "", true
}

// ── Response DTOs ───────────────────────────────────────────────────

type TeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Theme       string `json:"theme"`
	Description string `json:"description"`
	Role        string `json:"role"`
	MemberCount int64  `json:"member_count"`
	InviteLink  string `json:"invite_link"`
	CreatedAt   string `json:"created_at"`
}

type MemberResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

type InviteResponse struct {
	Invited       int `json:"invited"`
	AlreadyMember int `json:"already_member"`
}

type InviteLinkResponse struct {
	Link  string `json:"link"`
	Token string `json:"token"`
}

type FeedItemResponse struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	UserID     string   `json:"user_id"`
	UserName   string   `json:"user_name"`
	Message    string   `json:"message"`
	Title      string   `json:"title,omitempty"`
	Mentions   []string `json:"mentions,omitempty"`
	Resolved   *bool    `json:"resolved,omitempty"`
	ResolvedBy *string  `json:"resolved_by,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

type PaginatedFeedResponse struct {
	Data       []FeedItemResponse `json:"data"`
	Pagination PaginationMeta     `json:"pagination"`
}

type PaginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
