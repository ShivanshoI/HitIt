package organizations

// DetailsRequest payload expected for POST /api/organizations/details
type DetailsRequest struct {
	OrgID string `json:"orgId"`
}

// VerifyRequest payload expected for POST /api/organizations/verify
type VerifyRequest struct {
	OrgID string `json:"orgId"`
}

// OrganizationResponse is the unified wrapper for org-related GETs.
type OrganizationResponse struct {
	Success bool                `json:"success"`
	Org     OrganizationDetails `json:"org"`
}

// OrganizationDetails sets out the response structure for /api/organizations/details
type OrganizationDetails struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"member_count"`
	Role        string `json:"role"`
	IsVerified  bool   `json:"is_verified"`
}

// VerifyResponse sets out the response structure for /api/organizations/verify
type VerifyResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Org     VerifiedOrgDetails `json:"org"`
}

// VerifiedOrgDetails details for the verify response
type VerifiedOrgDetails struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsVerified bool   `json:"is_verified"`
}
