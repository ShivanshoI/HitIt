package users

// SignInRequest represents the payload expected for the POST /auth/sign-in endpoint.
// It maps the JSON body containing 'identifier' (such as email or phone) and 'password'.
type SignInRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// SignInResponse represents the successful response payload for the sign-in endpoint.
type SignInResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse is a dto for the User data safe for returning to client.
type UserResponse struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	NickName     string `json:"nick_name"`
	PhoneNumber  string `json:"phone_number"`
	EmailAddress   string  `json:"email_address"`
	OrganizationID *string `json:"organizationId,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}
