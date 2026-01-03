package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type Order struct {
	ent.Schema
}

func (Order) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("order_type").
			Values("DINE_IN", "TAKEOUT", "DELIVERY"),
		field.Enum("order_status").
			Values("OPEN", "CONFIRMED", "COMPLETED", "CANCELLED").
			Default("OPEN"),
		field.Enum("payment_status").
			Values("UNPAID", "PENDING", "PAID", "REFUNDED").
			Default("UNPAID"),
		field.UUID("restaurant_id", uuid.UUID{}).
			Comment("ID of the restaurant this order belongs to"),
	}
}

func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("restaurant", Restaurant.Type).
			Ref("orders").
			Unique().
			Required().
			Field("restaurant_id"),
		edge.To("order_items", OrderItem.Type),
	}
}
