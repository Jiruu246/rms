package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type MenuItem struct {
	ent.Schema
}

func (MenuItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (MenuItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Immutable().
			Unique(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Menu item name"),
		field.String("description").
			Default("").
			MaxLen(1000).
			Comment("Menu item description"),
		field.Float("price").
			Min(0).
			Comment("Menu item price"),
		field.String("image_url").
			Optional().
			Comment("URL of the menu item image"),
		field.Bool("is_available").
			Default(true).
			Comment("Whether the menu item is available"),
		field.UUID("restaurant_id", uuid.UUID{}).
			Comment("ID of the restaurant this menu item belongs to"),
		field.UUID("category_id", uuid.UUID{}).
			Optional().
			Comment("ID of the category this menu item belongs to"),
	}
}

func (MenuItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("restaurant", Restaurant.Type).
			Ref("menu_items").
			Unique().
			Required().
			Field("restaurant_id"),
		edge.From("category", Category.Type).
			Ref("menu_items").
			Unique().
			Field("category_id"),
		edge.To("modifiers", Modifier.Type),
		edge.To("order_items", OrderItem.Type),
	}
}
