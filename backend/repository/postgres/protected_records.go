package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
	authusecase "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

type row interface{ Scan(...any) error }
type tx interface {
	Exec(context.Context, string, ...any) (pgconnTag, error)
	QueryRow(context.Context, string, ...any) row
	Commit(context.Context) error
	Rollback(context.Context) error
}
type pool interface {
	Begin(context.Context) (tx, error)
}
type reader interface {
	QueryRow(context.Context, string, ...any) row
	Query(context.Context, string, ...any) (rows, error)
}
type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}
type pgconnTag interface{ RowsAffected() int64 }

type transactionContextKey struct{}

type pgxTransaction interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Commit(context.Context) error
	Rollback(context.Context) error
}

// Storeは保護記録のPostgreSQL実装です。
type Store struct {
	pool   pool
	reader reader
}

// NewStoreはpoolを注入してStoreを作成します。
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:   newPGXPool(pool),
		reader: newPGXPool(pool),
	}
}

type pgxPool struct {
	begin func(context.Context) (pgxTransaction, error)
	pool  *pgxpool.Pool
}

func newPGXPool(pool *pgxpool.Pool) pgxPool {
	return pgxPool{
		begin: func(ctx context.Context) (pgxTransaction, error) { return pool.Begin(ctx) },
		pool:  pool,
	}
}

func (p pgxPool) Begin(ctx context.Context) (tx, error) {
	value, err := p.begin(ctx)

	return pgxTx{value}, err
}
func (p pgxPool) QueryRow(ctx context.Context, sql string, args ...any) row {
	return p.pool.QueryRow(ctx, sql, args...)
}
func (p pgxPool) Query(ctx context.Context, sql string, args ...any) (rows, error) {
	//nolint:sqlclosecheck // 呼出元のread-only repositoryがrowsを必ずcloseする。
	return p.pool.Query(ctx, sql, args...)
}

type pgxTx struct{ tx pgxTransaction }

func (t pgxTx) Exec(ctx context.Context, sql string, args ...any) (pgconnTag, error) {
	return t.tx.Exec(ctx, sql, args...)
}
func (t pgxTx) QueryRow(ctx context.Context, sql string, args ...any) row {
	return t.tx.QueryRow(ctx, sql, args...)
}
func (t pgxTx) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t pgxTx) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }

