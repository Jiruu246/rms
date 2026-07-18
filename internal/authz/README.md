# Ownership / Authorization Framework

Status: **Restaurant is wired up** (see "Applied resources" below); every
other resource still has no ownership check and is wired up one at a time in
later iterations (see "Applying to a resource" below).

## Why this exists

Ownership checks were either missing or done ad hoc (e.g.
`restaurant_handler.go`'s `Get`/`Update`/`Delete` used to not check that the
caller owned the restaurant at all; `Create` read `claims` inline). As more
resources and roles are added, repeating that logic per-handler drifts out
of sync. This package gives every resource one shared place to ask "can this
actor do this action on this resource" and one shared place to change the
answer.

## Responsibility split

```
Middleware   -> Who is this? (validate JWT, build an Actor)
Service      -> What resource, in what business operation?
Authorizer   -> Is this Actor allowed to do this Action on this Resource?
Repository   -> Enforce the scope in the query itself (defense in depth)
```

Concretely:

- **Middleware** (`internal/middlewares/jwt.go`) authenticates and produces
  claims. It should *not* decide ownership — it doesn't know which resource
  is being touched or what the action is.
- **Handlers** stay thin: parse input, resolve the `Actor`, call the service.
  No `if resource.OwnerID != actor.UserID` checks in handler code.
- **Services** are where authorization happens: resolve the target's
  `Resource`, call `Authorizer.Authorize`, then call the repo. This is the
  layer that knows both the action being performed and how to look up the
  resource's ownership.
- **Repositories** should still scope mutating/reading queries by
  `restaurant_id` (or whatever the resolved scope is) as a second line of
  defense — not rely on the service check alone.

## Types (all in `internal/authz`)

| File | Type | Purpose |
|------|------|---------|
| `actor.go` | `Actor{UserID, Role}` | The authenticated caller. Built via `NewActorFromClaims(utils.JWTClaims)`. Carries no DB state. |
| `actor.go` | `WithActor` / `ActorFromContext` | Optional context plumbing, for when a plain `context.Context` needs to carry the actor deeper than a function argument. |
| `resource.go` | `Resource{Type, ID, RestaurantID, OwnerUserID, Attributes}` | Describes *what* is being acted on. `RestaurantID` is the tenant boundary — every entity in this system hangs off a restaurant (see root `CLAUDE.md`). `Attributes` is an escape hatch for policy rules that need something beyond ownership/restaurant. |
| `action.go` | `Action` (string) | Names the operation, `"resource:verb"` (e.g. `"menu_item:update"`). Constants are declared per-resource, next to that resource's service — not in this package. |
| `decision.go` | `Decision{Reason, Mode}`, `AccessMode` | What was decided and why, for audit logging. `nil` error from `Authorize` = allowed; the sentinel errors in `errors.go` are the only failure signal. |
| `authorizer.go` | `Authorizer` interface, `Request{Actor, Action, Resource}` | The single method services call: `Authorize(ctx, Request) (Decision, error)`. |
| `policy.go` | `PolicyAuthorizer` | Default `Authorizer` impl: admin role bypasses everything; otherwise `Resource.OwnerUserID == Actor.UserID` or deny. |
| `errors.go` | `ErrUnauthenticated`, `ErrForbidden`, `ErrNotFound`, `AsNotFound(err)` | Sentinels a handler maps to HTTP status. `AsNotFound` turns a forbidden decision into "not found" for endpoints that shouldn't reveal a resource exists. |

## Why `Resource` instead of `OwnerUserID` alone

`PolicyAuthorizer` only reads `Resource.RestaurantID`/`OwnerUserID` today, but
the field set anticipates the next step being **restaurant memberships**
(multiple users per restaurant with different roles) rather than a single
`owner_user_id`. When that lands, only `Resource` construction (in each
repo) and `PolicyAuthorizer.Authorize` change — `Action`, `Authorizer`, and
every call site stay the same.

## Applied resources

| Resource | Actions | Notes |
|----------|---------|-------|
| Restaurant | `restaurant:read`, `restaurant:update`, `restaurant:delete` (in `restaurant_service.go`) | Only the restaurant's `OwnerUserID` (== `Restaurant.user_id`) may read/update/delete it; `Create` has no check (any authenticated actor may create a restaurant they then own). `GetAll` is unscoped — it still lists every restaurant, not just the caller's own; not addressed by this pass. Resource resolution: `RestaurantRepository.GetAuthorizationResource`. Handler maps `authz.ErrForbidden` to 403; every other error keeps its pre-existing status per endpoint (404 for Get/Delete, 500 for Update — see `writeAuthzError` in `restaurant_handler.go`). No scoped repo queries (step 4 below) were added yet — the only defense is the service-layer `Authorize` call. |

## Applying to a resource (next iteration, per entity)

For each remaining resource (menu item, category, modifier, order, ...):

1. Declare its `Action` constants next to its service, e.g.
   ```go
   const (
       ActionUpdateRestaurant authz.Action = "restaurant:update"
       ActionDeleteRestaurant authz.Action = "restaurant:delete"
   )
   ```
2. Add a repo method that resolves a `authz.Resource` for a given ID — for
   restaurant this is trivial (`Resource{RestaurantID: id, OwnerUserID: row.UserID}`);
   for child entities it means joining up to the owning restaurant.
3. In the service method, before calling the repo mutation: resolve the
   `Resource`, call `authorizer.Authorize(ctx, authz.Request{...})`, return
   the error (mapped by the handler) if non-nil.
4. Add a scoped repo method (`UpdateInRestaurant(ctx, restaurantID, id, ...)`)
   so the query itself is filtered by the authorized scope — don't rely on
   the service check as the only safeguard.
5. In the handler, map `authz.ErrForbidden`/`authz.ErrUnauthenticated` to
   `utils.WriteForbidden`/`utils.WriteUnauthorized` (or run the result
   through `authz.AsNotFound` first — see "403 vs 404" below).

For child resources (menu item, category, ...), step 2's "join up to the
owning restaurant" is the main new work — `Resource.OwnerUserID` should
still resolve to the *restaurant's* owner, since none of these entities
have their own owner field.

## Decisions deferred to when we apply this

These are open questions the framework doesn't need answered yet, but each
resource's integration should make an explicit choice:

- **403 vs 404** — whether "not your restaurant" should read as forbidden or
  not-found, per endpoint. Use `authz.AsNotFound` where existence shouldn't
  leak.
- **Roles vs permissions** — `Actor.Role` is a single string today (matching
  the existing `JWTClaims.Role` claim). If more granular roles show up
  (manager, cashier, ...), prefer adding permission checks inside
  `PolicyAuthorizer` over branching on role strings at call sites.
- **Membership store** — if/when restaurants gain multiple owning users,
  `PolicyAuthorizer` gains a lookup (e.g. a `MembershipRepository`
  dependency) instead of every service doing its own membership check.
- **Audit logging** — `Decision.Reason`/`Mode` are already shaped for this;
  no audit sink is wired up yet.
