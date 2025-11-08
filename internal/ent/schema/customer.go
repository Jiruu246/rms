package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Customer struct {
	ent.Schema
}

// Fields of the Customer.
func (Customer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Customer name"),
		field.String("email").
			NotEmpty().
			Unique().
			Comment("Customer email"),
		field.String("phone_number").
			Default("").
			Comment("Customer phone number"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether the customer is active"),
		field.Time("created_at").
			Default(time.Now()).
			Immutable().
			Annotations(entsql.Annotation{
				Default: "CURRENT_TIMESTAMP",
			}).
			Comment("Creation timestamp"),
		field.String("password_hash").
			NotEmpty().
			Comment("Hashed password for authentication"),
	}
}

// Edges of the Customer.
func (Customer) Edges() []ent.Edge {
	return nil
}
