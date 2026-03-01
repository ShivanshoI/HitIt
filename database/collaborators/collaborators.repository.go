package collaborators

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const CollectionName = "collaborators"

type CollaboratorRepository struct {
	collection *mongo.Collection
}

func NewCollaboratorRepository(db *mongo.Database) *CollaboratorRepository {
	return &CollaboratorRepository{
		collection: db.Collection(CollectionName),
	}
}
