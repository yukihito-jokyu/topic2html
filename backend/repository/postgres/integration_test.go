package postgres

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

var integrationDatabaseURL = flag.String("integration-database-url", "", "PostgreSQL URL for repository integration tests")

// TestAdminAuthSchemaIntegrationは実PostgreSQLで保護記録を検証します。
func TestAdminAuthSchemaIntegration(t *testing.T) {
	if *integrationDatabaseURL == "" {
		t.Skip("integration-database-url is not configured")
	}
	ctx := context.Background()
	pool := newIntegrationPool(t, ctx)
	failingPool := newIntegrationPool(t, ctx)
	if err := applyAdminAuthSchemaWithDDL(ctx, newPGXPool(failingPool), adminAuthSchemaDDL+"\nCREATE TABLE ();\n"); err == nil {
		t.Fatal("failing migration DDL succeeded")
	}
	assertMigrationRollback(t, ctx, failingPool)
	if err := ApplyAdminAuthSchema(ctx, failingPool); err != nil {
		t.Fatalf("migration could not be rerun after rollback: %v", err)
	}
	if err := ApplyAdminAuthSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := ApplyAdminAuthSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	var version string
	if err := pool.QueryRow(ctx, `SELECT version FROM schema_migrations WHERE version = $1`, AdminAuthSchemaMigration).Scan(&version); err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	first := integrationTransaction(now, 1)
	second := integrationTransaction(now, 2)
	if err := store.ReplaceOAuthTransaction(ctx, nil, first); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceOAuthTransaction(ctx, first.ReferenceHash, second); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.ConsumeOAuthTransaction(ctx, first.ReferenceHash, first.StateHash, now); err != nil || found {
		t.Fatalf("invalidated transaction: found=%t err=%v", found, err)
	}
	if transaction, found, err := store.ConsumeOAuthTransaction(ctx, second.ReferenceHash, second.StateHash, now); err != nil || !found || transaction.ID != second.ID {
		t.Fatalf("consume transaction: found=%t id=%q err=%v", found, transaction.ID, err)
	}
	if _, found, err := store.ConsumeOAuthTransaction(ctx, second.ReferenceHash, second.StateHash, now); err != nil || found {
		t.Fatalf("reused transaction: found=%t err=%v", found, err)
	}
	concurrent := integrationTransaction(now, 3)
	if err := store.ReplaceOAuthTransaction(ctx, nil, concurrent); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan struct {
		found bool
		err   error
	}, 2)
	var consumers sync.WaitGroup
	for range 2 {
		consumers.Add(1)
		go func() {
			defer consumers.Done()
			<-start
			_, found, err := store.ConsumeOAuthTransaction(ctx, concurrent.ReferenceHash, concurrent.StateHash, now)
			results <- struct {
				found bool
				err   error
			}{found, err}
		}()
	}
	close(start)
	consumers.Wait()
	close(results)
	consumed := 0
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.found {
			consumed++
		}
	}
	if consumed != 1 {
		t.Fatalf("concurrent consume count = %d, want 1", consumed)
	}
	session := integrationSession(now, 4)
	if err := store.CreateAdminSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAdminSession(ctx, session); err == nil {
		t.Fatal("duplicate session was stored")
	}
	if found, ok, err := store.FindAdminSession(ctx, session.ReferenceHash); err != nil || !ok || !bytes.Equal(found.CSRFTokenHash, session.CSRFTokenHash) {
		t.Fatalf("find session: found=%t err=%v", ok, err)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE admin_state_change_probe (label TEXT PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	changeTime := now.Add(10 * time.Minute)
	if updated, err := store.RunAdminStateChange(ctx, session.ReferenceHash, changeTime, func(operationContext context.Context) error {
		transaction, ok := transactionFromContext(operationContext)
		if !ok {
			return errors.New("state change transaction is unavailable")
		}
		_, err := transaction.Exec(operationContext, `INSERT INTO admin_state_change_probe (label) VALUES ($1)`, "success")

		return err
	}); err != nil || !updated {
		t.Fatalf("admin state change: updated=%t err=%v", updated, err)
	}
	assertProbeCount(t, ctx, pool, "success", 1)
	failedSession := integrationSession(now, 7)
	if err := store.CreateAdminSession(ctx, failedSession); err != nil {
		t.Fatal(err)
	}
	if updated, err := store.RunAdminStateChange(ctx, failedSession.ReferenceHash, changeTime, func(operationContext context.Context) error {
		transaction, ok := transactionFromContext(operationContext)
		if !ok {
			return errors.New("state change transaction is unavailable")
		}
		if _, err := transaction.Exec(operationContext, `INSERT INTO admin_state_change_probe (label) VALUES ($1)`, "failure"); err != nil {
			return err
		}

		return errors.New("business update failed")
	}); err == nil || updated {
		t.Fatalf("failed admin state change: updated=%t err=%v", updated, err)
	}
	assertProbeCount(t, ctx, pool, "failure", 0)
	assertSessionIdleDeadline(t, ctx, pool, failedSession.ReferenceHash, failedSession.LastMutationAt, failedSession.IdleExpiresAt)
	if revoked, err := store.RevokeAdminSession(ctx, session.ReferenceHash, now); err != nil || !revoked {
		t.Fatalf("revoke session: revoked=%t err=%v", revoked, err)
	}
	if revoked, err := store.RevokeAdminSession(ctx, session.ReferenceHash, now); err != nil || revoked {
		t.Fatalf("repeat revoke: revoked=%t err=%v", revoked, err)
	}
	expiredTransaction := integrationTransaction(now.Add(-auth.ProtectedRecordRetention-auth.OAuthTransactionLifetime-time.Microsecond), 5)
	expiredSession := integrationSession(now.Add(-auth.ProtectedRecordRetention-auth.SessionAbsoluteLifetime-time.Microsecond), 6)
	boundaryTransaction := integrationTransaction(now.Add(-auth.ProtectedRecordRetention-auth.OAuthTransactionLifetime), 8)
	boundarySession := integrationSession(now.Add(-auth.ProtectedRecordRetention-auth.SessionAbsoluteLifetime), 9)
	if err := store.ReplaceOAuthTransaction(ctx, nil, expiredTransaction); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAdminSession(ctx, expiredSession); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceOAuthTransaction(ctx, nil, boundaryTransaction); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAdminSession(ctx, boundarySession); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteExpiredProtectedRecords(ctx, now); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		query string
		hash  auth.Hash
	}{
		{`SELECT COUNT(*) FROM admin_oauth_transactions WHERE reference_hash = $1`, expiredTransaction.ReferenceHash},
		{`SELECT COUNT(*) FROM admin_sessions WHERE reference_hash = $1`, expiredSession.ReferenceHash},
	} {
		var count int
		if err := pool.QueryRow(ctx, check.query, check.hash).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("expired protected record remains: %d", count)
		}
	}
	for _, check := range []struct {
		query string
		hash  auth.Hash
	}{
		{`SELECT COUNT(*) FROM admin_oauth_transactions WHERE reference_hash = $1`, boundaryTransaction.ReferenceHash},
		{`SELECT COUNT(*) FROM admin_sessions WHERE reference_hash = $1`, boundarySession.ReferenceHash},
	} {
		var count int
		if err := pool.QueryRow(ctx, check.query, check.hash).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("24-hour boundary record count = %d, want 1", count)
		}
	}
}

func newIntegrationPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	bootstrap, err := pgxpool.New(ctx, *integrationDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("topic2html_integration_%d", time.Now().UnixNano())
	if _, err := bootstrap.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		bootstrap.Close()
		t.Fatal(err)
	}
	bootstrap.Close()
	config, err := pgxpool.ParseConfig(*integrationDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = 1
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanup, err := pgxpool.New(ctx, *integrationDatabaseURL)
		if err == nil {
			_, _ = cleanup.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`)
			cleanup.Close()
		}
	})

	return pool
}

func assertMigrationRollback(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, relation := range []string{
		"schema_migrations", "admin_oauth_transactions", "admin_oauth_transactions_expires_at_idx",
		"admin_sessions", "admin_sessions_absolute_expires_at_idx", "admin_sessions_revoked_at_idx",
	} {
		var missing bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass($1) IS NULL`, relation).Scan(&missing); err != nil {
			t.Fatal(err)
		}
		if !missing {
			t.Fatalf("migration rollback left relation %q", relation)
		}
	}
}

func assertProbeCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string, want int) {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_state_change_probe WHERE label = $1`, label).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("probe count for %q = %d, want %d", label, count, want)
	}
}

func assertSessionIdleDeadline(t *testing.T, ctx context.Context, pool *pgxpool.Pool, hash auth.Hash, wantMutation, wantIdle time.Time) {
	t.Helper()
	var mutation, idle time.Time
	if err := pool.QueryRow(ctx, `SELECT last_mutation_at, idle_expires_at FROM admin_sessions WHERE reference_hash = $1`, hash).Scan(&mutation, &idle); err != nil {
		t.Fatal(err)
	}
	if !mutation.Equal(wantMutation) || !idle.Equal(wantIdle) {
		t.Fatalf("rolled back session timestamps = %s / %s, want %s / %s", mutation, idle, wantMutation, wantIdle)
	}
}

func integrationTransaction(createdAt time.Time, sequence int) auth.OAuthTransaction {
	return auth.OAuthTransaction{
		ID:                     integrationID(sequence),
		ReferenceHash:          integrationHash(sequence, 1),
		StateHash:              integrationHash(sequence, 2),
		NonceHash:              integrationHash(sequence, 3),
		PKCEVerifierCiphertext: auth.Ciphertext(integrationHash(sequence, 4)),
		ReturnPath:             "/admin",
		CreatedAt:              createdAt,
		ExpiresAt:              createdAt.Add(auth.OAuthTransactionLifetime),
	}
}

func integrationSession(createdAt time.Time, sequence int) auth.AdminSession {
	return auth.AdminSession{
		ID:                integrationID(sequence),
		ReferenceHash:     integrationHash(sequence, 5),
		AuthorizedEmail:   "admin@example.test",
		CSRFTokenHash:     integrationHash(sequence, 6),
		CreatedAt:         createdAt,
		LastMutationAt:    createdAt,
		AbsoluteExpiresAt: createdAt.Add(auth.SessionAbsoluteLifetime),
		IdleExpiresAt:     createdAt.Add(auth.SessionIdleLifetime),
	}
}

func integrationHash(sequence, field int) auth.Hash {
	return auth.Hash(fmt.Sprintf("integration-%d-%d-%d", time.Now().UnixNano(), sequence, field))
}

func integrationID(sequence int) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012x", time.Now().UnixNano()&0xffffffffff+int64(sequence))
}
