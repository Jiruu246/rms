package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/ent"
)

type entClientCtxKey struct{}

// clientFromContext resolves the *ent.Client a repo should use for this
// call: the transaction-scoped client if a Transactor.WithinTx is active on
// ctx, otherwise def (the repo's own default client). Repos are expected to
// call this at the top of every method instead of holding a fixed client.
func clientFromContext(ctx context.Context, def *ent.Client) *ent.Client {
	if txClient, inTx := txClientFromContext(ctx); inTx {
		return txClient
	}
	return def
}

// txClientFromContext reports whether ctx carries a transaction-scoped
// *ent.Client set by an active WithinTx call, returning it if so.
func txClientFromContext(ctx context.Context) (*ent.Client, bool) {
	txClient, ok := ctx.Value(entClientCtxKey{}).(*ent.Client)
	return txClient, ok
}

// EntTransactor is the Ent-backed Transactor. ent.Tx.Client() returns a
// *ent.Client scoped to the transaction — the same type repos already hold —
// so resolving "the current client" from context is a plain type switch, not
// a separate transactional code path.
type EntTransactor struct {
	client *ent.Client
}

func NewEntTransactor(client *ent.Client) *EntTransactor {
	return &EntTransactor{client: client}
}

// WithinTx implements Transactor. A nested call (ctx already carries a
// transaction-scoped client) runs fn inline against that same transaction;
// only the outermost call begins a real *ent.Tx and owns commit/rollback.
func (t *EntTransactor) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	if _, alreadyInTx := txClientFromContext(ctx); alreadyInTx {
		return fn(ctx)
	}

	tx, err := t.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	txCtx := context.WithValue(ctx, entClientCtxKey{}, tx.Client())

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
