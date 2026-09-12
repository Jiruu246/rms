package services

import "context"

// noopTransactor satisfies repos.Transactor for service unit tests that mock
// out every repo — there's no real storage engine backing them, so it just
// runs fn inline instead of opening a transaction.
type noopTransactor struct{}

func (noopTransactor) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
