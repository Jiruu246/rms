package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Category holds the schema definition for the Category entity.
type Category struct {
	ent.Schema
}

// Fields of the Category.
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
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Annotations(entsql.Annotation{
				Default: "CURRENT_TIMESTAMP",
			}).
			Comment("Creation timestamp"),
	}
}

// Edges of the Category.
func (Category) Edges() []ent.Edge {
	return nil
}
