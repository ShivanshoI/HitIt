package users

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// FindMultipleByIDs retrieves multiple users by their IDs
func (r *UserRepository) FindMultipleByIDs(ctx context.Context, ids []primitive.ObjectID) ([]User, error) {
	if len(ids) == 0 {
		return []User{}, nil
	}

	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}

	cursor, err := r.collection.Find(ctx, filter)
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

// FindByEmail retrieves a user by their email address
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"email_address": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfile updates only the display fields (first/last name, email, theme) for a user.
func (r *UserRepository) UpdateProfile(ctx context.Context, id, firstName, lastName, email, theme string) (*User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	setFields := bson.M{
		"first_name": firstName,
		"updated_at": time.Now(),
	}
	if email != "" {
		setFields["email_address"] = email
	}
	if lastName != "" {
		setFields["last_name"] = lastName
	}
	if theme != "" {
		setFields["theme"] = theme
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": setFields})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

// UpdatePassword replaces only the password field for a user.
func (r *UserRepository) UpdatePassword(ctx context.Context, id, newPassword string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"password":   newPassword,
			"updated_at": time.Now(),
		},
	})
	return err
}

// ExistsByEmailExcludingID checks if any other user already uses this email.
func (r *UserRepository) ExistsByEmailExcludingID(ctx context.Context, email, excludeID string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(excludeID)
	if err != nil {
		return false, err
	}
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"email_address": email,
		"_id":           bson.M{"$ne": objID},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email_address", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "phone_number", Value: 1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// CountByOrganizationID returns the number of users that belong to this org
func (r *UserRepository) CountByOrganizationID(ctx context.Context, orgID string) (int, error) {
	objID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return 0, err
	}
	count, err := r.collection.CountDocuments(ctx, bson.M{"organization_id": objID})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ConnectToOrganization updates the user's OrganizationID
func (r *UserRepository) ConnectToOrganization(ctx context.Context, userID string, orgID string) error {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	oID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": uID},
		bson.M{
			"$set": bson.M{
				"organization_id": oID,
				"updated_at":      time.Now(),
			},
		},
	)
	return err
}
