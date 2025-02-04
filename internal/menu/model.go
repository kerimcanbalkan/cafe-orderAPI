package menu

import "go.mongodb.org/mongo-driver/bson/primitive"

type MenuItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name"          json:"name"        validate:"required,min=2,max=60"`
	Description string             `bson:"description"   json:"description" validate:"required,min=5,max=150"`
	Price       float64            `bson:"price"         json:"price"       validate:"required,gt=0"`
	Category    string             `bson:"category"      json:"category"    validate:"required,min=2,max=60"`
	Img         string             `bson:"image"         json:"image"       validate:"required"`
}
