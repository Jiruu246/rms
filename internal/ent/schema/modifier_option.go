package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type ModifierOption struct {
	ent.Schema
}

func (ModifierOption) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (ModifierOption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Modifier option name"),
		field.Float("price").
			Default(0.0).
			Comment("Price of the modifier option"),
		field.String("image_url").
			Optional().
			Comment("Image URL for the modifier option"),
		field.Bool("available").
			Default(true).
			Comment("Whether the modifier option is available"),
		field.Bool("pre_select").
			Default(false).
			Comment("Whether the modifier option is pre-selected"),
		field.UUID("modifier_id", uuid.UUID{}).
			Comment("ID of the modifier this option belongs to"),
	}
}

func (ModifierOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("modifier", Modifier.Type).
			Ref("modifier_options").
			Unique().
			Required().
			Field("modifier_id"),
	}
}
