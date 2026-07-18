package authz

import (
	"context"

	"github.com/google/uuid"

	"github.com/Jiruu246/rms/pkg/utils"
)

// Actor is the authenticated identity performing a request. It is built once
// by auth middleware/handlers from validated JWT claims and carried by value
// from there on — it never triggers a DB lookup itself.
type Actor struct {
	UserID uuid.UUID
	Role   string
}

// NewActorFromClaims adapts the existing utils.JWTClaims (see pkg/utils/jwt_utils.go)
// into an Actor. This is the only place that should know about the JWT claims shape.
func NewActorFromClaims(claims utils.JWTClaims) Actor {
	return Actor{UserID: claims.UserID, Role: claims.Role}
}

func (a Actor) HasRole(role string) bool {
	return a.Role == role
}

type actorContextKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorContextKey{}).(Actor)
	return actor, ok
}
