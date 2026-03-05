package collaborators

type Collaborator struct {
}

type CreateCollaboratorDTO struct {
}

type ImportCollaboratorDTO struct {
	IDString string `json:"id_string"`
}

type CollaboratorResponse struct {
	UserID          string `json:"user_id"`
	Name            string `json:"name"`
	EmailAddress    string `json:"email_address,omitempty"`
	WritePermission bool   `json:"write_permission"`
}

type LinkPayload struct {
	EntityType string `json:"entity_type"`
	Permission bool   `json:"permission"`
	IDString   string `json:"id_string"`
	IsNew      bool   `json:"is_new"`
}
