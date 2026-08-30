package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// MediaAsset is the record of a finalized upload: an object that has been
// confirmed to exist in storage and is available for a resource to reference.
// uploaded_by_user_id is provenance only, not an access-control boundary —
// once an asset is attached to a resource (menu item, restaurant, ...), that
// resource's own ownership/restaurant scope governs who can see or change it.
type MediaAsset struct {
	ent.Schema
}

func (MediaAsset) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
		mixin.CreateTime{},
	}
}

func (MediaAsset) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("uploaded_by_user_id", uuid.UUID{}).
			Immutable().
			Comment("User who performed the upload; provenance only, not an ACL"),
		field.UUID("upload_id", uuid.UUID{}).
			Immutable().
			Unique().
			Comment("The media upload session that produced this asset; at most one asset per session"),
		field.String("storage_key").
			NotEmpty().
			MaxLen(1024).
			Immutable().
			Unique().
			Comment("Provider-neutral storage object key"),
		field.String("content_type").
			NotEmpty().
			MaxLen(255).
			Immutable(),
		field.Int64("size_bytes").
			Min(0).
			Immutable(),
		field.Enum("status").
			Values("active", "deleted").
			Default("active"),
	}
}

func (MediaAsset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("uploaded_by", User.Type).
			Ref("uploaded_media_assets").
			Unique().
			Required().
			Immutable().
			Field("uploaded_by_user_id"),
		edge.From("upload", MediaUpload.Type).
			Ref("asset").
			Unique().
			Required().
			Immutable().
			Field("upload_id"),
	}
}

func (MediaAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uploaded_by_user_id"),
	}
}
