package requests

type KeyValuePair struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type RequestResponse struct {
	ID           string         `json:"id"`
	CollectionID string         `json:"collection_id"`
	Name         string         `json:"name"`
	Method       string         `json:"method"`
	URL          string         `json:"url"`
	Headers      []KeyValuePair `json:"headers"`
	Params       []KeyValuePair `json:"params"`
	Body         string         `json:"body"`
	Auth         string         `json:"auth"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
}

type CreateRequestDTO struct {
	CollectionID string         `json:"collection_id"`
	Name         string         `json:"name"`
	Method       string         `json:"method"`
	URL          string         `json:"url"`
	Headers      []KeyValuePair `json:"headers"`
	Params       []KeyValuePair `json:"params"`
	Body         string         `json:"body"`
	Auth         string         `json:"auth"`
}

type RequestSummaryResponse struct {
	ID           string `json:"id"`
	CollectionID string `json:"collection_id"`
	Name         string `json:"name"`
	Method       string `json:"method"`
}
