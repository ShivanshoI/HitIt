package collections

type CollectionResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	UserID         string   `json:"user_id"`
	Tags           []string `json:"tags"`
	Default_Method string   `json:"default_method"`
	Accent_Color   string   `json:"accent_color"`
	Pattern        string   `json:"pattern"`
	TotalRequests  int      `json:"total_requests"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

type PaginatedCollectionResponse struct {
	Collections []CollectionResponse `json:"collections"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
	TotalPages  int64                `json:"total_pages"`
}

type CreateCollectionDTO struct {
	UserID         string   `json:"user_id"`
	Name           string   `json:"name"`
	Tags           []string `json:"tags"`
	Default_Method string   `json:"default_method"`
	Accent_Color   string   `json:"accent_color"`
	Pattern        string   `json:"pattern"`
}

func (d *CreateCollectionDTO) IsValid() bool {
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"OPTIONS": true,
		"HEAD":    true,
	}
	return validMethods[d.Default_Method]
}
