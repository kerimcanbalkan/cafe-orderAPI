package order

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"       json:"id"`
	Items       []menu.MenuItem    `bson:"items"               json:"items"       validate:"required"`
	TotalPrice  float32            `bson:"totalPrice"          json:"totalPrice"`
	TableNumber int                `bson:"tableNumber"         json:"tableNumber"`
	IsClosed    bool               `bson:"isClosed"            json:"isClosed"`
	ServedAt    *time.Time         `bson:"servedAt,omitempty"  json:"servedAt"`
	CreatedAt   time.Time          `bson:"createdAt"           json:"createdAt"`
	HandledBy   primitive.ObjectID `bson:"handledBy,omitempty" json:"handledBy"`
}
