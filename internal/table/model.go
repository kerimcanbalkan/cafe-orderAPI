package table

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"       json:"id"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
}
