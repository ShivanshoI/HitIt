package users

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const CollectionName = "users"

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection(CollectionName),
	}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *User) (*User, error) {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	log.Printf("[REPO] Attempting to insert user into collection: %s", r.collection.Name())
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("[REPO] Insert error: %v", err)
		return nil, err
	}
	log.Printf("[REPO] Successfully inserted user with ID: %s", user.ID.Hex())
	return user, nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user User
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByIdentifierAndPassword retrieves a user by their email or phone and plain text password
func (r *UserRepository) GetByIdentifierAndPassword(ctx context.Context, identifier, password string) (*User, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"email_address": identifier},
			{"phone_number": identifier},
		},
		"password": password,
	}

	var user User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No matching user found
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ExistsByIdentifier(ctx context.Context, identifier string) (bool, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"email_address": identifier},
			{"phone_number": identifier},
		},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, id string, user *User) (*User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	user.UpdatedAt = time.Now()
	update := bson.M{
		"$set": user,
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Delete removes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// List retrieves all users
func (r *UserRepository) List(ctx context.Context) ([]User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
