package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

type fakeTag int64

func (n fakeTag) RowsAffected() int64 { return int64(n) }

type fakeRow struct{ err error }

func (r fakeRow) Scan(...any) error { return r.err }

type fakePGXTx struct{}

func (fakePGXTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("UPDATE 1"), nil
}
func (fakePGXTx) QueryRow(context.Context, string, ...any) pgx.Row { return fakeRow{} }
func (fakePGXTx) Commit(context.Context) error                     { return nil }
func (fakePGXTx) Rollback(context.Context) error                   { return nil }

type fakeTx struct {
	exec   []error
	rows   []error
	commit error
	tag    pgconnTag
	calls  []string
}

func (t *fakeTx) Exec(_ context.Context, sql string, _ ...any) (pgconnTag, error) {
	t.calls = append(t.calls, sql)
	if len(t.exec) == 0 {
		if t.tag != nil {
			return t.tag, nil
		}

		return fakeTag(1), nil
	}
	e := t.exec[0]
	t.exec = t.exec[1:]
	if t.tag != nil {
		return t.tag, e
	}

	return fakeTag(1), e
}
func (t *fakeTx) QueryRow(_ context.Context, sql string, _ ...any) row {
	t.calls = append(t.calls, sql)
	if len(t.rows) == 0 {
		return fakeRow{}
	}
	e := t.rows[0]
	t.rows = t.rows[1:]

	return fakeRow{e}
}
func (t *fakeTx) Commit(context.Context) error   { return t.commit }
func (t *fakeTx) Rollback(context.Context) error { return nil }

type fakePool struct {
	transaction *fakeTx
	err         error
}

func (p fakePool) Begin(context.Context) (tx, error) { return p.transaction, p.err }
func testStore(tx *fakeTx) *Store {
	return &Store{
		pool: fakePool{
			transaction: tx,
		},
	}
}
func testTransaction() auth.OAuthTransaction {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	return auth.OAuthTransaction{
		ID:                     "00000000-0000-0000-0000-000000000001",
		ReferenceHash:          auth.Hash{1},
		StateHash:              auth.Hash{2},
		NonceHash:              auth.Hash{3},
		PKCEVerifierCiphertext: auth.Ciphertext{4},
		ReturnPath:             "/admin",
		CreatedAt:              now,
		ExpiresAt:              now.Add(auth.OAuthTransactionLifetime),
	}
}
func testSession() auth.AdminSession {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	return auth.AdminSession{
		ID:                  "00000000-0000-0000-0000-000000000001",
		ReferenceHash:       auth.Hash{1},
		AuthorizedEmail:     "admin@example.test",
		CSRFTokenHash:       auth.Hash{2},
		CSRFTokenCiphertext: auth.Ciphertext{3},
		CreatedAt:           now,
		LastMutationAt:      now,
		AbsoluteExpiresAt:   now.Add(auth.SessionAbsoluteLifetime),
		IdleExpiresAt:       now.Add(auth.SessionIdleLifetime),
	}
}

