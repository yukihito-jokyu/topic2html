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
	"github.com/yukihito-jokyu/topic2html/backend/domain/generation"
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
	if err := applyAdminAuthSchema(ctx, newPGXPool(failingPool)); err != nil {
		t.Fatalf("001 migration could not be rerun after rollback: %v", err)
	}
	legacyHash := integrationHash(99, 1)
	legacyNow := time.Now().UTC()
	if _, err := failingPool.Exec(ctx, `INSERT INTO admin_sessions (id, reference_hash, authorized_email, csrf_token_hash, created_at, last_mutation_at, absolute_expires_at, idle_expires_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, integrationID(99), legacyHash, "admin@example.test", integrationHash(99, 2), legacyNow, legacyNow, legacyNow.Add(auth.SessionAbsoluteLifetime), legacyNow.Add(auth.SessionIdleLifetime)); err != nil {
		t.Fatal(err)
	}
	if err := ApplyMigrations(ctx, failingPool); err != nil {
		t.Fatalf("002 migration could not be applied: %v", err)
	}
	var legacyRevoked *time.Time
	if err := failingPool.QueryRow(ctx, `SELECT revoked_at FROM admin_sessions WHERE reference_hash = $1`, legacyHash).Scan(&legacyRevoked); err != nil || legacyRevoked == nil {
		t.Fatalf("legacy session was not revoked: %v", err)
	}
	if err := ApplyMigrations(ctx, failingPool); err != nil {
		t.Fatalf("002 migration could not be reapplied: %v", err)
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
	if err := pool.QueryRow(ctx, `SELECT version FROM schema_migrations WHERE version = $1`, AdminSessionCSRFCiphertextMigration).Scan(&version); err != nil {
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
	if updated, err := store.RunAuthorizedAdminStateChange(ctx, session.ReferenceHash, session.AuthorizedEmail, session.CSRFTokenHash, changeTime, func(operationContext context.Context) error {
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
	assertSessionIdleDeadline(t, ctx, pool, session.ReferenceHash, changeTime, changeTime.Add(auth.SessionIdleLifetime))
	if updated, err := store.RunAuthorizedAdminStateChange(ctx, session.ReferenceHash, session.AuthorizedEmail, integrationHash(4, 8), changeTime.Add(time.Minute), func(operationContext context.Context) error {
		return errors.New("must not run")
	}); err != nil || updated {
		t.Fatalf("csrf mismatch state change: updated=%t err=%v", updated, err)
	}
	assertProbeCount(t, ctx, pool, "csrf-mismatch", 0)
	assertSessionIdleDeadline(t, ctx, pool, session.ReferenceHash, changeTime, changeTime.Add(auth.SessionIdleLifetime))
	failedSession := integrationSession(now, 7)
	if err := store.CreateAdminSession(ctx, failedSession); err != nil {
		t.Fatal(err)
	}
	if updated, err := store.RunAuthorizedAdminStateChange(ctx, failedSession.ReferenceHash, failedSession.AuthorizedEmail, failedSession.CSRFTokenHash, changeTime, func(operationContext context.Context) error {
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

// TestAdminSessionCSRFCiphertextMigrationRollbackIntegrationは002のDDL・legacy失効・migration記録が同一transactionでrollbackされることを実PostgreSQLで検証します。
func TestAdminSessionCSRFCiphertextMigrationRollbackIntegration(t *testing.T) {
	if *integrationDatabaseURL == "" {
		t.Skip("integration-database-url is not configured")
	}
	ctx := context.Background()
	pool := newIntegrationPool(t, ctx)
	if err := applyAdminAuthSchema(ctx, newPGXPool(pool)); err != nil {
		t.Fatal(err)
	}
	legacyHash := integrationHash(100, 1)
	legacyNow := time.Now().UTC()
	if _, err := pool.Exec(ctx, `INSERT INTO admin_sessions (id, reference_hash, authorized_email, csrf_token_hash, created_at, last_mutation_at, absolute_expires_at, idle_expires_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, integrationID(100), legacyHash, "admin@example.test", integrationHash(100, 2), legacyNow, legacyNow, legacyNow.Add(auth.SessionAbsoluteLifetime), legacyNow.Add(auth.SessionIdleLifetime)); err != nil {
		t.Fatal(err)
	}
	if err := applyMigrationWithDDL(ctx, newPGXPool(pool), AdminSessionCSRFCiphertextMigration, adminSessionCSRFCiphertextDDL+"\nSELECT 1 / 0;\n"); err == nil {
		t.Fatal("failing 002 migration succeeded")
	}
	var ciphertextColumnExists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'admin_sessions' AND column_name = 'csrf_token_ciphertext')`).Scan(&ciphertextColumnExists); err != nil {
		t.Fatal(err)
	}
	if ciphertextColumnExists {
		t.Fatal("rollback left csrf_token_ciphertext column")
	}
	var legacyRevoked *time.Time
	if err := pool.QueryRow(ctx, `SELECT revoked_at FROM admin_sessions WHERE reference_hash = $1`, legacyHash).Scan(&legacyRevoked); err != nil {
		t.Fatal(err)
	}
	if legacyRevoked != nil {
		t.Fatal("rollback left legacy session revoked")
	}
	var migrationCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, AdminSessionCSRFCiphertextMigration).Scan(&migrationCount); err != nil {
		t.Fatal(err)
	}
	if migrationCount != 0 {
		t.Fatalf("rollback left 002 migration record count=%d", migrationCount)
	}
}

func TestGenerationRequestSchemaIntegration(t *testing.T) {
	if *integrationDatabaseURL == "" {
		t.Skip("integration-database-url is not configured")
	}
	ctx := context.Background()
	pool := newIntegrationPool(t, ctx)
	if err := ApplyMigrations(ctx, pool); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	store := NewStore(pool)
	session := integrationSession(now, 101)
	if err := store.CreateAdminSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	request := integrationGenerationRequest(now, 102)
	if updated, err := store.RunAuthorizedAdminStateChange(ctx, session.ReferenceHash, session.AuthorizedEmail, session.CSRFTokenHash, now, func(operationContext context.Context) error {
		return store.CreateRunningGenerationRequest(operationContext, request)
	}); err != nil || !updated {
		t.Fatalf("T1: updated=%t err=%v", updated, err)
	}
	if err := store.RecordFailedGenerationAttempt(ctx, integrationGenerationFailedAttempt(now, request.ID, 1, 103)); err != nil {
		t.Fatalf("T2: %v", err)
	}
	success := integrationGenerationSucceededAttempt(now, request.ID, 2, 104)
	candidate := generation.Candidate{
		ID:          integrationID(105),
		RequestID:   request.ID,
		HTML:        "<!doctype html><html><head></head><body>safe</body></html>",
		ValidatedAt: now,
		CreatedAt:   now,
	}
	if err := store.CompleteGenerationSucceeded(ctx, success, candidate, now); err != nil {
		t.Fatalf("T3: %v", err)
	}
	record, found, err := store.FindGenerationRequest(ctx, request.ID)
	if err != nil || !found || record.Request.State != generation.StateCompletedSucceeded || len(record.Attempts) != 2 || record.Candidate == nil || record.Candidate.ID != candidate.ID || record.Candidate.HTML != "" {
		t.Fatalf("read-only record=%+v found=%t err=%v", record, found, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE generated_html_candidates SET html = $1 WHERE id = $2`, "changed", candidate.ID); err == nil {
		t.Fatal("candidate UPDATE succeeded")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM generated_html_candidates WHERE id = $1`, candidate.ID); err == nil {
		t.Fatal("candidate DELETE succeeded")
	}
	failedRequest := integrationGenerationRequest(now, 107)
	if updated, err := store.RunAuthorizedAdminStateChange(ctx, session.ReferenceHash, session.AuthorizedEmail, session.CSRFTokenHash, now.Add(time.Second), func(operationContext context.Context) error {
		return store.CreateRunningGenerationRequest(operationContext, failedRequest)
	}); err != nil || !updated {
		t.Fatalf("failed request T1: updated=%t err=%v", updated, err)
	}
	for number := int16(1); number < 4; number++ {
		if err := store.RecordFailedGenerationAttempt(ctx, integrationGenerationFailedAttempt(now, failedRequest.ID, number, 110+int(number))); err != nil {
			t.Fatalf("failed request T2 number=%d: %v", number, err)
		}
	}
	if err := store.CompleteGenerationFailed(ctx, integrationGenerationFailedAttempt(now, failedRequest.ID, 4, 114), now); err != nil {
		t.Fatalf("T4: %v", err)
	}
	failedRecord, found, err := store.FindGenerationRequest(ctx, failedRequest.ID)
	if err != nil || !found || failedRecord.Request.State != generation.StateCompletedFailed || len(failedRecord.Attempts) != 4 || failedRecord.Candidate != nil {
		t.Fatalf("failed record=%+v found=%t err=%v", failedRecord, found, err)
	}
}

func TestGenerationRequestSchemaRollbackAndConstraintsIntegration(t *testing.T) {
	if *integrationDatabaseURL == "" {
		t.Skip("integration-database-url is not configured")
	}
	ctx := context.Background()
	pool := newIntegrationPool(t, ctx)
	if err := applyAdminAuthSchema(ctx, newPGXPool(pool)); err != nil {
		t.Fatal(err)
	}
	if err := applyMigrationWithDDL(ctx, newPGXPool(pool), AdminSessionCSRFCiphertextMigration, adminSessionCSRFCiphertextDDL); err != nil {
		t.Fatal(err)
	}
	if err := applyMigrationWithDDL(ctx, newPGXPool(pool), GenerationRequestSchemaMigration, generationRequestSchemaDDL+"\nSELECT 1 / 0;"); err == nil {
		t.Fatal("failing 003 migration succeeded")
	}
	for _, relation := range []string{"generation_requests", "generation_attempts", "generated_html_candidates", "generation_attempts_request_number_idx", "generation_requests_source_created_idx"} {
		var missing bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass($1) IS NULL`, relation).Scan(&missing); err != nil || !missing {
			t.Fatalf("rollback left relation %q: missing=%t err=%v", relation, missing, err)
		}
	}
	var functionExists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'reject_generated_html_candidate_mutation')`).Scan(&functionExists); err != nil || functionExists {
		t.Fatalf("rollback left function: exists=%t err=%v", functionExists, err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, GenerationRequestSchemaMigration).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rollback left migration: count=%d err=%v", count, err)
	}
	if err := ApplyMigrations(ctx, pool); err != nil {
		t.Fatal(err)
	}
	for _, index := range []string{"generation_attempts_request_number_idx", "generation_requests_source_created_idx"} {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, index).Scan(&exists); err != nil || !exists {
			t.Fatalf("index %q does not exist: exists=%t err=%v", index, exists, err)
		}
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	request := integrationGenerationRequest(now, 201)
	if _, err := pool.Exec(ctx, `INSERT INTO generation_requests (id, kind, topic, state, created_at) VALUES ($1,$2,$3,$4,$5)`, request.ID, request.Kind, *request.Topic, request.State, request.CreatedAt); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, started_at, completed_at) VALUES ($1,$2,$3,$4,$5,$6)`, integrationID(202), request.ID, 1, generation.OutcomeSucceeded, now, now); err != nil {
		t.Fatal(err)
	}
	constraintCases := []struct {
		name      string
		statement string
	}{
		{
			name:      "request kind",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, created_at) VALUES ('00000000-0000-0000-0000-000000000208', 'other', 'topic', 'running', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "request state",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, created_at) VALUES ('00000000-0000-0000-0000-000000000209', 'initial', 'topic', 'other', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "initial blank topic",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, created_at) VALUES ('00000000-0000-0000-0000-000000000210', 'initial', ' ', 'running', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "initial missing topic",
			statement: `INSERT INTO generation_requests (id, kind, state, created_at) VALUES ('00000000-0000-0000-0000-000000000211', 'initial', 'running', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "initial source version",
			statement: `INSERT INTO generation_requests (id, kind, topic, source_version_id, state, created_at) VALUES ('00000000-0000-0000-0000-000000000212', 'initial', 'topic', '00000000-0000-0000-0000-000000000212', 'running', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "revision topic",
			statement: `INSERT INTO generation_requests (id, kind, topic, instructions, source_version_id, state, created_at) VALUES ('00000000-0000-0000-0000-000000000213', 'revision', 'topic', 'detail', '00000000-0000-0000-0000-000000000213', 'running', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "revision missing source version",
			statement: `INSERT INTO generation_requests (id, kind, instructions, state, created_at) VALUES ('00000000-0000-0000-0000-000000000214', 'revision', 'detail', 'running', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "revision blank instructions",
			statement: `INSERT INTO generation_requests (id, kind, instructions, source_version_id, state, created_at) VALUES ('00000000-0000-0000-0000-000000000215', 'revision', ' ', '00000000-0000-0000-0000-000000000215', 'running', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "succeeded request failure fields",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, final_failure_code, final_failure_summary, created_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000209', 'initial', 'topic', 'completed_succeeded', 'invalid_html', 'failure', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "running request failure fields",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, final_failure_code, final_failure_summary, created_at) VALUES ('00000000-0000-0000-0000-000000000216', 'initial', 'topic', 'running', 'invalid_html', 'failure', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "running request completion",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, created_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000217', 'initial', 'topic', 'running', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "succeeded request missing completion",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, created_at) VALUES ('00000000-0000-0000-0000-000000000218', 'initial', 'topic', 'completed_succeeded', CURRENT_TIMESTAMP)`,
		},
		{
			name:      "failed request missing failure fields",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, created_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000219', 'initial', 'topic', 'completed_failed', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "failed request missing failure summary",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, final_failure_code, created_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000220', 'initial', 'topic', 'completed_failed', 'invalid_html', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "failed request failure code",
			statement: `INSERT INTO generation_requests (id, kind, topic, state, final_failure_code, final_failure_summary, created_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000221', 'initial', 'topic', 'completed_failed', 'other', 'failure', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "attempt request foreign key",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000203', '00000000-0000-0000-0000-000000000203', 1, 'succeeded', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "attempt number unique",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000204', '` + request.ID + `', 1, 'succeeded', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "attempt number below range",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000222', '` + request.ID + `', 0, 'succeeded', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "attempt number above range",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000210', '` + request.ID + `', 5, 'succeeded', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "attempt outcome",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000223', '` + request.ID + `', 2, 'other', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "succeeded attempt failure fields",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, failure_code, failure_summary, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000211', '` + request.ID + `', 2, 'succeeded', 'invalid_html', 'failure', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "failed attempt missing failure code",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, failure_summary, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000224', '` + request.ID + `', 2, 'failed', 'failure', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "failed attempt missing failure summary",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, failure_code, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000225', '` + request.ID + `', 2, 'failed', 'invalid_html', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "failed attempt failure code",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, failure_code, failure_summary, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000226', '` + request.ID + `', 2, 'failed', 'other', 'failure', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
		{
			name:      "attempt completion before start",
			statement: `INSERT INTO generation_attempts (id, generation_request_id, attempt_number, outcome, started_at, completed_at) VALUES ('00000000-0000-0000-0000-000000000212', '` + request.ID + `', 2, 'succeeded', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP - INTERVAL '1 second')`,
		},
		{
			name:      "candidate empty HTML",
			statement: `INSERT INTO generated_html_candidates (id, generation_request_id, html, validated_at, created_at) VALUES ('00000000-0000-0000-0000-000000000205', '` + request.ID + `', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		},
	}
	for _, testCase := range constraintCases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := pool.Exec(ctx, testCase.statement); err == nil {
				t.Fatalf("constraint violation succeeded: %s", testCase.statement)
			}
		})
	}
	if _, err := pool.Exec(ctx, `INSERT INTO generated_html_candidates (id, generation_request_id, html, validated_at, created_at) VALUES ($1,$2,$3,$4,$5)`, integrationID(206), request.ID, "<html></html>", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO generated_html_candidates (id, generation_request_id, html, validated_at, created_at) VALUES ($1,$2,$3,$4,$5)`, integrationID(207), request.ID, "<html></html>", now, now); err == nil {
		t.Fatal("candidate one-to-one violation succeeded")
	}
}

func TestGenerationT1SessionRollbackIntegration(t *testing.T) {
	if *integrationDatabaseURL == "" {
		t.Skip("integration-database-url is not configured")
	}
	ctx := context.Background()
	pool := newIntegrationPool(t, ctx)
	if err := ApplyMigrations(ctx, pool); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	store := NewStore(pool)
	session := integrationSession(now, 401)
	if err := store.CreateAdminSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `CREATE FUNCTION reject_generation_request_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'reject'; END; $$; CREATE TRIGGER reject_generation_request_insert BEFORE INSERT ON generation_requests FOR EACH ROW EXECUTE FUNCTION reject_generation_request_insert();`); err != nil {
		t.Fatal(err)
	}
	request := integrationGenerationRequest(now, 402)
	changeAt := now.Add(time.Minute)
	if updated, err := store.RunAuthorizedAdminStateChange(ctx, session.ReferenceHash, session.AuthorizedEmail, session.CSRFTokenHash, changeAt, func(operationContext context.Context) error {
		return store.CreateRunningGenerationRequest(operationContext, request)
	}); err == nil || updated {
		t.Fatalf("T1 failure: updated=%t err=%v", updated, err)
	}
	assertSessionIdleDeadline(t, ctx, pool, session.ReferenceHash, session.LastMutationAt, session.IdleExpiresAt)
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM generation_requests WHERE id = $1`, request.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("T1 rollback request count=%d err=%v", count, err)
	}
}

