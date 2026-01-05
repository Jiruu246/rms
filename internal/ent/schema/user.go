package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.String("email").
			Optional().
			Unique(),
		field.Bool("email_verified").
			Default(false),
		field.String("phone_number").
			Default(""),
		field.Bool("is_active").
			Default(true),
		field.String("password_hash").
			Optional().
			Nillable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("restaurants", Restaurant.Type),
		edge.To("auth_providers", UserAuthProvider.Type),
		edge.To("refresh_tokens", RefreshToken.Type),
	}
}
