package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type Customer struct {
	ent.Schema
}

func (Customer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.UpdateTime{},
	}
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
		field.String("password_hash").
			NotEmpty().
			Comment("Hashed password for authentication"),
	}
}

// Edges of the Customer.
func (Customer) Edges() []ent.Edge {
	return nil
}
