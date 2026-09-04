package repos

import "context"

// Transactor runs fn atomically. It is the port services depend on to
// compose multiple repo calls — and calls into other services — into a
// single unit of work, without the service layer knowing anything about the
// underlying storage engine.
//
// Implementations must detect whether a transaction is already active on
// ctx: the outermost WithinTx call opens it and owns commit/rollback; any
// WithinTx call nested inside it (the same service calling itself, or a
// different service reached through the same ctx) just runs fn inline
// against the already-open transaction.
//
// Every call in the chain must forward the ctx it was given — swapping in a
// fresh context anywhere breaks atomicity for everything downstream, with no
// error raised.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(context.Context) error) error
}

// WithinTxResult wraps Transactor for callers that need to return a value
// alongside the error. It exists so a caller can't accidentally read result
// before checking err: result is only assigned once fn has succeeded, and is
// only returned once WithinTx itself has returned nil.
func WithinTxResult[T any](ctx context.Context, tr Transactor, fn func(context.Context) (T, error)) (T, error) {
	var result T

	callback := func(ctx context.Context) error {
		v, err := fn(ctx)
		if err != nil {
			return err
		}
		result = v
		return nil
	}

	err := tr.WithinTx(ctx, callback)
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}
