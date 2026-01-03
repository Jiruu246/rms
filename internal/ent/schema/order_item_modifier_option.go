package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type OrderItemModifierOption struct {
	ent.Schema
}

func (OrderItemModifierOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(
			"order_item_id",
			"modifier_option_id").
			Unique(),
	}
}

func (OrderItemModifierOption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("order_item_id", uuid.UUID{}),
		field.UUID("modifier_option_id", uuid.UUID{}),
		field.Int("quantity").
			Default(1).
			Min(1).
			Comment("Quantity of the modifier option selected"),
		field.String("option_name").
			NotEmpty().
			Comment("Snapshot of the modifier option name at the time of order"),
		field.Float("option_price").
			Comment("Snapshot of the modifier option price at the time of order"),
	}
}

func (OrderItemModifierOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order_item", OrderItem.Type).
			Ref("order_item_modifier_options").
			Unique().
			Required().
			Field("order_item_id"),
		edge.From("modifier_option", ModifierOption.Type).
			Ref("order_item_modifier_options").
			Unique().
			Required().
			Field("modifier_option_id"),
	}
}
