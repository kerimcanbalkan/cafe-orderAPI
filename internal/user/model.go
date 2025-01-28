package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name"          json:"name"      validate:"required,min=2,max=20"`
	Surname   string             `bson:"surname"       json:"surname"   validate:"required,min=2,max=20"`
	Gender    string             `bson:"gender"        json:"gender"    validate:"required,oneof=male female"`
	Email     string             `bson:"email"         json:"email"     validate:"required,email"`
	Username  string             `bson:"username"      json:"username"  validate:"required,min=2,max=20"`
	Password  string             `bson:"password"      json:"password"  validate:"required,min=8,max=20"`
	Role      string             `bson:"role"          json:"role"      validate:"required,oneof=admin cashier waiter"`
	CreatedAt time.Time          `bson:"created_at"    json:"createdAt"`
}
