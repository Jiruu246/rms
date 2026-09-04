package repos

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/enttest"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
)

// ent hardcodes "sqlite3" as both the Go sql driver name and its internal
// SQLite dialect flavor (see dialect.SQLite), so the pure-Go driver — which
// self-registers as "sqlite" — must additionally be registered under that
// exact name for ent.Open to recognize it.
func init() {
	sql.Register("sqlite3", &sqlite.Driver{})
}

func newTestClient(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	return client
}

// TestEntTransactor_Commit shows the target shape: two "repo calls" (a User
// create and a Restaurant create, standing in for two different repos) run
// against the client resolved from ctx and both commit together.
func TestEntTransactor_Commit(t *testing.T) {
	client := newTestClient(t)
	transactor := NewEntTransactor(client)
	ctx := context.Background()

	err := transactor.WithinTx(ctx, func(ctx context.Context) error {
		c := clientFromContext(ctx, client)

		user, err := c.User.Create().SetName("Alice").Save(ctx)
		if err != nil {
			return err
		}

		_, err = c.Restaurant.Create().
			SetName("Alice's Diner").
			SetPhone("555-0100").
			SetEmail("alice@example.com").
			SetAddress("1 Main St").
			SetCity("Springfield").
			SetState("IL").
			SetZipCode("00000").
			SetCountry("US").
			SetCurrency("USD").
			SetUserID(user.ID).
			Save(ctx)
		return err
	})

	require.NoError(t, err)
	count, err := client.Restaurant.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// TestEntTransactor_RollbackOnError shows that an error from the second
// "repo call" undoes the first one's write too — the whole closure is one
// atomic unit, not two independent operations.
func TestEntTransactor_RollbackOnError(t *testing.T) {
	client := newTestClient(t)
	transactor := NewEntTransactor(client)
	ctx := context.Background()

	sentinel := errors.New("simulated failure in second repo call")
	err := transactor.WithinTx(ctx, func(ctx context.Context) error {
		c := clientFromContext(ctx, client)

		if _, err := c.User.Create().SetName("Bob").Save(ctx); err != nil {
			return err
		}
		return sentinel
	})

	require.ErrorIs(t, err, sentinel)
	count, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, count, "the User created before the error must have been rolled back")
}

// TestEntTransactor_RollbackOnPanic shows a panicking closure still rolls
// back cleanly, and the panic itself is not swallowed.
func TestEntTransactor_RollbackOnPanic(t *testing.T) {
	client := newTestClient(t)
	transactor := NewEntTransactor(client)
	ctx := context.Background()

	require.Panics(t, func() {
		_ = transactor.WithinTx(ctx, func(ctx context.Context) error {
			c := clientFromContext(ctx, client)
			if _, err := c.User.Create().SetName("Carol").Save(ctx); err != nil {
				t.Fatal(err)
			}
			panic("boom")
		})
	})

	count, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, count, "the User created before the panic must have been rolled back")
}

// TestEntTransactor_Nested models a service calling another service inside
// the same WithinTx (e.g. ServiceA.DoThing calling ServiceB.DoOtherThing with
// the same ctx). The inner WithinTx must detect the transaction already in
// ctx and run inline — not open a second *ent.Tx — so an error from the
// inner call still rolls back everything the outer call had already done.
func TestEntTransactor_Nested(t *testing.T) {
	client := newTestClient(t)
	transactor := NewEntTransactor(client)
	ctx := context.Background()

	sentinel := errors.New("simulated failure in nested service call")

	outerServiceA := func(ctx context.Context) error {
		return transactor.WithinTx(ctx, func(ctx context.Context) error {
			c := clientFromContext(ctx, client)
			if _, err := c.User.Create().SetName("Dave").Save(ctx); err != nil {
				return err
			}

			innerServiceB := func(ctx context.Context) error {
				return transactor.WithinTx(ctx, func(ctx context.Context) error {
					return sentinel
				})
			}
			return innerServiceB(ctx)
		})
	}

	err := outerServiceA(ctx)

	require.ErrorIs(t, err, sentinel)
	count, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, count, "outer service's write must roll back when the nested service call fails")
}
