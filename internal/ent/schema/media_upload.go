package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type MediaUpload struct {
	ent.Schema
}

func (MediaUpload) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
		mixin.CreateTime{},
	}
}

func (MediaUpload) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("owner_id", uuid.UUID{}).
			Immutable().
			Comment("User who requested this upload session; only this user may consume it"),
		field.Enum("purpose").
			Values("menu_item_image", "restaurant_logo", "restaurant_cover_image").
			Immutable().
			Comment("What the resulting asset will be used for"),
		field.String("object_key").
			NotEmpty().
			MaxLen(1024).
			Immutable().
			Unique().
			Comment("Provider-neutral storage object key granted for this session"),
		field.Enum("status").
			Values("issued", "consumed", "failed").
			Default("issued").
			Comment("failed: the uploaded object was rejected at consume time (e.g. content-type/size mismatch) and cleaned up from storage; terminal, never retried. Expiry is not a distinct status — it is derived from expires_at against the current time"),
		field.Time("expires_at").
			Immutable().
			Comment("Grant deadline; consuming after this time is rejected"),
	}
}

func (MediaUpload) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("media_uploads").
			Unique().
			Required().
			Immutable().
			Field("owner_id"),
		edge.To("asset", MediaAsset.Type).
			Unique(),
	}
}

// Indexes support the two access patterns this session type needs: looking up
// a caller's own sessions by state, and (for a future expiry sweep — not
// implemented yet, see documentation) scanning issued sessions past expiry.
func (MediaUpload) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "status"),
		index.Fields("status", "expires_at"),
	}
}
