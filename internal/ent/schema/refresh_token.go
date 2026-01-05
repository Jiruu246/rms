package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type RefreshToken struct {
	ent.Schema
}

func (RefreshToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
}

func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.String("token").
			NotEmpty().
			Unique().
			Comment("The refresh token value"),
		field.Time("expires_at").
			Comment("When the refresh token expires"),
		field.Bool("revoked").
			Default(false).
			Comment("Whether the token has been revoked"),
		field.Time("revoked_at").
			Optional().
			Nillable().
			Comment("When the token was revoked"),
		field.UUID("replaced_by", uuid.New()).
			Optional().
			Nillable().
			Comment("ID of the token that replaced this one during rotation"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("When the token was last used"),
	}
}

// Edges of the RefreshToken.
func (RefreshToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("refresh_tokens").
			Field("user_id").
			Required().
			Immutable().
			Unique().
			Comment("The user this token belongs to"),
		edge.To("replaced_by_token", RefreshToken.Type).
			Field("replaced_by").
			Unique().
			Comment("Token that replaced this one"),
	}
}

func (RefreshToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("revoked", "expires_at"),
		index.Fields("expires_at"),
	}
}