func TestGenerationTransactionRollbackIntegration(t *testing.T) {
	if *integrationDatabaseURL == "" {
		t.Skip("integration-database-url is not configured")
	}
	ctx := context.Background()
	pool := newIntegrationPool(t, ctx)
	if err := ApplyMigrations(ctx, pool); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	store := NewStore(pool)
	request := integrationGenerationRequest(now, 301)
	insertRunningIntegrationRequest(t, ctx, pool, request)
	if _, err := pool.Exec(ctx, `CREATE FUNCTION reject_generation_attempt_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'reject'; END; $$; CREATE TRIGGER reject_generation_attempt_insert BEFORE INSERT ON generation_attempts FOR EACH ROW EXECUTE FUNCTION reject_generation_attempt_insert();`); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordFailedGenerationAttempt(ctx, integrationGenerationFailedAttempt(now, request.ID, 1, 302)); err == nil {
		t.Fatal("T2 failure succeeded")
	}
	assertGenerationCounts(t, ctx, pool, request.ID, 0, 0, generation.StateRunning)
	if _, err := pool.Exec(ctx, `DROP TRIGGER reject_generation_attempt_insert ON generation_attempts; DROP FUNCTION reject_generation_attempt_insert();`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `CREATE FUNCTION reject_generation_candidate_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'reject'; END; $$; CREATE TRIGGER reject_generation_candidate_insert BEFORE INSERT ON generated_html_candidates FOR EACH ROW EXECUTE FUNCTION reject_generation_candidate_insert();`); err != nil {
		t.Fatal(err)
	}
	success := integrationGenerationSucceededAttempt(now, request.ID, 1, 303)
	candidate := generation.Candidate{
		ID:          integrationID(304),
		RequestID:   request.ID,
		HTML:        "<html></html>",
		ValidatedAt: now,
		CreatedAt:   now,
	}
	if err := store.CompleteGenerationSucceeded(ctx, success, candidate, now); err == nil {
		t.Fatal("T3 failure succeeded")
	}
	assertGenerationCounts(t, ctx, pool, request.ID, 0, 0, generation.StateRunning)
	if _, err := pool.Exec(ctx, `DROP TRIGGER reject_generation_candidate_insert ON generated_html_candidates; DROP FUNCTION reject_generation_candidate_insert();`); err != nil {
		t.Fatal(err)
	}
	failedRequest := integrationGenerationRequest(now, 305)
	insertRunningIntegrationRequest(t, ctx, pool, failedRequest)
	for number := int16(1); number < 4; number++ {
		if err := store.RecordFailedGenerationAttempt(ctx, integrationGenerationFailedAttempt(now, failedRequest.ID, number, 305+int(number))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, `CREATE FUNCTION reject_generation_request_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'reject'; END; $$; CREATE TRIGGER reject_generation_request_update BEFORE UPDATE ON generation_requests FOR EACH ROW EXECUTE FUNCTION reject_generation_request_update();`); err != nil {
		t.Fatal(err)
	}
	if err := store.CompleteGenerationFailed(ctx, integrationGenerationFailedAttempt(now, failedRequest.ID, 4, 309), now); err == nil {
		t.Fatal("T4 failure succeeded")
	}
	assertGenerationCounts(t, ctx, pool, failedRequest.ID, 3, 0, generation.StateRunning)
}

func insertRunningIntegrationRequest(t *testing.T, ctx context.Context, pool *pgxpool.Pool, request generation.Request) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO generation_requests (id, kind, topic, state, created_at) VALUES ($1,$2,$3,$4,$5)`, request.ID, request.Kind, *request.Topic, request.State, request.CreatedAt); err != nil {
		t.Fatal(err)
	}
}

func assertGenerationCounts(t *testing.T, ctx context.Context, pool *pgxpool.Pool, requestID string, attempts, candidates int, state generation.State) {
	t.Helper()
	var actualAttempts, actualCandidates int
	var actualState generation.State
	if err := pool.QueryRow(ctx, `SELECT (SELECT COUNT(*) FROM generation_attempts WHERE generation_request_id = $1), (SELECT COUNT(*) FROM generated_html_candidates WHERE generation_request_id = $1), state FROM generation_requests WHERE id = $1`, requestID).Scan(&actualAttempts, &actualCandidates, &actualState); err != nil || actualAttempts != attempts || actualCandidates != candidates || actualState != state {
		t.Fatalf("attempts=%d candidates=%d state=%q err=%v", actualAttempts, actualCandidates, actualState, err)
	}
}

func integrationGenerationRequest(now time.Time, sequence int) generation.Request {
	topic := "Go"

	return generation.Request{
		ID:        integrationID(sequence),
		Kind:      generation.KindInitial,
		Topic:     &topic,
		State:     generation.StateRunning,
		CreatedAt: now,
	}
}

func integrationGenerationFailedAttempt(now time.Time, requestID string, number int16, sequence int) generation.Attempt {
	code := generation.FailureGenerationUnavailable
	summary := "生成サービスに接続できませんでした。"

	return generation.Attempt{
		ID:             integrationID(sequence),
		RequestID:      requestID,
		Number:         number,
		Outcome:        generation.OutcomeFailed,
		FailureCode:    &code,
		FailureSummary: &summary,
		StartedAt:      now,
		CompletedAt:    now,
	}
}

func integrationGenerationSucceededAttempt(now time.Time, requestID string, number int16, sequence int) generation.Attempt {
	return generation.Attempt{
		ID:          integrationID(sequence),
		RequestID:   requestID,
		Number:      number,
		Outcome:     generation.OutcomeSucceeded,
		StartedAt:   now,
		CompletedAt: now,
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
		ID:                  integrationID(sequence),
		ReferenceHash:       integrationHash(sequence, 5),
		AuthorizedEmail:     "admin@example.test",
		CSRFTokenHash:       integrationHash(sequence, 6),
		CSRFTokenCiphertext: auth.Ciphertext(integrationHash(sequence, 7)),
		CreatedAt:           createdAt,
		LastMutationAt:      createdAt,
		AbsoluteExpiresAt:   createdAt.Add(auth.SessionAbsoluteLifetime),
		IdleExpiresAt:       createdAt.Add(auth.SessionIdleLifetime),
	}
}

func integrationHash(sequence, field int) auth.Hash {
	return auth.Hash(fmt.Sprintf("integration-%d-%d-%d", time.Now().UnixNano(), sequence, field))
}

func integrationID(sequence int) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012x", time.Now().UnixNano()&0xffffffffff+int64(sequence))
}
