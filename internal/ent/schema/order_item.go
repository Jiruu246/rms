package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type OrderItem struct {
	ent.Schema
}

func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("quantity").
			Default(1).
			Min(1).
			Comment("Quantity of the menu item ordered"),
		field.String("special_instructions").
			Optional().
			Comment("Special instructions for the order item"),
		field.String("item_name").
			NotEmpty().
			Comment("Snapshot of the menu item name at the time of order"),
		field.Float("item_price").
			Comment("Snapshot of the menu item price at the time of order"),
		field.Int64("menu_item_id").
			Comment("ID of the menu item"),
		field.UUID("order_id", uuid.UUID{}).
			Comment("ID of the order this item belongs to"),
	}
}

func (OrderItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("order_items").
			Unique().
			Required().
			Field("order_id"),
		edge.From("menu_item", MenuItem.Type).
			Ref("order_items").
			Unique().
			Required().
			Field("menu_item_id"),
		edge.To("order_item_modifier_options", OrderItemModifierOption.Type),
	}
}
