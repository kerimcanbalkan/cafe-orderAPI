package order

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Items       []menu.MenuItem    `bson:"items"         json:"items"       validate:"required"`
	TotalPrice  float32            `bson:"totalPrice"    json:"totalPrice"`
	TableNumber int                `bson:"tableNumber"   json:"tableNumber"`
	Status      bool               `bson:"status"        json:"status"`
	Served      bool               `bson:"served"        json:"served"`
	CreatedAt   time.Time          `bson:"createdAt"     json:"createdAt"`
}
