package authz

import (
	"context"

	"github.com/google/uuid"
)

// RoleAdmin is the platform-wide bypass role, matched against Actor.Role
// (== the "role" claim already issued by GenerateJWT — see pkg/utils/jwt_utils.go).
const RoleAdmin = "admin"

// PolicyAuthorizer is the default Authorizer: platform admins may do
// anything; everyone else must own the resource (Resource.OwnerUserID ==
// Actor.UserID). Extend the checks here — not at call sites — when new
// roles, permissions, or restaurant memberships are introduced (see
// README.md "Extending the policy").
type PolicyAuthorizer struct{}

func NewPolicyAuthorizer() *PolicyAuthorizer {
	return &PolicyAuthorizer{}
}

func (p *PolicyAuthorizer) Authorize(ctx context.Context, req Request) (Decision, error) {
	if req.Actor.HasRole(RoleAdmin) {
		return Decision{Mode: AccessAdmin, Reason: "admin role"}, nil
	}

	if req.Resource.OwnerUserID != uuid.Nil && req.Resource.OwnerUserID == req.Actor.UserID {
		return Decision{Mode: AccessOwner, Reason: "resource owner"}, nil
	}

	return Decision{}, ErrForbidden
}