func TestStoreWrites(t *testing.T) {
	ctx := context.Background()
	if err := testStore(&fakeTx{}).ReplaceOAuthTransaction(ctx, nil, testTransaction()); err != nil {
		t.Fatal(err)
	}
	if err := testStore(&fakeTx{}).ReplaceOAuthTransaction(ctx, auth.Hash{9}, testTransaction()); err != nil {
		t.Fatal(err)
	}
	if err := testStore(&fakeTx{}).ReplaceOAuthTransaction(ctx, nil, auth.OAuthTransaction{}); err == nil {
		t.Fatal("wanted validation error")
	}
	if err := testStore(&fakeTx{}).CreateAdminSession(ctx, testSession()); err != nil {
		t.Fatal(err)
	}
	if err := testStore(&fakeTx{}).CreateAdminSession(ctx, auth.AdminSession{}); err == nil {
		t.Fatal("wanted validation error")
	}
	if err := testStore(&fakeTx{}).DeleteExpiredProtectedRecords(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestStoreReadsAndUpdates(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	hash := auth.Hash{1}
	if _, ok, err := testStore(&fakeTx{
		rows: []error{pgx.ErrNoRows},
	}).ConsumeOAuthTransaction(ctx, hash, hash, now); err != nil || ok {
		t.Fatalf("consume missing: %v %t", err, ok)
	}
	if _, ok, err := testStore(&fakeTx{}).ConsumeOAuthTransaction(ctx, hash, hash, now); err != nil || !ok {
		t.Fatalf("consume: %v %t", err, ok)
	}
	if _, ok, err := testStore(&fakeTx{
		rows: []error{pgx.ErrNoRows},
	}).FindAdminSession(ctx, hash); err != nil || ok {
		t.Fatalf("find missing: %v %t", err, ok)
	}
	if _, ok, err := testStore(&fakeTx{}).FindAdminSession(ctx, hash); err != nil || !ok {
		t.Fatalf("find: %v %t", err, ok)
	}
	for _, call := range []func() (bool, error){func() (bool, error) { return testStore(&fakeTx{}).RevokeAdminSession(ctx, hash, now) }} {
		if ok, err := call(); err != nil || !ok {
			t.Fatalf("update: %v %t", err, ok)
		}
	}
	if updated, err := testStore(&fakeTx{}).RunAuthorizedAdminStateChange(ctx, hash, "admin@example.test", hash, now, func(operationContext context.Context) error {
		if _, ok := transactionFromContext(operationContext); !ok {
			return errors.New("missing transaction")
		}

		return nil
	}); err != nil || !updated {
		t.Fatalf("authorized update: %v %t", err, updated)
	}
	if _, err := testStore(&fakeTx{}).RunAuthorizedAdminStateChange(ctx, hash, "admin@example.test", hash, now, nil); err == nil {
		t.Fatal("nil authorized admin state change operation succeeded")
	}
	if updated, err := testStore(&fakeTx{tag: fakeTag(0)}).RunAuthorizedAdminStateChange(ctx, hash, "admin@example.test", hash, now, func(context.Context) error { return nil }); err != nil || updated {
		t.Fatalf("missing authorized session state change: updated=%t err=%v", updated, err)
	}
}

func TestStoreFailures(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("boom")
	record := testTransaction()
	session := testSession()
	hash := auth.Hash{1}
	now := time.Now()
	for _, call := range []func() error{
		func() error {
			return (&Store{
				pool: fakePool{
					err: boom,
				},
			}).ReplaceOAuthTransaction(ctx, nil, record)
		},
		func() error {
			return testStore(&fakeTx{
				exec: []error{boom},
			}).ReplaceOAuthTransaction(ctx, nil, record)
		},
		func() error {
			return testStore(&fakeTx{
				exec: []error{boom},
			}).ReplaceOAuthTransaction(ctx, auth.Hash{9}, record)
		},
		func() error {
			return testStore(&fakeTx{
				commit: boom,
			}).ReplaceOAuthTransaction(ctx, nil, record)
		},
		func() error {
			return (&Store{
				pool: fakePool{
					err: boom,
				},
			}).CreateAdminSession(ctx, session)
		},
		func() error {
			return testStore(&fakeTx{
				exec: []error{boom},
			}).CreateAdminSession(ctx, session)
		},
		func() error {
			return testStore(&fakeTx{
				commit: boom,
			}).CreateAdminSession(ctx, session)
		},
		func() error {
			_, _, e := (&Store{
				pool: fakePool{
					err: boom,
				},
			}).ConsumeOAuthTransaction(ctx, hash, hash, now)

			return e
		},
		func() error {
			_, _, e := testStore(&fakeTx{
				rows: []error{boom},
			}).ConsumeOAuthTransaction(ctx, hash, hash, now)

			return e
		},
		func() error {
			_, _, e := testStore(&fakeTx{
				commit: boom,
			}).ConsumeOAuthTransaction(ctx, hash, hash, now)

			return e
		},
		func() error {
			_, _, e := (&Store{
				pool: fakePool{
					err: boom,
				},
			}).FindAdminSession(ctx, hash)

			return e
		},
		func() error {
			_, _, e := testStore(&fakeTx{
				rows: []error{boom},
			}).FindAdminSession(ctx, hash)

			return e
		},
		func() error {
			_, _, e := testStore(&fakeTx{
				commit: boom,
			}).FindAdminSession(ctx, hash)

			return e
		},
		func() error {
			_, e := (&Store{
				pool: fakePool{
					err: boom,
				},
			}).RevokeAdminSession(ctx, hash, now)

			return e
		},
		func() error {
			_, e := testStore(&fakeTx{
				exec: []error{boom},
			}).RevokeAdminSession(ctx, hash, now)

			return e
		},
		func() error {
			_, e := testStore(&fakeTx{
				commit: boom,
			}).RevokeAdminSession(ctx, hash, now)

			return e
		},
		func() error {
			_, e := (&Store{pool: fakePool{err: boom}}).RunAuthorizedAdminStateChange(ctx, hash, "admin@example.test", hash, now, func(context.Context) error { return nil })

			return e
		},
		func() error {
			_, e := testStore(&fakeTx{exec: []error{boom}}).RunAuthorizedAdminStateChange(ctx, hash, "admin@example.test", hash, now, func(context.Context) error { return nil })

			return e
		},
		func() error {
			_, e := testStore(&fakeTx{}).RunAuthorizedAdminStateChange(ctx, hash, "admin@example.test", hash, now, func(context.Context) error { return boom })

			return e
		},
		func() error {
			_, e := testStore(&fakeTx{commit: boom}).RunAuthorizedAdminStateChange(ctx, hash, "admin@example.test", hash, now, func(context.Context) error { return nil })

			return e
		},
		func() error {
			return (&Store{
				pool: fakePool{
					err: boom,
				},
			}).DeleteExpiredProtectedRecords(ctx, now)
		},
		func() error {
			return testStore(&fakeTx{
				exec: []error{boom},
			}).DeleteExpiredProtectedRecords(ctx, now)
		},
		func() error {
			return testStore(&fakeTx{
				exec: []error{nil, boom},
			}).DeleteExpiredProtectedRecords(ctx, now)
		},
		func() error {
			return testStore(&fakeTx{
				commit: boom,
			}).DeleteExpiredProtectedRecords(ctx, now)
		},
	} {
		if err := call(); !errors.Is(err, boom) {
			t.Fatalf("error=%v", err)
		}
	}
}

func TestPGXAdapters(t *testing.T) {
	ctx := context.Background()
	transaction, err := (pgxPool{
		begin: func(context.Context) (pgxTransaction, error) {
			return fakePGXTx{}, nil
		},
	}).Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.Exec(ctx, "UPDATE example"); err != nil {
		t.Fatal(err)
	}
	if err := transaction.QueryRow(ctx, "SELECT example").Scan(); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, "postgres://localhost:1/topic2html")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := NewStore(pool)
	if store == nil {
		t.Fatal("NewStore returned nil")
	}
	failedContext, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	if _, err := store.pool.Begin(failedContext); err == nil {
		t.Fatal("expected unavailable database error")
	}
	if err := ApplyAdminAuthSchema(failedContext, pool); err == nil {
		t.Fatal("expected unavailable database error")
	}
}
