package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type UserAuthProvider struct {
	ent.Schema
}

func (UserAuthProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (UserAuthProvider) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.String("provider").
			NotEmpty().
			MaxLen(100),
		field.String("provider_user_id").
			NotEmpty().
			MaxLen(255),
	}
}

func (UserAuthProvider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("auth_providers").
			Field("user_id").
			Required().
			Immutable().
			Unique(),
	}
}
