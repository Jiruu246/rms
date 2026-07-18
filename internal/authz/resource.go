package authz

import "github.com/google/uuid"

// Resource describes the object an Actor wants to act on: the authorization
// boundary (RestaurantID — see CLAUDE.md, every entity in this system is
// scoped to a restaurant), the owning user, and any extra attributes a
// policy needs. Callers build this from a repo lookup before calling
// Authorizer.Authorize; it is never persisted.
type Resource struct {
	Type         string
	ID           uuid.UUID
	RestaurantID uuid.UUID
	OwnerUserID  uuid.UUID
	Attributes   map[string]string
}
