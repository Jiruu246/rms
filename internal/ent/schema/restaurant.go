package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type Restaurant struct {
	ent.Schema
}

func (Restaurant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (Restaurant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).Unique().
			Immutable(),
		field.String("name").NotEmpty(),
		field.Text("description").Optional(),
		field.String("phone").NotEmpty(),
		field.String("email").NotEmpty(),
		field.String("address").NotEmpty(),
		field.String("city").NotEmpty(),
		field.String("state").NotEmpty(),
		field.String("zip_code").NotEmpty(),
		field.String("country").NotEmpty(),
		field.String("logo_url").Optional(),
		field.String("cover_image_url").Optional(),
		field.Enum("status").Values("active", "inactive", "closed").Default("active"),
		field.JSON("operating_hours", map[string]any{}).Optional(),
		field.String("currency"),
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (Restaurant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("restaurants").
			Unique().
			Required().
			Field("user_id"),
		edge.To("menu_items", MenuItem.Type),
		edge.To("categories", Category.Type),
		edge.To("modifiers", Modifier.Type),
	}
}
