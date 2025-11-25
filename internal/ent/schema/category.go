package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type Category struct {
	ent.Schema
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Category name"),
		field.String("description").
			Default("").
			MaxLen(1000).
			Comment("Category description"),
		field.Int("display_order").
			Default(0).
			Min(0).
			Comment("Display order for sorting"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether the category is active"),
	}
}

func (Category) Edges() []ent.Edge {
	return nil
}
