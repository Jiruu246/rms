# Ownership / Authorization Framework

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
  `restaurant_id` scope as a second line of
  defense — not rely on the service check alone.

## Types (all in `internal/authz`)

| File | Type | Purpose |
|------|------|---------|
| `actor.go` | `Actor{UserID, Role}` | The authenticated caller. Built via `NewActorFromClaims(utils.JWTClaims)`. Carries no DB state. |
| `actor.go` | `WithActor` / `ActorFromContext` | Optional context plumbing, for when a plain `context.Context` needs to carry the actor deeper than a function argument. |
| `resource.go` | `Resource{Type, ID, RestaurantID, OwnerUserID, Attributes}` | Describes *what* is being acted on. `RestaurantID` is the tenant boundary — every entity in this system hangs off a restaurant (see root `CLAUDE.md`). `Attributes` is an escape hatch for policy rules that need something beyond ownership/restaurant. |
| `action.go` | `Action` (string) | Names the operation, `"resource:verb"` (e.g. `"menu_item:update"`). Constants are declared per-resource, next to that resource's service — not in this package. |
| `decision.go` | `Decision{Reason, Mode}`, `AccessMode` | What was decided and why, for audit logging. `nil` error from `Authorize` = allowed; a denied decision fails with `apperr.ErrForbidden` (see `documentation/ErrorHandlingFramework.md`) — this package defines no sentinels of its own. |
| `authorizer.go` | `Authorizer` interface, `Request{Actor, Action, Resource}` | The single method services call: `Authorize(ctx, Request) (Decision, error)`. |
| `policy.go` | `PolicyAuthorizer` | Default `Authorizer` impl: admin role bypasses everything; otherwise `Resource.OwnerUserID == Actor.UserID` or deny (returns `apperr.ErrForbidden`). |

## Applying to a resource (per entity)

Category (`internal/services/category_service.go`) is the first entity
fully wired up — use it as the reference implementation. For each
remaining resource (menu item, modifier, order, ...):

1. Declare its `Action` constants next to its service, e.g.
   ```go
   const (
       ActionUpdateRestaurant authz.Action = "restaurant:update"
       ActionDeleteRestaurant authz.Action = "restaurant:delete"
   )
   ```
2. Make sure the entity's schema carries a denormalized `restaurant_id`
   column + edge straight to `Restaurant` — even for resources nested more
   than one level deep (e.g. `ModifierOption` under `Modifier`, `OrderItem`
   and `OrderItemModifierOption` under `Order`). Populate it from the
   parent's `restaurant_id` at creation time, rather than only storing the
   immediate parent's ID. Then add a repo method that resolves a
   `authz.Resource` for a given ID — for restaurant this is trivial
   (`Resource{RestaurantID: id, OwnerUserID: row.UserID}`); for every other
   entity it's a single join to `Restaurant` on that denormalized
   `restaurant_id` (see `CategoryRepository.GetAuthorizationResource` — ent
   compiles `QueryRestaurant()` to one SQL statement when the FK is present
   directly on the entity, not a walk up the parent chain). `OwnerUserID`
   only lives on `Restaurant`, so this join can't be skipped entirely — the
   denormalized column just keeps it a single hop instead of N, regardless
   of how deep the entity is nested.
3. The actual check takes one of two shapes, depending on what's being
   authorized:
   - **Against the entity's own resource** (read/update/delete an existing
     row): resolve it via the repo method from step 2, then call
     `authorizer.Authorize(ctx, authz.Request{...})` directly — see
     `categoryService.authorize`, a small per-service helper that does this
     and returns the resolved `Resource` so the caller can scope the
     follow-up repo query.
   - **Against a restaurant the entity doesn't exist under yet, or is being
     listed under** (create, or list-scoped-by-`restaurant_id`): don't
     duplicate the resolve-then-authorize dance in every entity service —
     call `RestaurantService.AuthorizeOwnership(ctx, actor, action,
     restaurantID)` instead. This is why `CategoryService` takes a
     `RestaurantService` dependency rather than a `RestaurantRepository` —
     see `categoryService.Create` and `categoryService.List`.
4. Add a scoped repo method (`UpdateInRestaurant(ctx, restaurantID, id, ...)`)
   so the query itself is filtered by the authorized scope — don't rely on
   the service check as the only safeguard.
5. In the handler, do nothing special — `apperr.WriteHTTPError(c.Writer, err,
   fallback)` already renders `apperr.ErrForbidden` as a 404 (see "403 vs
   404" below).

Menu item, modifier, and order already carry `restaurant_id` directly (see
their ent schemas), so for those step 2 is just the repo method. For
resources nested under *those* (`ModifierOption`, `OrderItem`,
`OrderItemModifierOption`, ...), adding the denormalized `restaurant_id`
column is the main new work — don't rely on walking
`child -> parent -> restaurant` through ent edges for something that will
run on every authorized request. `Resource.OwnerUserID` should still
resolve to the *restaurant's* owner, since none of these entities have
their own owner field.

## Decisions deferred to when we apply this

These are open questions the framework doesn't need answered yet, but each
resource's integration should make an explicit choice:

  (`errors.Is(err, apperr.ErrForbidden)` keeps working for logging/audit) —
  only the HTTP response collapses forbidden into not-found. There is no
  per-endpoint opt-out; if a future endpoint genuinely needs to reveal "this
  exists but you can't touch it" (e.g. a paywalled resource), that would be a
  deliberate new sentinel, not a flag on this one.
- **Roles vs permissions** — `Actor.Role` is a single string today (matching
  the existing `JWTClaims.Role` claim). If more granular roles show up
  (manager, cashier, ...), prefer adding permission checks inside
  `PolicyAuthorizer` over branching on role strings at call sites.
- **Membership store** — if/when restaurants gain multiple owning users,
  `PolicyAuthorizer` gains a lookup (e.g. a `MembershipRepository`
  dependency) instead of every service doing its own membership check.
- **Audit logging** — `Decision.Reason`/`Mode` are already shaped for this;
  no audit sink is wired up yet.