// ReplaceOAuthTransactionは旧cookie参照を無効化し、新規transactionを一つのtransactionで保存します。
func (s *Store) ReplaceOAuthTransaction(ctx context.Context, previousReferenceHash auth.Hash, record auth.OAuthTransaction) error {
	if err := record.Validate(); err != nil {
		return err
	}
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if len(previousReferenceHash) > 0 {
		if _, err := transaction.Exec(ctx, `UPDATE admin_oauth_transactions SET invalidated_at = $1 WHERE reference_hash = $2 AND consumed_at IS NULL AND invalidated_at IS NULL`, record.CreatedAt, previousReferenceHash); err != nil {
			return err
		}
	}
	_, err = transaction.Exec(ctx, `INSERT INTO admin_oauth_transactions (id, reference_hash, state_hash, nonce_hash, pkce_verifier_ciphertext, return_path, created_at, expires_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, record.ID, record.ReferenceHash, record.StateHash, record.NonceHash, record.PKCEVerifierCiphertext, record.ReturnPath, record.CreatedAt, record.ExpiresAt)
	if err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

// ConsumeOAuthTransactionは有効なtransactionを条件付き更新で一回だけ取得します。
func (s *Store) ConsumeOAuthTransaction(ctx context.Context, referenceHash, stateHash auth.Hash, now auth.Time) (auth.OAuthTransaction, bool, error) {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.OAuthTransaction{}, false, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	var record auth.OAuthTransaction
	err = transaction.QueryRow(ctx, `UPDATE admin_oauth_transactions SET consumed_at = $1 WHERE reference_hash = $2 AND state_hash = $3 AND consumed_at IS NULL AND invalidated_at IS NULL AND expires_at > $1 RETURNING id, reference_hash, state_hash, nonce_hash, pkce_verifier_ciphertext, return_path, created_at, expires_at`, now, referenceHash, stateHash).Scan(&record.ID, &record.ReferenceHash, &record.StateHash, &record.NonceHash, &record.PKCEVerifierCiphertext, &record.ReturnPath, &record.CreatedAt, &record.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.OAuthTransaction{}, false, nil
	}
	if err != nil {
		return auth.OAuthTransaction{}, false, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return auth.OAuthTransaction{}, false, err
	}

	return record, true, nil
}

func (s *Store) CreateAdminSession(ctx context.Context, session auth.AdminSession) error {
	if err := session.Validate(); err != nil {
		return err
	}
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	_, err = transaction.Exec(ctx, `INSERT INTO admin_sessions (id, reference_hash, authorized_email, csrf_token_hash, csrf_token_ciphertext, created_at, last_mutation_at, absolute_expires_at, idle_expires_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, session.ID, session.ReferenceHash, session.AuthorizedEmail, session.CSRFTokenHash, session.CSRFTokenCiphertext, session.CreatedAt, session.LastMutationAt, session.AbsoluteExpiresAt, session.IdleExpiresAt)
	if err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func (s *Store) FindAdminSession(ctx context.Context, referenceHash auth.Hash) (auth.AdminSession, bool, error) {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.AdminSession{}, false, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	var session auth.AdminSession
	err = transaction.QueryRow(ctx, `SELECT id, reference_hash, authorized_email, csrf_token_hash, csrf_token_ciphertext, created_at, last_mutation_at, absolute_expires_at, idle_expires_at, revoked_at FROM admin_sessions WHERE reference_hash = $1`, referenceHash).Scan(&session.ID, &session.ReferenceHash, &session.AuthorizedEmail, &session.CSRFTokenHash, &session.CSRFTokenCiphertext, &session.CreatedAt, &session.LastMutationAt, &session.AbsoluteExpiresAt, &session.IdleExpiresAt, &session.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.AdminSession{}, false, nil
	}
	if err != nil {
		return auth.AdminSession{}, false, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return auth.AdminSession{}, false, err
	}

	return session, true, nil
}

func (s *Store) RevokeAdminSession(ctx context.Context, referenceHash auth.Hash, now auth.Time) (bool, error) {
	return s.updateSession(ctx, `UPDATE admin_sessions SET revoked_at = $1 WHERE reference_hash = $2 AND revoked_at IS NULL AND absolute_expires_at > $1 AND idle_expires_at > $1`, now, referenceHash)
}

// RunAuthorizedAdminStateChangeは認可済みemailとCSRF hashを条件に業務更新を実行します。
func (s *Store) RunAuthorizedAdminStateChange(ctx context.Context, referenceHash auth.Hash, allowedEmail string, csrfTokenHash auth.Hash, now auth.Time, operation authusecase.AdminStateChangeOperation) (bool, error) {
	if operation == nil {
		return false, errors.New("admin state change operation is required")
	}
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	tag, err := transaction.Exec(ctx, `UPDATE admin_sessions SET last_mutation_at = $1, idle_expires_at = LEAST($1 + INTERVAL '30 minutes', absolute_expires_at) WHERE reference_hash = $2 AND authorized_email = $3 AND csrf_token_hash = $4 AND csrf_token_ciphertext IS NOT NULL AND revoked_at IS NULL AND absolute_expires_at > $1 AND idle_expires_at > $1`, now, referenceHash, allowedEmail, csrfTokenHash)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() != 1 {
		return false, nil
	}
	if err := operation(context.WithValue(ctx, transactionContextKey{}, transaction)); err != nil {
		return false, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}

func transactionFromContext(ctx context.Context) (tx, bool) {
	transaction, ok := ctx.Value(transactionContextKey{}).(tx)

	return transaction, ok
}
func (s *Store) updateSession(ctx context.Context, sql string, args ...any) (bool, error) {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	tag, err := transaction.Exec(ctx, sql, args...)
	if err != nil {
		return false, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return false, err
	}

	return tag.RowsAffected() == 1, nil
}

// DeleteExpiredProtectedRecordsは認証と独立した保守操作です。nowから24時間保持期間を固定算出します。
func (s *Store) DeleteExpiredProtectedRecords(ctx context.Context, now auth.Time) error {
	cutoff := now.Add(-auth.ProtectedRecordRetention)
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if _, err := transaction.Exec(ctx, `DELETE FROM admin_oauth_transactions WHERE expires_at < $1`, cutoff); err != nil {
		return err
	}
	if _, err := transaction.Exec(ctx, `DELETE FROM admin_sessions WHERE (revoked_at IS NOT NULL AND revoked_at < $1) OR absolute_expires_at < $1`, cutoff); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}
