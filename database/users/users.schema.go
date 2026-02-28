package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName    string             `bson:"first_name" json:"first_name"`
	LastName     *string            `bson:"last_name,omitempty" json:"last_name,omitempty"`
	NickName     *string            `bson:"nick_name,omitempty" json:"nick_name,omitempty"`
	PhoneNumber  *string            `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
	EmailAddress *string            `bson:"email_address,omitempty" json:"email_address,omitempty"`
	Password     string             `bson:"password" json:"password"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
