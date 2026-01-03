package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type Modifier struct {
	ent.Schema
}

func (Modifier) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (Modifier) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Modifier name"),
		field.Bool("required").
			Default(false).
			Comment("Whether the modifier is required"),
		field.Bool("multi_select").
			Default(false).
			Comment("Whether multiple selections are allowed"),
		field.Int("max").
			Default(1).
			Min(0).
			Comment("Maximum number of selections allowed"),
		field.UUID("restaurant_id", uuid.UUID{}).
			Comment("ID of the restaurant this modifier belongs to"),
	}
}

func (Modifier) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("restaurant", Restaurant.Type).
			Ref("modifiers").
			Unique().
			Required().
			Field("restaurant_id"),
		edge.To("modifier_options", ModifierOption.Type),
		edge.To("menu_items", MenuItem.Type),
	}
}
