# Transactor Framework

This package gives services a way to run repo calls — and calls into other
services — as one atomic unit, without the service layer importing or
knowing about `internal/ent`. It exists to solve cross-repo/cross-service
atomicity (e.g. "create an order, its items, and its modifier options, or
none of it") without threading `*ent.Tx` through every interface signature.

## Responsibility split

```
Service      -> composes repo + service calls inside Transactor.WithinTx
Transactor   -> port for "run this atomically"; the outermost call owns
                commit/rollback, any call nested inside it just runs inline
Repository   -> resolves the active client from ctx before every operation
                (not yet wired — repos still hold a fixed client today)
```

Concretely:

- **Services** never see `*ent.Client`/`*ent.Tx`. They depend on
  `repos.Transactor` and wrap a block of repo/service calls in
  `transactor.WithinTx(ctx, func(ctx context.Context) error { ... })`.
- **`EntTransactor`** is the only piece that knows Ent. It checks whether a
  transaction is already active on `ctx`: if so it just calls `fn(ctx)`
  inline (nested case); otherwise it begins a real `*ent.Tx`, stashes the
  tx-scoped client in a child context, and commits or rolls back based on
  what `fn` returns.
- **Repositories** stop holding a single fixed client for their whole
  lifetime and instead resolve "the client to use right now" from `ctx` at
  the top of every method, falling back to their own default client when no
  transaction is active. A repo never knows or cares whether it's inside a
  transaction.

## Types (all in `internal/repos`)

| File | Type | Purpose |
|------|------|---------|
| `transactor.go` | `Transactor` interface — `WithinTx(ctx, fn func(context.Context) error) error` | The port services depend on. Contains no Ent types, so importing it doesn't pull Ent into the service layer. |
| `transactor.go` | `WithinTxResult[T](ctx, tr, fn func(context.Context) (T, error)) (T, error)` | Wraps `Transactor` for callers that need a value back. Assigns the result only after `fn` succeeds, and only returns it once `WithinTx` itself returned nil — so "don't read the result until you've checked the error" is structural, not a convention to remember. |
| `ent_transactor.go` | `EntTransactor` (impl of `Transactor`) | Begins/commits/rolls back a real `*ent.Tx` for the outermost `WithinTx` call; detects nesting and runs `fn` inline otherwise. Panic-safe: recovers just long enough to roll back, then re-panics. |
| `ent_transactor.go` | `clientFromContext(ctx, def *ent.Client) *ent.Client` (unexported) | What a migrated repo method calls instead of using its held `client` field. Returns the tx-scoped client if one is active on `ctx`, otherwise `def`. Works because `ent.Tx.Client()` returns a plain `*ent.Client` — the same type every repo already holds — so there is no separate transactional code path to maintain. |

## Behavior contract

1. **Outermost call wins.** The first `WithinTx` in a chain opens the real
   transaction and owns commit/rollback. Any `WithinTx` nested inside it —
   the same service calling itself, or a different service reached through
   the same `ctx` — detects the transaction already in context and just
   executes inline.
2. **Context propagation is mandatory.** Every nested call, service-to-service
   or service-to-repo, must forward the *same* `ctx` it received. Swapping in
   `context.Background()` (or any fresh context) anywhere in the chain
   silently breaks atomicity downstream of that point — no error is raised.
3. **Rollback on error or panic.** A non-nil error from the closure, or a
   panic, rolls back the outermost transaction. A panic is re-raised after
   rollback, never swallowed.
4. **One shared client/DB only.** This only works for in-process calls
   sharing one Ent client / database connection. It cannot span separate
   services over the network or separate databases.

## Sample usage

These sketches show the target shape for `order_repo.go`/`OrderService`,
which are not migrated yet — see "Current state" above for what actually
exists today (`restaurant_repo.go`, `media_upload_repo.go`'s `Consume`,
`MediaService`, and `RestaurantService`).

**A migrated repo method.** `order_repo.go`'s `Create` today opens its own
`client.Tx(ctx)`. Once migrated, it just resolves whatever client is active
and stops managing transactions itself — the caller decides that:

```go
func (r *orderRepository) Create(ctx context.Context, data *CreateOrderData) (*dto.Order, error) {
	c := clientFromContext(ctx, r.client)

	ord, err := c.Order.Create().
		SetOrderType(order.OrderType(data.OrderType)).
		SetOrderStatus(order.OrderStatus(data.OrderStatus)).
		SetPaymentStatus(order.PaymentStatus(data.PaymentStatus)).
		SetRestaurantID(data.RestaurantID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range data.OrderItems {
		if _, err := c.OrderItem.Create().
			SetOrderID(ord.ID).
			SetQuantity(item.Quantity).
			SetMenuItemID(item.MenuItemID).
			SetItemName(item.ItemName).
			SetItemPrice(item.ItemPrice).
			Save(ctx); err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	return r.GetByID(ctx, ord.ID, WithOrderItems())
}
```

Called on its own (no active `WithinTx` on `ctx`), `clientFromContext` just
returns `r.client` — identical behavior to today, but the atomicity of the
`Order` + `OrderItem` writes now comes from whoever calls this inside a
`WithinTx`, not from the repo itself.

**A service composing two repos atomically.** `OrderService.CreateOrder`
wraps the order write and a stock decrement so both commit or neither does:

```go
type orderService struct {
	orderRepo    repos.OrderRepository
	menuItemRepo repos.MenuItemRepository
	transactor   repos.Transactor
}
```

Since `Create` needs to return the created order, use `WithinTxResult`
instead of `WithinTx` directly:

```go
func (s *orderService) CreateOrder(ctx context.Context, data *repos.CreateOrderData) (*dto.Order, error) {
	return repos.WithinTxResult(ctx, s.transactor, func(ctx context.Context) (*dto.Order, error) {
		for _, item := range data.OrderItems {
			if err := s.menuItemRepo.DecrementStock(ctx, item.MenuItemID, item.Quantity); err != nil {
				return nil, err // rolls back the whole transaction, including any order already created below
			}
		}
		return s.orderRepo.Create(ctx, data)
	})
}
```

**A service calling another service, same `ctx`.** `OrderService.CreateOrder`
also awards loyalty points through `LoyaltyService` — a different service,
same atomic unit, as long as `ctx` is forwarded unchanged:

```go
func (s *orderService) CreateOrder(ctx context.Context, data *repos.CreateOrderData) (*dto.Order, error) {
	return repos.WithinTxResult(ctx, s.transactor, func(ctx context.Context) (*dto.Order, error) {
		ord, err := s.orderRepo.Create(ctx, data)
		if err != nil {
			return nil, err
		}
		// LoyaltyService.AwardPoints opens its own WithinTx internally, but
		// detects the transaction already active on ctx and just runs
		// inline — no second transaction, no separate commit.
		if err := s.loyaltyService.AwardPoints(ctx, data.RestaurantID, ord.ID); err != nil {
			return nil, err // order creation above rolls back too
		}
		return ord, nil
	})
}
```

If `AwardPoints` had instead been called with `context.Background()` here,
it would silently run in its own separate transaction — a failure there
would no longer roll back the order. This is the one failure mode the
"context propagation is mandatory" rule in the behavior contract above
exists to prevent.
