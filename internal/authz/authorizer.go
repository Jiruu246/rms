package authz

import "context"

type Request struct {
	Actor    Actor
	Action   Action
	Resource Resource
}

// Authorizer decides whether an Actor may perform an Action on a Resource.
// A nil error means allowed (details in Decision); a non-nil error is always
// one of the sentinels in errors.go. Services call this after resolving the
// target's Resource and before touching the repository layer — it is the
// single place ownership/role rules live, so no handler or repo should
// reimplement this check inline.
type Authorizer interface {
	Authorize(ctx context.Context, req Request) (Decision, error)
}
