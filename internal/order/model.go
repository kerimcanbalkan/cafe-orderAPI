package order

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"       json:"id"`
	Items       []menu.MenuItem    `bson:"items"               json:"items"       validate:"required"`
	TotalPrice  float64            `bson:"totalPrice"          json:"totalPrice"`
	TableNumber uint8              `bson:"tableNumber"         json:"tableNumber"`
	ServedAt    *time.Time         `bson:"servedAt,omitempty"  json:"servedAt"`
	CreatedAt   time.Time          `bson:"createdAt"           json:"createdAt"`
	HandledBy   primitive.ObjectID `bson:"handledBy,omitempty" json:"handledBy"`
	ClosedAt    *time.Time         `bson:"closedAt,omitempty"  json:"closedAt"`
}
