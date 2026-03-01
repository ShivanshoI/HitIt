package collaborators

type Collaborator struct {
}

type CreateCollaboratorDTO struct {
}

type ImportCollaboratorDTO struct {
	IDString string `json:"id_string"`
}

type CollaboratorResponse struct {
}

type LinkPayload struct {
	EntityType string `json:"entity_type"`
	Permission bool   `json:"permission"`
	IDString   string `json:"id_string"`
}
