package order

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"       json:"id"`
	Items       []menu.MenuItem    `bson:"items"               json:"items"       validate:"required"`
	TotalPrice  float64            `bson:"total_price"          json:"totalPrice"`
	TableNumber uint8              `bson:"table_number"         json:"tableNumber"`
	ServedAt    *time.Time         `bson:"served_at,omitempty"  json:"servedAt"`
	CreatedAt   time.Time          `bson:"created_at"           json:"createdAt"`
	HandledBy   primitive.ObjectID `bson:"handled_by,omitempty" json:"handledBy"`
	ClosedAt    *time.Time         `bson:"closed_at,omitempty"  json:"closedAt"`
	ClosedBy    primitive.ObjectID `bson:"closed_by,omitempty"  json:"closedBy"`
}
